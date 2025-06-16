# 🚀 Claude Code Self-Monitoring Demo Guide

## Overview

This guide shows you how to set up **Claude Code to monitor its own codebase** in real-time, demonstrating the WebSocket functionality and production capabilities of Project Argus Enhanced.

**What you'll see:**
- Real-time WebSocket streaming of errors and process updates
- Live file monitoring and change detection
- Git status tracking
- Process monitoring with live output capture
- Error pattern detection across multiple languages
- Development server integration

---

## 🎯 Quick Start

### Step 1: Open a New WSL Terminal

Open a new Windows Subsystem for Linux terminal and navigate to your project:

```bash
cd "/mnt/c/Go Fiber Router Backend"
```

### Step 2: Start Claude Code Self-Monitoring

```bash
# Make scripts executable and start the demo
chmod +x test_self_monitoring.sh generate_activity.sh
./test_self_monitoring.sh
```

This will:
- ✅ Build and start the Claude Code intelligence service
- ✅ Configure it to monitor its own codebase
- ✅ Set up WebSocket endpoints for real-time streaming
- ✅ Display connection information and instructions

### Step 3: Open WebSocket Test Client

1. Open `websocket_test.html` in your web browser
2. The page will automatically test production readiness
3. Click **"Connect Error Stream"** and **"Connect Process Stream"**
4. You should see successful WebSocket connections

### Step 4: Generate Activity to Monitor

In another WSL terminal:

```bash
cd "/mnt/c/Go Fiber Router Backend"
./generate_activity.sh
```

Choose option **6 (Run all activities)** to see comprehensive monitoring in action!

---

## 🌐 What You'll See in Real-Time

### WebSocket Error Stream (`/ws/errors`)
```json
{
  "type": "error",
  "process_pid": 12345,
  "command": "go build test_error.go",
  "error_type": "compilation",
  "message": "syntax error: missing ')'",
  "timestamp": "2024-01-15T10:30:00Z",
  "severity": "error",
  "source": "stderr",
  "context": ["fmt.Println(\"Hello, World!\"", "// Missing closing parenthesis"]
}
```

### WebSocket Process Stream (`/ws/processes`)
```json
{
  "type": "process_update",
  "processes": [
    {
      "pid": 12345,
      "command": "go",
      "args": ["build", "test_error.go"],
      "status": "running",
      "start_time": "2024-01-15T10:30:00Z",
      "output_lines": ["# command-line-arguments", "./test_error.go:6:30: syntax error"]
    }
  ],
  "timestamp": "2024-01-15T10:30:05Z"
}
```

---

## 🎬 Demo Scenarios

### Scenario 1: File Creation and Error Detection

```bash
# In terminal 2 (activity generator)
./generate_activity.sh files
```

**What Claude Code detects:**
- ✅ New files created (Go, JavaScript, Python, TypeScript)
- ✅ Syntax errors in Go files
- ✅ Runtime errors in JavaScript/Python
- ✅ Type errors in TypeScript
- ✅ File structure changes

### Scenario 2: Development Workflow Simulation

```bash
# In terminal 2
./generate_activity.sh workflow
```

**What Claude Code monitors:**
- ✅ File modifications in real-time
- ✅ Git operations (add, status)
- ✅ Build processes and their output
- ✅ Compilation errors and warnings
- ✅ Runtime error detection

### Scenario 3: Continuous Activity

```bash
# In terminal 2
./generate_activity.sh continuous
```

**Live monitoring features:**
- ✅ Continuous file changes
- ✅ Periodic build attempts
- ✅ Real-time WebSocket broadcasts
- ✅ Process lifecycle management
- ✅ Automatic cleanup

### Scenario 4: API Testing

```bash
# In terminal 2
./generate_activity.sh api
```

**API interactions you'll see:**
- ✅ Process start/stop operations
- ✅ Error stream queries
- ✅ Project structure analysis
- ✅ Health monitoring
- ✅ Real-time metrics

---

## 🔧 Available Endpoints

### WebSocket Endpoints (Real-time)
| Endpoint | Purpose | Test URL |
|----------|---------|----------|
| `ws://localhost:3002/ws/errors` | Real-time error stream | Use WebSocket client |
| `ws://localhost:3002/ws/processes` | Process monitoring | Use WebSocket client |

### REST API Endpoints
| Endpoint | Purpose | Test Command |
|----------|---------|--------------|
| `GET /` | Server status | `curl http://localhost:3002/` |
| `GET /structure` | Project structure | `curl http://localhost:3002/structure` |
| `GET /health` | Project health | `curl http://localhost:3002/health` |
| `GET /errors` | Current errors | `curl http://localhost:3002/errors` |
| `GET /processes/monitored` | Active processes | `curl http://localhost:3002/processes/monitored` |
| `POST /processes/start` | Start process | See activity generator |
| `GET /errors/latest` | Recent errors | `curl http://localhost:3002/errors/latest` |

---

## 📊 Monitoring Capabilities

### Real-Time File Watching
- **File changes** detected instantly
- **New file creation** monitored
- **File deletion** tracked
- **Directory changes** observed

### Process Monitoring
- **Live stdout/stderr** capture
- **Process lifecycle** management
- **Exit code** tracking
- **Resource usage** monitoring

### Error Detection
- **Go compilation** errors
- **JavaScript runtime** errors
- **Python exceptions**
- **TypeScript type** errors
- **Build failures**

### Git Integration
- **Status changes** monitored
- **Commit detection**
- **Branch changes**
- **Staged files** tracking

---

## 🎮 Interactive Testing

### Using the WebSocket HTML Client

1. **Open `websocket_test.html`** in your browser
2. **Connect to both streams:**
   - Click "Connect Error Stream"
   - Click "Connect Process Stream"
3. **Generate activity:**
   - Click "Start Test Process"
   - Run `./generate_activity.sh` in terminal
4. **Watch real-time updates** in the message log

### Command Line Testing

```bash
# Test individual API endpoints
curl -s http://localhost:3002/health | jq
curl -s http://localhost:3002/structure | jq '.files | length'
curl -s http://localhost:3002/processes/monitored | jq

# Start a monitored process
curl -X POST -H "Content-Type: application/json" \
  -d '{"command":"go","args":["version"],"working_dir":"."}' \
  http://localhost:3002/processes/start

# Monitor server logs
tail -f claude-code.log
```

---

## 🛑 Stopping the Demo

### Stop the Server
```bash
# Kill the server using PID file
kill $(cat claude-code.pid 2>/dev/null)

# Or find and kill the process
pkill -f "claude-code"
```

### Cleanup Test Files
```bash
./generate_activity.sh cleanup
rm -f claude-code.log claude-code.pid claude-code
```

---

## 🚀 Expected Results

### ✅ Production Readiness Confirmed
- **Real WebSocket** implementation working
- **Live error detection** across multiple languages
- **Process monitoring** with real-time output
- **File system monitoring** operational
- **Git integration** functional
- **API endpoints** responding correctly

### ✅ Performance Metrics
- **WebSocket connections** stable and responsive
- **Real-time broadcasting** working
- **Error pattern matching** accurate
- **Resource management** efficient
- **Memory usage** controlled

### ✅ Security Features
- **Rate limiting** active
- **Command validation** working
- **Input sanitization** operational
- **Process isolation** maintained

---

## 🎉 Success Indicators

When everything is working correctly, you'll see:

1. **Server startup** with WebSocket endpoints logged
2. **WebSocket connections** established in HTML client
3. **Real-time messages** flowing in both streams
4. **File changes** detected instantly
5. **Process monitoring** capturing live output
6. **Error detection** working across languages
7. **API responses** returning valid JSON

**🎯 This demonstrates that Project Argus Enhanced is production-ready with real WebSocket functionality!**

---

## 💡 Pro Tips

- **Multiple browsers** can connect simultaneously to test scaling
- **Network tab** in browser dev tools shows WebSocket traffic
- **Server logs** (`claude-code.log`) show detailed internal operations
- **Activity generator** can run continuously for stress testing
- **WebSocket test client** provides production-ready monitoring interface

**🚀 Enjoy watching Claude Code monitor itself in real-time!** 