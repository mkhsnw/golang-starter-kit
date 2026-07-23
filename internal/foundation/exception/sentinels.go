package exception

// Standard immutable API error sentinels for common domain exception scenarios.
var (
	ErrNotFound     = New(NOT_FOUND, "Requested resource was not found")
	ErrUnauthorized = New(UNAUTHORIZED, "Authentication credentials are missing or invalid")
	ErrForbidden    = New(FORBIDDEN, "You do not have permission to perform this action")
	ErrValidation   = New(VALIDATION_ERROR, "Provided input data is invalid")
	ErrConflict     = New(DATABASE_ERROR, "Resource already exists or conflicts with existing state")
	ErrInternal     = New(INTERNAL_ERROR, "An unexpected internal server error occurred")
)
