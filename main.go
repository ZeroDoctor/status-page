package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/zerodoctor/go-status/handler"
	ppt "github.com/zerodoctor/goprettyprinter"
)

var env = ""

func init() {
	ppt.Init()
	ppt.LoggerFlags = ppt.FILE | ppt.LINE
	ppt.Infoln("Starting Server...")

	env := os.Getenv("env")
	envFile := "./" + env + ".env"
	if env == "" {
		env = "dev"
		envFile = env + ".env"
	}

	err := godotenv.Load(envFile)
	if err != nil {
		ppt.Errorln(err.Error())
		panic("Failed to load env file")
	}
}

func main() {
	dbh := handler.NewDBHandler()
	wh := handler.NewWebHandler(dbh)

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
	// TODO: create init route

	go wh.WsBroadcast()
	go SendFake(wh)

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
