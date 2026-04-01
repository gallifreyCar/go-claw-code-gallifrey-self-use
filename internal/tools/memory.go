package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/services/memory"
)

// Memory 记忆工具
type MemoryTool struct {
	manager *memory.Manager
}

// MemoryInput 记忆输入
type MemoryInput struct {
	Action  string   `json:"action"`   // add, get, list, search, delete
	Type    string   `json:"type"`     // user, feedback, project, reference
	Content string   `json:"content"`  // 内容
	Tags    []string `json:"tags"`     // 标签
	ID      string   `json:"id"`       // ID (for get/delete)
	Query   string   `json:"query"`    // 搜索查询
}

func NewMemoryTool(manager *memory.Manager) *MemoryTool {
	return &MemoryTool{manager: manager}
}

func (t *MemoryTool) Name() string {
	return "Memory"
}

func (t *MemoryTool) Description() string {
	return `Manage persistent memory across conversations.
Actions: add, get, list, search, delete
Types: user (user info), feedback (corrections), project (context), reference (pointers)`
}

func (t *MemoryTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"action": {
				"type": "string",
				"enum": ["add", "get", "list", "search", "delete"],
				"description": "Action to perform"
			},
			"type": {
				"type": "string",
				"enum": ["user", "feedback", "project", "reference"],
				"description": "Type of memory"
			},
			"content": {
				"type": "string",
				"description": "Memory content (for add)"
			},
			"tags": {
				"type": "array",
				"items": {"type": "string"},
				"description": "Tags for categorization"
			},
			"id": {
				"type": "string",
				"description": "Memory ID (for get/delete)"
			},
			"query": {
				"type": "string",
				"description": "Search query"
			}
		},
		"required": ["action"]
	}`)
}

func (t *MemoryTool) RequiresConfirmation() bool {
	return false
}

func (t *MemoryTool) Execute(ctx context.Context, input json.RawMessage) (*Result, error) {
	var params MemoryInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	switch params.Action {
	case "add":
		return t.add(params)
	case "get":
		return t.get(params)
	case "list":
		return t.list(params)
	case "search":
		return t.search(params)
	case "delete":
		return t.delete(params)
	default:
		return &Result{
			Output:  fmt.Sprintf("Unknown action: %s", params.Action),
			IsError: true,
		}, nil
	}
}

func (t *MemoryTool) add(params MemoryInput) (*Result, error) {
	if params.Content == "" {
		return &Result{
			Output:  "Content is required for add action",
			IsError: true,
		}, nil
	}

	mem, err := t.manager.Add(params.Type, params.Content, params.Tags)
	if err != nil {
		return &Result{
			Output:  err.Error(),
			IsError: true,
		}, nil
	}

	return &Result{
		Output: fmt.Sprintf("Memory added with ID: %s", mem.ID),
		Metadata: map[string]interface{}{
			"id":   mem.ID,
			"type": mem.Type,
		},
	}, nil
}

func (t *MemoryTool) get(params MemoryInput) (*Result, error) {
	if params.ID == "" {
		return &Result{
			Output:  "ID is required for get action",
			IsError: true,
		}, nil
	}

	mem := t.manager.Get(params.ID)
	if mem == nil {
		return &Result{
			Output: fmt.Sprintf("Memory not found: %s", params.ID),
			IsError: true,
		}, nil
	}

	return &Result{
		Output: fmt.Sprintf("ID: %s\nType: %s\nContent: %s\nTags: %v\nCreated: %s",
			mem.ID, mem.Type, mem.Content, mem.Tags, mem.CreatedAt.Format("2006-01-02")),
	}, nil
}

func (t *MemoryTool) list(params MemoryInput) (*Result, error) {
	memories := t.manager.List(params.Type)
	if len(memories) == 0 {
		return &Result{
			Output: "No memories found",
		}, nil
	}

	var output string
	for _, mem := range memories {
		output += fmt.Sprintf("- [%s] %s: %s\n", mem.ID, mem.Type, truncate(mem.Content, 50))
	}

	return &Result{
		Output: output,
		Metadata: map[string]interface{}{
			"count": len(memories),
		},
	}, nil
}

func (t *MemoryTool) search(params MemoryInput) (*Result, error) {
	if params.Query == "" {
		return &Result{
			Output:  "Query is required for search action",
			IsError: true,
		}, nil
	}

	memories := t.manager.Search(params.Query)
	if len(memories) == 0 {
		return &Result{
			Output: fmt.Sprintf("No memories found matching: %s", params.Query),
		}, nil
	}

	var output string
	for _, mem := range memories {
		output += fmt.Sprintf("- [%s] %s: %s\n", mem.ID, mem.Type, truncate(mem.Content, 50))
	}

	return &Result{
		Output: output,
		Metadata: map[string]interface{}{
			"count": len(memories),
		},
	}, nil
}

func (t *MemoryTool) delete(params MemoryInput) (*Result, error) {
	if params.ID == "" {
		return &Result{
			Output:  "ID is required for delete action",
			IsError: true,
		}, nil
	}

	if err := t.manager.Delete(params.ID); err != nil {
		return &Result{
			Output:  err.Error(),
			IsError: true,
		}, nil
	}

	return &Result{
		Output: fmt.Sprintf("Memory deleted: %s", params.ID),
	}, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
