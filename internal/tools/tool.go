package tools

import (
	"context"
	"encoding/json"
)

// Tool 工具接口
type Tool interface {
	// Name 工具名称
	Name() string

	// Description 工具描述
	Description() string

	// InputSchema 输入参数 JSON Schema
	InputSchema() json.RawMessage

	// Execute 执行工具
	Execute(ctx context.Context, input json.RawMessage) (*Result, error)

	// RequiresConfirmation 是否需要用户确认
	RequiresConfirmation() bool
}

// Result 工具执行结果
type Result struct {
	Output   string                 `json:"output"`
	IsError  bool                   `json:"is_error"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Definition 工具定义（用于 API）
type Definition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// ToDefinition 转换为 API 定义
func ToDefinition(t Tool) Definition {
	return Definition{
		Name:        t.Name(),
		Description: t.Description(),
		InputSchema: t.InputSchema(),
	}
}

// Registry 工具注册表
type Registry struct {
	tools map[string]Tool
}

// NewRegistry 创建工具注册表
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get 获取工具
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List 列出所有工具
func (r *Registry) List() []Tool {
	result := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

// Definitions 获取所有工具定义
func (r *Registry) Definitions() []Definition {
	result := make([]Definition, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, ToDefinition(t))
	}
	return result
}

// Execute 执行工具
func (r *Registry) Execute(ctx context.Context, name string, input json.RawMessage) (*Result, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, &ToolNotFoundError{Name: name}
	}
	return tool.Execute(ctx, input)
}

// ToolNotFoundError 工具未找到错误
type ToolNotFoundError struct {
	Name string
}

func (e *ToolNotFoundError) Error() string {
	return "tool not found: " + e.Name
}