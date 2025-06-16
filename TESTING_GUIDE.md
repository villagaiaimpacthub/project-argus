# Project Argus Enhanced - Complete Testing Guide

## 🚀 Prerequisites Setup

### Install Go (Required)
```bash
# For Ubuntu/WSL
sudo apt update
sudo apt install golang-go

# Verify installation
go version
```

### Install Dependencies
```bash
# Install required tools
sudo apt install curl jq

# Make CLI executable
chmod +x claude-query.sh test_argus.sh
```

## 🎯 Step-by-Step Testing

### 1. Start Project Argus Server
```bash
# Start the enhanced server
go run main.go .

# Expected output:
# Starting Project Argus monitoring for: /path/to/workspace
# Starting Project Argus on port :3002
```

### 2. Basic Health Check
```bash
# In a new terminal, test connectivity
curl http://localhost:3002/

# Should return JSON with service info
```

### 3. Test Original Intelligence Features

```bash
# Service status
./claude-query.sh status

# Project health summary
./claude-query.sh health

# Project structure
./claude-query.sh structure

# Active errors
./claude-query.sh errors

# Git status
./claude-query.sh git

# Recent changes
./claude-query.sh changes
```

### 4. Test NEW Process Monitoring Features

#### Monitor a Command
```bash
# Monitor a simple command
./claude-query.sh monitor "echo 'Hello World'"

# Monitor a script with errors
echo 'echo "Normal output"; echo "Error: Test error" >&2' > test.sh
chmod +x test.sh
./claude-query.sh monitor "./test.sh"
```

#### Check Monitored Processes
```bash
# List all monitored processes
./claude-query.sh monitored

# Get specific process logs (replace PID)
./claude-query.sh logs 1234

# Stop a process (replace PID)
./claude-query.sh stop 1234
```

### 5. Test NEW Development Server Integration

```bash
# Check dev server status
./claude-query.sh dev status

# Start development server (if package.json exists)
./claude-query.sh dev start npm

# Start other server types
./claude-query.sh dev start go      # For Go projects
./claude-query.sh dev start python  # For Python projects

# Stop servers
./claude-query.sh dev stop npm
```

### 6. Test NEW Real-Time Error Streaming

```bash
# Start real-time error monitoring
./claude-query.sh stream

# This will continuously poll for errors
# Press Ctrl+C to stop
```

### 7. Test NEW API Endpoints Directly

#### Process Management API
```bash
# Start a process via API
curl -X POST http://localhost:3002/processes/start \
  -H "Content-Type: application/json" \
  -d '{
    "command": "echo",
    "args": ["Hello", "API"],
    "working_dir": ".",
    "auto_restart": false
  }'

# Get monitored processes
curl http://localhost:3002/processes/monitored

# Get process output (replace PID)
curl http://localhost:3002/processes/1234/output

# Stop process (replace PID)
curl -X DELETE http://localhost:3002/processes/1234
```

#### Error Streaming API
```bash
# Get latest errors
curl http://localhost:3002/errors/latest?since=60s

# Stream errors via HTTP
curl http://localhost:3002/errors/stream
```

### 8. Test NEW WebSocket Endpoints

#### Using websocat (if available)
```bash
# Install websocat
cargo install websocat

# Connect to error stream
websocat ws://localhost:3002/ws/errors

# Connect to process stream
websocat ws://localhost:3002/ws/processes
```

#### Using JavaScript in browser
```javascript
// Open browser console and run:
const ws = new WebSocket('ws://localhost:3002/ws/errors');
ws.onmessage = (event) => console.log('Error:', JSON.parse(event.data));

// For process monitoring:
const wsProc = new WebSocket('ws://localhost:3002/ws/processes');
wsProc.onmessage = (event) => console.log('Process:', JSON.parse(event.data));
```

## 🧪 Comprehensive Test Suite

Run the automated test suite:
```bash
# Execute all tests
./test_argus.sh

# This will test:
# ✅ Server connectivity
# ✅ Basic intelligence features
# ✅ Process monitoring
# ✅ Development server integration
# ✅ Error streaming
# ✅ API endpoints
# ✅ File operations
# ✅ WebSocket connectivity
```

## 🎪 Demo Scenarios

### Scenario 1: Monitor a Failing Node.js App
```bash
# Create a failing Node.js script
echo 'console.log("Starting app...");
const x = y; // ReferenceError
console.error("App crashed!");
process.exit(1);' > failing-app.js

# Monitor it
./claude-query.sh monitor "node failing-app.js"

# Check error stream
./claude-query.sh stream
```

### Scenario 2: Development Workflow
```bash
# Start a development server
./claude-query.sh dev start npm

# Monitor the server in real-time
./claude-query.sh stream

# Check server status
./claude-query.sh dev status

# Stop when done
./claude-query.sh dev stop npm
```

### Scenario 3: Multi-Process Monitoring
```bash
# Monitor multiple processes
./claude-query.sh monitor "ping google.com"
./claude-query.sh monitor "echo 'Process 2'"
./claude-query.sh monitor "sleep 30"

# Check all monitored processes
./claude-query.sh monitored

# View logs from specific processes
./claude-query.sh logs <PID1>
./claude-query.sh logs <PID2>
```

## 🔍 Expected Results

### ✅ Success Indicators
- Server starts on port 3002
- All CLI commands return JSON responses
- Process monitoring captures stdout/stderr
- Error patterns are detected correctly
- WebSocket connections accept upgrades
- Development servers start/stop correctly
- Real-time streaming shows live updates

### ❌ Common Issues
- **Port 3002 in use**: Change port with `CLAUDE_INTEL_PORT=3003 go run main.go .`
- **Permission denied**: Run `chmod +x claude-query.sh test_argus.sh`
- **jq not found**: Install with `sudo apt install jq`
- **Go not found**: Install with `sudo apt install golang-go`

## 🚀 Performance Testing

### Load Testing
```bash
# Test multiple simultaneous processes
for i in {1..5}; do
  ./claude-query.sh monitor "echo 'Process $i'" &
done

# Check system can handle load
./claude-query.sh monitored
```

### Memory Testing
```bash
# Monitor a memory-intensive process
./claude-query.sh monitor "yes > /dev/null"

# Check resource usage
./claude-query.sh monitored
ps aux | grep argus
```

## 🎯 Feature Validation Checklist

- [ ] ✨ Real-time process monitoring
- [ ] ⚡ Live error detection and streaming
- [ ] 🔧 Development server integration
- [ ] 🌐 WebSocket real-time updates
- [ ] 📡 REST API process management
- [ ] 🔍 Enhanced file and project intelligence
- [ ] 🛡️ Security (rate limiting, command validation)
- [ ] 🧹 Resource cleanup and management
- [ ] 📊 Process metrics and monitoring
- [ ] 🔄 Auto-restart capabilities

## 🎉 Success Criteria

**Project Argus Enhanced should provide:**
1. **All-seeing monitoring** - Captures everything happening in development
2. **Real-time intelligence** - Immediate error notifications
3. **Development integration** - One-command server management
4. **Production-ready** - Secure, performant, and reliable
5. **Claude-friendly** - Perfect integration for AI-assisted development

---

**🚀 Project Argus: All-seeing project monitoring for Claude Code!** 