package tools

import (
	"context"
	"encoding/json"
	"os"
)

// Read 文件读取工具
type Read struct{}

// ReadInput 读取输入
type ReadInput struct {
	FilePath string `json:"file_path"`
	Offset   int    `json:"offset,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

func NewRead() *Read {
	return &Read{}
}

func (t *Read) Name() string {
	return "Read"
}

func (t *Read) Description() string {
	return `Read a file from the local filesystem.`
}

func (t *Read) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file_path": {
				"type": "string",
				"description": "The absolute path to the file to read"
			},
			"offset": {
				"type": "integer",
				"description": "Line number to start reading from"
			},
			"limit": {
				"type": "integer",
				"description": "Number of lines to read"
			}
		},
		"required": ["file_path"]
	}`)
}

func (t *Read) RequiresConfirmation() bool {
	return false
}

func (t *Read) Execute(ctx context.Context, input json.RawMessage) (*Result, error) {
	var params ReadInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(params.FilePath)
	if err != nil {
		return &Result{
			Output:  err.Error(),
			IsError: true,
		}, nil
	}

	return &Result{
		Output: string(data),
		Metadata: map[string]interface{}{
			"file_path": params.FilePath,
			"size":      len(data),
		},
	}, nil
}

// Write 文件写入工具
type Write struct{}

// WriteInput 写入输入
type WriteInput struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

func NewWrite() *Write {
	return &Write{}
}

func (t *Write) Name() string {
	return "Write"
}

func (t *Write) Description() string {
	return `Write a file to the local filesystem.`
}

func (t *Write) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file_path": {
				"type": "string",
				"description": "The absolute path to the file to write"
			},
			"content": {
				"type": "string",
				"description": "The content to write to the file"
			}
		},
		"required": ["file_path", "content"]
	}`)
}

func (t *Write) RequiresConfirmation() bool {
	return true
}

func (t *Write) Execute(ctx context.Context, input json.RawMessage) (*Result, error) {
	var params WriteInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	if err := os.WriteFile(params.FilePath, []byte(params.Content), 0644); err != nil {
		return &Result{
			Output:  err.Error(),
			IsError: true,
		}, nil
	}

	return &Result{
		Output: "File written successfully: " + params.FilePath,
		Metadata: map[string]interface{}{
			"file_path": params.FilePath,
			"size":      len(params.Content),
		},
	}, nil
}

// Edit 文件编辑工具
type Edit struct{}

// EditInput 编辑输入
type EditInput struct {
	FilePath   string `json:"file_path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

func NewEdit() *Edit {
	return &Edit{}
}

func (t *Edit) Name() string {
	return "Edit"
}

func (t *Edit) Description() string {
	return `Perform exact string replacements in files.`
}

func (t *Edit) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file_path": {
				"type": "string",
				"description": "The absolute path to the file to modify"
			},
			"old_string": {
				"type": "string",
				"description": "The text to replace"
			},
			"new_string": {
				"type": "string",
				"description": "The text to replace it with"
			},
			"replace_all": {
				"type": "boolean",
				"description": "Replace all occurrences"
			}
		},
		"required": ["file_path", "old_string", "new_string"]
	}`)
}

func (t *Edit) RequiresConfirmation() bool {
	return true
}

func (t *Edit) Execute(ctx context.Context, input json.RawMessage) (*Result, error) {
	var params EditInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(params.FilePath)
	if err != nil {
		return &Result{
			Output:  err.Error(),
			IsError: true,
		}, nil
	}

	content := string(data)
	if params.ReplaceAll {
		content = replaceAll(content, params.OldString, params.NewString)
	} else {
		content = replaceFirst(content, params.OldString, params.NewString)
	}

	if err := os.WriteFile(params.FilePath, []byte(content), 0644); err != nil {
		return &Result{
			Output:  err.Error(),
			IsError: true,
		}, nil
	}

	return &Result{
		Output: "File edited successfully: " + params.FilePath,
	}, nil
}

func replaceFirst(s, old, new string) string {
	idx := len(s)
	for i := 0; i <= len(s)-len(old); i++ {
		if s[i:i+len(old)] == old {
			idx = i
			break
		}
	}
	if idx < len(s) {
		return s[:idx] + new + s[idx+len(old):]
	}
	return s
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := -1
		for i := 0; i <= len(s)-len(old); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx < 0 {
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result + s
}