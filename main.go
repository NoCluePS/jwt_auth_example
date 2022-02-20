package main

import (
	"jwt_example/database"
	"jwt_example/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func intro(c *fiber.Ctx) error {
	c.JSON(fiber.Map{
		"message": "Hello to this auth test API",
	})

	return nil
}

func setupRoutes(app *fiber.App) {
	app.Get("/", intro)
	app.Post("/login", routes.Login)
	app.Post("/register", routes.Register)
	app.Get("/user", routes.GetUser)
} 

func main() {
	database.ConnectDb()
	app := fiber.New();

	app.Use(logger.New())
	setupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}