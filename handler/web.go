package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zerodoctor/go-status/model"
	ppt "github.com/zerodoctor/goprettyprinter"
)

// WebHandler :
type WebHandler struct {
	socket     websocket.Upgrader
	clients    map[string]*websocket.Conn
	programMap map[string]model.App
	dbHandler  *DBHandler

	Broadcast chan model.WebMsg
}

// NewWebHandler :
func NewWebHandler(dbHandler *DBHandler) *WebHandler {
	return &WebHandler{
		socket: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:    make(map[string]*websocket.Conn),
		Broadcast:  make(chan model.WebMsg, 1000),
		programMap: make(map[string]model.App),
		dbHandler:  dbHandler,
	}
}

// WsBroadcast :
func (wh *WebHandler) WsBroadcast() {

	buffMsg := make(map[string][]model.WebMsg)

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

						buffMsg[connID] = []model.WebMsg{}
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

	apps := []model.App{
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

	webMsg := model.WebMsg{}
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
	id := ctx.Query("id")
	name := ctx.Query("name")
	device := ctx.Query("device")
	ip := ctx.Query("ip") // TODO: maybe not do this
	app := wh.dbHandler.AppID(id, name, device, ip)

	// TODO: find -> increment and return/create session number

	wh.programMap[id] = app

	wh.dbHandler.SaveApp(app)

	wh.Broadcast <- model.WebMsg{
		Type: "app",
		Data: app,
	}

	ppt.Infoln("Registered new program:", name)
	ctx.String(http.StatusAccepted, app.ID)
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
		var web model.WebMsg

		err = json.Unmarshal(msg, &web)
		if err != nil {
			ppt.Errorln("failed to Unmarshal msg:\n\t", err.Error())
			continue
		}

		if web.Type == "logs" {
			if len(web.Logs) <= 0 {
				continue
			}

			logs := web.Logs
			id := logs[0].AppID
			if id == "" {
				ppt.Errorln("failed to receive id")
				continue
			}

			app := wh.programMap[id]
			app.Logs = append(app.Logs, logs...)
			wh.programMap[id] = app

			logs = wh.dbHandler.SaveLogs(logs)

			wh.Broadcast <- model.WebMsg{
				Type: "logs",
				Data: logs,
			}
		}

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
