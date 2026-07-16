package route

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/mkhsnw/golang-starter-kit/internal/delivery/http/controller"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type RouteConfig struct {
	App               *fiber.App
	UserController    *controller.UserController
	ProductController *controller.ProductController
	OrderController   *controller.OrderController
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
	c.setupUserRoutes(apiAuth)
	c.setupProductRoutes(apiAuth)
	c.setupOrderRoutes(apiAuth)
	// @InjectRouteSetup
}

func (c *RouteConfig) setupAuthRoutes(api fiber.Router) {
	authLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
	})

	api.Post("/register", c.UserController.Register, authLimiter)
	api.Post("/login", c.UserController.Login, authLimiter)
}

func (c *RouteConfig) setupUserRoutes(api fiber.Router) {
	users := api.Group("/users")

	users.Get("/current", c.UserController.Current)
}

func (c *RouteConfig) setupProductRoutes(api fiber.Router) {
	products := api.Group("/products")

	products.Post("/", c.ProductController.Create)
	products.Get("/", c.ProductController.GetAll)
	products.Get("/:id", c.ProductController.GetByID)
	products.Put("/:id", c.ProductController.Update)
	products.Delete("/:id", c.ProductController.Delete)
}

func (c *RouteConfig) setupOrderRoutes(api fiber.Router) {
	orders := api.Group("/orders")

	orders.Post("/", c.OrderController.Create)
	orders.Get("/", c.OrderController.GetAll)
	orders.Get("/:id", c.OrderController.GetByID)
	orders.Put("/:id", c.OrderController.Update)
	orders.Delete("/:id", c.OrderController.Delete)
}
