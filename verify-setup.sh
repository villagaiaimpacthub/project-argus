#!/bin/bash

echo "🔍 Verifying Claude Code Intelligence Service Setup..."
echo ""

# Check required files
required_files=("main.go" "go.mod" "claude-query.sh" "dashboard.html")
missing_files=()

for file in "${required_files[@]}"; do
    if [[ -f "$file" ]]; then
        size=$(stat -c%s "$file" 2>/dev/null || stat -f%z "$file" 2>/dev/null || echo "unknown")
        echo "✅ $file (${size} bytes)"
    else
        echo "❌ $file - MISSING"
        missing_files+=("$file")
    fi
done

echo ""

# Check file permissions
if [[ -f "claude-query.sh" ]]; then
    if [[ -x "claude-query.sh" ]]; then
        echo "✅ claude-query.sh is executable"
    else
        echo "⚠️  claude-query.sh needs execute permission"
        echo "   Run: chmod +x claude-query.sh"
    fi
fi

# Check Go installation
echo ""
echo "🔧 Checking dependencies..."
if command -v go &> /dev/null; then
    echo "✅ Go: $(go version)"
else
    echo "❌ Go not found - required for intelligence service"
fi

if command -v curl &> /dev/null; then
    echo "✅ curl: $(curl --version | head -1)"
else
    echo "❌ curl not found - required for CLI"
fi

if command -v jq &> /dev/null; then
    echo "✅ jq: $(jq --version)"
else
    echo "❌ jq not found - required for CLI"
fi

# Check go.mod
echo ""
echo "📦 Checking Go module..."
if [[ -f "go.mod" ]]; then
    echo "Module: $(head -1 go.mod)"
    if grep -q "github.com/gofiber/fiber/v2" go.mod; then
        echo "✅ Fiber dependency found"
    else
        echo "⚠️  Fiber dependency missing - run 'go mod tidy'"
    fi
fi

# Summary
echo ""
if [[ ${#missing_files[@]} -eq 0 ]]; then
    echo "🎉 All required files are present!"
    echo ""
    echo "🚀 Ready to start! Run one of:"
    echo "   bash setup.sh              # Automated setup"
    echo "   go run main.go .           # Start service directly"
    echo "   ./claude-query.sh help     # CLI help"
else
    echo "❌ Missing files: ${missing_files[*]}"
    echo "Please ensure all required files are created."
fi

echo ""
echo "📖 For complete setup instructions, see complete_readme.md" 