package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
			Foreground(lipgloss.Color("220")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0)
)

// 状态
type state int

const (
	stateInput state = iota
	stateProcessing
	stateToolConfirm
	stateDone
)

// Model TUI 模型
type Model struct {
	state       state
	input       textinput.Model
	output      strings.Builder
	messages    []string
	toolPending string
	err         error
	width       int
	height      int
}

// NewModel 创建 TUI 模型
func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "输入你的问题..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 70

	return Model{
		state:  stateInput,
		input:  ti,
		output: strings.Builder{},
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
				m.state = stateProcessing
				m.messages = append(m.messages, "用户: "+m.input.Value())
				return m, m.processInput()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		inputStyle = inputStyle.Width(min(msg.Width-4, 80))

	case ResponseMsg:
		m.output.WriteString(msg.Content)
		m.state = stateInput
		m.input.SetValue("")
		m.input.Focus()

	case ErrorMsg:
		m.err = msg.Err
		m.state = stateInput
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View 渲染
func (m Model) View() string {
	var b strings.Builder

	// 标题
	b.WriteString(titleStyle.Render("🚀 Go-Claw-Code"))
	b.WriteString("\n\n")

	// 消息历史
	for _, msg := range m.messages {
		b.WriteString(outputStyle.Render(msg))
		b.WriteString("\n")
	}

	// 状态
	switch m.state {
	case stateProcessing:
		b.WriteString(outputStyle.Render("⏳ 处理中..."))
		b.WriteString("\n")
	case stateInput:
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

func (m Model) processInput() tea.Cmd {
	return func() tea.Msg {
		// 这里会被实际的 Agent 调用替换
		return ResponseMsg{Content: "🤖 正在处理: " + m.input.Value()}
	}
}

// ResponseMsg 响应消息
type ResponseMsg struct {
	Content string
}

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
