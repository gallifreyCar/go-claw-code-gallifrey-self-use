package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/coordinator"
	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/prompt"
	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/services/api"
)

// 风格定义
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 1).
			Width(80)

	outputStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Width(80)

	toolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0)

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))
)

// 状态
type state int

const (
	stateInput state = iota
	stateProcessing
	stateToolConfirm
)

// Message 消息
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// Model TUI 模型
type Model struct {
	state      state
	input      textinput.Model
	messages   []Message
	currentOut strings.Builder
	agent      *coordinator.Agent
	provider   string
	model      string
	err        error
	width      int
	height     int

	// 工具确认
	pendingTool string
	pendingInput json.RawMessage
}

// NewModel 创建 TUI 模型
func NewModel(agent *coordinator.Agent, provider, model string) Model {
	ti := textinput.New()
	ti.Placeholder = "输入你的问题..."
	ti.Focus()
	ti.CharLimit = 2000
	ti.Width = 70

	return Model{
		state:    stateInput,
		input:    ti,
		agent:    agent,
		provider: provider,
		model:    model,
	}
}

// Init 初始化
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update 更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.state == stateInput && m.input.Value() != "" {
				return m.handleInput()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		inputStyle = inputStyle.Width(min(msg.Width-4, 80))

	case StreamTextMsg:
		m.currentOut.WriteString(msg.Text)
		return m, nil

	case StreamToolUseMsg:
		m.messages = append(m.messages, Message{
			Role:    "assistant",
			Content: m.currentOut.String() + fmt.Sprintf("\n🔧 Using tool: %s", msg.Name),
		})
		m.currentOut.Reset()
		return m, nil

	case StreamToolResultMsg:
		result := truncate(msg.Result, 100)
		if msg.IsError {
			m.messages = append(m.messages, Message{
				Role:    "system",
				Content: fmt.Sprintf("❌ Error: %s", result),
			})
		} else {
			m.messages = append(m.messages, Message{
				Role:    "system",
				Content: fmt.Sprintf("✅ Result: %s", result),
			})
		}
		return m, nil

	case ResponseCompleteMsg:
		if m.currentOut.Len() > 0 {
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: m.currentOut.String(),
			})
			m.currentOut.Reset()
		}
		m.state = stateInput
		m.input.SetValue("")
		m.input.Focus()
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		m.state = stateInput
		return m, nil
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View 渲染
func (m Model) View() string {
	var b strings.Builder

	// 标题
	title := fmt.Sprintf("🚀 Gallifrey Code (%s/%s)", m.provider, m.model)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// 消息历史
	for _, msg := range m.messages {
		var styled string
		switch msg.Role {
		case "user":
			styled = userStyle.Render("你: ") + msg.Content
		case "assistant":
			styled = assistantStyle.Render("AI: ") + msg.Content
		default:
			styled = msg.Content
		}
		b.WriteString(outputStyle.Render(styled))
		b.WriteString("\n")
	}

	// 当前输出
	if m.currentOut.Len() > 0 {
		b.WriteString(assistantStyle.Render("AI: "))
		b.WriteString(m.currentOut.String())
		b.WriteString("\n")
	}

	// 状态
	switch m.state {
	case stateProcessing:
		b.WriteString(outputStyle.Render("⏳ 处理中..."))
		b.WriteString("\n")
	case stateInput:
		b.WriteString(userStyle.Render("你: "))
		b.WriteString(inputStyle.Render(m.input.View()))
		b.WriteString("\n")
	}

	// 错误
	if m.err != nil {
		b.WriteString(errorStyle.Render("❌ 错误: "+m.err.Error()))
		b.WriteString("\n")
	}

	// 帮助
	b.WriteString(helpStyle.Render("Enter 发送 • Ctrl+C 退出"))

	return b.String()
}

func (m Model) handleInput() (tea.Model, tea.Cmd) {
	input := m.input.Value()
	m.messages = append(m.messages, Message{
		Role:    "user",
		Content: input,
	})
	m.state = stateProcessing
	m.input.Blur()

	return m, m.runAgent(input)
}

func (m Model) runAgent(input string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// 设置回调
		var mu sync.Mutex
		textChan := make(chan string, 100)
		done := make(chan struct{})

		m.agent.OnText(func(text string) {
			select {
			case textChan <- text:
			default:
			}
		})

		m.agent.OnToolUse(func(name, id string, input json.RawMessage) {
			mu.Lock()
			m.pendingTool = name
			m.pendingInput = input
			mu.Unlock()
		})

		m.agent.OnToolResult(func(id string, result string, isError bool) {
			mu.Lock()
			_ = id // 暂时不使用
			mu.Unlock()
		})

		// 构建系统提示
		pm := prompt.NewManager().SetDefault(prompt.DefaultPrompt())
		systemPrompt := pm.Build()

		// 启动 goroutine 处理流
		go func() {
			defer close(done)
			if err := m.agent.Run(ctx, systemPrompt, input); err != nil {
				// 错误处理
			}
		}()

		// 等待完成
		<-done

		return ResponseCompleteMsg{}
	}
}

// StreamTextMsg 流式文本消息
type StreamTextMsg struct {
	Text string
}

// StreamToolUseMsg 工具使用消息
type StreamToolUseMsg struct {
	Name string
	ID   string
	Input json.RawMessage
}

// StreamToolResultMsg 工具结果消息
type StreamToolResultMsg struct {
	ID      string
	Result  string
	IsError bool
}

// ResponseCompleteMsg 响应完成消息
type ResponseCompleteMsg struct{}

// ErrorMsg 错误消息
type ErrorMsg struct {
	Err error
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SetAgent 设置 Agent
func (m *Model) SetAgent(agent *coordinator.Agent) {
	m.agent = agent
}

// SetAPIInfo 设置 API 信息
func (m *Model) SetAPIInfo(provider, model string) {
	m.provider = provider
	m.model = model
}

// Ensure Model implements tea.Model
var _ tea.Model = (*Model)(nil)

// Ensure API types are used
var _ api.Client = (api.Client)(nil)
