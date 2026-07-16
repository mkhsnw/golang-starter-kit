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

		// Extract user ID string
		idStr, ok := claims["id"].(string)
		if !ok {
			return exception.Unauthorized("Invalid token claims: id must be a string")
		}

		ctx.Locals(util.ContextKeyUserID, idStr)
		ctx.Locals(util.ContextKeyUserEmail, claims["email"])

		return ctx.Next()
	}
}
