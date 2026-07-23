package mapper

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mkhsnw/rel/internal/foundation/response"
)

// MapValidation maps validator.ValidationErrors to a list of field error maps.
func MapValidation(err validator.ValidationErrors) []map[string]string {
	var errFields []map[string]string
	for _, errField := range err {
		errFields = append(errFields, map[string]string{
			"field":   errField.Field(),
			"message": fmt.Sprintf("failed on '%s' validation", errField.Tag()),
		})
	}
	return errFields
}

// MapPagination maps raw pagination data to response.Paging structure.
func MapPagination(page int, size int, totalItems int64, totalPages int, hasNext bool, hasPrev bool, nextCursor string) *response.Paging {
	return &response.Paging{
		Page:       page,
		Size:       size,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		NextCursor: nextCursor,
	}
}
