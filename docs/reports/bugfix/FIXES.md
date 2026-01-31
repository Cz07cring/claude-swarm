# Git Worktree 管理系统 - 修复文档

## 修复日期
2026-01-30

## 修复的问题

### 问题 1: 多个 swarm 进程同时运行
**现象：** 可以启动多个 swarm 实例，导致任务队列竞争和状态混乱

**根本原因：** 没有进程锁机制

**解决方案：**
- 添加 PID 文件锁 (`~/.claude-swarm/claude-swarm.pid`)
- 启动前检查是否已有进程在运行
- 停止时自动清理 PID 文件

**修改文件：**
- `cmd/swarm/start.go` - 添加 `checkPidLock()`, `writePidFile()`, `getPidFilePath()`
- `cmd/swarm/stop.go` - 添加 `cleanupPidFile()`

---

### 问题 2: 合并没有日志，无法追踪
**现象：** 任务完成后文件出现在 main 分支，但没有任何合并日志

**根本原因：**
1. 合并逻辑的触发条件不明确
2. 缺少详细日志

**解决方案：**
- 添加详细的合并日志：
  ```
  📊 agent-0: 检测到任务完成 (task: xxx)
  🔀 开始合并 agent-0 的工作到 main 分支...
  ✅ 合并成功 - 任务 xxx 已完成
  ```
- 在状态变化日志中包含任务信息

**修改文件：**
- `pkg/controller/coordinator.go` - `monitorAgent()` 函数

---

### 问题 3: 状态追踪不准确
**现象：** 日志显示 `state=working hasTask=false` 或 `state=idle hasTask=true`

**根本原因：**
- 调度器手动设置状态为 `Working`
- 监控器通过 Detector 分析输出更新状态
- 两者产生竞态条件，状态被覆盖

**解决方案：**
1. **移除调度器中手动设置状态**
   - 让 Detector 自然检测 agent 状态
   - 调度器只负责分配任务 (`CurrentTask`)

2. **智能状态更新逻辑**
   ```go
   // 如果有任务 + 之前不是 idle + 检测到 idle = 任务刚完成
   if currentTask != nil && prevState != idle && detectedState == idle {
       // 触发合并
   }
   ```

3. **状态和任务的一致性**
   - `CurrentTask` 由调度器设置
   - `State` 由 Detector 检测
   - 任务完成时，两者同时清理

**修改文件：**
- `pkg/controller/coordinator.go`
  - `monitorAgent()` - 智能状态更新
  - `runScheduler()` - 移除手动设置 Working 状态

---

## 修复后的工作流程

### 启动流程
```
1. 检查 PID 文件 → 如果存在且进程运行中，拒绝启动
2. 创建 PID 文件（写入当前进程 PID）
3. 创建 Worktrees 和分支
4. 启动 Agents
5. 启动监控和调度
```

### 任务执行流程
```
1. 调度器：发现 idle agent（state=idle && CurrentTask=nil）
2. 调度器：分配任务（设置 CurrentTask）
3. 调度器：发送任务描述给 agent
4. 监控器：检测到 agent 开始工作（state 变为 working）
5. 监控器：持续检测状态
6. 监控器：检测到任务完成（有任务 + 检测到 idle）
   📊 检测到任务完成
   🔀 开始合并
   ✅ 合并成功
7. 监控器：清空 CurrentTask，状态回到 idle
8. 循环到步骤 1
```

### 停止流程
```
1. 清理 Worktrees
2. 删除 Agent 分支
3. 删除 PID 文件
4. 终止 tmux 会话
```

---

## 测试方法

### 快速测试
```bash
# 1. 编译
./build.sh

# 2. 完整测试
chmod +x test-fixes.sh
./test-fixes.sh
```

### 手动测试
```bash
# 1. 测试 PID 锁
./swarm start --agents 2 &
sleep 3
./swarm start --agents 2  # 应该被阻止

# 2. 测试任务执行和合并
./swarm add-task "创建 test.txt 文件"
# 观察日志，应该看到合并日志

# 3. 测试清理
./swarm stop
# 验证 worktrees、分支、PID 文件都被清理
```

---

## 验证清单

- [ ] PID 文件锁正常工作
- [ ] 无法启动多个实例
- [ ] Worktrees 正确创建
- [ ] 任务正常执行
- [ ] 合并日志清晰可见
- [ ] 状态追踪准确（state 和 hasTask 一致）
- [ ] 停止时完全清理

---

## 已知限制

1. **单机限制**
   - PID 检查只在同一台机器有效
   - 如果在不同机器上启动，仍可能冲突

2. **异常退出**
   - 如果进程被 `kill -9` 强制杀死，PID 文件可能残留
   - 下次启动会检测进程是否真的在运行，自动清理

3. **Worktree 数量**
   - 建议最多 10 个 agents
   - 过多会影响性能

---

## 后续改进建议

1. **分布式锁**
   - 使用 Redis 或文件锁实现跨机器的互斥

2. **健康检查**
   - 定期检查 agents 是否正常工作
   - 自动重启卡死的 agents

3. **合并策略优化**
   - 支持批量合并（多个 agents 完成后一起合并）
   - 添加合并冲突的自动解决策略

4. **监控增强**
   - 添加 Prometheus metrics
   - 记录任务执行时间
   - 统计合并成功率
