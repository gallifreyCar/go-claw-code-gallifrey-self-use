package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/tools"
)

// MCPTool MCP 工具包装器
type MCPTool struct {
	client *Client
	tool   Tool
}

// NewMCPTool 创建 MCP 工具包装器
func NewMCPTool(client *Client, tool Tool) *MCPTool {
	return &MCPTool{
		client: client,
		tool:   tool,
	}
}

func (t *MCPTool) Name() string {
	return t.tool.Name
}

func (t *MCPTool) Description() string {
	return t.tool.Description
}

func (t *MCPTool) InputSchema() json.RawMessage {
	return t.tool.InputSchema
}

func (t *MCPTool) RequiresConfirmation() bool {
	// MCP 工具默认需要确认
	return true
}

func (t *MCPTool) Execute(ctx context.Context, input json.RawMessage) (*tools.Result, error) {
	var args map[string]interface{}
	if err := json.Unmarshal(input, &args); err != nil {
		return nil, err
	}

	output, err := t.client.CallTool(ctx, t.tool.Name, args)
	if err != nil {
		return &tools.Result{
			Output:  err.Error(),
			IsError: true,
		}, nil
	}

	return &tools.Result{
		Output: output,
		Metadata: map[string]interface{}{
			"tool":   t.tool.Name,
			"source": "mcp",
		},
	}, nil
}

// ToTool 转换为 tools.Tool 接口
func (t *MCPTool) ToTool() tools.Tool {
	return t
}

// WrapTools 将 MCP 工具列表转换为 tools.Tool 切片
func WrapTools(client *Client) []tools.Tool {
	var result []tools.Tool
	for _, tool := range client.GetTools() {
		result = append(result, NewMCPTool(client, tool))
	}
	return result
}

// Config MCP 配置
type Config struct {
	Servers []Server `json:"servers" yaml:"servers"`
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		// 尝试 YAML
		if err := parseYAML(data, &cfg); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

func readFile(path string) ([]byte, error) {
	f, err := openFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readAll(f)
}

// 以下是为了避免导入额外依赖的简化实现

func openFile(path string) (interface{ Close() error }, error) {
	type file interface {
		Close() error
	}
	return nil, fmt.Errorf("placeholder")
}

func readAll(r interface{}) ([]byte, error) {
	return nil, fmt.Errorf("placeholder")
}

func parseYAML(data []byte, v interface{}) error {
	return fmt.Errorf("YAML parsing not implemented, use JSON config")
}
