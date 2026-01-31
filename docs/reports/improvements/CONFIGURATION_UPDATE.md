# 配置系统更新 - 2026-01-30

## 更新概述

将 Gemini API Key 从硬编码/仅环境变量改为支持配置文件管理。

## 修改内容

### 1. 新增文件

#### `config.yaml.example` - 配置文件模板
- 完整的配置示例
- 包含所有可配置项和说明
- 用户需要复制为 `config.yaml` 并填入真实值

#### `pkg/config/config.go` - 配置管理包
- 支持 YAML 配置文件读取
- 多位置查找：指定路径 → `./config.yaml` → `~/.claude-swarm/config.yaml`
- 支持环境变量覆盖
- 提供默认值
- 配置验证

#### `CONFIG_GUIDE.md` - 配置指南文档
- 详细的配置说明
- 使用示例
- 故障排除

#### `CONFIGURATION_UPDATE.md` - 本文档
- 更新说明
- 测试方法

### 2. 修改文件

#### `.gitignore`
```diff
+ # 配置文件（包含敏感信息）
+ config.yaml
```

#### `pkg/orchestrator/brain.go`
```diff
+ import "os"

  func NewOrchestratorBrain(apiKey string, taskQueue *state.TaskQueue) (*OrchestratorBrain, error) {
-     // 初始化Gemini客户端（从环境变量读取API Key）
-     client, err := genai.NewClient(ctx, nil)
+     // 如果没有传入 apiKey，尝试从环境变量读取
+     if apiKey == "" {
+         apiKey = os.Getenv("GEMINI_API_KEY")
+         if apiKey == "" {
+             return nil, fmt.Errorf("gemini API key is required...")
+         }
+     }
+     
+     // 初始化Gemini客户端（使用 apiKey）
+     client, err := genai.NewClient(ctx, &genai.ClientConfig{
+         APIKey: apiKey,
+     })
```

#### `cmd/swarm/orchestrate.go`
```diff
+ import "github.com/yourusername/claude-swarm/pkg/config"

+ var configFilePath string

  func init() {
+     orchestrateCmd.Flags().StringVarP(&configFilePath, "config", "c", "", "配置文件路径")
  }

  func runOrchestrate(cmd *cobra.Command, args []string) {
+     // 加载配置（优先级：命令行参数 > 配置文件 > 环境变量）
+     cfg, err := config.Load(configFilePath)
+     
+     // 获取API Key（命令行参数最高优先级）
+     apiKey := geminiAPIKey
+     if apiKey == "" && cfg != nil {
+         apiKey = cfg.Gemini.APIKey
+     }
+     if apiKey == "" {
+         apiKey = os.Getenv("GEMINI_API_KEY")
+     }
  }
```

### 3. 依赖更新

添加了 YAML 解析库：
```bash
go get gopkg.in/yaml.v3
```

## 配置优先级

从高到低：
1. **命令行参数** `--api-key`
2. **配置文件** `config.yaml`
3. **环境变量** `GEMINI_API_KEY`

## 使用方法

### 方法 1: 配置文件（推荐）

```bash
# 1. 复制模板
cp config.yaml.example config.yaml

# 2. 编辑填入 API Key
vi config.yaml

# 3. 使用
swarm orchestrate "需求描述"
```

### 方法 2: 环境变量（兼容旧方式）

```bash
export GEMINI_API_KEY="your-key"
swarm orchestrate "需求描述"
```

### 方法 3: 命令行参数（临时使用）

```bash
swarm orchestrate --api-key "your-key" "需求描述"
```

## 测试步骤

### 测试 1: 配置文件

```bash
# 创建测试配置
cat > config.yaml << 'YAML'
gemini:
  api_key: "test-key-from-config"
  model: "gemini-3-flash-preview"
YAML

# 验证读取（应该提示使用 config.yaml）
./swarm orchestrate "测试需求" 2>&1 | grep -i config
```

### 测试 2: 环境变量

```bash
# 删除配置文件
rm -f config.yaml

# 设置环境变量
export GEMINI_API_KEY="test-key-from-env"

# 验证读取（应该使用环境变量）
./swarm orchestrate "测试需求" 2>&1 | grep -i env
```

### 测试 3: 命令行参数

```bash
# 命令行参数应该覆盖其他方式
./swarm orchestrate --api-key "test-key-from-cli" "测试需求"
```

### 测试 4: 优先级验证

```bash
# 创建配置文件
echo 'gemini:
  api_key: "config-key"' > config.yaml

# 设置环境变量
export GEMINI_API_KEY="env-key"

# 命令行参数（最高优先级）
./swarm orchestrate --api-key "cli-key" "测试" 2>&1

# 应该使用 "cli-key"
```

## 安全改进

1. ✅ **配置文件在 .gitignore** - 防止提交敏感信息
2. ✅ **仅示例文件入库** - `config.yaml.example`
3. ✅ **支持多种配置方式** - 适应不同环境
4. ✅ **清晰的错误提示** - 告诉用户如何配置

## 向后兼容

- ✅ 环境变量方式仍然有效
- ✅ 命令行参数方式仍然有效
- ✅ 不影响现有用户工作流
- ✅ 配置文件是可选的

## 文档更新

需要更新以下文档：
- [ ] README.md - 添加配置文件说明
- [x] CONFIG_GUIDE.md - 详细配置指南
- [ ] docs/GEMINI_SETUP.md - 更新 API Key 配置方法

## 相关文件

- `config.yaml.example` - 配置模板
- `CONFIG_GUIDE.md` - 配置指南
- `pkg/config/config.go` - 配置管理
- `pkg/orchestrator/brain.go` - API Key 使用
- `cmd/swarm/orchestrate.go` - 配置加载

## 验证清单

- [x] 编译通过
- [x] 配置文件示例已创建
- [x] .gitignore 已更新
- [x] API Key 实际被使用（不再是 nil）
- [x] 支持多种配置方式
- [x] 配置优先级正确
- [x] 文档已创建
- [ ] 实际测试（需要真实 API Key）
