package schemas

// HTTP status code constants used in standard responses.
// These map to the standard response definitions in the OpenAPI spec.
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusConflict            = 409
	StatusUnprocessableEntity = 422
	StatusTooManyRequests     = 429
	StatusInternalServerError = 500
	StatusServiceUnavailable  = 503
)

// StandardResponseCodes returns the list of HTTP status codes that every
// endpoint should document in its responses section.
func StandardResponseCodes() []int {
	return []int{
		StatusOK,
		StatusCreated,
		StatusBadRequest,
		StatusUnauthorized,
		StatusForbidden,
		StatusNotFound,
		StatusConflict,
		StatusUnprocessableEntity,
		StatusTooManyRequests,
		StatusInternalServerError,
		StatusServiceUnavailable,
	}
}

// CodeDescription returns a human-readable description for a status code.
func CodeDescription(code int) string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusCreated:
		return "Created"
	case StatusBadRequest:
		return "Bad Request"
	case StatusUnauthorized:
		return "Unauthorized"
	case StatusForbidden:
		return "Forbidden"
	case StatusNotFound:
		return "Not Found"
	case StatusConflict:
		return "Conflict"
	case StatusUnprocessableEntity:
		return "Unprocessable Entity"
	case StatusTooManyRequests:
		return "Too Many Requests"
	case StatusInternalServerError:
		return "Internal Server Error"
	case StatusServiceUnavailable:
		return "Service Unavailable"
	default:
		return "Unknown"
	}
}

// ErrorCodeForStatus maps an HTTP status code to the error code used in StandardError.
func ErrorCodeForStatus(code int) string {
	switch code {
	case StatusBadRequest:
		return CodeBadRequest
	case StatusUnauthorized:
		return CodeUnauthorized
	case StatusForbidden:
		return CodeForbidden
	case StatusNotFound:
		return CodeNotFound
	case StatusConflict:
		return CodeConflict
	case StatusUnprocessableEntity:
		return CodeUnprocessableEntity
	case StatusTooManyRequests:
		return CodeRateLimited
	case StatusInternalServerError:
		return CodeInternalError
	case StatusServiceUnavailable:
		return CodeServiceUnavailable
	default:
		return CodeInternalError
	}
}
