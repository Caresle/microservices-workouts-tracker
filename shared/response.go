package shared

type ApiResponse struct {
	Data      []any      `json:"data"`
	Errors    []APIError `json:"errors,omitempty"` // Array for multi-field errors
	Meta      *Meta      `json:"meta,omitempty"`
	Timestamp int64      `json:"timestamp"`
}

type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Field   string                 `json:"field,omitempty"` // For validation errors
	Details map[string]interface{} `json:"details,omitempty"`
}

type Meta struct {
	RequestID string      `json:"request_id"` // For distributed tracing
	Page      *Pagination `json:"page,omitempty"`
}

type Pagination struct {
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
}
