package model

import "database/sql/driver"

// Log :
type Log struct {
	Type       string `json:"log_type"`
	Msg        string `json:"msg"`
	LogTime    string `json:"log_time"`
	FileName   string `json:"file_name"`
	FuncName   string `json:"func_name"`
	LineNumber int    `json:"line_number"`
	Index      int64  `json:"index"`
	Session    int    `json:"session"`
	AppID      string `json:"app_id"`
	AppName    string `json:"app_name"`
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
