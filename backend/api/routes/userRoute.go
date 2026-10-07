package routes

import (
	"Server/controllers"
	"Server/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(app *fiber.App) {
	app.Get("/user/getUser/:id", controllers.GetUserByID)
	app.Patch("/user/update/:id", middleware.AuthMiddleware, controllers.UpdateUser)
	app.Patch("/user/:id/following", middleware.AuthMiddleware, controllers.FollowingUser)
	app.Get("/user/getSug", middleware.AuthMiddleware, controllers.GetSugUser)
	app.Delete("/user/delete/:id", middleware.AuthMiddleware, controllers.DeleteUser)
}
