package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	API        APIConfig     `mapstructure:"api"`
	Permission string        `mapstructure:"permission"`
	Log        LogConfig     `mapstructure:"log"`
	Memory     MemoryConfig  `mapstructure:"memory"`
	Agent      AgentConfig   `mapstructure:"agent"`
}

// APIConfig API 配置
type APIConfig struct {
	Provider  string          `mapstructure:"provider"`
	Anthropic AnthropicConfig `mapstructure:"anthropic"`
	OpenAI    OpenAIConfig    `mapstructure:"openai"`
}

// AnthropicConfig Anthropic 配置
type AnthropicConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// MemoryConfig 记忆配置
type MemoryConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// AgentConfig Agent 配置
type AgentConfig struct {
	MaxIterations int    `mapstructure:"max_iterations"`
	Timeout       string `mapstructure:"timeout"`
}

// Load 加载配置
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		home, _ := os.UserHomeDir()
		v.AddConfigPath(filepath.Join(home, ".config", "go-claw-code"))
		v.AddConfigPath(".")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// 环境变量
	v.SetEnvPrefix("GO_CLAW_CODE")
	v.AutomaticEnv()

	// 绑定环境变量
	_ = v.BindEnv("api.anthropic.api_key", "ANTHROPIC_API_KEY")
	_ = v.BindEnv("api.openai.api_key", "OPENAI_API_KEY")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// 展开环境变量
	expandEnvVars(&cfg)

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("api.provider", "anthropic")
	v.SetDefault("api.anthropic.model", "claude-sonnet-4-6")
	v.SetDefault("api.anthropic.base_url", "https://api.anthropic.com")
	v.SetDefault("api.openai.model", "gpt-4")
	v.SetDefault("api.openai.base_url", "https://api.openai.com/v1")
	v.SetDefault("permission", "ask")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("memory.enabled", true)
	v.SetDefault("agent.max_iterations", 100)
	v.SetDefault("agent.timeout", "10m")
}

func expandEnvVars(cfg *Config) {
	cfg.API.Anthropic.APIKey = expandEnv(cfg.API.Anthropic.APIKey)
	cfg.API.OpenAI.APIKey = expandEnv(cfg.API.OpenAI.APIKey)
}

func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		return os.Getenv(s[2 : len(s)-1])
	}
	return s
}

// GetAPIKey 获取当前提供商的 API Key
func (c *Config) GetAPIKey() string {
	switch c.API.Provider {
	case "openai":
		return c.API.OpenAI.APIKey
	default:
		return c.API.Anthropic.APIKey
	}
}

// GetModel 获取当前提供商的模型
func (c *Config) GetModel() string {
	switch c.API.Provider {
	case "openai":
		return c.API.OpenAI.Model
	default:
		return c.API.Anthropic.Model
	}
}