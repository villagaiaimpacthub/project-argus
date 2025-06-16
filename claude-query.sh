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
        echo -e "${RED}Error: Cannot connect to Project Argus${NC}"
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

# New process monitoring commands
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
        pid=$(echo "$response" | jq -r '.process.pid // "unknown"')
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
        process_count=$(echo "$processes_data" | jq -r '.count // 0')
        
        if [ "$process_count" -eq 0 ]; then
            echo -e "${YELLOW}No processes being monitored${NC}"
            echo -e "${BLUE}Use '$0 monitor \"command\"' to start monitoring${NC}"
        else
            echo -e "${GREEN}$process_count monitored process(es):${NC}\n"
            
            echo "$processes_data" | jq -r '.processes[] | "\(.pid) \(.command) \(.status) \(.start_time)"' | while read pid command status start_time; do
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
                echo "Usage: $0 dev start [npm|go|python|next|vite]"
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
                echo "Usage: $0 dev stop [npm|go|python|next|vite]"
                return 1
            fi
            
            response=$(curl -s -X POST "$BASE_URL/dev/stop/$server_type")
            echo -e "${YELLOW}Stopped $server_type development server${NC}"
            echo "$response" | jq . 2>/dev/null
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
    
    # Use curl to poll for errors (fallback when websocat is not available)
    while true; do
        latest_errors=$(query_api "/errors/latest?since=10s")
        
        if [ $? -eq 0 ]; then
            error_count=$(echo "$latest_errors" | jq '.error_count // 0')
            
            if [ "$error_count" -gt 0 ]; then
                echo "$latest_errors" | jq -r '.errors[] | "\(.timestamp) [\(.command)] \(.error_type): \(.message)"' | while read line; do
                    echo -e "${RED}🚨 $line${NC}"
                done
            fi
        fi
        
        sleep 2
    done
}

cmd_stop_process() {
    local pid="$1"
    
    if [ -z "$pid" ]; then
        echo -e "${RED}Error: Process PID required${NC}"
        echo "Usage: $0 stop <pid>"
        return 1
    fi
    
    print_header "Stopping Process (PID: $pid)"
    
    response=$(curl -s -X DELETE "$BASE_URL/processes/$pid")
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Process $pid stopped successfully${NC}"
        echo "$response" | jq . 2>/dev/null
    else
        echo -e "${RED}❌ Failed to stop process $pid${NC}"
        echo "$response" | jq . 2>/dev/null || echo "$response"
    fi
}

cmd_help() {
    echo -e "${WHITE}Project Argus - Claude Code Intelligence Tool${NC}"
    echo -e "${CYAN}================================================${NC}"
    echo ""
    echo -e "${YELLOW}Usage:${NC} $0 [command] [options]"
    echo ""
    echo -e "${YELLOW}📊 Project Intelligence Commands:${NC}"
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
    echo ""
    echo -e "${YELLOW}⚡ Process Monitoring Commands:${NC}"
    echo -e "  ${GREEN}monitor${NC} \"command\" - Start monitoring a command"
    echo -e "  ${GREEN}monitored${NC}    - Show monitored processes"
    echo -e "  ${GREEN}logs${NC} [pid]    - Show process output"
    echo -e "  ${GREEN}stop${NC} [pid]    - Stop a monitored process"
    echo -e "  ${GREEN}dev${NC} [start|stop|status] [type] - Manage dev servers"
    echo -e "  ${GREEN}stream${NC}       - Stream real-time errors"
    echo ""
    echo -e "${YELLOW}🔍 Search & File Commands:${NC}"
    echo -e "  ${GREEN}search${NC} \"query\" - Search across project"
    echo -e "  ${GREEN}file${NC} \"path\"   - Get file information"
    echo -e "  ${GREEN}help${NC}         - Show this help message"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 quick"
    echo "  $0 monitor \"npm run dev\""
    echo "  $0 dev start npm"
    echo "  $0 logs 1234"
    echo "  $0 stream"
    echo "  $0 search \"TODO\""
    echo ""
    echo -e "${BLUE}Service URL: $BASE_URL${NC}"
    echo -e "${PURPLE}All-seeing project monitoring for Claude Code${NC}"
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
        "monitor")
            cmd_monitor "$2"
            ;;
        "monitored")
            cmd_monitored_processes
            ;;
        "logs"|"log")
            cmd_process_logs "$2" "$3"
            ;;
        "stop")
            cmd_stop_process "$2"
            ;;
        "dev")
            cmd_dev_server "$2" "$3"
            ;;
        "stream")
            cmd_stream_errors
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