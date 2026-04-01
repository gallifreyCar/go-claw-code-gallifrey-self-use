package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/services/api"
	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/tools"
)

// Agent AI Agent
type Agent struct {
	client    api.Client
	tools     map[string]tools.Tool
	model     string
	maxIter   int
	messages  []api.Message

	// 回调
	onText       func(text string)
	onToolUse    func(name, id string, input json.RawMessage)
	onToolResult func(id string, result string, isError bool)
	onComplete  func(message *api.MessageResponse)
}

// NewAgent 创建 Agent
func NewAgent(client api.Client, model string) *Agent {
	return &Agent{
		client:  client,
		tools:     make(map[string]tools.Tool),
		model:     model,
		maxIter:   50,
		messages:  make([]api.Message, 0),
	}
}

// RegisterTool 注册工具
func (a *Agent) RegisterTool(tool tools.Tool) {
	a.tools[tool.Name()] = tool
}

// OnText 设置文本回调
func (a *Agent) OnText(fn func(text string)) {
	a.onText = fn
}

// OnToolUse 设置工具调用回调
func (a *Agent) OnToolUse(fn func(name, id string, input json.RawMessage)) {
	a.onToolUse = fn
}

// OnToolResult 设置工具结果回调
func (a *Agent) OnToolResult(fn func(id string, result string, isError bool)) {
	a.onToolResult = fn
}

// OnComplete 设置完成回调
func (a *Agent) OnComplete(fn func(message *api.MessageResponse)) {
	a.onComplete = fn
}

// Run 运行 Agent
func (a *Agent) Run(ctx context.Context, systemPrompt, input string) error {
	// 添加用户消息
	a.messages = append(a.messages, api.Message{
		Role: api.RoleUser,
		Content: []api.ContentBlock{
			{Type: api.ContentTypeText, Text: input},
		},
	})

	for i := 0; i < a.maxIter; i++ {
		// 构建 Tool 定义
		toolDefs := a.getToolDefinitions()

		// 创建请求
		req := &api.MessageRequest{
			Model:     a.model,
			MaxTokens: 4096,
			System:    systemPrompt,
			Messages:  a.messages,
			Tools:     toolDefs,
		}

		// 调用 API
		stream, err := a.client.CreateMessage(ctx, req)
		if err != nil {
			return fmt.Errorf("API call failed: %w", err)
		}

		// 处理流式响应
		message, toolCalls, err := a.processStream(stream)
		if err != nil {
			return err
		}

		// 如果没有 tool calls，完成
		if len(toolCalls) == 0 {
			if a.onComplete != nil && message != nil {
				a.onComplete(message)
			}
			return nil
		}

		// 执行工具调用
		toolResults := a.executeTools(ctx, toolCalls)

		// 添加助手消息
		a.messages = append(a.messages, api.Message{
			Role:    api.RoleAssistant,
			Content: toolCallsToContent(toolCalls),
		})

		// 添加工具结果消息
		a.messages = append(a.messages, api.Message{
			Role:    api.RoleUser,
			Content: toolResultsToContent(toolResults),
		})
	}

	return fmt.Errorf("max iterations (%d) reached", a.maxIter)
}

func (a *Agent) getToolDefinitions() []api.ToolDefinition {
	defs := make([]api.ToolDefinition, 0, len(a.tools))
	for _, t := range a.tools {
		defs = append(defs, api.ToolDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.InputSchema(),
		})
	}
	return defs
}

func (a *Agent) processStream(stream <-chan api.StreamEvent) (*api.MessageResponse, []api.ContentBlock, error) {
	var textBuilder strings.Builder
	var toolCalls []api.ContentBlock
	var currentToolCall *api.ContentBlock
	var inputBuilder strings.Builder

	for event := range stream {
		switch event.Type {
		case api.EventContentBlockStart:
			if event.Message != nil && len(event.Message.Content) > 0 {
				block := event.Message.Content[0]
				if block.Type == api.ContentTypeToolUse {
					currentToolCall = &block
					inputBuilder.Reset()
				}
			}

		case api.EventContentBlockDelta:
			if event.Delta != nil {
				if event.Delta.Type == "text_delta" {
					text := event.Delta.Text
					textBuilder.WriteString(text)
					if a.onText != nil {
						a.onText(text)
					}
				} else if event.Delta.Type == "input_json_delta" {
					if currentToolCall != nil {
						inputBuilder.WriteString(string(event.Delta.Partial))
					}
				}
			}

		case api.EventContentBlockStop:
			if currentToolCall != nil {
				currentToolCall.Input = json.RawMessage(inputBuilder.String())
				toolCalls = append(toolCalls, *currentToolCall)
				if a.onToolUse != nil {
					a.onToolUse(currentToolCall.Name, currentToolCall.ID, currentToolCall.Input)
				}
				currentToolCall = nil
			}

		case api.EventMessageStop:
			// 完成
		}
	}

	// 如果有文本输出，添加到 toolCalls
	if textBuilder.Len() > 0 {
		toolCalls = append([]api.ContentBlock{{
			Type: api.ContentTypeText,
			Text: textBuilder.String(),
		}}, toolCalls...)
	}

	return nil, toolCalls, nil
}

func (a *Agent) executeTools(ctx context.Context, toolCalls []api.ContentBlock) []tools.Result {
	var results []tools.Result
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, tc := range toolCalls {
		if tc.Type != api.ContentTypeToolUse {
			continue
		}

		wg.Add(1)
		go func(toolCall api.ContentBlock) {
			defer wg.Done()

			tool, ok := a.tools[toolCall.Name]
			if !ok {
				mu.Lock()
				results = append(results, tools.Result{
					Output:  fmt.Sprintf("Unknown tool: %s", toolCall.Name),
					IsError: true,
				})
				mu.Unlock()
				return
			}

			result, err := tool.Execute(ctx, toolCall.Input)
			if err != nil {
				result = &tools.Result{
					Output:  err.Error(),
					IsError: true,
				}
			}

			mu.Lock()
			results = append(results, *result)
			if a.onToolResult != nil {
				a.onToolResult(toolCall.ID, result.Output, result.IsError)
			}
			mu.Unlock()
		}(tc)
	}

	wg.Wait()
	return results
}

func toolCallsToContent(toolCalls []api.ContentBlock) []api.ContentBlock {
	return toolCalls
}

func toolResultsToContent(results []tools.Result) []api.ContentBlock {
	blocks := make([]api.ContentBlock, len(results))
	for i, r := range results {
		blocks[i] = api.ContentBlock{
			Type:    api.ContentTypeToolResult,
			Content: r.Output,
			IsError: r.IsError,
		}
	}
	return blocks
}