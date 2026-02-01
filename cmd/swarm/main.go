package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "swarm",
	Short: "Claude Agent Swarm - 多 Agent 协作开发环境",
	Long: `Claude Agent Swarm 是一个基于 Claude CLI 的多 Agent 协作系统。
它使用 Git worktree 隔离和直接 CLI 执行，实现任务自动分发和智能协助。`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
