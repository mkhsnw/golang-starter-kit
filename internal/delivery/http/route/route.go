package route

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/mkhsnw/golang-starter-kit/internal/delivery/http/controller"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type RouteConfig struct {
	App            *fiber.App
	UserController *controller.UserController
	// @InjectRouteStruct
	AuthMiddleware fiber.Handler
}

func (c *RouteConfig) SetupRoutes() {
	api := c.App.Group("/api/v1")

	// Setup Public Routes
	c.setupAuthRoutes(api)
	api.Get("/docs/*", adaptor.HTTPHandlerFunc(httpSwagger.WrapHandler))

	// Setup Protected Routes
	apiAuth := api.Group("/", c.AuthMiddleware)
	apiAuth.Post("/auth/logout", c.UserController.Logout) // Map logout here explicitly

	c.setupUserRoutes(apiAuth)
	// @InjectRouteSetup
}

func (c *RouteConfig) setupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")
	auth.Post("/register", c.UserController.Register)
	auth.Post("/login", c.UserController.Login)
	auth.Post("/refresh", c.UserController.RefreshToken)

	// Logout uses the parent apiAuth which has AuthMiddleware applied
	// It's mapped to /auth/logout
}

func (c *RouteConfig) setupUserRoutes(api fiber.Router) {
	users := api.Group("/users")

	users.Get("/current", c.UserController.Current)
}
