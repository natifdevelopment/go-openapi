package schemas

import "time"

// StandardError is the universal error response wrapper.
// Every failed endpoint returns this structure.
type StandardError struct {
	Status    bool       `json:"status" example:"false"`
	Error     ErrorDetail `json:"error"`
	TraceID   string     `json:"trace_id" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	Timestamp time.Time  `json:"timestamp" example:"2026-07-14T10:30:00Z"`
}

// ErrorDetail contains the structured error information.
type ErrorDetail struct {
	Code    string            `json:"code" example:"VALIDATION_ERROR"`
	Message string            `json:"message" example:"Invalid input parameters"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationError represents a single field validation failure.
type ValidationError struct {
	Field       string `json:"field" example:"email"`
	Rule        string `json:"rule" example:"required"`
	RuleMessage string `json:"rule_message" example:"email is required"`
}

// Error codes used in ErrorDetail.Code.
// These are constants to ensure consistency across services.
const (
	CodeValidationError      = "VALIDATION_ERROR"
	CodeUnauthorized         = "UNAUTHORIZED"
	CodeForbidden            = "FORBIDDEN"
	CodeNotFound             = "NOT_FOUND"
	CodeConflict             = "CONFLICT"
	CodeUnprocessableEntity  = "UNPROCESSABLE_ENTITY"
	CodeRateLimited          = "RATE_LIMITED"
	CodeInternalError        = "INTERNAL_ERROR"
	CodeServiceUnavailable   = "SERVICE_UNAVAILABLE"
	CodeBadRequest           = "BAD_REQUEST"
)

// NewError creates a StandardError with the given code and message.
func NewError(code string, message string) StandardError {
	return StandardError{
		Status: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
		TraceID:   "",
		Timestamp: time.Now().UTC(),
	}
}

// NewValidationError creates a StandardError with validation error details.
func NewValidationError(message string, details []ValidationError) StandardError {
	return StandardError{
		Status: false,
		Error: ErrorDetail{
			Code:    CodeValidationError,
			Message: message,
			Details: details,
		},
		TraceID:   "",
		Timestamp: time.Now().UTC(),
	}
}
