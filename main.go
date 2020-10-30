package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	ppt "github.com/zerodoctor/goprettyprinter"
)

func main() {

	ppt.Init()
	ppt.Infoln("Starting Server...")

	engine := html.New("./assets/html", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Render("index", fiber.Map{
			"Title": "Hello, World!",
		})
	})

	app.Listen(":3000")

}
