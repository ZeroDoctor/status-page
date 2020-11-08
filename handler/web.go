package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ppt "github.com/zerodoctor/goprettyprinter"
)

// App :
type App struct {
	ID   string
	Name string
}

// WebMsg :
type WebMsg struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WebHandler :
type WebHandler struct {
	socket    websocket.Upgrader
	clients   map[string]*websocket.Conn
	Broadcast chan WebMsg

	programMap map[string]App
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
					if strings.Contains(err.Error(), "broken pipe") ||
						strings.Contains(err.Error(), "connection timed out") ||
						strings.Contains(err.Error(), "no route to host") {

						fmt.Printf("Removing client %s\n", connID)
						delete(wh.clients, connID)
					}
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
			ID:   "asdfasdf",
			Name: "Test-1",
		},
		{
			ID:   "qwerqwer",
			Name: "Test-2",
		},
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

	if webMsg.Type == "init" {
		ppt.Infoln("Found connection:", webMsg.Data.(string))
		data := webMsg.Data.(string)
		wh.clients[data] = conn
	}
}

func (wh *WebHandler) NewProgram(ctx *gin.Context) {
	name := ctx.Query("name")
	id := RandString(10)

	program := App{
		ID:   id,
		Name: name,
	}
	wh.programMap[id] = program

	ppt.Infoln("Registered new program:", name)
}

func (wh *WebHandler) NewLog(ctx *gin.Context) {
	id := ctx.Query("id")

	// TODO: parse json
	// TODO: send log to socket

	ppt.Infoln("New Log from:", id)
}

func RandString(length int) string {

	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" +
		"<>?*&^%$#@!(){}[];:,.~`"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
