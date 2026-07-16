package controller

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/golang-starter-kit/internal/exception"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/usecase"
	"github.com/mkhsnw/golang-starter-kit/internal/util"
)

type OrderController struct {
	OrderUsecase usecase.OrderUsecaseInterface
	Validator    *validator.Validate
}

func NewOrderController(uc usecase.OrderUsecaseInterface, validate *validator.Validate) *OrderController {
	return &OrderController{OrderUsecase: uc, Validator: validate}
}

// Create godoc
// @Summary Create Order
// @Description Create a new Order
// @Tags orders
// @Accept json
// @Produce json
// @Param body body model.CreateOrderRequest true "Create Request"
// @Security BearerAuth
// @Success 201 {object} model.WebResponse[model.OrderResponse]
// @Failure 400 {object} model.WebResponse[any]
// @Router /orders [post]
func (c *OrderController) Create(ctx fiber.Ctx) error {
	var req model.CreateOrderRequest
	if err := util.ParseAndValidate(ctx, c.Validator, &req); err != nil {
		return err
	}

	response, err := c.OrderUsecase.Create(ctx.Context(), &req)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.OrderResponse]{Data: response})
}

// GetByID godoc
// @Summary Get Order by ID
// @Description Get a Order by its ID
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Security BearerAuth
// @Success 200 {object} model.WebResponse[model.OrderResponse]
// @Failure 404 {object} model.WebResponse[any]
// @Router /orders/{id} [get]
func (c *OrderController) GetByID(ctx fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return exception.NewResponseError(400, "BAD_REQUEST", "invalid id format")
	}

	response, err := c.OrderUsecase.GetByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.OrderResponse]{Data: response})
}

// GetAll godoc
// @Summary Get all orders
// @Description Get a paginated list of orders
// @Tags orders
// @Produce json
// @Param cursor query string false "Cursor ID" default("")
// @Param size query int false "Page size" default(10)
// @Security BearerAuth
// @Success 200 {object} model.WebResponse[[]model.OrderResponse]
// @Router /orders [get]
func (c *OrderController) GetAll(ctx fiber.Ctx) error {
	cursor := ctx.Query("cursor", "")
	size, _ := strconv.Atoi(ctx.Query("size", "10"))

	// Guard: Cegah abuse pagination
	if size < 1 || size > 100 {
		size = 10
	}

	responses, nextCursor, err := c.OrderUsecase.GetAll(ctx.Context(), cursor, size)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.OrderResponse]{
		Data: responses,
		Paging: &model.PageMetadata{
			Size:       size,
			NextCursor: nextCursor,
		},
	})
}

// Update godoc
// @Summary Update Order
// @Description Update an existing Order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param body body model.UpdateOrderRequest true "Update Request"
// @Security BearerAuth
// @Success 200 {object} model.WebResponse[model.OrderResponse]
// @Failure 400 {object} model.WebResponse[any]
// @Failure 404 {object} model.WebResponse[any]
// @Router /orders/{id} [put]
func (c *OrderController) Update(ctx fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return exception.NewResponseError(400, "BAD_REQUEST", "invalid id format")
	}

	var req model.UpdateOrderRequest
	if err := util.ParseAndValidate(ctx, c.Validator, &req); err != nil {
		return err
	}

	response, err := c.OrderUsecase.Update(ctx.Context(), id, &req)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.OrderResponse]{Data: response})
}

// Delete godoc
// @Summary Delete Order
// @Description Delete a Order by its ID
// @Tags orders
// @Param id path string true "Order ID"
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 404 {object} model.WebResponse[any]
// @Router /orders/{id} [delete]
func (c *OrderController) Delete(ctx fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return exception.NewResponseError(400, "BAD_REQUEST", "invalid id format")
	}

	if err := c.OrderUsecase.Delete(ctx.Context(), id); err != nil {
		return err
	}
	return ctx.Status(fiber.StatusNoContent).Send(nil)
}
