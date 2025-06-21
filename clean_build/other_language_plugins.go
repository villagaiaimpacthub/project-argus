package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// JavaPlugin implements language support for Java
type JavaPlugin struct {
	BaseLanguagePlugin
}

// NewJavaPlugin creates a new Java language plugin
func NewJavaPlugin() *JavaPlugin {
	return &JavaPlugin{
		BaseLanguagePlugin: BaseLanguagePlugin{
			Name:        "java",
			Extensions:  []string{".java"},
			ConfigFiles: []string{"pom.xml", "build.gradle", "build.gradle.kts"},
			ErrorPatterns: []ErrorPattern{
				{
					Pattern:   `(.+\.java):(\d+): error: (.+)`,
					Type:      "compile",
					Severity:  "error",
					Language:  "java",
					FileRegex: `(.+\.java):(\d+):`,
					LineRegex: `.+\.java:(\d+):`,
				},
			},
		},
	}
}

func (jp *JavaPlugin) AnalyzeErrors(projectPath string) ([]ErrorInfo, error) {
	var allErrors []ErrorInfo

	// Check for Maven
	if _, err := os.Stat(filepath.Join(projectPath, "pom.xml")); err == nil {
		if errors := jp.runMavenCompile(projectPath); len(errors) > 0 {
			allErrors = append(allErrors, errors...)
		}
	}

	// Check for Gradle
	if _, err := os.Stat(filepath.Join(projectPath, "build.gradle")); err == nil {
		if errors := jp.runGradleCompile(projectPath); len(errors) > 0 {
			allErrors = append(allErrors, errors...)
		}
	}

	return allErrors, nil
}

func (jp *JavaPlugin) runMavenCompile(projectPath string) []ErrorInfo {
	if !checkCommandExists("mvn") {
		return nil
	}

	output, err := runCommand("mvn", []string{"compile"}, projectPath, 120*time.Second)
	if err != nil && output != "" {
		return parseErrorWithRegex(output, jp.ErrorPatterns)
	}
	return nil
}

func (jp *JavaPlugin) runGradleCompile(projectPath string) []ErrorInfo {
	if !checkCommandExists("gradle") && !checkCommandExists("./gradlew") {
		return nil
	}

	cmd := "gradle"
	if _, err := os.Stat(filepath.Join(projectPath, "gradlew")); err == nil {
		cmd = "./gradlew"
	}

	output, err := runCommand(cmd, []string{"compileJava"}, projectPath, 120*time.Second)
	if err != nil && output != "" {
		return parseErrorWithRegex(output, jp.ErrorPatterns)
	}
	return nil
}

func (jp *JavaPlugin) GetDependencies(projectPath string) ([]DependencyInfo, error) {
	return []DependencyInfo{}, nil // Simplified implementation
}

func (jp *JavaPlugin) FindServices(projectPath string) ([]ServiceInfo, error) {
	defaultPorts := []int{8080, 8443, 9090}
	return detectRunningProcesses(projectPath, defaultPorts), nil
}

func (jp *JavaPlugin) RunLinter(projectPath string) ([]ErrorInfo, error) {
	return []ErrorInfo{}, nil
}

func (jp *JavaPlugin) RunTests(projectPath string) (*TestResults, error) {
	return &TestResults{LastRun: time.Now()}, nil
}

func (jp *JavaPlugin) ParseBuildOutput(output string) ([]ErrorInfo, error) {
	return parseErrorWithRegex(output, jp.ErrorPatterns), nil
}

// CSharpPlugin implements language support for C#
type CSharpPlugin struct {
	BaseLanguagePlugin
}

// NewCSharpPlugin creates a new C# language plugin
func NewCSharpPlugin() *CSharpPlugin {
	return &CSharpPlugin{
		BaseLanguagePlugin: BaseLanguagePlugin{
			Name:        "csharp",
			Extensions:  []string{".cs"},
			ConfigFiles: []string{".csproj", ".sln", "global.json"},
			ErrorPatterns: []ErrorPattern{
				{
					Pattern:   `(.+\.cs)\((\d+),(\d+)\): error (.+): (.+)`,
					Type:      "compile",
					Severity:  "error",
					Language:  "csharp",
					FileRegex: `(.+\.cs)\((\d+),(\d+)\):`,
					LineRegex: `.+\.cs\((\d+),\d+\):`,
				},
			},
		},
	}
}

func (csp *CSharpPlugin) AnalyzeErrors(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("dotnet") {
		return nil, fmt.Errorf("dotnet CLI not available")
	}

	output, err := runCommand("dotnet", []string{"build"}, projectPath, 120*time.Second)
	if err != nil && output != "" {
		return parseErrorWithRegex(output, csp.ErrorPatterns), nil
	}

	return []ErrorInfo{}, nil
}

func (csp *CSharpPlugin) GetDependencies(projectPath string) ([]DependencyInfo, error) {
	return []DependencyInfo{}, nil
}

func (csp *CSharpPlugin) FindServices(projectPath string) ([]ServiceInfo, error) {
	defaultPorts := []int{5000, 5001, 8080}
	return detectRunningProcesses(projectPath, defaultPorts), nil
}

func (csp *CSharpPlugin) RunLinter(projectPath string) ([]ErrorInfo, error) {
	return []ErrorInfo{}, nil
}

func (csp *CSharpPlugin) RunTests(projectPath string) (*TestResults, error) {
	return &TestResults{LastRun: time.Now()}, nil
}

func (csp *CSharpPlugin) ParseBuildOutput(output string) ([]ErrorInfo, error) {
	return parseErrorWithRegex(output, csp.ErrorPatterns), nil
}

// RustPlugin implements language support for Rust
type RustPlugin struct {
	BaseLanguagePlugin
}

// NewRustPlugin creates a new Rust language plugin
func NewRustPlugin() *RustPlugin {
	return &RustPlugin{
		BaseLanguagePlugin: BaseLanguagePlugin{
			Name:        "rust",
			Extensions:  []string{".rs"},
			ConfigFiles: []string{"Cargo.toml", "Cargo.lock"},
			ErrorPatterns: []ErrorPattern{
				{
					Pattern:  `error\[E\d+\]: (.+)`,
					Type:     "compile",
					Severity: "error",
					Language: "rust",
				},
				{
					Pattern:  `warning: (.+)`,
					Type:     "lint",
					Severity: "warning",
					Language: "rust",
				},
			},
		},
	}
}

func (rp *RustPlugin) AnalyzeErrors(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("cargo") {
		return nil, fmt.Errorf("cargo not available")
	}

	output, err := runCommand("cargo", []string{"check"}, projectPath, 120*time.Second)
	if err != nil && output != "" {
		return parseErrorWithRegex(output, rp.ErrorPatterns), nil
	}

	return []ErrorInfo{}, nil
}

func (rp *RustPlugin) GetDependencies(projectPath string) ([]DependencyInfo, error) {
	return []DependencyInfo{}, nil
}

func (rp *RustPlugin) FindServices(projectPath string) ([]ServiceInfo, error) {
	defaultPorts := []int{8080, 8000, 3000}
	return detectRunningProcesses(projectPath, defaultPorts), nil
}

func (rp *RustPlugin) RunLinter(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("cargo") {
		return nil, fmt.Errorf("cargo not available")
	}

	output, err := runCommand("cargo", []string{"clippy"}, projectPath, 60*time.Second)
	if err != nil && output != "" {
		return parseErrorWithRegex(output, rp.ErrorPatterns), nil
	}

	return []ErrorInfo{}, nil
}

func (rp *RustPlugin) RunTests(projectPath string) (*TestResults, error) {
	return &TestResults{LastRun: time.Now()}, nil
}

func (rp *RustPlugin) ParseBuildOutput(output string) ([]ErrorInfo, error) {
	return parseErrorWithRegex(output, rp.ErrorPatterns), nil
}

// PHPPlugin implements language support for PHP
type PHPPlugin struct {
	BaseLanguagePlugin
}

// NewPHPPlugin creates a new PHP language plugin
func NewPHPPlugin() *PHPPlugin {
	return &PHPPlugin{
		BaseLanguagePlugin: BaseLanguagePlugin{
			Name:        "php",
			Extensions:  []string{".php"},
			ConfigFiles: []string{"composer.json", "composer.lock"},
			ErrorPatterns: []ErrorPattern{
				{
					Pattern:  `Parse error: (.+) in (.+) on line (\d+)`,
					Type:     "syntax",
					Severity: "error",
					Language: "php",
				},
			},
		},
	}
}

func (pp *PHPPlugin) AnalyzeErrors(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("php") {
		return nil, fmt.Errorf("PHP not available")
	}

	// Run syntax check on PHP files
	phpFiles, _ := findFilesWithExtensions(projectPath, pp.Extensions)
	var errors []ErrorInfo

	for _, file := range phpFiles {
		output, err := runCommand("php", []string{"-l", file}, projectPath, 10*time.Second)
		if err != nil && output != "" {
			fileErrors := parseErrorWithRegex(output, pp.ErrorPatterns)
			errors = append(errors, fileErrors...)
		}
	}

	return errors, nil
}

func (pp *PHPPlugin) GetDependencies(projectPath string) ([]DependencyInfo, error) {
	return []DependencyInfo{}, nil
}

func (pp *PHPPlugin) FindServices(projectPath string) ([]ServiceInfo, error) {
	defaultPorts := []int{8000, 8080, 80}
	return detectRunningProcesses(projectPath, defaultPorts), nil
}

func (pp *PHPPlugin) RunLinter(projectPath string) ([]ErrorInfo, error) {
	return []ErrorInfo{}, nil
}

func (pp *PHPPlugin) RunTests(projectPath string) (*TestResults, error) {
	return &TestResults{LastRun: time.Now()}, nil
}

func (pp *PHPPlugin) ParseBuildOutput(output string) ([]ErrorInfo, error) {
	return parseErrorWithRegex(output, pp.ErrorPatterns), nil
}

// RubyPlugin implements language support for Ruby
type RubyPlugin struct {
	BaseLanguagePlugin
}

// NewRubyPlugin creates a new Ruby language plugin
func NewRubyPlugin() *RubyPlugin {
	return &RubyPlugin{
		BaseLanguagePlugin: BaseLanguagePlugin{
			Name:        "ruby",
			Extensions:  []string{".rb"},
			ConfigFiles: []string{"Gemfile", "Gemfile.lock"},
			ErrorPatterns: []ErrorPattern{
				{
					Pattern:  `(.+\.rb):(\d+): (.+)`,
					Type:     "syntax",
					Severity: "error",
					Language: "ruby",
				},
			},
		},
	}
}

func (rp *RubyPlugin) AnalyzeErrors(projectPath string) ([]ErrorInfo, error) {
	if !checkCommandExists("ruby") {
		return nil, fmt.Errorf("Ruby not available")
	}

	// Run syntax check on Ruby files
	rubyFiles, _ := findFilesWithExtensions(projectPath, rp.Extensions)
	var errors []ErrorInfo

	for _, file := range rubyFiles {
		output, err := runCommand("ruby", []string{"-c", file}, projectPath, 10*time.Second)
		if err != nil && output != "" {
			fileErrors := parseErrorWithRegex(output, rp.ErrorPatterns)
			errors = append(errors, fileErrors...)
		}
	}

	return errors, nil
}

func (rp *RubyPlugin) GetDependencies(projectPath string) ([]DependencyInfo, error) {
	return []DependencyInfo{}, nil
}

func (rp *RubyPlugin) FindServices(projectPath string) ([]ServiceInfo, error) {
	defaultPorts := []int{3000, 4567, 9292}
	return detectRunningProcesses(projectPath, defaultPorts), nil
}

func (rp *RubyPlugin) RunLinter(projectPath string) ([]ErrorInfo, error) {
	return []ErrorInfo{}, nil
}

func (rp *RubyPlugin) RunTests(projectPath string) (*TestResults, error) {
	return &TestResults{LastRun: time.Now()}, nil
}

func (rp *RubyPlugin) ParseBuildOutput(output string) ([]ErrorInfo, error) {
	return parseErrorWithRegex(output, rp.ErrorPatterns), nil
}
