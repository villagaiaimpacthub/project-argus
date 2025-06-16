# 👁️ Project Argus

> Project Intelligence Service for Claude Code  
> All-seeing project monitoring for Claude Code

![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)

## ✨ Features

- 🔍 **Real-time Project Monitoring** - Live file system watching
- 🚨 **Intelligent Error Detection** - Multi-language error scanning  
- 📁 **Project Structure Analysis** - Automatic project type detection
- 🔄 **Git Integration** - Branch status, commits, and dirty files
- 📋 **TODO/FIXME Scanning** - Code comment analysis
- 📦 **Dependency Management** - Version tracking and updates
- ⚡ **Process Monitoring** - Running services detection
- 🌐 **REST API** - Comprehensive endpoints for all data
- 💻 **CLI Interface** - Command-line tools for quick access
- 🎨 **Web Dashboard** - Beautiful real-time visualization

## 🚀 Quick Start

### Prerequisites
- Go 1.18 or higher
- Git (for repository monitoring)
- `jq` and `curl` (for CLI tools)

### Installation

#### Option 1: One-Line Setup
```bash
curl -sSL https://raw.githubusercontent.com/villagaiaimpacthub/project-argus/main/install.sh | bash
```

#### Option 2: Manual Installation
```bash
# Clone the repository
git clone https://github.com/villagaiaimpacthub/project-argus.git
cd project-argus

# Run setup script
bash setup.sh

# Or install manually
go mod tidy
chmod +x claude-query.sh
```

### Usage

1. **Start the Service**
```bash
# Monitor current directory
go run main.go .

# Monitor specific project
go run main.go /path/to/your/project

# Use custom port
CLAUDE_INTEL_PORT=8080 go run main.go .
```

2. **Access the Dashboard**
   - Open `dashboard.html` in your browser
   - Navigate to `http://localhost:3002`

3. **Use CLI Commands**
```bash
./claude-query.sh quick       # Project overview
./claude-query.sh errors      # Active errors
./claude-query.sh git         # Git status
./claude-query.sh changes     # Recent changes
./claude-query.sh search "term"  # Search project
```

## 📖 Documentation

### Configuration

Copy `config.example.json` to `config.json` and customize:

```json
{
  "server": {
    "port": 3002,
    "host": "localhost"
  },
  "monitoring": {
    "file_watch_interval": 2,
    "git_check_interval": 5
  }
}
```

### API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /snapshot` | Complete project snapshot |
| `GET /structure` | Project file structure |
| `GET /errors` | Active errors and warnings |
| `GET /git` | Git repository status |
| `GET /changes` | Recent file changes |
| `GET /todos` | TODO items in code |
| `GET /health` | Project health metrics |
| `GET /search?q=query` | Search across project |

### Claude AI Integration

Tell Claude AI:

```
I've set up Project Argus at http://localhost:3002. 

Use these commands to understand the project:
- `./claude-query.sh quick` - Project overview
- `./claude-query.sh errors` - Current errors  
- API: GET /errors, /structure, /git, /search?q=term

Always check the current state before working on the project.
```

## 🛠️ Development

### Building from Source
```bash
go build -o project-argus main.go
./project-argus /path/to/project
```

### Running Tests
```bash
go test ./...
```

### Contributing
1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📊 Supported Languages

- Go
- JavaScript/TypeScript
- Python
- Java
- Rust
- C/C++
- And more...

## 🤝 Community

- 📝 [Documentation](https://github.com/villagaiaimpacthub/project-argus/wiki)
- 🐛 [Bug Reports](https://github.com/villagaiaimpacthub/project-argus/issues)
- 💡 [Feature Requests](https://github.com/villagaiaimpacthub/project-argus/discussions)
- 💬 [Discord Community](https://discord.gg/your-server)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Fiber](https://github.com/gofiber/fiber) - Express-inspired web framework
- Inspired by the need for AI-aware development environments
- Thanks to all contributors and the open source community

---

**⭐ Star this repository if it helps supercharge your development with AI!** 