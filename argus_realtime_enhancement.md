# Project Argus Real-Time Console Enhancement Specification

## 🎯 Enhancement Objective

Extend the existing Project Argus system to capture **real-time console errors** from running processes and development servers, then stream them directly to Claude Code as they occur.

## 📋 Requirements for Claude 4 Sonnet

**Please read the existing Project Argus codebase and apply these enhancements. Do NOT rewrite the entire system - just add these new capabilities to the existing `main.go`, `argus-query.sh`, and `dashboard.html` files.**

---

## 🔧 Core Enhancements Needed

### 1. Real-Time Process Monitoring

Add these new types to the existing Go code:

```go
// Add to existing types in main.go

type ProcessMonitor struct {
	activeProcesses map[int]*MonitoredProcess
	errorStream     chan StreamError
	mutex           sync.RWMutex
}

type MonitoredProcess struct {
	PID         int       `json:"pid"`
	Command     string    `json:"command"`
	StartTime   time.Time `json:"start_time"`
	Status      string    `json:"status"` // running, stopped, error
	OutputLines []string  `json:"output_lines"`
	ErrorLines  []string  `json:"error_lines"`
	LastError   *StreamError `json:"last_error,omitempty"`
}

type StreamError struct {
	ProcessPID  int       `json:"process_pid"`
	Command     string    `json:"command"`
	ErrorType   string    `json:"error_type"` // runtime, compilation, test, server
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"` // error, warning, info
	Context     []string  `json:"context"` // surrounding lines
	Source      string    `json:"source"` // stdout, stderr
}

type ProcessCommand struct {
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	WorkingDir  string            `json:"working_dir"`
	Environment map[string]string `json:"environment"`
	AutoRestart bool              `json:"auto_restart"`
	ErrorPatterns []string        `json:"error_patterns"`
}
```

### 2. Process Monitoring Implementation

Add this new functionality to the `ProjectIntelligence` struct:

```go
// Add to ProjectIntelligence struct
processMonitor *ProcessMonitor

// Add to NewProjectIntelligence function
processMonitor: &ProcessMonitor{
	activeProcesses: make(map[int]*MonitoredProcess),
	errorStream:     make(chan StreamError, 1000),
},

// Add to StartWatching function
go pi.processMonitor.startMonitoring()
```

**New Methods to Add:**

```go
func (pm *ProcessMonitor) startMonitoring() {
	log.Println("Process monitor started")
	
	// Start error stream processor
	go pm.processErrorStream()
	
	// Monitor for common development processes
	go pm.autoDetectDevProcesses()
	
	// Cleanup stopped processes
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		pm.cleanupStoppedProcesses()
	}
}

func (pm *ProcessMonitor) processErrorStream() {
	for streamError := range pm.errorStream {
		log.Printf("Stream Error: %s - %s", streamError.Command, streamError.Message)
		// Add to main error watcher
		// Trigger WebSocket notifications
		pm.notifyClients(streamError)
	}
}

func (pm *ProcessMonitor) StartProcess(cmd ProcessCommand) (*MonitoredProcess, error) {
	// Start process with output monitoring
	// Attach stdout/stderr readers
	// Parse output for error patterns
	// Send errors to errorStream channel
}

func (pm *ProcessMonitor) StopProcess(pid int) error {
	// Stop monitoring and kill process
}

func (pm *ProcessMonitor) GetProcessOutput(pid int, lines int) ([]string, error) {
	// Return recent output lines from process
}

func (pm *ProcessMonitor) autoDetectDevProcesses() {
	// Automatically detect and monitor common dev commands:
	// - npm run dev, npm start, npm test
	// - go run main.go, go test ./...
	// - python app.py, python manage.py runserver
	// - yarn dev, yarn start
	// - next dev, vite dev
}

func (pm *ProcessMonitor) parseOutputForErrors(line string, source string) *StreamError {
	// Parse common error patterns:
	
	errorPatterns := map[string]string{
		// JavaScript/TypeScript errors
		`Error: (.+)`:                    "runtime",
		`TypeError: (.+)`:                "runtime", 
		`SyntaxError: (.+)`:              "compilation",
		`Module not found: (.+)`:         "compilation",
		`Failed to compile`:              "compilation",
		
		// Go errors  
		`panic: (.+)`:                    "runtime",
		`fatal error: (.+)`:              "runtime",
		`cannot find package "(.+)"`:     "compilation",
		
		// Python errors
		`Traceback \(most recent call last\)`: "runtime",
		`ImportError: (.+)`:              "compilation",
		`SyntaxError: (.+)`:              "compilation",
		
		// Test errors
		`FAIL (.+)`:                      "test",
		`Error: expect\((.+)\)`:          "test",
		
		// Server errors
		`EADDRINUSE (.+)`:                "server",
		`listen EADDRINUSE (.+)`:         "server",
		`Connection refused`:             "server",
	}
	
	// Return parsed StreamError or nil
}
```

### 3. WebSocket Real-Time Streaming

Add WebSocket support for real-time error streaming:

```go
// Add to imports
"github.com/gofiber/websocket/v2"

// Add to setupRoutes() method
is.app.Use("/ws", func(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
})

is.app.Get("/ws/errors", websocket.New(is.errorStreamHandler))
is.app.Get("/ws/processes", websocket.New(is.processStreamHandler))

// New WebSocket handlers
func (is *IntelligenceServer) errorStreamHandler(c *websocket.Conn) {
	// Stream real-time errors to connected clients
}

func (is *IntelligenceServer) processStreamHandler(c *websocket.Conn) {
	// Stream process status updates
}
```

### 4. New API Endpoints

Add these new routes to `setupRoutes()`:

```go
// Process monitoring routes
is.app.Get("/processes/monitored", is.monitoredProcessesHandler)
is.app.Post("/processes/start", is.startProcessHandler)
is.app.Delete("/processes/:pid", is.stopProcessHandler)
is.app.Get("/processes/:pid/output", is.processOutputHandler)

// Real-time error streaming  
is.app.Get("/errors/stream", is.errorStreamHandler)
is.app.Get("/errors/latest", is.latestErrorsHandler)

// Development server integration
is.app.Post("/dev/start/:type", is.startDevServerHandler) // npm, go, python
is.app.Post("/dev/stop/:type", is.stopDevServerHandler)
is.app.Get("/dev/status", is.devServerStatusHandler)
```

**New Handler Methods:**

```go
func (is *IntelligenceServer) startProcessHandler(c *fiber.Ctx) error {
	var cmd ProcessCommand
	if err := c.BodyParser(&cmd); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid command format"})
	}
	
	process, err := is.pi.processMonitor.StartProcess(cmd)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	return c.JSON(process)
}

func (is *IntelligenceServer) processOutputHandler(c *fiber.Ctx) error {
	pidStr := c.Params("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid PID"})
	}
	
	lines := c.QueryInt("lines", 50)
	output, err := is.pi.processMonitor.GetProcessOutput(pid, lines)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Process not found"})
	}
	
	return c.JSON(fiber.Map{
		"pid":    pid,
		"lines":  lines,
		"output": output,
	})
}

func (is *IntelligenceServer) startDevServerHandler(c *fiber.Ctx) error {
	serverType := c.Params("type")
	
	commands := map[string]ProcessCommand{
		"npm": {
			Command: "npm",
			Args:    []string{"run", "dev"},
			WorkingDir: is.pi.workspace,
			AutoRestart: true,
			ErrorPatterns: []string{"Error:", "Failed to compile", "Module not found"},
		},
		"go": {
			Command: "go",
			Args:    []string{"run", "main.go"},
			WorkingDir: is.pi.workspace,
			AutoRestart: true,
			ErrorPatterns: []string{"panic:", "fatal error:", "cannot find package"},
		},
		"python": {
			Command: "python",
			Args:    []string{"app.py"},
			WorkingDir: is.pi.workspace,
			AutoRestart: false,
			ErrorPatterns: []string{"Traceback", "ImportError:", "SyntaxError:"},
		},
	}
	
	cmd, exists := commands[serverType]
	if !exists {
		return c.Status(400).JSON(fiber.Map{"error": "Unknown server type"})
	}
	
	process, err := is.pi.processMonitor.StartProcess(cmd)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Started %s development server", serverType),
		"process": process,
	})
}

func (is *IntelligenceServer) latestErrorsHandler(c *fiber.Ctx) error {
	since := c.Query("since", "5m")
	duration, err := time.ParseDuration(since)
	if err != nil {
		duration = 5 * time.Minute
	}
	
	cutoff := time.Now().Add(-duration)
	
	// Get recent stream errors
	recentErrors := []StreamError{}
	// Filter errors since cutoff time
	
	return c.JSON(fiber.Map{
		"since":       since,
		"cutoff":      cutoff,
		"error_count": len(recentErrors),
		"errors":      recentErrors,
	})
}
```

### 5. Enhanced CLI Tool Features

Add these new commands to `argus-query.sh`:

```bash
# Add to cmd_help() function
echo -e "  ${GREEN}monitor${NC} \"command\" - Start monitoring a command"
echo -e "  ${GREEN}processes${NC}    - Show monitored processes"
echo -e "  ${GREEN}logs${NC} [pid]    - Show process output"
echo -e "  ${GREEN}dev${NC} [start|stop|status] [type] - Manage dev servers"
echo -e "  ${GREEN}stream${NC}       - Stream real-time errors"

# New command functions
cmd_monitor() {
    local command="$1"
    
    if [ -z "$command" ]; then
        echo -e "${RED}Error: Command to monitor is required${NC}"
        echo "Usage: $0 monitor \"npm run dev\""
        return 1
    fi
    
    print_header "Starting Process Monitor"
    
    # Parse command into parts
    IFS=' ' read -ra cmd_parts <<< "$command"
    
    # Create JSON payload
    payload=$(jq -n \
        --arg cmd "${cmd_parts[0]}" \
        --argjson args "$(printf '%s\n' "${cmd_parts[@]:1}" | jq -R . | jq -s .)" \
        --arg wd "$(pwd)" \
        '{
            command: $cmd,
            args: $args,
            working_dir: $wd,
            auto_restart: true,
            error_patterns: ["Error:", "error:", "ERROR", "Failed", "Exception"]
        }'
    )
    
    response=$(curl -s -X POST "$BASE_URL/processes/start" \
        -H "Content-Type: application/json" \
        -d "$payload")
    
    if [ $? -eq 0 ]; then
        pid=$(echo "$response" | jq -r '.pid // "unknown"')
        echo -e "${GREEN}✅ Started monitoring process PID: $pid${NC}"
        echo -e "${BLUE}Command: $command${NC}"
        echo -e "${YELLOW}Use '$0 logs $pid' to see output${NC}"
        echo -e "${YELLOW}Use '$0 stream' for real-time errors${NC}"
    else
        echo -e "${RED}❌ Failed to start monitoring${NC}"
        echo "$response" | jq . 2>/dev/null || echo "$response"
    fi
}

cmd_monitored_processes() {
    print_header "Monitored Processes"
    
    processes_data=$(query_api "/processes/monitored")
    
    if [ $? -eq 0 ]; then
        process_count=$(echo "$processes_data" | jq length)
        
        if [ "$process_count" -eq 0 ]; then
            echo -e "${YELLOW}No processes being monitored${NC}"
            echo -e "${BLUE}Use '$0 monitor \"command\"' to start monitoring${NC}"
        else
            echo -e "${GREEN}$process_count monitored process(es):${NC}\n"
            
            echo "$processes_data" | jq -r '.[] | "\(.pid) \(.command) \(.status) \(.start_time)"' | while read pid command status start_time; do
                formatted_time=$(date -d "$start_time" '+%H:%M:%S' 2>/dev/null || echo "$start_time")
                
                case "$status" in
                    "running")
                        echo -e "${GREEN}⚡ PID $pid: $command (started $formatted_time)${NC}"
                        ;;
                    "stopped")
                        echo -e "${YELLOW}⏸️  PID $pid: $command (stopped)${NC}"
                        ;;
                    "error")
                        echo -e "${RED}❌ PID $pid: $command (error)${NC}"
                        ;;
                    *)
                        echo -e "${BLUE}🔄 PID $pid: $command ($status)${NC}"
                        ;;
                esac
            done
        fi
    fi
}

cmd_process_logs() {
    local pid="$1"
    local lines="${2:-50}"
    
    if [ -z "$pid" ]; then
        echo -e "${RED}Error: Process PID required${NC}"
        echo "Usage: $0 logs <pid> [lines]"
        return 1
    fi
    
    print_header "Process Output (PID: $pid)"
    
    log_data=$(query_api "/processes/$pid/output?lines=$lines")
    
    if [ $? -eq 0 ]; then
        echo "$log_data" | jq -r '.output[]?' | while read line; do
            echo -e "${CYAN}$line${NC}"
        done
    else
        echo -e "${RED}❌ Process not found or no output available${NC}"
    fi
}

cmd_dev_server() {
    local action="$1"
    local server_type="$2"
    
    case "$action" in
        "start")
            if [ -z "$server_type" ]; then
                echo -e "${RED}Error: Server type required${NC}"
                echo "Usage: $0 dev start [npm|go|python]"
                return 1
            fi
            
            print_header "Starting $server_type Development Server"
            
            response=$(curl -s -X POST "$BASE_URL/dev/start/$server_type")
            
            if [ $? -eq 0 ]; then
                echo -e "${GREEN}✅ Started $server_type development server${NC}"
                echo "$response" | jq . 2>/dev/null
            else
                echo -e "${RED}❌ Failed to start $server_type server${NC}"
            fi
            ;;
            
        "stop")
            if [ -z "$server_type" ]; then
                echo -e "${RED}Error: Server type required${NC}"
                echo "Usage: $0 dev stop [npm|go|python]"
                return 1
            fi
            
            response=$(curl -s -X POST "$BASE_URL/dev/stop/$server_type")
            echo -e "${YELLOW}Stopped $server_type development server${NC}"
            ;;
            
        "status")
            print_header "Development Server Status"
            query_api "/dev/status" pretty
            ;;
            
        *)
            echo -e "${RED}Error: Unknown dev server action${NC}"
            echo "Usage: $0 dev [start|stop|status] [type]"
            ;;
    esac
}

cmd_stream_errors() {
    print_header "Real-Time Error Stream"
    echo -e "${YELLOW}Streaming errors... Press Ctrl+C to stop${NC}\n"
    
    # Use curl to stream errors (fallback to polling)
    if command -v websocat &> /dev/null; then
        # Use websocat for WebSocket if available
        websocat ws://localhost:3002/ws/errors
    else
        # Fallback to polling
        while true; do
            latest_errors=$(query_api "/errors/latest?since=10s")
            
            if [ $? -eq 0 ]; then
                error_count=$(echo "$latest_errors" | jq '.error_count // 0')
                
                if [ "$error_count" -gt 0 ]; then
                    echo "$latest_errors" | jq -r '.errors[] | "\(.timestamp) [\(.command)] \(.message)"' | while read line; do
                        echo -e "${RED}🚨 $line${NC}"
                    done
                fi
            fi
            
            sleep 2
        done
    fi
}

# Add to main() function cases
"monitor")
    cmd_monitor "$2"
    ;;
"processes"|"proc")
    cmd_monitored_processes
    ;;
"logs"|"log")
    cmd_process_logs "$2" "$3"
    ;;
"dev")
    cmd_dev_server "$2" "$3"
    ;;
"stream")
    cmd_stream_errors
    ;;
```

### 6. Enhanced Dashboard Features

Add to the existing `dashboard.html`:

```html
<!-- Add to the grid-3 section -->
<div class="card">
    <div class="card-header">
        <div class="card-title">⚡ Monitored Processes</div>
        <span class="btn copy-btn" onclick="copyEndpoint('/processes/monitored')">API</span>
    </div>
    <div class="scrollable" id="monitored-processes-list">
        <div style="text-align: center; color: #8b949e; padding: 20px;">No monitored processes</div>
    </div>
    <div style="margin-top: 10px; text-align: center;">
        <button class="btn btn-secondary" onclick="startDevServer()">Start Dev Server</button>
    </div>
</div>

<div class="card">
    <div class="card-header">
        <div class="card-title">🚨 Live Error Stream</div>
        <span class="btn copy-btn" onclick="toggleErrorStream()">Stream</span>
    </div>
    <div class="scrollable" id="error-stream-list">
        <div style="text-align: center; color: #8b949e; padding: 20px;">No real-time errors</div>
    </div>
    <div style="margin-top: 10px;">
        <input type="text" class="query-input" placeholder="Monitor command..." id="monitor-input" onkeypress="handleMonitorCommand(event)">
    </div>
</div>

<!-- Add to JavaScript section -->
<script>
let errorStreamWS = null;
let isStreamingErrors = false;

function updateMonitoredProcesses(processes) {
    const container = document.getElementById('monitored-processes-list');
    
    if (processes.length === 0) {
        container.innerHTML = '<div style="text-align: center; color: #8b949e; padding: 20px;">No monitored processes</div>';
        return;
    }
    
    container.innerHTML = processes.map(proc => `
        <div class="change-item">
            <div style="display: flex; justify-content: space-between; align-items: flex-start;">
                <div>
                    <div class="file-path">PID ${proc.pid}: ${proc.command}</div>
                    <div style="color: #8b949e; font-size: 0.8rem;">
                        Status: ${proc.status} • Started: ${new Date(proc.start_time).toLocaleTimeString()}
                    </div>
                </div>
                <button class="btn copy-btn" onclick="viewProcessLogs(${proc.pid})">Logs</button>
            </div>
        </div>
    `).join('');
}

function toggleErrorStream() {
    if (isStreamingErrors) {
        stopErrorStream();
    } else {
        startErrorStream();
    }
}

function startErrorStream() {
    if (!window.WebSocket) {
        // Fallback to polling
        startErrorPolling();
        return;
    }
    
    try {
        errorStreamWS = new WebSocket('ws://localhost:3002/ws/errors');
        
        errorStreamWS.onopen = function() {
            isStreamingErrors = true;
            console.log('Error stream connected');
            document.querySelector('[onclick="toggleErrorStream()"]').textContent = 'Stop';
        };
        
        errorStreamWS.onmessage = function(event) {
            const error = JSON.parse(event.data);
            addStreamError(error);
        };
        
        errorStreamWS.onclose = function() {
            isStreamingErrors = false;
            document.querySelector('[onclick="toggleErrorStream()"]').textContent = 'Stream';
        };
        
        errorStreamWS.onerror = function() {
            console.error('WebSocket error, falling back to polling');
            startErrorPolling();
        };
    } catch (error) {
        startErrorPolling();
    }
}

function stopErrorStream() {
    if (errorStreamWS) {
        errorStreamWS.close();
        errorStreamWS = null;
    }
    isStreamingErrors = false;
    document.querySelector('[onclick="toggleErrorStream()"]').textContent = 'Stream';
}

function addStreamError(error) {
    const container = document.getElementById('error-stream-list');
    
    const errorDiv = document.createElement('div');
    errorDiv.className = 'error-item';
    errorDiv.innerHTML = `
        <div style="display: flex; justify-content: space-between; align-items: flex-start;">
            <div>
                <div class="file-path">[${error.command}] ${error.error_type}</div>
                <div class="message">${error.message}</div>
            </div>
            <div class="timestamp">${new Date(error.timestamp).toLocaleTimeString()}</div>
        </div>
    `;
    
    container.insertBefore(errorDiv, container.firstChild);
    
    // Keep only last 20 errors
    while (container.children.length > 20) {
        container.removeChild(container.lastChild);
    }
}

function startErrorPolling() {
    if (isStreamingErrors) return;
    
    isStreamingErrors = true;
    document.querySelector('[onclick="toggleErrorStream()"]').textContent = 'Stop';
    
    const pollErrors = async () => {
        if (!isStreamingErrors) return;
        
        try {
            const response = await fetch(`${SERVER_URL}/errors/latest?since=10s`);
            const data = await response.json();
            
            if (data.errors && data.errors.length > 0) {
                data.errors.forEach(error => addStreamError(error));
            }
        } catch (error) {
            console.error('Error polling failed:', error);
        }
        
        setTimeout(pollErrors, 2000);
    };
    
    pollErrors();
}

function handleMonitorCommand(event) {
    if (event.key === 'Enter') {
        const command = event.target.value.trim();
        if (!command) return;
        
        startMonitoringCommand(command);
        event.target.value = '';
    }
}

async function startMonitoringCommand(command) {
    const parts = command.split(' ');
    
    const payload = {
        command: parts[0],
        args: parts.slice(1),
        working_dir: '.',
        auto_restart: true,
        error_patterns: ['Error:', 'error:', 'ERROR', 'Failed', 'Exception']
    };
    
    try {
        const response = await fetch(`${SERVER_URL}/processes/start`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload)
        });
        
        const result = await response.json();
        
        if (response.ok) {
            console.log('Started monitoring:', result);
            setTimeout(loadData, 1000); // Refresh dashboard
        } else {
            console.error('Failed to start monitoring:', result);
        }
    } catch (error) {
        console.error('Error starting monitoring:', error);
    }
}

function startDevServer() {
    const serverType = prompt('Enter server type (npm, go, python):');
    if (!serverType) return;
    
    fetch(`${SERVER_URL}/dev/start/${serverType}`, {
        method: 'POST'
    })
    .then(response => response.json())
    .then(result => {
        console.log('Dev server started:', result);
        setTimeout(loadData, 1000);
    })
    .catch(error => {
        console.error('Error starting dev server:', error);
    });
}

function viewProcessLogs(pid) {
    window.open(`${SERVER_URL}/processes/${pid}/output`, '_blank');
}

// Add to existing updateDashboard function
// Load monitored processes
fetch(`${SERVER_URL}/processes/monitored`)
    .then(response => response.json())
    .then(processes => updateMonitoredProcesses(processes))
    .catch(error => console.error('Error loading monitored processes:', error));
</script>
```

### 7. Usage Examples for Claude Code

Once enhanced, Claude Code can use these new capabilities:

```bash
# Start monitoring a development server
argus monitor "npm run dev"
argus monitor "go run main.go"
argus monitor "python app.py"

# Start predefined dev servers
argus dev start npm
argus dev start go  
argus dev start python

# View real-time errors as they happen
argus stream

# Check what processes are being monitored
argus processes

# View recent output from a specific process
argus logs 1234

# Get latest errors from last 5 minutes
curl http://localhost:3002/errors/latest?since=5m

# Start monitoring via API
curl -X POST http://localhost:3002/processes/start \
  -H "Content-Type: application/json" \
  -d '{"command":"npm","args":["run","dev"],"working_dir":".","auto_restart":true}'
```

## 📋 Implementation Checklist for Claude 4 Sonnet

**Please implement these enhancements by:**

✅ **Adding the new types and structs** to the existing `main.go`
✅ **Implementing the ProcessMonitor methods** 
✅ **Adding WebSocket support** for real-time streaming
✅ **Creating the new API endpoints** and handlers
✅ **Enhancing the CLI tool** with new commands  
✅ **Updating the dashboard** with real-time features
✅ **Adding error pattern matching** for different languages
✅ **Implementing process lifecycle management**

## 🎯 Expected Result

After implementation, Claude Code will have:

- **Real-time error notifications** as they occur in development servers
- **Process monitoring capabilities** for running commands
- **Live error streaming** via WebSocket or polling
- **Automatic dev server integration** for common frameworks
- **Process output inspection** for debugging
- **Enhanced CLI commands** for monitoring and control

This transforms Project Argus from a **periodic monitoring tool** into a **real-time project intelligence system** that immediately notifies Claude Code of any issues as they happen.

## 🛡️ **PRODUCTION-READY GUARDRAILS - MANDATORY REQUIREMENTS**

**Claude 4 Sonnet: You MUST implement ALL these production standards. No shortcuts, placeholders, or simplified implementations allowed.**

### **🔒 Error Handling Requirements**
```go
// REQUIRED: Every function must have comprehensive error handling
func (pm *ProcessMonitor) StartProcess(cmd ProcessCommand) (*MonitoredProcess, error) {
    // ✅ MUST validate all inputs
    if cmd.Command == "" {
        return nil, errors.New("command cannot be empty")
    }
    if cmd.WorkingDir == "" {
        cmd.WorkingDir = "."
    }
    
    // ✅ MUST check if command exists
    if _, err := exec.LookPath(cmd.Command); err != nil {
        return nil, fmt.Errorf("command not found: %s", cmd.Command)
    }
    
    // ✅ MUST validate working directory
    if _, err := os.Stat(cmd.WorkingDir); os.IsNotExist(err) {
        return nil, fmt.Errorf("working directory does not exist: %s", cmd.WorkingDir)
    }
    
    // ✅ MUST implement actual process monitoring with:
    // - Proper process lifecycle management
    // - Resource cleanup on process termination
    // - Goroutine leak prevention
    // - Memory management for output buffers
    // - Timeout handling for unresponsive processes
    
    // ❌ NO PLACEHOLDER CODE LIKE: return &MonitoredProcess{}, nil
}
```

### **🔧 Resource Management Requirements**
```go
// REQUIRED: Proper cleanup and resource management
func (pm *ProcessMonitor) cleanupStoppedProcesses() {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    for pid, process := range pm.activeProcesses {
        // ✅ MUST actually check if process is running
        if !pm.isProcessRunning(pid) {
            // ✅ MUST cleanup all associated resources
            pm.closeProcessChannels(process)
            pm.cleanupProcessFiles(process)
            delete(pm.activeProcesses, pid)
            
            log.Printf("Cleaned up stopped process PID %d", pid)
        }
    }
}

// ✅ MUST implement real process checking, not just return true
func (pm *ProcessMonitor) isProcessRunning(pid int) bool {
    // Actual implementation required - check /proc filesystem or use os.Process
    process, err := os.FindProcess(pid)
    if err != nil {
        return false
    }
    
    // Send signal 0 to check if process exists
    err = process.Signal(syscall.Signal(0))
    return err == nil
}
```

### **🔍 Input Validation & Security**
```go
// REQUIRED: Validate ALL user inputs
func (is *IntelligenceServer) startProcessHandler(c *fiber.Ctx) error {
    var cmd ProcessCommand
    if err := c.BodyParser(&cmd); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid JSON format",
            "details": err.Error(),
        })
    }
    
    // ✅ MUST sanitize and validate command
    if err := validateProcessCommand(cmd); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Invalid command",
            "details": err.Error(),
        })
    }
    
    // ✅ MUST implement rate limiting
    if !is.checkRateLimit(c.IP()) {
        return c.Status(429).JSON(fiber.Map{
            "error": "Rate limit exceeded",
        })
    }
    
    // ✅ MUST prevent command injection
    if containsUnsafeCharacters(cmd.Command) {
        return c.Status(400).JSON(fiber.Map{
            "error": "Command contains unsafe characters",
        })
    }
}

func validateProcessCommand(cmd ProcessCommand) error {
    // ✅ REQUIRED: Real validation logic
    if len(cmd.Command) > 1000 {
        return errors.New("command too long")
    }
    
    // ✅ Whitelist allowed commands for security
    allowedCommands := []string{"npm", "node", "go", "python", "yarn", "cargo"}
    if !contains(allowedCommands, cmd.Command) {
        return fmt.Errorf("command not allowed: %s", cmd.Command)
    }
    
    return nil
}
```

### **⚡ Performance Requirements**
```go
// REQUIRED: Efficient implementations with proper concurrency
func (pm *ProcessMonitor) processErrorStream() {
    // ✅ MUST use buffered channels to prevent blocking
    errorBuffer := make(chan StreamError, 1000)
    
    // ✅ MUST implement worker pool pattern for processing
    for i := 0; i < runtime.NumCPU(); i++ {
        go pm.errorWorker(errorBuffer)
    }
    
    // ✅ MUST implement backpressure handling
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
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
```

### **🗄️ Configuration Management**
```go
// REQUIRED: Configurable, not hardcoded values
type ProcessMonitorConfig struct {
    MaxProcesses        int           `json:"max_processes" default:"10"`
    OutputBufferSize    int           `json:"output_buffer_size" default:"1000"`
    ErrorStreamBuffer   int           `json:"error_stream_buffer" default:"1000"`
    ProcessTimeout      time.Duration `json:"process_timeout" default:"1h"`
    CleanupInterval     time.Duration `json:"cleanup_interval" default:"30s"`
    MaxOutputLines      int           `json:"max_output_lines" default:"10000"`
    AllowedCommands     []string      `json:"allowed_commands"`
    RateLimitPerMinute  int           `json:"rate_limit_per_minute" default:"10"`
}

// ✅ MUST load configuration from file/environment
func loadConfig() *ProcessMonitorConfig {
    config := &ProcessMonitorConfig{}
    
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
    
    return config
}
```

### **🧪 Testing Requirements**
```go
// REQUIRED: Include comprehensive test cases
func TestProcessMonitor_StartProcess(t *testing.T) {
    tests := []struct {
        name        string
        cmd         ProcessCommand
        expectError bool
        errorMsg    string
    }{
        {
            name: "valid command",
            cmd: ProcessCommand{
                Command:    "echo",
                Args:       []string{"hello"},
                WorkingDir: ".",
            },
            expectError: false,
        },
        {
            name: "empty command",
            cmd: ProcessCommand{
                Command: "",
            },
            expectError: true,
            errorMsg:    "command cannot be empty",
        },
        {
            name: "invalid working directory",
            cmd: ProcessCommand{
                Command:    "echo",
                WorkingDir: "/nonexistent",
            },
            expectError: true,
            errorMsg:    "working directory does not exist",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            pm := NewProcessMonitor()
            process, err := pm.StartProcess(tt.cmd)
            
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
                assert.Nil(t, process)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, process)
                assert.Greater(t, process.PID, 0)
            }
        })
    }
}
```

### **📊 Monitoring & Observability**
```go
// REQUIRED: Comprehensive logging and metrics
func (pm *ProcessMonitor) StartProcess(cmd ProcessCommand) (*MonitoredProcess, error) {
    startTime := time.Now()
    
    log.Printf("Starting process monitoring: command=%s, args=%v, workdir=%s", 
        cmd.Command, cmd.Args, cmd.WorkingDir)
    
    defer func() {
        duration := time.Since(startTime)
        log.Printf("Process start completed in %v", duration)
    }()
    
    // ✅ MUST implement metrics collection
    pm.metrics.ProcessStartAttempts.Inc()
    
    // Implementation...
    
    if err != nil {
        pm.metrics.ProcessStartFailures.Inc()
        log.Printf("Failed to start process: %v", err)
        return nil, err
    }
    
    pm.metrics.ActiveProcesses.Inc()
    log.Printf("Successfully started process PID %d", process.PID)
    
    return process, nil
}
```

### **🔒 Security Requirements**
```go
// REQUIRED: Security measures
func (is *IntelligenceServer) setupSecurity() {
    // ✅ MUST implement CORS properly
    is.app.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:*"},
        AllowMethods:     []string{"GET", "POST", "DELETE"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
        AllowCredentials: false,
        MaxAge:           86400,
    }))
    
    // ✅ MUST implement rate limiting
    is.app.Use(limiter.New(limiter.Config{
        Max:        10,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
    }))
    
    // ✅ MUST validate content length
    is.app.Use(func(c *fiber.Ctx) error {
        if c.Request().Header.ContentLength() > 1024*1024 { // 1MB limit
            return c.Status(413).JSON(fiber.Map{
                "error": "Request too large",
            })
        }
        return c.Next()
    })
}
```

### **🚀 Production Deployment Considerations**
```go
// REQUIRED: Production-ready features
func (is *IntelligenceServer) Start(port string) error {
    // ✅ MUST implement graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        log.Println("Gracefully shutting down...")
        
        // Cleanup all monitored processes
        is.pi.processMonitor.StopAllProcesses()
        
        // Close WebSocket connections
        is.closeAllWebSockets()
        
        // Shutdown server
        is.app.Shutdown()
    }()
    
    // ✅ MUST implement health checks
    is.app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status":    "healthy",
            "timestamp": time.Now(),
            "uptime":    time.Since(startTime),
            "processes": len(is.pi.processMonitor.activeProcesses),
        })
    })
    
    log.Printf("Starting Project Argus on port %s", port)
    return is.app.Listen(":" + port)
}
```

## ❌ **FORBIDDEN SHORTCUTS**

**Claude 4 Sonnet: You are PROHIBITED from:**

- ❌ **Placeholder implementations** that return empty/mock data
- ❌ **TODO comments** instead of actual code
- ❌ **Simplified error handling** with just `log.Println(err)`
- ❌ **Missing input validation** or security checks
- ❌ **Hardcoded values** instead of configuration
- ❌ **Memory leaks** from unclosed channels/goroutines
- ❌ **Race conditions** from missing mutex protection
- ❌ **Blocking operations** on main threads
- ❌ **Missing cleanup** for processes/resources
- ❌ **Poor error messages** without context

## ✅ **MANDATORY IMPLEMENTATIONS**

**You MUST provide complete, working code for:**

1. **Real process monitoring** with actual stdout/stderr capture
2. **WebSocket implementation** with proper connection management
3. **Error pattern matching** with regex and language-specific rules
4. **Resource cleanup** with process lifecycle management
5. **Input validation** with security considerations
6. **Configuration management** with file and environment support
7. **Comprehensive logging** with structured output
8. **Rate limiting** and security middleware
9. **Graceful shutdown** with proper cleanup
10. **Error handling** with meaningful messages and recovery

---

**Important**: This enhancement builds on the existing Project Argus codebase. Please read the existing code first, then integrate these new features without breaking the current functionality.

**PRODUCTION REQUIREMENT**: Every line of code must be production-ready, secure, and properly tested. No development shortcuts or placeholder code allowed.