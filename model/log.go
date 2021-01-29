package model

import "database/sql/driver"

// Log :
type Log struct {
	Type       string `db:"type" json:"log_type"`
	Msg        string `db:"msg" json:"msg"`
	LogTime    string `db:"log_time" json:"log_time"`
	FileName   string `db:"file_name" json:"file_name"`
	FuncName   string `db:"func_name" json:"func_name"`
	LineNumber int    `db:"line_number" json:"line_number"`
	Index      int64  `db:"id" json:"index"`
	Session    int    `db:"session" json:"session"`
	AppID      string `db:"app_id" json:"app_id"`
	AppName    string `db:"app_name" json:"app_name"`
}

// Value :
func (l Log) Value() (driver.Value, error) {
	return []driver.Value{
		l.Type,
		l.Msg,
		l.LogTime,
		l.FileName,
		l.FuncName,
		l.LineNumber,
		l.AppID,
		l.AppName,
		l.Session,
	}, nil
}

// App :
type App struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IPAddress  string `json:"ip_address"`
	DeviceName string `json:"device_name"`
	Session    int    `json:"session"`
	Logs       []Log  `json:"-"` // make db later
}

// Value :
func (a App) Value() (driver.Value, error) {
	return []driver.Value{
		a.ID,
		a.Name,
		a.IPAddress,
		a.DeviceName,
		a.Session,
	}, nil
}
