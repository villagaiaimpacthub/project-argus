#!/bin/bash

# Claude Code Intelligence Service Setup Script
echo "🚀 Setting up Claude Code Intelligence Service..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "📦 Installing Go..."
    
    # Try different methods to install Go
    if command -v apt &> /dev/null; then
        echo "Using apt package manager..."
        sudo apt update
        sudo apt install -y golang-go
    elif command -v brew &> /dev/null; then
        echo "Using Homebrew..."
        brew install go
    else
        echo "⚠️  Please install Go manually from https://golang.org/dl/"
        echo "Then run this script again."
        exit 1
    fi
else
    echo "✅ Go is already installed: $(go version)"
fi

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod tidy

# Make CLI script executable
echo "🔧 Making CLI script executable..."
chmod +x claude-query.sh

# Check if jq is installed (required for CLI)
if ! command -v jq &> /dev/null; then
    echo "📦 Installing jq..."
    if command -v apt &> /dev/null; then
        sudo apt install -y jq
    elif command -v brew &> /dev/null; then
        brew install jq
    else
        echo "⚠️  Please install jq manually for the CLI to work"
    fi
else
    echo "✅ jq is already installed"
fi

# Check if curl is installed
if ! command -v curl &> /dev/null; then
    echo "📦 Installing curl..."
    if command -v apt &> /dev/null; then
        sudo apt install -y curl
    elif command -v brew &> /dev/null; then
        brew install curl
    fi
else
    echo "✅ curl is already installed"
fi

echo ""
echo "🎉 Setup complete! To start the service:"
echo ""
echo "1. Start the service:"
echo "   go run main.go ."
echo ""
echo "2. In another terminal, test it:"
echo "   ./claude-query.sh quick"
echo ""
echo "3. Open dashboard.html in your browser"
echo ""
echo "4. Tell Claude Code to use: http://localhost:3002"
echo ""

# Try to start a quick test
echo "🧪 Testing Go compilation..."
if go build -o claude-intelligence main.go; then
    echo "✅ Go compilation successful!"
    echo "   You can now run: ./claude-intelligence"
    rm -f claude-intelligence  # Clean up test binary
else
    echo "❌ Go compilation failed. Check the error messages above."
fi

echo ""
echo "📖 For full integration instructions, see: complete_readme.md" 