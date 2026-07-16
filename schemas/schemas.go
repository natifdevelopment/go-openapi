package schemas

import (
	"github.com/natifdevelopment/go-openapi/oas"
)

// SharedSchemaNames lists schema names that are identical across all services
// and should be deduplicated (not prefixed) during merge.
var SharedSchemaNames = []string{
	"StandardResponse",
	"StandardPagination",
	"StandardError",
	"ErrorDetail",
	"ValidationError",
}

// IsSharedSchema returns true if the schema name is in the shared set.
func IsSharedSchema(name string) bool {
	for _, s := range SharedSchemaNames {
		if s == name {
			return true
		}
	}
	return false
}

// BuildSharedSchemas returns the OpenAPI schema definitions for all shared
// (cross-service) schemas as a map[string]interface{}.
func BuildSharedSchemas() map[string]interface{} {
	return map[string]interface{}{
		"StandardPagination": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page":        map[string]interface{}{"type": "integer", "example": 1},
				"page_size":   map[string]interface{}{"type": "integer", "example": 10},
				"sort":        map[string]interface{}{"type": "string", "example": "-created_at"},
				"total_rows":  map[string]interface{}{"type": "integer", "format": "int64", "example": 150},
				"total_pages": map[string]interface{}{"type": "integer", "example": 15},
			},
		},
		"ValidationError": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"field":        map[string]interface{}{"type": "string", "example": "email"},
				"rule":         map[string]interface{}{"type": "string", "example": "required"},
				"rule_message": map[string]interface{}{"type": "string", "example": "email is required"},
			},
		},
		"ErrorDetail": map[string]interface{}{
			"type":     "object",
			"required": []interface{}{"code", "message"},
			"properties": map[string]interface{}{
				"code":    map[string]interface{}{"type": "string", "example": "VALIDATION_ERROR"},
				"message": map[string]interface{}{"type": "string", "example": "Invalid input parameters"},
				"details": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"$ref": "#/components/schemas/ValidationError"},
				},
			},
		},
		"StandardError": map[string]interface{}{
			"type":     "object",
			"required": []interface{}{"status", "error", "trace_id", "timestamp"},
			"properties": map[string]interface{}{
				"status":    map[string]interface{}{"type": "boolean", "example": false},
				"error":     map[string]interface{}{"$ref": "#/components/schemas/ErrorDetail"},
				"trace_id":  map[string]interface{}{"type": "string", "format": "uuid"},
				"timestamp": map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"StandardResponse": map[string]interface{}{
			"type":     "object",
			"required": []interface{}{"status", "trace_id", "timestamp"},
			"properties": map[string]interface{}{
				"status":     map[string]interface{}{"type": "boolean", "example": true},
				"message":    map[string]interface{}{"type": "string", "example": "OK"},
				"data":       map[string]interface{}{"type": "object", "description": "Response payload"},
				"pagination": map[string]interface{}{"$ref": "#/components/schemas/StandardPagination"},
				"trace_id":   map[string]interface{}{"type": "string", "format": "uuid"},
				"timestamp":  map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
	}
}

// BuildSecuritySchemes returns the OpenAPI security scheme definitions.
func BuildSecuritySchemes() map[string]interface{} {
	return map[string]interface{}{
		SecurityBearerAuth: map[string]interface{}{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
			"description":  "JWT Bearer token. Obtained from /auth/v1/auth/login.",
		},
		SecurityApiKeyAuth: map[string]interface{}{
			"type":        "apiKey",
			"in":          "header",
			"name":        "X-API-Key",
			"description": "API key for gateway-level authentication and admin endpoints.",
		},
		SecurityBasicAuth: map[string]interface{}{
			"type":        "http",
			"scheme":      "basic",
			"description": "Basic authentication (optional, for legacy integrations).",
		},
		SecurityOAuth2: map[string]interface{}{
			"type":        "oauth2",
			"description": "OAuth2 authorization code flow (future-ready).",
			"flows": map[string]interface{}{
				"authorizationCode": map[string]interface{}{
					"authorizationUrl": "/auth/v1/oauth/authorize",
					"tokenUrl":         "/auth/v1/oauth/token",
					"refreshUrl":       "/auth/v1/oauth/refresh",
					"scopes": map[string]interface{}{
						"read":  "Read access",
						"write": "Write access",
						"admin": "Administrative access",
					},
				},
			},
		},
	}
}

// StandardTags returns the ordered list of tags for the aggregated spec.
func StandardTags() []interface{} {
	return []interface{}{
		map[string]interface{}{"name": "Authentication", "description": "User authentication, login, OTP, PIN, biometric, device management"},
		map[string]interface{}{"name": "User Management", "description": "User accounts, access levels, access groups"},
		map[string]interface{}{"name": "Member", "description": "Member management"},
		map[string]interface{}{"name": "Koperasi", "description": "Koperasi management"},
		map[string]interface{}{"name": "Order", "description": "Order management"},
		map[string]interface{}{"name": "Product", "description": "Product management"},
		map[string]interface{}{"name": "Payment", "description": "Payment management"},
		map[string]interface{}{"name": "Notification", "description": "Notification management"},
		map[string]interface{}{"name": "Inventory", "description": "Stock and inventory management"},
		map[string]interface{}{"name": "Report", "description": "Report generation and export"},
		map[string]interface{}{"name": "Dashboard", "description": "Dashboard analytics and visualization"},
		map[string]interface{}{"name": "Background Jobs", "description": "Scheduled cron jobs and background tasks"},
		map[string]interface{}{"name": "Gateway", "description": "Gateway-level endpoints (health, metrics)"},
		map[string]interface{}{"name": "Health Check", "description": "Service health check endpoints"},
	}
}

// StandardErrorResponse returns a response object for an error status code.
func StandardErrorResponse(code int, desc string) map[string]interface{} {
	return map[string]interface{}{
		"description": desc,
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{"$ref": "#/components/schemas/StandardError"},
			},
		},
	}
}

// StandardSuccessResponse returns a 200 response object wrapping StandardResponse.
func StandardSuccessResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "OK",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
			},
		},
	}
}

// StandardErrorResponsesMap returns error responses keyed by status code string.
func StandardErrorResponsesMap() map[string]interface{} {
	responses := make(map[string]interface{})
	for _, code := range StandardResponseCodes() {
		if code == StatusOK || code == StatusCreated {
			continue
		}
		desc := CodeDescription(code)
		responses[intToStr(code)] = StandardErrorResponse(code, desc)
	}
	return responses
}

// intToStr converts int to string without importing strconv (avoid cycle issues).
func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	negative := i < 0
	if negative {
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// Ensure oas package is used (for future integration).
var _ = oas.New
