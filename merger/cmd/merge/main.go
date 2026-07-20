// Package main is the CLI entry point for the bbo-openapi-merger.
//
// Usage:
//
//	bbo-openapi-merger -config services.yaml -o docs/openapi-aggregated.json
//	bbo-openapi-merger -config services.yaml -validate-only
//	bbo-openapi-merger -config services.yaml -o out.json -yaml out.yaml -report report.json
//
// Environment variables:
//
//	GATEWAY_URL  - Gateway base URL (default: http://localhost:8888)
//	GIT_TAG      - Version string (default: 1.0.0)
package main

import (
	"log"
	"flag"
	"fmt"
	"os"

	"github.com/natifdevelopment/go-openapi/merger"
)

func main() {
	configPath := flag.String("config", "services.yaml", "Path to merge config (services.yaml)")
	outputPath := flag.String("o", "docs/openapi-aggregated.json", "Output JSON path")
	yamlPath := flag.String("yaml", "", "Output YAML path (empty = no YAML output)")
	reportPath := flag.String("report", "docs/merge-report.json", "Merge report path")
	validateOnly := flag.Bool("validate-only", false, "Validate only, don't write output")
	verbose := flag.Bool("v", false, "Verbose output")
	strict := flag.Bool("strict", false, "Fail on warnings too (not just errors)")
	flag.Parse()

	m, err := merger.New(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Override config with CLI flags
	if *outputPath != "" {
		m.Config.Output.Path = *outputPath
	}
	if *yamlPath != "" {
		m.Config.Output.YAMLPath = *yamlPath
	}
	if *reportPath != "" {
		m.Config.Output.ReportPath = *reportPath
	}

	if *verbose {
		fmt.Printf("Config: %s\n", *configPath)
		fmt.Printf("Output: %s\n", m.Config.Output.Path)
		if m.Config.Output.YAMLPath != "" {
			fmt.Printf("YAML:   %s\n", m.Config.Output.YAMLPath)
		}
		fmt.Printf("Report: %s\n", m.Config.Output.ReportPath)
		fmt.Printf("Services: %d\n", len(m.Config.Services))
		for _, svc := range m.Config.Services {
			fmt.Printf("  - %s (%s)\n", svc.Key, svc.SpecPath)
		}
		log.Println()
	}

	if *validateOnly {
		if err := m.ValidateOnly(); err != nil {
			fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
			os.Exit(1)
		}
		log.Println("Validation passed.")
		return
	}

	if err := m.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Merge failed: %v\n", err)
		os.Exit(1)
	}

	// Check strict mode
	if *strict && m.HasWarnings() {
		fmt.Fprintf(os.Stderr, "Strict mode: %d warnings treated as errors\n", m.WarningCount())
		os.Exit(1)
	}

	log.Println("Merge completed successfully.")
}
