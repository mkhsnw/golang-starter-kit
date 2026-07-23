package response

import (
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/mkhsnw/rel/internal/foundation/exception"
)

// Response wraps the standard API response structure.
type Response struct {
	Data   interface{}         `json:"data,omitempty"`
	Error  *exception.APIError `json:"error,omitempty"`
	Meta   *Meta               `json:"meta,omitempty"`
	Paging *Paging             `json:"paging,omitempty"`
}

// Builder helps construct HTTP responses fluently.
type Builder struct {
	response Response
	status   int
	stream   io.Reader
}

// OK creates a new response with status 200 OK.
func OK(data interface{}) *Builder {
	return &Builder{
		response: Response{Data: data},
		status:   fiber.StatusOK,
	}
}

// Created creates a new response with status 201 Created.
func Created(data interface{}) *Builder {
	return &Builder{
		response: Response{Data: data},
		status:   fiber.StatusCreated,
	}
}

// Paginate creates a new paginated response.
func Paginate(data interface{}) *Builder {
	return &Builder{
		response: Response{Data: data},
		status:   fiber.StatusOK,
	}
}

// Page creates a new paginated response (alias for Paginate).
func Page(data interface{}) *Builder {
	return Paginate(data)
}

// NoContent creates a new response with status 204 No Content.
func NoContent() *Builder {
	return &Builder{
		status: fiber.StatusNoContent,
	}
}

// Accepted creates a new response with status 202 Accepted.
func Accepted(data interface{}) *Builder {
	return &Builder{
		response: Response{Data: data},
		status:   fiber.StatusAccepted,
	}
}

// Stream creates a new response that sends an io.Reader stream.
func Stream(r io.Reader) *Builder {
	return &Builder{
		status: fiber.StatusOK,
		stream: r,
	}
}

// Error creates a new error response.
func Error(err error) *Builder {
	apiErr, ok := err.(*exception.APIError)
	if !ok {
		apiErr = exception.New(exception.INTERNAL_ERROR, err.Error())
	}
	return &Builder{
		response: Response{Error: apiErr},
		status:   apiErr.Status,
	}
}

// Meta adds metadata to the response.
func (b *Builder) Meta(meta *Meta) *Builder {
	b.response.Meta = meta
	return b
}

// Paging adds pagination details to the response.
func (b *Builder) Paging(paging *Paging) *Builder {
	b.response.Paging = paging
	return b
}

// Send sends the constructed response via Fiber context.
func (b *Builder) Send(c fiber.Ctx) error {
	if b.stream != nil {
		c.Status(b.status)
		return c.SendStream(b.stream)
	}
	if b.status == fiber.StatusNoContent {
		return c.SendStatus(b.status)
	}
	return c.Status(b.status).JSON(b.response)
}
