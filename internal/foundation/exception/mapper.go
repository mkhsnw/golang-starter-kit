package exception

import (
	"github.com/gofiber/fiber/v3"
)

// MapCodeToStatus maps application error codes to HTTP status codes.
func MapCodeToStatus(code string) int {
	switch code {
	case USER_NOT_FOUND, NOT_FOUND:
		return fiber.StatusNotFound
	case EMAIL_ALREADY_EXISTS, INVALID_PASSWORD:
		return fiber.StatusConflict // Or StatusBadRequest depending on semantics
	case TOKEN_EXPIRED, INVALID_TOKEN, UNAUTHORIZED:
		return fiber.StatusUnauthorized
	case FORBIDDEN:
		return fiber.StatusForbidden
	case VALIDATION_ERROR:
		return fiber.StatusBadRequest
	case DATABASE_ERROR, INTERNAL_ERROR:
		return fiber.StatusInternalServerError
	default:
		return fiber.StatusInternalServerError
	}
}
