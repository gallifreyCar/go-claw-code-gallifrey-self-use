# Go-Claw-Code

> Go 语言实现的 Claude Code 风格 AI 编程助手

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

基于 [Claude Code](https://github.com/anthropics/claude-code) 泄漏源码架构分析，用 Go 语言实现的 AI 编程助手 CLI 工具。

## 特性

- 🚀 **单二进制部署** - 纯 Go 实现，无需运行时环境
- 💬 **双模式输出** - 同时支持 TUI 界面和 `--print` 无头模式
- 🔌 **多 API 支持** - 支持 Anthropic API 和 OpenAI 兼容 API
- 🛠️ **丰富工具集** - Bash、文件操作、搜索等内置工具
- 🤖 **多 Agent 协调** - 支持并行 Worker Agent 编排
- 📝 **智能记忆** - 自动记忆整合系统

## 架构

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

## 安装

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/gallifreyCar/go-claw-code-gallifrey-self-use.git
cd go-claw-code-gallifrey-self-use

# 构建
make build

# 安装到 $GOPATH/bin
make install
```

### 使用 go install

```bash
go install github.com/gallifreyCar/go-claw-code-gallifrey-self-use/cmd/go-claw-code@latest
```

## 快速开始

### 配置

创建配置文件 `~/.config/go-claw-code/config.yaml`:

```yaml
# API 配置
api:
  provider: anthropic  # anthropic 或 openai
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-sonnet-4-6
  openai:
    api_key: ${OPENAI_API_KEY}
    base_url: https://api.openai.com/v1  # 可自定义
    model: gpt-4

# 权限模式
permission: ask  # ask, allow, deny

# 日志级别
log_level: info
```

或使用环境变量:

```bash
export ANTHROPIC_API_KEY=your-api-key
export OPENAI_API_KEY=your-openai-key
```

### 使用

```bash
# TUI 交互模式
go-claw-code

# 无头模式（单次问答）
go-claw-code -p "帮我写一个 hello world 程序"

# 指定模型
go-claw-code -p "解释这段代码" --model claude-opus-4-6

# 使用 OpenAI 兼容 API
go-claw-code -p "你好" --provider openai

# 查看帮助
go-claw-code --help
```

## 内置工具

| 工具 | 说明 |
|------|------|
| `Bash` | 执行 shell 命令 |
| `Read` | 读取文件内容 |
| `Edit` | 编辑文件（字符串替换） |
| `Write` | 写入文件 |
| `Glob` | 文件模式搜索 |
| `Grep` | 文件内容搜索 |
| `Agent` | 创建子 Agent |
| `Task` | 任务管理 |

## 项目结构

```
go-claw-code-gallifrey-self-use/
├── cmd/go-claw-code/        # CLI 入口
├── internal/
│   ├── cli/                 # CLI 命令定义
│   ├── tui/                 # TUI 界面 (bubbletea)
│   ├── coordinator/         # Agent 协调器
│   ├── tools/               # 工具实现
│   ├── services/            # 服务层
│   │   ├── api/             # API 客户端
│   │   ├── auth/            # 认证
│   │   ├── memory/          # 记忆系统
│   │   └── permission/      # 权限管理
│   ├── prompt/              # System Prompt 管理
│   └── config/              # 配置管理
├── pkg/                     # 可导出的包
│   ├── types/               # 公共类型
│   └── utils/               # 工具函数
├── configs/                 # 配置文件示例
├── go.mod
├── go.sum
└── Makefile
```

## 开发

```bash
# 安装依赖
make deps

# 运行测试
make test

# 代码检查
make lint

# 构建
make build
```

## 路线图

- [x] 项目初始化
- [x] 核心框架（CLI、配置、API 客户端）
- [x] 工具系统（Bash, Read, Write, Edit, Glob, Grep）
- [x] Agent 核心循环
- [x] TUI 界面
- [ ] 多 Agent 协调
- [ ] Memory 系统
- [ ] MCP 支持

## 致谢

本项目架构设计参考了 2026年3月31日泄漏的 Claude Code v2.1.88 源码分析。

## License

MIT License