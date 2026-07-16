// Package swaggerui provides a reusable Gin handler for serving Swagger UI
// with an embedded OpenAPI specification.
//
// Usage in the gateway:
//
//	import openapiswagger "github.com/natifdevelopment/go-openapi/swaggerui"
//
//	r.GET("/swagger/*any", openapiswagger.Handler())
//	r.GET("/swagger/doc.json", openapiswagger.SpecHandler())
//
// The spec is embedded via go:embed in the gateway's main package and passed
// to this package via SetSpec().
package swaggerui

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// aggregatedSpec holds the embedded OpenAPI spec bytes.
// Set via SetSpec() before the server starts.
var aggregatedSpec []byte

// SetSpec sets the aggregated OpenAPI spec bytes.
// Call this in main.go before starting the server:
//
//	//go:embed docs/openapi-aggregated.json
//	var specBytes []byte
//	openapiswagger.SetSpec(specBytes)
func SetSpec(spec []byte) {
	aggregatedSpec = spec
}

// SpecHandler returns the raw OpenAPI JSON spec.
// Mount at: GET /swagger/doc.json
func SpecHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(aggregatedSpec) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  false,
				"message": "OpenAPI spec not available (not embedded at build time)",
			})
			return
		}
		c.Data(http.StatusOK, "application/json", aggregatedSpec)
	}
}

// Handler returns the Swagger UI handler.
// Mount at: GET /swagger/*any
// The Swagger UI will fetch the spec from /swagger/doc.json by default.
func Handler() gin.HandlerFunc {
	return ginSwagger.CustomWrapHandler(&ginSwagger.Config{
		URL:   "/swagger/doc.json",
		Title: "BBO API Gateway — Documentation",
	}, swaggerFiles.Handler)
}

// HandlerWithConfig returns a Swagger UI handler with custom configuration.
// url: the URL where Swagger UI should fetch the spec from.
// title: the browser tab title.
func HandlerWithConfig(url, title string) gin.HandlerFunc {
	if url == "" {
		url = "/swagger/doc.json"
	}
	if title == "" {
		title = "BBO API Gateway — Documentation"
	}
	return ginSwagger.CustomWrapHandler(&ginSwagger.Config{
		URL:   url,
		Title: title,
	}, swaggerFiles.Handler)
}

// IsSpecLoaded returns true if the aggregated spec has been set.
func IsSpecLoaded() bool {
	return len(aggregatedSpec) > 0
}

// SpecSize returns the size of the loaded spec in bytes.
func SpecSize() int {
	return len(aggregatedSpec)
}

// SpecPath returns the path where the spec is served.
func SpecPath() string {
	return "/swagger/doc.json"
}

// UIPath returns the path where Swagger UI is served.
func UIPath() string {
	return "/swagger/index.html"
}

// IsSwaggerRequest checks if a request path is for Swagger UI or spec.
func IsSwaggerRequest(path string) bool {
	return strings.HasPrefix(path, "/swagger/")
}
