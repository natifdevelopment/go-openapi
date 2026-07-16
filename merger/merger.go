// Package merger implements the OpenAPI specification merger that combines
// per-service OpenAPI specs into a single aggregated spec for the API gateway.
package merger

import (
	"fmt"

	"github.com/natifdevelopment/go-openapi/oas"
	"github.com/natifdevelopment/go-openapi/schemas"
)

// Merger is the main merge orchestrator.
type Merger struct {
	Config    *Config
	specs     map[string]*oas.Spec
	conflicts []Conflict
	warnings  []ValidationResult
	errors    []ValidationResult
	pathOwner map[string]string
}

// New creates a new Merger from a config file path.
func New(configPath string) (*Merger, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return &Merger{
		Config:    cfg,
		specs:     make(map[string]*oas.Spec),
		pathOwner: make(map[string]string),
	}, nil
}

// Run executes the full merge pipeline and writes output.
func (m *Merger) Run() error {
	if err := m.LoadAll(); err != nil {
		return err
	}
	m.RewriteAllPaths()
	m.PrefixAllSchemas()
	merged := m.Merge()
	m.Validate(merged)
	m.PrintReport()

	if len(m.errors) > 0 {
		return fmt.Errorf("merge failed with %d validation errors", len(m.errors))
	}

	if err := m.WriteOutput(merged); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

// ValidateOnly runs the merge pipeline but does not write output.
func (m *Merger) ValidateOnly() error {
	if err := m.LoadAll(); err != nil {
		return err
	}
	m.RewriteAllPaths()
	m.PrefixAllSchemas()
	merged := m.Merge()
	m.Validate(merged)
	m.PrintReport()
	if len(m.errors) > 0 {
		return fmt.Errorf("validation failed with %d errors", len(m.errors))
	}
	return nil
}

// HasWarnings returns true if there are any validation warnings.
func (m *Merger) HasWarnings() bool { return len(m.warnings) > 0 }

// WarningCount returns the number of validation warnings.
func (m *Merger) WarningCount() int { return len(m.warnings) }

// Merge combines all loaded specs into a single OpenAPI spec.
func (m *Merger) Merge() *oas.Spec {
	merged := oas.New()
	merged.Set("openapi", "3.0.3")
	merged.Set("info", map[string]interface{}{
		"title":       m.Config.Gateway.Title,
		"description": m.Config.Gateway.Description,
		"version":     m.Config.Gateway.Version,
		"contact":     map[string]interface{}{"name": "BBO Team"},
		"license":     map[string]interface{}{"name": "MIT"},
	})
	merged.Set("paths", map[string]interface{}{})
	merged.Set("components", map[string]interface{}{
		"schemas":         schemas.BuildSharedSchemas(),
		"securitySchemes": schemas.BuildSecuritySchemes(),
		"responses":       map[string]interface{}{},
		"parameters":      map[string]interface{}{},
		"requestBodies":   map[string]interface{}{},
	})
	merged.Set("tags", schemas.StandardTags())
	merged.Set("servers", []interface{}{
		map[string]interface{}{
			"url":         m.Config.Gateway.ServerURL,
			"description": "BBO API Gateway",
		},
	})

	m.mergeComponents(merged)
	m.mergePaths(merged)
	merged.SetTags(m.mergeTags())
	m.injectSecurity(merged)

	return merged
}

// mergeComponents merges schemas, responses, parameters, requestBodies from all specs.
func (m *Merger) mergeComponents(merged *oas.Spec) {
	mergedSchemas := merged.Schemas()
	mergedResponses := map[string]interface{}{}
	mergedParams := map[string]interface{}{}
	mergedReqBodies := map[string]interface{}{}
	mergedSecSchemes := merged.SecuritySchemes()

	for svcKey, spec := range m.specs {
		svc := m.Config.GetService(svcKey)
		comp := spec.Components()
		if comp == nil {
			continue
		}

		// Schemas
		if ss := spec.Schemas(); ss != nil {
			for name, schema := range ss {
				if schemas.IsSharedSchema(name) {
					continue
				}
				if _, exists := mergedSchemas[name]; exists {
					m.addConflict("duplicate_schema", svcKey, name,
						fmt.Sprintf("schema '%s' already exists", name))
					continue
				}
				mergedSchemas[name] = schema
			}
		}

		// Security schemes (dedup by key)
		if secSchemes, ok := comp["securitySchemes"].(map[string]interface{}); ok {
			for key, scheme := range secSchemes {
				if _, exists := mergedSecSchemes[key]; !exists {
					mergedSecSchemes[key] = scheme
				}
			}
		}

		// Responses (prefix non-shared)
		if resps, ok := comp["responses"].(map[string]interface{}); ok {
			for name, resp := range resps {
				finalName := name
				if !isSharedResponse(name) {
					finalName = svc.SchemaPrefix + name
				}
				if _, exists := mergedResponses[finalName]; !exists {
					mergedResponses[finalName] = resp
				}
			}
		}

		// Parameters (prefix non-shared)
		if params, ok := comp["parameters"].(map[string]interface{}); ok {
			for name, param := range params {
				finalName := name
				if !isSharedParam(name) {
					finalName = svc.SchemaPrefix + name
				}
				if _, exists := mergedParams[finalName]; !exists {
					mergedParams[finalName] = param
				}
			}
		}

		// RequestBodies (prefix)
		if rbs, ok := comp["requestBodies"].(map[string]interface{}); ok {
			for name, rb := range rbs {
				finalName := svc.SchemaPrefix + name
				if _, exists := mergedReqBodies[finalName]; !exists {
					mergedReqBodies[finalName] = rb
				}
			}
		}
	}

	comp := merged.Components()
	comp["schemas"] = mergedSchemas
	comp["securitySchemes"] = mergedSecSchemes
	comp["responses"] = mergedResponses
	comp["parameters"] = mergedParams
	comp["requestBodies"] = mergedReqBodies
}

// mergePaths merges all paths from all specs into the merged spec.
func (m *Merger) mergePaths(merged *oas.Spec) {
	mergedPaths := merged.Paths()
	if mergedPaths == nil {
		mergedPaths = map[string]interface{}{}
	}
	pathMethodOwner := make(map[string]map[string]string)

	for svcKey, spec := range m.specs {
		paths := spec.Paths()
		if paths == nil {
			continue
		}
		for path, pathItemVal := range paths {
			pathItem, ok := pathItemVal.(map[string]interface{})
			if !ok {
				continue
			}

			if _, exists := pathMethodOwner[path]; !exists {
				pathMethodOwner[path] = make(map[string]string)
			}

			for method, opVal := range oas.GetOperations(pathItem) {
				if opVal == nil {
					continue
				}
				if existingSvc, exists := pathMethodOwner[path][method]; exists {
					m.addConflict("duplicate_path", svcKey, path,
						fmt.Sprintf("path '%s %s' already defined by service '%s'",
							method, path, existingSvc))
					continue
				}
				pathMethodOwner[path][method] = svcKey
				m.pathOwner[path] = svcKey
			}

			if existing, ok := mergedPaths[path].(map[string]interface{}); ok {
				mergePathItemMap(existing, pathItem)
			} else {
				mergedPaths[path] = pathItem
			}
		}
	}

	merged.SetPaths(mergedPaths)
}

// mergeTags collects and deduplicates tags from all specs.
func (m *Merger) mergeTags() []interface{} {
	tagMap := make(map[string]map[string]interface{})

	for _, tag := range schemas.StandardTags() {
		if t, ok := tag.(map[string]interface{}); ok {
			name, _ := t["name"].(string)
			tagMap[name] = t
		}
	}

	for _, spec := range m.specs {
		for _, tag := range spec.Tags() {
			t, ok := tag.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := t["name"].(string)
			if existing, ok := tagMap[name]; ok {
				existingDesc, _ := existing["description"].(string)
				newDesc, _ := t["description"].(string)
				if len(newDesc) > len(existingDesc) {
					existing["description"] = newDesc
				}
			} else {
				tagMap[name] = t
			}
		}
	}

	return sortTagsMap(tagMap)
}

// injectSecurity assigns security requirements to each operation.
func (m *Merger) injectSecurity(merged *oas.Spec) {
	paths := merged.Paths()
	for path, pathItemVal := range paths {
		pathItem, ok := pathItemVal.(map[string]interface{})
		if !ok {
			continue
		}

		svcKey, ok := m.pathOwner[path]
		if !ok {
			continue
		}
		svc := m.Config.GetService(svcKey)
		publicSet := buildPublicPathSet(svc)

		for _, op := range oas.GetOperations(pathItem) {
			if op == nil {
				continue
			}
			if _, hasSec := op["security"]; hasSec {
				continue
			}

			if publicSet[path] {
				op["security"] = []interface{}{}
			} else {
				op["security"] = []interface{}{
					map[string]interface{}{svc.DefaultSecurity: []interface{}{}},
				}
			}
		}
	}
}

// WriteOutput writes the merged spec to JSON/YAML and the report.
func (m *Merger) WriteOutput(spec *oas.Spec) error {
	if err := spec.WriteJSONFile(m.Config.Output.Path); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}
	if m.Config.Output.YAMLPath != "" {
		if err := spec.WriteYAMLFile(m.Config.Output.YAMLPath); err != nil {
			return fmt.Errorf("write YAML: %w", err)
		}
	}
	if err := m.WriteReport(spec); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

// mergePathItemMap merges operations from src into dst (non-overwriting).
func mergePathItemMap(dst, src map[string]interface{}) {
	for _, method := range oas.HTTPMethods {
		if _, ok := src[method]; ok {
			if _, exists := dst[method]; !exists {
				dst[method] = src[method]
			}
		}
	}
	if params, ok := src["parameters"].([]interface{}); ok {
		if existing, ok := dst["parameters"].([]interface{}); ok {
			dst["parameters"] = append(existing, params...)
		} else {
			dst["parameters"] = params
		}
	}
}

func isSharedResponse(name string) bool {
	for _, s := range []string{"StandardResponse", "StandardError", "ValidationError", "ErrorDetail"} {
		if s == name {
			return true
		}
	}
	return false
}

func isSharedParam(name string) bool {
	for _, s := range []string{"PageParam", "PageSizeParam", "SortParam", "IdParam"} {
		if s == name {
			return true
		}
	}
	return false
}

func buildPublicPathSet(svc *ServiceConfig) map[string]bool {
	set := make(map[string]bool)
	for _, p := range svc.PublicPaths {
		set[p] = true
	}
	return set
}

func sortTagsMap(tagMap map[string]map[string]interface{}) []interface{} {
	names := make([]string, 0, len(tagMap))
	for name := range tagMap {
		names = append(names, name)
	}
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	tags := make([]interface{}, 0, len(names))
	for _, name := range names {
		tags = append(tags, tagMap[name])
	}
	return tags
}
