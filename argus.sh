#!/bin/bash

# Project Argus Launcher Script
# Makes it easy to monitor any project directory

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ASCII Art Banner
echo -e "${PURPLE}"
echo "  ╔═══════════════════════════════════════════════════════════╗"
echo "  ║                    🚀 PROJECT ARGUS                      ║"
echo "  ║              Real-time Project Intelligence               ║"
echo "  ╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Default values
WORKSPACE=""
PORT="3002"
HELP=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help|help)
            HELP=true
            shift
            ;;
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -w|--workspace)
            WORKSPACE="$2"
            shift 2
            ;;
        *)
            if [[ -z "$WORKSPACE" ]]; then
                WORKSPACE="$1"
            fi
            shift
            ;;
    esac
done

# Show help if requested
if [[ "$HELP" == "true" ]]; then
    echo -e "${CYAN}USAGE:${NC}"
    echo "  ./argus.sh [OPTIONS] [WORKSPACE_PATH]"
    echo ""
    echo -e "${CYAN}OPTIONS:${NC}"
    echo "  -h, --help              Show this help message"
    echo "  -p, --port PORT         Port to run on (default: 3002)"
    echo "  -w, --workspace PATH    Project directory to monitor"
    echo ""
    echo -e "${CYAN}EXAMPLES:${NC}"
    echo "  ./argus.sh                           # Monitor current directory"
    echo "  ./argus.sh /path/to/project          # Monitor specific directory"
    echo "  ./argus.sh --port 3003 ../my-app    # Monitor on custom port"
    echo "  ./argus.sh -w ~/projects/task-dash   # Using --workspace flag"
    echo ""
    echo -e "${CYAN}QUICK START:${NC}"
    echo "  1. Run this script with your project path"
    echo "  2. Open http://localhost:$PORT in your browser"
    echo "  3. Or open websocket_test.html for the test dashboard"
    echo ""
    exit 0
fi

# Use current directory if no workspace specified
if [[ -z "$WORKSPACE" ]]; then
    WORKSPACE="."
fi

# Convert to absolute path
WORKSPACE=$(realpath "$WORKSPACE" 2>/dev/null || echo "$WORKSPACE")

# Validate workspace exists
if [[ ! -d "$WORKSPACE" ]]; then
    echo -e "${RED}❌ Error: Directory '$WORKSPACE' does not exist${NC}"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Error: Go is not installed or not in PATH${NC}"
    echo -e "${YELLOW}💡 Install Go from: https://golang.org/dl/${NC}"
    exit 1
fi

# Check if main.go exists
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ ! -f "$SCRIPT_DIR/main.go" ]]; then
    echo -e "${RED}❌ Error: main.go not found in script directory${NC}"
    echo -e "${YELLOW}💡 Make sure you're running this from the Project Argus directory${NC}"
    exit 1
fi

echo -e "${GREEN}🎯 Target Directory:${NC} $WORKSPACE"
echo -e "${GREEN}🌐 Server Port:${NC} $PORT"
echo -e "${GREEN}📂 Project Type:${NC} $(detect_project_type "$WORKSPACE")"
echo ""

# Detect project type for better UX
function detect_project_type() {
    local dir="$1"
    if [[ -f "$dir/package.json" ]]; then
        echo "Node.js/JavaScript"
    elif [[ -f "$dir/go.mod" ]]; then
        echo "Go"
    elif [[ -f "$dir/requirements.txt" ]] || [[ -f "$dir/pyproject.toml" ]]; then
        echo "Python"
    elif [[ -f "$dir/Cargo.toml" ]]; then
        echo "Rust"
    elif [[ -f "$dir/composer.json" ]]; then
        echo "PHP"
    elif [[ -f "$dir/pom.xml" ]]; then
        echo "Java (Maven)"
    elif [[ -f "$dir/build.gradle" ]]; then
        echo "Java (Gradle)"
    elif [[ -f "$dir/Gemfile" ]]; then
        echo "Ruby"
    else
        echo "Generic"
    fi
}

echo -e "${BLUE}🚀 Starting Project Argus...${NC}"
echo ""

# Set environment variables and start
export ARGUS_WORKSPACE="$WORKSPACE"
export ARGUS_PORT="$PORT"

# Start the server
cd "$SCRIPT_DIR"
echo -e "${CYAN}📡 Dashboard will be available at: ${GREEN}http://localhost:$PORT${NC}"
echo -e "${CYAN}🎨 Test Dashboard: Open ${GREEN}websocket_test.html${CYAN} in your browser${NC}"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop monitoring${NC}"
echo ""

go run main.go "$WORKSPACE" 