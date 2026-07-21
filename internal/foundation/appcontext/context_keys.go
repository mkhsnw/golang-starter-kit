package appcontext

import "github.com/gofiber/fiber/v3"

type ContextKey string

const (
	ContextKeyUserID    ContextKey = "userId"
	ContextKeyUserEmail ContextKey = "userEmail"
	ContextKeyRequestID ContextKey = "requestid"
)

// GetFiberUserID extracts the user ID as string (UUID) from the fiber context
func GetFiberUserID(ctx fiber.Ctx) (string, bool) {
	val := ctx.Locals(ContextKeyUserID)
	if val == nil {
		return "", false
	}
	if id, ok := val.(string); ok {
		return id, true
	}
	return "", false
}
