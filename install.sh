#!/bin/bash

# Project Argus Installation Script  
# Usage: curl -sSL https://raw.githubusercontent.com/villagaiaimpacthub/project-argus/main/install.sh | bash

set -e

echo "👁️ Installing Project Argus..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.18+ first:"
    echo "   - Ubuntu/Debian: sudo apt install golang-go"
    echo "   - macOS: brew install go"
    echo "   - Windows: Download from https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | cut -c3-)
REQUIRED_VERSION="1.18"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go version $GO_VERSION is too old. Please upgrade to Go 1.18+"
    exit 1
fi

# Create installation directory
INSTALL_DIR="$HOME/project-argus"
if [ -d "$INSTALL_DIR" ]; then
    echo "📁 Updating existing installation..."
    cd "$INSTALL_DIR"
    git pull
else
    echo "📁 Creating installation directory..."
    git clone https://github.com/villagaiaimpacthub/project-argus.git "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Make CLI executable
chmod +x claude-query.sh

# Check for required tools
echo "🔧 Checking required tools..."
if ! command -v jq &> /dev/null; then
    echo "⚠️  jq not found. Install with:"
    echo "   - Ubuntu/Debian: sudo apt install jq"
    echo "   - macOS: brew install jq"
fi

if ! command -v curl &> /dev/null; then
    echo "⚠️  curl not found. Install with:"
    echo "   - Ubuntu/Debian: sudo apt install curl"
    echo "   - macOS: brew install curl"
fi

# Create symlinks for global access (optional)
if [ -w "/usr/local/bin" ]; then
    ln -sf "$INSTALL_DIR/claude-query.sh" "/usr/local/bin/claude-query"
    echo "🔗 Created global symlink: claude-query"
fi

# Copy example config
if [ ! -f config.json ]; then
    cp config.example.json config.json
    echo "⚙️  Created config.json from example"
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "🚀 Quick Start:"
echo "   cd $INSTALL_DIR"
echo "   go run main.go ."
echo ""
echo "💻 CLI Usage:"
echo "   ./claude-query.sh quick"
echo "   ./claude-query.sh errors"
echo ""
echo "🌐 Web Dashboard:"
echo "   Open dashboard.html in your browser"
echo ""
echo "📖 Documentation:"
echo "   https://github.com/villagaiaimpacthub/project-argus#readme" 