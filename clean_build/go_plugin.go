package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GoPlugin implements language support for Go
type GoPlugin struct {
	BaseLanguagePlugin
}

// NewGoPlugin creates a new Go language plugin
func NewGoPlugin() *GoPlugin {
	return &GoPlugin{
		BaseLanguagePlugin: BaseLanguagePlugin{
			Name:        "go",
			Extensions:  []string{".go"},
			ConfigFiles: []string{"go.mod", "go.sum", "go.work"},
			ErrorPatterns: []ErrorPattern{
				{
					Pattern:     `(.+\.go):(\d+):(\d+): (.+)`,
					Type:        "compile",
					Severity:    "error",
					Language:    "go",
					FileRegex:   `(.+\.go):(\d+):(\d+):`,
					LineRegex:   `.+\.go:(\d+):\d+:`,
					ColumnRegex: `.+\.go:\d+:(\d+):`,
				},
				{
					Pattern:  `panic: .+`,
					Type:     "runtime",
					Severity: "error",
					Language: "go",
				},
				{
					Pattern:  `cannot find package .+`,
					Type:     "import",
					Severity: "error",
					Language: "go",
				},
				{
					Pattern:  `undefined: .+`,
					Type:     "compile",
					Severity: "error",
					Language: "go",
				},
			},
		},
	}
}

// AnalyzeErrors performs comprehensive error analysis for Go projects
func (gp *GoPlugin) AnalyzeErrors(projectPath string) ([]ErrorInfo, error) {
	var allErrors []ErrorInfo

	// 1. Run go build
	if buildErrors := gp.runGoBuild(projectPath); len(buildErrors) > 0 {
		allErrors = append(allErrors, buildErrors...)
	}

	// 2. Run go vet
	if vetErrors := gp.runGoVet(projectPath); len(vetErrors) > 0 {
		allErrors = append(allErrors, vetErrors...)
	}

	// 3. Run golangci-lint if available
	if lintErrors, err := gp.runGolangciLint(projectPath); err == nil {
		allErrors = append(allErrors, lintErrors...)
	}

	return allErrors, nil
}

// runGoBuild executes go build and parses errors
func (gp *GoPlugin) runGoBuild(projectPath string) []ErrorInfo {
	if !checkCommandExists("go") {
		return nil
	}

	output, err := runCommand("go", []string{"build", "./..."}, projectPath, 60*time.Second)
	if err != nil && output != "" {
		return parseErrorWithRegex(output, gp.ErrorPatterns)
	}

	return nil
}

// runGoVet executes go vet
func (gp *GoPlugin) runGoVet(projectPath string) []ErrorInfo {
	if !checkCommandExists("go") {
		return nil
	}

	output, err := runCommand("go", []string{"vet", "./..."}, projectPath, 30*time.Second)
	if err != nil && output != "" {
		return parseErrorWithRegex(output, gp.ErrorPatterns)
	}

	return nil
}

// runGolangciLint executes golangci-lint
func (gp *GoPlugin) runGolangciLint(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("golangci-lint") {
		return nil, fmt.Errorf("golangci-lint not available")
	}

	output, err := runCommand("golangci-lint", []string{"run"}, projectPath, 60*time.Second)
	if err != nil && output != "" {
		return parseErrorWithRegex(output, gp.ErrorPatterns), nil
	}

	return nil, err
}

// GetDependencies analyzes Go dependencies from go.mod
func (gp *GoPlugin) GetDependencies(projectPath string) ([]DependencyInfo, error) {
	goModPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		return nil, fmt.Errorf("go.mod not found")
	}

	var deps []DependencyInfo
	file, err := os.Open(goModPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Parse require statements
		if strings.HasPrefix(line, "require ") || (strings.Contains(line, " v") && !strings.HasPrefix(line, "module")) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := parts[0]
				if name == "require" && len(parts) >= 3 {
					name = parts[1]
				}

				version := "latest"
				for _, part := range parts {
					if strings.HasPrefix(part, "v") {
						version = part
						break
					}
				}

				deps = append(deps, DependencyInfo{
					Name:    name,
					Version: version,
					Type:    "direct",
					Source:  "go.mod",
				})
			}
		}
	}

	return deps, nil
}

// FindServices discovers running Go services
func (gp *GoPlugin) FindServices(projectPath string) ([]ServiceInfo, error) {
	var services []ServiceInfo

	defaultPorts := []int{8080, 8000, 3000, 9000}
	runningServices := detectRunningProcesses(projectPath, defaultPorts)
	services = append(services, runningServices...)

	// Detect Go web frameworks
	if framework := gp.detectGoFramework(projectPath); framework != "" {
		service := ServiceInfo{
			ID:        fmt.Sprintf("go-%s", framework),
			Name:      fmt.Sprintf("Go %s Service", framework),
			Language:  "go",
			Framework: framework,
			Port:      8080,
			Status:    "stopped",
		}

		if isPortInUse(8080) {
			service.Status = "running"
		}

		services = append(services, service)
	}

	return services, nil
}

// detectGoFramework detects Go web frameworks
func (gp *GoPlugin) detectGoFramework(projectPath string) string {
	goFiles, _ := findFilesWithExtensions(projectPath, gp.Extensions)

	for _, file := range goFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "github.com/gin-gonic/gin") {
			return "gin"
		} else if strings.Contains(contentStr, "github.com/gofiber/fiber") {
			return "fiber"
		} else if strings.Contains(contentStr, "github.com/labstack/echo") {
			return "echo"
		}
	}

	return ""
}

// RunLinter executes Go linters
func (gp *GoPlugin) RunLinter(projectPath string) ([]ErrorInfo, error) {
	return gp.runGolangciLint(projectPath)
}

// RunTests executes Go tests
func (gp *GoPlugin) RunTests(projectPath string) (*TestResults, error) {
	if !checkCommandExists("go") {
		return nil, fmt.Errorf("go not available")
	}

	output, err := runCommand("go", []string{"test", "-v", "./..."}, projectPath, 120*time.Second)

	results := &TestResults{
		LastRun: time.Now(),
	}

	// Parse go test output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "PASS") {
			results.PassedTests++
		} else if strings.Contains(line, "FAIL") {
			results.FailedTests++
		}
	}

	results.TotalTests = results.PassedTests + results.FailedTests

	return results, err
}

// ParseBuildOutput parses Go build output for errors
func (gp *GoPlugin) ParseBuildOutput(output string) ([]ErrorInfo, error) {
	return parseErrorWithRegex(output, gp.ErrorPatterns), nil
}
