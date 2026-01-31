# 代码质量改进记录 - 方案A (Critical 问题)

**日期**: 2026-01-30  
**改进类型**: Critical 缺陷修复

---

## ✅ 已完成的修复 (3/3)

### 1. ✅ 修复 taskqueue 文件锁缺失 (#1 - Critical)

**问题**: 多进程并发时，内存锁无法保护跨进程的文件访问

**影响**:
- 任务重复分配给多个 agent
- 数据文件损坏
- ABA 问题（读-修改-写竞态）

**解决方案**:
1. **添加文件锁 (flock)**
   - 读取时使用共享锁 (LOCK_SH)
   - 写入时使用独占锁 (LOCK_EX)
   - 自动解锁 (defer)

2. **原子写入**
   - 写入临时文件 (.tmp)
   - 原子 rename 覆盖目标文件
   - 失败时自动清理

3. **路径处理增强**
   - 修复 `~/` 路径边界条件
   - 添加空路径验证
   - 处理 `~` 单独情况

**修改文件**:
- `pkg/state/taskqueue.go` (+50 行)
  - 新增 `lockFile *os.File` 字段
  - 新增 `Close()` 方法
  - 改进 `NewTaskQueue()` 路径处理
  - 修改 `load()` 添加共享锁
  - 修改 `save()` 添加独占锁和原子写入

**代码片段**:
```go
// 读取时使用共享锁
func (tq *TaskQueue) load() error {
    if err := syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_SH); err != nil {
        return fmt.Errorf("failed to acquire read lock: %w", err)
    }
    defer syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_UN)
    // ...
}

// 写入时使用独占锁 + 原子写入
func (tq *TaskQueue) save() error {
    if err := syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_EX); err != nil {
        return fmt.Errorf("failed to acquire write lock: %w", err)
    }
    defer syscall.Flock(int(tq.lockFile.Fd()), syscall.LOCK_UN)
    
    // 原子写入
    tmpFile := tq.filePath + ".tmp"
    os.WriteFile(tmpFile, data, 0644)
    os.Rename(tmpFile, tq.filePath)  // 原子操作
}
```

---

### 2. ✅ 修复 coordinator 死锁风险 (#2 - Critical)

**问题**: monitorAgent 释放锁执行合并时，状态可能被其他 goroutine 修改

**影响**:
- 任务丢失（CurrentTask 被覆盖）
- 状态不一致
- 竞态条件

**解决方案**:
1. **保存状态快照**
   - 在释放锁前保存 taskID

2. **重新验证状态**
   - 重新获取锁后，验证 CurrentTask.ID 是否仍为预期值
   - 如果状态已变，记录警告而不是覆盖

3. **添加 continue 跳过**
   - 任务完成处理后，跳过正常状态更新逻辑

**修改文件**:
- `pkg/controller/coordinator.go:255-300` (~+20 行)

**代码片段**:
```go
// 保存状态快照
taskID := currentTask.ID
agent.mu.Unlock()

// 执行耗时操作
mergeErr := c.mergeAgentWork(agent)

// 重新获取锁并验证
agent.mu.Lock()
if agent.Status.CurrentTask != nil && agent.Status.CurrentTask.ID == taskID {
    // 状态仍然有效，安全更新
    agent.Status.CurrentTask = nil
    agent.Status.State = models.AgentStateIdle
} else {
    // 状态已变，记录警告
    log.Printf("⚠️  任务状态在合并过程中已变更")
}
agent.mu.Unlock()
continue  // 跳过正常更新
```

---

### 3. ✅ 添加 goroutine panic 恢复 (#3 - Critical)

**问题**: 后台 goroutine panic 导致 wg.Done() 不被调用

**影响**:
- goroutine 泄漏
- `Stop()` 永久阻塞在 `wg.Wait()`
- 整个系统挂起

**解决方案**:
1. **在所有后台 goroutine 添加 recover**
   - monitorAgent (每个 agent 一个)
   - runScheduler (单例)
   - runRescue (单例)

2. **确保 wg.Done() 总是调用**
   - 使用嵌套 defer
   - recover 在外层 defer 中

3. **记录 panic 信息**
   - 记录 panic 值
   - 提示查看 runtime 堆栈

**修改文件**:
- `pkg/controller/coordinator.go` (3个函数)
  - `monitorAgent()` :223
  - `runScheduler()` :326
  - `runRescue()` :386

**代码片段**:
```go
func (c *Coordinator) monitorAgent(agent *Agent) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("❌ PANIC in monitorAgent for %s: %v", agent.ID, r)
        }
        c.wg.Done()  // 总是调用
    }()
    
    // ... 原有逻辑
}
```

---

## 📊 改进统计

- **修复文件数**: 2
- **修改行数**: ~100 行
- **新增代码**: ~70 行
- **测试状态**: ✅ 编译通过

---

## 🔒 安全性提升

1. **多进程安全**: 文件锁保护跨进程并发
2. **数据一致性**: 原子写入防止部分写入
3. **并发安全**: 状态验证防止竞态条件
4. **系统稳定性**: panic 恢复防止系统挂起

---

## 🧪 建议测试

### 测试 1: 多进程并发
```bash
# 终端 1
./swarm start --agents 2

# 终端 2 (同时启动)
./swarm start --agents 2

# 应该看到 PID 锁阻止第二个实例
```

### 测试 2: 并发任务分配
```bash
# 快速添加多个任务
for i in {1..10}; do
  ./swarm add-task "Task $i" &
done
wait

# 检查任务队列，应该有 10 个任务，无重复
./swarm status
```

### 测试 3: 状态一致性
```bash
# 启动 swarm 并添加多个任务
./swarm start --agents 3
for i in {1..5}; do
  ./swarm add-task "Task $i"
done

# 观察日志，不应该有 "任务状态已变更" 警告
tmux attach -t claude-swarm
```

---

## 📋 待办事项 (High 优先级)

剩余 5 个问题建议继续修复：

- [ ] #4 - 改进 git 命令错误处理 (High)
- [ ] #5 - 添加 Gemini API 超时和重试 (High)
- [ ] #6 - 修复冲突解决超时控制 (High)
- [ ] #7 - 修复调度器 TOCTOU 问题 (Medium)
- [ ] #8 - 改进 worktree 清理逻辑 (Medium)

---

## 🎯 下一步

**选项 1**: 继续修复 High 优先级问题 (#4, #5, #6)  
**选项 2**: 提交当前修复，测试验证  
**选项 3**: 运行完整测试套件

