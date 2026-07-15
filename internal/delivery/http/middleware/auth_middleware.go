package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mkhsnw/golang-starter-kit/internal/exception"
	"github.com/mkhsnw/golang-starter-kit/internal/util"
)

func NewAuthMiddleware(jwtSecret string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			return exception.Unauthorized("Missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return exception.Unauthorized("Invalid authorization header format")
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, exception.Unauthorized("Invalid token signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return exception.Unauthorized("Invalid or expired token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return exception.Unauthorized("Invalid token claims")
		}

		// Inject user info into fiber Context
		ctx.Locals(util.ContextKeyUserID, claims["id"])
		ctx.Locals(util.ContextKeyUserEmail, claims["email"])

		return ctx.Next()
	}
}
