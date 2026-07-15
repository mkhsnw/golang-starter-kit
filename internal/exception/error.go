package exception

import "github.com/gofiber/fiber/v3"

type ResponseError struct {
	Code    int
	AppCode string
	Message string
}

func (e *ResponseError) Error() string {
	return e.Message
}

func NewResponseError(code int, appCode string, message string) *ResponseError {
	return &ResponseError{
		Code:    code,
		AppCode: appCode,
		Message: message,
	}
}

func BadRequest(message string) *ResponseError {
	return &ResponseError{Code: fiber.StatusBadRequest, AppCode: "BAD_REQUEST", Message: message}
}

func Unauthorized(message string) *ResponseError {
	return &ResponseError{Code: fiber.StatusUnauthorized, AppCode: "UNAUTHORIZED", Message: message}
}

func Forbidden(message string) *ResponseError {
	return &ResponseError{Code: fiber.StatusForbidden, AppCode: "FORBIDDEN", Message: message}
}

func NotFound(message string) *ResponseError {
	return &ResponseError{Code: fiber.StatusNotFound, AppCode: "NOT_FOUND", Message: message}
}

func Conflict(message string) *ResponseError {
	return &ResponseError{Code: fiber.StatusConflict, AppCode: "CONFLICT", Message: message}
}
