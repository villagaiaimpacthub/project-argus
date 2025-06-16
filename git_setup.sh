#!/bin/bash

echo "🚀 Setting up Project Argus Git Repository..."

# Initialize git repository with main branch
echo "📋 Initializing git repository..."
git init -b main

# Configure git for cross-platform compatibility  
echo "🔧 Configuring git settings..."
git config core.filemode false
git config core.autocrlf false

# Add the existing GitHub repository as remote
echo "🌐 Adding Project Argus GitHub repository..."
git remote add origin https://github.com/villagaiaimpacthub/project-argus.git

# Add all files to staging
echo "➕ Adding all files..."
git add .

# Create the initial commit
echo "💾 Creating initial commit..."
git commit -m "Initial commit: Project Argus - Go Fiber monitoring system

- Real-time project intelligence and monitoring platform
- WebSocket streams for live file, process, and git monitoring  
- Beautiful dark mode monitoring dashboard with 1-2-3-4 workflow
- Comprehensive test suite with self-monitoring capabilities
- Build, error, and process tracking with REST API endpoints
- Auto-detection of development processes and workspace monitoring
- Enterprise-grade real-time monitoring with WebSocket streaming

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>"

# Try to push to the GitHub repository
echo "🚀 Pushing to GitHub..."
if git push -u origin main; then
    echo "✅ Successfully pushed to GitHub!"
else
    echo "⚠️ Push failed, trying with pull first..."
    git pull origin main --allow-unrelated-histories
    git push origin main
fi

echo "🎉 Project Argus is now live on GitHub!"
echo "🌐 Repository: https://github.com/villagaiaimpacthub/project-argus" 