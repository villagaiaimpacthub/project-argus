package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// JavaScriptPlugin implements language support for JavaScript
type JavaScriptPlugin struct {
	BaseLanguagePlugin
}

// NewJavaScriptPlugin creates a new JavaScript language plugin
func NewJavaScriptPlugin() *JavaScriptPlugin {
	return &JavaScriptPlugin{
		BaseLanguagePlugin: BaseLanguagePlugin{
			Name:        "javascript",
			Extensions:  []string{".js", ".jsx", ".mjs", ".cjs"},
			ConfigFiles: []string{"package.json", ".eslintrc.js", ".eslintrc.json", "babel.config.js", "webpack.config.js"},
			ErrorPatterns: []ErrorPattern{
				{
					Pattern:     `SyntaxError: .+`,
					Type:        "syntax",
					Severity:    "error",
					Language:    "javascript",
					FileRegex:   `at (.+):(\d+):(\d+)`,
					LineRegex:   `at .+:(\d+):\d+`,
					ColumnRegex: `at .+:\d+:(\d+)`,
				},
				{
					Pattern:     `TypeError: .+`,
					Type:        "runtime",
					Severity:    "error",
					Language:    "javascript",
					FileRegex:   `at (.+):(\d+):(\d+)`,
					LineRegex:   `at .+:(\d+):\d+`,
					ColumnRegex: `at .+:\d+:(\d+)`,
				},
				{
					Pattern:     `ReferenceError: .+`,
					Type:        "runtime",
					Severity:    "error",
					Language:    "javascript",
					FileRegex:   `at (.+):(\d+):(\d+)`,
					LineRegex:   `at .+:(\d+):\d+`,
					ColumnRegex: `at .+:\d+:(\d+)`,
				},
				{
					Pattern:     `Module not found: .+`,
					Type:        "build",
					Severity:    "error",
					Language:    "javascript",
					FileRegex:   `in (.+)`,
					LineRegex:   ``,
					ColumnRegex: ``,
				},
				{
					Pattern:     `Failed to compile`,
					Type:        "build",
					Severity:    "error",
					Language:    "javascript",
				},
				{
					Pattern:     `warning .+`,
					Type:        "lint",
					Severity:    "warning",
					Language:    "javascript",
					FileRegex:   `(.+):\d+:\d+`,
					LineRegex:   `.+:(\d+):\d+`,
					ColumnRegex: `.+:\d+:(\d+)`,
				},
			},
		},
	}
}

// AnalyzeErrors performs comprehensive error analysis for JavaScript projects
func (jsp *JavaScriptPlugin) AnalyzeErrors(projectPath string) ([]ErrorInfo, error) {
	var allErrors []ErrorInfo

	// 1. Run ESLint if available
	if eslintErrors, err := jsp.runESLint(projectPath); err == nil {
		allErrors = append(allErrors, eslintErrors...)
	}

	// 2. Check for syntax errors by attempting to parse JS files
	if syntaxErrors := jsp.checkSyntaxErrors(projectPath); len(syntaxErrors) > 0 {
		allErrors = append(allErrors, syntaxErrors...)
	}

	// 3. Check build outputs if build tools are present
	if buildErrors := jsp.checkBuildErrors(projectPath); len(buildErrors) > 0 {
		allErrors = append(allErrors, buildErrors...)
	}

	// 4. Scan log files for runtime errors
	if runtimeErrors := jsp.scanLogFiles(projectPath); len(runtimeErrors) > 0 {
		allErrors = append(allErrors, runtimeErrors...)
	}

	return allErrors, nil
}

// runESLint executes ESLint and parses the output
func (jsp *JavaScriptPlugin) runESLint(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("npx") {
		return nil, fmt.Errorf("npx not available")
	}

	// Check if ESLint is configured
	eslintConfigs := []string{".eslintrc.js", ".eslintrc.json", ".eslintrc.yml", "eslint.config.js"}
	hasESLint := false
	for _, config := range eslintConfigs {
		if _, err := os.Stat(filepath.Join(projectPath, config)); err == nil {
			hasESLint = true
			break
		}
	}

	if !hasESLint {
		return nil, fmt.Errorf("no ESLint configuration found")
	}

	// Run ESLint with JSON output
	output, err := runCommand("npx", []string{"eslint", ".", "--format", "json"}, projectPath, 30*time.Second)
	if err != nil && output == "" {
		return nil, err
	}

	return jsp.parseESLintOutput(output)
}

// parseESLintOutput parses ESLint JSON output
func (jsp *JavaScriptPlugin) parseESLintOutput(output string) ([]ErrorInfo, error) {
	var eslintResults []struct {
		FilePath string `json:"filePath"`
		Messages []struct {
			RuleID    string `json:"ruleId"`
			Severity  int    `json:"severity"`
			Message   string `json:"message"`
			Line      int    `json:"line"`
			Column    int    `json:"column"`
			NodeType  string `json:"nodeType"`
			MessageID string `json:"messageId"`
		} `json:"messages"`
	}

	if err := json.Unmarshal([]byte(output), &eslintResults); err != nil {
		// If JSON parsing fails, try to parse as text
		return parseErrorWithRegex(output, jsp.ErrorPatterns), nil
	}

	var errors []ErrorInfo
	for _, result := range eslintResults {
		for _, msg := range result.Messages {
			errors = append(errors, ErrorInfo{
				Source:    "eslint",
				File:      result.FilePath,
				Line:      msg.Line,
				Column:    msg.Column,
				Type:      "lint",
				Message:   msg.Message,
				Code:      msg.RuleID,
				Timestamp: time.Now(),
			})
		}
	}

	return errors, nil
}

// checkSyntaxErrors checks for JavaScript syntax errors
func (jsp *JavaScriptPlugin) checkSyntaxErrors(projectPath string) []ErrorInfo {
	if !checkCommandExists("node") {
		return nil
	}

	var errors []ErrorInfo
	jsFiles, err := findFilesWithExtensions(projectPath, jsp.Extensions)
	if err != nil {
		return nil
	}

	for _, file := range jsFiles {
		// Skip node_modules and other build directories
		if strings.Contains(file, "node_modules") || strings.Contains(file, "dist") || strings.Contains(file, "build") {
			continue
		}

		// Try to parse the file with Node.js
		output, err := runCommand("node", []string{"--check", file}, projectPath, 5*time.Second)
		if err != nil && output != "" {
			// Parse Node.js syntax error output
			if syntaxError := jsp.parseNodeSyntaxError(output, file); syntaxError != nil {
				errors = append(errors, *syntaxError)
			}
		}
	}

	return errors
}

// parseNodeSyntaxError parses Node.js syntax error output
func (jsp *JavaScriptPlugin) parseNodeSyntaxError(output, file string) *ErrorInfo {
	// Node.js syntax error format: "SyntaxError: Unexpected token ... at line:column"
	syntaxErrorRegex := regexp.MustCompile(`SyntaxError: (.+)`)
	locationRegex := regexp.MustCompile(`at .*:(\d+):(\d+)`)

	var message string
	var line, column int

	if matches := syntaxErrorRegex.FindStringSubmatch(output); len(matches) > 1 {
		message = matches[1]
	}

	if matches := locationRegex.FindStringSubmatch(output); len(matches) > 2 {
		line, _ = strconv.Atoi(matches[1])
		column, _ = strconv.Atoi(matches[2])
	}

	if message != "" {
		relPath, _ := filepath.Rel(filepath.Dir(file), file)
		return &ErrorInfo{
			Source:    "node",
			File:      relPath,
			Line:      line,
			Column:    column,
			Type:      "syntax",
			Message:   message,
			Timestamp: time.Now(),
		}
	}

	return nil
}

// checkBuildErrors checks for build-related errors
func (jsp *JavaScriptPlugin) checkBuildErrors(projectPath string) []ErrorInfo {
	var errors []ErrorInfo

	// Check package.json for build scripts
	packageJSONPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err != nil {
		return errors
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}

	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return errors
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return errors
	}

	// Try running build command if it exists
	if _, exists := pkg.Scripts["build"]; exists && checkCommandExists("npm") {
		output, err := runCommand("npm", []string{"run", "build"}, projectPath, 60*time.Second)
		if err != nil && output != "" {
			buildErrors := parseErrorWithRegex(output, jsp.ErrorPatterns)
			errors = append(errors, buildErrors...)
		}
	}

	return errors
}

// scanLogFiles scans for JavaScript runtime errors in log files
func (jsp *JavaScriptPlugin) scanLogFiles(projectPath string) []ErrorInfo {
	var errors []ErrorInfo

	// Common log file patterns
	logPatterns := []string{
		"*.log",
		"logs/*.log",
		"log/*.log",
		".next/*.log",
		"npm-debug.log*",
		"yarn-debug.log*",
		"yarn-error.log*",
	}

	for _, pattern := range logPatterns {
		matches, _ := filepath.Glob(filepath.Join(projectPath, pattern))
		for _, logFile := range matches {
			logErrors := scanFileForPatterns(logFile, jsp.ErrorPatterns)
			errors = append(errors, logErrors...)
		}
	}

	return errors
}

// GetDependencies analyzes JavaScript dependencies
func (jsp *JavaScriptPlugin) GetDependencies(projectPath string) ([]DependencyInfo, error) {
	packageJSONPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err != nil {
		return nil, fmt.Errorf("package.json not found")
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, err
	}

	var deps []DependencyInfo

	// Add production dependencies
	for name, version := range pkg.Dependencies {
		deps = append(deps, DependencyInfo{
			Name:    name,
			Version: version,
			Type:    "direct",
			Source:  "package.json",
		})
	}

	// Add development dependencies
	for name, version := range pkg.DevDependencies {
		deps = append(deps, DependencyInfo{
			Name:    name,
			Version: version,
			Type:    "dev",
			Source:  "package.json",
		})
	}

	return deps, nil
}

// FindServices discovers running JavaScript services
func (jsp *JavaScriptPlugin) FindServices(projectPath string) ([]ServiceInfo, error) {
	var services []ServiceInfo

	// Check for common JavaScript development ports
	defaultPorts := []int{3000, 3001, 8080, 8000, 9000, 5000, 4200}
	runningServices := detectRunningProcesses(projectPath, defaultPorts)
	services = append(services, runningServices...)

	// Analyze package.json for service information
	packageJSONPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err == nil {
		if pkgServices := jsp.analyzePackageJSONServices(packageJSONPath); len(pkgServices) > 0 {
			services = append(services, pkgServices...)
		}
	}

	return services, nil
}

// analyzePackageJSONServices extracts service information from package.json
func (jsp *JavaScriptPlugin) analyzePackageJSONServices(packageJSONPath string) []ServiceInfo {
	var services []ServiceInfo

	var pkg struct {
		Name    string            `json:"name"`
		Scripts map[string]string `json:"scripts"`
		Dependencies map[string]string `json:"dependencies"`
	}

	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return services
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return services
	}

	// Detect framework and default port
	framework := "unknown"
	defaultPort := 3000

	// Check dependencies for framework detection
	if _, hasReact := pkg.Dependencies["react"]; hasReact {
		framework = "react"
	} else if _, hasVue := pkg.Dependencies["vue"]; hasVue {
		framework = "vue"
		defaultPort = 8080
	} else if _, hasAngular := pkg.Dependencies["@angular/core"]; hasAngular {
		framework = "angular"
		defaultPort = 4200
	} else if _, hasExpress := pkg.Dependencies["express"]; hasExpress {
		framework = "express"
	} else if _, hasNext := pkg.Dependencies["next"]; hasNext {
		framework = "next"
	}

	// Check if service is likely running
	status := "stopped"
	if isPortInUse(defaultPort) {
		status = "running"
	}

	service := ServiceInfo{
		ID:        fmt.Sprintf("js-%s", pkg.Name),
		Name:      pkg.Name,
		Language:  "javascript",
		Framework: framework,
		Port:      defaultPort,
		Status:    status,
		ConfigFile: "package.json",
	}

	// Add start command if available
	if startCmd, exists := pkg.Scripts["start"]; exists {
		service.StartCommand = fmt.Sprintf("npm run start (%s)", startCmd)
	} else if devCmd, exists := pkg.Scripts["dev"]; exists {
		service.StartCommand = fmt.Sprintf("npm run dev (%s)", devCmd)
	}

	services = append(services, service)
	return services
}

// RunLinter executes the linter for JavaScript
func (jsp *JavaScriptPlugin) RunLinter(projectPath string) ([]ErrorInfo, error) {
	return jsp.runESLint(projectPath)
}

// RunTests executes tests for JavaScript projects
func (jsp *JavaScriptPlugin) RunTests(projectPath string) (*TestResults, error) {
	packageJSONPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err != nil {
		return nil, fmt.Errorf("package.json not found")
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}

	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, err
	}

	// Check if test script exists
	testCmd, hasTest := pkg.Scripts["test"]
	if !hasTest {
		return nil, fmt.Errorf("no test script found in package.json")
	}

	// Run tests
	output, err := runCommand("npm", []string{"test"}, projectPath, 120*time.Second)
	
	// Parse test results (this is a simplified parser - real implementation would be more robust)
	results := &TestResults{
		LastRun: time.Now(),
	}

	// Try to parse Jest output format
	if strings.Contains(testCmd, "jest") || strings.Contains(output, "jest") {
		results = jsp.parseJestOutput(output)
	}

	return results, err
}

// parseJestOutput parses Jest test output
func (jsp *JavaScriptPlugin) parseJestOutput(output string) *TestResults {
	results := &TestResults{
		LastRun: time.Now(),
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Look for Jest summary line: "Tests: X failed, Y passed, Z total"
		if strings.Contains(line, "Tests:") && strings.Contains(line, "total") {
			// Parse the summary
			parts := strings.Split(line, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.Contains(part, "failed") {
					fmt.Sscanf(part, "%d failed", &results.FailedTests)
				} else if strings.Contains(part, "passed") {
					fmt.Sscanf(part, "%d passed", &results.PassedTests)
				} else if strings.Contains(part, "total") {
					fmt.Sscanf(part, "%d total", &results.TotalTests)
				}
			}
		}
	}

	return results
}

// ParseBuildOutput parses build output for errors
func (jsp *JavaScriptPlugin) ParseBuildOutput(output string) ([]ErrorInfo, error) {
	return parseErrorWithRegex(output, jsp.ErrorPatterns), nil
}