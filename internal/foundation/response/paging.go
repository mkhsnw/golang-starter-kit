package response

type Paging struct {
	Page int `json:"page"`
	Size int `json:"size"`

	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`

	HasNext bool `json:"has_next"`
	HasPrev bool `json:"has_prev"`

	NextCursor string `json:"next_cursor,omitempty"`
}
