package user

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// SetupRoutes configures all routes for the User module
func SetupRoutes(api fiber.Router, userController *UserController, authMiddleware fiber.Handler) {
	// Swagger
	api.Get("/docs/*", adaptor.HTTPHandlerFunc(httpSwagger.WrapHandler))

	// Auth routes (Public)
	auth := api.Group("/auth")
	auth.Post("/register", userController.Register)
	auth.Post("/login", userController.Login)
	auth.Post("/refresh", userController.RefreshToken)

	// Protected routes
	apiAuth := api.Group("/", authMiddleware)
	apiAuth.Post("/auth/logout", userController.Logout)

	users := apiAuth.Group("/users")
	users.Get("/current", userController.Current)
}
