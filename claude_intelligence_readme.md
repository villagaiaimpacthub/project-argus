# Claude Code Project Intelligence Service

A comprehensive monitoring and intelligence service that gives Claude Code deeper insight into your development environment, specifically designed for **Claude Code in WSL + Cursor IDE on Windows VM** setups.

## 🎯 Problem This Solves

- **Context Gap**: Claude Code can't see your full project structure efficiently
- **Manual Updates**: You have to constantly tell Claude Code what's happening in your project
- **Integration Limitation**: Can't integrate Claude Code natively with Cursor IDE yet (Windows VM limitation)
- **Slow Iteration**: Claude Code works slower without real-time project awareness

## 🚀 What This Service Provides

### Real-Time Project Intelligence
- **Live file structure monitoring** with language detection
- **Error tracking** from TypeScript, Go, Python compilers
- **Git status monitoring** including branch, commits, and dirty state
- **TODO/FIXME scanning** across your entire codebase
- **Build process monitoring** and status tracking
- **Dependency analysis** for Node.js, Go, Python projects
- **Process monitoring** for project-related services

### Claude Code Integration
- **REST API endpoints** that Claude Code can query directly
- **CLI tool** for quick project queries from terminal
- **Real-time dashboard** for visual monitoring
- **Search functionality** across files, errors, and TODOs
- **Project health scoring** with actionable insights

## 📋 Prerequisites

- **Go 1.19+** (for the intelligence service)
- **jq** (for CLI tool JSON parsing)
- **curl** (for API requests)
- **Git** (for repository monitoring)

### Install Dependencies

**Ubuntu/WSL:**
```bash
# Install Go
sudo apt update
sudo apt install golang-go jq curl git

# Or install specific Go version
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

**macOS:**
```bash
brew install go jq curl git
```

## 🛠️ Setup Instructions

### 1. Create Project Directory
```bash
mkdir claude-intelligence
cd claude-intelligence
```

### 2. Initialize Go Module
```bash
go mod init claude-intelligence
go get github.com/gofiber/fiber/v2
```

### 3. Create the Service
Save the Go code as `main.go`, then run:

```bash
# Start monitoring your current project
go run main.go .

# Or monitor a specific project directory
go run main.go /path/to/your/project

# Or monitor from your WSL project root
go run main.go /mnt/c/Users/YourName/Projects/your-project
```

### 4. Set Up CLI Tool
```bash
# Make the CLI tool executable
chmod +x claude-query.sh

# Create symlink for easy access (optional)
sudo ln -s $(pwd)/claude-query.sh /usr/local/bin/claude-query
```

### 5. Open Dashboard
Save the dashboard HTML and open in your browser:
```
http://localhost:3002
```

## 🎮 Usage Examples

### For Claude Code in Terminal

**Quick project overview:**
```bash
./claude-query.sh quick
```

**Check for errors:**
```bash
./claude-query.sh errors
```

**See recent changes:**
```bash
./claude-query.sh changes
```

**Search for specific items:**
```bash
./claude-query.sh search \"TODO authentication\"
./claude-query.sh search \"error handling\"
```

**Get file information:**
```bash
./claude-query.sh file \"src/main.go\"
./claude-query.sh file \"package.json\"
```

### For Claude Code API Integration

Claude Code can directly query these endpoints:

```bash
# Get complete project snapshot
curl http://localhost:3002/snapshot

# Get only errors
curl http://localhost:3002/errors

# Get project structure
curl http://localhost:3002/structure

# Search across project
curl http://localhost:3002/search?q=authentication

# Get specific file content
curl http://localhost:3002/files/src/main.go/content

# Get git status
curl http://localhost:3002/git

# Get recent changes
curl http://localhost:3002/changes
```

## 🔧 Integration with Claude Code

### Method 1: CLI Integration
Tell Claude Code to use the CLI tool:

```
When working on this project, you can use the claude-query tool to get real-time information:

- Run `./claude-query.sh quick` for project overview
- Run `./claude-query.sh errors` to see current errors
- Run `./claude-query.sh changes` for recent file changes
- Run `./claude-query.sh search \"term\"` to search the project

Please use these commands to stay updated on the project state.
```

### Method 2: Direct API Usage
Claude Code can make HTTP requests:

```
The project intelligence service is running at http://localhost:3002

Key endpoints for you to use:
- GET /snapshot - Complete project state
- GET /errors - Current compilation/lint errors  
- GET /structure - File structure and project type
- GET /git - Git repository status
- GET /search?q=term - Search across project
- GET /changes - Recent file modifications

Use these to understand the current project state before making changes.
```

### Method 3: Automated Context
Set up a script that automatically provides context:

```bash
#!/bin/bash
# auto-context.sh - Run before Claude Code sessions

echo "=== PROJECT CONTEXT FOR CLAUDE ===" 
./claude-query.sh quick
echo ""
echo "=== RECENT ERRORS ==="
./claude-query.sh errors
echo ""
echo "=== RECENT CHANGES ==="
./claude-query.sh changes
echo ""
echo "=== CURRENT GIT STATUS ==="
./claude-query.sh git
```

## 📊 Dashboard Features

The web dashboard provides:

- **Real-time project health score** (0-100)
- **Active errors and warnings** with file locations
- **Recent file changes** with timestamps
- **Git repository status** including branch and changes
- **TODO items** found in your code
- **Language breakdown** of your project
- **Running processes** related to your project
- **Quick search** across all project data
- **API endpoint reference** for Claude Code integration

## 🚀 Advanced Configuration

### Custom File Patterns
Modify the `detectLanguage()` function to add support for more file types:

```go
langMap := map[string]string{
    ".vue":    "vue",
    ".svelte": "svelte", 
    ".jsx":    "react",
    ".tsx":    "react-ts",
    // Add your custom mappings
}
```

### Custom Error Patterns
Add more error detection in `scanLogFiles()`:

```go
errorPatterns := []string{
    "error", "exception", "fatal", 
    "panic", "failed", "denied",
    // Add your patterns
}
```

### Performance Tuning
Adjust monitoring intervals in the code:

```go
// File watching interval
time.Sleep(5 * time.Second)  // Increase for less frequent checks

// Git status checking  
time.Sleep(15 * time.Second) // Increase for better performance

// Error scanning
time.Sleep(10 * time.Second) // Increase if CPU usage is high
```

## 🔍 API Reference

### Core Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Service status and endpoint list |
| `/snapshot` | GET | Complete project snapshot |
| `/structure` | GET | Project file structure |
| `/errors` | GET | Active errors and warnings |
| `/git` | GET | Git repository status |
| `/changes` | GET | Recent file changes |
| `/todos` | GET | TODO items in code |
| `/dependencies` | GET | Project dependencies |
| `/processes` | GET | Running processes |
| `/health` | GET | Project health metrics |

### Search & Query

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/search?q=term` | GET | Search across project |
| `/files/{path}` | GET | File information |
| `/files/{path}/content` | GET | File content |
| `/query/errors-by-file` | GET | Errors grouped by file |
| `/query/recent-activity` | GET | Recent project activity |
| `/query/project-stats` | GET | Project statistics |

### Actions

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/refresh` | POST | Trigger manual refresh |
| `/analyze` | POST | Trigger deep analysis |

## 🐛 Troubleshooting

### Service Won't Start
```bash
# Check if port is in use
netstat -tulpn | grep :3002

# Kill existing process
pkill -f "claude-intelligence"

# Check Go installation
go version
```

### CLI Tool Not Working
```bash
# Check dependencies
which jq curl
jq --version

# Test service connection
curl http://localhost:3002/

# Check file permissions
chmod +x claude-query.sh
```

### Missing Project Data
```bash
# Verify workspace path
ls -la /path/to/your/project

# Check service logs
go run main.go . 2>&1 | tail -f

# Trigger manual refresh
curl -X POST http://localhost:3002/refresh
```

### Performance Issues
```bash
# Check system resources
top | grep claude-intelligence

# Reduce monitoring frequency in code
# Increase sleep intervals in main.go

# Exclude large directories
# Add to skip patterns in analyzeProjectStructure()
```

## 💡 Tips for Claude Code

### Best Practices
1. **Start each session** by running `./claude-query.sh quick`
2. **Check for errors** before making changes: `./claude-query.sh errors`
3. **Search existing code** before implementing: `./claude-query.sh search \"function name\"`
4. **Monitor git status** during development: `./claude-query.sh git`
5. **Review TODOs** for context: `./claude-query.sh todos`

### Useful Commands
```bash
# Before starting work
./claude-query.sh quick && ./claude-query.sh errors

# When looking for implementation examples  
./claude-query.sh search \"authentication\"
./claude-query.sh search \"database connection\"

# When debugging
./claude-query.sh errors
./claude-query.sh search \"error\"
./claude-query.sh processes

# When planning changes
./claude-query.sh structure
./claude-query.sh dependencies
./claude-query.sh git
```

## 🔄 Keeping It Updated

### Auto-start with System
Add to your `~/.bashrc` or `~/.zshrc`:

```bash
# Auto-start Claude Intelligence (optional)
if ! pgrep -f "claude-intelligence" > /dev/null; then
    cd /path/to/claude-intelligence
    nohup go run main.go /path/to/your/project > /dev/null 2>&1 &
fi
```

### Update Dependencies
```bash
go mod tidy
go mod download
```

## 🎉 You're Ready!

Your Claude Code now has comprehensive project intelligence! The service monitors your project 24/7 and provides Claude Code with real-time insights, dramatically improving development speed and context awareness.

**Next Steps:**
1. Start the service in your project directory
2. Open the dashboard to verify everything works
3. Test the CLI tool with `./claude-query.sh quick`
4. Tell Claude Code about the available commands
5. Enjoy much faster and more contextual development!

---

**Questions or Issues?** The service logs detailed information to help debug any problems. Check the console output where you started the Go service for troubleshooting information.