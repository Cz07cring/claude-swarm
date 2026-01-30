# 配置文件指南

## 快速开始

1. **复制配置文件模板**
   ```bash
   cp config.yaml.example config.yaml
   ```

2. **编辑配置文件**
   ```bash
   vi config.yaml  # 或使用你喜欢的编辑器
   ```

3. **填入 Gemini API Key**
   ```yaml
   gemini:
     api_key: "your-actual-api-key-here"
   ```

4. **完成！** 现在可以使用配置文件

## 配置优先级

配置加载的优先级（从高到低）：

1. **命令行参数** - 最高优先级
   ```bash
   swarm orchestrate --api-key "xxx" "需求描述"
   ```

2. **配置文件** 
   - 使用 `--config` 指定的路径
   - 当前目录的 `config.yaml`
   - `~/.claude-swarm/config.yaml`

3. **环境变量** - 最低优先级
   ```bash
   export GEMINI_API_KEY="your-key"
   ```

## 配置文件位置

系统会按以下顺序查找配置文件：

1. `--config` 参数指定的路径
2. `./config.yaml` （当前目录）
3. `~/.claude-swarm/config.yaml` （用户主目录）

## 完整配置说明

### Gemini API 配置

```yaml
gemini:
  # Gemini API Key (必填)
  # 获取地址: https://ai.google.dev/
  api_key: "AIzaSy..."
  
  # 使用的模型 (可选，默认: gemini-3-flash-preview)
  model: "gemini-3-flash-preview"
  
  # API 超时时间（秒）(可选，默认: 30)
  timeout: 30
```

### Swarm 配置

```yaml
swarm:
  # 默认启动的 Agent 数量 (可选，默认: 3)
  default_agents: 3
  
  # 监控间隔（秒）(可选，默认: 5)
  monitor_interval: 5
  
  # tmux 会话名称 (可选，默认: claude-swarm)
  session_name: "claude-swarm"
  
  # 任务队列文件路径 (可选)
  task_queue_path: "~/.claude-swarm/tasks.json"
```

### Git 配置

```yaml
git:
  # Git 仓库路径 (可选，默认: 当前目录)
  repo_path: "."
  
  # Worktrees 目录 (可选，默认: .worktrees)
  worktrees_dir: ".worktrees"
  
  # 主分支名称 (可选，默认: main)
  main_branch: "main"
```

## 使用示例

### 示例 1: 使用配置文件

```bash
# 1. 创建配置文件
cp config.yaml.example config.yaml

# 2. 编辑并填入 API Key
vi config.yaml

# 3. 直接使用（自动读取 config.yaml）
swarm orchestrate "实现用户登录功能"
```

### 示例 2: 指定配置文件路径

```bash
# 使用自定义位置的配置文件
swarm orchestrate --config /path/to/my-config.yaml "需求描述"
```

### 示例 3: 命令行覆盖配置

```bash
# 即使配置文件中有 API Key，这里的参数会覆盖它
swarm orchestrate --api-key "temp-key" "需求描述"
```

### 示例 4: 环境变量作为后备

```bash
# 如果没有配置文件，使用环境变量
export GEMINI_API_KEY="your-key"
swarm orchestrate "需求描述"
```

## 安全建议

1. **不要提交配置文件到 Git**
   - `config.yaml` 已在 `.gitignore` 中
   - 只提交 `config.yaml.example` 模板

2. **保护 API Key**
   ```bash
   # 设置文件权限（仅所有者可读写）
   chmod 600 config.yaml
   ```

3. **使用环境变量（生产环境推荐）**
   ```bash
   # 在服务器上使用环境变量而不是配置文件
   export GEMINI_API_KEY="your-key"
   ```

4. **多环境配置**
   ```bash
   # 开发环境
   config.dev.yaml
   
   # 生产环境
   config.prod.yaml
   
   # 使用时指定
   swarm orchestrate --config config.dev.yaml "需求"
   ```

## 故障排除

### 问题: 找不到配置文件

**错误信息**: `⚠️ 配置文件加载失败`

**解决方案**:
1. 检查配置文件是否存在
   ```bash
   ls -la config.yaml
   ```

2. 检查配置文件路径
   ```bash
   swarm orchestrate --config ./config.yaml "需求"
   ```

3. 使用环境变量作为后备
   ```bash
   export GEMINI_API_KEY="your-key"
   ```

### 问题: API Key 无效

**错误信息**: `❌ 请提供Gemini API Key`

**解决方案**:
1. 确认配置文件中有 `api_key`
   ```bash
   grep api_key config.yaml
   ```

2. 确认 API Key 格式正确（以 `AIza` 开头）

3. 测试 API Key
   ```bash
   swarm orchestrate --api-key "your-key" "测试需求"
   ```

### 问题: YAML 格式错误

**错误信息**: `failed to parse config file`

**解决方案**:
1. 检查 YAML 缩进（使用空格，不要用 Tab）
2. 检查引号是否配对
3. 使用 YAML 验证工具
   ```bash
   # 在线验证: https://www.yamllint.com/
   # 或使用命令行工具
   yamllint config.yaml
   ```

## 参考

- [Gemini API 文档](https://ai.google.dev/)
- [YAML 语法指南](https://yaml.org/)
- [项目 README](README.md)
