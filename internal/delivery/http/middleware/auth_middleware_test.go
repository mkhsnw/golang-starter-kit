package middleware_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mkhsnw/golang-starter-kit/internal/config"
	"github.com/mkhsnw/golang-starter-kit/internal/delivery/http/middleware"
	"github.com/stretchr/testify/assert"
)

func setupTestApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: config.NewErrorHandler(),
	})
	app.Use(middleware.NewAuthMiddleware("secret_key"))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("success")
	})
	return app
}

func generateValidToken() string {
	claims := jwt.MapClaims{
		"id":    "123",
		"email": "test@example.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte("secret_key"))
	return t
}

func generateInvalidTypeToken() string {
	claims := jwt.MapClaims{
		"id":    true, // neither float64 nor string
		"email": "test@example.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte("secret_key"))
	return t
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	app := setupTestApp()
	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	app := setupTestApp()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+generateValidToken())
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	app := setupTestApp()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token_xyz")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthMiddleware_InvalidTypeToken(t *testing.T) {
	app := setupTestApp()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+generateInvalidTypeToken())
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
