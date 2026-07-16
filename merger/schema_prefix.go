package merger

import (
	"strings"

	"github.com/natifdevelopment/go-openapi/oas"
)

// PrefixAllSchemas applies schema name prefixing to all loaded specs.
func (m *Merger) PrefixAllSchemas() {
	for svcKey, spec := range m.specs {
		svc := m.Config.GetService(svcKey)
		if svc == nil || svc.SchemaPrefix == "" {
			continue
		}
		m.prefixSchemasForService(svcKey, spec, svc)
	}
}

func (m *Merger) prefixSchemasForService(svcKey string, spec *oas.Spec, svc *ServiceConfig) {
	schemasMap := spec.Schemas()
	if schemasMap == nil {
		return
	}

	newSchemas := make(map[string]interface{})
	refUpdates := make(map[string]string)

	for name, schema := range schemasMap {
		if isSharedSchemaName(name) {
			newSchemas[name] = schema
			continue
		}
		newName := svc.SchemaPrefix + name
		newSchemas[newName] = schema
		refUpdates["#/components/schemas/"+name] = "#/components/schemas/" + newName
	}

	spec.SetSchemas(newSchemas)
	m.updateAllRefsInSpec(spec, refUpdates)
}

func (m *Merger) updateAllRefsInSpec(spec *oas.Spec, refUpdates map[string]string) {
	// Update refs in paths
	paths := spec.Paths()
	for _, pathItemVal := range paths {
		if pathItem, ok := pathItemVal.(map[string]interface{}); ok {
			updateRefsInMap(pathItem, refUpdates)
		}
	}

	// Update refs in components
	comp := spec.Components()
	if comp == nil {
		return
	}
	for _, componentMap := range comp {
		if cm, ok := componentMap.(map[string]interface{}); ok {
			updateRefsInMap(cm, refUpdates)
		}
	}
}

func updateRefsInMap(obj map[string]interface{}, refUpdates map[string]string) {
	for key, val := range obj {
		switch v := val.(type) {
		case string:
			if key == "$ref" {
				if newRef, ok := refUpdates[v]; ok {
					obj[key] = newRef
				}
			}
		case map[string]interface{}:
			updateRefsInMap(v, refUpdates)
		case []interface{}:
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					updateRefsInMap(itemMap, refUpdates)
				}
			}
		}
	}
}

// updateRefsInSchemaMap is a helper that recursively updates $ref in a schema map.
func updateRefsInSchemaMap(schema map[string]interface{}, refUpdates map[string]string) {
	for key, val := range schema {
		switch v := val.(type) {
		case string:
			if key == "$ref" {
				if newRef, ok := refUpdates[v]; ok {
					schema[key] = newRef
				}
			}
		case map[string]interface{}:
			updateRefsInSchemaMap(v, refUpdates)
		case []interface{}:
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					updateRefsInSchemaMap(itemMap, refUpdates)
				}
			}
		}
	}
}

// refSchemaName extracts the schema name from a $ref string.
func refSchemaName(ref string) string {
	return strings.TrimPrefix(ref, "#/components/schemas/")
}
