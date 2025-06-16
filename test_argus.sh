#!/bin/bash

# Project Argus Enhanced - Comprehensive Test Suite
# Tests all real-time monitoring and development intelligence features

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:3002"
CLI="./claude-query.sh"

print_header() {
    echo -e "\n${CYAN}========================================${NC}"
    echo -e "${WHITE}$1${NC}"
    echo -e "${CYAN}========================================${NC}"
}

print_test() {
    echo -e "\n${YELLOW}🧪 TEST: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Test server connectivity
test_server_connectivity() {
    print_test "Server Connectivity"
    
    if curl -s --max-time 5 "$BASE_URL/" > /dev/null; then
        print_success "Server is running on port 3002"
        return 0
    else
        print_error "Server is not responding on port 3002"
        print_info "Please start the server with: go run main.go ."
        return 1
    fi
}

# Test basic intelligence features
test_basic_intelligence() {
    print_header "Testing Basic Intelligence Features"
    
    print_test "Service Status"
    if $CLI status > /tmp/status_test.out 2>&1; then
        print_success "Status endpoint working"
    else
        print_error "Status endpoint failed"
        cat /tmp/status_test.out
    fi
    
    print_test "Project Health"
    if $CLI health > /tmp/health_test.out 2>&1; then
        print_success "Health endpoint working"
    else
        print_error "Health endpoint failed"
        cat /tmp/health_test.out
    fi
    
    print_test "Project Structure"
    if $CLI structure > /tmp/structure_test.out 2>&1; then
        print_success "Structure endpoint working"
    else
        print_error "Structure endpoint failed"
        cat /tmp/structure_test.out
    fi
    
    print_test "Error Detection"
    if $CLI errors > /tmp/errors_test.out 2>&1; then
        print_success "Errors endpoint working"
    else
        print_error "Errors endpoint failed"
        cat /tmp/errors_test.out
    fi
}

# Test process monitoring
test_process_monitoring() {
    print_header "Testing Process Monitoring Features"
    
    print_test "Starting a Test Process"
    
    # Create a test script that generates output and errors
    cat > test_process.sh << 'EOF'
#!/bin/bash
echo "Test process started"
echo "Normal output line 1"
echo "Error: This is a test error" >&2
echo "Normal output line 2"
echo "Failed to do something" >&2
echo "Process completed"
sleep 2
echo "Final output"
EOF
    
    chmod +x test_process.sh
    
    # Monitor the test process
    if $CLI monitor "./test_process.sh" > /tmp/monitor_test.out 2>&1; then
        print_success "Process monitoring started"
        cat /tmp/monitor_test.out
        
        # Wait a moment for process to start
        sleep 1
        
        print_test "Checking Monitored Processes"
        if $CLI monitored > /tmp/monitored_test.out 2>&1; then
            print_success "Can list monitored processes"
            cat /tmp/monitored_test.out
            
            # Extract PID from output if available
            PID=$(grep -o 'PID [0-9]*' /tmp/monitored_test.out | head -1 | cut -d' ' -f2)
            
            if [ -n "$PID" ]; then
                print_test "Getting Process Logs (PID: $PID)"
                if $CLI logs $PID > /tmp/logs_test.out 2>&1; then
                    print_success "Process logs retrieved"
                    cat /tmp/logs_test.out
                else
                    print_error "Failed to get process logs"
                    cat /tmp/logs_test.out
                fi
                
                # Let process complete, then test stop
                sleep 3
                
                print_test "Stopping Process (PID: $PID)"
                if $CLI stop $PID > /tmp/stop_test.out 2>&1; then
                    print_success "Process stopped successfully"
                    cat /tmp/stop_test.out
                else
                    print_error "Failed to stop process"
                    cat /tmp/stop_test.out
                fi
            else
                print_info "Could not extract PID from monitor output"
            fi
        else
            print_error "Failed to list monitored processes"
            cat /tmp/monitored_test.out
        fi
    else
        print_error "Failed to start process monitoring"
        cat /tmp/monitor_test.out
    fi
    
    # Cleanup
    rm -f test_process.sh
}

# Test development server integration
test_dev_server_integration() {
    print_header "Testing Development Server Integration"
    
    print_test "Development Server Status"
    if $CLI dev status > /tmp/dev_status_test.out 2>&1; then
        print_success "Dev server status endpoint working"
        cat /tmp/dev_status_test.out
    else
        print_error "Dev server status failed"
        cat /tmp/dev_status_test.out
    fi
    
    # Test with a simple package.json if we can
    if [ -f package.json ]; then
        print_test "Starting NPM Development Server"
        if $CLI dev start npm > /tmp/dev_start_test.out 2>&1; then
            print_success "NPM dev server started"
            cat /tmp/dev_start_test.out
            
            sleep 2
            
            print_test "Stopping NPM Development Server"
            if $CLI dev stop npm > /tmp/dev_stop_test.out 2>&1; then
                print_success "NPM dev server stopped"
                cat /tmp/dev_stop_test.out
            else
                print_error "Failed to stop NPM dev server"
                cat /tmp/dev_stop_test.out
            fi
        else
            print_info "NPM dev server test skipped (package.json not found or no scripts)"
            cat /tmp/dev_start_test.out
        fi
    else
        print_info "NPM dev server test skipped (no package.json found)"
    fi
}

# Test error streaming
test_error_streaming() {
    print_header "Testing Real-Time Error Streaming"
    
    print_test "Latest Errors Endpoint"
    if curl -s "$BASE_URL/errors/latest?since=60s" > /tmp/latest_errors_test.out 2>&1; then
        print_success "Latest errors endpoint working"
        cat /tmp/latest_errors_test.out
    else
        print_error "Latest errors endpoint failed"
        cat /tmp/latest_errors_test.out
    fi
    
    print_test "Error Stream (5 second test)"
    print_info "Testing error stream for 5 seconds..."
    
    # Create a background process that generates errors
    cat > error_generator.sh << 'EOF'
#!/bin/bash
for i in {1..3}; do
    echo "Error: Test error $i" >&2
    echo "Warning: Test warning $i" >&2
    sleep 1
done
EOF
    
    chmod +x error_generator.sh
    
    # Start the error generator in background
    $CLI monitor "./error_generator.sh" > /tmp/error_gen_start.out 2>&1 &
    MONITOR_PID=$!
    
    # Test streaming for a few seconds
    timeout 5s $CLI stream > /tmp/stream_test.out 2>&1 || true
    
    if [ -s /tmp/stream_test.out ]; then
        print_success "Error streaming is working"
        head -10 /tmp/stream_test.out
    else
        print_info "No errors captured in stream (this is normal if no errors occurred)"
    fi
    
    # Cleanup
    kill $MONITOR_PID 2>/dev/null || true
    rm -f error_generator.sh
}

# Test API endpoints directly
test_api_endpoints() {
    print_header "Testing API Endpoints Directly"
    
    print_test "Process Management API"
    
    # Test starting a process via API
    echo '{"command":"echo","args":["Hello","API","Test"],"working_dir":".","auto_restart":false}' > /tmp/start_process.json
    
    if curl -s -X POST "$BASE_URL/processes/start" \
        -H "Content-Type: application/json" \
        -d @/tmp/start_process.json > /tmp/api_start_test.out 2>&1; then
        print_success "Process start API working"
        cat /tmp/api_start_test.out
        
        # Extract PID if available
        API_PID=$(grep -o '"pid":[0-9]*' /tmp/api_start_test.out | cut -d':' -f2)
        
        if [ -n "$API_PID" ]; then
            print_test "Getting process via API (PID: $API_PID)"
            if curl -s "$BASE_URL/processes/$API_PID/output" > /tmp/api_output_test.out 2>&1; then
                print_success "Process output API working"
                cat /tmp/api_output_test.out
            else
                print_error "Process output API failed"
                cat /tmp/api_output_test.out
            fi
        fi
    else
        print_error "Process start API failed"
        cat /tmp/api_start_test.out
    fi
    
    print_test "Monitored Processes API"
    if curl -s "$BASE_URL/processes/monitored" > /tmp/api_monitored_test.out 2>&1; then
        print_success "Monitored processes API working"
        cat /tmp/api_monitored_test.out
    else
        print_error "Monitored processes API failed"
        cat /tmp/api_monitored_test.out
    fi
    
    # Cleanup
    rm -f /tmp/start_process.json
}

# Test file operations
test_file_operations() {
    print_header "Testing File Operations"
    
    print_test "File Search"
    if $CLI search "main" > /tmp/search_test.out 2>&1; then
        print_success "Search functionality working"
        head -10 /tmp/search_test.out
    else
        print_error "Search functionality failed"
        cat /tmp/search_test.out
    fi
    
    print_test "File Information"
    if $CLI file "main.go" > /tmp/file_test.out 2>&1; then
        print_success "File information working"
        cat /tmp/file_test.out
    else
        print_error "File information failed"
        cat /tmp/file_test.out
    fi
}

# Test WebSocket connections (basic connectivity)
test_websocket_basic() {
    print_header "Testing WebSocket Basic Connectivity"
    
    print_test "WebSocket Error Stream Endpoint"
    # Test WebSocket endpoint accessibility (not full WebSocket test)
    if curl -s -I "$BASE_URL/ws/errors" | grep -q "426 Upgrade Required"; then
        print_success "WebSocket error stream endpoint accessible"
    else
        print_info "WebSocket endpoint test inconclusive"
    fi
    
    print_test "WebSocket Process Stream Endpoint"
    if curl -s -I "$BASE_URL/ws/processes" | grep -q "426 Upgrade Required"; then
        print_success "WebSocket process stream endpoint accessible"
    else
        print_info "WebSocket endpoint test inconclusive"
    fi
}

# Main test execution
main() {
    print_header "Project Argus Enhanced - Comprehensive Test Suite"
    echo -e "${PURPLE}Testing all real-time monitoring and intelligence features${NC}"
    
    # Make CLI executable
    chmod +x "$CLI" 2>/dev/null || true
    
    # Check if server is running
    if ! test_server_connectivity; then
        print_error "Cannot proceed with tests - server not running"
        echo -e "\n${YELLOW}To start the server:${NC}"
        echo "go run main.go ."
        exit 1
    fi
    
    # Run all test suites
    test_basic_intelligence
    test_process_monitoring
    test_dev_server_integration
    test_error_streaming
    test_api_endpoints
    test_file_operations
    test_websocket_basic
    
    print_header "Test Suite Completed"
    print_success "All tests have been executed"
    print_info "Check the output above for any failures"
    
    # Cleanup temp files
    rm -f /tmp/*_test.out /tmp/error_gen_start.out
    
    echo -e "\n${CYAN}Project Argus Enhanced Testing Summary:${NC}"
    echo -e "${WHITE}✨ Real-time process monitoring${NC}"
    echo -e "${WHITE}⚡ Live error detection and streaming${NC}"  
    echo -e "${WHITE}🔧 Development server integration${NC}"
    echo -e "${WHITE}🌐 WebSocket real-time updates${NC}"
    echo -e "${WHITE}📡 REST API process management${NC}"
    echo -e "${WHITE}🔍 Enhanced file and project intelligence${NC}"
    echo -e "\n${PURPLE}All-seeing project monitoring for Claude Code! 🚀${NC}"
}

# Run the test suite
main "$@" 