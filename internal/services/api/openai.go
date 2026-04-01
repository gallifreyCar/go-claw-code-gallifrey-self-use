package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

// OpenAIClient OpenAI 兼容 API 客户端
type OpenAIClient struct {
	config     OpenAIConfig
	httpClient *http.Client
}

// NewOpenAIClient 创建 OpenAI 客户端
func NewOpenAIClient(cfg OpenAIConfig) *OpenAIClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4"
	}
	return &OpenAIClient{
		config:     cfg,
		httpClient: &http.Client{},
	}
}

func (c *OpenAIClient) GetProvider() string {
	return "openai"
}

// OpenAI 请求类型
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Tools       []openAITool    `json:"tools,omitempty"`
	Stream      bool            `json:"stream"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role    string          `json:"role"`
	Content interface{}     `json:"content"`
	Name    string          `json:"name,omitempty"`
	ToolID  string          `json:"tool_call_id,omitempty"`
}

type openAITool struct {
	Type     string          `json:"type"`
	Function openAIFunction  `json:"function"`
}

type openAIFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// OpenAI 响应类型
type openAIResponse struct {
	ID      string         `json:"id"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
}

type openAIChoice struct {
	Index        int            `json:"index"`
	Message      openAIMessage  `json:"message"`
	Delta        openAIDelta    `json:"delta"`
	FinishReason string         `json:"finish_reason"`
}

type openAIDelta struct {
	Role      string              `json:"role,omitempty"`
	Content   string              `json:"content,omitempty"`
	ToolCalls []openAIToolCallDelta `json:"tool_calls,omitempty"`
}

type openAIToolCallDelta struct {
	Index int                 `json:"index"`
	ID    string              `json:"id,omitempty"`
	Type  string              `json:"type,omitempty"`
	Function openAIFunctionDelta `json:"function,omitempty"`
}

type openAIFunctionDelta struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CreateMessage 创建消息（流式）
func (c *OpenAIClient) CreateMessage(ctx context.Context, req *MessageRequest) (<-chan StreamEvent, error) {
	openAIReq := c.convertRequest(req)
	openAIReq.Stream = true

	body, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	eventChan := make(chan StreamEvent, 100)

	go func() {
		defer close(eventChan)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				eventChan <- StreamEvent{Type: EventMessageStop}
				break
			}

			var openAIResp openAIResponse
			if err := json.Unmarshal([]byte(data), &openAIResp); err != nil {
				continue
			}

			events := c.convertStreamResponse(&openAIResp)
			for _, event := range events {
				eventChan <- event
			}
		}
	}()

	return eventChan, nil
}

// CreateMessageSync 创建消息（同步）
func (c *OpenAIClient) CreateMessageSync(ctx context.Context, req *MessageRequest) (*MessageResponse, error) {
	openAIReq := c.convertRequest(req)
	openAIReq.Stream = false

	body, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return c.convertResponse(&openAIResp), nil
}

func (c *OpenAIClient) convertRequest(req *MessageRequest) *openAIRequest {
	messages := make([]openAIMessage, 0, len(req.Messages)+1)

	// System prompt
	if req.System != "" {
		messages = append(messages, openAIMessage{
			Role:    "system",
			Content: req.System,
		})
	}

	// 消息
	for _, msg := range req.Messages {
		content := ""
		for _, block := range msg.Content {
			if block.Type == ContentTypeText {
				content += block.Text
			}
		}
		messages = append(messages, openAIMessage{
			Role:    string(msg.Role),
			Content: content,
		})
	}

	// 工具
	tools := make([]openAITool, 0, len(req.Tools))
	for _, t := range req.Tools {
		tools = append(tools, openAITool{
			Type: "function",
			Function: openAIFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}

	return &openAIRequest{
		Model:     c.config.Model,
		Messages:  messages,
		Tools:     tools,
		MaxTokens: req.MaxTokens,
	}
}

func (c *OpenAIClient) convertResponse(resp *openAIResponse) *MessageResponse {
	if len(resp.Choices) == 0 {
		return &MessageResponse{}
	}

	choice := resp.Choices[0]
	content := []ContentBlock{}

	if msg, ok := choice.Message.Content.(string); ok && msg != "" {
		content = append(content, ContentBlock{
			Type: ContentTypeText,
			Text: msg,
		})
	}

	return &MessageResponse{
		ID:      resp.ID,
		Role:    RoleAssistant,
		Content: content,
		Model:   c.config.Model,
		Usage: Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
		Stop: choice.FinishReason,
	}
}

func (c *OpenAIClient) convertStreamResponse(resp *openAIResponse) []StreamEvent {
	if len(resp.Choices) == 0 {
		return nil
	}

	choice := resp.Choices[0]
	var events []StreamEvent

	if choice.Delta.Content != "" {
		events = append(events, StreamEvent{
			Type:  EventContentBlockDelta,
			Index: 0,
			Delta: &StreamDelta{
				Type: "text_delta",
				Text: choice.Delta.Content,
			},
		})
	}

	if choice.FinishReason != "" {
		events = append(events, StreamEvent{
			Type: EventMessageStop,
		})
	}

	return events
}