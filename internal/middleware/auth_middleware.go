package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	appcontext "github.com/mkhsnw/golang-starter-kit/internal/foundation/appcontext"
	"github.com/mkhsnw/golang-starter-kit/internal/foundation/exception"
)

func NewAuthMiddleware(jwtSecret string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			return exception.New(exception.UNAUTHORIZED, "Missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return exception.New(exception.UNAUTHORIZED, "Invalid authorization header format")
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, exception.New(exception.UNAUTHORIZED, "Invalid token signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return exception.New(exception.UNAUTHORIZED, "Invalid or expired token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return exception.New(exception.UNAUTHORIZED, "Invalid token claims")
		}

		// Extract user ID string
		idStr, ok := claims["id"].(string)
		if !ok {
			return exception.New(exception.UNAUTHORIZED, "Invalid token claims: id must be a string")
		}

		ctx.Locals(appcontext.ContextKeyUserID, idStr)
		ctx.Locals(appcontext.ContextKeyUserEmail, claims["email"])

		if reqCtx, ok := ctx.Locals("requestContext").(*appcontext.RequestContext); ok {
			reqCtx.UserID = idStr
		}

		return ctx.Next()
	}
}
