package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lukerops/pluxy/cmd/handlers"
)

func main() {
	app := fiber.New()

	app.Get("/", handlers.Index)

	app.Listen(":8080")
}
