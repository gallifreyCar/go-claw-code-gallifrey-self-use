# Gallifrey Code

English | [简体中文](./README.md)

> AI Coding Assistant in Go (Learning Project)

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> ⚠️ **Disclaimer**: This project is for learning and research purposes only. The project was inspired by public AI coding assistant architecture analysis. All code is original implementation and does not contain any third-party intellectual property content. Do not use this project for commercial purposes.

## About

Gallifrey Code is an AI-powered coding assistant CLI tool implemented in Go, designed to explore and learn AI Agent architecture. The project demonstrates how to build an AI application with tool calling, streaming responses, and multi-agent coordination.

## Features

- 🚀 **Single Binary Deployment** - Pure Go implementation, no runtime required
- 💬 **Dual Output Modes** - TUI interface and `--print` headless mode
- 🔌 **Multi-API Support** - Anthropic API and OpenAI-compatible APIs
- 🛠️ **Rich Toolset** - Bash, file operations, search and more built-in tools
- 🤖 **Multi-Agent Coordination** - Parallel worker agent orchestration
- 📝 **Smart Memory** - Automatic memory integration system

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                             │
│  (cobra CLI, TUI with bubbletea, --print mode)              │
├─────────────────────────────────────────────────────────────┤
│                     Coordinator Layer                        │
│  (Single Agent / Multi-Agent orchestration)                  │
├─────────────────────────────────────────────────────────────┤
│                       Tool Layer                             │
│  (Bash, Read, Edit, Write, Glob, Grep, MCP, Agent, Task)    │
├─────────────────────────────────────────────────────────────┤
│                      Service Layer                           │
│  (API Client, Auth, Memory, Permission, Telemetry)          │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                      │
│  (Config, Logger, Context, State Manager)                    │
└─────────────────────────────────────────────────────────────┘
```

## Installation

### Build from Source

```bash
# Clone repository
git clone https://github.com/gallifreyCar/go-claw-code-gallifrey-self-use.git
cd go-claw-code-gallifrey-self-use

# Build
make build

# Install to $GOPATH/bin
make install
```

### Using go install

```bash
go install github.com/gallifreyCar/go-claw-code-gallifrey-self-use/cmd/gallifrey-code@latest
```

## Quick Start

### Configuration

Create config file at `~/.config/gallifrey-code/config.yaml`:

```yaml
# API Configuration
api:
  provider: anthropic  # anthropic or openai
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-sonnet-4-6
  openai:
    api_key: ${OPENAI_API_KEY}
    base_url: https://api.openai.com/v1  # Customizable
    model: gpt-4

# Permission mode
permission: ask  # ask, allow, deny

# Log level
log_level: info
```

Or use environment variables:

```bash
export ANTHROPIC_API_KEY=your-api-key
export OPENAI_API_KEY=your-openai-key
```

### Usage

```bash
# TUI interactive mode
gallifrey-code

# Headless mode (single query)
gallifrey-code -p "Help me write a hello world program"

# Specify model
gallifrey-code -p "Explain this code" --model claude-opus-4-6

# Use OpenAI-compatible API
gallifrey-code -p "Hello" --provider openai

# Show help
gallifrey-code --help
```

## Built-in Tools

| Tool | Description |
|------|-------------|
| `Bash` | Execute shell commands |
| `Read` | Read file contents |
| `Edit` | Edit files (string replacement) |
| `Write` | Write files |
| `Glob` | File pattern search |
| `Grep` | File content search |
| `Agent` | Create sub-agents |
| `Task` | Task management |
| `Memory` | Memory management |

## Project Structure

```
go-claw-code-gallifrey-self-use/
├── cmd/gallifrey-code/      # CLI entry point
├── internal/
│   ├── cli/                 # CLI command definitions
│   ├── tui/                 # TUI interface (bubbletea)
│   ├── coordinator/         # Agent coordinator
│   ├── tools/               # Tool implementations
│   ├── services/            # Service layer
│   │   ├── api/             # API clients
│   │   ├── auth/            # Authentication
│   │   ├── memory/          # Memory system
│   │   └── permission/      # Permission management
│   ├── prompt/              # System Prompt management
│   └── config/              # Configuration management
├── pkg/                     # Exportable packages
│   ├── types/               # Public types
│   └── utils/               # Utility functions
├── configs/                 # Config examples
├── go.mod
├── go.sum
└── Makefile
```

## Development

```bash
# Install dependencies
make deps

# Run tests
make test

# Code linting
make lint

# Build
make build
```

## Learning Value

This project is suitable for the following learning scenarios:

1. **AI Agent Architecture Design** - Understand how to design a complete AI Agent system
2. **Go CLI Development** - Learn cobra, bubbletea and other libraries
3. **Streaming API Processing** - Master streaming response handling
4. **Tool System Design** - Understand extensible tool interface design
5. **Multi-Agent Coordination** - Explore parallel agent orchestration

## Roadmap

- [x] Project initialization
- [x] Core framework (CLI, config, API clients)
- [x] Tool system (Bash, Read, Write, Edit, Glob, Grep, Memory)
- [x] Agent core loop
- [x] TUI interface
- [x] Memory system
- [x] Multi-agent coordination
- [x] MCP support

## License

MIT License

## Disclaimer

This project is a personal learning project for educational and research purposes only. The author is not responsible for any consequences arising from the use of this project. When using this project, please comply with relevant laws and regulations and terms of service.
