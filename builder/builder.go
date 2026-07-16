// Package builder provides a programmatic OpenAPI spec builder that generates
// path and schema definitions from Go struct types via reflection.
//
// This is primarily used by bbo-api which has 200+ modules using a generic
// BaseRouter pattern. Instead of annotating each handler manually with swaggo
// comments, the builder reflects the Model and Request structs to auto-generate
// the standard CRUD endpoints per module.
package builder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/natifdevelopment/go-openapi/schemas"
)

// SpecBuilder accumulates paths, schemas, and tags during the build process.
// After registering all modules, call Build() to get the final OpenAPI spec
// as a map[string]interface{} ready for JSON/YAML serialization.
type SpecBuilder struct {
	Title       string
	Description string
	Version     string
	BasePath    string // e.g., "/api/v1"

	paths   map[string]interface{} // path → PathItem
	schemas map[string]interface{} // schemaName → schema
	tags    *TagRegistry
}

// NewSpecBuilder creates a new SpecBuilder with the given metadata.
func NewSpecBuilder(title, description, version, basePath string) *SpecBuilder {
	return &SpecBuilder{
		Title:       title,
		Description: description,
		Version:     version,
		BasePath:    basePath,
		paths:       make(map[string]interface{}),
		schemas:     make(map[string]interface{}),
		tags:        NewTagRegistry(),
	}
}

// RegisterModule registers a module with its model and request types.
// modulePath is the URL path segment (e.g., "/access-level").
// tag is the Swagger UI group name (e.g., "User Management").
// modelType and requestType are the reflect.Type of the Model and Request structs.
func (b *SpecBuilder) RegisterModule(modulePath, tag string, modelType, requestType reflect.Type) {
	b.tags.Register(tag, "")

	modelSchemaName := b.registerSchema(modelType)
	requestSchemaName := b.registerSchema(requestType)

	crudPaths := GenerateCRUDPaths(b.BasePath+modulePath, tag, modelSchemaName, requestSchemaName)
	for path, pathItem := range crudPaths {
		b.paths[path] = pathItem
	}
}

// RegisterCustomPath adds a custom (non-CRUD) path to the spec.
func (b *SpecBuilder) RegisterCustomPath(path, method, tag, summary, description string,
	requestSchemaName string, responseSchemaName string, security []string) {

	b.tags.Register(tag, "")

	operation := map[string]interface{}{
		"tags":        []interface{}{tag},
		"summary":     summary,
		"description": description,
		"responses":   b.buildStandardResponses(responseSchemaName),
	}

	if requestSchemaName != "" {
		operation["requestBody"] = map[string]interface{}{
			"required": true,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{"$ref": "#/components/schemas/" + requestSchemaName},
				},
			},
		}
	}

	if len(security) > 0 {
		secReqs := []interface{}{}
		for _, s := range security {
			secReqs = append(secReqs, map[string]interface{}{s: []interface{}{}})
		}
		operation["security"] = secReqs
	} else {
		operation["security"] = []interface{}{}
	}

	pathItem, ok := b.paths[path].(map[string]interface{})
	if !ok {
		pathItem = make(map[string]interface{})
	}
	pathItem[strings.ToLower(method)] = operation
	b.paths[path] = pathItem
}

// Build returns the final OpenAPI 3.0 spec as a map[string]interface{}.
func (b *SpecBuilder) Build() map[string]interface{} {
	// Merge shared schemas with custom schemas
	allSchemas := schemas.BuildSharedSchemas()
	for name, schema := range b.schemas {
		if !schemas.IsSharedSchema(name) {
			allSchemas[name] = schema
		}
	}

	spec := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       b.Title,
			"description": b.Description,
			"version":     b.Version,
		},
		"paths": b.paths,
		"components": map[string]interface{}{
			"schemas":         allSchemas,
			"securitySchemes": schemas.BuildSecuritySchemes(),
		},
		"tags":    b.tags.Sorted(),
		"servers": []interface{}{map[string]interface{}{"url": "/", "description": "Service local"}},
	}

	return spec
}

// registerSchema converts a Go struct type to an OpenAPI schema and registers it.
// Returns the schema name (the struct name).
func (b *SpecBuilder) registerSchema(t reflect.Type) string {
	if t == nil {
		return ""
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}

	name := t.Name()
	if name == "" {
		return ""
	}

	if _, exists := b.schemas[name]; exists {
		return name
	}
	if schemas.IsSharedSchema(name) {
		return name
	}

	schema := StructToSchemaMap(t)
	if schema != nil {
		b.schemas[name] = schema
	}
	return name
}

// buildStandardResponses creates the standard response set for an operation.
func (b *SpecBuilder) buildStandardResponses(dataSchemaName string) map[string]interface{} {
	responses := make(map[string]interface{})

	responses["200"] = map[string]interface{}{
		"description": "OK",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": b.wrapInStandardResponse(dataSchemaName),
			},
		},
	}
	responses["201"] = map[string]interface{}{
		"description": "Created",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": b.wrapInStandardResponse(dataSchemaName),
			},
		},
	}

	for _, code := range schemas.StandardResponseCodes() {
		if code == 200 || code == 201 {
			continue
		}
		desc := schemas.CodeDescription(code)
		responses[fmt.Sprintf("%d", code)] = schemas.StandardErrorResponse(code, desc)
	}

	return responses
}

// wrapInStandardResponse creates a StandardResponse schema ref wrapping the data schema.
func (b *SpecBuilder) wrapInStandardResponse(dataSchemaName string) map[string]interface{} {
	if dataSchemaName == "" {
		return map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"}
	}
	return map[string]interface{}{
		"allOf": []interface{}{
			map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
			map[string]interface{}{
				"properties": map[string]interface{}{
					"data": map[string]interface{}{"$ref": "#/components/schemas/" + dataSchemaName},
				},
			},
		},
	}
}
