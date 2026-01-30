#!/bin/bash

echo "🔨 编译 Claude Swarm..."
cd "/Users/ring/Documents/公司源码/ringsite/claude-swarm"

# 编译
if go build -o swarm ./cmd/swarm; then
    echo "✅ 编译成功！"
    echo
    ls -lh swarm
    echo
    echo "运行以下命令测试："
    echo "  ./swarm --help"
    echo "  ./swarm start --agents 2"
else
    echo "❌ 编译失败"
    exit 1
fi
