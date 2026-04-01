package tools

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

// Bash Bash 工具
type Bash struct {
	timeout time.Duration
}

// BashInput Bash 输入参数
type BashInput struct {
	Command   string `json:"command"`
	Timeout   int    `json:"timeout,omitempty"`
	Sandbox   bool   `json:"sandbox,omitempty"`
	Threshold int    `json:"threshold,omitempty"`
}

// NewBash 创建 Bash 工具
func NewBash() *Bash {
	return &Bash{
		timeout: 120 * time.Second,
	}
}

func (t *Bash) Name() string {
	return "Bash"
}

func (t *Bash) Description() string {
	return `Execute a bash command in a persistent shell session.

Usage notes:
- Use this tool to execute shell commands
- Commands run in the user's current working directory
- Returns command output and exit code`
}

func (t *Bash) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {
				"type": "string",
				"description": "The command to execute"
			},
			"timeout": {
				"type": "integer",
				"description": "Timeout in milliseconds (default 120000)"
			}
		},
		"required": ["command"]
	}`)
}

func (t *Bash) RequiresConfirmation() bool {
	return true // 危险操作需要确认
}

func (t *Bash) Execute(ctx context.Context, input json.RawMessage) (*Result, error) {
	var params BashInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	timeout := t.timeout
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", params.Command)
	output, err := cmd.CombinedOutput()

	result := &Result{
		Output:  string(output),
		IsError: err != nil,
		Metadata: map[string]interface{}{
			"command": params.Command,
		},
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.Output = "Command timed out after " + timeout.String()
		result.IsError = true
	}

	return result, nil
}

// 安全命令检查
func isSafeCommand(cmd string) bool {
	safeCommands := []string{
		"ls", "cat", "head", "tail", "grep", "find", "pwd",
		"git status", "git log", "git diff", "git branch",
		"npm list", "go list", "go mod",
	}

	cmdLower := strings.ToLower(strings.TrimSpace(cmd))
	for _, safe := range safeCommands {
		if strings.HasPrefix(cmdLower, safe) {
			return true
		}
	}
	return false
}