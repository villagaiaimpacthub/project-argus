#!/bin/bash

# Enhanced Claude Code Project Intelligence Query Tool with Multi-Language Support
# Usage: ./claude-query-enhanced.sh [command] [options]

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

# Enhanced language commands

cmd_languages() {
    print_header "Detected Programming Languages"
    
    languages_data=$(query_api "/api/languages")
    
    if [ $? -eq 0 ]; then
        language_count=$(echo "$languages_data" | jq -r '.count // 0')
        
        if [ "$language_count" -eq 0 ]; then
            echo -e "${YELLOW}No languages detected yet${NC}"
        else
            echo -e "${GREEN}Found $language_count programming language(s):${NC}\n"
            
            echo "$languages_data" | jq -r '.languages[] | "\(.language.name) - \(.file_count) files (\(.line_count) lines)"' | while read line; do
                echo -e "${CYAN}🔧 $line${NC}"
            done
            
            echo -e "\n${BLUE}Frameworks detected:${NC}"
            echo "$languages_data" | jq -r '.languages[] | select(.frameworks | length > 0) | .frameworks[] | "\(.language): \(.name)"' | while read framework; do
                echo -e "${PURPLE}  📦 $framework${NC}"
            done
        fi
    fi
}

cmd_topology() {
    print_header "Project Topology & Architecture"
    
    topology_data=$(query_api "/api/topology")
    
    if [ $? -eq 0 ]; then
        languages_count=$(echo "$topology_data" | jq '.languages | length')
        services_count=$(echo "$topology_data" | jq '.services | length')
        
        echo -e "${GREEN}Languages: $languages_count${NC}"
        echo -e "${GREEN}Services: $services_count${NC}"
        
        print_section "Detected Services"
        echo "$topology_data" | jq -r '.services[] | "\(.name) (\(.language)/\(.framework)) - Port \(.port) [\(.status)]"' | while read service; do
            echo -e "${BLUE}⚡ $service${NC}"
        done
        
        print_section "Language Distribution"
        echo "$topology_data" | jq -r '.languages[] | "\(.language.name): \(.file_count) files"' | while read lang; do
            echo -e "${CYAN}📝 $lang${NC}"
        done
    fi
}

cmd_lang_errors() {
    local language="$1"
    
    if [ -z "$language" ]; then
        echo -e "${RED}Error: Language name required${NC}"
        echo "Usage: $0 lang-errors [javascript|typescript|python|go|java|etc.]"
        return 1
    fi
    
    print_header "Language-Specific Errors: $language"
    
    errors_data=$(query_api "/api/languages/$language/errors")
    
    if [ $? -eq 0 ]; then
        error_count=$(echo "$errors_data" | jq -r '.count // 0')
        
        if [ "$error_count" -eq 0 ]; then
            echo -e "${GREEN}✅ No $language errors detected!${NC}"
        else
            echo -e "${RED}Found $error_count $language error(s):${NC}\n"
            
            echo "$errors_data" | jq -r '.errors[] | "\(.file):\(.line): \(.type) - \(.message)"' | while read error; do
                echo -e "${RED}❌ $error${NC}"
            done
        fi
    else
        echo -e "${YELLOW}Language '$language' not found or not supported${NC}"
    fi
}

cmd_lang_deps() {
    local language="$1"
    
    if [ -z "$language" ]; then
        echo -e "${RED}Error: Language name required${NC}"
        echo "Usage: $0 lang-deps [javascript|typescript|python|go|java|etc.]"
        return 1
    fi
    
    print_header "Dependencies for $language"
    
    deps_data=$(query_api "/api/languages/$language/dependencies")
    
    if [ $? -eq 0 ]; then
        dep_count=$(echo "$deps_data" | jq -r '.count // 0')
        
        if [ "$dep_count" -eq 0 ]; then
            echo -e "${YELLOW}No dependencies found for $language${NC}"
        else
            echo -e "${BLUE}$dep_count dependencies found:${NC}\n"
            
            echo "$deps_data" | jq -r '.dependencies[] | "\(.name) \(.version) [\(.type)] - \(.source)"' | while read dep; do
                echo -e "${CYAN}📦 $dep${NC}"
            done
        fi
    fi
}

cmd_lang_services() {
    local language="$1"
    
    if [ -z "$language" ]; then
        echo -e "${RED}Error: Language name required${NC}"
        echo "Usage: $0 lang-services [javascript|typescript|python|go|java|etc.]"
        return 1
    fi
    
    print_header "Services for $language"
    
    services_data=$(query_api "/api/languages/$language/services")
    
    if [ $? -eq 0 ]; then
        service_count=$(echo "$services_data" | jq -r '.count // 0')
        
        if [ "$service_count" -eq 0 ]; then
            echo -e "${YELLOW}No services found for $language${NC}"
        else
            echo -e "${BLUE}$service_count service(s) found:${NC}\n"
            
            echo "$services_data" | jq -r '.services[] | "\(.name) - Port \(.port) [\(.status)]"' | while read service; do
                echo -e "${GREEN}⚡ $service${NC}"
            done
        fi
    fi
}

cmd_lint() {
    local language="$1"
    
    if [ -z "$language" ]; then
        echo -e "${RED}Error: Language name required${NC}"
        echo "Usage: $0 lint [javascript|typescript|python|go|java|etc.]"
        return 1
    fi
    
    print_header "Running Linter for $language"
    
    lint_data=$(curl -s -X POST "$BASE_URL/api/languages/$language/lint")
    
    if [ $? -eq 0 ]; then
        error_count=$(echo "$lint_data" | jq -r '.count // 0')
        
        if [ "$error_count" -eq 0 ]; then
            echo -e "${GREEN}✅ No linting errors found!${NC}"
        else
            echo -e "${YELLOW}Found $error_count linting issue(s):${NC}\n"
            
            echo "$lint_data" | jq -r '.lint_errors[] | "\(.file):\(.line): \(.message)"' | while read error; do
                echo -e "${YELLOW}⚠️  $error${NC}"
            done
        fi
    else
        echo -e "${RED}Failed to run linter for $language${NC}"
    fi
}

cmd_test() {
    local language="$1"
    
    if [ -z "$language" ]; then
        echo -e "${RED}Error: Language name required${NC}"
        echo "Usage: $0 test [javascript|typescript|python|go|java|etc.]"
        return 1
    fi
    
    print_header "Running Tests for $language"
    
    test_data=$(curl -s -X POST "$BASE_URL/api/languages/$language/test")
    
    if [ $? -eq 0 ]; then
        total_tests=$(echo "$test_data" | jq -r '.test_results.total_tests // 0')
        passed_tests=$(echo "$test_data" | jq -r '.test_results.passed_tests // 0')
        failed_tests=$(echo "$test_data" | jq -r '.test_results.failed_tests // 0')
        
        echo -e "${BLUE}Test Results:${NC}"
        echo -e "${GREEN}  Passed: $passed_tests${NC}"
        echo -e "${RED}  Failed: $failed_tests${NC}"
        echo -e "${CYAN}  Total: $total_tests${NC}"
        
        if [ "$failed_tests" -gt 0 ]; then
            echo -e "\n${RED}Test failures detected${NC}"
        else
            echo -e "\n${GREEN}✅ All tests passed!${NC}"
        fi
    else
        echo -e "${RED}Failed to run tests for $language${NC}"
    fi
}

cmd_overview() {
    print_header "Comprehensive Project Overview"
    
    overview_data=$(query_api "/api/project-overview")
    
    if [ $? -eq 0 ]; then
        primary_lang=$(echo "$overview_data" | jq -r '.languages.primary // "unknown"')
        lang_count=$(echo "$overview_data" | jq -r '.languages.count // 0')
        total_files=$(echo "$overview_data" | jq -r '.structure.total_files // 0')
        health_score=$(echo "$overview_data" | jq -r '.health.score // 0')
        error_count=$(echo "$overview_data" | jq -r '.errors.count // 0')
        service_count=$(echo "$overview_data" | jq -r '.services.count // 0')
        
        echo -e "${BLUE}Project Type: Multi-language development project${NC}"
        echo -e "${CYAN}Primary Language: $primary_lang${NC}"
        echo -e "${CYAN}Languages Detected: $lang_count${NC}"
        echo -e "${CYAN}Total Files: $total_files${NC}"
        echo -e "${CYAN}Services Running: $service_count${NC}"
        
        # Health indicator
        if [ "$health_score" -ge 80 ]; then
            echo -e "${GREEN}Health Score: $health_score/100 ✅${NC}"
        elif [ "$health_score" -ge 50 ]; then
            echo -e "${YELLOW}Health Score: $health_score/100 ⚠️${NC}"
        else
            echo -e "${RED}Health Score: $health_score/100 ❌${NC}"
        fi
        
        if [ "$error_count" -gt 0 ]; then
            echo -e "${RED}Active Errors: $error_count${NC}"
        else
            echo -e "${GREEN}No Active Errors ✅${NC}"
        fi
        
        print_section "Detected Languages"
        echo "$overview_data" | jq -r '.languages.detected[] | "\(.language.name): \(.file_count) files"' | while read lang; do
            echo -e "${CYAN}  🔧 $lang${NC}"
        done
        
        if [ "$service_count" -gt 0 ]; then
            print_section "Running Services"
            echo "$overview_data" | jq -r '.services.detected[] | "\(.name) - \(.language)/\(.framework) on port \(.port)"' | while read service; do
                echo -e "${GREEN}  ⚡ $service${NC}"
            done
        fi
    fi
}

cmd_analyze_all() {
    print_header "Analyzing All Languages"
    
    echo -e "${BLUE}🔍 Starting comprehensive multi-language analysis...${NC}"
    
    analyze_data=$(curl -s -X POST "$BASE_URL/api/analyze/all-languages")
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Analysis started successfully${NC}"
        echo -e "${YELLOW}📊 Check individual language results with:${NC}"
        echo -e "  ${CYAN}$0 lang-errors [language]${NC}"
        echo -e "  ${CYAN}$0 overview${NC}"
    else
        echo -e "${RED}❌ Failed to start analysis${NC}"
    fi
}

# Enhanced help command
cmd_help() {
    echo -e "${WHITE}Project Argus - Universal Development Intelligence Tool${NC}"
    echo -e "${CYAN}=====================================================${NC}"
    echo ""
    echo -e "${YELLOW}Usage:${NC} $0 [command] [options]"
    echo ""
    echo -e "${YELLOW}🌐 Multi-Language Intelligence Commands:${NC}"
    echo -e "  ${GREEN}languages${NC}     - Show all detected programming languages"
    echo -e "  ${GREEN}topology${NC}      - Display project architecture and relationships"
    echo -e "  ${GREEN}overview${NC}      - Comprehensive project overview"
    echo -e "  ${GREEN}analyze-all${NC}   - Analyze all languages comprehensively"
    echo ""
    echo -e "${YELLOW}🔧 Language-Specific Commands:${NC}"
    echo -e "  ${GREEN}lang-errors${NC} [lang] - Show errors for specific language"
    echo -e "  ${GREEN}lang-deps${NC} [lang]   - Show dependencies for language"
    echo -e "  ${GREEN}lang-services${NC} [lang] - Show services for language"
    echo -e "  ${GREEN}lint${NC} [lang]        - Run linter for specific language"
    echo -e "  ${GREEN}test${NC} [lang]        - Run tests for specific language"
    echo ""
    echo -e "${YELLOW}📊 Standard Intelligence Commands:${NC}"
    echo -e "  ${GREEN}quick${NC}         - Quick project overview"
    echo -e "  ${GREEN}health${NC}        - Project health summary"
    echo -e "  ${GREEN}errors${NC}        - Show all active errors"
    echo -e "  ${GREEN}structure${NC}     - Project file structure"
    echo -e "  ${GREEN}dependencies${NC}  - All project dependencies"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 languages                    # Show detected languages"
    echo "  $0 lang-errors typescript       # TypeScript-specific errors"
    echo "  $0 lint javascript              # Run ESLint"
    echo "  $0 test python                  # Run Python tests"
    echo "  $0 topology                     # Project architecture"
    echo "  $0 overview                     # Complete overview"
    echo ""
    echo -e "${YELLOW}Supported Languages:${NC}"
    echo -e "  ${CYAN}JavaScript, TypeScript, Python, Go, Java, C#, Rust, PHP, Ruby${NC}"
    echo ""
    echo -e "${BLUE}Service URL: $BASE_URL${NC}"
    echo -e "${PURPLE}Universal development monitoring for Claude Code${NC}"
}

# Main command router
main() {
    case "${1:-help}" in
        "languages"|"langs"|"l")
            cmd_languages
            ;;
        "topology"|"topo"|"arch")
            cmd_topology
            ;;
        "lang-errors"|"lerrors"|"le")
            cmd_lang_errors "$2"
            ;;
        "lang-deps"|"ldeps"|"ld")
            cmd_lang_deps "$2"
            ;;
        "lang-services"|"lservices"|"ls")
            cmd_lang_services "$2"
            ;;
        "lint"|"linter")
            cmd_lint "$2"
            ;;
        "test"|"tests")
            cmd_test "$2"
            ;;
        "overview"|"ov"|"summary")
            cmd_overview
            ;;
        "analyze-all"|"analyze"|"aa")
            cmd_analyze_all
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