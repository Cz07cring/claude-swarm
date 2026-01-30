package main

import (
	"fmt"
	"github.com/yourusername/claude-swarm/pkg/analyzer"
)

func main() {
	detector := analyzer.NewDetector()

	// 模拟 Claude 的空闲输出
	output := `╭─── Claude Code v2.1.25 ──────────────────────────────────────────────────────╮
│                                            │ Tips for getting started        │
│                Welcome back!               │ Run /init to create a CLAUDE.m… │
│                                            │ ─────────────────────────────── │
│                                            │ Recent activity                 │
│                   ▐▛███▜▌                  │ No recent activity              │
│                  ▝▜█████▛▘                 │                                 │
│                    ▘▘ ▝▝                   │                                 │
│          Sonnet 4.5 · Claude API           │                                 │
│ ~/Documents/公司源码/ringsite/claude-swarm │                                 │
╰──────────────────────────────────────────────────────────────────────────────╯

────────────────────────────────────────────────────────────────────────────────
❯ Try "fix lint errors"
────────────────────────────────────────────────────────────────────────────────
  ? for shortcuts`

	state := detector.Analyze(output)
	fmt.Printf("Detected state: %s\n", state)

	if state == "idle" {
		fmt.Println("✅ Correctly detected as IDLE")
	} else {
		fmt.Println("❌ Failed to detect as IDLE")
	}
}
