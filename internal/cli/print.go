package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/config"
	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/coordinator"
	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/prompt"
	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/services/api"
	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/tools"
)

func runPrintMode(args []string) {
	// 加载配置
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 命令行覆盖
	if model != "" {
		cfg.API.Anthropic.Model = model
		cfg.API.OpenAI.Model = model
	}
	if provider != "" {
		cfg.API.Provider = provider
	}

	// 检查 API Key
	if cfg.GetAPIKey() == "" {
		fmt.Fprintln(os.Stderr, "Error: No API key configured. Set ANTHROPIC_API_KEY or OPENAI_API_KEY environment variable.")
		os.Exit(1)
	}

	// 创建客户端
	var client api.Client
	switch cfg.API.Provider {
	case "openai":
		client = api.NewOpenAIClient(api.OpenAIConfig{
			APIKey:  cfg.API.OpenAI.APIKey,
			Model:   cfg.API.OpenAI.Model,
			BaseURL: cfg.API.OpenAI.BaseURL,
		})
	default:
		client = api.NewAnthropicClient(api.AnthropicConfig{
			APIKey:  cfg.API.Anthropic.APIKey,
			Model:   cfg.API.Anthropic.Model,
			BaseURL: cfg.API.Anthropic.BaseURL,
		})
	}

	// 创建 Agent
	agent := coordinator.NewAgent(client, cfg.GetModel())

	// 注册工具
	agent.RegisterTool(tools.NewBash())
	agent.RegisterTool(tools.NewRead())
	agent.RegisterTool(tools.NewWrite())
	agent.RegisterTool(tools.NewEdit())
	agent.RegisterTool(tools.NewGlob())
	agent.RegisterTool(tools.NewGrep())

	// 设置回调
	agent.OnText(func(text string) {
		fmt.Print(text)
	})

	agent.OnToolUse(func(name, id string, input json.RawMessage) {
		fmt.Printf("\n🔧 Using tool: %s\n", name)
	})

	agent.OnToolResult(func(id string, result string, isError bool) {
		if isError {
			fmt.Printf("❌ Tool error: %s\n", result)
		} else {
			fmt.Printf("✅ Tool result: %s\n", truncate(result, 200))
		}
	})

	// 构建输入
	input := args[0]
	for _, arg := range args[1:] {
		input += " " + arg
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 处理中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nInterrupted by user")
		cancel()
	}()

	// 构建 System Prompt
	pm := prompt.NewManager().SetDefault(prompt.DefaultPrompt())
	systemPrompt := pm.Build()

	// 运行 Agent
	fmt.Printf("🤖 Processing with %s (%s)...\n\n", client.GetProvider(), cfg.GetModel())

	if err := agent.Run(ctx, systemPrompt, input); err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n\n✅ Done!")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}