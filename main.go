package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/valyala/fasthttp"
	ppt "github.com/zerodoctor/goprettyprinter"
)

// TODO: use gin instead

// Program :
type Program struct {
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
	socket    websocket.FastHTTPUpgrader
	clients   map[string]*websocket.Conn
	Broadcast chan WebMsg
}

func main() {

	ppt.Init()
	ppt.Infoln("Starting Server...")

	engine := html.New("./assets/html", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	wh := &WebHandler{
		socket: websocket.FastHTTPUpgrader{
			CheckOrigin: func(r *fasthttp.RequestCtx) bool {
				return true
			},
		},
		clients:   make(map[string]*websocket.Conn),
		Broadcast: make(chan WebMsg, 1000),
	}

	app.Static("/css", "./assets/css/")
	app.Static("/js", "./assets/js/")
	app.Static("/img", "./assets/img/")

	app.Get("/", wh.RenderIndex)
	app.Get("/ws", wh.Websocket)

	go func() {
		time.Sleep(time.Second * 3)
		fmt.Println("Sending")
		wh.Broadcast <- WebMsg{
			Type: "ping",
			Data: "Hello there",
		}
	}()

	go wh.WsBroadcast()

	app.Listen(":3000")
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
func (wh *WebHandler) RenderIndex(ctx *fiber.Ctx) error {

	programs := []Program{
		{
			ID:   "asdfasdf",
			Name: "Test-1",
		},
		{
			ID:   "qwerqwer",
			Name: "Test-2",
		},
	}

	return ctx.Render("index", fiber.Map{
		"Title":    "Hello, World!",
		"Programs": programs,
	})
}

// Websocket :
func (wh *WebHandler) Websocket(ctx *fiber.Ctx) error {
	err := wh.socket.Upgrade(ctx.Context(), func(conn *websocket.Conn) {
		defer conn.Close()

		for {
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
			time.Sleep(time.Second * 1)
		}
	})
	if err != nil {
		ppt.Errorln("failed to create upgrade from context:\n\t", err.Error())
		return err
	}

	return nil
}
