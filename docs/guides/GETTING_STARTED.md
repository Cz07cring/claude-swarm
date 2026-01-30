# 开始使用 Claude Agent Swarm

## 🎉 恭喜！MVP 已完成

Claude Agent Swarm MVP 版本已经开发完成并可以使用了！

## 📦 项目内容

### 核心模块
- ✅ **tmux 管理** - 会话和窗格控制
- ✅ **任务队列** - JSON 文件存储
- ✅ **状态检测** - 正则模式匹配
- ✅ **协调器** - 调度、监控、救援
- ✅ **CLI** - 完整的命令行工具

### 文档
- ✅ **完整架构文档** - `docs/architecture/full-plan.md`
- ✅ **MVP 实施指南** - `docs/guides/mvp-guide.md`
- ✅ **快速开始** - `docs/guides/quickstart.md`
- ✅ **项目总结** - `docs/PROJECT_SUMMARY.md`
- ✅ **README** - `README.md`

## 🚀 立即开始

### 1. 检查环境

```bash
# 检查 Go
go version  # 应该显示 go1.21+

# 检查 tmux
tmux -V     # 应该显示 tmux 3.x+

# 检查 Claude
claude --version
```

### 2. 构建项目

项目已经构建完成，二进制文件位于：`./swarm`

如果需要重新构建：
```bash
go build -o swarm ./cmd/swarm
```

### 3. 第一次运行

```bash
# 1. 启动集群（3 个 Agent）
./swarm start

# 输出：
# 🚀 启动 Claude Agent Swarm...
# ✓ Created tmux session: claude-swarm
# ✓ Started agent-0 in pane 0
# ✓ Started agent-1 in pane 1
# ✓ Started agent-2 in pane 2
# ✓ Coordinator running...
```

**在新终端窗口：**

```bash
# 2. 添加测试任务
./swarm add-task "列出当前目录的文件"

# 3. 查看状态
./swarm status

# 4. 查看实时输出
tmux attach -t claude-swarm

# 5. 退出 tmux（不停止）
# 按 Ctrl+B 然后按 D

# 6. 停止集群
./swarm stop
```

## 📚 学习路径

1. **快速入门**
   - 阅读 `docs/guides/quickstart.md`
   - 运行基础示例
   - 熟悉 CLI 命令

2. **深入理解**
   - 阅读 `docs/architecture/full-plan.md`
   - 了解完整架构设计
   - 查看 `docs/PROJECT_SUMMARY.md` 了解实现细节

3. **高级使用**
   - 阅读 `docs/guides/mvp-guide.md`
   - 探索源码
   - 贡献改进

## 🧪 测试场景

### 场景 1: 基础功能测试

```bash
# 启动
./swarm start -n 2

# 添加简单任务
./swarm add-task "echo 'Hello from Agent'"
./swarm add-task "pwd"
./swarm add-task "date"

# 查看状态
./swarm status

# 观察实时输出
tmux attach -t claude-swarm
```

### 场景 2: 自动确认测试

```bash
./swarm start

# 安全任务（应自动确认）
./swarm add-task "创建一个新的 README 文件"

# 危险任务（不应自动确认）
./swarm add-task "删除所有临时文件"

# 附加到 tmux 观察行为
tmux attach -t claude-swarm
```

### 场景 3: 并行处理测试

```bash
./swarm start -n 3

# 添加多个任务
for i in {1..5}; do
  ./swarm add-task "处理任务 $i"
done

# 监控进度
watch -n 2 ./swarm status
```

## 🔍 调试技巧

### 查看详细日志

协调器会在控制台输出日志：
```
📋 Assigned task task-xxx to agent-0: 任务描述
✅ Auto-confirmed for agent-1
❌ agent-2 encountered an error
```

### 手动检查 tmux

```bash
# 列出所有会话
tmux ls

# 列出窗格
tmux list-panes -t claude-swarm

# 捕获窗格输出
tmux capture-pane -p -t claude-swarm:0.0
```

### 检查任务队列

```bash
# 查看任务文件
cat ~/.claude-swarm/tasks.json

# 格式化查看
cat ~/.claude-swarm/tasks.json | python -m json.tool
```

## 🐛 常见问题

### Q: Agent 没有响应？

```bash
# 附加到 tmux
tmux attach -t claude-swarm

# 手动在窗格中重启 claude
# 按 Ctrl+C，然后输入: claude
```

### Q: 会话已存在？

```bash
# 手动终止
tmux kill-session -t claude-swarm

# 重新启动
./swarm start
```

### Q: 任务队列损坏？

```bash
# 删除任务文件
rm ~/.claude-swarm/tasks.json

# 重新启动会自动创建
./swarm start
```

## 📝 下一步计划

### 短期（1-2 周）
- [ ] 增加更多测试用例
- [ ] 优化状态检测准确性
- [ ] 添加错误重试机制
- [ ] 改进日志格式

### 中期（1-2 月）
- [ ] 实现 Git worktree 管理
- [ ] 添加 SQLite 数据库
- [ ] 开发 TUI 仪表板
- [ ] 支持任务依赖

### 长期（3-6 月）
- [ ] 智能调度算法
- [ ] P2P 救援机制
- [ ] Windows 支持
- [ ] Docker 镜像
- [ ] Homebrew 发布

## 🤝 贡献指南

欢迎贡献！请：

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📧 获取帮助

- **GitHub Issues**: 报告 bug 或请求功能
- **文档**: 查看 `docs/` 目录
- **示例**: 查看 `README.md` 中的示例

## 🎯 项目目标

1. ✅ **MVP 完成** - 验证核心概念
2. 🔄 **功能完善** - 添加 Git、数据库、TUI
3. 📦 **发布到社区** - GitHub, Homebrew
4. 🌟 **收集反馈** - 改进和优化

---

**开始你的 Agent 协作之旅吧！** 🐝

如有问题，请查看文档或提交 Issue。
