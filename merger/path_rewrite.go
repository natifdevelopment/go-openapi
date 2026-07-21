package merger

import (
	"strings"

	"github.com/natifdevelopment/go-openapi/oas"
)

// RewriteAllPaths applies path rewrite rules to all loaded specs.
func (m *Merger) RewriteAllPaths() {
	for svcKey, spec := range m.specs {
		svc := m.Config.GetService(svcKey)
		if svc == nil || len(svc.PathRewrite) == 0 {
			continue
		}
		m.rewritePathsForService(svcKey, spec, svc)
	}
}

func (m *Merger) rewritePathsForService(_ string, spec *oas.Spec, svc *ServiceConfig) {
	paths := spec.Paths()
	if paths == nil {
		return
	}

	newPaths := make(map[string]interface{})
	for oldPath, pathItemVal := range paths {
		newPath := oldPath
		for _, rule := range svc.PathRewrite {
			if strings.HasPrefix(oldPath, rule.From) {
				newPath = strings.Replace(oldPath, rule.From, rule.To, 1)
				break
			}
		}

		if pathItem, ok := pathItemVal.(map[string]interface{}); ok {
			m.updatePathRefs(pathItem, svc)
		}
		newPaths[newPath] = pathItemVal
	}
	spec.SetPaths(newPaths)
}

func (m *Merger) updatePathRefs(pathItem map[string]interface{}, svc *ServiceConfig) {
	for _, op := range oas.GetOperations(pathItem) {
		if op == nil {
			continue
		}
		// Update parameter refs
		if params, ok := op["parameters"].([]interface{}); ok {
			for _, p := range params {
				if param, ok := p.(map[string]interface{}); ok {
					m.updateMapRefs(param, svc)
				}
			}
		}
		// Update request body refs
		if rb, ok := op["requestBody"].(map[string]interface{}); ok {
			m.updateMapRefs(rb, svc)
		}
		// Update response refs
		if responses, ok := op["responses"].(map[string]interface{}); ok {
			for _, respVal := range responses {
				if resp, ok := respVal.(map[string]interface{}); ok {
					if ref, ok := resp["$ref"].(string); ok {
						resp["$ref"] = m.rewriteRef(ref, svc)
					}
					m.updateMapRefs(resp, svc)
				}
			}
		}
	}
}

func (m *Merger) updateMapRefs(obj map[string]interface{}, svc *ServiceConfig) {
	for key, val := range obj {
		switch v := val.(type) {
		case string:
			if key == "$ref" {
				obj[key] = m.rewriteRef(v, svc)
			}
		case map[string]interface{}:
			m.updateMapRefs(v, svc)
		case []interface{}:
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					m.updateMapRefs(itemMap, svc)
				}
			}
		}
	}
}

func (m *Merger) rewriteRef(ref string, svc *ServiceConfig) string {
	if !strings.HasPrefix(ref, "#/components/schemas/") {
		return ref
	}
	schemaName := strings.TrimPrefix(ref, "#/components/schemas/")
	if isSharedSchemaName(schemaName) {
		return ref
	}
	return "#/components/schemas/" + svc.SchemaPrefix + schemaName
}

func isSharedSchemaName(name string) bool {
	shared := []string{"StandardResponse", "StandardPagination", "StandardError", "ErrorDetail", "ValidationError"}
	for _, s := range shared {
		if s == name {
			return true
		}
	}
	return false
}
