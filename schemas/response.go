// Package schemas defines the standard response and error schemas used across
// all BBO microservices. These structs serve dual purpose:
//
//  1. Runtime: services import and use them as response wrappers.
//  2. OpenAPI: the builder/merger references them to generate shared components.
//
// When modifying these structs, ensure the OpenAPI schema definition in
// schemas.go stays in sync.
package schemas

import "time"

// StandardResponse is the universal success response wrapper.
// Every successful endpoint returns this structure.
type StandardResponse struct {
	Status     bool                `json:"status" example:"true"`
	Message    string              `json:"message,omitempty" example:"OK"`
	Data       interface{}         `json:"data,omitempty"`
	Pagination *StandardPagination `json:"pagination,omitempty"`
	TraceID    string              `json:"trace_id" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	Timestamp  time.Time           `json:"timestamp" example:"2026-07-14T10:30:00Z"`
}

// StandardPagination is the pagination metadata included in list responses.
type StandardPagination struct {
	Page       int    `json:"page" example:"1"`
	PageSize   int    `json:"page_size" example:"10"`
	Sort       string `json:"sort,omitempty" example:"-created_at"`
	TotalRows  int64  `json:"total_rows" example:"150"`
	TotalPages int    `json:"total_pages" example:"15"`
}

// NewSuccess creates a StandardResponse with status=true and current timestamp.
func NewSuccess(data interface{}) StandardResponse {
	return StandardResponse{
		Status:    true,
		Message:   "OK",
		Data:      data,
		TraceID:   "", // set by middleware
		Timestamp: time.Now().UTC(),
	}
}

// NewPaginated creates a StandardResponse with pagination metadata.
func NewPaginated(data interface{}, pagination StandardPagination) StandardResponse {
	return StandardResponse{
		Status:     true,
		Message:    "OK",
		Data:       data,
		Pagination: &pagination,
		TraceID:    "",
		Timestamp:  time.Now().UTC(),
	}
}
