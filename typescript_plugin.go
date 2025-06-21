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

// TypeScriptPlugin implements language support for TypeScript
type TypeScriptPlugin struct {
	BaseLanguagePlugin
}

// NewTypeScriptPlugin creates a new TypeScript language plugin
func NewTypeScriptPlugin() *TypeScriptPlugin {
	return &TypeScriptPlugin{
		BaseLanguagePlugin: BaseLanguagePlugin{
			Name:        "typescript",
			Extensions:  []string{".ts", ".tsx"},
			ConfigFiles: []string{"tsconfig.json", "tsconfig.build.json", "tslint.json"},
			ErrorPatterns: []ErrorPattern{
				{
					Pattern:     `error TS\d+: .+`,
					Type:        "syntax",
					Severity:    "error",
					Language:    "typescript",
					FileRegex:   `(.+)\(\d+,\d+\):`,
					LineRegex:   `.+\((\d+),\d+\):`,
					ColumnRegex: `.+\(\d+,(\d+)\):`,
				},
				{
					Pattern:     `\(\d+,\d+\): error TS\d+: .+`,
					Type:        "type",
					Severity:    "error",
					Language:    "typescript",
					FileRegex:   `(.+)\(\d+,\d+\):`,
					LineRegex:   `.+\((\d+),\d+\):`,
					ColumnRegex: `.+\(\d+,(\d+)\):`,
				},
				{
					Pattern:     `Cannot find module .+`,
					Type:        "import",
					Severity:    "error",
					Language:    "typescript",
					FileRegex:   `in (.+)`,
				},
				{
					Pattern:     `Property .+ does not exist on type .+`,
					Type:        "type",
					Severity:    "error",
					Language:    "typescript",
				},
				{
					Pattern:     `Argument of type .+ is not assignable to parameter of type .+`,
					Type:        "type",
					Severity:    "error",
					Language:    "typescript",
				},
				{
					Pattern:     `warning TS\d+: .+`,
					Type:        "lint",
					Severity:    "warning",
					Language:    "typescript",
					FileRegex:   `(.+)\(\d+,\d+\):`,
					LineRegex:   `.+\((\d+),\d+\):`,
					ColumnRegex: `.+\(\d+,(\d+)\):`,
				},
			},
		},
	}
}

// AnalyzeErrors performs comprehensive error analysis for TypeScript projects
func (tsp *TypeScriptPlugin) AnalyzeErrors(projectPath string) ([]ErrorInfo, error) {
	var allErrors []ErrorInfo

	// 1. Run TypeScript compiler check
	if tscErrors, err := tsp.runTypeScriptCompiler(projectPath); err == nil {
		allErrors = append(allErrors, tscErrors...)
	}

	// 2. Run TSLint/ESLint if available
	if lintErrors, err := tsp.runTypeScriptLinter(projectPath); err == nil {
		allErrors = append(allErrors, lintErrors...)
	}

	// 3. Check for type definition issues
	if typeErrors := tsp.checkTypeDefinitions(projectPath); len(typeErrors) > 0 {
		allErrors = append(allErrors, typeErrors...)
	}

	// 4. Inherit JavaScript error checking for JS files in TS projects
	jsPlugin := NewJavaScriptPlugin()
	if jsErrors, err := jsPlugin.AnalyzeErrors(projectPath); err == nil {
		// Filter to only include JS-specific errors
		for _, err := range jsErrors {
			if strings.HasSuffix(err.File, ".js") || strings.HasSuffix(err.File, ".jsx") {
				allErrors = append(allErrors, err)
			}
		}
	}

	return allErrors, nil
}

// runTypeScriptCompiler runs tsc for type checking
func (tsp *TypeScriptPlugin) runTypeScriptCompiler(projectPath string) ([]ErrorInfo, error) {
	// Check if TypeScript is available
	tscCommands := [][]string{
		{"npx", "tsc", "--noEmit", "--pretty", "false"},
		{"tsc", "--noEmit", "--pretty", "false"},
	}

	var output string
	var err error

	for _, cmd := range tscCommands {
		if checkCommandExists(cmd[0]) {
			output, err = runCommand(cmd[0], cmd[1:], projectPath, 60*time.Second)
			break
		}
	}

	if err != nil && output == "" {
		return nil, fmt.Errorf("TypeScript compiler not available")
	}

	return tsp.parseTypeScriptOutput(output), nil
}

// parseTypeScriptOutput parses TypeScript compiler output
func (tsp *TypeScriptPlugin) parseTypeScriptOutput(output string) []ErrorInfo {
	var errors []ErrorInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// TypeScript error format: file.ts(line,col): error TSxxxx: message
		tsErrorRegex := regexp.MustCompile(`(.+)\((\d+),(\d+)\): (error|warning) (TS\d+): (.+)`)
		matches := tsErrorRegex.FindStringSubmatch(line)

		if len(matches) == 7 {
			line, _ := strconv.Atoi(matches[2])
			column, _ := strconv.Atoi(matches[3])
			_ = matches[4]
			code := matches[5]
			message := matches[6]

			errorType := "type"
			if strings.Contains(message, "Cannot find module") {
				errorType = "import"
			} else if strings.Contains(message, "syntax") {
				errorType = "syntax"
			}

			errors = append(errors, ErrorInfo{
				Source:    "tsc",
				File:      matches[1],
				Line:      line,
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

// runTypeScriptLinter runs TSLint or ESLint for TypeScript
func (tsp *TypeScriptPlugin) runTypeScriptLinter(projectPath string) ([]ErrorInfo, error) {
	var allErrors []ErrorInfo

	// Try TSLint first (legacy)
	if tslintErrors, err := tsp.runTSLint(projectPath); err == nil {
		allErrors = append(allErrors, tslintErrors...)
	}

	// Try ESLint with TypeScript support
	if eslintErrors, err := tsp.runESLintForTypeScript(projectPath); err == nil {
		allErrors = append(allErrors, eslintErrors...)
	}

	return allErrors, nil
}

// runTSLint executes TSLint
func (tsp *TypeScriptPlugin) runTSLint(projectPath string) ([]ErrorInfo, error) {
	// Check if TSLint is configured
	tslintConfig := filepath.Join(projectPath, "tslint.json")
	if _, err := os.Stat(tslintConfig); err != nil {
		return nil, fmt.Errorf("tslint.json not found")
	}

	if !checkCommandExists("npx") {
		return nil, fmt.Errorf("npx not available")
	}

	output, err := runCommand("npx", []string{"tslint", "--format", "json", "**/*.ts"}, projectPath, 30*time.Second)
	if err != nil && output == "" {
		return nil, err
	}

	return tsp.parseTSLintOutput(output)
}

// parseTSLintOutput parses TSLint JSON output
func (tsp *TypeScriptPlugin) parseTSLintOutput(output string) ([]ErrorInfo, error) {
	var tslintResults []struct {
		Name      string `json:"name"`
		RuleName  string `json:"ruleName"`
		StartPos  struct {
			Line      int `json:"line"`
			Character int `json:"character"`
		} `json:"startPosition"`
		EndPos struct {
			Line      int `json:"line"`
			Character int `json:"character"`
		} `json:"endPosition"`
		Failure   string `json:"failure"`
		RuleSeverity string `json:"ruleSeverity"`
	}

	if err := json.Unmarshal([]byte(output), &tslintResults); err != nil {
		// If JSON parsing fails, try regex parsing
		return parseErrorWithRegex(output, tsp.ErrorPatterns), nil
	}

	var errors []ErrorInfo
	for _, result := range tslintResults {
		errors = append(errors, ErrorInfo{
			Source:    "tslint",
			File:      result.Name,
			Line:      result.StartPos.Line + 1, // TSLint uses 0-based line numbers
			Column:    result.StartPos.Character + 1,
			Type:      "lint",
			Message:   result.Failure,
			Code:      result.RuleName,
			Timestamp: time.Now(),
		})
	}

	return errors, nil
}

// runESLintForTypeScript runs ESLint with TypeScript support
func (tsp *TypeScriptPlugin) runESLintForTypeScript(projectPath string) ([]ErrorInfo, error) {
	// Check for ESLint config with TypeScript support
	eslintConfigs := []string{".eslintrc.js", ".eslintrc.json", ".eslintrc.yml"}
	hasESLint := false
	var configContent string

	for _, config := range eslintConfigs {
		configPath := filepath.Join(projectPath, config)
		if content, err := os.ReadFile(configPath); err == nil {
			configContent = string(content)
			// Check if TypeScript parser is configured
			if strings.Contains(configContent, "@typescript-eslint") || strings.Contains(configContent, "typescript") {
				hasESLint = true
				break
			}
		}
	}

	if !hasESLint {
		return nil, fmt.Errorf("no ESLint configuration with TypeScript support found")
	}

	if !checkCommandExists("npx") {
		return nil, fmt.Errorf("npx not available")
	}

	// Run ESLint with TypeScript extensions
	output, err := runCommand("npx", []string{"eslint", ".", "--ext", ".ts,.tsx", "--format", "json"}, projectPath, 30*time.Second)
	if err != nil && output == "" {
		return nil, err
	}

	// Use JavaScript plugin's ESLint parser since the format is the same
	jsPlugin := NewJavaScriptPlugin()
	return jsPlugin.parseESLintOutput(output)
}

// checkTypeDefinitions checks for type definition issues
func (tsp *TypeScriptPlugin) checkTypeDefinitions(projectPath string) []ErrorInfo {
	var errors []ErrorInfo

	// Check for missing @types packages
	packageJSONPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err != nil {
		return errors
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return errors
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return errors
	}

	// Common packages that need @types
	typesNeeded := map[string]string{
		"node":    "@types/node",
		"express": "@types/express",
		"lodash":  "@types/lodash",
		"react":   "@types/react",
		"jest":    "@types/jest",
	}

	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}

	for dep := range pkg.Dependencies {
		if typePkg, needed := typesNeeded[dep]; needed {
			if _, hasTypes := allDeps[typePkg]; !hasTypes {
				errors = append(errors, ErrorInfo{
					Source:    "typescript",
					File:      "package.json",
					Type:      "type",
					Message:   fmt.Sprintf("Missing type definitions for '%s'. Consider installing '%s'", dep, typePkg),
					Timestamp: time.Now(),
				})
			}
		}
	}

	return errors
}

// GetDependencies analyzes TypeScript dependencies
func (tsp *TypeScriptPlugin) GetDependencies(projectPath string) ([]DependencyInfo, error) {
	// TypeScript projects typically use npm/yarn, so delegate to JavaScript plugin
	jsPlugin := NewJavaScriptPlugin()
	deps, err := jsPlugin.GetDependencies(projectPath)
	if err != nil {
		return nil, err
	}

	// Add TypeScript-specific dependency information
	for i, dep := range deps {
		if strings.HasPrefix(dep.Name, "@types/") {
			deps[i].Type = "types"
		} else if dep.Name == "typescript" {
			deps[i].Type = "compiler"
		}
	}

	return deps, nil
}

// FindServices discovers running TypeScript services
func (tsp *TypeScriptPlugin) FindServices(projectPath string) ([]ServiceInfo, error) {
	// TypeScript services are typically run through npm/node, so use JavaScript discovery
	jsPlugin := NewJavaScriptPlugin()
	services, err := jsPlugin.FindServices(projectPath)
	if err != nil {
		return nil, err
	}

	// Update language field for TypeScript projects
	for i := range services {
		// Check if the project has TypeScript files
		if tsp.Detect(projectPath) {
			services[i].Language = "typescript"
		}
	}

	return services, nil
}

// RunLinter executes the linter for TypeScript
func (tsp *TypeScriptPlugin) RunLinter(projectPath string) ([]ErrorInfo, error) {
	return tsp.runTypeScriptLinter(projectPath)
}

// RunTests executes tests for TypeScript projects
func (tsp *TypeScriptPlugin) RunTests(projectPath string) (*TestResults, error) {
	// TypeScript tests are typically run through npm scripts, so delegate to JavaScript plugin
	jsPlugin := NewJavaScriptPlugin()
	return jsPlugin.RunTests(projectPath)
}

// ParseBuildOutput parses TypeScript build output for errors
func (tsp *TypeScriptPlugin) ParseBuildOutput(output string) ([]ErrorInfo, error) {
	errors := parseErrorWithRegex(output, tsp.ErrorPatterns)
	
	// Also parse JavaScript build errors since TS projects often have mixed content
	jsPlugin := NewJavaScriptPlugin()
	jsErrors, _ := jsPlugin.ParseBuildOutput(output)
	errors = append(errors, jsErrors...)
	
	return errors, nil
}