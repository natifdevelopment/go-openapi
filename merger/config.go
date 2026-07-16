package merger

import (
	"fmt"
	"os"
	"strings"

	"github.com/natifdevelopment/go-openapi/oas"
	"gopkg.in/yaml.v3"
)

// Config is the declarative merge configuration loaded from services.yaml.
type Config struct {
	Version         string               `yaml:"version"`
	Gateway         GatewayConfig        `yaml:"gateway"`
	Services        []ServiceConfig      `yaml:"services"`
	SharedSchemas   []string             `yaml:"shared_schemas"`
	SecuritySchemes map[string]SecScheme `yaml:"security_schemes"`
	Validation      ValidationConfig     `yaml:"validation"`
	Output          OutputConfig         `yaml:"output"`
}

// GatewayConfig defines the aggregated spec metadata.
type GatewayConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	ServerURL   string `yaml:"server_url"`
}

// ServiceConfig defines a single service's merge configuration.
type ServiceConfig struct {
	Key             string        `yaml:"key"`
	Name            string        `yaml:"name"`
	SpecPath        string        `yaml:"spec_path"`
	Module          string        `yaml:"module"`
	TagPrefix       string        `yaml:"tag_prefix"`
	SchemaPrefix    string        `yaml:"schema_prefix"`
	PathRewrite     []PathRewrite `yaml:"path_rewrite"`
	DefaultSecurity string        `yaml:"default_security"`
	PublicPaths     []string      `yaml:"public_paths"`
	Required        bool          `yaml:"required"`
}

// PathRewrite defines a path prefix rewrite rule.
type PathRewrite struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// SecScheme is a simplified security scheme definition for YAML config.
type SecScheme struct {
	Type         string `yaml:"type"`
	Scheme       string `yaml:"scheme"`
	In           string `yaml:"in"`
	Name         string `yaml:"name"`
	BearerFormat string `yaml:"bearerFormat"`
	Description  string `yaml:"description"`
}

// ValidationConfig defines which validation rules to enforce.
type ValidationConfig struct {
	FailOn []string `yaml:"fail_on"`
	WarnOn []string `yaml:"warn_on"`
	Linter string   `yaml:"linter"`
}

// OutputConfig defines the output file paths and format.
type OutputConfig struct {
	Format     string `yaml:"format"`
	Path       string `yaml:"path"`
	YAMLPath   string `yaml:"yaml_path"`
	ReportPath string `yaml:"report_path"`
	Pretty     bool   `yaml:"pretty"`
}

// LoadConfig reads and parses a services.yaml config file.
// Environment variables in the format ${VAR_NAME} are expanded.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expand environment variables
	expanded := expandEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	for i := range cfg.Services {
		if cfg.Services[i].Required == false && cfg.Services[i].SpecPath != "" {
			// required defaults to true if spec_path is set
			cfg.Services[i].Required = true
		}
	}
	if cfg.Output.Path == "" {
		cfg.Output.Path = "docs/openapi-aggregated.json"
	}
	if cfg.Output.Format == "" {
		cfg.Output.Format = "json"
	}
	cfg.Output.Pretty = true

	return &cfg, nil
}

// GetService returns the ServiceConfig for a given key.
func (c *Config) GetService(key string) *ServiceConfig {
	for i := range c.Services {
		if c.Services[i].Key == key {
			return &c.Services[i]
		}
	}
	return nil
}

// SharedSchemaSet returns the shared schema names as a set for quick lookup.
func (c *Config) SharedSchemaSet() map[string]bool {
	set := make(map[string]bool)
	for _, s := range c.SharedSchemas {
		set[s] = true
	}
	// Always include the standard shared schemas
	for _, s := range []string{"StandardResponse", "StandardPagination", "StandardError", "ErrorDetail", "ValidationError"} {
		set[s] = true
	}
	return set
}

// expandEnvVars replaces ${VAR_NAME} patterns with environment variable values.
func expandEnvVars(s string) string {
	return os.Expand(s, func(key string) string {
		// Handle ${VAR:-default} syntax
		if idx := strings.Index(key, ":-"); idx >= 0 {
			varName := key[:idx]
			defaultVal := key[idx+2:]
			val := os.Getenv(varName)
			if val == "" {
				return defaultVal
			}
			return val
		}
		return os.Getenv(key)
	})
}

// LoadAll loads all service specs from their configured paths.
func (m *Merger) LoadAll() error {
	for _, svc := range m.Config.Services {
		if _, err := os.Stat(svc.SpecPath); os.IsNotExist(err) {
			if svc.Required {
				return m.loadError(svc, "spec file not found")
			}
			m.warnings = append(m.warnings, ValidationResult{
				Rule:     "missing_spec",
				Severity: "warning",
				Message:  fmt.Sprintf("service '%s' spec not found at %s, skipping", svc.Key, svc.SpecPath),
			})
			continue
		}

		spec, err := oas.LoadFromFile(svc.SpecPath)
		if err != nil {
			return m.loadError(svc, fmt.Sprintf("failed to parse: %v", err))
		}

		m.specs[svc.Key] = spec
	}
	return nil
}

func (m *Merger) loadError(svc ServiceConfig, detail string) error {
	return fmt.Errorf("service '%s' (%s): %s", svc.Key, svc.SpecPath, detail)
}
