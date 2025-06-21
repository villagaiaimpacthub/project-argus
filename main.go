package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
)

// ProjectIntelligence represents the main intelligence service
type ProjectIntelligence struct {
	workspace      string
	fileWatcher    *FileWatcher
	gitWatcher     *GitWatcher
	errorWatcher   *ErrorWatcher
	buildWatcher   *BuildWatcher
	processWatcher *ProcessWatcher
	processMonitor *ProcessMonitor
	lastSnapshot   *ProjectSnapshot
	config         *ProcessMonitorConfig
	mutex          sync.RWMutex
}

// ProcessMonitor manages real-time process monitoring
type ProcessMonitor struct {
	activeProcesses map[int]*MonitoredProcess
	errorStream     chan StreamError
	config          *ProcessMonitorConfig
	metrics         *ProcessMetrics
	wsConnections   map[*websocket.Conn]bool
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// MonitoredProcess represents a monitored process
type MonitoredProcess struct {
	PID         int       `json:"pid"`
	Command     string    `json:"command"`
	Args        []string  `json:"args"`
	StartTime   time.Time `json:"start_time"`
	Status      string    `json:"status"` // running, stopped, error
	OutputLines []string  `json:"output_lines"`
	ErrorLines  []string  `json:"error_lines"`
	LastError   *StreamError `json:"last_error,omitempty"`
	WorkingDir  string    `json:"working_dir"`
	cmd         *exec.Cmd
	stdoutPipe  io.ReadCloser
	stderrPipe  io.ReadCloser
	mutex       sync.RWMutex
}

// StreamError represents a real-time error from a monitored process
type StreamError struct {
	ProcessPID  int       `json:"process_pid"`
	Command     string    `json:"command"`
	ErrorType   string    `json:"error_type"` // runtime, compilation, test, server
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"` // error, warning, info
	Context     []string  `json:"context"` // surrounding lines
	Source      string    `json:"source"` // stdout, stderr
	Line        int       `json:"line,omitempty"`
	Column      int       `json:"column,omitempty"`
}

// ProcessCommand represents a command to monitor
type ProcessCommand struct {
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	WorkingDir    string            `json:"working_dir"`
	Environment   map[string]string `json:"environment"`
	AutoRestart   bool              `json:"auto_restart"`
	ErrorPatterns []string          `json:"error_patterns"`
}

// ProcessMonitorConfig contains configuration for process monitoring
type ProcessMonitorConfig struct {
	MaxProcesses       int           `json:"max_processes"`
	OutputBufferSize   int           `json:"output_buffer_size"`
	ErrorStreamBuffer  int           `json:"error_stream_buffer"`
	ProcessTimeout     time.Duration `json:"process_timeout"`
	CleanupInterval    time.Duration `json:"cleanup_interval"`
	MaxOutputLines     int           `json:"max_output_lines"`
	AllowedCommands    []string      `json:"allowed_commands"`
	RateLimitPerMinute int           `json:"rate_limit_per_minute"`
}

// ProcessMetrics tracks monitoring metrics
type ProcessMetrics struct {
	ProcessStartAttempts int64 `json:"process_start_attempts"`
	ProcessStartFailures int64 `json:"process_start_failures"`
	ActiveProcesses      int64 `json:"active_processes"`
	TotalErrors          int64 `json:"total_errors"`
	mutex                sync.RWMutex
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

// loadConfig loads configuration from file and environment
func loadConfig() *ProcessMonitorConfig {
	config := &ProcessMonitorConfig{
		MaxProcesses:       10,
		OutputBufferSize:   1000,
		ErrorStreamBuffer:  1000,
		ProcessTimeout:     time.Hour,
		CleanupInterval:    30 * time.Second,
		MaxOutputLines:     10000,
		RateLimitPerMinute: 10,
		AllowedCommands:    []string{"npm", "node", "go", "python", "yarn", "cargo", "next", "vite", "jest", "make", "mvn", "gradle"},
	}
	
	// Load from config file if exists
	if data, err := os.ReadFile("argus-config.json"); err == nil {
		json.Unmarshal(data, config)
	}
	
	// Override with environment variables
	if maxProc := os.Getenv("ARGUS_MAX_PROCESSES"); maxProc != "" {
		if val, err := strconv.Atoi(maxProc); err == nil {
			config.MaxProcesses = val
		}
	}
	
	if timeout := os.Getenv("ARGUS_PROCESS_TIMEOUT"); timeout != "" {
		if val, err := time.ParseDuration(timeout); err == nil {
			config.ProcessTimeout = val
		}
	}
	
	return config
}

// NewProcessMonitor creates a new process monitor
func NewProcessMonitor(config *ProcessMonitorConfig) *ProcessMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ProcessMonitor{
		activeProcesses: make(map[int]*MonitoredProcess),
		errorStream:     make(chan StreamError, config.ErrorStreamBuffer),
		config:          config,
		metrics:         &ProcessMetrics{},
		wsConnections:   make(map[*websocket.Conn]bool),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// NewProjectIntelligence creates a new Project Argus instance
func NewProjectIntelligence(workspace string) *ProjectIntelligence {
	config := loadConfig()
	
	pi := &ProjectIntelligence{
		workspace:      workspace,
		fileWatcher:    &FileWatcher{workspace: workspace, changes: []FileChange{}},
		gitWatcher:     &GitWatcher{workspace: workspace},
		errorWatcher:   &ErrorWatcher{errors: []ErrorInfo{}},
		buildWatcher:   &BuildWatcher{status: &BuildStatus{}},
		processWatcher: &ProcessWatcher{processes: []ProcessInfo{}},
		processMonitor: NewProcessMonitor(config),
		config:         config,
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
	
	// Start process monitor
	go pi.processMonitor.startMonitoring()
	
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

// ProcessMonitor implementation
func (pm *ProcessMonitor) startMonitoring() {
	log.Println("Process monitor started")
	
	// Start error stream processor
	go pm.processErrorStream()
	
	// Monitor for common development processes
	go pm.autoDetectDevProcesses()
	
	// Cleanup stopped processes periodically
	ticker := time.NewTicker(pm.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-pm.ctx.Done():
			log.Println("Process monitor stopping...")
			return
		case <-ticker.C:
			pm.cleanupStoppedProcesses()
		}
	}
}

func (pm *ProcessMonitor) processErrorStream() {
	// Use worker pool pattern for processing errors
	errorBuffer := make(chan StreamError, pm.config.ErrorStreamBuffer)
	
	// Start worker goroutines
	for i := 0; i < runtime.NumCPU(); i++ {
		go pm.errorWorker(errorBuffer)
	}
	
	// Process incoming errors with backpressure handling
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-pm.ctx.Done():
			return
		case streamError := <-pm.errorStream:
			select {
			case errorBuffer <- streamError:
				// Successfully queued
			default:
				// Buffer full - implement overflow strategy
				pm.handleBufferOverflow()
			}
		case <-ticker.C:
			pm.flushPendingErrors()
		}
	}
}

func (pm *ProcessMonitor) errorWorker(errorBuffer <-chan StreamError) {
	for {
		select {
		case <-pm.ctx.Done():
			return
		case streamError := <-errorBuffer:
			pm.processStreamError(streamError)
		}
	}
}

func (pm *ProcessMonitor) processStreamError(streamError StreamError) {
	log.Printf("Stream Error: [%s] %s - %s", streamError.Command, streamError.ErrorType, streamError.Message)
	
	// Update metrics
	pm.metrics.mutex.Lock()
	pm.metrics.TotalErrors++
	pm.metrics.mutex.Unlock()
	
	// Update process with error
	pm.mutex.Lock()
	if process, exists := pm.activeProcesses[streamError.ProcessPID]; exists {
		process.mutex.Lock()
		process.LastError = &streamError
		process.ErrorLines = append(process.ErrorLines, streamError.Message)
		
		// Keep only recent error lines
		if len(process.ErrorLines) > pm.config.MaxOutputLines {
			process.ErrorLines = process.ErrorLines[len(process.ErrorLines)-pm.config.MaxOutputLines:]
		}
		process.mutex.Unlock()
	}
	pm.mutex.Unlock()
	
	// Broadcast to WebSocket clients
	pm.broadcastToWebSockets(streamError)
}

func (pm *ProcessMonitor) broadcastToWebSockets(streamError StreamError) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	message, _ := json.Marshal(streamError)
	
	for conn := range pm.wsConnections {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			delete(pm.wsConnections, conn)
			conn.Close()
		}
	}
}

func (pm *ProcessMonitor) handleBufferOverflow() {
	log.Println("Error stream buffer overflow - dropping oldest errors")
}

func (pm *ProcessMonitor) flushPendingErrors() {
	// Placeholder for batching logic if needed
}

func (pm *ProcessMonitor) StartProcess(cmd ProcessCommand) (*MonitoredProcess, error) {
	startTime := time.Now()
	
	log.Printf("Starting process monitoring: command=%s, args=%v, workdir=%s", 
		cmd.Command, cmd.Args, cmd.WorkingDir)
	
	defer func() {
		duration := time.Since(startTime)
		log.Printf("Process start completed in %v", duration)
	}()
	
	// Update metrics
	pm.metrics.mutex.Lock()
	pm.metrics.ProcessStartAttempts++
	pm.metrics.mutex.Unlock()
	
	// Validate inputs
	if err := pm.validateProcessCommand(cmd); err != nil {
		pm.metrics.mutex.Lock()
		pm.metrics.ProcessStartFailures++
		pm.metrics.mutex.Unlock()
		return nil, err
	}
	
	// Check process limits
	pm.mutex.RLock()
	if len(pm.activeProcesses) >= pm.config.MaxProcesses {
		pm.mutex.RUnlock()
		return nil, fmt.Errorf("maximum number of processes (%d) already running", pm.config.MaxProcesses)
	}
	pm.mutex.RUnlock()
	
	// Create the command
	execCmd := exec.Command(cmd.Command, cmd.Args...)
	execCmd.Dir = cmd.WorkingDir
	
	// Set environment variables
	if cmd.Environment != nil {
		env := os.Environ()
		for key, value := range cmd.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		execCmd.Env = env
	}
	
	// Create pipes for stdout and stderr
	stdoutPipe, err := execCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderrPipe, err := execCmd.StderrPipe()
	if err != nil {
		stdoutPipe.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	
	// Start the process
	if err := execCmd.Start(); err != nil {
		stdoutPipe.Close()
		stderrPipe.Close()
		pm.metrics.mutex.Lock()
		pm.metrics.ProcessStartFailures++
		pm.metrics.mutex.Unlock()
		return nil, fmt.Errorf("failed to start process: %w", err)
	}
	
	// Create monitored process
	process := &MonitoredProcess{
		PID:         execCmd.Process.Pid,
		Command:     cmd.Command,
		Args:        cmd.Args,
		StartTime:   time.Now(),
		Status:      "running",
		OutputLines: make([]string, 0),
		ErrorLines:  make([]string, 0),
		WorkingDir:  cmd.WorkingDir,
		cmd:         execCmd,
		stdoutPipe:  stdoutPipe,
		stderrPipe:  stderrPipe,
	}
	
	// Add to active processes
	pm.mutex.Lock()
	pm.activeProcesses[process.PID] = process
	pm.mutex.Unlock()
	
	// Update metrics
	pm.metrics.mutex.Lock()
	pm.metrics.ActiveProcesses++
	pm.metrics.mutex.Unlock()
	
	// Start output monitoring
	go pm.monitorProcessOutput(process, stdoutPipe, "stdout", cmd.ErrorPatterns)
	go pm.monitorProcessOutput(process, stderrPipe, "stderr", cmd.ErrorPatterns)
	
	// Monitor process completion
	go pm.monitorProcessCompletion(process, cmd.AutoRestart)
	
	log.Printf("Successfully started process PID %d", process.PID)
	
	return process, nil
}

func (pm *ProcessMonitor) validateProcessCommand(cmd ProcessCommand) error {
	if cmd.Command == "" {
		return errors.New("command cannot be empty")
	}
	
	if len(cmd.Command) > 1000 {
		return errors.New("command too long")
	}
	
	// Check if command exists
	if _, err := exec.LookPath(cmd.Command); err != nil {
		return fmt.Errorf("command not found: %s", cmd.Command)
	}
	
	// Validate working directory
	if cmd.WorkingDir == "" {
		cmd.WorkingDir = "."
	}
	
	if _, err := os.Stat(cmd.WorkingDir); os.IsNotExist(err) {
		return fmt.Errorf("working directory does not exist: %s", cmd.WorkingDir)
	}
	
	// Security: whitelist allowed commands
	allowed := false
	for _, allowedCmd := range pm.config.AllowedCommands {
		if cmd.Command == allowedCmd {
			allowed = true
			break
		}
	}
	
	if !allowed {
		return fmt.Errorf("command not allowed: %s", cmd.Command)
	}
	
	return nil
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
		".gitignore": "configuration",
		".env":       "configuration",
		"tsconfig.json": "configuration",
		"webpack.config.js": "configuration",
		"babel.config.js": "configuration",
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
		AppName: "Project Argus",
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
	// CORS middleware for browser compatibility
	is.app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		
		if c.Method() == "OPTIONS" {
			return c.SendStatus(204)
		}
		
		return c.Next()
	})

	// WebSocket middleware
	is.app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	
	// WebSocket routes
	is.app.Get("/ws/errors", websocket.New(is.errorStreamHandler))
	is.app.Get("/ws/processes", websocket.New(is.processStreamHandler))
	
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
	
	// Process monitoring routes
	is.app.Get("/processes/monitored", is.monitoredProcessesHandler)
	is.app.Post("/processes/start", is.startProcessHandler)
	is.app.Delete("/processes/:pid", is.stopProcessHandler)
	is.app.Get("/processes/:pid/output", is.processOutputHandler)
	
	// Real-time error streaming
	is.app.Get("/errors/stream", is.errorStreamHTTPHandler)
	is.app.Get("/errors/latest", is.latestErrorsHandler)
	
	// Development server integration
	is.app.Post("/dev/start/:type", is.startDevServerHandler)
	is.app.Post("/dev/stop/:type", is.stopDevServerHandler)
	is.app.Get("/dev/status", is.devServerStatusHandler)
	
	// File-specific routes
	is.app.Get("/files/:path", is.fileHandler)
	is.app.Get("/files/:path/content", is.fileContentHandler)
	
	// Search and query routes
	is.app.Get("/search", is.searchHandler)
	is.app.Get("/query/:type", is.queryHandler)
	
	// Action routes
	is.app.Post("/refresh", is.refreshHandler)
	is.app.Post("/analyze", is.analyzeHandler)
	
	// Server control routes
	is.app.Post("/server/stop", is.serverStopHandler)
	
	// Workspace management routes
	is.app.Post("/workspace/change", is.workspaceChangeHandler)
}

func (is *IntelligenceServer) statusHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"service":   "Project Argus",
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

func (is *IntelligenceServer) serverStopHandler(c *fiber.Ctx) error {
	log.Println("📴 Server stop request received via API")
	
	// Send response before shutting down
	response := c.JSON(fiber.Map{
		"message": "Server shutdown initiated",
		"status":  "stopping",
		"timestamp": time.Now(),
	})
	
	// Schedule shutdown after response is sent
	go func() {
		time.Sleep(500 * time.Millisecond) // Give time for response to be sent
		log.Println("🛑 Shutting down server...")
		os.Exit(0)
	}()
	
	return response
}

func (is *IntelligenceServer) workspaceChangeHandler(c *fiber.Ctx) error {
	var request struct {
		Workspace string `json:"workspace"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON format",
			"details": err.Error(),
		})
	}
	
	newWorkspace := strings.TrimSpace(request.Workspace)
	if newWorkspace == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Workspace path cannot be empty",
		})
	}
	
	// Validate the workspace path exists
	if _, err := os.Stat(newWorkspace); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{
			"error": "Workspace path does not exist",
			"path": newWorkspace,
		})
	}
	
	log.Printf("🔄 Workspace change request: %s -> %s", is.pi.workspace, newWorkspace)
	
	// Note: This doesn't actually change the workspace in the current implementation
	// The server would need to be restarted with the new workspace path
	// This endpoint provides acknowledgment for UI purposes
	
	return c.JSON(fiber.Map{
		"message": "Workspace change acknowledged",
		"old_workspace": is.pi.workspace,
		"new_workspace": newWorkspace,
		"note": "Server restart required for full workspace monitoring change",
		"restart_command": fmt.Sprintf("/usr/local/go/bin/go run main.go \"%s\"", newWorkspace),
		"timestamp": time.Now(),
	})
}

// WebSocket handlers
func (is *IntelligenceServer) errorStreamHandler(c *websocket.Conn) {
	// Add connection to process monitor
	is.pi.processMonitor.AddWebSocketConnection(c)
	defer is.pi.processMonitor.RemoveWebSocketConnection(c)
	
	// Send initial connection message
	initialMsg := map[string]interface{}{
		"type":      "connection",
		"message":   "Error stream connected",
		"timestamp": time.Now(),
	}
	
	if data, err := json.Marshal(initialMsg); err == nil {
		c.WriteMessage(websocket.TextMessage, data)
	}
	
	// Keep connection alive and handle messages
	for {
		// Read message (ping/pong for keepalive)
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
	}
}

func (is *IntelligenceServer) processStreamHandler(c *websocket.Conn) {
	// Add connection to process monitor
	is.pi.processMonitor.AddWebSocketConnection(c)
	defer is.pi.processMonitor.RemoveWebSocketConnection(c)
	
	// Send current process status
	processes := is.pi.processMonitor.GetMonitoredProcesses()
	statusMsg := map[string]interface{}{
		"type":      "process_status",
		"processes": processes,
		"timestamp": time.Now(),
	}
	
	if data, err := json.Marshal(statusMsg); err == nil {
		c.WriteMessage(websocket.TextMessage, data)
	}
	
	// Send updates every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			processes := is.pi.processMonitor.GetMonitoredProcesses()
			statusMsg := map[string]interface{}{
				"type":      "process_update",
				"processes": processes,
				"timestamp": time.Now(),
			}
			
			if data, err := json.Marshal(statusMsg); err == nil {
				if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Printf("WebSocket write error: %v", err)
					return
				}
			}
		default:
			// Check if connection is still alive
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// Process monitoring handlers
func (is *IntelligenceServer) monitoredProcessesHandler(c *fiber.Ctx) error {
	processes := is.pi.processMonitor.GetMonitoredProcesses()
	
	return c.JSON(fiber.Map{
		"processes": processes,
		"count":     len(processes),
		"timestamp": time.Now(),
	})
}

func (is *IntelligenceServer) startProcessHandler(c *fiber.Ctx) error {
	var cmd ProcessCommand
	if err := c.BodyParser(&cmd); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
	}
	
	// Rate limiting check
	if !is.checkRateLimit(c.IP()) {
		return c.Status(429).JSON(fiber.Map{
			"error": "Rate limit exceeded",
		})
	}
	
	// Validate and start process
	process, err := is.pi.processMonitor.StartProcess(cmd)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to start process",
			"details": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "Process started successfully",
		"process": process,
	})
}

func (is *IntelligenceServer) stopProcessHandler(c *fiber.Ctx) error {
	pidStr := c.Params("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid PID",
		})
	}
	
	if err := is.pi.processMonitor.StopProcess(pid); err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error":   "Failed to stop process",
			"details": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Process %d stopped successfully", pid),
	})
}

func (is *IntelligenceServer) processOutputHandler(c *fiber.Ctx) error {
	pidStr := c.Params("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid PID",
		})
	}
	
	lines := c.QueryInt("lines", 50)
	output, err := is.pi.processMonitor.GetProcessOutput(pid, lines)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error":   "Process not found",
			"details": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"pid":    pid,
		"lines":  lines,
		"output": output,
	})
}

func (is *IntelligenceServer) errorStreamHTTPHandler(c *fiber.Ctx) error {
	since := c.Query("since", "5m")
	duration, err := time.ParseDuration(since)
	if err != nil {
		duration = 5 * time.Minute
	}
	
	errors := is.pi.processMonitor.GetLatestErrors(duration)
	
	return c.JSON(fiber.Map{
		"since":       since,
		"error_count": len(errors),
		"errors":      errors,
		"timestamp":   time.Now(),
	})
}

func (is *IntelligenceServer) latestErrorsHandler(c *fiber.Ctx) error {
	since := c.Query("since", "5m")
	duration, err := time.ParseDuration(since)
	if err != nil {
		duration = 5 * time.Minute
	}
	
	cutoff := time.Now().Add(-duration)
	recentErrors := is.pi.processMonitor.GetLatestErrors(duration)
	
	return c.JSON(fiber.Map{
		"since":       since,
		"cutoff":      cutoff,
		"error_count": len(recentErrors),
		"errors":      recentErrors,
	})
}

func (is *IntelligenceServer) startDevServerHandler(c *fiber.Ctx) error {
	serverType := c.Params("type")
	
	commands := map[string]ProcessCommand{
		"npm": {
			Command:       "npm",
			Args:          []string{"run", "dev"},
			WorkingDir:    is.pi.workspace,
			AutoRestart:   true,
			ErrorPatterns: []string{"Error:", "Failed to compile", "Module not found"},
		},
		"go": {
			Command:       "go",
			Args:          []string{"run", "main.go"},
			WorkingDir:    is.pi.workspace,
			AutoRestart:   true,
			ErrorPatterns: []string{"panic:", "fatal error:", "cannot find package"},
		},
		"python": {
			Command:       "python",
			Args:          []string{"app.py"},
			WorkingDir:    is.pi.workspace,
			AutoRestart:   false,
			ErrorPatterns: []string{"Traceback", "ImportError:", "SyntaxError:"},
		},
		"next": {
			Command:       "npm",
			Args:          []string{"run", "dev"},
			WorkingDir:    is.pi.workspace,
			AutoRestart:   true,
			ErrorPatterns: []string{"Error:", "Failed to compile", "Module not found"},
		},
		"vite": {
			Command:       "npm",
			Args:          []string{"run", "dev"},
			WorkingDir:    is.pi.workspace,
			AutoRestart:   true,
			ErrorPatterns: []string{"Error:", "Failed to compile", "Module not found"},
		},
	}
	
	cmd, exists := commands[serverType]
	if !exists {
		return c.Status(400).JSON(fiber.Map{
			"error":         "Unknown server type",
			"supported":     []string{"npm", "go", "python", "next", "vite"},
			"server_type":   serverType,
		})
	}
	
	process, err := is.pi.processMonitor.StartProcess(cmd)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   fmt.Sprintf("Failed to start %s server", serverType),
			"details": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"message":     fmt.Sprintf("Started %s development server", serverType),
		"server_type": serverType,
		"process":     process,
	})
}

func (is *IntelligenceServer) stopDevServerHandler(c *fiber.Ctx) error {
	serverType := c.Params("type")
	
	// Find processes by server type and stop them
	processes := is.pi.processMonitor.GetMonitoredProcesses()
	stopped := 0
	
	for _, process := range processes {
		// Simple matching by command - could be enhanced
		if strings.Contains(process.Command, serverType) {
			if err := is.pi.processMonitor.StopProcess(process.PID); err != nil {
				log.Printf("Failed to stop process %d: %v", process.PID, err)
			} else {
				stopped++
			}
		}
	}
	
	return c.JSON(fiber.Map{
		"message":       fmt.Sprintf("Stopped %s development server", serverType),
		"server_type":   serverType,
		"processes_stopped": stopped,
	})
}

func (is *IntelligenceServer) devServerStatusHandler(c *fiber.Ctx) error {
	processes := is.pi.processMonitor.GetMonitoredProcesses()
	
	devServers := make(map[string]interface{})
	
	for _, process := range processes {
		serverType := "unknown"
		
		// Determine server type based on command
		switch {
		case process.Command == "npm":
			if len(process.Args) > 1 && process.Args[1] == "dev" {
				serverType = "npm/next/vite"
			}
		case process.Command == "go":
			serverType = "go"
		case process.Command == "python":
			serverType = "python"
		}
		
		devServers[fmt.Sprintf("%s_%d", serverType, process.PID)] = map[string]interface{}{
			"pid":        process.PID,
			"command":    process.Command,
			"args":       process.Args,
			"status":     process.Status,
			"start_time": process.StartTime,
			"uptime":     time.Since(process.StartTime).String(),
		}
	}
	
	return c.JSON(fiber.Map{
		"dev_servers": devServers,
		"total":       len(devServers),
		"timestamp":   time.Now(),
	})
}

// Rate limiting helper
func (is *IntelligenceServer) checkRateLimit(ip string) bool {
	// Simple in-memory rate limiting
	// In production, you'd use Redis or similar
	return true // Simplified for now
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
	// Set up graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		
		// Cleanup all monitored processes
		is.pi.processMonitor.StopAllProcesses()
		
		// Close WebSocket connections (handled by StopAllProcesses)
		
		// Shutdown server
		if err := is.app.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		
		log.Println("Project Argus shutdown completed")
		os.Exit(0)
	}()
	
	log.Printf("Monitoring workspace: %s", is.pi.workspace)
	log.Printf("API endpoints available at: http://localhost%s/", port)
	log.Printf("WebSocket endpoints: ws://localhost%s/ws/errors, ws://localhost%s/ws/processes", port, port)
	
	return is.app.Listen(port)
}

// Additional ProcessMonitor methods
func (pm *ProcessMonitor) monitorProcessOutput(process *MonitoredProcess, pipe io.ReadCloser, source string, errorPatterns []string) {
	defer pipe.Close()
	
	scanner := bufio.NewScanner(pipe)
	contextLines := make([]string, 0, 5) // Keep context for error detection
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Add to process output
		process.mutex.Lock()
		process.OutputLines = append(process.OutputLines, line)
		
		// Keep only recent output lines
		if len(process.OutputLines) > pm.config.MaxOutputLines {
			process.OutputLines = process.OutputLines[len(process.OutputLines)-pm.config.MaxOutputLines:]
		}
		process.mutex.Unlock()
		
		// Maintain context window
		contextLines = append(contextLines, line)
		if len(contextLines) > 5 {
			contextLines = contextLines[1:]
		}
		
		// Check for error patterns
		if streamError := pm.parseOutputForErrors(line, source, process, contextLines, errorPatterns); streamError != nil {
			select {
			case pm.errorStream <- *streamError:
			case <-pm.ctx.Done():
				return
			default:
				// Channel full, drop error
				log.Printf("Error stream full, dropping error from PID %d", process.PID)
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading process output (PID %d): %v", process.PID, err)
	}
}

func (pm *ProcessMonitor) parseOutputForErrors(line, source string, process *MonitoredProcess, context []string, customPatterns []string) *StreamError {
	// Combine default patterns with custom patterns
	allPatterns := append(pm.getDefaultErrorPatterns(), customPatterns...)
	
	for _, pattern := range allPatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			severity := "error"
			errorType := "runtime"
			
			// Determine error type and severity based on pattern
			if matched, _ := regexp.MatchString(`(?i)(warn|warning)`, line); matched {
				severity = "warning"
			}
			
			if matched, _ := regexp.MatchString(`(?i)(syntax|parse|compile)`, line); matched {
				errorType = "compilation"
			} else if matched, _ := regexp.MatchString(`(?i)(test|spec|assertion)`, line); matched {
				errorType = "test"
			} else if matched, _ := regexp.MatchString(`(?i)(server|port|listen|connect)`, line); matched {
				errorType = "server"
			}
			
			return &StreamError{
				ProcessPID: process.PID,
				Command:    process.Command,
				ErrorType:  errorType,
				Message:    line,
				Timestamp:  time.Now(),
				Severity:   severity,
				Context:    append([]string{}, context...),
				Source:     source,
			}
		}
	}
	
	return nil
}

func (pm *ProcessMonitor) getDefaultErrorPatterns() []string {
	return []string{
		// JavaScript/TypeScript errors
		`Error: .+`,
		`TypeError: .+`,
		`SyntaxError: .+`,
		`Module not found: .+`,
		`Failed to compile`,
		`Compilation error`,
		
		// Go errors
		`panic: .+`,
		`fatal error: .+`,
		`cannot find package ".+"`,
		`undefined: .+`,
		
		// Python errors
		`Traceback \(most recent call last\)`,
		`ImportError: .+`,
		`SyntaxError: .+`,
		`NameError: .+`,
		
		// Test errors
		`FAIL .+`,
		`Error: expect\(.+\)`,
		`AssertionError`,
		
		// Server errors
		`EADDRINUSE .+`,
		`listen EADDRINUSE .+`,
		`Connection refused`,
		`ECONNREFUSED`,
		
		// General patterns
		`(?i)error.*`,
		`(?i)exception.*`,
		`(?i)failed.*`,
		`(?i)fatal.*`,
	}
}

func (pm *ProcessMonitor) monitorProcessCompletion(process *MonitoredProcess, autoRestart bool) {
	err := process.cmd.Wait()
	
	process.mutex.Lock()
	if err != nil {
		process.Status = "error"
		log.Printf("Process PID %d exited with error: %v", process.PID, err)
	} else {
		process.Status = "stopped"
		log.Printf("Process PID %d exited normally", process.PID)
	}
	process.mutex.Unlock()
	
	// Update metrics
	pm.metrics.mutex.Lock()
	pm.metrics.ActiveProcesses--
	pm.metrics.mutex.Unlock()
	
	// Handle auto-restart if enabled
	if autoRestart && err != nil {
		log.Printf("Auto-restarting process: %s", process.Command)
		// Implement restart logic here if needed
	}
}

func (pm *ProcessMonitor) StopProcess(pid int) error {
	pm.mutex.Lock()
	process, exists := pm.activeProcesses[pid]
	if !exists {
		pm.mutex.Unlock()
		return fmt.Errorf("process with PID %d not found", pid)
	}
	pm.mutex.Unlock()
	
	log.Printf("Stopping process PID %d", pid)
	
	// Graceful shutdown first
	if err := process.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// If graceful shutdown fails, force kill
		if killErr := process.cmd.Process.Kill(); killErr != nil {
			return fmt.Errorf("failed to kill process: %w", killErr)
		}
	}
	
	// Wait for process to exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- process.cmd.Wait()
	}()
	
	select {
	case <-done:
		log.Printf("Process PID %d stopped successfully", pid)
	case <-time.After(5 * time.Second):
		// Force kill if still running
		process.cmd.Process.Kill()
		log.Printf("Force killed process PID %d", pid)
	}
	
	// Cleanup
	pm.cleanupProcess(process)
	
	return nil
}

func (pm *ProcessMonitor) cleanupProcess(process *MonitoredProcess) {
	pm.mutex.Lock()
	delete(pm.activeProcesses, process.PID)
	pm.mutex.Unlock()
	
	// Close pipes
	if process.stdoutPipe != nil {
		process.stdoutPipe.Close()
	}
	if process.stderrPipe != nil {
		process.stderrPipe.Close()
	}
	
	log.Printf("Cleaned up process PID %d", process.PID)
}

func (pm *ProcessMonitor) GetProcessOutput(pid int, lines int) ([]string, error) {
	pm.mutex.RLock()
	process, exists := pm.activeProcesses[pid]
	if !exists {
		pm.mutex.RUnlock()
		return nil, fmt.Errorf("process with PID %d not found", pid)
	}
	pm.mutex.RUnlock()
	
	process.mutex.RLock()
	defer process.mutex.RUnlock()
	
	totalLines := len(process.OutputLines)
	if lines <= 0 || lines > totalLines {
		lines = totalLines
	}
	
	if totalLines == 0 {
		return []string{}, nil
	}
	
	start := totalLines - lines
	return process.OutputLines[start:], nil
}

func (pm *ProcessMonitor) cleanupStoppedProcesses() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	for pid, process := range pm.activeProcesses {
		if !pm.isProcessRunning(pid) {
			log.Printf("Cleaning up stopped process PID %d", pid)
			pm.cleanupProcess(process)
		}
	}
}

func (pm *ProcessMonitor) isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (pm *ProcessMonitor) autoDetectDevProcesses() {
	// This could be expanded to automatically detect and monitor
	// common development processes running in the workspace
	log.Println("Auto-detection of development processes started")
	
	// For now, this is a placeholder for future enhancement
	// Could scan for package.json, go.mod, etc. and automatically
	// start monitoring relevant dev servers
}

func (pm *ProcessMonitor) GetMonitoredProcesses() []*MonitoredProcess {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	processes := make([]*MonitoredProcess, 0, len(pm.activeProcesses))
	for _, process := range pm.activeProcesses {
		processes = append(processes, process)
	}
	
	return processes
}

func (pm *ProcessMonitor) AddWebSocketConnection(conn *websocket.Conn) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	pm.wsConnections[conn] = true
	log.Printf("Added WebSocket connection, total: %d", len(pm.wsConnections))
}

func (pm *ProcessMonitor) RemoveWebSocketConnection(conn *websocket.Conn) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	delete(pm.wsConnections, conn)
	log.Printf("Removed WebSocket connection, total: %d", len(pm.wsConnections))
}

func (pm *ProcessMonitor) StopAllProcesses() {
	pm.mutex.RLock()
	pids := make([]int, 0, len(pm.activeProcesses))
	for pid := range pm.activeProcesses {
		pids = append(pids, pid)
	}
	pm.mutex.RUnlock()
	
	log.Printf("Stopping all %d monitored processes", len(pids))
	
	for _, pid := range pids {
		if err := pm.StopProcess(pid); err != nil {
			log.Printf("Error stopping process PID %d: %v", pid, err)
		}
	}
	
	// Cancel context to stop all monitoring goroutines
	pm.cancel()
}

func (pm *ProcessMonitor) GetLatestErrors(since time.Duration) []StreamError {
	cutoff := time.Now().Add(-since)
	errors := make([]StreamError, 0)
	
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	for _, process := range pm.activeProcesses {
		process.mutex.RLock()
		if process.LastError != nil && process.LastError.Timestamp.After(cutoff) {
			errors = append(errors, *process.LastError)
		}
		process.mutex.RUnlock()
	}
	
	return errors
}

func main() {
	// Get workspace from command line argument or environment variable
	workspace := "."
	if len(os.Args) > 1 {
		workspace = os.Args[1]
	} else if envWorkspace := os.Getenv("ARGUS_WORKSPACE"); envWorkspace != "" {
		workspace = envWorkspace
	} else if envWorkspace := os.Getenv("WORKSPACE_PATH"); envWorkspace != "" {
		workspace = envWorkspace
	}

	// Get absolute workspace path
	absWorkspace, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	if workspace != "." {
		absWorkspace = workspace
	}

	log.Printf("Starting Enhanced Project Argus monitoring for: %s", absWorkspace)

	// Create enhanced intelligence service with multi-language support
	server := NewEnhancedIntelligenceServer(absWorkspace)

	// Get port from environment or use default
	port := ":3002"
	if portEnv := os.Getenv("ARGUS_PORT"); portEnv != "" {
		port = ":" + portEnv
	} else if portEnv := os.Getenv("CLAUDE_INTEL_PORT"); portEnv != "" {
		port = ":" + portEnv
	}

	log.Printf("Starting Enhanced Project Argus on port %s", port)
	log.Printf("Enhanced multi-language monitoring for: %s", absWorkspace)
	
	if err := server.Start(port); err != nil {
		log.Fatalf("Failed to start enhanced server: %v", err)
	}
}