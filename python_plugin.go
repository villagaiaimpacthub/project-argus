package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PythonPlugin implements language support for Python
type PythonPlugin struct {
	BaseLanguagePlugin
}

// NewPythonPlugin creates a new Python language plugin
func NewPythonPlugin() *PythonPlugin {
	return &PythonPlugin{
		BaseLanguagePlugin: BaseLanguagePlugin{
			Name:        "python",
			Extensions:  []string{".py", ".pyw", ".pyx", ".pyi"},
			ConfigFiles: []string{"requirements.txt", "pyproject.toml", "setup.py", "setup.cfg", "Pipfile", "poetry.lock", "tox.ini"},
			ErrorPatterns: []ErrorPattern{
				{
					Pattern:     `SyntaxError: .+`,
					Type:        "syntax",
					Severity:    "error",
					Language:    "python",
					FileRegex:   `File "(.+)", line (\d+)`,
					LineRegex:   `line (\d+)`,
				},
				{
					Pattern:     `IndentationError: .+`,
					Type:        "syntax",
					Severity:    "error",
					Language:    "python",
					FileRegex:   `File "(.+)", line (\d+)`,
					LineRegex:   `line (\d+)`,
				},
				{
					Pattern:     `NameError: .+`,
					Type:        "runtime",
					Severity:    "error",
					Language:    "python",
					FileRegex:   `File "(.+)", line (\d+)`,
					LineRegex:   `line (\d+)`,
				},
				{
					Pattern:     `TypeError: .+`,
					Type:        "runtime",
					Severity:    "error",
					Language:    "python",
					FileRegex:   `File "(.+)", line (\d+)`,
					LineRegex:   `line (\d+)`,
				},
				{
					Pattern:     `AttributeError: .+`,
					Type:        "runtime",
					Severity:    "error",
					Language:    "python",
					FileRegex:   `File "(.+)", line (\d+)`,
					LineRegex:   `line (\d+)`,
				},
				{
					Pattern:     `ImportError: .+`,
					Type:        "import",
					Severity:    "error",
					Language:    "python",
					FileRegex:   `File "(.+)", line (\d+)`,
					LineRegex:   `line (\d+)`,
				},
				{
					Pattern:     `ModuleNotFoundError: .+`,
					Type:        "import",
					Severity:    "error",
					Language:    "python",
					FileRegex:   `File "(.+)", line (\d+)`,
					LineRegex:   `line (\d+)`,
				},
				{
					Pattern:     `FAILED .+ - .+`,
					Type:        "test",
					Severity:    "error",
					Language:    "python",
				},
				{
					Pattern:     `ERROR .+ - .+`,
					Type:        "test",
					Severity:    "error",
					Language:    "python",
				},
			},
		},
	}
}

// AnalyzeErrors performs comprehensive error analysis for Python projects
func (pp *PythonPlugin) AnalyzeErrors(projectPath string) ([]ErrorInfo, error) {
	var allErrors []ErrorInfo

	// 1. Run syntax checks
	if syntaxErrors := pp.checkSyntaxErrors(projectPath); len(syntaxErrors) > 0 {
		allErrors = append(allErrors, syntaxErrors...)
	}

	// 2. Run linting tools (pylint, flake8, etc.)
	if lintErrors, err := pp.runPythonLinters(projectPath); err == nil {
		allErrors = append(allErrors, lintErrors...)
	}

	// 3. Run type checking with mypy if available
	if typeErrors, err := pp.runMyPy(projectPath); err == nil {
		allErrors = append(allErrors, typeErrors...)
	}

	// 4. Scan log files for runtime errors
	if runtimeErrors := pp.scanLogFiles(projectPath); len(runtimeErrors) > 0 {
		allErrors = append(allErrors, runtimeErrors...)
	}

	// 5. Check import issues
	if importErrors := pp.checkImportIssues(projectPath); len(importErrors) > 0 {
		allErrors = append(allErrors, importErrors...)
	}

	return allErrors, nil
}

// checkSyntaxErrors checks for Python syntax errors
func (pp *PythonPlugin) checkSyntaxErrors(projectPath string) []ErrorInfo {
	var errors []ErrorInfo

	if !checkCommandExists("python") && !checkCommandExists("python3") {
		return errors
	}

	pythonCmd := "python"
	if checkCommandExists("python3") {
		pythonCmd = "python3"
	}

	pyFiles, err := findFilesWithExtensions(projectPath, pp.Extensions)
	if err != nil {
		return errors
	}

	for _, file := range pyFiles {
		// Skip virtual environments and cache directories
		if strings.Contains(file, "venv") || strings.Contains(file, "__pycache__") || 
		   strings.Contains(file, ".env") || strings.Contains(file, "site-packages") {
			continue
		}

		// Check syntax with python -m py_compile
		output, err := runCommand(pythonCmd, []string{"-m", "py_compile", file}, projectPath, 10*time.Second)
		if err != nil && output != "" {
			if syntaxError := pp.parsePythonSyntaxError(output, file); syntaxError != nil {
				errors = append(errors, *syntaxError)
			}
		}
	}

	return errors
}

// parsePythonSyntaxError parses Python syntax error output
func (pp *PythonPlugin) parsePythonSyntaxError(output, file string) *ErrorInfo {
	// Python syntax error format: 'File "filename", line X\n    SyntaxError: ...'
	fileRegex := regexp.MustCompile(`File "(.+)", line (\d+)`)
	errorRegex := regexp.MustCompile(`(SyntaxError|IndentationError): (.+)`)

	var fileName string
	var line int
	var errorType, message string

	lines := strings.Split(output, "\n")
	for i, lineText := range lines {
		if matches := fileRegex.FindStringSubmatch(lineText); len(matches) == 3 {
			fileName = matches[1]
			line, _ = strconv.Atoi(matches[2])
		}
		
		if matches := errorRegex.FindStringSubmatch(lineText); len(matches) == 3 {
			errorType = matches[1]
			message = matches[2]
			break
		}
		
		// Sometimes the error is on the next line
		if i+1 < len(lines) {
			if matches := errorRegex.FindStringSubmatch(lines[i+1]); len(matches) == 3 {
				errorType = matches[1]
				message = matches[2]
				break
			}
		}
	}

	if fileName != "" && message != "" {
		relPath, _ := filepath.Rel(filepath.Dir(file), fileName)
		return &ErrorInfo{
			Source:    "python",
			File:      relPath,
			Line:      line,
			Type:      "syntax",
			Message:   fmt.Sprintf("%s: %s", errorType, message),
			Timestamp: time.Now(),
		}
	}

	return nil
}

// runPythonLinters runs various Python linting tools
func (pp *PythonPlugin) runPythonLinters(projectPath string) ([]ErrorInfo, error) {
	var allErrors []ErrorInfo

	// Try pylint
	if pylintErrors, err := pp.runPyLint(projectPath); err == nil {
		allErrors = append(allErrors, pylintErrors...)
	}

	// Try flake8
	if flake8Errors, err := pp.runFlake8(projectPath); err == nil {
		allErrors = append(allErrors, flake8Errors...)
	}

	// Try ruff (modern fast linter)
	if ruffErrors, err := pp.runRuff(projectPath); err == nil {
		allErrors = append(allErrors, ruffErrors...)
	}

	return allErrors, nil
}

// runPyLint executes pylint
func (pp *PythonPlugin) runPyLint(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("pylint") {
		return nil, fmt.Errorf("pylint not available")
	}

	// Run pylint with JSON output
	output, err := runCommand("pylint", []string{"--output-format=json", "."}, projectPath, 60*time.Second)
	if err != nil && output == "" {
		return nil, err
	}

	return pp.parsePyLintOutput(output)
}

// parsePyLintOutput parses pylint JSON output
func (pp *PythonPlugin) parsePyLintOutput(output string) ([]ErrorInfo, error) {
	// Pylint JSON format parsing would go here
	// For now, use regex parsing as fallback
	return parseErrorWithRegex(output, pp.ErrorPatterns), nil
}

// runFlake8 executes flake8
func (pp *PythonPlugin) runFlake8(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("flake8") {
		return nil, fmt.Errorf("flake8 not available")
	}

	output, err := runCommand("flake8", []string{"."}, projectPath, 30*time.Second)
	if err != nil && output == "" {
		return nil, err
	}

	return pp.parseFlake8Output(output), nil
}

// parseFlake8Output parses flake8 output
func (pp *PythonPlugin) parseFlake8Output(output string) []ErrorInfo {
	var errors []ErrorInfo
	lines := strings.Split(output, "\n")

	// Flake8 format: filename:line:column: error_code error_message
	flake8Regex := regexp.MustCompile(`(.+):(\d+):(\d+): (\w+) (.+)`)

	for _, line := range lines {
		if matches := flake8Regex.FindStringSubmatch(line); len(matches) == 6 {
			lineNum, _ := strconv.Atoi(matches[2])
			column, _ := strconv.Atoi(matches[3])
			code := matches[4]
			message := matches[5]

			errorType := "lint"

			errors = append(errors, ErrorInfo{
				Source:    "flake8",
				File:      matches[1],
				Line:      lineNum,
				Column:    column,
				Type:      errorType,
				Message:   message,
				Code:      code,
				Timestamp: time.Now(),
			})
		}
	}

	return errors
}

// runRuff executes ruff linter
func (pp *PythonPlugin) runRuff(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("ruff") {
		return nil, fmt.Errorf("ruff not available")
	}

	output, err := runCommand("ruff", []string{"check", ".", "--output-format=json"}, projectPath, 30*time.Second)
	if err != nil && output == "" {
		return nil, err
	}

	// Ruff has JSON output format, but for simplicity using regex for now
	return parseErrorWithRegex(output, pp.ErrorPatterns), nil
}

// runMyPy executes mypy for type checking
func (pp *PythonPlugin) runMyPy(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("mypy") {
		return nil, fmt.Errorf("mypy not available")
	}

	output, err := runCommand("mypy", []string{"."}, projectPath, 60*time.Second)
	if err != nil && output == "" {
		return nil, err
	}

	return pp.parseMyPyOutput(output), nil
}

// parseMyPyOutput parses mypy output
func (pp *PythonPlugin) parseMyPyOutput(output string) []ErrorInfo {
	var errors []ErrorInfo
	lines := strings.Split(output, "\n")

	// MyPy format: filename:line: error: message
	mypyRegex := regexp.MustCompile(`(.+):(\d+): (error|warning|note): (.+)`)

	for _, line := range lines {
		if matches := mypyRegex.FindStringSubmatch(line); len(matches) == 5 {
			lineNum, _ := strconv.Atoi(matches[2])
			severity := matches[3]
			message := matches[4]

			errorType := "type"
			if severity == "note" {
				severity = "info"
			}

			errors = append(errors, ErrorInfo{
				Source:    "mypy",
				File:      matches[1],
				Line:      lineNum,
				Type:      errorType,
				Message:   message,
				Timestamp: time.Now(),
			})
		}
	}

	return errors
}

// scanLogFiles scans for Python runtime errors in log files
func (pp *PythonPlugin) scanLogFiles(projectPath string) []ErrorInfo {
	var errors []ErrorInfo

	logPatterns := []string{
		"*.log",
		"logs/*.log",
		"log/*.log",
		"django.log",
		"app.log",
		"error.log",
	}

	for _, pattern := range logPatterns {
		matches, _ := filepath.Glob(filepath.Join(projectPath, pattern))
		for _, logFile := range matches {
			logErrors := scanFileForPatterns(logFile, pp.ErrorPatterns)
			errors = append(errors, logErrors...)
		}
	}

	return errors
}

// checkImportIssues checks for import-related issues
func (pp *PythonPlugin) checkImportIssues(projectPath string) []ErrorInfo {
	var errors []ErrorInfo

	pyFiles, err := findFilesWithExtensions(projectPath, pp.Extensions)
	if err != nil {
		return errors
	}

	for _, file := range pyFiles {
		if strings.Contains(file, "venv") || strings.Contains(file, "__pycache__") {
			continue
		}

		importErrors := pp.analyzeImports(file)
		errors = append(errors, importErrors...)
	}

	return errors
}

// analyzeImports analyzes import statements in Python files
func (pp *PythonPlugin) analyzeImports(filePath string) []ErrorInfo {
	var errors []ErrorInfo

	file, err := os.Open(filePath)
	if err != nil {
		return errors
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Check for import statements
		if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
			// Basic import validation - this could be much more sophisticated
			if strings.Contains(line, "..") && !strings.Contains(line, "...") {
				// Potentially problematic relative import
				relPath, _ := filepath.Rel(filepath.Dir(filePath), filePath)
				errors = append(errors, ErrorInfo{
					Source:    "python",
					File:      relPath,
					Line:      lineNum,
					Type:      "import",
					Message:   "Potentially problematic relative import: " + line,
					Timestamp: time.Now(),
				})
			}
		}
	}

	return errors
}

// GetDependencies analyzes Python dependencies
func (pp *PythonPlugin) GetDependencies(projectPath string) ([]DependencyInfo, error) {
	var deps []DependencyInfo

	// Check requirements.txt
	reqPath := filepath.Join(projectPath, "requirements.txt")
	if reqDeps, err := pp.parseRequirementsTxt(reqPath); err == nil {
		deps = append(deps, reqDeps...)
	}

	// Check pyproject.toml
	pyprojectPath := filepath.Join(projectPath, "pyproject.toml")
	if pyprojectDeps, err := pp.parsePyProjectToml(pyprojectPath); err == nil {
		deps = append(deps, pyprojectDeps...)
	}

	// Check Pipfile
	pipfilePath := filepath.Join(projectPath, "Pipfile")
	if pipfileDeps, err := pp.parsePipfile(pipfilePath); err == nil {
		deps = append(deps, pipfileDeps...)
	}

	return deps, nil
}

// parseRequirementsTxt parses requirements.txt file
func (pp *PythonPlugin) parseRequirementsTxt(reqPath string) ([]DependencyInfo, error) {
	file, err := os.Open(reqPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []DependencyInfo
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse package==version format
		parts := regexp.MustCompile(`([^=<>!]+)[=<>!]+(.+)`).FindStringSubmatch(line)
		if len(parts) == 3 {
			deps = append(deps, DependencyInfo{
				Name:    strings.TrimSpace(parts[1]),
				Version: strings.TrimSpace(parts[2]),
				Type:    "direct",
				Source:  "requirements.txt",
			})
		} else {
			// Package without version
			deps = append(deps, DependencyInfo{
				Name:    line,
				Version: "latest",
				Type:    "direct",
				Source:  "requirements.txt",
			})
		}
	}

	return deps, nil
}

// parsePyProjectToml parses pyproject.toml file (simplified)
func (pp *PythonPlugin) parsePyProjectToml(pyprojectPath string) ([]DependencyInfo, error) {
	content, err := os.ReadFile(pyprojectPath)
	if err != nil {
		return nil, err
	}

	// This is a simplified parser - a real implementation would use a TOML library
	var deps []DependencyInfo
	lines := strings.Split(string(content), "\n")
	inDependencies := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "[tool.poetry.dependencies]" || line == "[project.dependencies]" {
			inDependencies = true
			continue
		}
		
		if strings.HasPrefix(line, "[") && inDependencies {
			inDependencies = false
			continue
		}
		
		if inDependencies && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[0])
				version := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				
				deps = append(deps, DependencyInfo{
					Name:    name,
					Version: version,
					Type:    "direct",
					Source:  "pyproject.toml",
				})
			}
		}
	}

	return deps, nil
}

// parsePipfile parses Pipfile (simplified)
func (pp *PythonPlugin) parsePipfile(pipfilePath string) ([]DependencyInfo, error) {
	content, err := os.ReadFile(pipfilePath)
	if err != nil {
		return nil, err
	}

	// Simplified Pipfile parsing
	var deps []DependencyInfo
	lines := strings.Split(string(content), "\n")
	inPackages := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "[packages]" {
			inPackages = true
			continue
		}
		
		if strings.HasPrefix(line, "[") && inPackages {
			inPackages = false
			continue
		}
		
		if inPackages && strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[0])
				version := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				
				deps = append(deps, DependencyInfo{
					Name:    name,
					Version: version,
					Type:    "direct",
					Source:  "Pipfile",
				})
			}
		}
	}

	return deps, nil
}

// FindServices discovers running Python services
func (pp *PythonPlugin) FindServices(projectPath string) ([]ServiceInfo, error) {
	var services []ServiceInfo

	// Check for common Python web framework ports
	defaultPorts := []int{8000, 5000, 8080, 9000, 8888}
	runningServices := detectRunningProcesses(projectPath, defaultPorts)
	services = append(services, runningServices...)

	// Detect framework from dependencies and files
	if framework := pp.detectPythonFramework(projectPath); framework != "" {
		service := ServiceInfo{
			ID:        fmt.Sprintf("python-%s", framework),
			Name:      fmt.Sprintf("Python %s Application", framework),
			Language:  "python",
			Framework: framework,
			Port:      pp.getFrameworkDefaultPort(framework),
			Status:    "stopped",
		}

		// Check if service is running
		if isPortInUse(service.Port) {
			service.Status = "running"
		}

		services = append(services, service)
	}

	return services, nil
}

// detectPythonFramework detects which Python framework is being used
func (pp *PythonPlugin) detectPythonFramework(projectPath string) string {
	// Check for Django
	if _, err := os.Stat(filepath.Join(projectPath, "manage.py")); err == nil {
		return "django"
	}

	// Check for Flask app files
	if pp.hasFlaskApp(projectPath) {
		return "flask"
	}

	// Check for FastAPI
	if pp.hasFastAPIApp(projectPath) {
		return "fastapi"
	}

	return ""
}

// hasFlaskApp checks for Flask application indicators
func (pp *PythonPlugin) hasFlaskApp(projectPath string) bool {
	pyFiles, _ := findFilesWithExtensions(projectPath, pp.Extensions)
	
	for _, file := range pyFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		
		contentStr := string(content)
		if strings.Contains(contentStr, "from flask import") || strings.Contains(contentStr, "Flask(__name__)") {
			return true
		}
	}
	
	return false
}

// hasFastAPIApp checks for FastAPI application indicators
func (pp *PythonPlugin) hasFastAPIApp(projectPath string) bool {
	pyFiles, _ := findFilesWithExtensions(projectPath, pp.Extensions)
	
	for _, file := range pyFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		
		contentStr := string(content)
		if strings.Contains(contentStr, "from fastapi import") || strings.Contains(contentStr, "FastAPI()") {
			return true
		}
	}
	
	return false
}

// getFrameworkDefaultPort returns the default port for a framework
func (pp *PythonPlugin) getFrameworkDefaultPort(framework string) int {
	switch framework {
	case "django":
		return 8000
	case "flask":
		return 5000
	case "fastapi":
		return 8000
	default:
		return 8000
	}
}

// RunLinter executes Python linters
func (pp *PythonPlugin) RunLinter(projectPath string) ([]ErrorInfo, error) {
	return pp.runPythonLinters(projectPath)
}

// RunTests executes Python tests
func (pp *PythonPlugin) RunTests(projectPath string) (*TestResults, error) {
	// Try pytest first
	if checkCommandExists("pytest") {
		return pp.runPyTest(projectPath)
	}

	// Fall back to unittest
	if checkCommandExists("python") || checkCommandExists("python3") {
		return pp.runUnittest(projectPath)
	}

	return nil, fmt.Errorf("no Python test runner available")
}

// runPyTest executes pytest
func (pp *PythonPlugin) runPyTest(projectPath string) (*TestResults, error) {
	output, err := runCommand("pytest", []string{"--tb=short", "-v"}, projectPath, 120*time.Second)
	
	results := &TestResults{
		LastRun: time.Now(),
	}

	// Parse pytest output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "passed") && strings.Contains(line, "failed") {
			// Parse summary line
			fmt.Sscanf(line, "%d failed, %d passed", &results.FailedTests, &results.PassedTests)
			results.TotalTests = results.FailedTests + results.PassedTests
		}
	}

	return results, err
}

// runUnittest executes Python unittest
func (pp *PythonPlugin) runUnittest(projectPath string) (*TestResults, error) {
	pythonCmd := "python"
	if checkCommandExists("python3") {
		pythonCmd = "python3"
	}

	output, err := runCommand(pythonCmd, []string{"-m", "unittest", "discover", "-v"}, projectPath, 120*time.Second)
	
	results := &TestResults{
		LastRun: time.Now(),
	}

	// Parse unittest output (simplified)
	if strings.Contains(output, "OK") {
		results.PassedTests = 1 // Simplified - real implementation would count tests
		results.TotalTests = 1
	} else if strings.Contains(output, "FAILED") {
		results.FailedTests = 1
		results.TotalTests = 1
	}

	return results, err
}

// ParseBuildOutput parses Python build output for errors
func (pp *PythonPlugin) ParseBuildOutput(output string) ([]ErrorInfo, error) {
	return parseErrorWithRegex(output, pp.ErrorPatterns), nil
}