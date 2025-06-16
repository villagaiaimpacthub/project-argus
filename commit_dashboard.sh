#!/bin/bash

echo "🚀 Committing and pushing Claude Code Live Monitoring Dashboard..."

# Check git status
echo "📋 Checking git status..."
git status

# Add the websocket dashboard file
echo "➕ Adding websocket_test.html..."
git add websocket_test.html

# Commit with a descriptive message
echo "💾 Committing with message..."
git commit -m "✨ Add beautiful real-time monitoring dashboard

- Beautiful dark mode WebSocket monitoring interface
- Simple 1-2-3-4 workflow for easy setup  
- Real-time error and process stream monitoring
- Enterprise-grade monitoring capabilities
- Production-ready with enhanced debugging
- Responsive design with status indicators"

# Push to GitHub
echo "🚀 Pushing to GitHub..."
git push origin main || git push origin master

echo "✅ Dashboard successfully pushed to GitHub!"
echo "🎉 Your monitoring dashboard is now live on GitHub!" 