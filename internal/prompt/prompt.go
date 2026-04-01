package prompt

import (
	"strings"
)

// Manager System Prompt 管理器
type Manager struct {
	// 优先级从高到低
	override    string // Priority 0: 强制覆盖
	coordinator string // Priority 1: 协调模式
	agent       string // Priority 2: 子 Agent
	custom      string // Priority 3: 用户自定义
	defaultP    string // Priority 4: 默认
	append      string // 始终追加

	// 缓存边界
	dynamicBoundary string
}

// NewManager 创建 Prompt 管理器
func NewManager() *Manager {
	return &Manager{
		dynamicBoundary: "__SYSTEM_PROMPT_DYNAMIC_BOUNDARY__",
	}
}

// SetOverride 设置覆盖 Prompt
func (m *Manager) SetOverride(prompt string) *Manager {
	m.override = prompt
	return m
}

// SetCoordinator 设置协调模式 Prompt
func (m *Manager) SetCoordinator(prompt string) *Manager {
	m.coordinator = prompt
	return m
}

// SetAgent 设置子 Agent Prompt
func (m *Manager) SetAgent(prompt string) *Manager {
	m.agent = prompt
	return m
}

// SetCustom 设置自定义 Prompt
func (m *Manager) SetCustom(prompt string) *Manager {
	m.custom = prompt
	return m
}

// SetDefault 设置默认 Prompt
func (m *Manager) SetDefault(prompt string) *Manager {
	m.defaultP = prompt
	return m
}

// Append 追加 Prompt
func (m *Manager) Append(prompt string) *Manager {
	m.append = prompt
	return m
}

// Build 构建 System Prompt
func (m *Manager) Build() string {
	var prompt string

	// 按优先级选择
	switch {
	case m.override != "":
		prompt = m.override
	case m.coordinator != "":
		prompt = m.coordinator
	case m.agent != "":
		prompt = m.agent
	case m.custom != "":
		prompt = m.custom
	default:
		prompt = m.defaultP
	}

	// 追加
	if m.append != "" {
		prompt += "\n\n" + m.append
	}

	return prompt
}

// BuildWithBoundary 构建带缓存边界的 System Prompt
func (m *Manager) BuildWithBoundary(dynamicContent string) string {
	staticPart := m.Build()
	if dynamicContent == "" {
		return staticPart
	}

	return staticPart + "\n\n" + m.dynamicBoundary + "\n\n" + dynamicContent
}

// SplitByBoundary 按边界分割 Prompt
func (m *Manager) SplitByBoundary(fullPrompt string) (static, dynamic string) {
	parts := strings.SplitN(fullPrompt, m.dynamicBoundary, 2)
	static = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		dynamic = strings.TrimSpace(parts[1])
	}
	return
}

// DefaultPrompt 返回默认 System Prompt
func DefaultPrompt() string {
	return `You are an interactive agent that helps users with software engineering tasks.

## System Rules

- Use dedicated tools (Read/Edit/Glob/Grep) instead of Bash for file operations
- Always read files before modifying them
- Don't add features or refactoring unless explicitly asked
- Avoid OWASP Top 10 vulnerabilities

## Task Execution

1. Understand the request before acting
2. Use appropriate tools for each operation
3. Verify changes before completing
4. Keep responses concise and focused

## Output Format

- No emojis unless explicitly requested
- Use file_path:line_number format for code references
- If you can say it in one sentence, don't use three`
}
