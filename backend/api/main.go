package main

import (
	"Server/database"
	"Server/routes"
	"log"
	"os"

	_ "Server/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
)

// @title Golang Fiber with MongoDB for Synoebook
// @version 1.0
// @description This is Swagger docs for rest api golang fiber
// @host localhost:5000
// @BasePath /
// @schemes http
// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the token
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or could not be loaded, using environment variables")
	}
	database.Connect()
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
	}))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome bro!")
	})
	// Serve swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Register API routes
	routes.SetupAuthRoutes(app)
	routes.SetupUserRoutes(app)
	routes.SetupPostRoutes(app)
	routes.SetupChatRoutes(app)
	routes.SetupNotificationRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	if port[0] != ':' {
		port = ":" + port
	}

	log.Fatal(app.Listen(port))
}
