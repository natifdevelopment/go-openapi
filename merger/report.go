package merger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/natifdevelopment/go-openapi/oas"
)

// MergeReport is the structured output report from a merge run.
type MergeReport struct {
	Timestamp     string                 `json:"timestamp"`
	MergerVersion string                 `json:"merger_version"`
	ConfigVersion string                 `json:"config_version"`
	Summary       ReportSummary          `json:"summary"`
	PerService    map[string]ServiceStat `json:"per_service"`
	SharedSchemas []string               `json:"shared_schemas_deduplicated"`
	Conflicts     []Conflict             `json:"conflicts"`
	Validation    ReportValidation       `json:"validation"`
	OutputFiles   []string               `json:"output_files"`
}

type ReportSummary struct {
	ServicesMerged  int `json:"services_merged"`
	TotalPaths      int `json:"total_paths"`
	TotalSchemas    int `json:"total_schemas"`
	TotalTags       int `json:"total_tags"`
	TotalOperations int `json:"total_operations"`
}

type ServiceStat struct {
	Paths        int    `json:"paths"`
	Schemas      int    `json:"schemas"`
	Operations   int    `json:"operations"`
	PathRewrites int    `json:"path_rewrites"`
	SchemaPrefix string `json:"schema_prefix"`
}

type ReportValidation struct {
	Errors   int                `json:"errors"`
	Warnings int                `json:"warnings"`
	Details  []ValidationResult `json:"details"`
}

// BuildReport generates the merge report from the merged spec.
func (m *Merger) BuildReport(spec *oas.Spec) *MergeReport {
	report := &MergeReport{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		MergerVersion: "1.0.0",
		ConfigVersion: m.Config.Version,
		PerService:    make(map[string]ServiceStat),
		SharedSchemas: []string{"StandardResponse", "StandardPagination", "StandardError", "ErrorDetail", "ValidationError"},
		OutputFiles:   []string{m.Config.Output.Path},
	}

	if m.Config.Output.YAMLPath != "" {
		report.OutputFiles = append(report.OutputFiles, m.Config.Output.YAMLPath)
	}
	if m.Config.Output.ReportPath != "" {
		report.OutputFiles = append(report.OutputFiles, m.Config.Output.ReportPath)
	}

	paths := spec.Paths()
	report.Summary.TotalPaths = len(paths)
	totalOps := 0
	for _, pi := range paths {
		if pathItem, ok := pi.(map[string]interface{}); ok {
			for _, op := range oas.GetOperations(pathItem) {
				if op != nil {
					totalOps++
				}
			}
		}
	}
	report.Summary.TotalOperations = totalOps
	report.Summary.TotalSchemas = len(spec.Schemas())
	report.Summary.TotalTags = len(spec.Tags())
	report.Summary.ServicesMerged = len(m.specs)

	for svcKey, spec := range m.specs {
		svc := m.Config.GetService(svcKey)
		stat := ServiceStat{SchemaPrefix: svc.SchemaPrefix}
		svcPaths := spec.Paths()
		stat.Paths = len(svcPaths)
		for _, pi := range svcPaths {
			if pathItem, ok := pi.(map[string]interface{}); ok {
				for _, op := range oas.GetOperations(pathItem) {
					if op != nil {
						stat.Operations++
					}
				}
			}
		}
		stat.Schemas = len(spec.Schemas())
		stat.PathRewrites = len(svc.PathRewrite)
		report.PerService[svcKey] = stat
	}

	report.Conflicts = m.conflicts
	report.Validation.Errors = len(m.errors)
	report.Validation.Warnings = len(m.warnings)
	report.Validation.Details = append(report.Validation.Details, m.errors...)
	report.Validation.Details = append(report.Validation.Details, m.warnings...)

	return report
}

// PrintReport prints a summary of the merge to stdout.
func (m *Merger) PrintReport() {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  BBO OpenAPI Merger — Report")
	fmt.Println("═══════════════════════════════════════════════════════════════")

	fmt.Printf("\n  Services merged: %d\n", len(m.specs))
	for svcKey := range m.specs {
		svc := m.Config.GetService(svcKey)
		fmt.Printf("    • %-20s → %s (prefix: %s)\n", svcKey, svc.Name, svc.SchemaPrefix)
	}

	totalPaths, totalOps := 0, 0
	for _, spec := range m.specs {
		for _, pi := range spec.Paths() {
			totalPaths++
			if pathItem, ok := pi.(map[string]interface{}); ok {
				for _, op := range oas.GetOperations(pathItem) {
					if op != nil {
						totalOps++
					}
				}
			}
		}
	}
	fmt.Printf("\n  Total paths:      %d\n", totalPaths)
	fmt.Printf("  Total operations: %d\n", totalOps)

	fmt.Printf("\n  Conflicts: %d\n", len(m.conflicts))
	for _, c := range m.conflicts {
		fmt.Printf("    ✗ [%s] %s: %s\n", c.Type, c.Service, c.Detail)
	}

	fmt.Printf("\n  Validation:\n")
	fmt.Printf("    Errors:   %d\n", len(m.errors))
	for _, e := range m.errors {
		pathInfo := ""
		if e.Path != "" {
			pathInfo = fmt.Sprintf(" [%s %s]", e.Method, e.Path)
		}
		fmt.Printf("    ✗ [%s]%s %s\n", e.Rule, pathInfo, e.Message)
	}
	fmt.Printf("    Warnings: %d\n", len(m.warnings))
	for _, w := range m.warnings {
		pathInfo := ""
		if w.Path != "" {
			pathInfo = fmt.Sprintf(" [%s %s]", w.Method, w.Path)
		}
		fmt.Printf("    ⚠ [%s]%s %s\n", w.Rule, pathInfo, w.Message)
	}

	fmt.Println("\n═══════════════════════════════════════════════════════════════")
}

// WriteReport writes the merge report to a JSON file.
func (m *Merger) WriteReport(spec *oas.Spec) error {
	if m.Config.Output.ReportPath == "" {
		return nil
	}
	report := m.BuildReport(spec)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.Config.Output.ReportPath, data, 0644)
}
