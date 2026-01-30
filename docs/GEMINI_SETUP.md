# Gemini API 配置指南

## AI主脑系统使用Gemini API

Claude Agent Swarm v2.0的AI主脑使用Google Gemini API进行智能需求分析和任务拆分。

## 获取Gemini API Key

1. 访问 [Google AI Studio](https://ai.google.dev/)
2. 点击 "Get API Key"
3. 创建或选择项目
4. 复制API Key

## 配置API Key

### 方式1：环境变量（推荐）
```bash
export GEMINI_API_KEY="your-api-key-here"
```

### 方式2：命令行参数
```bash
swarm orchestrate --api-key "your-api-key-here" "你的需求"
```

## 验证配置

测试API Key是否正常工作：
```bash
swarm orchestrate "实现一个Hello World程序"
```

## 模型选择

默认使用 `gemini-1.5-flash-latest` 模型（快速、低成本）。

如果遇到模型不可用的错误，可用的模型包括：
- `gemini-pro`
- `gemini-1.5-pro`
- `gemini-1.5-flash`
- `gemini-1.5-flash-latest`

## 常见问题

### 错误：models/xxx is not found
**原因：** 模型名称不正确或API Key没有该模型的访问权限

**解决方案：**
1. 确认API Key正确
2. 检查Google AI Studio中的配额和权限
3. 尝试使用其他模型名称

### 错误：googleapi: Error 400
**原因：** API请求格式错误

**解决方案：**
1. 检查Go SDK版本：`go list -m github.com/google/generative-ai-go`
2. 更新SDK：`go get -u github.com/google/generative-ai-go`

### 错误：googleapi: Error 429
**原因：** API配额超限

**解决方案：**
1. 等待配额重置
2. 升级到付费计划
3. 使用其他API Key

## API配额

免费版Gemini API限制：
- 每分钟：60次请求
- 每天：1500次请求

如需更高配额，请访问 [Google Cloud Console](https://console.cloud.google.com/)

## 费用估算

Gemini 1.5 Flash定价（截至2024年）：
- 输入：免费（低于一定token数）
- 输出：免费（低于一定token数）

详细定价请查看：[Gemini API Pricing](https://ai.google.dev/pricing)
