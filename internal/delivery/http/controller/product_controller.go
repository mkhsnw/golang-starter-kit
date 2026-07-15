package controller

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/usecase"
	"github.com/mkhsnw/golang-starter-kit/internal/util"
)

type ProductController struct {
	ProductUsecase usecase.ProductUsecaseInterface
	Validator      *validator.Validate
}

func NewProductController(uc usecase.ProductUsecaseInterface, validate *validator.Validate) *ProductController {
	return &ProductController{ProductUsecase: uc, Validator: validate}
}

// Create godoc
// @Summary Create Product
// @Description Create a new Product
// @Tags products
// @Accept json
// @Produce json
// @Param body body model.CreateProductRequest true "Create Request"
// @Security BearerAuth
// @Success 201 {object} model.WebResponse[model.ProductResponse]
// @Failure 400 {object} model.WebResponse[any]
// @Router /products [post]
func (c *ProductController) Create(ctx fiber.Ctx) error {
	var req model.CreateProductRequest
	if err := util.ParseAndValidate(ctx, c.Validator, &req); err != nil {
		return err
	}

	response, err := c.ProductUsecase.Create(ctx.Context(), &req)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.ProductResponse]{Data: response})
}

// GetByID godoc
// @Summary Get Product by ID
// @Description Get a Product by its ID
// @Tags products
// @Produce json
// @Param id path int true "Product ID"
// @Security BearerAuth
// @Success 200 {object} model.WebResponse[model.ProductResponse]
// @Failure 404 {object} model.WebResponse[any]
// @Router /products/{id} [get]
func (c *ProductController) GetByID(ctx fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return err
	}

	response, err := c.ProductUsecase.GetByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.ProductResponse]{Data: response})
}

// GetAll godoc
// @Summary Get all products
// @Description Get a paginated list of products
// @Tags products
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Security BearerAuth
// @Success 200 {object} model.WebResponse[[]model.ProductResponse]
// @Router /products [get]
func (c *ProductController) GetAll(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))

	// Guard: Cegah abuse pagination
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	responses, total, err := c.ProductUsecase.GetAll(ctx.Context(), page, size)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[[]model.ProductResponse]{
		Data: responses,
		Paging: &model.PageMetadata{
			Page:      page,
			Size:      size,
			TotalItem: total,
			TotalPage: (total + int64(size) - 1) / int64(size),
		},
	})
}

// Update godoc
// @Summary Update Product
// @Description Update an existing Product
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param body body model.UpdateProductRequest true "Update Request"
// @Security BearerAuth
// @Success 200 {object} model.WebResponse[model.ProductResponse]
// @Failure 400 {object} model.WebResponse[any]
// @Failure 404 {object} model.WebResponse[any]
// @Router /products/{id} [put]
func (c *ProductController) Update(ctx fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return err
	}

	var req model.UpdateProductRequest
	if err := util.ParseAndValidate(ctx, c.Validator, &req); err != nil {
		return err
	}

	response, err := c.ProductUsecase.Update(ctx.Context(), id, &req)
	if err != nil {
		return err
	}
	return ctx.JSON(model.WebResponse[*model.ProductResponse]{Data: response})
}

// Delete godoc
// @Summary Delete Product
// @Description Delete a Product by its ID
// @Tags products
// @Param id path int true "Product ID"
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 404 {object} model.WebResponse[any]
// @Router /products/{id} [delete]
func (c *ProductController) Delete(ctx fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return err
	}

	if err := c.ProductUsecase.Delete(ctx.Context(), id); err != nil {
		return err
	}
	return ctx.Status(fiber.StatusNoContent).Send(nil)
}
