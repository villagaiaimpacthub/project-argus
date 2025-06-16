#!/bin/bash

echo "🚀 Setting up Git repository and pushing Claude Code Live Monitoring Dashboard..."

# Initialize git repository if it doesn't exist
echo "📋 Initializing git repository..."
git init

# Add all files to git
echo "➕ Adding all project files..."
git add .

# Initial commit
echo "💾 Creating initial commit..."
git commit -m "🎉 Initial commit: Claude Code Live Monitoring Dashboard

✨ Features:
- Beautiful dark mode WebSocket monitoring interface
- Real-time error and process stream monitoring  
- Simple 1-2-3-4 workflow for easy setup
- Enterprise-grade monitoring capabilities
- Production-ready with enhanced debugging
- Responsive design with status indicators
- Go Fiber backend with WebSocket streaming
- Project Argus monitoring system"

# Set up main branch (modern git default)
echo "🔧 Setting up main branch..."
git branch -M main

echo "🌐 Ready to add remote repository!"
echo ""
echo "Next steps:"
echo "1. Create a new repository on GitHub"
echo "2. Copy the repository URL"
echo "3. Run: git remote add origin <your-github-repo-url>"
echo "4. Run: git push -u origin main"
echo ""
echo "Or if you already have a GitHub repo URL, provide it and I'll complete the setup!" 