package controller_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/golang-starter-kit/internal/config"
	"github.com/mkhsnw/golang-starter-kit/internal/delivery/http/controller"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	ucmocks "github.com/mkhsnw/golang-starter-kit/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Helper ──────────────────────────────────────────────────────────────────

func setupOrderController(t *testing.T) (*fiber.App, *ucmocks.OrderUsecaseInterface) {
	t.Helper()
	ucMock := new(ucmocks.OrderUsecaseInterface)
	validate := validator.New()
	ctrl := controller.NewOrderController(ucMock, validate)

	// Use the same error handler as production (config.NewErrorHandler)
	// so that validator errors return 400, not 500.
	app := fiber.New(fiber.Config{
		ErrorHandler: config.NewErrorHandler(),
	})
	app.Post("/orders", ctrl.Create)
	app.Get("/orders", ctrl.GetAll)
	app.Get("/orders/:id", ctrl.GetByID)
	app.Put("/orders/:id", ctrl.Update)
	app.Delete("/orders/:id", ctrl.Delete)

	return app, ucMock
}

// ─── Create ───────────────────────────────────────────────────────────────────

// validCreateOrderBody returns a minimal valid JSON body for Create.
func validCreateOrderBody() []byte {
	b, _ := json.Marshal(map[string]interface{}{
		"user_id":    "test",
		"product_id": "test",
		"amount":     1,
		"total":      1,
	})
	return b
}

func TestOrderController_Create_Success(t *testing.T) {
	app, ucMock := setupOrderController(t)

	ucMock.On("Create", mock.Anything, mock.Anything).
		Return(&model.OrderResponse{}, nil).Once()

	req := httptest.NewRequest("POST", "/orders", bytes.NewReader(validCreateOrderBody()))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode, "Response status is not 201 Created. Please ensure validCreateOrderBody() returns a valid JSON payload that satisfies the validation rules.")
	ucMock.AssertExpectations(t)
}
func TestOrderController_Create_ValidationError(t *testing.T) {
	app, ucMock := setupOrderController(t)

	// Send empty body — validator should reject it
	req := httptest.NewRequest("POST", "/orders", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	ucMock.AssertNotCalled(t, "Create")
}

// ─── GetAll ───────────────────────────────────────────────────────────────────

func TestOrderController_GetAll_Success(t *testing.T) {
	app, ucMock := setupOrderController(t)

	ucMock.On("GetAll", mock.Anything, "", 10).
		Return([]model.OrderResponse{}, (*string)(nil), nil).Once()

	req := httptest.NewRequest("GET", "/orders?cursor=&size=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	ucMock.AssertExpectations(t)
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func TestOrderController_GetByID_Success(t *testing.T) {
	app, ucMock := setupOrderController(t)

	ucMock.On("GetByID", mock.Anything, "uuid-1").
		Return(&model.OrderResponse{}, nil).Once()

	req := httptest.NewRequest("GET", "/orders/uuid-1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	ucMock.AssertExpectations(t)
}

func TestOrderController_GetByID_NotFound(t *testing.T) {
	app, ucMock := setupOrderController(t)

	ucMock.On("GetByID", mock.Anything, "uuid-404").
		Return(nil, errors.New("not found")).Once()

	req := httptest.NewRequest("GET", "/orders/uuid-404", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	ucMock.AssertExpectations(t)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestOrderController_Delete_Success(t *testing.T) {
	app, ucMock := setupOrderController(t)

	ucMock.On("Delete", mock.Anything, "uuid-1").Return(nil).Once()

	req := httptest.NewRequest("DELETE", "/orders/uuid-1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	ucMock.AssertExpectations(t)
}
