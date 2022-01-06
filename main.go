package main

import (
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware/cors"
	"github.com/kamil5b/API-Obat/database"
	"github.com/kamil5b/API-Obat/routes"
	"github.com/kamil5b/API-Obat/utils"
)

func main() {
	database.Connect()
	app := fiber.New()
	origin := utils.GoDotEnvVariable("VIEW_URL")
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     origin,
		AllowMethods:     "GET,POST,PUT,DELETE",
	}))

	routes.Setup(app)

	err := app.Listen("192.168.1.18:8000")
	if err != nil {
		fmt.Println(err)
		fmt.Scan()
	}
}
