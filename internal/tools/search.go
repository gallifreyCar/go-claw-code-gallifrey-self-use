package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Glob 文件搜索工具
type Glob struct{}

// GlobInput Glob 输入参数
type GlobInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
}

func NewGlob() *Glob {
	return &Glob{}
}

func (t *Glob) Name() string {
	return "Glob"
}

func (t *Glob) Description() string {
	return `Fast file pattern matching tool. Supports glob patterns like "**/*.js" or "src/**/*.ts".`
}

func (t *Glob) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"pattern": {
				"type": "string",
				"description": "The glob pattern to match files against"
			},
			"path": {
				"type": "string",
				"description": "The directory to search in (default: current directory)"
			}
		},
		"required": ["pattern"]
	}`)
}

func (t *Glob) RequiresConfirmation() bool {
	return false
}

func (t *Glob) Execute(ctx context.Context, input json.RawMessage) (*Result, error) {
	var params GlobInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}

	var matches []string
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(path, filePath)
		if err != nil {
			return nil
		}

		matched, err := filepath.Match(params.Pattern, filepath.Base(relPath))
		if err != nil {
			return nil
		}
		if matched {
			matches = append(matches, filePath)
		}

		// 支持双星号模式
		if strings.Contains(params.Pattern, "**") {
			// 简单的 ** 支持
			patternParts := strings.Split(params.Pattern, "**")
			if len(patternParts) == 2 {
				prefix := patternParts[0]
				suffix := patternParts[1]
				relPath = filepath.ToSlash(relPath)
				if strings.HasPrefix(relPath, prefix) && strings.HasSuffix(relPath, suffix) {
					matches = append(matches, filePath)
				}
			}
		}

		return nil
	})

	if err != nil {
		return &Result{
			Output:  err.Error(),
			IsError: true,
		}, nil
	}

	if len(matches) == 0 {
		return &Result{
			Output: "No files found matching pattern: " + params.Pattern,
		}, nil
	}

	return &Result{
		Output: strings.Join(matches, "\n"),
		Metadata: map[string]interface{}{
			"pattern": params.Pattern,
			"count":   len(matches),
		},
	}, nil
}

// Grep 内容搜索工具
type Grep struct{}

// GrepInput Grep 输入参数
type GrepInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
	Glob    string `json:"glob,omitempty"`
	ICase   bool   `json:"i,omitempty"`
}

func NewGrep() *Grep {
	return &Grep{}
}

func (t *Grep) Name() string {
	return "Grep"
}

func (t *Grep) Description() string {
	return `A powerful search tool built on ripgrep-style patterns.
Searches for patterns within file contents.`
}

func (t *Grep) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"pattern": {
				"type": "string",
				"description": "The regular expression pattern to search for"
			},
			"path": {
				"type": "string",
				"description": "File or directory to search in"
			},
			"glob": {
				"type": "string",
				"description": "Glob pattern to filter files"
			},
			"i": {
				"type": "boolean",
				"description": "Case insensitive search"
			}
		},
		"required": ["pattern"]
	}`)
}

func (t *Grep) RequiresConfirmation() bool {
	return false
}

func (t *Grep) Execute(ctx context.Context, input json.RawMessage) (*Result, error) {
	var params GrepInput
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}

	pattern := params.Pattern
	if params.ICase {
		pattern = strings.ToLower(pattern)
	}

	var results []string

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		// 检查 glob 过滤
		if params.Glob != "" {
			matched, err := filepath.Match(params.Glob, filepath.Base(filePath))
			if err != nil || !matched {
				return nil
			}
		}

		file, err := os.Open(filePath)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			searchLine := line
			if params.ICase {
				searchLine = strings.ToLower(line)
			}

			if strings.Contains(searchLine, pattern) {
				results = append(results, fmt.Sprintf("%s:%d:%s", filePath, lineNum, line))
			}
		}

		return nil
	})

	if err != nil {
		return &Result{
			Output:  err.Error(),
			IsError: true,
		}, nil
	}

	if len(results) == 0 {
		return &Result{
			Output: "No matches found for pattern: " + params.Pattern,
		}, nil
	}

	return &Result{
		Output: strings.Join(results, "\n"),
		Metadata: map[string]interface{}{
			"pattern": params.Pattern,
			"count":   len(results),
		},
	}, nil
}
