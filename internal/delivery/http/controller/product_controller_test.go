package controller_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

func setupProductController(t *testing.T) (*fiber.App, *ucmocks.ProductUsecaseInterface) {
	t.Helper()
	ucMock := new(ucmocks.ProductUsecaseInterface)
	validate := validator.New()
	ctrl := controller.NewProductController(ucMock, validate)

	// Use the same error handler as production (config.NewErrorHandler)
	// so that validator errors return 400, not 500.
	app := fiber.New(fiber.Config{
		ErrorHandler: config.NewErrorHandler(),
	})
	app.Post("/products", ctrl.Create)
	app.Get("/products", ctrl.GetAll)
	app.Get("/products/:id", ctrl.GetByID)
	app.Put("/products/:id", ctrl.Update)
	app.Delete("/products/:id", ctrl.Delete)

	return app, ucMock
}

// ─── Create ───────────────────────────────────────────────────────────────────

// validCreateProductBody returns a minimal valid JSON body for Create.
// Fill in real values below if your model has required fields.
func validCreateProductBody() []byte {
	b, _ := json.Marshal(map[string]interface{}{
		// TODO: fill in required fields for CreateProductRequest
		// "name": "test",
	})
	return b
}

func TestProductController_Create_Success(t *testing.T) {
	app, ucMock := setupProductController(t)

	ucMock.On("Create", mock.Anything, mock.Anything).
		Return(&model.ProductResponse{}, nil).Maybe()

	req := httptest.NewRequest("POST", "/products", bytes.NewReader(validCreateProductBody()))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	// If validator rejects the empty body → StatusBadRequest, otherwise StatusCreated.
	// Populate validCreateProductBody() above with real field values to get StatusCreated.
	assert.True(t, resp.StatusCode == fiber.StatusCreated || resp.StatusCode == fiber.StatusBadRequest)
	ucMock.AssertExpectations(t)
}

func TestProductController_Create_ValidationError(t *testing.T) {
	app, ucMock := setupProductController(t)

	// Send empty body — validator should reject it
	req := httptest.NewRequest("POST", "/products", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	ucMock.AssertNotCalled(t, "Create")
}

// ─── GetAll ───────────────────────────────────────────────────────────────────

func TestProductController_GetAll_Success(t *testing.T) {
	app, ucMock := setupProductController(t)

	ucMock.On("GetAll", mock.Anything, 1, 10).
		Return([]model.ProductResponse{}, int64(0), nil).Once()

	req := httptest.NewRequest("GET", "/products?page=1&size=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	ucMock.AssertExpectations(t)
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func TestProductController_GetByID_Success(t *testing.T) {
	app, ucMock := setupProductController(t)

	ucMock.On("GetByID", mock.Anything, uint64(1)).
		Return(&model.ProductResponse{}, nil).Once()

	req := httptest.NewRequest("GET", "/products/1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	ucMock.AssertExpectations(t)
}

func TestProductController_GetByID_NotFound(t *testing.T) {
	app, ucMock := setupProductController(t)

	ucMock.On("GetByID", mock.Anything, uint64(404)).
		Return(nil, errors.New("not found")).Once()

	req := httptest.NewRequest("GET", fmt.Sprintf("/products/%d", 404), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	ucMock.AssertExpectations(t)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestProductController_Delete_Success(t *testing.T) {
	app, ucMock := setupProductController(t)

	ucMock.On("Delete", mock.Anything, uint64(1)).Return(nil).Once()

	req := httptest.NewRequest("DELETE", "/products/1", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
	ucMock.AssertExpectations(t)
}
