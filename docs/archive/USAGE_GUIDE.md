# Claude Swarm 使用技巧和最佳实践

## 📚 目录

1. [如何写好需求描述](#如何写好需求描述)
2. [审批环节技巧](#审批环节技巧)
3. [完整工作流程](#完整工作流程)
4. [实战场景示例](#实战场景示例)
5. [常见问题和避坑指南](#常见问题和避坑指南)
6. [性能优化技巧](#性能优化技巧)

---

## 如何写好需求描述

### ✅ 好的需求描述特征

**1. 明确具体，包含核心功能**
```bash
# ✅ 好例子
swarm orchestrate "实现一个Todo应用，包括：添加任务、标记完成、删除任务、任务持久化到文件"

# ❌ 差例子
swarm orchestrate "做一个Todo应用"  # 太笼统
```

**2. 说明技术栈或实现方式**
```bash
# ✅ 好例子
swarm orchestrate "用Go实现RESTful API服务器，包括用户注册、登录（JWT认证）、个人资料管理"

# ✅ 更好例子
swarm orchestrate "用Gin框架实现RESTful API，PostgreSQL存储，JWT认证，包括用户注册、登录、个人资料CRUD"
```

**3. 提及关键的非功能需求**
```bash
# ✅ 包含性能、安全等要求
swarm orchestrate "实现文件上传服务，支持图片（jpg/png）最大5MB，上传前压缩，添加访问权限控制"
```

**4. 分层次描述（对复杂需求）**
```bash
# ✅ 结构化描述
swarm orchestrate "电商购物车系统：
核心功能：
- 添加商品到购物车
- 修改商品数量
- 删除商品
- 计算总价（含优惠券）

技术要求：
- Go + Redis存储
- RESTful API
- 单元测试覆盖80%以上
"
```

### 📏 需求描述的黄金法则

| 要素 | 说明 | 示例 |
|------|------|------|
| **What** | 要做什么功能 | "实现用户认证系统" |
| **How** | 用什么技术/方式 | "使用JWT + PostgreSQL" |
| **Why** | 核心目标（可选） | "提供安全的API访问控制" |
| **Constraints** | 限制条件 | "响应时间<100ms，支持并发1000用户" |

---

## 审批环节技巧

### 🔍 审批时检查清单

当AI主脑分析完需求后，你会看到：
```
选项:
  1. ✅ 批准并创建任务队列
  2. ❌ 拒绝（取消创建）
  3. 📝 查看详细信息
```

**建议操作流程：**

#### 第一步：先选 `3` 查看详细信息

查看以下关键点：

**✓ 模块拆分是否合理**
- 模块之间是否低耦合？
- 有没有遗漏的模块？
- 模块粒度是否合适（不要太大也不要太小）？

**✓ 任务描述是否清晰**
- 每个任务是否具体可执行？
- 任务描述是否包含要创建的文件名和函数名？
- 预计时间是否合理（30分钟-2小时）？

**✓ 依赖关系是否正确**
- 依赖链是否形成DAG（无循环依赖）？
- 可并行的任务是否被错误标记为串行？
- 基础任务（如配置、工具函数）是否优先级最高？

**✓ 文件路径是否合理**
- 文件路径是否符合项目结构？
- 有没有命名冲突？

#### 第二步：根据检查结果决定

**如果满意** → 选 `1` 批准
```bash
请选择 [1/2/3]: 1
✅ 已批准！开始创建任务队列...
```

**如果不满意** → 选 `2` 拒绝，重新调整需求描述
```bash
请选择 [1/2/3]: 2
❌ 已拒绝。
💡 提示：您可以修改需求描述后重新运行 orchestrate
```

然后修改需求描述，重新运行：
```bash
# 根据第一次的分析结果，调整需求描述更明确
swarm orchestrate "实现Todo应用，包括：
1. todo.go: 定义Task结构体和切片存储
2. crud.go: 实现AddTask、DeleteTask、MarkDone函数
3. main.go: CLI命令行交互（cobra框架）
4. storage.go: JSON文件持久化
"
```

### 🚀 快速模式：跳过审批

如果你对需求描述很有信心，可以使用 `--auto-approve`：
```bash
swarm orchestrate --auto-approve "实现简单的HTTP服务器，返回Hello World"
```

⚠️ **注意**：只在简单、低风险的需求时使用，避免AI误解需求导致返工。

---

## 完整工作流程

### 🎯 标准工作流（推荐）

```bash
# 1. 设置Gemini API Key（首次）
export GEMINI_API_KEY="your-api-key"

# 2. AI主脑分析需求
swarm orchestrate "你的需求描述"

# 3. 查看详细分析（选项3）
请选择 [1/2/3]: 3
# 仔细检查模块、任务、依赖

# 4. 批准创建任务队列（选项1）
返回审批选项...
请选择 [1/2/3]: 1

# 5. 启动Agent集群开始工作 (V2)
swarm start-v2 --agents 5

# 6. 实时查看Agent输出
# V2 输出会直接显示在终端
# 或者查看日志文件：tail -f /tmp/swarm-v2-run.log

# 7. 监控任务进度
# 查看任务文件状态
cat ~/.claude-swarm/tasks.json | jq '.tasks[] | {id, status}'

# 8. 查看完成的任务
swarm list

# 9. 集成测试和验证
go test ./...
```

### ⚡ 快速开发流程（简单需求）

```bash
# 一键式：分析 → 批准 → 启动
swarm orchestrate --auto-approve --auto-start --agents 3 "实现Hello World HTTP服务器"

# 等同于：
# 1. orchestrate + 自动批准
# 2. 自动调用 swarm start --agents 3
```

---

## 实战场景示例

### 场景1：从零开发新功能

**需求**：为现有项目添加用户认证功能

```bash
# Step 1: 先用 orchestrate 分析
swarm orchestrate "为现有Go项目添加用户认证：
- 用户注册（邮箱+密码）
- 登录返回JWT token
- Token验证中间件
- 使用现有的PostgreSQL数据库
- 集成到现有的Gin路由
"

# Step 2: 审批时检查
# - 是否正确识别了现有项目结构？
# - 数据库表设计是否合理？
# - 是否考虑了安全性（密码加密、token过期）？

# Step 3: 批准后启动
swarm start --agents 4
```

### 场景2：大型项目拆解

**需求**：开发一个完整的博客系统

```bash
# 策略：分阶段开发

# 阶段1：核心功能
swarm orchestrate "博客系统MVP - 阶段1：
- 文章CRUD（标题、内容、作者）
- 分类管理
- PostgreSQL存储
- RESTful API（Gin）
"

# 等阶段1完成后，再进行阶段2
swarm orchestrate "博客系统 - 阶段2（基于阶段1）：
- 评论系统
- 用户点赞
- 文章搜索
- 添加缓存（Redis）
"
```

### 场景3：Bug修复和优化

```bash
# 场景：发现性能问题

# 1. 先用orchestrate分析优化方案
swarm orchestrate "优化现有API性能：
问题：
- /api/users接口响应慢（>2s）
- 数据库查询N+1问题

目标：
- 响应时间降到<200ms
- 添加Redis缓存
- 优化数据库查询（预加载关联数据）
- 添加性能测试
"

# 2. AI会分析出具体的优化任务
# 3. 审批后让Agent集群并行优化
```

### 场景4：测试和文档补充

```bash
# 为现有代码补充测试
swarm orchestrate "为pkg/auth模块补充完整测试：
- 单元测试（所有函数）
- 集成测试（数据库交互）
- 覆盖率达到80%以上
- 使用testify断言库
"

# 补充文档
swarm orchestrate "为项目补充文档：
- API文档（Swagger格式）
- 架构设计文档
- 部署指南
- 开发者快速上手指南
"
```

---

## 常见问题和避坑指南

### ❌ 问题1：AI拆分的任务太大或太小

**现象**：
- 任务太大：单个任务预计需要4-5小时
- 任务太小：一个函数拆成多个任务

**解决方案**：
```bash
# 在需求描述中明确任务粒度
swarm orchestrate "实现用户管理API：
...（功能描述）

任务拆分要求：
- 每个任务30分钟-2小时完成
- 每个任务对应一个独立的.go文件或一组相关函数
"
```

### ❌ 问题2：依赖关系混乱

**现象**：任务之间相互依赖，无法并行

**解决方案**：
```bash
# 明确指出可并行的模块
swarm orchestrate "电商系统，包含以下独立模块（可并行开发）：
1. 用户模块（独立）
2. 商品模块（独立）
3. 订单模块（依赖用户和商品）
4. 支付模块（依赖订单）
"
```

### ❌ 问题3：AI不理解现有项目结构

**现象**：生成的文件路径不符合项目规范

**解决方案**：
```bash
# 在需求描述中说明项目结构
swarm orchestrate "为现有项目添加缓存层：

现有项目结构：
- cmd/server/main.go (入口)
- internal/service/ (业务逻辑)
- internal/repository/ (数据访问)
- pkg/cache/ (公共缓存包 - 要在这里添加)

要求：
- 在pkg/cache/创建redis.go实现Redis客户端
- 在internal/service/中添加缓存逻辑
"
```

### ❌ 问题4：Gemini API配额超限

**现象**：
```
Error 429: Resource exhausted
```

**解决方案**：
1. 等待配额重置（免费版：每分钟60次）
2. 使用 `--auto-approve` 减少API调用
3. 升级到付费计划

### ❌ 问题5：任务队列创建后想修改

**现象**：批准后发现任务有问题

**解决方案**：
```bash
# 1. 查看当前任务队列
swarm list

# 2. 清空任务队列（如果支持）
# TODO: 未来版本会支持 swarm queue clear

# 当前版本：手动删除任务队列文件
rm ~/.claude-swarm/task_queue.json

# 3. 重新运行orchestrate
swarm orchestrate "修改后的需求描述"
```

---

## 性能优化技巧

### 🚀 提高开发效率

**1. 合理设置Agent数量**

```bash
# 根据任务数量和复杂度调整

# 简单需求（3-5个任务）
swarm start --agents 3

# 中等需求（5-10个任务）
swarm start --agents 5

# 复杂需求（10+个任务）
swarm start --agents 8

# 最大不超过10个（避免冲突）
```

**2. 利用任务并行度**

```bash
# 设计需求时考虑并行性
swarm orchestrate "实现三个独立服务（可并行）：
1. 用户服务（端口8001）
2. 商品服务（端口8002）
3. 订单服务（端口8003）
"

# 而不是串行依赖：
# ❌ "先实现用户服务，基于用户服务实现商品服务，基于商品服务实现订单服务"
```

**3. 任务粒度控制**

| 任务规模 | 预计时间 | Agent数量 | 适用场景 |
|---------|---------|-----------|---------|
| 小 | 30min-1h | 3-5 | 快速原型、简单功能 |
| 中 | 1h-2h | 5-7 | 标准功能开发 |
| 大 | 2h+ | 7-10 | 复杂系统、多模块 |

### 💰 降低成本

**1. 使用自动批准（降低API调用）**
```bash
# 对于熟悉的需求模式
swarm orchestrate --auto-approve "标准CRUD API"
```

**2. 批量处理相似任务**
```bash
# 一次性处理多个相似功能
swarm orchestrate "为5个实体添加CRUD API：用户、商品、订单、评论、收藏"
```

**3. 复用任务模板**
```bash
# 建立需求描述模板文件
cat > templates/rest-api.txt <<EOF
实现 {{EntityName}} 的RESTful API：
- GET /{{entity}} - 列表
- GET /{{entity}}/:id - 详情
- POST /{{entity}} - 创建
- PUT /{{entity}}/:id - 更新
- DELETE /{{entity}}/:id - 删除
使用Gin框架，PostgreSQL存储
EOF

# 使用时替换占位符
sed 's/{{EntityName}}/用户/g; s/{{entity}}/users/g' templates/rest-api.txt | \
  xargs -0 swarm orchestrate
```

---

## 高级技巧

### 🎨 自定义AI主脑行为

编辑 `pkg/orchestrator/brain.go` 中的 prompt 来调整AI行为：

```go
// 例如：强制要求添加测试
func (b *OrchestratorBrain) buildAnalysisPrompt(requirement string) string {
    return fmt.Sprintf(`...

额外要求：
- 每个功能任务必须包含对应的单元测试任务
- 测试覆盖率要达到80%以上
- 使用表驱动测试（table-driven tests）

%s`, requirement)
}
```

### 📊 监控和分析

```bash
# 查看任务执行统计
swarm status --verbose

# 导出任务执行报告（未来功能）
# swarm report --format json > report.json

# 分析Agent效率
# swarm analytics
```

### 🔄 集成到CI/CD

```yaml
# .github/workflows/auto-develop.yml
name: Auto Develop Feature
on:
  issues:
    types: [labeled]

jobs:
  develop:
    if: github.event.label.name == 'auto-develop'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Claude Swarm
        run: |
          export GEMINI_API_KEY=${{ secrets.GEMINI_API_KEY }}
          swarm orchestrate --auto-approve "${{ github.event.issue.title }}"
          swarm start --agents 5
          # 等待完成并创建PR
```

---

## 最佳实践总结

### ✅ DO（推荐做法）

1. **需求描述清晰具体**，包含技术栈和关键约束
2. **先查看详细信息**（选项3），再决定是否批准
3. **合理设置Agent数量**，不是越多越好
4. **分阶段开发**，先MVP再迭代
5. **保持任务独立性**，减少依赖
6. **添加测试要求**，确保代码质量

### ❌ DON'T（避免做法）

1. ❌ 需求描述过于笼统："做一个网站"
2. ❌ 盲目批准，不检查任务拆分
3. ❌ 所有需求一次性开发，没有优先级
4. ❌ 忽略依赖关系，导致任务阻塞
5. ❌ 过度使用自动批准，导致返工
6. ❌ Agent数量过多（>10），增加冲突风险

---

## 下一步

- 🚀 尝试第一个简单需求： `swarm orchestrate "实现Hello World HTTP服务器"`
- 📖 阅读完整文档： [README.md](../README.md)
- 🔧 配置Gemini API： [GEMINI_SETUP.md](./GEMINI_SETUP.md)
- 🐛 遇到问题？提交Issue： [GitHub Issues](https://github.com/Cz07cring/claude-swarm/issues)

---

**Happy Swarming! 🐝✨**
