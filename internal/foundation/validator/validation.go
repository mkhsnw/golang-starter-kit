package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func ParseAndValidate(ctx fiber.Ctx, validate *validator.Validate, request any) error {
	if err := ctx.Bind().Body(request); err != nil {
		return err
	}
	if err := validate.Struct(request); err != nil {
		return err
	}
	return nil
}
