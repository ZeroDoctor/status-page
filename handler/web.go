package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ppt "github.com/zerodoctor/goprettyprinter"
)

// Log :
type Log struct {
	Type       string `json:"type"`
	Msg        string `json:"msg"`
	LogTime    string `json:"log_time"`
	FileName   string `json:"file_name"`
	FuncName   string `json:"func_name"`
	LineNumber int    `json:"line_number"`
	Index      int    `json:"index"`
	AppID      string `json:"app_id"`
	AppName    string `json:"app_name"`
}

// App :
type App struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Logs []Log  `json:"-"` // make db later
}

// WebMsg :
type WebMsg struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WebHandler :
type WebHandler struct {
	socket     websocket.Upgrader
	clients    map[string]*websocket.Conn
	programMap map[string]App

	Broadcast chan WebMsg
}

// NewWebHandler :
func NewWebHandler() *WebHandler {
	return &WebHandler{
		socket: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:    make(map[string]*websocket.Conn),
		Broadcast:  make(chan WebMsg, 1000),
		programMap: make(map[string]App),
	}
}

// WsBroadcast :
func (wh *WebHandler) WsBroadcast() {

	buffMsg := make(map[string][]WebMsg)

	for msg := range wh.Broadcast {

		for connID, conn := range wh.clients {

			if _, ok := wh.clients[connID]; ok {
				err := conn.WriteJSON(msg)
				if err != nil {
					ppt.Warnf("Failed to write to client: %s\n", err)

					fmt.Printf("Removing client %s\n", connID)
					delete(wh.clients, connID)
				} else {
					if len(buffMsg[connID]) > 0 {
						for _, m := range buffMsg[connID] {
							conn.WriteJSON(m)
						}

						buffMsg[connID] = []WebMsg{}
					}
				}
			} else {
				buffMsg[connID] = append(buffMsg[connID], msg)
			}
		}
	}
}

// RenderIndex :
func (wh *WebHandler) RenderIndex(ctx *gin.Context) {

	apps := []App{
		{
			ID:   "zxcvzcv",
			Name: "Hello There",
		},
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"Title": "Hello, World!",
		"Apps":  apps,
	})
}

// Websocket :
func (wh *WebHandler) Websocket(ctx *gin.Context) {

	conn, err := wh.socket.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println("Upgrade Error:", err)
		return
	}

	_, p, err := conn.ReadMessage()
	if err != nil {
		ppt.Errorln("failed to read message:\n\t", err.Error())
		return
	}

	webMsg := WebMsg{}
	err = json.Unmarshal(p, &webMsg)
	if err != nil {
		ppt.Errorln("failed to Unmarshal:\n\t", err.Error())
		return
	}

	ppt.Infoln("Found connection:", webMsg.Data.(string))
	data := webMsg.Data.(string)
	wh.clients[data] = conn

	if webMsg.Type == "web" {

	}

	if webMsg.Type == "client" {
		go wh.ReadMessage(conn)
	}

}

// NewApp :
func (wh *WebHandler) NewApp(ctx *gin.Context) {
	name := ctx.Query("name")
	id := RandString(10)

	program := App{
		ID:   id,
		Name: name,
	}
	wh.programMap[id] = program

	wh.Broadcast <- WebMsg{
		Type: "app",
		Data: program,
	}

	ppt.Infoln("Registered new program:", name)
	ctx.String(http.StatusAccepted, id)
}

// ReadMessage :
func (wh *WebHandler) ReadMessage(conn *websocket.Conn) {
	defer conn.Close()

	for {

		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// * assume its a log for now
		var log Log

		err = json.Unmarshal(msg, &log)
		if err != nil {
			ppt.Errorln("failed to Unmarshal msg:\n\t", err.Error())
			continue
		}

		id := log.AppID

		app := wh.programMap[id]
		log.Index = len(app.Logs)
		app.Logs = append(app.Logs, log)
		wh.programMap[id] = app

		wh.Broadcast <- WebMsg{
			Type: "log",
			Data: log,
		}
	}
}

// NewLog :
func (wh *WebHandler) NewLog(ctx *gin.Context) {
	id := ctx.Query("id")

	var log Log
	err := ctx.BindJSON(&log)
	if err != nil {
		ppt.Errorln("failed to parse json to log:\n\t", err.Error())
		return
	}

	app := wh.programMap[id]
	log.Index = len(app.Logs)
	app.Logs = append(app.Logs, log)
	wh.programMap[id] = app

	wh.Broadcast <- WebMsg{
		Type: "log",
		Data: log,
	}
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~"

// RandString :
func RandString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
