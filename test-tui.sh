#!/bin/bash

# TUI Monitor 测试脚本
# 此脚本帮助验证 TUI Monitor 功能

set -e

echo "======================================"
echo "Claude Agent Swarm - TUI Monitor 测试"
echo "======================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查二进制文件
if [ ! -f "./swarm" ]; then
    echo -e "${RED}❌ 错误: 找不到 swarm 二进制文件${NC}"
    echo "   请先运行: go build -o swarm ./cmd/swarm"
    exit 1
fi

echo -e "${GREEN}✓${NC} 找到 swarm 二进制文件"
echo ""

# 清理旧的状态文件
echo "清理旧的状态文件..."
rm -f ~/.claude-swarm/tasks.json
rm -f ~/.claude-swarm/agents.json
rm -f ~/.claude-swarm/tasks.json.lock
rm -f ~/.claude-swarm/agents.json.lock
echo -e "${GREEN}✓${NC} 状态文件已清理"
echo ""

# 创建测试任务文件
echo "创建测试数据..."
mkdir -p ~/.claude-swarm

# 创建示例任务
cat > ~/.claude-swarm/tasks.json << 'EOF'
{
  "tasks": [
    {
      "id": "task-001",
      "description": "实现用户认证功能",
      "status": "pending",
      "created_at": "2026-01-30T10:00:00Z",
      "updated_at": "2026-01-30T10:00:00Z"
    },
    {
      "id": "task-002",
      "description": "创建数据库架构",
      "status": "in_progress",
      "assignee_id": "agent-0",
      "created_at": "2026-01-30T10:01:00Z",
      "updated_at": "2026-01-30T10:05:00Z"
    },
    {
      "id": "task-003",
      "description": "编写单元测试",
      "status": "completed",
      "assignee_id": "agent-1",
      "created_at": "2026-01-30T10:02:00Z",
      "updated_at": "2026-01-30T10:10:00Z"
    },
    {
      "id": "task-004",
      "description": "实现 API 端点",
      "status": "pending",
      "created_at": "2026-01-30T10:03:00Z",
      "updated_at": "2026-01-30T10:03:00Z"
    },
    {
      "id": "task-005",
      "description": "添加日志记录",
      "status": "failed",
      "assignee_id": "agent-2",
      "created_at": "2026-01-30T10:04:00Z",
      "updated_at": "2026-01-30T10:12:00Z"
    }
  ]
}
EOF

# 创建示例 Agent 状态
cat > ~/.claude-swarm/agents.json << 'EOF'
{
  "agents": [
    {
      "agent_id": "agent-0",
      "state": "working",
      "current_task": {
        "id": "task-002",
        "description": "创建数据库架构",
        "status": "in_progress",
        "assignee_id": "agent-0",
        "created_at": "2026-01-30T10:01:00Z",
        "updated_at": "2026-01-30T10:05:00Z"
      },
      "last_update": "2026-01-30T10:05:30Z",
      "output": "正在创建数据库表...\nCREATE TABLE users (\n  id SERIAL PRIMARY KEY,\n  username VARCHAR(50) UNIQUE NOT NULL,\n  email VARCHAR(100) UNIQUE NOT NULL\n);\n表创建成功"
    },
    {
      "agent_id": "agent-1",
      "state": "idle",
      "last_update": "2026-01-30T10:10:30Z",
      "output": "任务已完成，等待新任务..."
    },
    {
      "agent_id": "agent-2",
      "state": "error",
      "current_task": {
        "id": "task-005",
        "description": "添加日志记录",
        "status": "failed",
        "assignee_id": "agent-2",
        "created_at": "2026-01-30T10:04:00Z",
        "updated_at": "2026-01-30T10:12:00Z"
      },
      "last_update": "2026-01-30T10:12:30Z",
      "output": "错误: 无法导入日志库\nImportError: No module named 'logging'\n请安装依赖: pip install logging"
    },
    {
      "agent_id": "agent-3",
      "state": "idle",
      "last_update": "2026-01-30T10:00:00Z",
      "output": ""
    },
    {
      "agent_id": "agent-4",
      "state": "waiting_confirm",
      "last_update": "2026-01-30T10:13:00Z",
      "output": "准备提交更改到 git\n是否继续? (y/n)"
    }
  ],
  "updated_at": "2026-01-30T10:13:00Z"
}
EOF

echo -e "${GREEN}✓${NC} 测试数据已创建"
echo ""

# 显示测试数据统计
echo "测试数据统计:"
echo "  - 任务数量: 5"
echo "    ○ 待处理: 2"
echo "    ● 进行中: 1"
echo "    ✓ 已完成: 1"
echo "    ✗ 失败: 1"
echo "  - Agent 数量: 5"
echo "    - 空闲: 2"
echo "    - 工作中: 1"
echo "    - 等待确认: 1"
echo "    - 错误: 1"
echo ""

# 提示用户
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}准备启动 TUI Monitor${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "即将启动 TUI 监控面板，您将看到："
echo ""
echo "  📋 左侧: 任务列表"
echo "     - 5 个任务，不同状态"
echo "     - 使用 j/k 或 ↑/↓ 导航"
echo ""
echo "  👥 中间: Agent 网格"
echo "     - 5 个 Agent 的实时状态"
echo "     - 颜色编码显示不同状态"
echo ""
echo "  📝 右侧: 日志查看器"
echo "     - 显示选中 Agent 的输出"
echo "     - 自动滚动显示"
echo ""
echo "  ⌨️  键盘操作:"
echo "     - Tab: 切换面板"
echo "     - j/k 或 ↑/↓: 上下导航"
echo "     - h/l 或 ←/→: 左右导航 (Agent 网格)"
echo "     - q: 退出"
echo ""
echo -e "${GREEN}测试要点:${NC}"
echo "  1. 验证三个面板都正确显示"
echo "  2. 测试 Tab 键切换面板"
echo "  3. 在任务列表中导航，观察日志更新"
echo "  4. 在 Agent 网格中选择不同 Agent"
echo "  5. 观察不同状态的颜色编码"
echo "  6. 按 q 退出并确认程序正常关闭"
echo ""
read -p "按 Enter 启动 TUI Monitor..."

# 启动 TUI Monitor
./swarm monitor

# 测试完成
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}测试完成！${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "如果一切正常，您应该已经体验到:"
echo "  ✓ 三窗格布局"
echo "  ✓ 任务状态可视化"
echo "  ✓ Agent 状态网格"
echo "  ✓ 日志查看器"
echo "  ✓ 键盘导航"
echo ""
echo "下一步:"
echo "  1. 使用真实 swarm 实例测试:"
echo "     ./swarm start --agents 3"
echo "     ./swarm monitor"
echo ""
echo "  2. 添加任务并观察实时更新:"
echo "     ./swarm add-task \"测试任务\""
echo ""
echo "  3. 查看文档:"
echo "     cat docs/TUI_MONITOR.md"
echo ""
