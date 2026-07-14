package route

import "github.com/gofiber/fiber/v3"

type RouteConfig struct {
	App *fiber.App
}

func (c *RouteConfig) SetupRoutes() {
	c.App.Get("/health", func(ctx fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"status": "ok",
		})
	})
}
