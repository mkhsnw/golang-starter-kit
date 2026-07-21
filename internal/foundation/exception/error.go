package exception

import (
	"fmt"
)

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Details interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Is supports the errors.Is interface for robust error comparison.
func (e *APIError) Is(target error) bool {
	t, ok := target.(*APIError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// New creates a new APIError. It automatically looks up the HTTP status code from the mapper.
func New(code string, message string, details ...interface{}) *APIError {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	return &APIError{
		Code:    code,
		Message: message,
		Status:  MapCodeToStatus(code),
		Details: detail,
	}
}
