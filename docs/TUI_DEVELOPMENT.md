# TUI Monitor 开发者文档

## 概述

TUI Monitor 是基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 框架构建的终端用户界面，用于实时监控 Claude Agent Swarm 的运行状态。本文档面向希望理解、修改或扩展 TUI Monitor 的开发者。

## 技术栈

### 核心框架

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** v0.23+
  - Elm 架构风格的 TUI 框架
  - 基于消息传递的状态管理
  - 声明式 UI 渲染

- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** v0.7+
  - 终端样式库
  - 类似 CSS 的样式系统
  - 布局和对齐工具

- **Go 1.21+**
  - 核心编程语言
  - 并发和 goroutine 支持

### 相关组件

- `pkg/state/taskqueue.go` - 任务队列管理
- `pkg/state/agentstate.go` - Agent 状态持久化
- `internal/models/` - 数据模型定义

## 架构设计

### Bubble Tea 模型

TUI Monitor 遵循 Bubble Tea 的 Elm 架构模式：

```
┌─────────────────────────────────────────┐
│           Bubble Tea Runtime            │
│                                         │
│  ┌──────────┐    ┌──────────┐          │
│  │   Init   │───▶│  Update  │◀─┐       │
│  └──────────┘    └──────────┘  │       │
│                        │        │       │
│                        ▼        │       │
│                   ┌──────────┐  │       │
│                   │   View   │  │       │
│                   └──────────┘  │       │
│                        │        │       │
│                        ▼        │       │
│                   ┌──────────┐  │       │
│                   │ Terminal │  │       │
│                   └──────────┘  │       │
│                        │        │       │
│                        ▼        │       │
│                   ┌──────────┐  │       │
│                   │   User   │──┘       │
│                   └──────────┘          │
└─────────────────────────────────────────┘
```

**关键概念**:
- **Model**: 应用状态（`Dashboard` 结构体）
- **Message**: 事件（键盘输入、定时器等）
- **Update**: 状态更新逻辑
- **View**: 渲染函数（生成字符串）

### 组件结构

```
pkg/tui/
├── dashboard.go      # 主应用模型和协调器
├── tasklist.go       # 任务列表视图组件
├── agentgrid.go      # Agent 网格视图组件
├── logviewer.go      # 日志查看器组件
└── styles.go         # Lipgloss 样式定义
```

#### 文件职责

| 文件 | 职责 | 关键类型 |
|------|------|----------|
| `dashboard.go` | 主模型、消息处理、布局协调 | `Dashboard`, `tickMsg` |
| `tasklist.go` | 任务列表渲染、导航逻辑 | `TaskListView` |
| `agentgrid.go` | Agent 网格渲染、2D 导航 | `AgentGridView` |
| `logviewer.go` | 日志显示、滚动逻辑 | `LogViewerView` |
| `styles.go` | 全局样式常量 | `lipgloss.Style` 变量 |

### 数据流

```
┌─────────────────────────────────────────────────────────┐
│                    Coordinator                          │
│                   (pkg/controller)                      │
└──────────────┬──────────────────────┬───────────────────┘
               │ (每 2 秒写入)         │
               ▼                      ▼
     ┌─────────────────┐    ┌─────────────────┐
     │  agents.json    │    │   tasks.json    │
     └─────────────────┘    └─────────────────┘
               │                      │
               │ (每 2 秒读取)         │
               ▼                      ▼
     ┌─────────────────────────────────────────┐
     │          TUI Monitor (Dashboard)        │
     │  ┌───────────┐ ┌───────────┐ ┌────────┐│
     │  │ TaskList  │ │AgentGrid  │ │LogView ││
     │  └───────────┘ └───────────┘ └────────┘│
     └─────────────────────────────────────────┘
                       │
                       ▼
                  Terminal UI
```

**关键点**:
- TUI 是只读的（不修改状态文件）
- 通过文件系统与 Coordinator 通信（无网络或 IPC）
- 文件锁确保并发安全（由 `pkg/state` 处理）

## 核心组件详解

### 1. Dashboard（主模型）

位置：`pkg/tui/dashboard.go`

#### 职责
- 管理整个 TUI 应用状态
- 协调三个子视图（Tasks、Agents、Logs）
- 处理用户输入和定时器事件
- 控制面板激活状态和焦点

#### 关键字段

```go
type Dashboard struct {
    taskQueue    *state.TaskQueue              // 任务队列管理器
    getAgentsFn  func() []*models.AgentStatus  // Agent 状态获取函数
    taskList     *TaskListView                 // 任务列表子组件
    agentGrid    *AgentGridView                // Agent 网格子组件
    logViewer    *LogViewerView                // 日志查看器子组件
    activePane   ActivePane                    // 当前激活的面板
    width        int                           // 终端宽度
    height       int                           // 终端高度
    tasks        []*models.Task                // 任务数据缓存
    agents       []*models.AgentStatus         // Agent 数据缓存
    quitting     bool                          // 退出标志
    updateTicker *time.Ticker                  // 更新定时器
}
```

#### 消息类型

```go
type tickMsg time.Time  // 定时器消息（每 2 秒）
tea.KeyMsg              // 键盘输入消息
tea.WindowSizeMsg       // 窗口大小变化消息
```

#### 生命周期方法

1. **Init()**
   ```go
   func (m *Dashboard) Init() tea.Cmd
   ```
   - 初始化应用状态
   - 启动定时器
   - 加载初始数据

2. **Update(msg tea.Msg)**
   ```go
   func (m *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd)
   ```
   - 处理消息
   - 更新状态
   - 返回新模型和命令

3. **View()**
   ```go
   func (m *Dashboard) View() string
   ```
   - 渲染 UI
   - 组合三个面板
   - 返回字符串输出

#### 关键方法

**refreshData()**
```go
func (m *Dashboard) refreshData()
```
从状态文件读取最新数据并更新子组件。

**handleKeyPress(msg tea.KeyMsg)**
```go
func (m *Dashboard) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd)
```
处理键盘输入：
- `Tab`/`Shift+Tab`: 切换面板
- `hjkl`/方向键: 导航
- `q`/`Ctrl+C`: 退出

**updateLogViewer()**
```go
func (m *Dashboard) updateLogViewer()
```
根据当前激活面板和选中项，更新日志查看器内容。

### 2. TaskListView（任务列表）

位置：`pkg/tui/tasklist.go`

#### 职责
- 渲染任务列表
- 上下导航
- 任务选择
- 状态指示器

#### 关键方法

**Render()**
```go
func (v *TaskListView) Render() string
```
渲染逻辑：
1. 遍历任务列表
2. 为每个任务生成状态图标
3. 高亮选中项
4. 处理滚动（窗口化显示）

**状态图标映射**:
```go
○  TaskStatusPending       (灰色)
●  TaskStatusInProgress    (蓝色)
✓  TaskStatusCompleted     (绿色)
✗  TaskStatusFailed        (红色)
```

**MoveUp() / MoveDown()**
```go
func (v *TaskListView) MoveUp()
func (v *TaskListView) MoveDown()
```
上下移动选中索引，带边界检查。

### 3. AgentGridView（Agent 网格）

位置：`pkg/tui/agentgrid.go`

#### 职责
- 以 3x3 网格显示 Agent
- 2D 导航（上下左右）
- Agent 状态颜色编码
- 当前任务显示

#### 网格布局

```
┌─────────┐ ┌─────────┐ ┌─────────┐
│ Agent-0 │ │ Agent-1 │ │ Agent-2 │
│    ●    │ │    ○    │ │    ?    │
│ Working │ │  Idle   │ │Waiting  │
│ Task-1  │ │   ---   │ │ Task-3  │
└─────────┘ └─────────┘ └─────────┘
┌─────────┐ ┌─────────┐ ┌─────────┐
│ Agent-3 │ │ Agent-4 │ │ Agent-5 │
│    ●    │ │    ✗    │ │    ●    │
│ Working │ │  Error  │ │ Working │
│ Task-5  │ │ Task-2  │ │ Task-6  │
└─────────┘ └─────────┘ └─────────┘
```

#### 导航逻辑

```go
// 上下导航：行切换（±3）
func (v *AgentGridView) MoveUp()    { v.selectedIndex -= v.cols }
func (v *AgentGridView) MoveDown()  { v.selectedIndex += v.cols }

// 左右导航：列切换（±1）
func (v *AgentGridView) MoveLeft()  { v.selectedIndex-- }
func (v *AgentGridView) MoveRight() { v.selectedIndex++ }
```

带边界检查，防止越界。

#### 状态颜色

```go
AgentStateIdle           → 灰色边框
AgentStateWorking        → 蓝色边框
AgentStateWaitingConfirm → 黄色边框
AgentStateError/Stuck    → 红色边框
```

### 4. LogViewerView（日志查看器）

位置：`pkg/tui/logviewer.go`

#### 职责
- 显示选中 Agent 的输出
- 自动滚动到最新日志
- 处理长行截断
- 显示 Agent 状态和任务信息

#### 渲染结构

```
┌─────────────────────────────────┐
│ agent-0                         │  ← Agent ID（粗体、主色）
│ Working                         │  ← 状态（带颜色）
│                                 │
│ Task:                           │  ← 任务描述（如果有）
│ 实现用户认证功能                 │
│                                 │
│ Output:                         │  ← 日志标题
│ Reading file: auth.go           │  ← 日志内容
│ Analyzing code...               │    （滚动窗口）
│ Writing test: auth_test.go      │
│ ...                             │
└─────────────────────────────────┘
```

#### 滚动逻辑

```go
// 计算可用行数
availableLines := v.height - headerLines

// 显示最后 N 行
start := len(lines) - availableLines
if start < 0 { start = 0 }
displayLines := lines[start:]
```

## 样式系统

### Lipgloss 基础

位置：`pkg/tui/styles.go`

#### 样式定义模式

```go
// 1. 定义颜色
colorPrimary := lipgloss.Color("86")  // Cyan

// 2. 创建样式
titleStyle := lipgloss.NewStyle().
    Bold(true).
    Foreground(colorPrimary).
    MarginBottom(1)

// 3. 使用样式
rendered := titleStyle.Render("My Title")
```

### 颜色方案

```go
colorPrimary   = "86"   // Cyan   - 主色（标题、激活边框）
colorSuccess   = "42"   // Green  - 成功（已完成任务）
colorWarning   = "226"  // Yellow - 警告（等待确认）
colorError     = "196"  // Red    - 错误（失败、卡住）
colorIdle      = "240"  // Gray   - 空闲（待处理）
colorWorking   = "39"   // Blue   - 工作中
colorBorder    = "238"  // Dark gray - 普通边框
colorText      = "252"  // Light gray - 普通文本
colorHighlight = "219"  // Pink   - 高亮选中
```

### 布局技巧

#### 水平布局

```go
left := "Left Panel"
right := "Right Panel"
combined := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
```

#### 垂直布局

```go
header := "Header"
body := "Body"
combined := lipgloss.JoinVertical(lipgloss.Left, header, body)
```

#### 对齐和尺寸

```go
centered := lipgloss.NewStyle().
    Width(40).
    Height(10).
    Align(lipgloss.Center, lipgloss.Center).
    Render("Centered Text")
```

## 开发指南

### 添加新面板

假设你想添加第四个面板 "Statistics"：

#### 1. 创建组件文件

**pkg/tui/statistics.go**:
```go
package tui

import "github.com/charmbracelet/lipgloss"

type StatisticsView struct {
    width  int
    height int
    stats  map[string]interface{}
}

func NewStatisticsView(width, height int) *StatisticsView {
    return &StatisticsView{
        width:  width,
        height: height,
        stats:  make(map[string]interface{}),
    }
}

func (v *StatisticsView) Update(stats map[string]interface{}) {
    v.stats = stats
}

func (v *StatisticsView) Render() string {
    // 实现渲染逻辑
    return "Statistics Panel"
}
```

#### 2. 更新 Dashboard

**pkg/tui/dashboard.go**:
```go
// 添加 Pane 枚举
const (
    PaneTasks ActivePane = iota
    PaneAgents
    PaneLogs
    PaneStatistics  // 新增
)

// 添加字段
type Dashboard struct {
    // ... 现有字段
    statisticsView *StatisticsView
}

// 初始化组件
func (m *Dashboard) initializeViews() {
    // ... 现有代码
    m.statisticsView = NewStatisticsView(width, height)
}

// 更新 View() 方法
func (m *Dashboard) View() string {
    // 渲染新面板
    statsPane := m.renderPane("Statistics", m.statisticsView.Render(), ...)

    // 更新布局（例如 2x2 网格）
    topRow := lipgloss.JoinHorizontal(lipgloss.Top, taskPane, agentPane)
    bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, logPane, statsPane)
    content := lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)

    return content
}
```

#### 3. 添加导航逻辑

```go
func (m *Dashboard) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "tab":
        // 循环切换 4 个面板
        m.activePane = (m.activePane + 1) % 4
    // ... 其他按键处理
    }
}
```

### 自定义样式主题

#### 创建主题文件

**pkg/tui/themes.go**:
```go
package tui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
    Primary   lipgloss.Color
    Success   lipgloss.Color
    Warning   lipgloss.Color
    Error     lipgloss.Color
    // ... 其他颜色
}

var (
    DefaultTheme = Theme{
        Primary: lipgloss.Color("86"),
        Success: lipgloss.Color("42"),
        // ...
    }

    DarkTheme = Theme{
        Primary: lipgloss.Color("39"),
        Success: lipgloss.Color("35"),
        // ...
    }
)

func ApplyTheme(theme Theme) {
    colorPrimary = theme.Primary
    colorSuccess = theme.Success
    // ... 重新定义所有样式
}
```

#### 使用主题

```go
// 在 Dashboard.Init() 中
func (m *Dashboard) Init() tea.Cmd {
    ApplyTheme(DarkTheme)  // 应用暗色主题
    // ...
}
```

### 添加交互功能

#### 示例：任务过滤

**1. 添加过滤状态**:
```go
type TaskListView struct {
    // ... 现有字段
    filter TaskStatus  // 过滤器
}
```

**2. 添加过滤逻辑**:
```go
func (v *TaskListView) SetFilter(filter TaskStatus) {
    v.filter = filter
}

func (v *TaskListView) Render() string {
    filteredTasks := v.tasks
    if v.filter != "" {
        filteredTasks = filterTasks(v.tasks, v.filter)
    }
    // ... 渲染过滤后的任务
}
```

**3. 添加快捷键**:
```go
func (m *Dashboard) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "p":  // 显示 pending 任务
        m.taskList.SetFilter(TaskStatusPending)
    case "c":  // 显示 completed 任务
        m.taskList.SetFilter(TaskStatusCompleted)
    case "a":  // 显示所有任务
        m.taskList.SetFilter("")
    // ...
    }
}
```

### 性能优化技巧

#### 1. 避免频繁重新分配

```go
// 不好：每次 Update 都创建新切片
func (v *TaskListView) Update(tasks []*models.Task) {
    v.tasks = make([]*models.Task, len(tasks))
    copy(v.tasks, tasks)
}

// 好：复用现有切片
func (v *TaskListView) Update(tasks []*models.Task) {
    v.tasks = tasks  // 直接引用
}
```

#### 2. 缓存渲染结果

```go
type TaskListView struct {
    // ...
    cachedRender string
    dirty        bool
}

func (v *TaskListView) Update(tasks []*models.Task) {
    v.tasks = tasks
    v.dirty = true  // 标记为脏
}

func (v *TaskListView) Render() string {
    if !v.dirty {
        return v.cachedRender  // 返回缓存
    }

    // 重新渲染
    v.cachedRender = v.renderInternal()
    v.dirty = false
    return v.cachedRender
}
```

#### 3. 限制日志大小

```go
const MaxLogLines = 1000

func (v *LogViewerView) Update(agent *models.AgentStatus) {
    lines := strings.Split(agent.Output, "\n")
    if len(lines) > MaxLogLines {
        // 只保留最后 1000 行
        lines = lines[len(lines)-MaxLogLines:]
        agent.Output = strings.Join(lines, "\n")
    }
    v.agent = agent
}
```

## 调试技巧

### 1. 日志调试

Bubble Tea 应用无法直接使用 `fmt.Println`，需要写入文件：

```go
// 在 dashboard.go 开头
var debugLog *os.File

func init() {
    debugLog, _ = os.OpenFile("/tmp/tui-debug.log",
        os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

func (m *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    fmt.Fprintf(debugLog, "Received message: %T\n", msg)
    // ...
}
```

查看日志：
```bash
tail -f /tmp/tui-debug.log
```

### 2. 状态检查器

创建一个 `debug.go` 文件：

```go
// +build debug

package tui

func (m *Dashboard) DebugString() string {
    return fmt.Sprintf(
        "Active: %v, Tasks: %d, Agents: %d, Selected: %d",
        m.activePane, len(m.tasks), len(m.agents), m.taskList.selectedIndex,
    )
}
```

在 View() 中添加：
```go
func (m *Dashboard) View() string {
    // ...
    debug := m.DebugString()
    return lipgloss.JoinVertical(lipgloss.Left, title, content, help, debug)
}
```

### 3. 性能分析

```bash
# 启用 pprof
go run -tags debug ./cmd/swarm monitor

# 在另一个终端
go tool pprof http://localhost:6060/debug/pprof/heap
```

## 测试

### 单元测试示例

**pkg/tui/tasklist_test.go**:
```go
package tui

import (
    "testing"
    "github.com/yourusername/claude-swarm/internal/models"
)

func TestTaskListView_Navigation(t *testing.T) {
    view := NewTaskListView(80, 20)

    tasks := []*models.Task{
        {ID: "1", Description: "Task 1"},
        {ID: "2", Description: "Task 2"},
        {ID: "3", Description: "Task 3"},
    }
    view.Update(tasks)

    // 初始选中第一项
    if view.selectedIndex != 0 {
        t.Errorf("Expected selectedIndex 0, got %d", view.selectedIndex)
    }

    // 向下移动
    view.MoveDown()
    if view.selectedIndex != 1 {
        t.Errorf("Expected selectedIndex 1, got %d", view.selectedIndex)
    }

    // 向上移动
    view.MoveUp()
    if view.selectedIndex != 0 {
        t.Errorf("Expected selectedIndex 0, got %d", view.selectedIndex)
    }

    // 边界检查：不能向上超出
    view.MoveUp()
    if view.selectedIndex != 0 {
        t.Errorf("Expected selectedIndex to stay at 0, got %d", view.selectedIndex)
    }
}

func TestTaskListView_Render(t *testing.T) {
    view := NewTaskListView(80, 20)

    tasks := []*models.Task{
        {ID: "1", Description: "Test Task", Status: models.TaskStatusPending},
    }
    view.Update(tasks)

    output := view.Render()

    // 检查包含任务描述
    if !strings.Contains(output, "Test Task") {
        t.Error("Rendered output should contain task description")
    }

    // 检查包含状态图标
    if !strings.Contains(output, "○") {
        t.Error("Rendered output should contain pending status icon")
    }
}
```

### 集成测试

**测试整个 Dashboard**:
```go
func TestDashboard_Integration(t *testing.T) {
    // 创建模拟数据
    taskQueue := state.NewTaskQueue("/tmp/test-tasks.json")
    getAgentsFn := func() []*models.AgentStatus {
        return []*models.AgentStatus{
            {AgentID: "agent-0", State: models.AgentStateIdle},
        }
    }

    dashboard := NewDashboard(taskQueue, getAgentsFn)

    // 模拟窗口大小消息
    dashboard, _ = dashboard.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

    // 测试渲染
    output := dashboard.View()
    if output == "" {
        t.Error("Dashboard should render non-empty output")
    }
}
```

## 常见问题

### Q: 如何修改刷新间隔？

**A**: 修改 `dashboard.go` 中的 `tickCmd()` 函数：

```go
func tickCmd() tea.Cmd {
    return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

// 改为 1 秒
func tickCmd() tea.Cmd {
    return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

### Q: 如何支持鼠标交互？

**A**: Bubble Tea 已经启用了鼠标支持（`tea.WithMouseCellMotion()`），你需要处理鼠标消息：

```go
func (m *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        // 处理鼠标点击
        if msg.Type == tea.MouseLeft {
            // 根据坐标确定点击的面板
            // ...
        }
    // ...
    }
}
```

### Q: 如何添加帮助菜单？

**A**: 创建一个新的视图组件：

```go
type HelpView struct {
    visible bool
}

func (v *HelpView) Toggle() {
    v.visible = !v.visible
}

func (v *HelpView) Render() string {
    if !v.visible {
        return ""
    }

    help := `
快捷键帮助：
  Tab       切换面板
  j/k       上下导航
  h/l       左右导航
  q         退出
  ?         显示/隐藏帮助
`

    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Padding(1).
        Render(help)
}
```

在 `Dashboard` 中添加 `?` 快捷键处理：

```go
case "?":
    m.helpView.Toggle()
```

### Q: 如何支持配置文件？

**A**: 创建配置结构和加载函数：

```go
// pkg/tui/config.go
type Config struct {
    RefreshInterval time.Duration `json:"refresh_interval"`
    Theme           string        `json:"theme"`
    GridCols        int           `json:"grid_cols"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return DefaultConfig(), nil  // 使用默认值
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

func DefaultConfig() *Config {
    return &Config{
        RefreshInterval: 2 * time.Second,
        Theme:           "default",
        GridCols:        3,
    }
}
```

## 贡献指南

### 代码风格

1. **遵循 Go 标准**:
   - 运行 `gofmt` 和 `golint`
   - 使用有意义的变量名

2. **注释**:
   - 所有导出函数必须有 godoc 注释
   - 复杂逻辑添加内联注释

3. **错误处理**:
   - 不要忽略错误
   - 使用 `fmt.Errorf` 包装错误

### 提交 PR

1. Fork 仓库
2. 创建特性分支（`git checkout -b feature/my-feature`）
3. 添加测试
4. 提交代码（`git commit -m "Add: my feature"`）
5. 推送分支（`git push origin feature/my-feature`）
6. 创建 Pull Request

## 参考资源

- [Bubble Tea 文档](https://github.com/charmbracelet/bubbletea)
- [Lipgloss 示例](https://github.com/charmbracelet/lipgloss/tree/master/examples)
- [Charm 官方教程](https://charm.sh/docs/)
- [Go TUI 最佳实践](https://github.com/awesome-go/awesome-go#terminal)

## 联系

如有开发相关问题，请：
- 提交 GitHub Issue（技术问题）
- 查看 Discussions（设计讨论）
- 参考 [CONTRIBUTING.md](../CONTRIBUTING.md)

---

**最后更新**: 2026-01-31
**维护者**: Claude Agent Swarm Team
