# LLM Member(大模型会员系统)

一个用 Golang 实现的大模型代理服务，支持多个大模型提供商，包含完整的管理界面。

## 功能特性

- 🤖 **多模型支持**: 支持 OpenAI、Claude、通义千问、豆包、智谱清言、Grok、Gemini、OpenRouter、SiliconFlow 等 10+ 大模型提供商
- 🔐 **认证授权**: 基于 Token 的身份认证系统
- 📊 **管理界面**: 现代化的 Web 管理界面
- 📈 **统计分析**: 详细的使用统计和图表展示
- 📝 **请求日志**: 完整的 API 调用历史记录
- 🎛️ **参数配置**: 灵活的模型参数配置
- 💬 **聊天测试**: 内置的聊天测试功能

## 快速开始

### 1. 环境要求

- Go 1.24+
- 现代浏览器

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境变量

复制 `.env.example` 文件为 `.env` 并配置您的 API 密钥：

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置您需要使用的模型 API 密钥。您可以只配置需要使用的模型，不需要配置所有模型。

参考配置示例：
```bash
# 服务器配置
APP_PORT=8080
APP_MODE=debug
DATA_PATH=/data

# 管理员配置
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123

# 只需要配置您要使用的模型 API 密钥
OPENAI_API_KEY=your_openai_api_key_here
CLAUDE_API_KEY=your_claude_api_key_here
QWEN_API_KEY=your_qwen_api_key_here
# ... 其他模型配置
```

### 4. 运行服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

### 5. 访问管理界面

打开浏览器访问 `http://localhost:8080`，使用以下默认凭据登录：

- 用户名: `admin`
- 密码: `admin123`

## API 使用

### 认证

所有 API 请求都需要在 Header 中包含认证 Token：

```
Authorization: Bearer <your_token>
```

### 聊天完成 API

```bash
curl -X POST http://localhost:8080/compatible-v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_token>" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ],
    "temperature": 0.7,
    "max_tokens": 1000
  }'
```

### 获取可用模型

```bash
curl -X GET http://localhost:8080/api/models \
  -H "Authorization: Bearer <your_token>"
```

### 获取统计信息

```bash
curl -X GET http://localhost:8080/api/stats \
  -H "Authorization: Bearer <your_token>"
```

### 获取请求日志

```bash
curl -X GET "http://localhost:8080/api/logs?page=1&page_size=20" \
  -H "Authorization: Bearer <your_token>"
```

## 项目结构

```
llm-member/
├── main.go                 # 主程序入口
├── go.mod                  # Go 模块文件
├── .env                    # 环境变量配置
├── README.md               # 项目说明
├── internal/               # 内部包
│   ├── config/            # 配置管理
│   ├── database/          # 数据库初始化
│   ├── handlers/          # HTTP 处理器
│   ├── middleware/        # 中间件
│   ├── models/           # 数据模型
│   └── services/         # 业务服务
│       ├── auth.go       # 认证服务
│       ├── llm.go        # 大模型服务
│       └── log.go        # 日志服务
└── web/                   # 前端文件
    ├── index.html         # 主页面
    └── app.js            # 前端逻辑
```

## 支持的模型

### OpenAI
- GPT-4
- GPT-4 Turbo
- GPT-4o
- GPT-4o Mini
- GPT-3.5 Turbo

### Claude (Anthropic)
- Claude 3.5 Sonnet
- Claude 3.5 Haiku
- Claude 3 Opus

### 通义千问 (阿里云)
- 通义千问 Turbo
- 通义千问 Plus
- 通义千问 Max
- 通义千问 2.5 72B

### 豆包 (字节跳动)
- 豆包 Pro 4K
- 豆包 Pro 32K
- 豆包 Lite 4K

### 智谱清言 (智谱AI)
- GLM-4
- GLM-4 Plus
- GLM-4 Air
- GLM-4 Flash

### Grok (xAI)
- Grok Beta
- Grok Vision Beta

### Gemini (Google)
- Gemini 1.5 Pro
- Gemini 1.5 Flash
- Gemini Pro

### OpenRouter
- GPT-4o (OpenRouter)
- Claude 3.5 Sonnet (OpenRouter)
- Gemini Pro 1.5 (OpenRouter)

### SiliconFlow
- DeepSeek Chat
- Qwen2.5 72B
- Llama 3.1 405B

### OpenAI-Like (自定义)
- 支持任何兼容 OpenAI API 格式的自定义模型

## 管理界面功能

### 仪表板
- 总请求数统计
- Token 使用统计
- 成功率监控
- 平均响应时间
- 模型使用分布图表
- 提供商使用分布图表

### 聊天测试
- 选择不同模型进行测试
- 调整 Temperature、Max Tokens 等参数
- 实时查看响应结果和统计信息

### 请求日志
- 查看所有 API 请求历史
- 按时间、状态、模型等筛选
- 详细的请求和响应信息

### 模型管理
- 查看所有可用模型
- 模型状态监控
- 提供商信息展示

## 安全注意事项

1. **修改默认密码**: 部署前请务必修改默认的管理员密码
2. **API 密钥保护**: 确保 API 密钥安全存储，不要提交到版本控制
3. **HTTPS**: 生产环境建议使用 HTTPS
4. **防火墙**: 适当配置防火墙规则

## 开发

### 添加新的模型提供商

1. 在 `internal/services/llm.go` 中添加新的提供商支持
2. 更新 `getProviderFromModel` 方法
3. 在 `GetModels` 方法中添加新模型
4. 更新环境变量配置

### 自定义前端

前端使用原生 JavaScript 和 Tailwind CSS 构建，可以根据需要进行定制。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！