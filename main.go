package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fasthttp/websocket"
	"github.com/gin-gonic/gin"
	"github.com/zerodoctor/go-status/handler"
	ppt "github.com/zerodoctor/goprettyprinter"
)

func main() {

	ppt.Init()
	ppt.Infoln("Starting Server...")

	wh := &handler.WebHandler{
		socket: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:   make(map[string]*websocket.Conn),
		Broadcast: make(chan WebMsg, 1000),
	}

	router := gin.Default()

	router.Static("/css", "./assets/css/")
	router.Static("/js", "./assets/js/")
	router.Static("/img", "./assets/img/")

	html := template.Must(loadTemplates("./assets/html"))
	router.SetHTMLTemplate(html)

	router.GET("/", wh.RenderIndex)
	router.GET("/ws", wh.Websocket)
	router.POST("/new/program", wh.NewProgram)
	router.POST("/new/log", wh.NewLog)

	go wh.WsBroadcast()

	go wh.SendFake()

	router.Run(":3000")
}

func loadTemplates(path string) (*template.Template, error) {

	var files []string
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if strings.Contains(path, ".html") {
				files = append(files, path)
				fmt.Println("found file:", path, info.Size())
			}
			return nil
		})
	if err != nil {
		ppt.Errorln(err)
	}

	return template.ParseFiles(files...)
}
