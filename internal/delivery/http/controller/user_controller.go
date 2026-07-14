package controller

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/usecase"
	"github.com/mkhsnw/golang-starter-kit/internal/util"
)

type UserController struct {
	UserUsecase *usecase.UserUsecase
	Validator   *validator.Validate
}

func NewUserController(userUsecase *usecase.UserUsecase, validate *validator.Validate) *UserController {
	return &UserController{
		UserUsecase: userUsecase,
		Validator:   validate,
	}
}

func (c *UserController) Register(ctx fiber.Ctx) error {
	var req model.RegisterRequest
	if err := util.ParseAndValidate(ctx, c.Validator, &req); err != nil {
		return err
	}

	response, err := c.UserUsecase.Register(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.UserResponse]{
		Data: response,
	})
}

func (c *UserController) Login(ctx fiber.Ctx) error {
	var req model.LoginRequest
	if err := util.ParseAndValidate(ctx, c.Validator, &req); err != nil {
		return err
	}

	response, err := c.UserUsecase.Login(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.TokenResponse]{
		Data: response,
	})
}

func (c *UserController) Current(ctx fiber.Ctx) error {
	// We extract ID from fiber Locals which is injected by AuthMiddleware
	userId := ctx.Locals("userId")
	
	return ctx.JSON(model.WebResponse[fiber.Map]{
		Data: fiber.Map{
			"id": userId,
			"message": "This is a protected route. Hello user!",
		},
	})
}
