package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/tui"
	"github.com/spf13/cobra"
)

var (
	// 构建信息
	version   string
	buildTime string

	// 全局标志
	cfgFile   string
	provider  string
	model     string
	printMode bool
)

// Execute 执行 CLI
func Execute(ver, build string) error {
	version = ver
	buildTime = build

	rootCmd := &cobra.Command{
		Use:   "go-claw-code",
		Short: "AI coding assistant in Go",
		Long: `Go-Claw-Code is an AI-powered coding assistant that helps you
write, understand, and modify code through natural language commands.

Inspired by Claude Code, implemented in Go for single-binary deployment.`,
		Version: version,
		Run:     runRoot,
	}

	// 全局标志
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/go-claw-code/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&provider, "provider", "P", "", "API provider (anthropic or openai)")
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "", "model to use")
	rootCmd.Flags().BoolVarP(&printMode, "print", "p", false, "print mode (non-interactive)")

	// 添加子命令
	rootCmd.AddCommand(versionCmd)

	return rootCmd.Execute()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("go-claw-code %s\n", version)
		fmt.Printf("  Build: %s\n", buildTime)
	},
}

func runRoot(cmd *cobra.Command, args []string) {
	if printMode {
		// 非交互模式
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: --print mode requires a prompt")
			os.Exit(1)
		}
		runPrintMode(args)
	} else {
		// TUI 交互模式
		if len(args) > 0 {
			// 有参数时也进入 print 模式
			runPrintMode(args)
			return
		}
		runTUI()
	}
}

func runTUI() {
	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}