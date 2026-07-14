package exception

import "github.com/gofiber/fiber/v3"

type ResponseError struct {
	Code    int
	Message string
}

func (e *ResponseError) Error() string {
	return e.Message
}

func NewResponseError(code int, message string) *ResponseError {
	return &ResponseError{
		Code:    code,
		Message: message,
	}
}

func BadRequest(message string) *ResponseError {
	return &ResponseError{Code: fiber.StatusBadRequest, Message: message}
}

func Unauthorized(message string) *ResponseError {
	return &ResponseError{Code: fiber.StatusUnauthorized, Message: message}
}

func Forbidden(message string) *ResponseError {
	return &ResponseError{Code: fiber.StatusForbidden, Message: message}
}

func NotFound(message string) *ResponseError {
	return &ResponseError{Code: fiber.StatusNotFound, Message: message}
}

func Conflict(message string) *ResponseError {
	return &ResponseError{Code: fiber.StatusConflict, Message: message}
}
