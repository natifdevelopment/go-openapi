package builder

import (
	"github.com/natifdevelopment/go-openapi/schemas"
)

// AssignSecurity returns the security requirements for a given path.
// Public paths get empty security (no auth required).
// All other paths get the default security scheme.
func AssignSecurity(path string, publicPaths map[string]bool, defaultSecurity string) []interface{} {
	if publicPaths[path] {
		return []interface{}{}
	}
	return []interface{}{
		map[string]interface{}{defaultSecurity: []interface{}{}},
	}
}

// DefaultSecurityForPath determines the security scheme for a path based on
// common patterns in the BBO platform.
func DefaultSecurityForPath(path string) string {
	publicAuthPaths := []string{
		"/auth/v1/auth/login",
		"/auth/v1/auth/login/init",
		"/auth/v1/auth/login/otp",
		"/auth/v1/auth/exchange",
		"/auth/v1/auth/captcha/init",
		"/auth/v1/auth/activate/validate",
		"/auth/v1/auth/activate",
		"/auth/v1/auth/forgotpassword",
		"/auth/v1/auth/forgotpassword/otp",
		"/auth/v1/auth/changepassword/otp",
		"/auth/v1/auth/changepassword/default",
		"/auth/v1/auth/changepassword/stop-all-session",
		"/auth/v1/auth/helpdesk",
	}

	for _, p := range publicAuthPaths {
		if path == p {
			return ""
		}
	}

	if path == "/health" || path == "/" {
		return ""
	}

	if len(path) > 7 && path[:7] == "/_jobs/" {
		return schemas.SecurityApiKeyAuth
	}

	if path == "/metrics" {
		return schemas.SecurityApiKeyAuth
	}

	return schemas.SecurityBearerAuth
}
