package handler

import (
	"database/sql/driver"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // sqlite db

	ppt "github.com/zerodoctor/goprettyprinter"
)

// DBHandler :
type DBHandler struct {
	local *sqlx.DB
}

// NewDBHandler :
func NewDBHandler() *DBHandler {

	local, err := sqlx.Connect("sqlite3", "./db/local.db")
	if err != nil {
		ppt.Errorln("failed to connect to db:\n\t", err.Error())
		os.Exit(1)
	}

	schema := `
		PRAGMA foreign_keys = 1;
		CREATE TABLE IF NOT EXISTS apps (
			id          TEXT,
			name        TEXT,
			ip_address  TEXT,
			device_name TEXT,
			session     INTEGER,

			PRIMARY KEY (id)
		);
	`

	_, err = local.Exec(schema)
	if err != nil {
		ppt.Errorln("failed to create table apps:\n\t", err.Error())
		os.Exit(1)
	}

	schema = `
		CREATE TABLE IF NOT EXISTS logs (
			type        TEXT,
			msg         TEXT,
			log_time    TIMESTAMP,
			file_name   TEXT,
			func_name   TEXT,
			line_number INTEGER,
			'index'     INTEGER,
			app_id      TEXT,
			app_name    TEXT,
			session     INTEGER,

			FOREIGN KEY (app_id) REFERENCES apps( id )
		);
	`

	_, err = local.Exec(schema)
	if err != nil {
		ppt.Errorln("failed to create table apps:\n\t", err.Error())
		os.Exit(1)
	}

	return &DBHandler{
		local: local,
	}
}

// GenerateAppID :
func (dbh *DBHandler) GenerateAppID(name, device, ip string) App {
	var app App

	id := RandString(12)
	for app.Name != "" {
		query := "SELECT * FROM apps WHERE app_id = ?;"
		dbh.local.Select(&app, query, id)
		id = RandString(12)
	}

	ppt.Infoln("Generated app id:", id)

	return App{
		ID:         id,
		Name:       name,
		DeviceName: device,
		IPAddress:  ip,
		Session:    0,
	}
}

// AppID :
func (dbh *DBHandler) AppID(id, name, device, ip string) App {

	var app App

	if id == "" {
		return dbh.GenerateAppID(name, device, ip)
	}

	query := "SELECT * FROM apps WHERE app_id = ?;"
	dbh.local.Select(&app, query, id)

	if app.Name == "" {
		ppt.Warnln("failed to find app id:", id, "generating id...")
		return dbh.GenerateAppID(name, device, ip)
	}

	ppt.Infoln("Found app id:", id)

	return app
}

// SaveApp :
func (dbh *DBHandler) SaveApp(app App) {
	template := `
		INSERT INTO apps (
			id, name,
			ip_address,
			device_name,
			session
		) VALUES ` + printRows(1) + `
		ON CONFLICT(id) DO UPDATE SET
			name        = excluded.name,
			ip_address  = excluded.ip_address,
			device_name = excluded.device_name,
			session     = excluded.session
		`

	var values []driver.Valuer
	values = append(values, app)

	dbh.execQuery(values, template)
}

// SaveLogs :
func (dbh *DBHandler) SaveLogs(logs []Log) {
	template := `
		INSERT INTO logs (
			type, msg, log_time,
			file_name, func_name,
			line_number, 'index', 
			app_id, app_name, session
		) VALUES
	` + printRows(len(logs))

	var values []driver.Valuer
	for _, log := range logs {
		values = append(values, log)
	}

	dbh.execQuery(values, template)
}

// printRows : creates a string of '(?)' for each item submited to database
func printRows(total int) string {
	var str strings.Builder

	for i := 0; i < total; i++ {
		str.WriteString(" (?),")
	}

	return str.String()[:str.Len()-1] // remove trailing comma before returning
}

// transact : used to force connection reuse
func (dbh *DBHandler) transact(fn func(*sqlx.Tx) error) error {
	tx, err := dbh.local.Beginx()
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (dbh *DBHandler) execQuery(items []driver.Valuer, template string) {
	params := make([]interface{}, len(items))
	for i, b := range items {
		params[i] = b
	}

	query, args, err := sqlx.In(template, params...)
	if err != nil {
		ppt.Errorf("Failed to bind query:\n\t%+v\n", err)
		return
	}

	query = dbh.local.Rebind(query)

	err = dbh.transact(func(tx *sqlx.Tx) error {
		_, err = tx.Exec(query, args...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		ppt.Errorf("failed writing(exec) to table:\n\terror: %s\n", err.Error())
		return
	}
}
