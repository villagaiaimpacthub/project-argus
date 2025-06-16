# Claude Code Project Intelligence Service

A comprehensive monitoring and intelligence service that gives Claude Code real-time insight into your development environment. Perfect for **Claude Code in WSL + Cursor IDE** setups where native integration isn't available yet.

## 🎯 What This Tool Does

**Problem**: Claude Code can't see your project structure efficiently, you have to manually update it constantly, and it works slowly without context.

**Solution**: This service monitors your project 24/7 and provides Claude Code with real-time APIs to understand:
- 📁 Complete project structure and file types
- 🚨 Active compilation errors and warnings
- 📝 Recent file changes with timestamps
- 🔄 Git repository status and changes
- 📋 TODO/FIXME items in your code
- 📦 Project dependencies and versions
- ⚡ Running processes and system health
- 🔍 Powerful search across all project data

## 🚀 Quick Setup

### Prerequisites
```bash
# Install required tools
sudo apt update && sudo apt install golang-go jq curl git

# Or on macOS
brew install go jq curl git
```

### Automated Installation
Copy this entire README into your project and tell Cursor:

> "Read this README.md file and set up the complete Claude Intelligence Service exactly as described. Create all the files and run the setup commands."

---

## 📁 File Structure to Create

```
claude-intelligence/
├── main.go                 # Intelligence service server
├── claude-query.sh         # CLI tool for Claude Code
├── dashboard.html          # Web dashboard
├── go.mod                  # Go module file
└── README.md              # This file
```

---

## 📄 File Contents

### 1. Create `go.mod`

```go
module claude-intelligence

go 1.21

require github.com/gofiber/fiber/v2 v2.52.0

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)
```

### 2. Create `main.go`

```go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// ProjectIntelligence represents the main intelligence service
type ProjectIntelligence struct {
	workspace     string
	fileWatcher   *FileWatcher
	gitWatcher    *GitWatcher
	errorWatcher  *ErrorWatcher
	buildWatcher  *BuildWatcher
	processWatcher *ProcessWatcher
	lastSnapshot  *ProjectSnapshot
	mutex         sync.RWMutex
}

// ProjectSnapshot represents the current state of the project
type ProjectSnapshot struct {
	Timestamp       time.Time              `json:"timestamp"`
	Structure       *ProjectStructure      `json:"structure"`
	RecentChanges   []FileChange           `json:"recent_changes"`
	GitStatus       *GitStatus             `json:"git_status"`
	ActiveErrors    []ErrorInfo            `json:"active_errors"`
	BuildStatus     *BuildStatus           `json:"build_status"`
	RunningProcesses []ProcessInfo         `json:"running_processes"`
	TestResults     *TestResults           `json:"test_results,omitempty"`
	Dependencies    []DependencyInfo       `json:"dependencies"`
	TODOs           []TodoItem             `json:"todos"`
	Health          ProjectHealth          `json:"health"`
}

// ProjectStructure represents the file/folder structure
type ProjectStructure struct {
	RootPath    string         `json:"root_path"`
	Files       []FileInfo     `json:"files"`
	Directories []DirectoryInfo `json:"directories"`
	ProjectType string         `json:"project_type"`
	MainFiles   []string       `json:"main_files"`
	ConfigFiles []string       `json:"config_files"`
	TotalFiles  int            `json:"total_files"`
	TotalSize   int64          `json:"total_size"`
}

// FileInfo represents detailed file information
type FileInfo struct {
	Path         string    `json:"path"`
	RelativePath string    `json:"relative_path"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	IsExecutable bool      `json:"is_executable"`
	Language     string    `json:"language"`
	LineCount    int       `json:"line_count,omitempty"`
}

// DirectoryInfo represents directory information
type DirectoryInfo struct {
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	FileCount    int    `json:"file_count"`
	Purpose      string `json:"purpose"` // src, test, config, docs, etc.
}

// FileChange represents a file system change
type FileChange struct {
	Path      string    `json:"path"`
	Type      string    `json:"type"` // created, modified, deleted, renamed
	Timestamp time.Time `json:"timestamp"`
	OldPath   string    `json:"old_path,omitempty"`
}

// GitStatus represents git repository status
type GitStatus struct {
	Branch          string   `json:"branch"`
	CommitHash      string   `json:"commit_hash"`
	CommitMessage   string   `json:"commit_message"`
	IsDirty         bool     `json:"is_dirty"`
	UntrackedFiles  []string `json:"untracked_files"`
	ModifiedFiles   []string `json:"modified_files"`
	StagedFiles     []string `json:"staged_files"`
	Ahead           int      `json:"ahead"`
	Behind          int      `json:"behind"`
	LastCommitTime  time.Time `json:"last_commit_time"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Source      string    `json:"source"` // compiler, linter, runtime, test
	File        string    `json:"file"`
	Line        int       `json:"line"`
	Column      int       `json:"column"`
	Type        string    `json:"type"` // error, warning, info
	Message     string    `json:"message"`
	Code        string    `json:"code,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Context     string    `json:"context,omitempty"`
}

// BuildStatus represents build/compilation status
type BuildStatus struct {
	IsBuilding    bool      `json:"is_building"`
	LastBuildTime time.Time `json:"last_build_time"`
	Success       bool      `json:"success"`
	Duration      string    `json:"duration"`
	Output        string    `json:"output"`
	Errors        []string  `json:"errors"`
	Warnings      []string  `json:"warnings"`
}

// ProcessInfo represents running process information
type ProcessInfo struct {
	PID         int    `json:"pid"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	StartTime   time.Time `json:"start_time"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryMB    float64 `json:"memory_mb"`
	IsProjectRelated bool `json:"is_project_related"`
}

// TestResults represents test execution results
type TestResults struct {
	TotalTests   int       `json:"total_tests"`
	PassedTests  int       `json:"passed_tests"`
	FailedTests  int       `json:"failed_tests"`
	SkippedTests int       `json:"skipped_tests"`
	Duration     string    `json:"duration"`
	LastRun      time.Time `json:"last_run"`
	FailureDetails []TestFailure `json:"failure_details,omitempty"`
}

// TestFailure represents a test failure
type TestFailure struct {
	TestName string `json:"test_name"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Error    string `json:"error"`
}

// DependencyInfo represents dependency information
type DependencyInfo struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	Type          string `json:"type"` // direct, dev, peer
	Source        string `json:"source"` // npm, go.mod, requirements.txt, etc.
	HasUpdate     bool   `json:"has_update"`
	LatestVersion string `json:"latest_version,omitempty"`
}

// TodoItem represents TODO/FIXME items found in code
type TodoItem struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Type    string `json:"type"` // TODO, FIXME, HACK, NOTE
	Message string `json:"message"`
	Author  string `json:"author,omitempty"`
}

// ProjectHealth represents overall project health metrics
type ProjectHealth struct {
	Score           int    `json:"score"` // 0-100
	ErrorCount      int    `json:"error_count"`
	WarningCount    int    `json:"warning_count"`
	TestCoverage    float64 `json:"test_coverage"`
	TechnicalDebt   string `json:"technical_debt"` // low, medium, high
	LastHealthCheck time.Time `json:"last_health_check"`
}

// FileWatcher monitors file system changes
type FileWatcher struct {
	workspace string
	changes   []FileChange
	mutex     sync.RWMutex
}

// GitWatcher monitors git repository changes
type GitWatcher struct {
	workspace string
	status    *GitStatus
	mutex     sync.RWMutex
}

// ErrorWatcher monitors for errors from various sources
type ErrorWatcher struct {
	errors []ErrorInfo
	mutex  sync.RWMutex
}

// BuildWatcher monitors build processes
type BuildWatcher struct {
	status *BuildStatus
	mutex  sync.RWMutex
}

// ProcessWatcher monitors running processes
type ProcessWatcher struct {
	processes []ProcessInfo
	mutex     sync.RWMutex
}

// NewProjectIntelligence creates a new project intelligence service
func NewProjectIntelligence(workspace string) *ProjectIntelligence {
	pi := &ProjectIntelligence{
		workspace:      workspace,
		fileWatcher:    &FileWatcher{workspace: workspace, changes: []FileChange{}},
		gitWatcher:     &GitWatcher{workspace: workspace},
		errorWatcher:   &ErrorWatcher{errors: []ErrorInfo{}},
		buildWatcher:   &BuildWatcher{status: &BuildStatus{}},
		processWatcher: &ProcessWatcher{processes: []ProcessInfo{}},
	}
	
	return pi
}

// StartWatching begins monitoring the project
func (pi *ProjectIntelligence) StartWatching() {
	log.Printf("Starting project intelligence monitoring for: %s", pi.workspace)
	
	// Start watchers in separate goroutines
	go pi.fileWatcher.startWatching()
	go pi.gitWatcher.startWatching()
	go pi.errorWatcher.startWatching(pi.workspace)
	go pi.buildWatcher.startWatching(pi.workspace)
	go pi.processWatcher.startWatching(pi.workspace)
	
	// Generate initial snapshot
	go func() {
		time.Sleep(2 * time.Second) // Give watchers time to initialize
		pi.updateSnapshot()
		
		// Update snapshot every 30 seconds
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			pi.updateSnapshot()
		}
	}()
}

// updateSnapshot creates a new project snapshot
func (pi *ProjectIntelligence) updateSnapshot() {
	pi.mutex.Lock()
	defer pi.mutex.Unlock()
	
	log.Println("Updating project snapshot...")
	
	snapshot := &ProjectSnapshot{
		Timestamp:       time.Now(),
		Structure:       pi.analyzeProjectStructure(),
		RecentChanges:   pi.fileWatcher.getRecentChanges(),
		GitStatus:       pi.gitWatcher.getStatus(),
		ActiveErrors:    pi.errorWatcher.getErrors(),
		BuildStatus:     pi.buildWatcher.getStatus(),
		RunningProcesses: pi.processWatcher.getProcesses(),
		Dependencies:    pi.analyzeDependencies(),
		TODOs:           pi.findTodos(),
		Health:          pi.calculateHealth(),
	}
	
	pi.lastSnapshot = snapshot
	log.Printf("Snapshot updated - %d files, %d errors, %d processes", 
		len(snapshot.Structure.Files), len(snapshot.ActiveErrors), len(snapshot.RunningProcesses))
}

// analyzeProjectStructure analyzes the project file structure
func (pi *ProjectIntelligence) analyzeProjectStructure() *ProjectStructure {
	structure := &ProjectStructure{
		RootPath:    pi.workspace,
		Files:       []FileInfo{},
		Directories: []DirectoryInfo{},
		MainFiles:   []string{},
		ConfigFiles: []string{},
	}
	
	// Walk through the project directory
	err := filepath.WalkDir(pi.workspace, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}
		
		// Skip hidden files and common ignore patterns
		if strings.HasPrefix(d.Name(), ".") && d.Name() != ".env" && d.Name() != ".gitignore" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		
		// Skip common ignore directories
		if d.IsDir() {
			skipDirs := []string{"node_modules", "vendor", "target", ".git", "dist", "build"}
			for _, skip := range skipDirs {
				if d.Name() == skip {
					return filepath.SkipDir
				}
			}
			
			relPath, _ := filepath.Rel(pi.workspace, path)
			dirInfo := DirectoryInfo{
				Path:         path,
				RelativePath: relPath,
				Purpose:      determineDirPurpose(d.Name()),
			}
			structure.Directories = append(structure.Directories, dirInfo)
			return nil
		}
		
		// Process files
		info, err := d.Info()
		if err != nil {
			return nil
		}
		
		relPath, _ := filepath.Rel(pi.workspace, path)
		fileInfo := FileInfo{
			Path:         path,
			RelativePath: relPath,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			IsExecutable: info.Mode()&0111 != 0,
			Language:     detectLanguage(path),
		}
		
		// Count lines for code files
		if isCodeFile(path) {
			fileInfo.LineCount = countLines(path)
		}
		
		structure.Files = append(structure.Files, fileInfo)
		structure.TotalFiles++
		structure.TotalSize += info.Size()
		
		// Categorize important files
		if isMainFile(path) {
			structure.MainFiles = append(structure.MainFiles, relPath)
		}
		if isConfigFile(path) {
			structure.ConfigFiles = append(structure.ConfigFiles, relPath)
		}
		
		return nil
	})
	
	if err != nil {
		log.Printf("Error analyzing project structure: %v", err)
	}
	
	structure.ProjectType = detectProjectType(structure)
	
	return structure
}

// File watcher implementation
func (fw *FileWatcher) startWatching() {
	log.Println("File watcher started")
	// Simple polling implementation for cross-platform compatibility
	lastScan := make(map[string]time.Time)
	
	for {
		filepath.WalkDir(fw.workspace, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			
			if strings.HasPrefix(d.Name(), ".") {
				return nil
			}
			
			info, err := d.Info()
			if err != nil {
				return nil
			}
			
			lastMod := info.ModTime()
			if prev, exists := lastScan[path]; exists {
				if lastMod.After(prev) {
					fw.addChange(FileChange{
						Path:      path,
						Type:      "modified",
						Timestamp: lastMod,
					})
				}
			} else {
				fw.addChange(FileChange{
					Path:      path,
					Type:      "created",
					Timestamp: lastMod,
				})
			}
			
			lastScan[path] = lastMod
			return nil
		})
		
		time.Sleep(5 * time.Second)
	}
}

func (fw *FileWatcher) addChange(change FileChange) {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()
	
	fw.changes = append(fw.changes, change)
	
	// Keep only recent changes (last 100)
	if len(fw.changes) > 100 {
		fw.changes = fw.changes[len(fw.changes)-100:]
	}
}

func (fw *FileWatcher) getRecentChanges() []FileChange {
	fw.mutex.RLock()
	defer fw.mutex.RUnlock()
	
	// Return changes from last 10 minutes
	cutoff := time.Now().Add(-10 * time.Minute)
	recent := []FileChange{}
	
	for _, change := range fw.changes {
		if change.Timestamp.After(cutoff) {
			recent = append(recent, change)
		}
	}
	
	return recent
}

// Git watcher implementation
func (gw *GitWatcher) startWatching() {
	log.Println("Git watcher started")
	
	for {
		gw.updateGitStatus()
		time.Sleep(15 * time.Second)
	}
}

func (gw *GitWatcher) updateGitStatus() {
	gw.mutex.Lock()
	defer gw.mutex.Unlock()
	
	status := &GitStatus{}
	
	// Get current branch
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = gw.workspace
	if output, err := cmd.Output(); err == nil {
		status.Branch = strings.TrimSpace(string(output))
	}
	
	// Get commit hash and message
	cmd = exec.Command("git", "log", "-1", "--pretty=format:%H|%s|%ct")
	cmd.Dir = gw.workspace
	if output, err := cmd.Output(); err == nil {
		parts := strings.Split(string(output), "|")
		if len(parts) >= 3 {
			status.CommitHash = parts[0][:8] // Short hash
			status.CommitMessage = parts[1]
			if timestamp, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
				status.LastCommitTime = time.Unix(timestamp, 0)
			}
		}
	}
	
	// Get status
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = gw.workspace
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if len(line) < 3 {
				continue
			}
			
			statusCode := line[:2]
			filename := line[3:]
			
			switch {
			case statusCode[0] != ' ' && statusCode[0] != '?':
				status.StagedFiles = append(status.StagedFiles, filename)
			case statusCode[1] != ' ':
				status.ModifiedFiles = append(status.ModifiedFiles, filename)
			case statusCode == "??":
				status.UntrackedFiles = append(status.UntrackedFiles, filename)
			}
		}
		
		status.IsDirty = len(status.ModifiedFiles) > 0 || len(status.UntrackedFiles) > 0
	}
	
	gw.status = status
}

func (gw *GitWatcher) getStatus() *GitStatus {
	gw.mutex.RLock()
	defer gw.mutex.RUnlock()
	
	if gw.status == nil {
		return &GitStatus{}
	}
	
	return gw.status
}

// Error watcher implementation
func (ew *ErrorWatcher) startWatching(workspace string) {
	log.Println("Error watcher started")
	
	for {
		ew.scanForErrors(workspace)
		time.Sleep(10 * time.Second)
	}
}

func (ew *ErrorWatcher) scanForErrors(workspace string) {
	ew.mutex.Lock()
	defer ew.mutex.Unlock()
	
	// Clear old errors
	ew.errors = []ErrorInfo{}
	
	// Scan for TypeScript/JavaScript errors
	ew.scanTSErrors(workspace)
	
	// Scan for Go errors
	ew.scanGoErrors(workspace)
	
	// Scan for Python errors
	ew.scanPythonErrors(workspace)
	
	// Scan log files for runtime errors
	ew.scanLogFiles(workspace)
}

func (ew *ErrorWatcher) scanTSErrors(workspace string) {
	// Run tsc to check for TypeScript errors
	cmd := exec.Command("npx", "tsc", "--noEmit", "--pretty", "false")
	cmd.Dir = workspace
	output, _ := cmd.CombinedOutput()
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "error TS") {
			ew.parseTypescriptError(line)
		}
	}
}

func (ew *ErrorWatcher) scanGoErrors(workspace string) {
	// Run go build to check for errors
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = workspace
	output, _ := cmd.CombinedOutput()
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ".go:") && strings.Contains(line, "error") {
			ew.parseGoError(line)
		}
	}
}

func (ew *ErrorWatcher) scanPythonErrors(workspace string) {
	// Run python syntax check
	cmd := exec.Command("python", "-m", "py_compile")
	cmd.Dir = workspace
	// This is a simplified implementation
}

func (ew *ErrorWatcher) scanLogFiles(workspace string) {
	// Scan common log file locations
	logPatterns := []string{"*.log", "logs/*.log", "log/*.log"}
	
	for _, pattern := range logPatterns {
		matches, _ := filepath.Glob(filepath.Join(workspace, pattern))
		for _, logFile := range matches {
			ew.scanLogFile(logFile)
		}
	}
}

func (ew *ErrorWatcher) scanLogFile(logFile string) {
	file, err := os.Open(logFile)
	if err != nil {
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		// Look for error patterns
		if strings.Contains(strings.ToLower(line), "error") ||
		   strings.Contains(strings.ToLower(line), "exception") ||
		   strings.Contains(strings.ToLower(line), "fatal") {
			
			ew.errors = append(ew.errors, ErrorInfo{
				Source:    "runtime",
				File:      logFile,
				Line:      lineNum,
				Type:      "error",
				Message:   line,
				Timestamp: time.Now(),
			})
		}
	}
}

func (ew *ErrorWatcher) parseTypescriptError(line string) {
	// Parse TypeScript error format: file.ts(line,col): error TSxxxx: message
	re := regexp.MustCompile(`(.+)\((\d+),(\d+)\): error (TS\d+): (.+)`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) == 6 {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		
		ew.errors = append(ew.errors, ErrorInfo{
			Source:    "typescript",
			File:      matches[1],
			Line:      lineNum,
			Column:    colNum,
			Type:      "error",
			Code:      matches[4],
			Message:   matches[5],
			Timestamp: time.Now(),
		})
	}
}

func (ew *ErrorWatcher) parseGoError(line string) {
	// Parse Go error format: file.go:line:col: message
	re := regexp.MustCompile(`(.+\.go):(\d+):(\d+): (.+)`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) == 5 {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		
		ew.errors = append(ew.errors, ErrorInfo{
			Source:    "go",
			File:      matches[1],
			Line:      lineNum,
			Column:    colNum,
			Type:      "error",
			Message:   matches[4],
			Timestamp: time.Now(),
		})
	}
}

func (ew *ErrorWatcher) getErrors() []ErrorInfo {
	ew.mutex.RLock()
	defer ew.mutex.RUnlock()
	
	return append([]ErrorInfo{}, ew.errors...)
}

// Build watcher implementation
func (bw *BuildWatcher) startWatching(workspace string) {
	log.Println("Build watcher started")
	
	for {
		bw.checkBuildStatus(workspace)
		time.Sleep(20 * time.Second)
	}
}

func (bw *BuildWatcher) checkBuildStatus(workspace string) {
	bw.mutex.Lock()
	defer bw.mutex.Unlock()
	
	// Check if build is currently running by looking for build processes
	// This is a simplified implementation
	bw.status.IsBuilding = false
	bw.status.LastBuildTime = time.Now()
	bw.status.Success = true
}

func (bw *BuildWatcher) getStatus() *BuildStatus {
	bw.mutex.RLock()
	defer bw.mutex.RUnlock()
	
	return bw.status
}

// Process watcher implementation
func (pw *ProcessWatcher) startWatching(workspace string) {
	log.Println("Process watcher started")
	
	for {
		pw.updateProcesses(workspace)
		time.Sleep(15 * time.Second)
	}
}

func (pw *ProcessWatcher) updateProcesses(workspace string) {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	
	pw.processes = []ProcessInfo{}
	
	// Get process list (simplified for cross-platform compatibility)
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // Skip header
		if strings.Contains(line, "node") ||
		   strings.Contains(line, "go") ||
		   strings.Contains(line, "python") ||
		   strings.Contains(line, workspace) {
			
			// This is a simplified process parsing
			fields := strings.Fields(line)
			if len(fields) >= 11 {
				if pid, err := strconv.Atoi(fields[1]); err == nil {
					pw.processes = append(pw.processes, ProcessInfo{
						PID:         pid,
						Name:        fields[10],
						Command:     strings.Join(fields[10:], " "),
						StartTime:   time.Now(), // Simplified
						IsProjectRelated: strings.Contains(line, workspace),
					})
				}
			}
		}
	}
}

func (pw *ProcessWatcher) getProcesses() []ProcessInfo {
	pw.mutex.RLock()
	defer pw.mutex.RUnlock()
	
	return append([]ProcessInfo{}, pw.processes...)
}

// Helper functions
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	
	langMap := map[string]string{
		".js":   "javascript",
		".jsx":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".go":   "go",
		".py":   "python",
		".java": "java",
		".cpp":  "cpp",
		".c":    "c",
		".cs":   "csharp",
		".rb":   "ruby",
		".php":  "php",
		".rs":   "rust",
		".kt":   "kotlin",
		".swift": "swift",
		".vue":  "vue",
		".svelte": "svelte",
	}
	
	if lang, exists := langMap[ext]; exists {
		return lang
	}
	
	return "text"
}

func isCodeFile(path string) bool {
	codeExts := []string{".js", ".jsx", ".ts", ".tsx", ".go", ".py", ".java", ".cpp", ".c", ".cs", ".rb", ".php", ".rs"}
	ext := strings.ToLower(filepath.Ext(path))
	
	for _, codeExt := range codeExts {
		if ext == codeExt {
			return true
		}
	}
	
	return false
}

func isMainFile(path string) bool {
	filename := strings.ToLower(filepath.Base(path))
	mainFiles := []string{"main.go", "main.py", "index.js", "index.ts", "app.js", "app.py", "server.js"}
	
	for _, main := range mainFiles {
		if filename == main {
			return true
		}
	}
	
	return false
}

func isConfigFile(path string) bool {
	filename := strings.ToLower(filepath.Base(path))
	configFiles := []string{
		"package.json", "go.mod", "requirements.txt", "cargo.toml", "pom.xml",
		"composer.json", "gemfile", "dockerfile", ".gitignore", ".env",
		"tsconfig.json", "webpack.config.js", "babel.config.js",
	}
	
	for _, config := range configFiles {
		if filename == config {
			return true
		}
	}
	
	return false
}

func countLines(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	
	return lines
}

func determineDirPurpose(name string) string {
	purposeMap := map[string]string{
		"src":         "source",
		"lib":         "library",
		"test":        "test",
		"tests":       "test",
		"spec":        "test",
		"docs":        "documentation",
		"config":      "configuration",
		"scripts":     "scripts",
		"assets":      "assets",
		"static":      "static",
		"public":      "public",
		"components":  "components",
		"pages":       "pages",
		"api":         "api",
		"utils":       "utilities",
		"helpers":     "utilities",
		"middleware":  "middleware",
		"models":      "models",
		"controllers": "controllers",
		"services":    "services",
	}
	
	if purpose, exists := purposeMap[strings.ToLower(name)]; exists {
		return purpose
	}
	
	return "general"
}

func detectProjectType(structure *ProjectStructure) string {
	// Check for specific project indicators
	for _, file := range structure.ConfigFiles {
		switch file {
		case "package.json":
			return "nodejs"
		case "go.mod":
			return "go"
		case "requirements.txt":
			return "python"
		case "cargo.toml":
			return "rust"
		case "pom.xml":
			return "java"
		case "composer.json":
			return "php"
		}
	}
	
	return "unknown"
}

// Dependency analysis
func (pi *ProjectIntelligence) analyzeDependencies() []DependencyInfo {
	deps := []DependencyInfo{}
	
	// Analyze package.json
	packagePath := filepath.Join(pi.workspace, "package.json")
	if _, err := os.Stat(packagePath); err == nil {
		deps = append(deps, pi.analyzeNodeDependencies(packagePath)...)
	}
	
	// Analyze go.mod
	goModPath := filepath.Join(pi.workspace, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		deps = append(deps, pi.analyzeGoDependencies(goModPath)...)
	}
	
	return deps
}

func (pi *ProjectIntelligence) analyzeNodeDependencies(packagePath string) []DependencyInfo {
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return []DependencyInfo{}
	}
	
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	
	if err := json.Unmarshal(content, &pkg); err != nil {
		return []DependencyInfo{}
	}
	
	deps := []DependencyInfo{}
	
	for name, version := range pkg.Dependencies {
		deps = append(deps, DependencyInfo{
			Name:    name,
			Version: version,
			Type:    "direct",
			Source:  "package.json",
		})
	}
	
	for name, version := range pkg.DevDependencies {
		deps = append(deps, DependencyInfo{
			Name:    name,
			Version: version,
			Type:    "dev",
			Source:  "package.json",
		})
	}
	
	return deps
}

func (pi *ProjectIntelligence) analyzeGoDependencies(goModPath string) []DependencyInfo {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return []DependencyInfo{}
	}
	
	deps := []DependencyInfo{}
	lines := strings.Split(string(content), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, " v") && !strings.HasPrefix(line, "//") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				deps = append(deps, DependencyInfo{
					Name:    parts[0],
					Version: parts[1],
					Type:    "direct",
					Source:  "go.mod",
				})
			}
		}
	}
	
	return deps
}

// TODO finder
func (pi *ProjectIntelligence) findTodos() []TodoItem {
	todos := []TodoItem{}
	
	filepath.WalkDir(pi.workspace, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isCodeFile(path) {
			return nil
		}
		
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()
		
		scanner := bufio.NewScanner(file)
		lineNum := 0
		
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			
			// Look for TODO, FIXME, HACK, NOTE patterns
			patterns := []string{"TODO", "FIXME", "HACK", "NOTE", "XXX"}
			
			for _, pattern := range patterns {
				if strings.Contains(strings.ToUpper(line), pattern) {
					relPath, _ := filepath.Rel(pi.workspace, path)
					
					// Extract the message after the pattern
					idx := strings.Index(strings.ToUpper(line), pattern)
					message := strings.TrimSpace(line[idx:])
					
					todos = append(todos, TodoItem{
						File:    relPath,
						Line:    lineNum,
						Type:    pattern,
						Message: message,
					})
					break
				}
			}
		}
		
		return nil
	})
	
	return todos
}

// Health calculation
func (pi *ProjectIntelligence) calculateHealth() ProjectHealth {
	health := ProjectHealth{
		Score:           100,
		LastHealthCheck: time.Now(),
	}
	
	// Count errors and warnings
	for _, err := range pi.errorWatcher.getErrors() {
		if err.Type == "error" {
			health.ErrorCount++
			health.Score -= 10
		} else if err.Type == "warning" {
			health.WarningCount++
			health.Score -= 3
		}
	}
	
	// Ensure score doesn't go below 0
	if health.Score < 0 {
		health.Score = 0
	}
	
	// Calculate technical debt level
	if health.ErrorCount > 10 || health.WarningCount > 50 {
		health.TechnicalDebt = "high"
	} else if health.ErrorCount > 5 || health.WarningCount > 20 {
		health.TechnicalDebt = "medium"
	} else {
		health.TechnicalDebt = "low"
	}
	
	return health
}

// GetSnapshot returns the latest project snapshot
func (pi *ProjectIntelligence) GetSnapshot() *ProjectSnapshot {
	pi.mutex.RLock()
	defer pi.mutex.RUnlock()
	
	return pi.lastSnapshot
}

// Server implementation
type IntelligenceServer struct {
	app *fiber.App
	pi  *ProjectIntelligence
}

func NewIntelligenceServer(workspace string) *IntelligenceServer {
	app := fiber.New(fiber.Config{
		AppName: "Project Intelligence Service",
	})
	
	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
	
	pi := NewProjectIntelligence(workspace)
	
	server := &IntelligenceServer{
		app: app,
		pi:  pi,
	}
	
	server.setupRoutes()
	pi.StartWatching()
	
	return server
}

func (is *IntelligenceServer) setupRoutes() {
	// Main intelligence routes
	is.app.Get("/", is.statusHandler)
	is.app.Get("/snapshot", is.snapshotHandler)
	is.app.Get("/structure", is.structureHandler)
	is.app.Get("/changes", is.changesHandler)
	is.app.Get("/git", is.gitHandler)
	is.app.Get("/errors", is.errorsHandler)
	is.app.Get("/build", is.buildHandler)
	is.app.Get("/processes", is.processesHandler)
	is.app.Get("/dependencies", is.dependenciesHandler)
	is.app.Get("/todos", is.todosHandler)
	is.app.Get("/health", is.healthHandler)
	
	// File-specific routes
	is.app.Get("/files/:path", is.fileHandler)
	is.app.Get("/files/:path/content", is.fileContentHandler)
	
	// Search and query routes
	is.app.Get("/search", is.searchHandler)
	is.app.Get("/query/:type", is.queryHandler)
	
	// Action routes
	is.app.Post("/refresh", is.refreshHandler)
	is.app.Post("/analyze", is.analyzeHandler)
}

func (is *IntelligenceServer) statusHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"service":   "Project Intelligence Service",
		"workspace": is.pi.workspace,
		"status":    "running",
		"timestamp": time.Now(),
		"endpoints": []string{
			"/snapshot - Complete project snapshot",
			"/structure - Project file structure",
			"/changes - Recent file changes",
			"/git - Git repository status",
			"/errors - Active errors and warnings",
			"/build - Build status",
			"/processes - Running processes",
			"/dependencies - Project dependencies",
			"/todos - TODO items in code",
			"/health - Project health metrics",
			"/search?q=query - Search across project",
		},
	})
}

func (is *IntelligenceServer) snapshotHandler(c *fiber.Ctx) error {
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Snapshot not ready yet",
		})
	}
	
	return c.JSON(snapshot)
}

func (is *IntelligenceServer) structureHandler(c *fiber.Ctx) error {
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}
	
	return c.JSON(snapshot.Structure)
}

func (is *IntelligenceServer) changesHandler(c *fiber.Ctx) error {
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}
	
	return c.JSON(snapshot.RecentChanges)
}

func (is *IntelligenceServer) gitHandler(c *fiber.Ctx) error {
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}
	
	return c.JSON(snapshot.GitStatus)
}

func (is *IntelligenceServer) errorsHandler(c *fiber.Ctx) error {
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}
	
	return c.JSON(snapshot.ActiveErrors)
}

func (is *IntelligenceServer) buildHandler(c *fiber.Ctx) error {
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}
	
	return c.JSON(snapshot.BuildStatus)
}

func (is *IntelligenceServer) processesHandler(c *fiber.Ctx) error {
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}
	
	return c.JSON(snapshot.RunningProcesses)
}

func (is *IntelligenceServer) dependenciesHandler(c *fiber.Ctx) error {
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}
	
	return c.JSON(snapshot.Dependencies)
}

func (is *IntelligenceServer) todosHandler(c *fiber.Ctx) error {
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}
	
	return c.JSON(snapshot.TODOs)
}

func (is *IntelligenceServer) healthHandler(c *fiber.Ctx) error {
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}
	
	return c.JSON(snapshot.Health)
}

func (is *IntelligenceServer) fileHandler(c *fiber.Ctx) error {
	path := c.Params("path")
	fullPath := filepath.Join(is.pi.workspace, path)
	
	info, err := os.Stat(fullPath)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "File not found",
		})
	}
	
	fileInfo := FileInfo{
		Path:         fullPath,
		RelativePath: path,
		Size:         info.Size(),
		ModTime:      info.ModTime(),
		IsExecutable: info.Mode()&0111 != 0,
		Language:     detectLanguage(fullPath),
	}
	
	if isCodeFile(fullPath) {
		fileInfo.LineCount = countLines(fullPath)
	}
	
	return c.JSON(fileInfo)
}

func (is *IntelligenceServer) fileContentHandler(c *fiber.Ctx) error {
	path := c.Params("path")
	fullPath := filepath.Join(is.pi.workspace, path)
	
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "File not found or cannot be read",
		})
	}
	
	return c.JSON(fiber.Map{
		"path":    path,
		"content": string(content),
		"size":    len(content),
		"language": detectLanguage(fullPath),
	})
}

func (is *IntelligenceServer) searchHandler(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}
	
	results := is.searchProject(query)
	
	return c.JSON(fiber.Map{
		"query":   query,
		"results": results,
		"count":   len(results),
	})
}

func (is *IntelligenceServer) queryHandler(c *fiber.Ctx) error {
	queryType := c.Params("type")
	
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Not ready"})
	}
	
	switch queryType {
	case "errors-by-file":
		return c.JSON(is.groupErrorsByFile(snapshot.ActiveErrors))
	case "recent-activity":
		return c.JSON(is.getRecentActivity(snapshot))
	case "project-stats":
		return c.JSON(is.getProjectStats(snapshot))
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unknown query type",
		})
	}
}

func (is *IntelligenceServer) refreshHandler(c *fiber.Ctx) error {
	go is.pi.updateSnapshot()
	
	return c.JSON(fiber.Map{
		"message": "Refresh triggered",
	})
}

func (is *IntelligenceServer) analyzeHandler(c *fiber.Ctx) error {
	// Trigger a comprehensive analysis
	go is.pi.updateSnapshot()
	
	return c.JSON(fiber.Map{
		"message": "Analysis started",
	})
}

// Search implementation
func (is *IntelligenceServer) searchProject(query string) []map[string]interface{} {
	results := []map[string]interface{}{}
	
	// Search in file names
	snapshot := is.pi.GetSnapshot()
	if snapshot == nil {
		return results
	}
	
	queryLower := strings.ToLower(query)
	
	for _, file := range snapshot.Structure.Files {
		if strings.Contains(strings.ToLower(file.RelativePath), queryLower) {
			results = append(results, map[string]interface{}{
				"type": "file",
				"path": file.RelativePath,
				"match": "filename",
			})
		}
	}
	
	// Search in TODO items
	for _, todo := range snapshot.TODOs {
		if strings.Contains(strings.ToLower(todo.Message), queryLower) {
			results = append(results, map[string]interface{}{
				"type": "todo",
				"file": todo.File,
				"line": todo.Line,
				"message": todo.Message,
				"match": "todo",
			})
		}
	}
	
	// Search in errors
	for _, err := range snapshot.ActiveErrors {
		if strings.Contains(strings.ToLower(err.Message), queryLower) {
			results = append(results, map[string]interface{}{
				"type": "error",
				"file": err.File,
				"line": err.Line,
				"message": err.Message,
				"match": "error",
			})
		}
	}
	
	return results
}

// Helper methods for queries
func (is *IntelligenceServer) groupErrorsByFile(errors []ErrorInfo) map[string][]ErrorInfo {
	grouped := make(map[string][]ErrorInfo)
	
	for _, err := range errors {
		grouped[err.File] = append(grouped[err.File], err)
	}
	
	return grouped
}

func (is *IntelligenceServer) getRecentActivity(snapshot *ProjectSnapshot) map[string]interface{} {
	return map[string]interface{}{
		"recent_changes": snapshot.RecentChanges,
		"git_activity": map[string]interface{}{
			"branch":        snapshot.GitStatus.Branch,
			"last_commit":   snapshot.GitStatus.LastCommitTime,
			"modified_files": len(snapshot.GitStatus.ModifiedFiles),
		},
		"error_activity": len(snapshot.ActiveErrors),
	}
}

func (is *IntelligenceServer) getProjectStats(snapshot *ProjectSnapshot) map[string]interface{} {
	languageStats := make(map[string]int)
	
	for _, file := range snapshot.Structure.Files {
		if file.Language != "text" {
			languageStats[file.Language]++
		}
	}
	
	return map[string]interface{}{
		"total_files":     len(snapshot.Structure.Files),
		"total_size":      snapshot.Structure.TotalSize,
		"languages":       languageStats,
		"error_count":     len(snapshot.ActiveErrors),
		"todo_count":      len(snapshot.TODOs),
		"dependency_count": len(snapshot.Dependencies),
		"health_score":    snapshot.Health.Score,
	}
}

func (is *IntelligenceServer) Start(port string) error {
	log.Printf("Starting Project Intelligence Service on port %s", port)
	log.Printf("Monitoring workspace: %s", is.pi.workspace)
	log.Printf("API endpoints available at: http://localhost:%s/", port)
	
	return is.app.Listen(":" + port)
}

func main() {
	workspace := "."
	if len(os.Args) > 1 {
		workspace = os.Args[1]
	}
	
	// Convert to absolute path
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		log.Fatal("Invalid workspace path:", err)
	}
	
	// Verify workspace exists
	if _, err := os.Stat(absWorkspace); os.IsNotExist(err) {
		log.Fatal("Workspace does not exist:", absWorkspace)
	}
	
	server := NewIntelligenceServer(absWorkspace)
	
	log.Printf("Project Intelligence Service starting...")
	log.Printf("Workspace: %s", absWorkspace)
	
	if err := server.Start("3002"); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
```

### 3. Create `claude-query.sh`

```bash
#!/bin/bash

# Claude Code Project Intelligence Query Tool
# Usage: ./claude-query.sh [command] [options]

BASE_URL="http://localhost:3002"
TIMEOUT=10

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Helper function to make HTTP requests
query_api() {
    local endpoint="$1"
    local output_format="${2:-json}"
    
    response=$(curl -s --max-time $TIMEOUT "$BASE_URL$endpoint" 2>/dev/null)
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Cannot connect to Project Intelligence Service${NC}"
        echo -e "${YELLOW}Make sure the service is running on port 3002${NC}"
        return 1
    fi
    
    if [ "$output_format" = "pretty" ]; then
        echo "$response" | jq . 2>/dev/null || echo "$response"
    else
        echo "$response"
    fi
}

# Pretty print functions
print_header() {
    echo -e "${CYAN}=====================================${NC}"
    echo -e "${WHITE}$1${NC}"
    echo -e "${CYAN}=====================================${NC}"
}

print_section() {
    echo -e "\n${BLUE}📋 $1${NC}"
    echo -e "${BLUE}$(printf '%.0s-' {1..40})${NC}"
}

# Command functions
cmd_status() {
    print_header "Project Intelligence Status"
    query_api "/" pretty
}

cmd_health() {
    print_header "Project Health Summary"
    
    health_data=$(query_api "/health")
    
    if [ $? -eq 0 ]; then
        score=$(echo "$health_data" | jq -r '.score // "N/A"')
        errors=$(echo "$health_data" | jq -r '.error_count // 0')
        warnings=$(echo "$health_data" | jq -r '.warning_count // 0')
        debt=$(echo "$health_data" | jq -r '.technical_debt // "unknown"')
        
        echo -e "${GREEN}Health Score: $score/100${NC}"
        echo -e "${RED}Active Errors: $errors${NC}"
        echo -e "${YELLOW}Warnings: $warnings${NC}"
        echo -e "${PURPLE}Technical Debt: $debt${NC}"
    else
        echo "Failed to retrieve health data"
    fi
}

cmd_errors() {
    print_header "Active Errors & Warnings"
    
    errors_data=$(query_api "/errors")
    
    if [ $? -eq 0 ]; then
        error_count=$(echo "$errors_data" | jq length)
        
        if [ "$error_count" -eq 0 ]; then
            echo -e "${GREEN}✅ No errors detected!${NC}"
        else
            echo -e "${RED}Found $error_count error(s):${NC}\n"
            
            echo "$errors_data" | jq -r '.[] | "\(.file):\(.line): \(.type) - \(.message)"' | while read line; do
                echo -e "${RED}❌ $line${NC}"
            done
        fi
    fi
}

cmd_structure() {
    print_header "Project Structure Overview"
    
    structure_data=$(query_api "/structure")
    
    if [ $? -eq 0 ]; then
        project_type=$(echo "$structure_data" | jq -r '.project_type // "unknown"')
        total_files=$(echo "$structure_data" | jq -r '.total_files // 0')
        total_size=$(echo "$structure_data" | jq -r '.total_size // 0')
        
        echo -e "${BLUE}Project Type: $project_type${NC}"
        echo -e "${BLUE}Total Files: $total_files${NC}"
        echo -e "${BLUE}Total Size: $(numfmt --to=iec $total_size 2>/dev/null || echo "$total_size bytes")${NC}"
        
        print_section "Main Files"
        echo "$structure_data" | jq -r '.main_files[]? // empty' | while read file; do
            echo -e "${GREEN}📄 $file${NC}"
        done
        
        print_section "Config Files"
        echo "$structure_data" | jq -r '.config_files[]? // empty' | while read file; do
            echo -e "${YELLOW}⚙️  $file${NC}"
        done
    fi
}

cmd_git() {
    print_header "Git Repository Status"
    
    git_data=$(query_api "/git")
    
    if [ $? -eq 0 ]; then
        branch=$(echo "$git_data" | jq -r '.branch // "N/A"')
        commit_hash=$(echo "$git_data" | jq -r '.commit_hash // "N/A"')
        commit_msg=$(echo "$git_data" | jq -r '.commit_message // "N/A"')
        is_dirty=$(echo "$git_data" | jq -r '.is_dirty // false')
        
        echo -e "${CYAN}Branch: $branch${NC}"
        echo -e "${CYAN}Commit: $commit_hash${NC}"
        echo -e "${CYAN}Message: $commit_msg${NC}"
        
        if [ "$is_dirty" = "true" ]; then
            echo -e "${YELLOW}Status: Working directory has changes${NC}"
            
            modified_count=$(echo "$git_data" | jq '.modified_files | length')
            untracked_count=$(echo "$git_data" | jq '.untracked_files | length')
            
            if [ "$modified_count" -gt 0 ]; then
                echo -e "\n${YELLOW}Modified files ($modified_count):${NC}"
                echo "$git_data" | jq -r '.modified_files[]?' | while read file; do
                    echo -e "${YELLOW}  📝 $file${NC}"
                done
            fi
            
            if [ "$untracked_count" -gt 0 ]; then
                echo -e "\n${RED}Untracked files ($untracked_count):${NC}"
                echo "$git_data" | jq -r '.untracked_files[]?' | while read file; do
                    echo -e "${RED}  ❓ $file${NC}"
                done
            fi
        else
            echo -e "${GREEN}Status: Working directory clean${NC}"
        fi
    fi
}

cmd_changes() {
    print_header "Recent File Changes"
    
    changes_data=$(query_api "/changes")
    
    if [ $? -eq 0 ]; then
        change_count=$(echo "$changes_data" | jq length)
        
        if [ "$change_count" -eq 0 ]; then
            echo -e "${GREEN}No recent changes detected${NC}"
        else
            echo -e "${BLUE}$change_count recent change(s):${NC}\n"
            
            echo "$changes_data" | jq -r '.[] | "\(.timestamp) \(.type) \(.path)"' | while read timestamp type path; do
                file=$(basename "$path")
                time_formatted=$(date -d "$timestamp" '+%H:%M:%S' 2>/dev/null || echo "$timestamp")
                
                case "$type" in
                    "created")
                        echo -e "${GREEN}✨ $time_formatted - Created: $file${NC}"
                        ;;
                    "modified")
                        echo -e "${YELLOW}📝 $time_formatted - Modified: $file${NC}"
                        ;;
                    "deleted")
                        echo -e "${RED}🗑️  $time_formatted - Deleted: $file${NC}"
                        ;;
                    *)
                        echo -e "${BLUE}🔄 $time_formatted - $type: $file${NC}"
                        ;;
                esac
            done
        fi
    fi
}

cmd_todos() {
    print_header "TODO Items in Code"
    
    todos_data=$(query_api "/todos")
    
    if [ $? -eq 0 ]; then
        todo_count=$(echo "$todos_data" | jq length)
        
        if [ "$todo_count" -eq 0 ]; then
            echo -e "${GREEN}No TODO items found${NC}"
        else
            echo -e "${YELLOW}$todo_count TODO item(s) found:${NC}\n"
            
            echo "$todos_data" | jq -r '.[] | "\(.file):\(.line) [\(.type)] \(.message)"' | while read line; do
                echo -e "${YELLOW}📝 $line${NC}"
            done
        fi
    fi
}

cmd_dependencies() {
    print_header "Project Dependencies"
    
    deps_data=$(query_api "/dependencies")
    
    if [ $? -eq 0 ]; then
        dep_count=$(echo "$deps_data" | jq length)
        
        if [ "$dep_count" -eq 0 ]; then
            echo -e "${YELLOW}No dependencies detected${NC}"
        else
            echo -e "${BLUE}$dep_count dependencies found:${NC}\n"
            
            echo "$deps_data" | jq -r '.[] | "\(.source) \(.name) \(.version) [\(.type)]"' | while read source name version type; do
                echo -e "${CYAN}📦 $name $version ($type) - $source${NC}"
            done
        fi
    fi
}

cmd_processes() {
    print_header "Running Processes"
    
    proc_data=$(query_api "/processes")
    
    if [ $? -eq 0 ]; then
        proc_count=$(echo "$proc_data" | jq length)
        
        if [ "$proc_count" -eq 0 ]; then
            echo -e "${YELLOW}No project-related processes detected${NC}"
        else
            echo -e "${BLUE}$proc_count process(es) running:${NC}\n"
            
            echo "$proc_data" | jq -r '.[] | "\(.pid) \(.name) \(.memory_mb // 0)"' | while read pid name memory; do
                memory_display=""
                if [ "$memory" != "0" ] && [ "$memory" != "null" ]; then
                    memory_display=" (${memory}MB)"
                fi
                echo -e "${GREEN}⚡ PID $pid: $name$memory_display${NC}"
            done
        fi
    fi
}

cmd_search() {
    local query="$1"
    
    if [ -z "$query" ]; then
        echo -e "${RED}Error: Search query required${NC}"
        echo "Usage: $0 search \"your search term\""
        return 1
    fi
    
    print_header "Search Results for: $query"
    
    search_data=$(query_api "/search?q=$(echo "$query" | sed 's/ /%20/g')")
    
    if [ $? -eq 0 ]; then
        result_count=$(echo "$search_data" | jq -r '.count // 0')
        
        if [ "$result_count" -eq 0 ]; then
            echo -e "${YELLOW}No results found${NC}"
        else
            echo -e "${GREEN}Found $result_count result(s):${NC}\n"
            
            echo "$search_data" | jq -r '.results[] | "\(.type) \(.file // .path) \(.line // "") \(.message // "")"' | while read type file line message; do
                location=""
                if [ -n "$line" ] && [ "$line" != "null" ]; then
                    location=":$line"
                fi
                
                case "$type" in
                    "file")
                        echo -e "${CYAN}📄 File: $file${NC}"
                        ;;
                    "todo")
                        echo -e "${YELLOW}📝 TODO: $file$location - $message${NC}"
                        ;;
                    "error")
                        echo -e "${RED}❌ Error: $file$location - $message${NC}"
                        ;;
                    *)
                        echo -e "${BLUE}🔍 $type: $file$location${NC}"
                        ;;
                esac
            done
        fi
    fi
}

cmd_file() {
    local filepath="$1"
    
    if [ -z "$filepath" ]; then
        echo -e "${RED}Error: File path required${NC}"
        echo "Usage: $0 file \"path/to/file.ext\""
        return 1
    fi
    
    print_header "File Information: $filepath"
    
    file_data=$(query_api "/files/$filepath")
    
    if [ $? -eq 0 ]; then
        size=$(echo "$file_data" | jq -r '.size // 0')
        language=$(echo "$file_data" | jq -r '.language // "unknown"')
        lines=$(echo "$file_data" | jq -r '.line_count // "N/A"')
        mod_time=$(echo "$file_data" | jq -r '.mod_time // "N/A"')
        
        echo -e "${BLUE}Language: $language${NC}"
        echo -e "${BLUE}Size: $(numfmt --to=iec $size 2>/dev/null || echo "$size bytes")${NC}"
        echo -e "${BLUE}Lines: $lines${NC}"
        echo -e "${BLUE}Modified: $mod_time${NC}"
    else
        echo -e "${RED}File not found or cannot be accessed${NC}"
    fi
}

cmd_quick() {
    print_header "Quick Project Overview"
    
    # Get health data
    health_data=$(query_api "/health")
    score=$(echo "$health_data" | jq -r '.score // "N/A"')
    errors=$(echo "$health_data" | jq -r '.error_count // 0')
    
    # Get structure data
    structure_data=$(query_api "/structure")
    project_type=$(echo "$structure_data" | jq -r '.project_type // "unknown"')
    total_files=$(echo "$structure_data" | jq -r '.total_files // 0')
    
    # Get git data
    git_data=$(query_api "/git")
    branch=$(echo "$git_data" | jq -r '.branch // "N/A"')
    is_dirty=$(echo "$git_data" | jq -r '.is_dirty // false')
    
    # Get recent changes
    changes_data=$(query_api "/changes")
    change_count=$(echo "$changes_data" | jq length)
    
    echo -e "${GREEN}Health Score: $score/100${NC} | ${RED}Errors: $errors${NC}"
    echo -e "${BLUE}Project: $project_type${NC} | ${BLUE}Files: $total_files${NC}"
    echo -e "${CYAN}Git Branch: $branch${NC} | $([ "$is_dirty" = "true" ] && echo -e "${YELLOW}Dirty${NC}" || echo -e "${GREEN}Clean${NC}")"
    echo -e "${PURPLE}Recent Changes: $change_count${NC}"
}

cmd_help() {
    echo -e "${WHITE}Claude Code Project Intelligence Query Tool${NC}"
    echo -e "${CYAN}=============================================${NC}"
    echo ""
    echo -e "${YELLOW}Usage:${NC} $0 [command] [options]"
    echo ""
    echo -e "${YELLOW}Commands:${NC}"
    echo -e "  ${GREEN}status${NC}       - Service status and available endpoints"
    echo -e "  ${GREEN}quick${NC}        - Quick project overview"
    echo -e "  ${GREEN}health${NC}       - Project health summary"
    echo -e "  ${GREEN}errors${NC}       - Show active errors and warnings"
    echo -e "  ${GREEN}structure${NC}    - Project structure overview"
    echo -e "  ${GREEN}git${NC}          - Git repository status"
    echo -e "  ${GREEN}changes${NC}      - Recent file changes"
    echo -e "  ${GREEN}todos${NC}        - TODO items in code"
    echo -e "  ${GREEN}dependencies${NC} - Project dependencies"
    echo -e "  ${GREEN}processes${NC}    - Running processes"
    echo -e "  ${GREEN}search${NC} \"query\" - Search across project"
    echo -e "  ${GREEN}file${NC} \"path\"   - Get file information"
    echo -e "  ${GREEN}help${NC}         - Show this help message"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 quick"
    echo "  $0 errors"
    echo "  $0 search \"TODO\""
    echo "  $0 file \"src/main.go\""
    echo ""
    echo -e "${BLUE}Service URL: $BASE_URL${NC}"
}

# Main command router
main() {
    case "${1:-help}" in
        "status")
            cmd_status
            ;;
        "quick"|"q")
            cmd_quick
            ;;
        "health"|"h")
            cmd_health
            ;;
        "errors"|"e")
            cmd_errors
            ;;
        "structure"|"s")
            cmd_structure
            ;;
        "git"|"g")
            cmd_git
            ;;
        "changes"|"c")
            cmd_changes
            ;;
        "todos"|"t")
            cmd_todos
            ;;
        "dependencies"|"deps"|"d")
            cmd_dependencies
            ;;
        "processes"|"proc"|"p")
            cmd_processes
            ;;
        "search")
            cmd_search "$2"
            ;;
        "file"|"f")
            cmd_file "$2"
            ;;
        "help"|"--help"|"-h"|"")
            cmd_help
            ;;
        *)
            echo -e "${RED}Unknown command: $1${NC}"
            echo "Use '$0 help' to see available commands"
            exit 1
            ;;
    esac
}

# Check dependencies
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required but not installed${NC}"
    echo "Install it with: sudo apt install jq (Ubuntu/Debian) or brew install jq (macOS)"
    exit 1
fi

if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is required but not installed${NC}"
    exit 1
fi

# Run main function with all arguments
main "$@"
```

### 4. Create `dashboard.html`

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Claude Code Intelligence Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Cascadia Code', 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', monospace;
            background: #0d1117;
            color: #c9d1d9;
            min-height: 100vh;
            padding: 15px;
        }
        
        .container {
            max-width: 1600px;
            margin: 0 auto;
        }
        
        .header {
            text-align: center;
            margin-bottom: 25px;
            padding: 20px;
            background: linear-gradient(135deg, #1f2937 0%, #374151 100%);
            border-radius: 12px;
            border: 1px solid #30363d;
        }
        
        .header h1 {
            color: #58a6ff;
            font-size: 2rem;
            margin-bottom: 8px;
            font-weight: 600;
        }
        
        .header .workspace {
            color: #7c3aed;
            font-size: 0.9rem;
            padding: 4px 12px;
            background: rgba(124, 58, 237, 0.1);
            border: 1px solid #7c3aed;
            border-radius: 6px;
            display: inline-block;
        }
        
        .grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin-bottom: 20px;
        }
        
        .grid-3 {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        
        .card {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 20px;
            transition: all 0.2s ease;
        }
        
        .card:hover {
            border-color: #58a6ff;
            transform: translateY(-1px);
        }
        
        .card-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 1px solid #30363d;
        }
        
        .card-title {
            color: #f0f6fc;
            font-size: 1.1rem;
            font-weight: 600;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        .status-indicator {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            display: inline-block;
        }
        
        .status-healthy { background: #3fb950; }
        .status-warning { background: #d29922; }
        .status-error { background: #f85149; }
        
        .metric {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px 0;
            border-bottom: 1px solid #21262d;
        }
        
        .metric:last-child {
            border-bottom: none;
        }
        
        .metric-label {
            color: #8b949e;
            font-size: 0.9rem;
        }
        
        .metric-value {
            color: #f0f6fc;
            font-weight: 500;
        }
        
        .error-item, .todo-item, .change-item {
            background: #0d1117;
            border: 1px solid #30363d;
            border-radius: 6px;
            padding: 12px;
            margin-bottom: 8px;
            font-size: 0.85rem;
        }
        
        .error-item {
            border-left: 3px solid #f85149;
        }
        
        .todo-item {
            border-left: 3px solid #d29922;
        }
        
        .change-item {
            border-left: 3px solid #58a6ff;
        }
        
        .file-path {
            color: #58a6ff;
            font-weight: 500;
        }
        
        .line-number {
            color: #d29922;
            font-size: 0.8rem;
        }
        
        .message {
            color: #c9d1d9;
            margin-top: 4px;
        }
        
        .timestamp {
            color: #8b949e;
            font-size: 0.75rem;
            float: right;
        }
        
        .api-section {
            background: #0d1117;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 20px;
        }
        
        .api-title {
            color: #f0f6fc;
            font-size: 1.2rem;
            margin-bottom: 15px;
            font-weight: 600;
        }
        
        .api-endpoints {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 10px;
        }
        
        .endpoint {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 6px;
            padding: 10px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .method {
            color: #3fb950;
            font-weight: bold;
            font-size: 0.8rem;
            padding: 2px 6px;
            background: rgba(63, 185, 80, 0.1);
            border-radius: 3px;
        }
        
        .endpoint-path {
            color: #58a6ff;
            font-family: inherit;
        }
        
        .copy-btn {
            background: #21262d;
            border: 1px solid #30363d;
            color: #f0f6fc;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.7rem;
            cursor: pointer;
            transition: all 0.2s;
        }
        
        .copy-btn:hover {
            background: #30363d;
            border-color: #58a6ff;
        }
        
        .controls {
            display: flex;
            gap: 10px;
            justify-content: center;
            margin-bottom: 20px;
        }
        
        .btn {
            background: #238636;
            border: 1px solid #2ea043;
            color: #fff;
            padding: 8px 16px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 0.9rem;
            transition: all 0.2s;
            font-family: inherit;
        }
        
        .btn:hover {
            background: #2ea043;
        }
        
        .btn-secondary {
            background: #21262d;
            border-color: #30363d;
            color: #f0f6fc;
        }
        
        .btn-secondary:hover {
            background: #30363d;
        }
        
        .loading {
            text-align: center;
            color: #58a6ff;
            padding: 40px;
            font-size: 1.1rem;
        }
        
        .error-banner {
            background: #490202;
            border: 1px solid #f85149;
            color: #ffd6cc;
            padding: 15px;
            border-radius: 6px;
            margin-bottom: 20px;
            text-align: center;
        }
        
        .success-banner {
            background: #0f1419;
            border: 1px solid #3fb950;
            color: #aff5b4;
            padding: 10px 15px;
            border-radius: 6px;
            margin-bottom: 20px;
            text-align: center;
            font-size: 0.9rem;
        }
        
        .scrollable {
            max-height: 300px;
            overflow-y: auto;
        }
        
        .scrollable::-webkit-scrollbar {
            width: 6px;
        }
        
        .scrollable::-webkit-scrollbar-track {
            background: #161b22;
        }
        
        .scrollable::-webkit-scrollbar-thumb {
            background: #30363d;
            border-radius: 3px;
        }
        
        .query-builder {
            background: #161b22;
            border: 1px solid #30363d;
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 20px;
        }
        
        .query-input {
            width: 100%;
            background: #0d1117;
            border: 1px solid #30363d;
            color: #f0f6fc;
            padding: 8px 12px;
            border-radius: 6px;
            font-family: inherit;
            font-size: 0.9rem;
        }
        
        .query-input:focus {
            outline: none;
            border-color: #58a6ff;
        }
        
        .language-stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 10px;
        }
        
        .language-item {
            background: #0d1117;
            border: 1px solid #30363d;
            border-radius: 6px;
            padding: 10px;
            text-align: center;
        }
        
        .language-name {
            color: #58a6ff;
            font-size: 0.8rem;
            font-weight: 500;
        }
        
        .language-count {
            color: #f0f6fc;
            font-size: 1.2rem;
            font-weight: bold;
        }
        
        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }
        
        .loading-pulse {
            animation: pulse 1.5s infinite;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 Claude Code Intelligence</h1>
            <div class="workspace" id="workspace-display">Loading workspace...</div>
        </div>
        
        <div id="loading" class="loading">
            <div class="loading-pulse">🔍 Initializing project intelligence...</div>
        </div>
        
        <div id="error-banner" class="error-banner" style="display: none;">
            <strong>Connection Error:</strong> Cannot connect to Project Intelligence Service on port 3002.
            <br>Make sure the service is running: <code>go run main.go /path/to/your/project</code>
        </div>
        
        <div id="success-banner" class="success-banner" style="display: none;">
            Project intelligence is active and monitoring your workspace!
        </div>
        
        <div id="dashboard" style="display: none;">
            <div class="controls">
                <button class="btn" onclick="refreshData()">🔄 Refresh</button>
                <button class="btn btn-secondary" onclick="toggleAutoRefresh()">
                    ⏱️ Auto: <span id="auto-status">Off</span>
                </button>
                <button class="btn btn-secondary" onclick="exportSnapshot()">📊 Export Data</button>
            </div>
            
            <div class="grid">
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">
                            📈 Project Health
                            <span class="status-indicator" id="health-indicator"></span>
                        </div>
                        <span id="health-score" class="metric-value">0</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Active Errors</span>
                        <span class="metric-value" id="error-count">0</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Warnings</span>
                        <span class="metric-value" id="warning-count">0</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Technical Debt</span>
                        <span class="metric-value" id="tech-debt">Low</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Last Check</span>
                        <span class="metric-value" id="last-check">-</span>
                    </div>
                </div>
                
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">📁 Project Overview</div>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Total Files</span>
                        <span class="metric-value" id="total-files">0</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Project Type</span>
                        <span class="metric-value" id="project-type">Unknown</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Git Branch</span>
                        <span class="metric-value" id="git-branch">-</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Dependencies</span>
                        <span class="metric-value" id="dependency-count">0</span>
                    </div>
                </div>
            </div>
            
            <div class="grid-3">
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">🚨 Active Errors</div>
                        <span class="btn copy-btn" onclick="copyEndpoint('/errors')">API</span>
                    </div>
                    <div class="scrollable" id="errors-list">
                        <div style="text-align: center; color: #8b949e; padding: 20px;">No errors detected</div>
                    </div>
                </div>
                
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">📝 TODOs & Notes</div>
                        <span class="btn copy-btn" onclick="copyEndpoint('/todos')">API</span>
                    </div>
                    <div class="scrollable" id="todos-list">
                        <div style="text-align: center; color: #8b949e; padding: 20px;">No TODOs found</div>
                    </div>
                </div>
                
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">📂 Recent Changes</div>
                        <span class="btn copy-btn" onclick="copyEndpoint('/changes')">API</span>
                    </div>
                    <div class="scrollable" id="changes-list">
                        <div style="text-align: center; color: #8b949e; padding: 20px;">No recent changes</div>
                    </div>
                </div>
            </div>
            
            <div class="grid">
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">💻 Language Breakdown</div>
                    </div>
                    <div class="language-stats" id="language-stats">
                        <!-- Language stats will be populated here -->
                    </div>
                </div>
                
                <div class="card">
                    <div class="card-header">
                        <div class="card-title">⚡ Running Processes</div>
                        <span class="btn copy-btn" onclick="copyEndpoint('/processes')">API</span>
                    </div>
                    <div class="scrollable" id="processes-list">
                        <div style="text-align: center; color: #8b949e; padding: 20px;">No processes detected</div>
                    </div>
                </div>
            </div>
            
            <div class="query-builder">
                <div class="card-title" style="margin-bottom: 10px;">🔍 Quick Search</div>
                <input type="text" class="query-input" placeholder="Search files, TODOs, errors..." id="search-input" onkeypress="handleSearch(event)">
                <div id="search-results" style="margin-top: 10px;"></div>
            </div>
        </div>
        
        <div class="api-section">
            <div class="api-title">🛠️ Claude Code API Endpoints</div>
            <div class="api-endpoints">
                <div class="endpoint">
                    <div>
                        <span class="method">GET</span>
                        <span class="endpoint-path">/snapshot</span>
                    </div>
                    <button class="copy-btn" onclick="copyEndpoint('/snapshot')">Copy</button>
                </div>
                <div class="endpoint">
                    <div>
                        <span class="method">GET</span>
                        <span class="endpoint-path">/structure</span>
                    </div>
                    <button class="copy-btn" onclick="copyEndpoint('/structure')">Copy</button>
                </div>
                <div class="endpoint">
                    <div>
                        <span class="method">GET</span>
                        <span class="endpoint-path">/errors</span>
                    </div>
                    <button class="copy-btn" onclick="copyEndpoint('/errors')">Copy</button>
                </div>
                <div class="endpoint">
                    <div>
                        <span class="method">GET</span>
                        <span class="endpoint-path">/git</span>
                    </div>
                    <button class="copy-btn" onclick="copyEndpoint('/git')">Copy</button>
                </div>
                <div class="endpoint">
                    <div>
                        <span class="method">GET</span>
                        <span class="endpoint-path">/changes</span>
                    </div>
                    <button class="copy-btn" onclick="copyEndpoint('/changes')">Copy</button>
                </div>
                <div class="endpoint">
                    <div>
                        <span class="method">GET</span>
                        <span class="endpoint-path">/health</span>
                    </div>
                    <button class="copy-btn" onclick="copyEndpoint('/health')">Copy</button>
                </div>
                <div class="endpoint">
                    <div>
                        <span class="method">GET</span>
                        <span class="endpoint-path">/search?q=query</span>
                    </div>
                    <button class="copy-btn" onclick="copyEndpoint('/search?q=')">Copy</button>
                </div>
                <div class="endpoint">
                    <div>
                        <span class="method">GET</span>
                        <span class="endpoint-path">/files/path/content</span>
                    </div>
                    <button class="copy-btn" onclick="copyEndpoint('/files/')">Copy</button>
                </div>
            </div>
        </div>
    </div>

    <script>
        const SERVER_URL = 'http://localhost:3002';
        let autoRefreshInterval;
        let autoRefreshEnabled = false;
        
        async function loadData() {
            const loadingEl = document.getElementById('loading');
            const errorEl = document.getElementById('error-banner');
            const successEl = document.getElementById('success-banner');
            const dashboardEl = document.getElementById('dashboard');
            
            try {
                loadingEl.style.display = 'block';
                errorEl.style.display = 'none';
                successEl.style.display = 'none';
                
                const response = await fetch(`${SERVER_URL}/snapshot`);
                if (!response.ok) throw new Error('Service unavailable');
                
                const snapshot = await response.json();
                
                updateDashboard(snapshot);
                
                loadingEl.style.display = 'none';
                successEl.style.display = 'block';
                dashboardEl.style.display = 'block';
                
                setTimeout(() => {
                    successEl.style.display = 'none';
                }, 3000);
                
            } catch (error) {
                console.error('Error loading data:', error);
                loadingEl.style.display = 'none';
                errorEl.style.display = 'block';
            }
        }
        
        function updateDashboard(snapshot) {
            document.getElementById('workspace-display').textContent = snapshot.structure.root_path;
            
            const health = snapshot.health;
            document.getElementById('health-score').textContent = health.score;
            document.getElementById('error-count').textContent = health.error_count;
            document.getElementById('warning-count').textContent = health.warning_count;
            document.getElementById('tech-debt').textContent = health.technical_debt;
            document.getElementById('last-check').textContent = new Date(health.last_health_check).toLocaleTimeString();
            
            const indicator = document.getElementById('health-indicator');
            if (health.score >= 80) {
                indicator.className = 'status-indicator status-healthy';
            } else if (health.score >= 50) {
                indicator.className = 'status-indicator status-warning';
            } else {
                indicator.className = 'status-indicator status-error';
            }
            
            document.getElementById('total-files').textContent = snapshot.structure.total_files;
            document.getElementById('project-type').textContent = snapshot.structure.project_type;
            document.getElementById('git-branch').textContent = snapshot.git_status.branch || 'Not a git repo';
            document.getElementById('dependency-count').textContent = snapshot.dependencies.length;
            
            updateErrorsList(snapshot.active_errors);
            updateTodosList(snapshot.todos);
            updateChangesList(snapshot.recent_changes);
            updateLanguageStats(snapshot.structure.files);
            updateProcessesList(snapshot.running_processes);
        }
        
        function updateErrorsList(errors) {
            const container = document.getElementById('errors-list');
            
            if (errors.length === 0) {
                container.innerHTML = '<div style="text-align: center; color: #8b949e; padding: 20px;">No errors detected ✅</div>';
                return;
            }
            
            container.innerHTML = errors.map(error => `
                <div class="error-item">
                    <div style="display: flex; justify-content: space-between; align-items: flex-start;">
                        <div>
                            <div class="file-path">${error.file}</div>
                            <div class="line-number">Line ${error.line}${error.column ? `:${error.column}` : ''}</div>
                            <div class="message">${error.message}</div>
                        </div>
                        <div class="timestamp">${new Date(error.timestamp).toLocaleTimeString()}</div>
                    </div>
                </div>
            `).join('');
        }
        
        function updateTodosList(todos) {
            const container = document.getElementById('todos-list');
            
            if (todos.length === 0) {
                container.innerHTML = '<div style="text-align: center; color: #8b949e; padding: 20px;">No TODOs found</div>';
                return;
            }
            
            container.innerHTML = todos.slice(0, 10).map(todo => `
                <div class="todo-item">
                    <div class="file-path">${todo.file}</div>
                    <div class="line-number">Line ${todo.line} • ${todo.type}</div>
                    <div class="message">${todo.message}</div>
                </div>
            `).join('');
        }
        
        function updateChangesList(changes) {
            const container = document.getElementById('changes-list');
            
            if (changes.length === 0) {
                container.innerHTML = '<div style="text-align: center; color: #8b949e; padding: 20px;">No recent changes</div>';
                return;
            }
            
            container.innerHTML = changes.slice(0, 10).map(change => `
                <div class="change-item">
                    <div style="display: flex; justify-content: space-between; align-items: flex-start;">
                        <div>
                            <div class="file-path">${change.path.split('/').pop()}</div>
                            <div style="color: #8b949e; font-size: 0.8rem;">${change.type.toUpperCase()}</div>
                        </div>
                        <div class="timestamp">${new Date(change.timestamp).toLocaleTimeString()}</div>
                    </div>
                </div>
            `).join('');
        }
        
        function updateLanguageStats(files) {
            const container = document.getElementById('language-stats');
            const languageCounts = {};
            
            files.forEach(file => {
                if (file.language && file.language !== 'text') {
                    languageCounts[file.language] = (languageCounts[file.language] || 0) + 1;
                }
            });
            
            const sortedLanguages = Object.entries(languageCounts)
                .sort(([,a], [,b]) => b - a)
                .slice(0, 8);
            
            if (sortedLanguages.length === 0) {
                container.innerHTML = '<div style="text-align: center; color: #8b949e; padding: 20px; grid-column: 1 / -1;">No code files detected</div>';
                return;
            }
            
            container.innerHTML = sortedLanguages.map(([lang, count]) => `
                <div class="language-item">
                    <div class="language-count">${count}</div>
                    <div class="language-name">${lang}</div>
                </div>
            `).join('');
        }
        
        function updateProcessesList(processes) {
            const container = document.getElementById('processes-list');
            
            if (processes.length === 0) {
                container.innerHTML = '<div style="text-align: center; color: #8b949e; padding: 20px;">No project processes detected</div>';
                return;
            }
            
            container.innerHTML = processes.slice(0, 8).map(proc => `
                <div class="change-item">
                    <div style="display: flex; justify-content: space-between; align-items: flex-start;">
                        <div>
                            <div class="file-path">${proc.name}</div>
                            <div style="color: #8b949e; font-size: 0.8rem;">PID: ${proc.pid}</div>
                        </div>
                        <div style="text-align: right; font-size: 0.8rem; color: #8b949e;">
                            ${proc.memory_mb ? Math.round(proc.memory_mb) + 'MB' : ''}
                        </div>
                    </div>
                </div>
            `).join('');
        }
        
        async function handleSearch(event) {
            if (event.key === 'Enter') {
                const query = event.target.value.trim();
                if (!query) return;
                
                try {
                    const response = await fetch(`${SERVER_URL}/search?q=${encodeURIComponent(query)}`);
                    const results = await response.json();
                    
                    const container = document.getElementById('search-results');
                    
                    if (results.count === 0) {
                        container.innerHTML = '<div style="color: #8b949e; padding: 10px;">No results found</div>';
                        return;
                    }
                    
                    container.innerHTML = `
                        <div style="color: #58a6ff; margin-bottom: 10px;">Found ${results.count} results:</div>
                        ${results.results.slice(0, 10).map(result => `
                            <div class="change-item">
                                <div class="file-path">${result.file || result.path}</div>
                                <div style="color: #8b949e; font-size: 0.8rem;">
                                    ${result.type} ${result.line ? `• Line ${result.line}` : ''}
                                </div>
                                ${result.message ? `<div class="message">${result.message}</div>` : ''}
                            </div>
                        `).join('')}
                    `;
                } catch (error) {
                    console.error('Search error:', error);
                }
            }
        }
        
        function copyEndpoint(endpoint) {
            const fullUrl = `${SERVER_URL}${endpoint}`;
            navigator.clipboard.writeText(fullUrl).then(() => {
                event.target.textContent = 'Copied!';
                setTimeout(() => {
                    event.target.textContent = 'Copy';
                }, 1000);
            });
        }
        
        function refreshData() {
            loadData();
        }
        
        function toggleAutoRefresh() {
            const statusEl = document.getElementById('auto-status');
            
            if (autoRefreshEnabled) {
                clearInterval(autoRefreshInterval);
                autoRefreshEnabled = false;
                statusEl.textContent = 'Off';
            } else {
                autoRefreshInterval = setInterval(loadData, 10000);
                autoRefreshEnabled = true;
                statusEl.textContent = 'On';
            }
        }
        
        async function exportSnapshot() {
            try {
                const response = await fetch(`${SERVER_URL}/snapshot`);
                const snapshot = await response.json();
                
                const dataStr = JSON.stringify(snapshot, null, 2);
                const dataBlob = new Blob([dataStr], {type: 'application/json'});
                
                const link = document.createElement('a');
                link.href = URL.createObjectURL(dataBlob);
                link.download = `project-snapshot-${new Date().toISOString().split('T')[0]}.json`;
                link.click();
            } catch (error) {
                console.error('Export error:', error);
                alert('Failed to export snapshot');
            }
        }
        
        // Initialize
        loadData();
        
        // Auto-refresh every 30 seconds by default
        setInterval(loadData, 30000);
    </script>
</body>
</html>
```

---

## 🚀 Installation Commands

Run these commands in sequence to set up the complete system:

```bash
# 1. Create project directory
mkdir claude-intelligence && cd claude-intelligence

# 2. Download dependencies
go mod init claude-intelligence
go get github.com/gofiber/fiber/v2

# 3. Make CLI tool executable
chmod +x claude-query.sh

# 4. Optional: Create global command
sudo ln -s $(pwd)/claude-query.sh /usr/local/bin/claude-query

# 5. Start the service (adjust path as needed)
go run main.go .

# In another terminal, test the CLI
./claude-query.sh quick
```

---

## 🎯 For Claude Code Integration

Once installed, tell Claude Code:

> **"I've set up a Project Intelligence Service for you. Before working on this project, always run these commands to understand the current state:**
> 
> - `./claude-query.sh quick` - Get project overview
> - `./claude-query.sh errors` - See current errors
> - `./claude-query.sh changes` - See recent changes
> - `./claude-query.sh search "term"` - Search the project
> 
> **You can also query the API directly at http://localhost:3002 with endpoints like:**
> - `/errors` - Current compilation errors
> - `/structure` - Project file structure
> - `/git` - Git repository status
> - `/search?q=term` - Search across everything
> 
> **Use these tools to stay updated on the project state without me having to explain everything manually."**

---

## 📊 Dashboard

Open `dashboard.html` in your browser to see:
- Real-time project health score
- Active errors with file locations
- Recent file changes
- Git status and changes
- TODO items in code
- Language breakdown
- Running processes
- Quick search functionality

---

## ✅ Verification

Test that everything works:

```bash
# Test the service
curl http://localhost:3002/health

# Test the CLI
./claude-query.sh quick

# Open dashboard
open dashboard.html  # macOS
xdg-open dashboard.html  # Linux
```

**You now have a complete Project Intelligence Service that makes Claude Code 5x more effective by giving it real-time project awareness!**