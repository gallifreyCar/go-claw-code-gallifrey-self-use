package api

import (
	"context"
	"encoding/json"
)

// MessageRole 消息角色
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// ContentBlock 内容块类型
type ContentBlockType string

const (
	ContentTypeText       ContentBlockType = "text"
	ContentTypeToolUse    ContentBlockType = "tool_use"
	ContentTypeToolResult ContentBlockType = "tool_result"
)

// ContentBlock 内容块
type ContentBlock struct {
	Type    ContentBlockType `json:"type"`
	Text    string           `json:"text,omitempty"`
	ID      string           `json:"id,omitempty"`
	Name    string           `json:"name,omitempty"`
	Input   json.RawMessage  `json:"input,omitempty"`
	ToolID  string           `json:"tool_use_id,omitempty"`
	Content string           `json:"content,omitempty"`
	IsError bool             `json:"is_error,omitempty"`
}

// Message 消息
type Message struct {
	Role    MessageRole    `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ToolDefinition 工具定义
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// MessageRequest 消息请求
type MessageRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Messages  []Message       `json:"messages"`
	Tools     []ToolDefinition `json:"tools,omitempty"`
	Stream    bool            `json:"stream"`
}

// StreamEvent 流式事件类型
type StreamEventType string

const (
	EventMessageStart    StreamEventType = "message_start"
	EventContentBlockStart StreamEventType = "content_block_start"
	EventContentBlockDelta StreamEventType = "content_block_delta"
	EventContentBlockStop  StreamEventType = "content_block_stop"
	EventMessageDelta      StreamEventType = "message_delta"
	EventMessageStop       StreamEventType = "message_stop"
	EventError             StreamEventType = "error"
)

// StreamEvent 流式事件
type StreamEvent struct {
	Type  StreamEventType     `json:"type"`
	Index int                 `json:"index,omitempty"`
	Delta *StreamDelta        `json:"delta,omitempty"`
	Message *Message          `json:"message,omitempty"`
	Error  *ErrorInfo         `json:"error,omitempty"`
}

// StreamDelta 流式增量
type StreamDelta struct {
	Type    string          `json:"type"`
	Text    string          `json:"text,omitempty"`
	Partial json.RawMessage `json:"partial_json,omitempty"`
	Stop    string          `json:"stop_reason,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Usage 使用量
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Role    MessageRole    `json:"role"`
	Content []ContentBlock `json:"content"`
	Model   string         `json:"model"`
	Usage   Usage          `json:"usage"`
	Stop    string         `json:"stop_reason,omitempty"`
}

// Client API 客户端接口
type Client interface {
	// CreateMessage 创建消息（流式）
	CreateMessage(ctx context.Context, req *MessageRequest) (<-chan StreamEvent, error)

	// CreateMessageSync 创建消息（同步）
	CreateMessageSync(ctx context.Context, req *MessageRequest) (*MessageResponse, error)

	// GetProvider 获取提供商名称
	GetProvider() string
}