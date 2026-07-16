// Package httpresponse provides Gin helper functions that return responses
// conforming to the StandardError and StandardResponse OpenAPI schemas.
//
// All helpers automatically extract the trace_id from the Gin context
// (set by the go-tracing GinMiddleware) and include it in the response.
package httpresponse

import (
	"net/http"

	"github.com/natifdevelopment/go-openapi/schemas"
	"github.com/gin-gonic/gin"
)

// traceIDFromContext extracts the trace_id from the Gin context.
// The go-tracing GinMiddleware stores it under the "trace_id" key.
func traceIDFromContext(c *gin.Context) string {
	return c.GetString("trace_id")
}

// ─── Success responses ───────────────────────────────────────────────────────

// SendSuccess sends a 200 response with status=true and message="OK".
func SendSuccess(c *gin.Context) {
	resp := schemas.NewSuccess(nil)
	resp.TraceID = traceIDFromContext(c)
	c.JSON(http.StatusOK, resp)
}

// SendSuccessWithData sends a 200 response with the given data payload.
func SendSuccessWithData(c *gin.Context, data interface{}) {
	resp := schemas.NewSuccess(data)
	resp.TraceID = traceIDFromContext(c)
	c.JSON(http.StatusOK, resp)
}

// SendSuccessWithMessage sends a 200 response with a custom message and data.
func SendSuccessWithMessage(c *gin.Context, message string, data interface{}) {
	resp := schemas.NewSuccess(data)
	resp.Message = message
	resp.TraceID = traceIDFromContext(c)
	c.JSON(http.StatusOK, resp)
}

// SendCreated sends a 201 response with the given data payload.
func SendCreated(c *gin.Context, data interface{}) {
	resp := schemas.NewSuccess(data)
	resp.Message = "Created"
	resp.TraceID = traceIDFromContext(c)
	c.JSON(http.StatusCreated, resp)
}

// SendPaginated sends a 200 response with data and pagination metadata.
func SendPaginated(c *gin.Context, data interface{}, pagination schemas.StandardPagination) {
	resp := schemas.NewPaginated(data, pagination)
	resp.TraceID = traceIDFromContext(c)
	c.JSON(http.StatusOK, resp)
}

// ─── Error responses ─────────────────────────────────────────────────────────

// sendError is the internal helper that builds and sends a StandardError.
func sendError(c *gin.Context, httpStatus int, code string, message string) {
	resp := schemas.NewError(code, message)
	resp.TraceID = traceIDFromContext(c)
	c.JSON(httpStatus, resp)
}

// SendError sends an error response with the given HTTP status, code, and message.
func SendError(c *gin.Context, httpStatus int, code string, message string) {
	sendError(c, httpStatus, code, message)
}

// SendBadRequest sends a 400 error response.
func SendBadRequest(c *gin.Context, message string) {
	sendError(c, http.StatusBadRequest, schemas.CodeBadRequest, message)
}

// SendUnauthorized sends a 401 error response.
func SendUnauthorized(c *gin.Context, message string) {
	sendError(c, http.StatusUnauthorized, schemas.CodeUnauthorized, message)
}

// SendForbidden sends a 403 error response.
func SendForbidden(c *gin.Context, message string) {
	sendError(c, http.StatusForbidden, schemas.CodeForbidden, message)
}

// SendNotFound sends a 404 error response.
func SendNotFound(c *gin.Context, message string) {
	sendError(c, http.StatusNotFound, schemas.CodeNotFound, message)
}

// SendConflict sends a 409 error response.
func SendConflict(c *gin.Context, message string) {
	sendError(c, http.StatusConflict, schemas.CodeConflict, message)
}

// SendUnprocessableEntity sends a 422 error response.
func SendUnprocessableEntity(c *gin.Context, message string) {
	sendError(c, http.StatusUnprocessableEntity, schemas.CodeUnprocessableEntity, message)
}

// SendRateLimited sends a 429 error response.
func SendRateLimited(c *gin.Context, message string) {
	sendError(c, http.StatusTooManyRequests, schemas.CodeRateLimited, message)
}

// SendInternalServerError sends a 500 error response.
func SendInternalServerError(c *gin.Context, message string) {
	sendError(c, http.StatusInternalServerError, schemas.CodeInternalError, message)
}

// SendServiceUnavailable sends a 503 error response.
func SendServiceUnavailable(c *gin.Context, message string) {
	sendError(c, http.StatusServiceUnavailable, schemas.CodeServiceUnavailable, message)
}

// SendValidationError sends a 400 error response with field-level validation details.
func SendValidationError(c *gin.Context, message string, details []schemas.ValidationError) {
	resp := schemas.NewValidationError(message, details)
	resp.TraceID = traceIDFromContext(c)
	c.JSON(http.StatusBadRequest, resp)
}

// SendErrorForStatus sends an error response using the standard error code
// mapped to the given HTTP status code. This is a convenience function
// when you only have the HTTP status code.
func SendErrorForStatus(c *gin.Context, httpStatus int, message string) {
	sendError(c, httpStatus, schemas.ErrorCodeForStatus(httpStatus), message)
}
