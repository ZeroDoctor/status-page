package handler

import (
	"database/sql/driver"
	"fmt"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgresql db

	"github.com/zerodoctor/go-status/model"
	ppt "github.com/zerodoctor/goprettyprinter"
)

// DBHandler :
type DBHandler struct {
	local *sqlx.DB
}

// NewDBHandler :
func NewDBHandler() *DBHandler {

	user := os.Getenv("psql_user")
	password := os.Getenv("psql_password")
	host := os.Getenv("psql_host")
	dbname := os.Getenv("dbname")
	sslMode := os.Getenv("psql_sslMode")

	local, err := sqlx.Connect("postgres", "user="+user+" password="+password+" host="+host+" dbname="+dbname+" sslmode="+sslMode)
	if err != nil {
		ppt.Errorln("failed to connect to db:\n\t", err.Error())
		os.Exit(1)
	}

	schema := `
		CREATE TABLE IF NOT EXISTS apps (
			id          TEXT,
			name        TEXT,
			ip_address  TEXT,
			device_name TEXT,
			session     INTEGER,
			CONSTRAINT apps_pk PRIMARY KEY (id)
		);
	`

	_, err = local.Exec(schema)
	if err != nil {
		ppt.Errorln("failed to create table apps:\n\t", err.Error())
		os.Exit(1)
	}

	schema = `
		CREATE TABLE IF NOT EXISTS logs (
			id          SERIAL,
			type        TEXT,
			msg         TEXT,
			log_time    TIMESTAMP,
			file_name   TEXT,
			func_name   TEXT,
			line_number INTEGER,
			app_id      TEXT REFERENCES apps,
			app_name    TEXT,
			session     INTEGER,
			CONSTRAINT logs_pk PRIMARY KEY (id)
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
func (dbh *DBHandler) GenerateAppID(name, device, ip string) model.App {
	var app model.App

	id := RandString(12)
	for app.Name != "" {
		query := "SELECT * FROM apps WHERE app_id = ?;"
		dbh.local.Select(&app, query, id)
		id = RandString(12)
	}

	ppt.Infoln("Generated app id:", id)

	return model.App{
		ID:         id,
		Name:       name,
		DeviceName: device,
		IPAddress:  ip,
		Session:    0,
	}
}

// AppID :
func (dbh *DBHandler) AppID(id, name, device, ip string) model.App {

	var app model.App

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
func (dbh *DBHandler) SaveApp(app model.App) {
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

	dbh.execHandler(values, template)
}

// SaveLogs :
func (dbh *DBHandler) SaveLogs(logs []model.Log) []model.Log {
	template := `
		INSERT INTO logs (
			type, msg, log_time,
			file_name, func_name,
			line_number, app_id, 
			app_name, session
		) VALUES
	` + printRows(len(logs)) + `
		RETURNING id;
	`

	var values []driver.Valuer
	for _, log := range logs {
		values = append(values, log)
	}

	ids := dbh.queryHandler(values, template)
	for i := 0; i < len(ids); i++ {
		logs[i].Index = ids[i].(int64)
	}

	return logs
}

// FetchLogs :
func (dbh *DBHandler) FetchLogs() []model.Log {
	var logs []model.Log

	query := "SELECT * FROM logs ORDER BY log_time ASC LIMIT 50;"
	err := dbh.local.Select(&logs, query)
	if err != nil {
		ppt.Errorln("failed to query all logs:\n\t", err.Error())
	}

	return logs
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

func (dbh *DBHandler) execHandler(items []driver.Valuer, template string) {
	params := make([]interface{}, len(items))
	for i, b := range items {
		params[i] = b
	}

	query, args, err := sqlx.In(template, params...)
	if err != nil {
		ppt.Errorf("failed to bind query [exec]:\n\t%s\n", err.Error())
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

func (dbh *DBHandler) queryHandler(items []driver.Valuer, template string) []interface{} {
	params := make([]interface{}, len(items))
	for i, b := range items {
		params[i] = b
	}

	query, args, err := sqlx.In(template, params...)
	if err != nil {
		ppt.Errorf("failed to bind query [query]:\n\t%s\n", err.Error())
		return nil
	}

	query = dbh.local.Rebind(query)

	var interList []interface{}
	err = dbh.transact(func(tx *sqlx.Tx) error {
		rows, err := tx.Query(query, args...)
		if err != nil {
			return fmt.Errorf("failed writing(query) to table: \n\terror: %s", err.Error())
		}

		var inter interface{}
		for rows.Next() {
			err := rows.Scan(&inter)
			if err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan row - aborting process: %s", err.Error())
			}
			interList = append(interList, inter)
		}
		rows.Close()
		err = rows.Err()
		if err != nil {
			return fmt.Errorf("aborting - failed on processing rows: %s", err.Error())
		}

		return nil
	})
	if err != nil {
		ppt.Errorf("failed writing(query) to table\n\terror:%s\n", err.Error())
	}

	return interList
}
