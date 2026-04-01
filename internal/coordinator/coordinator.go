package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/services/api"
	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/tools"
)

// Coordinator 多 Agent 协调器
type Coordinator struct {
	client api.Client
	model  string
	tools  map[string]tools.Tool

	// 子 Agent 池
	workers sync.Map

	// 回调
	onWorkerStart   func(id, task string)
	onWorkerDone    func(id string, result string)
	onWorkerError   func(id string, err error)
}

// NewCoordinator 创建协调器
func NewCoordinator(client api.Client, model string) *Coordinator {
	return &Coordinator{
		client: client,
		model:  model,
		tools:  make(map[string]tools.Tool),
	}
}

// RegisterTool 注册工具
func (c *Coordinator) RegisterTool(tool tools.Tool) {
	c.tools[tool.Name()] = tool
}

// SpawnWorker 创建子 Agent
func (c *Coordinator) SpawnWorker(ctx context.Context, id, systemPrompt, input string) (*Agent, error) {
	agent := NewAgent(c.client, c.model)

	// 注册相同的工具
	for _, tool := range c.tools {
		agent.RegisterTool(tool)
	}

	c.workers.Store(id, agent)

	if c.onWorkerStart != nil {
		c.onWorkerStart(id, input)
	}

	return agent, nil
}

// RunWorker 运行子 Agent
func (c *Coordinator) RunWorker(ctx context.Context, id, systemPrompt, input string) (string, error) {
	agent, err := c.SpawnWorker(ctx, id, systemPrompt, input)
	if err != nil {
		return "", err
	}

	var result string
	var resultMu sync.Mutex

	agent.OnText(func(text string) {
		resultMu.Lock()
		result += text
		resultMu.Unlock()
	})

	if err := agent.Run(ctx, systemPrompt, input); err != nil {
		if c.onWorkerError != nil {
			c.onWorkerError(id, err)
		}
		return "", err
	}

	if c.onWorkerDone != nil {
		c.onWorkerDone(id, result)
	}

	return result, nil
}

// RunWorkers 并行运行多个子 Agent
func (c *Coordinator) RunWorkers(ctx context.Context, tasks []WorkerTask) []WorkerResult {
	var wg sync.WaitGroup
	results := make([]WorkerResult, len(tasks))
	resultChan := make(chan WorkerResult, len(tasks))

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t WorkerTask) {
			defer wg.Done()

			result := WorkerResult{
				ID:    t.ID,
				Error: nil,
			}

			output, err := c.RunWorker(ctx, t.ID, t.SystemPrompt, t.Input)
			if err != nil {
				result.Error = err
				result.Output = err.Error()
			} else {
				result.Output = output
			}

			resultChan <- result
		}(i, task)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	i := 0
	for result := range resultChan {
		results[i] = result
		i++
	}

	return results
}

// OnWorkerStart 设置子 Agent 启动回调
func (c *Coordinator) OnWorkerStart(fn func(id, task string)) {
	c.onWorkerStart = fn
}

// OnWorkerDone 设置子 Agent 完成回调
func (c *Coordinator) OnWorkerDone(fn func(id, result string)) {
	c.onWorkerDone = fn
}

// OnWorkerError 设置子 Agent 错误回调
func (c *Coordinator) OnWorkerError(fn func(id string, err error)) {
	c.onWorkerError = fn
}

// WorkerTask 子 Agent 任务
type WorkerTask struct {
	ID           string
	SystemPrompt string
	Input        string
}

// WorkerResult 子 Agent 结果
type WorkerResult struct {
	ID     string
	Output string
	Error  error
}

// AgentTool 子 Agent 工具（让 Agent 可以创建子 Agent）
type AgentTool struct {
	coordinator *Coordinator
}

// NewAgentTool 创建 Agent 工具
func NewAgentTool(coordinator *Coordinator) *AgentTool {
	return &AgentTool{coordinator: coordinator}
}

func (t *AgentTool) Name() string {
	return "Agent"
}

func (t *AgentTool) Description() string {
	return `Launch a new agent to handle complex, multi-step tasks autonomously.
The agent will work independently and return results when complete.`
}

func (t *AgentTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"task": {
				"type": "string",
				"description": "Description of the task for the agent"
			},
			"system_prompt": {
				"type": "string",
				"description": "Custom system prompt for the agent"
			}
		},
		"required": ["task"]
	}`)
}

func (t *AgentTool) RequiresConfirmation() bool {
	return false
}

func (t *AgentTool) Execute(ctx context.Context, input json.RawMessage) (*tools.Result, error) {
	var params struct {
		Task         string `json:"task"`
		SystemPrompt string `json:"system_prompt"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	systemPrompt := params.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant that completes tasks efficiently."
	}

	id := fmt.Sprintf("worker-%d", len(params.Task))

	result, err := t.coordinator.RunWorker(ctx, id, systemPrompt, params.Task)
	if err != nil {
		return &tools.Result{
			Output:  err.Error(),
			IsError: true,
		}, nil
	}

	return &tools.Result{
		Output: result,
		Metadata: map[string]interface{}{
			"id":   id,
			"task": params.Task,
		},
	}, nil
}
