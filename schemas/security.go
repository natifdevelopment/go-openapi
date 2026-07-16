package schemas

// Security scheme key constants.
// These match the keys used in the OpenAPI components.securitySchemes
// and in the @Security swaggo annotations.
const (
	SecurityBearerAuth = "BearerAuth"
	SecurityApiKeyAuth = "ApiKeyAuth"
	SecurityBasicAuth  = "BasicAuth"
	SecurityOAuth2     = "OAuth2"
)

// SecuritySchemeInfo describes a security scheme for documentation purposes.
type SecuritySchemeInfo struct {
	Key         string
	Type        string // http, apiKey, oauth2
	Scheme      string // bearer, basic (for type=http)
	In          string // header, query, cookie (for type=apiKey)
	Name        string // header/param name (for type=apiKey)
	BearerFormat string // JWT (for type=http, scheme=bearer)
	Description string
}

// DefaultSecuritySchemes returns the security schemes used by the BBO platform.
// The merger uses this to populate components.securitySchemes in the aggregated spec.
var DefaultSecuritySchemes = map[string]SecuritySchemeInfo{
	SecurityBearerAuth: {
		Key:          SecurityBearerAuth,
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT Bearer token. Obtained from /auth/v1/auth/login.",
	},
	SecurityApiKeyAuth: {
		Key:         SecurityApiKeyAuth,
		Type:        "apiKey",
		In:          "header",
		Name:        "X-API-Key",
		Description: "API key for gateway-level authentication and admin endpoints.",
	},
	SecurityBasicAuth: {
		Key:         SecurityBasicAuth,
		Type:        "http",
		Scheme:      "basic",
		Description: "Basic authentication (optional, for legacy integrations).",
	},
	SecurityOAuth2: {
		Key:         SecurityOAuth2,
		Type:        "oauth2",
		Description: "OAuth2 authorization code flow (future-ready).",
	},
}
