package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Memory 记忆条目
type Memory struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`      // user, feedback, project, reference
	Content   string                 `json:"content"`
	Tags      []string               `json:"tags,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Manager 记忆管理器
type Manager struct {
	path    string
	memories []Memory
	mu      sync.RWMutex
}

// NewManager 创建记忆管理器
func NewManager(path string) (*Manager, error) {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".local", "share", "go-claw-code", "memory")
	}

	m := &Manager{
		path:     path,
		memories: make([]Memory, 0),
	}

	if err := m.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return m, nil
}

// load 加载记忆
func (m *Manager) load() error {
	data, err := os.ReadFile(m.memoryFile())
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &m.memories)
}

// save 保存记忆
func (m *Manager) save() error {
	if err := os.MkdirAll(m.path, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.memories, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.memoryFile(), data, 0644)
}

func (m *Manager) memoryFile() string {
	return filepath.Join(m.path, "memories.json")
}

// Add 添加记忆
func (m *Manager) Add(memType, content string, tags []string) (*Memory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	mem := Memory{
		ID:        generateID(),
		Type:      memType,
		Content:   content,
		Tags:      tags,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m.memories = append(m.memories, mem)

	if err := m.save(); err != nil {
		return nil, err
	}

	return &mem, nil
}

// Get 获取记忆
func (m *Manager) Get(id string) *Memory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := range m.memories {
		if m.memories[i].ID == id {
			return &m.memories[i]
		}
	}
	return nil
}

// List 列出记忆
func (m *Manager) List(memType string) []Memory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if memType == "" {
		return m.memories
	}

	var result []Memory
	for _, mem := range m.memories {
		if mem.Type == memType {
			result = append(result, mem)
		}
	}
	return result
}

// Search 搜索记忆
func (m *Manager) Search(query string) []Memory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Memory
	for _, mem := range m.memories {
		if contains(mem.Content, query) || containsSlice(mem.Tags, query) {
			result = append(result, mem)
		}
	}
	return result
}

// Delete 删除记忆
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.memories {
		if m.memories[i].ID == id {
			m.memories = append(m.memories[:i], m.memories[i+1:]...)
			return m.save()
		}
	}
	return nil
}

// Update 更新记忆
func (m *Manager) Update(id, content string, tags []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.memories {
		if m.memories[i].ID == id {
			m.memories[i].Content = content
			m.memories[i].Tags = tags
			m.memories[i].UpdatedAt = time.Now()
			return m.save()
		}
	}
	return nil
}

// Recent 获取最近记忆
func (m *Manager) Recent(n int) []Memory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if n >= len(m.memories) {
		return m.memories
	}

	return m.memories[len(m.memories)-n:]
}

func generateID() string {
	return time.Now().Format("20060102150405") + randomSuffix(6)
}

func randomSuffix(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().Nanosecond()%len(letters)]
	}
	return string(b)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsSlice(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
