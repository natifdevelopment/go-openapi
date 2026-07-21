package merger

import (
	"fmt"
	"strings"

	"github.com/natifdevelopment/go-openapi/oas"
)

// ValidationResult represents a single validation finding.
type ValidationResult struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Path     string `json:"path,omitempty"`
	Method   string `json:"method,omitempty"`
	Message  string `json:"message"`
}

// Validate runs all validation rules against the merged spec.
func (m *Merger) Validate(spec *oas.Spec) {
	m.checkDuplicatePaths(spec)
	m.checkBrokenReferences(spec)
	m.checkMissingTags(spec)
	m.checkMissingSecurity(spec)
	m.checkMissingSummary(spec)
	m.checkMissingDescription(spec)
}

func (m *Merger) checkDuplicatePaths(_ *oas.Spec) {
	for _, c := range m.conflicts {
		if c.Type == "duplicate_path" {
			m.errors = append(m.errors, ValidationResult{
				Rule: "duplicate_path", Severity: "error",
				Path: c.Path, Message: c.Detail,
			})
		}
	}
}

func (m *Merger) checkBrokenReferences(spec *oas.Spec) {
	validSchemas := make(map[string]bool)
	for name := range spec.Schemas() {
		validSchemas["#/components/schemas/"+name] = true
	}

	paths := spec.Paths()
	for path, pathItemVal := range paths {
		pathItem, ok := pathItemVal.(map[string]interface{})
		if !ok {
			continue
		}
		for method, op := range oas.GetOperations(pathItem) {
			if op == nil {
				continue
			}
			m.checkRefsInOpMap(op, path, method, validSchemas)
		}
	}
}

func (m *Merger) checkRefsInOpMap(op map[string]interface{}, path, method string, validSchemas map[string]bool) {
	m.checkRefsInMapRecursive(op, path, method, validSchemas)
}

func (m *Merger) checkRefsInMapRecursive(obj map[string]interface{}, path, method string, validSchemas map[string]bool) {
	for key, val := range obj {
		switch v := val.(type) {
		case string:
			if key == "$ref" && strings.Contains(v, "/schemas/") {
				if !validSchemas[v] {
					m.errors = append(m.errors, ValidationResult{
						Rule: "broken_reference", Severity: "error",
						Path: path, Method: method,
						Message: fmt.Sprintf("broken $ref: %s", v),
					})
				}
			}
		case map[string]interface{}:
			m.checkRefsInMapRecursive(v, path, method, validSchemas)
		case []interface{}:
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					m.checkRefsInMapRecursive(itemMap, path, method, validSchemas)
				}
			}
		}
	}
}

func (m *Merger) checkMissingTags(spec *oas.Spec) {
	for path, pathItemVal := range spec.Paths() {
		pathItem, ok := pathItemVal.(map[string]interface{})
		if !ok {
			continue
		}
		for method, op := range oas.GetOperations(pathItem) {
			if op == nil {
				continue
			}
			tags, ok := op["tags"]
			if !ok || tags == nil {
				m.errors = append(m.errors, ValidationResult{
					Rule: "missing_tag", Severity: "error",
					Path: path, Method: method, Message: "operation has no tags",
				})
			}
		}
	}
}

func (m *Merger) checkMissingSecurity(spec *oas.Spec) {
	for path, pathItemVal := range spec.Paths() {
		pathItem, ok := pathItemVal.(map[string]interface{})
		if !ok {
			continue
		}
		for method, op := range oas.GetOperations(pathItem) {
			if op == nil {
				continue
			}
			if _, hasSec := op["security"]; !hasSec {
				m.errors = append(m.errors, ValidationResult{
					Rule: "missing_security", Severity: "error",
					Path: path, Method: method,
					Message: "operation has no security requirement (use empty array for public)",
				})
			}
		}
	}
}

func (m *Merger) checkMissingSummary(spec *oas.Spec) {
	for path, pathItemVal := range spec.Paths() {
		pathItem, ok := pathItemVal.(map[string]interface{})
		if !ok {
			continue
		}
		for method, op := range oas.GetOperations(pathItem) {
			if op == nil {
				continue
			}
			summary, _ := op["summary"].(string)
			if summary == "" {
				m.warnings = append(m.warnings, ValidationResult{
					Rule: "missing_summary", Severity: "warning",
					Path: path, Method: method, Message: "operation has no summary",
				})
			}
		}
	}
}

func (m *Merger) checkMissingDescription(spec *oas.Spec) {
	for path, pathItemVal := range spec.Paths() {
		pathItem, ok := pathItemVal.(map[string]interface{})
		if !ok {
			continue
		}
		for method, op := range oas.GetOperations(pathItem) {
			if op == nil {
				continue
			}
			desc, _ := op["description"].(string)
			if desc == "" {
				m.warnings = append(m.warnings, ValidationResult{
					Rule: "missing_description", Severity: "warning",
					Path: path, Method: method, Message: "operation has no description",
				})
			}
		}
	}
}
