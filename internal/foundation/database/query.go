package database

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// QueryFilter defines standard criteria for paginated list APIs.
type QueryFilter struct {
	Page    int            `json:"page"`
	Size    int            `json:"size"`
	Search  string         `json:"search"`
	SortBy  string         `json:"sort_by"`
	SortDir string         `json:"sort_dir"` // "asc" or "desc"
	Filters map[string]any `json:"filters"`
}

// Normalize ensures safe defaults for page, size, and sort direction.
func (q *QueryFilter) Normalize() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 {
		q.Size = 10
	}
	if q.Size > 100 {
		q.Size = 100
	}
	q.SortDir = strings.ToLower(q.SortDir)
	if q.SortDir != "asc" && q.SortDir != "desc" {
		q.SortDir = "desc"
	}
}

// ApplyQueryFilter applies search, field filters, and sorting to a GORM query.
func ApplyQueryFilter(db *gorm.DB, q QueryFilter, searchColumns ...string) *gorm.DB {
	q.Normalize()

	// Apply Search
	if q.Search != "" && len(searchColumns) > 0 {
		var clauses []string
		var args []any
		likePattern := "%" + q.Search + "%"
		for _, col := range searchColumns {
			clauses = append(clauses, fmt.Sprintf("%s LIKE ?", col))
			args = append(args, likePattern)
		}
		db = db.Where("("+strings.Join(clauses, " OR ")+")", args...)
	}

	// Apply Field Filters
	for col, val := range q.Filters {
		if val != nil && val != "" {
			db = db.Where(fmt.Sprintf("%s = ?", col), val)
		}
	}

	// Apply Ordering
	if q.SortBy != "" {
		db = db.Order(fmt.Sprintf("%s %s", q.SortBy, strings.ToUpper(q.SortDir)))
	} else {
		db = db.Order(fmt.Sprintf("created_at %s", strings.ToUpper(q.SortDir)))
	}

	return db
}
