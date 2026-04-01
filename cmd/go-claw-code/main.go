package main

import (
	"fmt"
	"os"

	"github.com/gallifreyCar/go-claw-code-gallifrey-self-use/internal/cli"
)

// 构建时注入
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	if err := cli.Execute(Version, BuildTime); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}