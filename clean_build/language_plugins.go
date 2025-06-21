package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// LanguagePlugin defines the interface for language-specific analysis
type LanguagePlugin interface {
	GetName() string
	Detect(projectPath string) bool
	AnalyzeErrors(projectPath string) ([]ErrorInfo, error)
	GetDependencies(projectPath string) ([]DependencyInfo, error)
	FindServices(projectPath string) ([]ServiceInfo, error)
	GetErrorPatterns() []ErrorPattern
	RunLinter(projectPath string) ([]ErrorInfo, error)
	RunTests(projectPath string) (*TestResults, error)
	ParseBuildOutput(output string) ([]ErrorInfo, error)
}

// BaseLanguagePlugin provides common functionality for all language plugins
type BaseLanguagePlugin struct {
	Name          string
	Extensions    []string
	ConfigFiles   []string
	ErrorPatterns []ErrorPattern
	mutex         sync.RWMutex
}

// GetName returns the plugin name
func (blp *BaseLanguagePlugin) GetName() string {
	return blp.Name
}

// Detect checks if the language is present in the project
func (blp *BaseLanguagePlugin) Detect(projectPath string) bool {
	// Check for config files
	for _, configFile := range blp.ConfigFiles {
		if _, err := os.Stat(filepath.Join(projectPath, configFile)); err == nil {
			return true
		}
	}

	// Check for files with matching extensions
	found := false
	filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, targetExt := range blp.Extensions {
			if ext == targetExt {
				found = true
				return fmt.Errorf("found") // Break the walk
			}
		}
		return nil
	})

	return found
}

// GetErrorPatterns returns the language-specific error patterns
func (blp *BaseLanguagePlugin) GetErrorPatterns() []ErrorPattern {
	return blp.ErrorPatterns
}

// LanguagePluginManager manages all language plugins
type LanguagePluginManager struct {
	plugins map[string]LanguagePlugin
	mutex   sync.RWMutex
}

// NewLanguagePluginManager creates a new plugin manager with all language plugins
func NewLanguagePluginManager() *LanguagePluginManager {
	manager := &LanguagePluginManager{
		plugins: make(map[string]LanguagePlugin),
	}

	// Register all language plugins
	manager.RegisterPlugin(NewJavaScriptPlugin())
	manager.RegisterPlugin(NewTypeScriptPlugin())
	manager.RegisterPlugin(NewPythonPlugin())
	manager.RegisterPlugin(NewGoPlugin())
	manager.RegisterPlugin(NewJavaPlugin())
	manager.RegisterPlugin(NewCSharpPlugin())
	manager.RegisterPlugin(NewRustPlugin())
	manager.RegisterPlugin(NewPHPPlugin())
	manager.RegisterPlugin(NewRubyPlugin())

	return manager
}

// RegisterPlugin adds a new language plugin
func (lpm *LanguagePluginManager) RegisterPlugin(plugin LanguagePlugin) {
	lpm.mutex.Lock()
	defer lpm.mutex.Unlock()
	lpm.plugins[plugin.GetName()] = plugin
}

// DetectLanguages returns all detected languages in the project
func (lpm *LanguagePluginManager) DetectLanguages(projectPath string) []LanguagePlugin {
	lpm.mutex.RLock()
	defer lpm.mutex.RUnlock()

	var detected []LanguagePlugin
	for _, plugin := range lpm.plugins {
		if plugin.Detect(projectPath) {
			detected = append(detected, plugin)
		}
	}
	return detected
}

// AnalyzeAllErrors runs error analysis for all detected languages
func (lpm *LanguagePluginManager) AnalyzeAllErrors(projectPath string) ([]ErrorInfo, error) {
	detected := lpm.DetectLanguages(projectPath)
	var allErrors []ErrorInfo

	for _, plugin := range detected {
		errors, err := plugin.AnalyzeErrors(projectPath)
		if err == nil {
			allErrors = append(allErrors, errors...)
		}
	}

	return allErrors, nil
}

// GetPlugin returns a specific language plugin
func (lpm *LanguagePluginManager) GetPlugin(name string) (LanguagePlugin, bool) {
	lpm.mutex.RLock()
	defer lpm.mutex.RUnlock()
	plugin, exists := lpm.plugins[name]
	return plugin, exists
}

// Universal helper functions for all plugins

// runCommand executes a command and returns the output
func runCommand(command string, args []string, workDir string, timeout time.Duration) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

// parseErrorWithRegex parses error output using regex patterns
func parseErrorWithRegex(output string, patterns []ErrorPattern) []ErrorInfo {
	var errors []ErrorInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern.Pattern, line); matched {
				errorInfo := ErrorInfo{
					Source:    pattern.Language,
					Type:      pattern.Type,
					Message:   line,
					Timestamp: time.Now(),
				}

				// Extract file, line, and column if patterns are provided
				if pattern.FileRegex != "" {
					if re, err := regexp.Compile(pattern.FileRegex); err == nil {
						if matches := re.FindStringSubmatch(line); len(matches) > 1 {
							errorInfo.File = matches[1]
						}
					}
				}

				if pattern.LineRegex != "" {
					if re, err := regexp.Compile(pattern.LineRegex); err == nil {
						if matches := re.FindStringSubmatch(line); len(matches) > 1 {
							if lineNum, err := regexp.MatchString(`\d+`, matches[1]); err == nil && lineNum {
								fmt.Sscanf(matches[1], "%d", &errorInfo.Line)
							}
						}
					}
				}

				if pattern.ColumnRegex != "" {
					if re, err := regexp.Compile(pattern.ColumnRegex); err == nil {
						if matches := re.FindStringSubmatch(line); len(matches) > 1 {
							if colNum, err := regexp.MatchString(`\d+`, matches[1]); err == nil && colNum {
								fmt.Sscanf(matches[1], "%d", &errorInfo.Column)
							}
						}
					}
				}

				errors = append(errors, errorInfo)
				break // Only match the first pattern per line
			}
		}
	}

	return errors
}

// scanFileForPatterns scans a single file for error patterns
func scanFileForPatterns(filePath string, patterns []ErrorPattern) []ErrorInfo {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var errors []ErrorInfo
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern.Pattern, line); matched {
				relPath, _ := filepath.Rel(filepath.Dir(filePath), filePath)
				errors = append(errors, ErrorInfo{
					Source:    pattern.Language,
					File:      relPath,
					Line:      lineNum,
					Type:      pattern.Type,
					Message:   strings.TrimSpace(line),
					Timestamp: time.Now(),
				})
				break
			}
		}
	}

	return errors
}

// findFilesWithExtensions finds all files with specific extensions
func findFilesWithExtensions(projectPath string, extensions []string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			// Skip common ignore directories
			skipDirs := []string{"node_modules", "vendor", "target", "dist", "build", "__pycache__", ".git"}
			for _, skip := range skipDirs {
				if d.Name() == skip {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, targetExt := range extensions {
			if ext == targetExt {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

// checkCommandExists verifies if a command is available in PATH
func checkCommandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// parsePackageManagerOutput parses dependency information from package manager output
func parsePackageManagerOutput(output string, packageManager string) []DependencyInfo {
	var deps []DependencyInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Different parsing logic based on package manager
		switch packageManager {
		case "npm":
			// Parse npm list output
			if strings.Contains(line, "@") {
				parts := strings.Split(line, "@")
				if len(parts) >= 2 {
					name := strings.TrimSpace(parts[0])
					version := strings.TrimSpace(parts[1])
					deps = append(deps, DependencyInfo{
						Name:    name,
						Version: version,
						Source:  "npm",
						Type:    "direct",
					})
				}
			}
		case "pip":
			// Parse pip list output
			if strings.Contains(line, "==") {
				parts := strings.Split(line, "==")
				if len(parts) == 2 {
					deps = append(deps, DependencyInfo{
						Name:    strings.TrimSpace(parts[0]),
						Version: strings.TrimSpace(parts[1]),
						Source:  "pip",
						Type:    "direct",
					})
				}
			}
		case "go":
			// Parse go list output
			if strings.HasPrefix(line, "go.mod") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				deps = append(deps, DependencyInfo{
					Name:    parts[0],
					Version: parts[1],
					Source:  "go.mod",
					Type:    "direct",
				})
			}
		}
	}

	return deps
}

// detectRunningProcesses finds running processes related to the project
func detectRunningProcesses(projectPath string, defaultPorts []int) []ServiceInfo {
	var services []ServiceInfo

	// Check for processes listening on default ports
	for _, port := range defaultPorts {
		if isPortInUse(port) {
			service := ServiceInfo{
				ID:     fmt.Sprintf("service-%d", port),
				Name:   fmt.Sprintf("Service on port %d", port),
				Port:   port,
				Status: "running",
			}
			services = append(services, service)
		}
	}

	return services
}

// isPortInUse checks if a port is currently in use
func isPortInUse(port int) bool {
	// Try different commands based on OS
	commands := [][]string{
		{"netstat", "-an"},
		{"ss", "-tuln"},
		{"lsof", "-i", fmt.Sprintf(":%d", port)},
	}

	for _, cmd := range commands {
		if checkCommandExists(cmd[0]) {
			output, err := runCommand(cmd[0], cmd[1:], ".", 5*time.Second)
			if err == nil && strings.Contains(output, fmt.Sprintf(":%d", port)) {
				return true
			}
		}
	}

	return false
}
