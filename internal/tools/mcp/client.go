package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Server MCP 服务器配置
type Server struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// Tool MCP 工具定义
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// Client MCP 客户端
type Client struct {
	server  Server
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  *bufio.Reader
	mu      sync.Mutex
	nextID  int
	tools   []Tool
}

// NewClient 创建 MCP 客户端
func NewClient(server Server) *Client {
	return &Client{
		server: server,
		nextID: 1,
	}
}

// Start 启动 MCP 服务器
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cmd := exec.CommandContext(ctx, c.server.Command, c.server.Args...)

	// 设置环境变量
	cmd.Env = os.Environ()
	for k, v := range c.server.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	c.cmd = cmd
	c.stdin = stdin
	c.stdout = bufio.NewReader(stdout)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	// 初始化连接
	if err := c.initialize(); err != nil {
		return fmt.Errorf("failed to initialize MCP: %w", err)
	}

	return nil
}

// Stop 停止 MCP 服务器
func (c *Client) Stop() error {
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

// Request 发送请求
func (c *Client) Request(method string, params interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID
	c.nextID++

	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		req["params"] = params
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 发送请求
	if _, err := fmt.Fprintf(c.stdin, "%s\n", data); err != nil {
		return nil, err
	}

	// 读取响应
	line, err := c.stdout.ReadString('\n')
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", resp.Error.Message)
	}

	return resp.Result, nil
}

// initialize 初始化 MCP 连接
func (c *Client) initialize() error {
	result, err := c.Request("initialize", map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"clientInfo": map[string]string{
			"name":    "go-claw-code",
			"version": "1.0.0",
		},
	})
	if err != nil {
		return err
	}

	// 解析工具列表
	var initResult struct {
		Capabilities struct {
			Tools struct {
				ListChanged bool `json:"listChanged"`
			} `json:"tools"`
		} `json:"capabilities"`
	}

	_ = json.Unmarshal(result, &initResult)

	// 获取工具列表
	return c.listTools()
}

// listTools 获取工具列表
func (c *Client) listTools() error {
	result, err := c.Request("tools/list", nil)
	if err != nil {
		return err
	}

	var listResult struct {
		Tools []Tool `json:"tools"`
	}

	if err := json.Unmarshal(result, &listResult); err != nil {
		return err
	}

	c.tools = listResult.Tools
	return nil
}

// CallTool 调用工具
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	result, err := c.Request("tools/call", map[string]interface{}{
		"name":      name,
		"arguments": args,
	})
	if err != nil {
		return "", err
	}

	var callResult struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}

	if err := json.Unmarshal(result, &callResult); err != nil {
		return "", err
	}

	var texts []string
	for _, c := range callResult.Content {
		if c.Type == "text" {
			texts = append(texts, c.Text)
		}
	}

	return strings.Join(texts, "\n"), nil
}

// GetTools 获取工具列表
func (c *Client) GetTools() []Tool {
	return c.tools
}

// Manager MCP 管理器
type Manager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

// NewManager 创建 MCP 管理器
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
	}
}

// AddServer 添加 MCP 服务器
func (m *Manager) AddServer(ctx context.Context, server Server) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client := NewClient(server)
	if err := client.Start(ctx); err != nil {
		return err
	}

	m.clients[server.Name] = client
	return nil
}

// RemoveServer 移除 MCP 服务器
func (m *Manager) RemoveServer(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.clients[name]; ok {
		if err := client.Stop(); err != nil {
			return err
		}
		delete(m.clients, name)
	}

	return nil
}

// GetClient 获取 MCP 客户端
func (m *Manager) GetClient(name string) *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[name]
}

// GetAllTools 获取所有 MCP 工具
func (m *Manager) GetAllTools() map[string][]Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]Tool)
	for name, client := range m.clients {
		result[name] = client.GetTools()
	}
	return result
}
