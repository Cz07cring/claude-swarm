# 智能确认系统文档

## 概述

Claude Agent Swarm 实现了智能确认系统，能够自动识别 Claude Code 的确认提示，并根据安全策略决定是否自动确认。

## 支持的确认格式

### 1. 选项列表格式

```
Do you want to proceed?
 ❯ 1. Yes
   2. No
```

**检测模式：** `❯ 1. Yes` 或 `1. Yes`
**发送内容：** `1`

### 2. yes/no 格式

```
Proceed with this operation? (yes/no)
```

**检测模式：** `yes/no`、`[yes/no]`、`(yes/no)`
**发送内容：** `yes`

### 3. (y/n) 格式

```
Continue? (y/n)
```

**检测模式：** `(y/n)`
**发送内容：** `y`

### 4. 计划确认

```
Proceed with this plan?
```

**检测模式：** `proceed with this plan`
**发送内容：** `yes`

## 安全检查机制

### 危险关键词检测

系统会检查最近50行的上下文，查找以下危险关键词：

- `delete` - 删除操作
- `remove` - 移除操作
- `drop` - 丢弃操作
- `force` - 强制操作
- `destructive` - 破坏性操作
- `rm -rf` - 强制删除
- `git reset --hard` - 强制重置
- `git push --force` - 强制推送
- `truncate` - 截断
- `destroy` - 销毁

**如果检测到任何危险关键词，拒绝自动确认。**

### 上下文分析

对于选项列表格式，系统会分析上下文是否包含安全操作：

**安全操作关键词：**
- `create` - 创建
- `read` - 读取
- `analyze` - 分析
- `show` - 显示
- `display` - 展示
- `list` - 列出
- `get` - 获取
- `fetch` - 拉取
- `view` - 查看
- `check` - 检查

**逻辑：**
- 如果上下文包含安全操作 **且** 不包含危险关键词 → 自动确认
- 否则 → 需要人工确认

### 特殊情况

#### 1. 文件覆盖

```
File already exists. Overwrite? (yes/no)
```

**处理：** 拒绝自动确认（需要人工确认）

#### 2. 计划中的危险操作

```
Proceed with this plan?
Plan:
1. Delete old files
2. Create new files
```

**处理：** 即使是计划确认，如果包含危险操作也拒绝自动确认

## 工作流程

```
┌─────────────────────────┐
│  监控器每3秒检查Agent    │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐
│  检测到等待确认状态？    │
└───────────┬─────────────┘
            │ Yes
            ▼
┌─────────────────────────┐
│  ShouldConfirm() 分析    │
│  1. 检查危险关键词       │
│  2. 分析确认格式         │
│  3. 检查上下文安全性     │
└───────────┬─────────────┘
            │
     ┌──────┴──────┐
     │             │
     ▼             ▼
  安全          不安全
     │             │
     ▼             ▼
  自动确认     人工确认
  发送对应输入   记录日志
```

## 日志输出

### 自动确认成功

```
✅ Auto-confirmed for agent-0 (sent: 1, reason: 安全检查通过)
```

**信息：**
- `agent-0` - Agent ID
- `sent: 1` - 发送的输入
- `reason` - 确认原因

### 需要人工确认

```
⚠️  agent-1 waiting for confirmation (reason: 检测到危险操作或无法判断安全性)
   请手动确认: tmux send-keys -t claude-swarm:0.1 "1" Enter
```

**信息：**
- `reason` - 拒绝自动确认的原因
- 提供手动确认命令

## 使用示例

### 示例 1: 安全操作 - 自动确认

**任务：** "创建一个新的配置文件"

**Claude 提示：**
```
I'll create a new config file. Do you want to proceed?
 ❯ 1. Yes
   2. No
```

**系统行为：**
- ✅ 检测到确认提示
- ✅ 上下文包含 "create"（安全操作）
- ✅ 没有危险关键词
- ✅ 自动发送 "1"

**日志：**
```
✅ Auto-confirmed for agent-0 (sent: 1, reason: 安全检查通过)
```

### 示例 2: 危险操作 - 人工确认

**任务：** "删除所有临时文件"

**Claude 提示：**
```
I'll delete all temporary files. Proceed with this plan?
Plan:
1. Find all *.tmp files
2. Delete them with rm -rf
```

**系统行为：**
- ❌ 检测到 "delete" 和 "rm -rf"
- ❌ 拒绝自动确认

**日志：**
```
⚠️  agent-0 waiting for confirmation (reason: 检测到危险操作或无法判断安全性)
   请手动确认: tmux send-keys -t claude-swarm:0.0 "yes" Enter
```

### 示例 3: 文件覆盖 - 人工确认

**任务：** "更新 README 文件"

**Claude 提示：**
```
File README.md already exists. Overwrite? (yes/no)
```

**系统行为：**
- ❌ 检测到 "overwrite"
- ❌ 拒绝自动确认

**日志：**
```
⚠️  agent-1 waiting for confirmation (reason: 检测到危险操作或无法判断安全性)
   请手动确认: tmux send-keys -t claude-swarm:0.1 "yes" Enter
```

## 手动确认方法

如果系统拒绝自动确认，你可以：

### 方法 1: 通过 tmux 发送命令

```bash
# 查看是哪个 Agent 需要确认
./swarm status

# 发送确认（假设是 agent-0，pane 索引 0）
tmux send-keys -t claude-swarm:0.0 "1" Enter

# 或者发送 yes
tmux send-keys -t claude-swarm:0.0 "yes" Enter
```

### 方法 2: 附加到 tmux 手动输入

```bash
# 附加到会话
tmux attach -t claude-swarm

# 在需要确认的窗格中输入 1 或 yes
# 按 Ctrl+B 然后 D 退出
```

## 配置选项

### 调整监控间隔

```bash
# 更频繁地检查（默认3秒）
./swarm start -i 2
```

### 禁用自动确认

目前没有配置选项禁用，但可以通过修改代码实现：

```go
// pkg/controller/coordinator.go
// 在 runRescue 函数中注释掉自动确认逻辑
```

## 最佳实践

1. **明确任务描述** - 避免模糊的任务，如"清理项目"
2. **分解危险操作** - 将包含删除的任务单独执行
3. **监控日志** - 定期查看哪些操作需要人工确认
4. **附加 tmux** - 对于关键操作，建议附加到 tmux 实时查看

## 故障排除

### 问题：应该自动确认但没有

**可能原因：**
- 上下文包含危险关键词
- 确认格式未被识别

**解决方案：**
1. 检查日志中的拒绝原因
2. 手动确认
3. 如果是误判，可以调整 `DangerKeywords` 列表

### 问题：自动确认了不应该确认的操作

**可能原因：**
- 危险关键词列表不完整
- 上下文分析逻辑有缺陷

**解决方案：**
1. 立即停止操作（Ctrl+C）
2. 更新 `DangerKeywords` 列表
3. 改进 `SafeToConfirm` 逻辑
4. 提交 Issue 报告

## 未来改进

- [ ] 支持更多确认格式
- [ ] 机器学习模型判断安全性
- [ ] 用户自定义安全策略
- [ ] 记录确认历史供审计
- [ ] Web UI 远程人工确认

---

**版本：** v0.2.0
**更新日期：** 2026-01-30
