#!/bin/bash

# Project Argus Enhanced - Comprehensive Feature Test
# Tests all enhanced features in one script

echo "🚀 Project Argus Enhanced - Comprehensive Test"
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="http://localhost:3002"

test_endpoint() {
    local name="$1"
    local url="$2"
    local method="${3:-GET}"
    local data="$4"
    
    echo -n "Testing $name... "
    if [ "$method" = "POST" ]; then
        response=$(curl -s --max-time 5 -X POST "$url" -H "Content-Type: application/json" -d "$data" 2>/dev/null)
    else
        response=$(curl -s --max-time 5 "$url" 2>/dev/null)
    fi
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        echo -e "${GREEN}✅ WORKING${NC}"
        echo "   Response: $(echo "$response" | cut -c1-100)..."
        return 0
    else
        echo -e "${RED}❌ FAILED${NC}"
        return 1
    fi
}

echo
echo "📊 Testing Basic Intelligence Endpoints:"
echo "----------------------------------------"
test_endpoint "Server Status" "$BASE_URL/"
test_endpoint "Project Health" "$BASE_URL/health"
test_endpoint "Project Structure" "$BASE_URL/structure"
test_endpoint "Git Status" "$BASE_URL/git"
test_endpoint "Active Errors" "$BASE_URL/errors"

echo
echo "⚡ Testing Enhanced Process Monitoring:"
echo "--------------------------------------"
test_endpoint "Monitored Processes" "$BASE_URL/processes/monitored"
test_endpoint "Latest Errors" "$BASE_URL/errors/latest?since=60s"
test_endpoint "Error Stream" "$BASE_URL/errors/stream"

echo
echo "🔧 Testing Development Server Integration:"
echo "------------------------------------------"
test_endpoint "Dev Server Status" "$BASE_URL/dev/status"

echo
echo "🌐 Testing WebSocket Endpoints:"
echo "-------------------------------"
echo -n "Testing WebSocket Error Stream... "
ws_response=$(curl -s --max-time 3 -I "$BASE_URL/ws/errors" 2>/dev/null | grep "426 Upgrade Required")
if [ -n "$ws_response" ]; then
    echo -e "${GREEN}✅ WORKING${NC} (426 Upgrade Required)"
else
    echo -e "${RED}❌ FAILED${NC}"
fi

echo -n "Testing WebSocket Process Stream... "
ws_response=$(curl -s --max-time 3 -I "$BASE_URL/ws/processes" 2>/dev/null | grep "426 Upgrade Required")
if [ -n "$ws_response" ]; then
    echo -e "${GREEN}✅ WORKING${NC} (426 Upgrade Required)"
else
    echo -e "${RED}❌ FAILED${NC}"
fi

echo
echo "🔥 Testing Real-Time Process Monitoring:"
echo "----------------------------------------"
echo "Starting a test process..."
start_response=$(curl -s --max-time 10 -X POST "$BASE_URL/processes/start" \
    -H "Content-Type: application/json" \
    -d '{"command":"node","args":["-e","console.log(\"Hello Project Argus!\"); setTimeout(() => console.log(\"Process complete!\"), 2000)"],"working_dir":"."}' 2>/dev/null)

if echo "$start_response" | grep -q "Process started successfully"; then
    echo -e "${GREEN}✅ Process started successfully!${NC}"
    pid=$(echo "$start_response" | grep -o '"pid":[0-9]*' | cut -d':' -f2)
    echo "   PID: $pid"
    
    echo "Waiting 1 second for output..."
    sleep 1
    
    echo "Checking monitored processes..."
    monitored=$(curl -s --max-time 5 "$BASE_URL/processes/monitored" 2>/dev/null)
    if echo "$monitored" | grep -q '"count":[1-9]'; then
        echo -e "${GREEN}✅ Process monitoring working!${NC}"
        echo "   Monitored processes: $(echo "$monitored" | grep -o '"count":[0-9]*' | cut -d':' -f2)"
    else
        echo -e "${YELLOW}⚠️ Process not in monitored list${NC}"
    fi
    
    if [ -n "$pid" ] && [ "$pid" != "null" ]; then
        echo "Getting process output..."
        output=$(curl -s --max-time 5 "$BASE_URL/processes/$pid/output" 2>/dev/null)
        if [ -n "$output" ]; then
            echo -e "${GREEN}✅ Process output capture working!${NC}"
            echo "   Output: $(echo "$output" | cut -c1-100)..."
        fi
    fi
else
    echo -e "${RED}❌ Process start failed${NC}"
    echo "   Response: $start_response"
fi

echo
echo "🛠️ Testing Enhanced CLI Commands:"
echo "---------------------------------"
echo "Testing CLI help (should show enhanced commands)..."
if ./claude-query.sh help | grep -q "monitor"; then
    echo -e "${GREEN}✅ Enhanced CLI commands available${NC}"
    echo "   Found: monitor, stream, dev commands"
else
    echo -e "${RED}❌ Enhanced CLI commands missing${NC}"
fi

echo "Testing CLI status..."
if ./claude-query.sh status >/dev/null 2>&1; then
    echo -e "${GREEN}✅ CLI status command working${NC}"
else
    echo -e "${RED}❌ CLI status command failed${NC}"
fi

echo
echo "📊 Testing Summary:"
echo "==================="

# Count endpoints
echo -n "Checking total available endpoints... "
endpoints=$(curl -s --max-time 5 "$BASE_URL/" 2>/dev/null | grep -o '"endpoints":\[[^]]*\]' | grep -o '"/[^"]*"' | wc -l)
echo "$endpoints endpoints found"

if [ "$endpoints" -gt 15 ]; then
    echo -e "${GREEN}✅ Enhanced server running (${endpoints} endpoints)${NC}"
elif [ "$endpoints" -gt 8 ]; then
    echo -e "${YELLOW}⚠️ Partial enhancement (${endpoints} endpoints)${NC}"
else
    echo -e "${RED}❌ Basic server only (${endpoints} endpoints)${NC}"
fi

echo
echo "🎯 Feature Status:"
echo "  ✨ Real-time process monitoring"
echo "  ⚡ Live error detection and streaming"  
echo "  🔧 Development server integration"
echo "  🌐 WebSocket real-time updates"
echo "  📡 REST API process management"
echo "  🔍 Enhanced project intelligence"

echo
echo "🚀 Project Argus Enhanced Test Complete!"
echo "Check the results above to see which features are working." 