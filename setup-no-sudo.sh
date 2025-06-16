#!/bin/bash

# Claude Code Intelligence Service - No-Sudo Setup
echo "🚀 Setting up Claude Code Intelligence Service (No sudo required)..."

# Create local bin directory
mkdir -p ~/bin
export PATH="$HOME/bin:$PATH"

# Function to download and install Go locally
install_go_local() {
    echo "📦 Installing Go locally..."
    
    GO_VERSION="1.21.0"
    GO_OS="linux"
    GO_ARCH="amd64"
    
    # Download Go
    cd ~
    wget -q "https://golang.org/dl/go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
    
    # Extract to home directory
    rm -rf ~/go
    tar -xzf "go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
    
    # Add Go to PATH
    export PATH="$HOME/go/bin:$PATH"
    echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc
    
    # Clean up
    rm "go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
    
    echo "✅ Go installed locally: $(~/go/bin/go version)"
}

# Function to install jq locally
install_jq_local() {
    echo "📦 Installing jq locally..."
    
    # Download jq binary
    wget -q -O ~/bin/jq "https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64"
    chmod +x ~/bin/jq
    
    echo "✅ jq installed locally"
}

# Check if Go is available
if ! command -v go &> /dev/null; then
    install_go_local
else
    echo "✅ Go is already available: $(go version)"
fi

# Check if jq is available
if ! command -v jq &> /dev/null; then
    install_jq_local
else
    echo "✅ jq is already available"
fi

# curl is usually pre-installed in WSL, but check anyway
if ! command -v curl &> /dev/null; then
    echo "⚠️  curl not found. You may need to install it manually."
else
    echo "✅ curl is available"
fi

# Navigate to project directory
cd "$(dirname "$0")"

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod tidy

# Make CLI script executable
echo "🔧 Making CLI script executable..."
chmod +x claude-query.sh

# Test compilation
echo "🧪 Testing Go compilation..."
if go build -o claude-intelligence main.go; then
    echo "✅ Go compilation successful!"
    echo "   You can now run: ./claude-intelligence"
    rm -f claude-intelligence  # Clean up test binary
else
    echo "❌ Go compilation failed. Check the error messages above."
    exit 1
fi

echo ""
echo "🎉 Setup complete! No sudo required!"
echo ""
echo "To start the service:"
echo "   go run main.go ."
echo ""
echo "To test the CLI:"
echo "   ./claude-query.sh quick"
echo ""
echo "To open dashboard:"
echo "   Open dashboard.html in your browser"
echo ""
echo "📝 Note: Run 'source ~/.bashrc' or restart your terminal to ensure PATH is updated" 