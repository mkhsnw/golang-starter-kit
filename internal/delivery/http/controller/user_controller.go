package controller

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/golang-starter-kit/internal/model"
	"github.com/mkhsnw/golang-starter-kit/internal/usecase"
	"github.com/mkhsnw/golang-starter-kit/internal/util"
)

type UserController struct {
	UserUsecase usecase.UserUsecaseInterface
	Validator   *validator.Validate
}

func NewUserController(userUsecase usecase.UserUsecaseInterface, validate *validator.Validate) *UserController {
	return &UserController{
		UserUsecase: userUsecase,
		Validator:   validate,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email, and password
// @Tags users
// @Accept json
// @Produce json
// @Param body body model.RegisterRequest true "Register Request"
// @Success 201 {object} model.WebResponse[model.UserResponse]
// @Failure 400 {object} model.WebResponse[any]
// @Router /register [post]
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

// Login godoc
// @Summary Login user
// @Description Authenticate user and get JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Login Request"
// @Success 200 {object} model.WebResponse[model.TokenResponse]
// @Failure 401 {object} model.WebResponse[any]
// @Router /login [post]
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

// Current godoc
// @Summary Get current user
// @Description Get current authenticated user details
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.WebResponse[model.UserResponse]
// @Failure 401 {object} model.WebResponse[any]
// @Router /users/current [get]
func (c *UserController) Current(ctx fiber.Ctx) error {
	userId, ok := util.GetUserID(ctx)
	if !ok {
		return fiber.ErrUnauthorized
	}

	response, err := c.UserUsecase.GetCurrentUser(ctx.Context(), userId)
	if err != nil {
		return err
	}

	return ctx.JSON(model.WebResponse[*model.UserResponse]{
		Data: response,
	})
}
