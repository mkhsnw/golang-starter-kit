package model

type WebResponse[T any] struct {
	Data   T             `json:"data,omitempty"`
	Paging *PageMetadata `json:"paging,omitempty"`
	Error  *ErrorDetail  `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type PageMetadata struct {
	Page       int     `json:"page,omitempty"`
	Size       int     `json:"size"`
	TotalItem  int64   `json:"total_item,omitempty"`
	TotalPage  int64   `json:"total_page,omitempty"`
	NextCursor *string `json:"next_cursor,omitempty"`
}
