#!/bin/bash

# Claude Code Self-Monitoring Demo
# Test Project Argus Enhanced by having it monitor its own codebase

echo "🚀 Claude Code Self-Monitoring Demo"
echo "======================================"
echo "Testing Project Argus Enhanced by monitoring its own codebase!"
echo

# Configuration
PORT="3002"
WORKSPACE_PATH="/mnt/c/Go Fiber Router Backend"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')] ✅${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%H:%M:%S')] ⚠️${NC} $1"
}

log_error() {
    echo -e "${RED}[$(date +'%H:%M:%S')] ❌${NC} $1"
}

# Step 1: Check if server is running
check_server() {
    log "Checking if Claude Code intelligence service is running..."
    
    if timeout 3 bash -c "</dev/tcp/localhost/$PORT" 2>/dev/null; then
        log_success "Claude Code server is running on port $PORT"
        return 0
    else
        log_warning "Claude Code server not detected on port $PORT"
        return 1
    fi
}

# Step 2: Start the server if needed
start_server() {
    log "Starting Claude Code intelligence service..."
    log "Workspace: $WORKSPACE_PATH"
    
    # Check if main.go exists
    if [ ! -f "main.go" ]; then
        log_error "main.go not found in current directory"
        exit 1
    fi
    
    # Build and run
    log "Building Go application..."
    if go build -o claude-code main.go; then
        log_success "Build successful"
        
        log "Starting Claude Code intelligence service in background..."
        nohup ./claude-code > claude-code.log 2>&1 &
        SERVER_PID=$!
        echo $SERVER_PID > claude-code.pid
        
        log "Server PID: $SERVER_PID"
        log "Waiting for server to start..."
        
        # Wait for server to be ready
        for i in {1..10}; do
            if timeout 1 bash -c "</dev/tcp/localhost/$PORT" 2>/dev/null; then
                log_success "Claude Code server is ready!"
                return 0
            fi
            sleep 1
        done
        
        log_error "Server failed to start within 10 seconds"
        return 1
    else
        log_error "Build failed"
        return 1
    fi
}

# Step 3: Test basic endpoints
test_endpoints() {
    log "Testing Claude Code API endpoints..."
    
    # Test basic intelligence
    if command -v curl > /dev/null 2>&1; then
        log "Testing server status..."
        if curl -s "http://localhost:$PORT/" > /dev/null; then
            log_success "Server status endpoint working"
        fi
        
        log "Testing project structure analysis..."
        if curl -s "http://localhost:$PORT/structure" > /dev/null; then
            log_success "Project structure endpoint working"
        fi
        
        log "Testing WebSocket endpoints (should return 426)..."
        ws_status=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT/ws/errors")
        if [ "$ws_status" = "426" ]; then
            log_success "WebSocket error stream endpoint ready (426 Upgrade Required)"
        fi
        
        ws_proc_status=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$PORT/ws/processes")
        if [ "$ws_proc_status" = "426" ]; then
            log_success "WebSocket process stream endpoint ready (426 Upgrade Required)"
        fi
    fi
}

# Step 4: Instructions for WebSocket testing
show_instructions() {
    echo
    echo "=============================================="
    echo "🎯 Claude Code Self-Monitoring Instructions"
    echo "=============================================="
    echo
    log_success "Claude Code is now monitoring its own codebase!"
    echo
    echo "📡 WebSocket Endpoints Available:"
    echo "  • ws://localhost:$PORT/ws/errors     - Real-time error stream"
    echo "  • ws://localhost:$PORT/ws/processes  - Real-time process monitoring"
    echo
    echo "🌐 Test with the WebSocket HTML client:"
    echo "  1. Open: websocket_test.html in your browser"
    echo "  2. Click 'Connect Error Stream' and 'Connect Process Stream'"
    echo "  3. Click 'Start Test Process' to generate activity"
    echo
    echo "🔧 REST API Endpoints:"
    echo "  • http://localhost:$PORT/             - Server status"
    echo "  • http://localhost:$PORT/structure    - Project structure"
    echo "  • http://localhost:$PORT/health       - Project health"
    echo "  • http://localhost:$PORT/errors       - Current errors"
    echo "  • http://localhost:$PORT/processes/monitored - Monitored processes"
    echo
    echo "🧪 Generate some activity to test monitoring:"
    echo "  # In another terminal:"
    echo "  cd '$WORKSPACE_PATH'"
    echo "  echo 'package main' > test_file.go"
    echo "  go build main.go"
    echo "  git status"
    echo
    echo "🎮 Watch real-time monitoring in action:"
    echo "  • File changes will be detected"
    echo "  • Build processes will be monitored"
    echo "  • Git status changes will be tracked"
    echo "  • WebSocket clients will receive live updates"
    echo
    echo "🛑 To stop the server:"
    echo "  kill \$(cat claude-code.pid 2>/dev/null)"
    echo
    echo "=============================================="
}

# Step 5: Monitor the logs
monitor_logs() {
    if [ -f "claude-code.log" ]; then
        echo
        log "Showing recent server activity..."
        echo "=================================="
        tail -n 20 claude-code.log
        echo "=================================="
        echo
        log "To monitor live logs: tail -f claude-code.log"
    fi
}

# Main execution
main() {
    echo
    log "Starting Claude Code Self-Monitoring Demo..."
    
    # Check current directory
    if [ ! -f "main.go" ]; then
        log_error "Please run this script from the Claude Code project directory"
        log_error "Expected to find main.go in current directory"
        exit 1
    fi
    
    # Check if server is already running
    if check_server; then
        log_success "Using existing Claude Code server"
    else
        log "Starting new Claude Code server..."
        if ! start_server; then
            log_error "Failed to start Claude Code server"
            exit 1
        fi
    fi
    
    # Test endpoints
    test_endpoints
    
    # Show instructions
    show_instructions
    
    # Monitor initial logs
    monitor_logs
    
    echo
    log_success "🎉 Claude Code Self-Monitoring Demo is ready!"
    log "Open websocket_test.html in your browser to see real-time monitoring"
    echo
}

# Execute
main "$@" 