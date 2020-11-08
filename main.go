package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zerodoctor/go-status/handler"
	ppt "github.com/zerodoctor/goprettyprinter"
)

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

func main() {

	ppt.Init()
	ppt.Infoln("Starting Server...")

	wh := handler.NewWebHandler()

	router := gin.Default()

	html := template.Must(loadTemplates("./assets/html"))
	router.SetHTMLTemplate(html)

	router.Static("/css", "./assets/css/")
	router.Static("/js", "./assets/js/")
	router.Static("/img", "./assets/img/")
	router.StaticFile("/favicon.ico", "./assets/favicon.ico")

	router.GET("/", wh.RenderIndex)
	router.GET("/ws", wh.Websocket)
	router.POST("/new/app", wh.NewApp)
	router.POST("/new/log", wh.NewLog)

	go wh.WsBroadcast()
	go SendFake(wh)

	router.Run(":3000")
}
