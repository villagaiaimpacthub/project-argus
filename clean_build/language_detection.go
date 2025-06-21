package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// LanguageSupport represents comprehensive language ecosystem support
type LanguageSupport struct {
	Name            string   `json:"name"`
	Extensions      []string `json:"extensions"`
	ConfigFiles     []string `json:"config_files"`
	LintTools       []string `json:"lint_tools"`
	TestFrameworks  []string `json:"test_frameworks"`
	PackageManagers []string `json:"package_managers"`
	BuildTools      []string `json:"build_tools"`
	Frameworks      []string `json:"frameworks"`
	DefaultPorts    []int    `json:"default_ports"`
}

// FrameworkInfo represents detected framework information
type FrameworkInfo struct {
	Name         string            `json:"name"`
	Language     string            `json:"language"`
	Version      string            `json:"version"`
	ConfigFile   string            `json:"config_file"`
	DefaultPort  int               `json:"default_port"`
	DevCommand   string            `json:"dev_command"`
	BuildCommand string            `json:"build_command"`
	TestCommand  string            `json:"test_command"`
	Metadata     map[string]string `json:"metadata"`
}

// BuildTool represents build system information
type BuildTool struct {
	Name        string   `json:"name"`
	Language    string   `json:"language"`
	ConfigFiles []string `json:"config_files"`
	Commands    []string `json:"commands"`
	OutputDir   string   `json:"output_dir"`
}

// LanguageDetector manages multi-language project analysis
type LanguageDetector struct {
	Languages    map[string]*LanguageSupport `json:"languages"`
	Frameworks   []FrameworkInfo             `json:"frameworks"`
	BuildSystems []BuildTool                 `json:"build_systems"`
	workspace    string
	mutex        sync.RWMutex
}

// DetectedLanguage represents a language found in the project
type DetectedLanguage struct {
	Language       *LanguageSupport `json:"language"`
	FileCount      int              `json:"file_count"`
	LineCount      int              `json:"line_count"`
	MainFiles      []string         `json:"main_files"`
	ConfigFiles    []string         `json:"config_files"`
	Frameworks     []FrameworkInfo  `json:"frameworks"`
	BuildTools     []BuildTool      `json:"build_tools"`
	HasTests       bool             `json:"has_tests"`
	HasLinting     bool             `json:"has_linting"`
	PackageManager string           `json:"package_manager"`
}

// ErrorPattern represents language-specific error detection patterns
type ErrorPattern struct {
	Pattern     string `json:"pattern"`
	Type        string `json:"type"`     // "syntax", "runtime", "lint", "test", "build"
	Severity    string `json:"severity"` // "error", "warning", "info"
	LineRegex   string `json:"line_regex"`
	ColumnRegex string `json:"column_regex"`
	FileRegex   string `json:"file_regex"`
	Language    string `json:"language"`
}

// ServiceInfo represents detected service information
type ServiceInfo struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Language     string            `json:"language"`
	Framework    string            `json:"framework"`
	Port         int               `json:"port"`
	Status       string            `json:"status"` // "running", "stopped", "error"
	ProcessID    int               `json:"process_id"`
	ConfigFile   string            `json:"config_file"`
	Endpoints    []APIEndpoint     `json:"endpoints"`
	Environment  map[string]string `json:"environment"`
	StartCommand string            `json:"start_command"`
}

// APIEndpoint represents discovered API endpoints
type APIEndpoint struct {
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Handler string            `json:"handler"`
	File    string            `json:"file"`
	Line    int               `json:"line"`
	Params  []string          `json:"params"`
	Headers map[string]string `json:"headers"`
}

// ProjectTopology represents the overall project structure and relationships
type ProjectTopology struct {
	Languages []DetectedLanguage `json:"languages"`
	Services  []ServiceInfo      `json:"services"`
	Databases []DatabaseInfo     `json:"databases"`
	APIs      []APIEndpoint      `json:"apis"`
	Relations []ServiceRelation  `json:"relations"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// ServiceRelation represents relationships between project components
type ServiceRelation struct {
	From         string    `json:"from"`
	To           string    `json:"to"`
	Type         string    `json:"type"`       // "api_call", "database_query", "import", "dependency"
	Protocol     string    `json:"protocol"`   // "http", "grpc", "direct", "file_system"
	Confidence   float64   `json:"confidence"` // 0.0-1.0 confidence level
	Evidence     []string  `json:"evidence"`   // Files/lines that show this relationship
	LastDetected time.Time `json:"last_detected"`
}

// DatabaseInfo represents database connection information
type DatabaseInfo struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"` // "postgresql", "mysql", "mongodb", "redis"
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	Database     string            `json:"database"`
	Status       string            `json:"status"`
	UsedBy       []string          `json:"used_by"`       // Service IDs that use this database
	ConfigSource string            `json:"config_source"` // Where the config was found
	Environment  map[string]string `json:"environment"`
}

// NewLanguageDetector creates a new language detector with predefined language support
func NewLanguageDetector(workspace string) *LanguageDetector {
	detector := &LanguageDetector{
		Languages:    make(map[string]*LanguageSupport),
		Frameworks:   []FrameworkInfo{},
		BuildSystems: []BuildTool{},
		workspace:    workspace,
	}

	detector.initializeLanguageSupport()
	return detector
}

// initializeLanguageSupport sets up comprehensive language support definitions
func (ld *LanguageDetector) initializeLanguageSupport() {
	// JavaScript/TypeScript ecosystem
	ld.Languages["javascript"] = &LanguageSupport{
		Name:            "JavaScript",
		Extensions:      []string{".js", ".jsx", ".mjs", ".cjs"},
		ConfigFiles:     []string{"package.json", ".eslintrc.js", ".eslintrc.json", "babel.config.js", "webpack.config.js"},
		LintTools:       []string{"eslint", "jshint", "prettier", "biome"},
		TestFrameworks:  []string{"jest", "mocha", "jasmine", "vitest", "cypress", "playwright"},
		PackageManagers: []string{"npm", "yarn", "pnpm", "bun"},
		BuildTools:      []string{"webpack", "rollup", "parcel", "esbuild", "vite"},
		Frameworks:      []string{"react", "vue", "angular", "express", "nextjs", "nuxt", "svelte"},
		DefaultPorts:    []int{3000, 3001, 8080, 8000},
	}

	ld.Languages["typescript"] = &LanguageSupport{
		Name:            "TypeScript",
		Extensions:      []string{".ts", ".tsx"},
		ConfigFiles:     []string{"tsconfig.json", "tsconfig.build.json", "tslint.json"},
		LintTools:       []string{"tslint", "eslint", "prettier", "biome"},
		TestFrameworks:  []string{"jest", "mocha", "vitest", "cypress", "playwright"},
		PackageManagers: []string{"npm", "yarn", "pnpm", "bun"},
		BuildTools:      []string{"tsc", "webpack", "rollup", "vite", "esbuild"},
		Frameworks:      []string{"react", "vue", "angular", "express", "nestjs", "nextjs"},
		DefaultPorts:    []int{3000, 3001, 8080, 8000},
	}

	// Python ecosystem
	ld.Languages["python"] = &LanguageSupport{
		Name:            "Python",
		Extensions:      []string{".py", ".pyw", ".pyx", ".pyi"},
		ConfigFiles:     []string{"requirements.txt", "pyproject.toml", "setup.py", "setup.cfg", "Pipfile", "poetry.lock"},
		LintTools:       []string{"pylint", "flake8", "black", "mypy", "ruff", "bandit"},
		TestFrameworks:  []string{"pytest", "unittest", "nose2", "doctest"},
		PackageManagers: []string{"pip", "poetry", "conda", "pipenv"},
		BuildTools:      []string{"setuptools", "poetry", "flit", "wheel"},
		Frameworks:      []string{"django", "flask", "fastapi", "tornado", "pyramid", "bottle"},
		DefaultPorts:    []int{8000, 5000, 8080, 9000},
	}

	// Go ecosystem
	ld.Languages["go"] = &LanguageSupport{
		Name:            "Go",
		Extensions:      []string{".go"},
		ConfigFiles:     []string{"go.mod", "go.sum", "go.work"},
		LintTools:       []string{"golint", "golangci-lint", "staticcheck", "vet"},
		TestFrameworks:  []string{"testing", "testify", "ginkgo", "gomega"},
		PackageManagers: []string{"go"},
		BuildTools:      []string{"go"},
		Frameworks:      []string{"gin", "echo", "fiber", "chi", "gorilla", "beego"},
		DefaultPorts:    []int{8080, 8000, 3000, 9000},
	}

	// Java ecosystem
	ld.Languages["java"] = &LanguageSupport{
		Name:            "Java",
		Extensions:      []string{".java"},
		ConfigFiles:     []string{"pom.xml", "build.gradle", "build.gradle.kts", "settings.gradle"},
		LintTools:       []string{"checkstyle", "pmd", "spotbugs", "errorprone"},
		TestFrameworks:  []string{"junit", "testng", "spock", "mockito"},
		PackageManagers: []string{"maven", "gradle"},
		BuildTools:      []string{"maven", "gradle", "ant"},
		Frameworks:      []string{"spring", "spring-boot", "quarkus", "micronaut", "struts"},
		DefaultPorts:    []int{8080, 8443, 9090, 8000},
	}

	// C# ecosystem
	ld.Languages["csharp"] = &LanguageSupport{
		Name:            "C#",
		Extensions:      []string{".cs", ".csx"},
		ConfigFiles:     []string{".csproj", ".sln", "global.json", "Directory.Build.props"},
		LintTools:       []string{"stylecop", "editorconfig", "sonaranalyzer"},
		TestFrameworks:  []string{"xunit", "nunit", "mstest", "specflow"},
		PackageManagers: []string{"nuget", "dotnet"},
		BuildTools:      []string{"msbuild", "dotnet"},
		Frameworks:      []string{"aspnet", "blazor", "maui", "wpf", "winforms"},
		DefaultPorts:    []int{5000, 5001, 8080, 443},
	}

	// Rust ecosystem
	ld.Languages["rust"] = &LanguageSupport{
		Name:            "Rust",
		Extensions:      []string{".rs"},
		ConfigFiles:     []string{"Cargo.toml", "Cargo.lock", "rust-toolchain.toml"},
		LintTools:       []string{"clippy", "rustfmt", "miri"},
		TestFrameworks:  []string{"cargo-test", "proptest", "quickcheck"},
		PackageManagers: []string{"cargo"},
		BuildTools:      []string{"cargo", "rustc"},
		Frameworks:      []string{"actix", "rocket", "warp", "axum", "tide"},
		DefaultPorts:    []int{8080, 8000, 3000, 8888},
	}

	// PHP ecosystem
	ld.Languages["php"] = &LanguageSupport{
		Name:            "PHP",
		Extensions:      []string{".php", ".phtml", ".php3", ".php4", ".php5"},
		ConfigFiles:     []string{"composer.json", "composer.lock", "phpunit.xml", "psalm.xml"},
		LintTools:       []string{"phpstan", "psalm", "phpcs", "phpmd"},
		TestFrameworks:  []string{"phpunit", "pest", "codeception", "behat"},
		PackageManagers: []string{"composer"},
		BuildTools:      []string{"composer", "phing"},
		Frameworks:      []string{"laravel", "symfony", "codeigniter", "yii", "cakephp"},
		DefaultPorts:    []int{8000, 8080, 80, 443},
	}

	// Ruby ecosystem
	ld.Languages["ruby"] = &LanguageSupport{
		Name:            "Ruby",
		Extensions:      []string{".rb", ".rbw", ".rake", ".gemspec"},
		ConfigFiles:     []string{"Gemfile", "Gemfile.lock", "Rakefile", ".rubocop.yml"},
		LintTools:       []string{"rubocop", "reek", "flog", "brakeman"},
		TestFrameworks:  []string{"rspec", "minitest", "test-unit", "cucumber"},
		PackageManagers: []string{"gem", "bundler"},
		BuildTools:      []string{"rake", "bundler"},
		Frameworks:      []string{"rails", "sinatra", "hanami", "grape", "cuba"},
		DefaultPorts:    []int{3000, 4567, 8080, 9292},
	}
}

// DetectLanguages analyzes the workspace and returns detected languages
func (ld *LanguageDetector) DetectLanguages() ([]DetectedLanguage, error) {
	ld.mutex.Lock()
	defer ld.mutex.Unlock()

	detected := make(map[string]*DetectedLanguage)

	// Walk through the workspace to analyze files
	err := filepath.WalkDir(ld.workspace, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}

		// Skip hidden directories and common ignore patterns
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && d.Name() != ".env" {
				return filepath.SkipDir
			}
			skipDirs := []string{"node_modules", "vendor", "target", "dist", "build", "__pycache__", "venv", "env"}
			for _, skip := range skipDirs {
				if d.Name() == skip {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Analyze files by extension
		ext := strings.ToLower(filepath.Ext(path))
		relPath, _ := filepath.Rel(ld.workspace, path)

		// Check for config files
		filename := strings.ToLower(d.Name())
		for langName, lang := range ld.Languages {
			// Check extensions
			for _, langExt := range lang.Extensions {
				if ext == langExt {
					if detected[langName] == nil {
						detected[langName] = &DetectedLanguage{
							Language:    lang,
							FileCount:   0,
							LineCount:   0,
							MainFiles:   []string{},
							ConfigFiles: []string{},
							Frameworks:  []FrameworkInfo{},
							BuildTools:  []BuildTool{},
						}
					}
					detected[langName].FileCount++
					detected[langName].LineCount += ld.countFileLines(path)

					// Check if it's a main file
					if ld.isMainFile(filename, langName) {
						detected[langName].MainFiles = append(detected[langName].MainFiles, relPath)
					}
					break
				}
			}

			// Check config files
			for _, configFile := range lang.ConfigFiles {
				if filename == strings.ToLower(configFile) {
					if detected[langName] == nil {
						detected[langName] = &DetectedLanguage{
							Language:    lang,
							FileCount:   0,
							LineCount:   0,
							MainFiles:   []string{},
							ConfigFiles: []string{},
							Frameworks:  []FrameworkInfo{},
							BuildTools:  []BuildTool{},
						}
					}
					detected[langName].ConfigFiles = append(detected[langName].ConfigFiles, relPath)

					// Detect frameworks and package managers
					ld.analyzeConfigFile(path, filename, detected[langName])
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	// Convert map to slice and sort by file count
	result := make([]DetectedLanguage, 0, len(detected))
	for _, lang := range detected {
		// Detect testing and linting setup
		lang.HasTests = ld.hasTestFiles(lang)
		lang.HasLinting = ld.hasLintingSetup(lang)
		result = append(result, *lang)
	}

	// Sort by file count (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].FileCount > result[j].FileCount
	})

	return result, nil
}

// countFileLines counts lines in a file
func (ld *LanguageDetector) countFileLines(path string) int {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return len(strings.Split(string(content), "\n"))
}

// isMainFile checks if a file is considered a main entry point
func (ld *LanguageDetector) isMainFile(filename, language string) bool {
	mainFiles := map[string][]string{
		"javascript": {"index.js", "app.js", "main.js", "server.js"},
		"typescript": {"index.ts", "app.ts", "main.ts", "server.ts"},
		"python":     {"main.py", "app.py", "manage.py", "__init__.py"},
		"go":         {"main.go", "cmd.go"},
		"java":       {"Main.java", "Application.java", "App.java"},
		"csharp":     {"Program.cs", "Startup.cs", "Main.cs"},
		"rust":       {"main.rs", "lib.rs"},
		"php":        {"index.php", "app.php", "bootstrap.php"},
		"ruby":       {"app.rb", "main.rb", "config.ru"},
	}

	if mains, exists := mainFiles[language]; exists {
		for _, main := range mains {
			if strings.ToLower(filename) == strings.ToLower(main) {
				return true
			}
		}
	}
	return false
}

// analyzeConfigFile extracts framework and build tool information from config files
func (ld *LanguageDetector) analyzeConfigFile(path, filename string, detected *DetectedLanguage) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	contentStr := string(content)
	relPath, _ := filepath.Rel(ld.workspace, path)

	switch filename {
	case "package.json":
		ld.analyzePackageJSON(contentStr, relPath, detected)
	case "requirements.txt", "pyproject.toml":
		ld.analyzePythonDeps(contentStr, relPath, detected)
	case "go.mod":
		ld.analyzeGoMod(contentStr, relPath, detected)
	case "pom.xml":
		ld.analyzeJavaMaven(contentStr, relPath, detected)
	case "composer.json":
		ld.analyzePHPComposer(contentStr, relPath, detected)
	case "gemfile":
		ld.analyzeRubyGemfile(contentStr, relPath, detected)
	}
}

// analyzePackageJSON extracts JavaScript/TypeScript framework and dependency information
func (ld *LanguageDetector) analyzePackageJSON(content, configFile string, detected *DetectedLanguage) {
	var pkg struct {
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal([]byte(content), &pkg); err != nil {
		return
	}

	// Detect frameworks
	frameworks := map[string]FrameworkInfo{
		"react": {
			Name: "React", Language: "javascript", DefaultPort: 3000,
			DevCommand: "npm start", BuildCommand: "npm run build", TestCommand: "npm test",
		},
		"vue": {
			Name: "Vue.js", Language: "javascript", DefaultPort: 8080,
			DevCommand: "npm run serve", BuildCommand: "npm run build", TestCommand: "npm test",
		},
		"angular": {
			Name: "Angular", Language: "typescript", DefaultPort: 4200,
			DevCommand: "ng serve", BuildCommand: "ng build", TestCommand: "ng test",
		},
		"express": {
			Name: "Express", Language: "javascript", DefaultPort: 3000,
			DevCommand: "npm start", BuildCommand: "npm run build", TestCommand: "npm test",
		},
		"next": {
			Name: "Next.js", Language: "javascript", DefaultPort: 3000,
			DevCommand: "npm run dev", BuildCommand: "npm run build", TestCommand: "npm test",
		},
		"nuxt": {
			Name: "Nuxt.js", Language: "javascript", DefaultPort: 3000,
			DevCommand: "npm run dev", BuildCommand: "npm run build", TestCommand: "npm test",
		},
	}

	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}

	for dep := range allDeps {
		if fw, exists := frameworks[dep]; exists {
			fw.ConfigFile = configFile
			fw.Version = allDeps[dep]
			detected.Frameworks = append(detected.Frameworks, fw)
		}
	}

	// Detect package manager from lock files
	detected.PackageManager = "npm" // default
	if _, err := os.Stat(filepath.Join(ld.workspace, "yarn.lock")); err == nil {
		detected.PackageManager = "yarn"
	} else if _, err := os.Stat(filepath.Join(ld.workspace, "pnpm-lock.yaml")); err == nil {
		detected.PackageManager = "pnpm"
	} else if _, err := os.Stat(filepath.Join(ld.workspace, "bun.lockb")); err == nil {
		detected.PackageManager = "bun"
	}
}

// analyzePythonDeps extracts Python framework information
func (ld *LanguageDetector) analyzePythonDeps(content, configFile string, detected *DetectedLanguage) {
	frameworks := map[string]FrameworkInfo{
		"django": {
			Name: "Django", Language: "python", DefaultPort: 8000,
			DevCommand: "python manage.py runserver", BuildCommand: "", TestCommand: "python manage.py test",
		},
		"flask": {
			Name: "Flask", Language: "python", DefaultPort: 5000,
			DevCommand: "flask run", BuildCommand: "", TestCommand: "python -m pytest",
		},
		"fastapi": {
			Name: "FastAPI", Language: "python", DefaultPort: 8000,
			DevCommand: "uvicorn main:app --reload", BuildCommand: "", TestCommand: "python -m pytest",
		},
	}

	contentLower := strings.ToLower(content)
	for fwName, fw := range frameworks {
		if strings.Contains(contentLower, fwName) {
			fw.ConfigFile = configFile
			detected.Frameworks = append(detected.Frameworks, fw)
		}
	}

	detected.PackageManager = "pip"
	if strings.Contains(configFile, "pyproject.toml") && strings.Contains(content, "[tool.poetry]") {
		detected.PackageManager = "poetry"
	}
}

// analyzeGoMod extracts Go framework information
func (ld *LanguageDetector) analyzeGoMod(content, configFile string, detected *DetectedLanguage) {
	frameworks := map[string]FrameworkInfo{
		"gin": {
			Name: "Gin", Language: "go", DefaultPort: 8080,
			DevCommand: "go run main.go", BuildCommand: "go build", TestCommand: "go test ./...",
		},
		"echo": {
			Name: "Echo", Language: "go", DefaultPort: 8080,
			DevCommand: "go run main.go", BuildCommand: "go build", TestCommand: "go test ./...",
		},
		"fiber": {
			Name: "Fiber", Language: "go", DefaultPort: 3000,
			DevCommand: "go run main.go", BuildCommand: "go build", TestCommand: "go test ./...",
		},
	}

	contentLower := strings.ToLower(content)
	for fwName, fw := range frameworks {
		if strings.Contains(contentLower, fwName) {
			fw.ConfigFile = configFile
			detected.Frameworks = append(detected.Frameworks, fw)
		}
	}

	detected.PackageManager = "go"
}

// analyzeJavaMaven extracts Java framework information from Maven pom.xml
func (ld *LanguageDetector) analyzeJavaMaven(content, configFile string, detected *DetectedLanguage) {
	frameworks := map[string]FrameworkInfo{
		"spring-boot": {
			Name: "Spring Boot", Language: "java", DefaultPort: 8080,
			DevCommand: "mvn spring-boot:run", BuildCommand: "mvn package", TestCommand: "mvn test",
		},
		"quarkus": {
			Name: "Quarkus", Language: "java", DefaultPort: 8080,
			DevCommand: "mvn quarkus:dev", BuildCommand: "mvn package", TestCommand: "mvn test",
		},
	}

	contentLower := strings.ToLower(content)
	for fwName, fw := range frameworks {
		if strings.Contains(contentLower, fwName) {
			fw.ConfigFile = configFile
			detected.Frameworks = append(detected.Frameworks, fw)
		}
	}

	detected.PackageManager = "maven"
}

// analyzePHPComposer extracts PHP framework information
func (ld *LanguageDetector) analyzePHPComposer(content, configFile string, detected *DetectedLanguage) {
	frameworks := map[string]FrameworkInfo{
		"laravel": {
			Name: "Laravel", Language: "php", DefaultPort: 8000,
			DevCommand: "php artisan serve", BuildCommand: "", TestCommand: "php artisan test",
		},
		"symfony": {
			Name: "Symfony", Language: "php", DefaultPort: 8000,
			DevCommand: "symfony server:start", BuildCommand: "", TestCommand: "php bin/phpunit",
		},
	}

	contentLower := strings.ToLower(content)
	for fwName, fw := range frameworks {
		if strings.Contains(contentLower, fwName) {
			fw.ConfigFile = configFile
			detected.Frameworks = append(detected.Frameworks, fw)
		}
	}

	detected.PackageManager = "composer"
}

// analyzeRubyGemfile extracts Ruby framework information
func (ld *LanguageDetector) analyzeRubyGemfile(content, configFile string, detected *DetectedLanguage) {
	frameworks := map[string]FrameworkInfo{
		"rails": {
			Name: "Ruby on Rails", Language: "ruby", DefaultPort: 3000,
			DevCommand: "rails server", BuildCommand: "", TestCommand: "rails test",
		},
		"sinatra": {
			Name: "Sinatra", Language: "ruby", DefaultPort: 4567,
			DevCommand: "ruby app.rb", BuildCommand: "", TestCommand: "ruby test.rb",
		},
	}

	contentLower := strings.ToLower(content)
	for fwName, fw := range frameworks {
		if strings.Contains(contentLower, fwName) {
			fw.ConfigFile = configFile
			detected.Frameworks = append(detected.Frameworks, fw)
		}
	}

	detected.PackageManager = "bundler"
}

// hasTestFiles checks if the project has test files
func (ld *LanguageDetector) hasTestFiles(detected *DetectedLanguage) bool {
	testPatterns := []string{
		"test", "tests", "spec", "specs", "__tests__",
		".test.", ".spec.", "_test.", "_spec.",
	}

	err := filepath.WalkDir(ld.workspace, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		name := strings.ToLower(d.Name())
		for _, pattern := range testPatterns {
			if strings.Contains(name, pattern) {
				return fmt.Errorf("found") // Use error to break early
			}
		}
		return nil
	})

	return err != nil // If we found tests, err will not be nil
}

// hasLintingSetup checks if the project has linting configuration
func (ld *LanguageDetector) hasLintingSetup(detected *DetectedLanguage) bool {
	for _, tool := range detected.Language.LintTools {
		for _, configFile := range detected.ConfigFiles {
			if strings.Contains(strings.ToLower(configFile), strings.ToLower(tool)) {
				return true
			}
		}
	}
	return false
}

// GetProjectTopology returns the complete project topology
func (ld *LanguageDetector) GetProjectTopology() (*ProjectTopology, error) {
	languages, err := ld.DetectLanguages()
	if err != nil {
		return nil, err
	}

	topology := &ProjectTopology{
		Languages: languages,
		Services:  []ServiceInfo{},
		Databases: []DatabaseInfo{},
		APIs:      []APIEndpoint{},
		Relations: []ServiceRelation{},
		UpdatedAt: time.Now(),
	}

	// TODO: Implement service discovery, API endpoint detection, and relationship mapping
	// This will be expanded in future phases

	return topology, nil
}
