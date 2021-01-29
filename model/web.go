package model

// WebMsg :
type WebMsg struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	Logs []Log       `json:"logs"`
}
