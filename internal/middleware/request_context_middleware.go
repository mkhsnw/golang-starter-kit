package middleware

import (
	stdcontext "context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/mkhsnw/golang-starter-kit/internal/foundation/appcontext" // package appcontext
)

// RequestContextMiddleware initializes the RequestContext and injects it into Fiber's UserContext.
func RequestContextMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		reqID := c.Get("X-Request-ID")
		if reqID == "" {
			id, _ := uuid.NewV7()
			reqID = id.String()
		}

		traceID := c.Get("X-Trace-ID")
		if traceID == "" {
			id, _ := uuid.NewV7()
			traceID = id.String()
		}

		reqCtx := &appcontext.RequestContext{
			RequestID: reqID,
			TraceID:   traceID,
			IPAddress: c.IP(),
			UserAgent: c.Get("User-Agent"),
		}

		c.Locals("requestContext", reqCtx)

		return c.Next()
	}
}

// GetContext is a helper for controllers to extract the standard context with RequestContext injected.
func GetContext(c fiber.Ctx) stdcontext.Context {
	ctx := c.Context()
	if reqCtx, ok := c.Locals("requestContext").(*appcontext.RequestContext); ok {
		return appcontext.WithRequestContext(ctx, reqCtx)
	}
	return ctx
}
