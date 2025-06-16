#!/bin/bash

# Generate Development Activity for Claude Code Monitoring
# Creates various activities to demonstrate real-time monitoring capabilities

echo "🎬 Generating Development Activity for Claude Code Monitoring"
echo "============================================================="

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

log_activity() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')] 🎬${NC} $1"
}

# Function to create test files with different content
create_test_files() {
    log_activity "Creating test files for monitoring..."
    
    # Go file with syntax error
    log "Creating Go file with syntax error..."
    cat > test_error.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!"
    // Missing closing parenthesis - syntax error
}
EOF

    # JavaScript file with error
    log "Creating JavaScript file with error..."
    cat > test_error.js << 'EOF'
console.log("Starting JavaScript test");

function testFunction() {
    let x = 5;
    let y = undefined;
    console.log(x + y.length); // This will cause a runtime error
}

testFunction();
EOF

    # Python file with error
    log "Creating Python file with error..."
    cat > test_error.py << 'EOF'
print("Starting Python test")

def test_function():
    x = [1, 2, 3]
    print(x[10])  # IndexError

test_function()
EOF

    # TypeScript file with error
    log "Creating TypeScript file with error..."
    cat > test_error.ts << 'EOF'
interface User {
    name: string;
    age: number;
}

const user: User = {
    name: "John",
    // Missing age property - type error
};

console.log(user.email); // Property doesn't exist
EOF

    # Valid Go file
    log "Creating valid Go file..."
    cat > test_valid.go << 'EOF'
package main

import (
    "fmt"
    "time"
)

func main() {
    fmt.Println("This is a valid Go program")
    fmt.Printf("Current time: %v\n", time.Now())
}
EOF

    log_activity "Test files created successfully"
}

# Function to simulate development workflow
simulate_development() {
    log_activity "Simulating development workflow..."
    
    # 1. File modifications
    log "Modifying existing files..."
    echo "// Modified at $(date)" >> test_valid.go
    sleep 2
    
    # 2. Git operations
    if command -v git > /dev/null 2>&1; then
        log "Performing git operations..."
        git add test_*.go test_*.js 2>/dev/null || true
        git status
        sleep 2
    fi
    
    # 3. Build attempts (will generate errors)
    log "Attempting to build files with errors..."
    
    echo "Building Go file with syntax error..."
    go build test_error.go 2>&1 || true
    sleep 2
    
    echo "Building valid Go file..."
    go build test_valid.go 2>&1 || true
    sleep 2
    
    # 4. Run JavaScript with Node.js (if available)
    if command -v node > /dev/null 2>&1; then
        log "Running JavaScript file with error..."
        node test_error.js 2>&1 || true
        sleep 2
    fi
    
    # 5. Run Python file (if available)
    if command -v python3 > /dev/null 2>&1; then
        log "Running Python file with error..."
        python3 test_error.py 2>&1 || true
        sleep 2
    elif command -v python > /dev/null 2>&1; then
        log "Running Python file with error..."
        python test_error.py 2>&1 || true
        sleep 2
    fi
    
    log_activity "Development workflow simulation complete"
}

# Function to generate continuous activity
generate_continuous_activity() {
    log_activity "Starting continuous activity generation..."
    log "Press Ctrl+C to stop"
    
    local counter=1
    while true; do
        echo
        log "Activity round $counter"
        
        # Modify a file
        echo "// Update $counter at $(date)" >> test_valid.go
        sleep 3
        
        # Attempt builds
        if [ $((counter % 3)) -eq 0 ]; then
            log "Building with errors (round $counter)..."
            go build test_error.go 2>&1 || true
        else
            log "Building valid code (round $counter)..."
            go build test_valid.go 2>&1 || true
        fi
        
        sleep 5
        
        # Create/delete temporary files
        temp_file="temp_$counter.txt"
        echo "Temporary content $counter" > $temp_file
        sleep 2
        rm -f $temp_file
        
        counter=$((counter + 1))
        sleep 5
    done
}

# Function to test WebSocket endpoints via curl
test_websockets() {
    log_activity "Testing WebSocket endpoints..."
    
    # Test WebSocket upgrade requests
    log "Testing WebSocket error stream endpoint..."
    curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" http://localhost:3002/ws/errors 2>/dev/null || true
    
    sleep 2
    
    log "Testing WebSocket process stream endpoint..."
    curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" http://localhost:3002/ws/processes 2>/dev/null || true
}

# Function to make API calls to generate activity
test_api_endpoints() {
    log_activity "Testing API endpoints to generate activity..."
    
    if command -v curl > /dev/null 2>&1; then
        log "Testing various API endpoints..."
        
        # Basic endpoints
        curl -s http://localhost:3002/ | head -5
        echo "..."
        sleep 1
        
        curl -s http://localhost:3002/structure | head -10
        echo "..."
        sleep 1
        
        curl -s http://localhost:3002/health
        echo
        sleep 1
        
        curl -s http://localhost:3002/errors
        echo
        sleep 1
        
        # Process monitoring endpoints
        curl -s http://localhost:3002/processes/monitored
        echo
        sleep 1
        
        # Try to start a monitored process
        log "Starting a monitored process..."
        curl -X POST -H "Content-Type: application/json" -d '{
            "command": "go",
            "args": ["version"],
            "working_dir": ".",
            "auto_restart": false
        }' http://localhost:3002/processes/start
        echo
        sleep 2
        
        log_activity "API endpoint testing complete"
    else
        log "curl not available, skipping API tests"
    fi
}

# Cleanup function
cleanup() {
    log_activity "Cleaning up test files..."
    rm -f test_*.go test_*.js test_*.py test_*.ts temp_*.txt 2>/dev/null || true
    rm -f test_valid test_error 2>/dev/null || true
    log_activity "Cleanup complete"
}

# Main menu
show_menu() {
    echo
    echo "=================================================="
    echo "🎬 Claude Code Activity Generator"
    echo "=================================================="
    echo "Choose an activity to demonstrate monitoring:"
    echo
    echo "1. Create test files with errors"
    echo "2. Simulate development workflow"
    echo "3. Generate continuous activity"
    echo "4. Test WebSocket endpoints"
    echo "5. Test API endpoints"
    echo "6. Run all activities"
    echo "7. Cleanup test files"
    echo "8. Exit"
    echo
    echo -n "Enter your choice (1-8): "
}

# Main execution
main() {
    # Check if Claude Code server is running
    if ! timeout 3 bash -c "</dev/tcp/localhost/3002" 2>/dev/null; then
        echo
        log "⚠️ Claude Code server not detected on port 3002"
        log "Please run: ./test_self_monitoring.sh first"
        echo
        exit 1
    fi
    
    log_activity "Claude Code server detected - ready to generate activity!"
    
    if [ $# -eq 0 ]; then
        # Interactive mode
        while true; do
            show_menu
            read -r choice
            
            case $choice in
                1)
                    create_test_files
                    ;;
                2)
                    simulate_development
                    ;;
                3)
                    generate_continuous_activity
                    ;;
                4)
                    test_websockets
                    ;;
                5)
                    test_api_endpoints
                    ;;
                6)
                    create_test_files
                    sleep 2
                    simulate_development
                    sleep 2
                    test_api_endpoints
                    ;;
                7)
                    cleanup
                    ;;
                8)
                    log_activity "Goodbye!"
                    break
                    ;;
                *)
                    echo "Invalid choice. Please try again."
                    ;;
            esac
            
            echo
            echo "Press Enter to continue..."
            read -r
        done
    else
        # Command line mode
        case $1 in
            "files")
                create_test_files
                ;;
            "workflow")
                simulate_development
                ;;
            "continuous")
                generate_continuous_activity
                ;;
            "websockets")
                test_websockets
                ;;
            "api")
                test_api_endpoints
                ;;
            "all")
                create_test_files
                sleep 2
                simulate_development
                sleep 2
                test_api_endpoints
                ;;
            "cleanup")
                cleanup
                ;;
            *)
                echo "Usage: $0 [files|workflow|continuous|websockets|api|all|cleanup]"
                exit 1
                ;;
        esac
    fi
}

# Trap Ctrl+C for cleanup
trap cleanup EXIT

# Execute
main "$@" 