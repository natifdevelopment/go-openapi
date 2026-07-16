// Package oas provides minimal OpenAPI 3.0 type definitions using plain Go
// maps and structs. This avoids heavy external dependencies (kin-openapi)
// while still providing type-safe helpers for the merger and builder.
//
// The spec is represented as Spec (a typed wrapper around map[string]interface{})
// which can be loaded from JSON/YAML and manipulated programmatically.
package oas

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Spec is a generic OpenAPI 3.0 specification represented as a map.
// This allows flexible manipulation without rigid struct definitions.
type Spec struct {
	data map[string]interface{}
}

// New creates an empty Spec.
func New() *Spec {
	return &Spec{data: make(map[string]interface{})}
}

// Data returns the underlying map.
func (s *Spec) Data() map[string]interface{} {
	return s.data
}

// LoadFromJSONFile loads a spec from a JSON file.
func LoadFromJSONFile(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFromJSON(data)
}

// LoadFromJSON loads a spec from JSON bytes.
func LoadFromJSON(data []byte) (*Spec, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &Spec{data: m}, nil
}

// LoadFromYAMLFile loads a spec from a YAML file.
func LoadFromYAMLFile(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFromYAML(data)
}

// LoadFromYAML loads a spec from YAML bytes.
func LoadFromYAML(data []byte) (*Spec, error) {
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	// YAML unmarshals to map[string]interface{} with nested map[interface{}]interface{}
	normalizeYAMLMap(m)
	return &Spec{data: m}, nil
}

// LoadFromFile auto-detects JSON vs YAML from file extension.
func LoadFromFile(path string) (*Spec, error) {
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		return LoadFromYAMLFile(path)
	}
	return LoadFromJSONFile(path)
}

// ToJSON marshals the spec to JSON bytes.
func (s *Spec) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s.data, "", "  ")
}

// ToYAML marshals the spec to YAML bytes.
func (s *Spec) ToYAML() ([]byte, error) {
	return yaml.Marshal(s.data)
}

// WriteJSONFile writes the spec to a JSON file.
func (s *Spec) WriteJSONFile(path string) error {
	data, err := s.ToJSON()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// WriteYAMLFile writes the spec to a YAML file.
func (s *Spec) WriteYAMLFile(path string) error {
	data, err := s.ToYAML()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Get returns a value from the spec by key.
func (s *Spec) Get(key string) interface{} {
	return s.data[key]
}

// GetString returns a string value from the spec by key.
func (s *Spec) GetString(key string) string {
	v, ok := s.data[key].(string)
	if !ok {
		return ""
	}
	return v
}

// Set sets a value in the spec by key.
func (s *Spec) Set(key string, value interface{}) {
	s.data[key] = value
}

// Paths returns the paths map.
func (s *Spec) Paths() map[string]interface{} {
	paths, ok := s.data["paths"].(map[string]interface{})
	if !ok {
		return nil
	}
	return paths
}

// SetPaths sets the paths map.
func (s *Spec) SetPaths(paths map[string]interface{}) {
	s.data["paths"] = paths
}

// Components returns the components map.
func (s *Spec) Components() map[string]interface{} {
	comp, ok := s.data["components"].(map[string]interface{})
	if !ok {
		return nil
	}
	return comp
}

// Schemas returns the components.schemas map.
func (s *Spec) Schemas() map[string]interface{} {
	comp := s.Components()
	if comp == nil {
		return nil
	}
	schemas, ok := comp["schemas"].(map[string]interface{})
	if !ok {
		return nil
	}
	return schemas
}

// SetSchemas sets the components.schemas map.
func (s *Spec) SetSchemas(schemas map[string]interface{}) {
	comp := s.Components()
	if comp == nil {
		comp = make(map[string]interface{})
		s.data["components"] = comp
	}
	comp["schemas"] = schemas
}

// SecuritySchemes returns the components.securitySchemes map.
func (s *Spec) SecuritySchemes() map[string]interface{} {
	comp := s.Components()
	if comp == nil {
		return nil
	}
	ss, ok := comp["securitySchemes"].(map[string]interface{})
	if !ok {
		return nil
	}
	return ss
}

// Tags returns the tags array.
func (s *Spec) Tags() []interface{} {
	tags, ok := s.data["tags"].([]interface{})
	if !ok {
		return nil
	}
	return tags
}

// SetTags sets the tags array.
func (s *Spec) SetTags(tags []interface{}) {
	s.data["tags"] = tags
}

// Servers returns the servers array.
func (s *Spec) Servers() []interface{} {
	servers, ok := s.data["servers"].([]interface{})
	if !ok {
		return nil
	}
	return servers
}

// SetServers sets the servers array.
func (s *Spec) SetServers(servers []interface{}) {
	s.data["servers"] = servers
}

// Info returns the info map.
func (s *Spec) Info() map[string]interface{} {
	info, ok := s.data["info"].(map[string]interface{})
	if !ok {
		return nil
	}
	return info
}

// SetInfo sets the info map.
func (s *Spec) SetInfo(info map[string]interface{}) {
	s.data["info"] = info
}

// SortedPathKeys returns path keys sorted alphabetically.
func (s *Spec) SortedPathKeys() []string {
	paths := s.Paths()
	keys := make([]string, 0, len(paths))
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// HTTPMethods returns the standard HTTP methods that can appear in a path item.
var HTTPMethods = []string{"get", "post", "put", "delete", "patch", "head", "options"}

// GetOperations returns all operations from a path item as method → operation map.
func GetOperations(pathItem map[string]interface{}) map[string]map[string]interface{} {
	ops := make(map[string]map[string]interface{})
	for _, method := range HTTPMethods {
		if op, ok := pathItem[method].(map[string]interface{}); ok {
			ops[method] = op
		}
	}
	return ops
}

// normalizeYAMLMap recursively converts map[interface{}]interface{} to map[string]interface{}.
// YAML unmarshaling produces interface{} keys for maps, which JSON doesn't support.
func normalizeYAMLMap(m map[string]interface{}) {
	for k, v := range m {
		m[k] = normalizeYAMLValue(v)
	}
}

func normalizeYAMLValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range val {
			newMap[key.(string)] = normalizeYAMLValue(value)
		}
		return newMap
	case map[string]interface{}:
		normalizeYAMLMap(val)
		return val
	case []interface{}:
		for i, item := range val {
			val[i] = normalizeYAMLValue(item)
		}
		return val
	default:
		return v
	}
}

// DeepClone creates a deep copy of a map (for spec cloning).
func DeepClone(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	clone := make(map[string]interface{}, len(m))
	for k, v := range m {
		clone[k] = deepCloneValue(v)
	}
	return clone
}

func deepCloneValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		return DeepClone(val)
	case []interface{}:
		clone := make([]interface{}, len(val))
		for i, item := range val {
			clone[i] = deepCloneValue(item)
		}
		return clone
	default:
		return v
	}
}
