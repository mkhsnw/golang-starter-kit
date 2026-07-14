package route

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/golang-starter-kit/internal/delivery/http/controller"
)

type RouteConfig struct {
	App            *fiber.App
	UserController *controller.UserController
	AuthMiddleware fiber.Handler
}

func (c *RouteConfig) SetupRoutes() {
	c.App.Get("/health", func(ctx fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"status": "ok",
		})
	})

	api := c.App.Group("/api")
	
	// Setup Routes
	c.setupAuthRoutes(api)
	c.setupUserRoutes(api)
}

func (c *RouteConfig) setupAuthRoutes(api fiber.Router) {
	api.Post("/register", c.UserController.Register)
	api.Post("/login", c.UserController.Login)
}

func (c *RouteConfig) setupUserRoutes(api fiber.Router) {
	users := api.Group("/users")
	users.Use(c.AuthMiddleware)
	
	users.Get("/current", c.UserController.Current)
}
