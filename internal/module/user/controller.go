package user

import (
	playvalidator "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/rel/internal/foundation/appcontext"
	"github.com/mkhsnw/rel/internal/foundation/response"
	"github.com/mkhsnw/rel/internal/foundation/validator"
	"github.com/mkhsnw/rel/internal/module/user/dto"
)

type UserController struct {
	UserService *UserService
	Validator   *playvalidator.Validate
}

func NewUserController(userService *UserService, validate *playvalidator.Validate) *UserController {
	return &UserController{
		UserService: userService,
		Validator:   validate,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email, and password
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "Register Request"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /auth/register [post]
func (c *UserController) Register(ctx fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := validator.ParseAndValidate(ctx, c.Validator, &req); err != nil {
		return err
	}

	res, err := c.UserService.Register(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return response.Created(res).Send(ctx)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and get JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "Login Request"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (c *UserController) Login(ctx fiber.Ctx) error {
	var req dto.LoginRequest
	if err := validator.ParseAndValidate(ctx, c.Validator, &req); err != nil {
		return err
	}

	res, err := c.UserService.Login(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return response.OK(res).Send(ctx)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags users
// @Accept json
// @Produce json
// @Param body body map[string]string true "Refresh Token Request (requires refresh_token field)"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/refresh [post]
func (c *UserController) RefreshToken(ctx fiber.Ctx) error {
	var req map[string]string
	if err := ctx.Bind().JSON(&req); err != nil {
		return fiber.ErrBadRequest
	}

	refreshToken, ok := req["refresh_token"]
	if !ok || refreshToken == "" {
		return fiber.ErrBadRequest
	}

	res, err := c.UserService.RefreshToken(ctx.Context(), refreshToken)
	if err != nil {
		return err
	}

	return response.OK(res).Send(ctx)
}

// Logout godoc
// @Summary Logout user
// @Description Revoke all refresh tokens for the current user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/logout [post]
func (c *UserController) Logout(ctx fiber.Ctx) error {
	userId, ok := appcontext.GetFiberUserID(ctx)
	if !ok {
		return fiber.ErrUnauthorized
	}

	if err := c.UserService.Logout(ctx.Context(), userId); err != nil {
		return err
	}

	return response.OK("Logged out successfully").Send(ctx)
}

// Current godoc
// @Summary Get current user
// @Description Get current authenticated user details
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /users/current [get]
func (c *UserController) Current(ctx fiber.Ctx) error {
	userId, ok := appcontext.GetFiberUserID(ctx)
	if !ok {
		return fiber.ErrUnauthorized
	}

	res, err := c.UserService.GetCurrentUser(ctx.Context(), userId)
	if err != nil {
		return err
	}

	return response.OK(res).Send(ctx)
}
