# LLM Member - 支持LLM Proxy的会员管理系统

一个专为开发者和企业打造的LLM代理服务平台，提供完整的会员管理体系和灵活的模型接入管理。

## 🎯 产品定位

**面向用户：**
- 👥 **会员注册登录**: 完整的用户注册、登录体系
- 💳 **充值续费**: 支持多种支付方式的充值和套餐续费
- 📊 **使用统计**: 清晰的用量统计和消费记录
- 🔒 **权限管理**: 基于套餐的服务权限控制

**面向管理：**
- 📦 **套餐管理**: 灵活设置会员套餐和计费规则
- 🤖 **模型接入**: 支持多种大模型的统一接入和管理
- 💰 **在线支付**: 集成多种支付方式，支持自动续费
- 👨‍💼 **会员管理**: 完整的会员信息、订单、日志管理
- 📈 **数据分析**: 收入统计、增长分析、运营数据监控

**解决什么问题？**
- ❌ 不用处理不同模型的接口差异
- ❌ 不用担心模型服务的稳定性和切换
- ❌ 不用从零开始构建用户管理和计费系统
- ❌ 不用在业务侧频繁升级调整模型适配
- ❌ 不用担心 API KEY 在业务侧泄漏

## 💡 核心价值

### 🎭 **对用户完全透明的AI服务**
您的用户只需要：
1. 注册账号
2. 选择套餐
3. 开始使用

**用户无需知道：**
- 什么是GPT、Claude、通义千问
- 什么是Token、Temperature、Max Tokens
- 哪个模型适合什么场景
- 如何申请API密钥

### ⚙️ **对管理侧可灵活调整上游模型**
- **智能路由**: 根据负载和成本自动选择最优模型
- **策略配置**: 灵活配置不同场景下的模型使用策略
- **实时切换**: 支持在线调整模型配置，无需重启服务
- **成本控制**: 精细化的成本控制和预算管理

### 🔧 **开箱即用，兼容OpenAI接口**
只要是适配OpenAI的接口，直接接入即可使用，无需任何修改：

```javascript
// 标准OpenAI接口调用方式
fetch('/v1/chat/completions', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer user_token',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    model: 'gpt-3.5-turbo',
    messages: [{
      role: 'user',
      content: '帮我写一份产品介绍'
    }]
  })
})
```

## 🚀 快速开始

### 1. 部署服务

```bash
# 克隆项目
git clone https://github.com/OptLTD/llm-member.git
cd llm-member

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，配置您的模型API密钥

# 启动服务
go run main.go
```

### 2. 配置您的产品

1. **在您的产品中集成LLM Member认证服务：**

增加登录注册功能，用户可以通过LLM Member认证服务进行注册和登录。
Authorization Url: `https://your-domain.com/authorization`

2. **在您的产品中存储用户认证信息：**
在LLM Member 认证成功后，如果设置了`Callback Url`，认证服务会将用户信息返回给您的产品。您需要在您的产品中存储这些信息，以便后续的API调用。

payload:
```json
{
  "token": "sk-xxxxxxxxxxxxxxxx", // 临时 token
  "sign": "sn-xxxxxxxxxxxxxxxxx", // 签名信息
  "time": "2023-12-31T23:59:59Z"  // current time
}
```


2.1 **Web Callback设置**：
Web App Callback Url: `https://your-web-app.com/auth-callback`
实际回调请求为：
`https://your-web-app.com/auth-callback?token=sk-...&sign=...&time=...`

2.2 **Mobile App Callback设置**：
Mobile App Callback Url: `x-you-app://auth-callback`
实际回调请求为：
`x-you-app://auth-callback?token=sk-...&sign=...&time=...`

2.3 **Desktop App Callback设置**：
Mobile App Callback Url: `x-you-app://auth-callback`
实际回调请求为：
`x-you-app://auth-callback?token=sk-...&sign=...&time=...`

2.4 通过`token`获取用户信息：
获取用户信息接口：`https://your-domain.com/v1/verify-token`
```js
const resp = fetch(`https://your-domain.com/v1/verify-token`, {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ${token}',
    'Content-Type': 'application/json'
  },
})
```
响应体：
```json
{
  "email": "user@example.com",
  "username": "User Name",
  "user_plan": "basic",
  "api_token": "sk-xxxxxxxxxxxxxxxx",
  "expire_at": "2023-12-31T23:59:59Z"
}
```

3. **在您的产品中集成LLM Member服务：**

3.1 **请求大模型**：
Base Url：`https://your-domain.com/v1`
实例请求：
```js
// 标准OpenAI接口调用方式
fetch(`${baseURL}/chat/completions`, {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${api_token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    model: 'gpt-3.5-turbo',
    messages: [{
      role: 'user',
      content: '帮我写一份产品介绍'
    }]
  })
})
```

3.2 **获取用户信息**：
获取用户信息接口：`https://your-domain.com/v1/user-profile`
```js
const resp = fetch(`https://your-domain.com/v1/user-profile`, {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${api_token}`,
    'Content-Type': 'application/json'
  },
})
```
响应体：
```json
{
  "email": "user@example.com",
  "username": "User Name",
  "user_plan": "basic",
  "expire_at": "2023-12-31T23:59:59Z"
}
```

3.3 **获取使用统计信息**：
获取用户信息接口：`https://your-domain.com/api/usage`
```js
cosnt resp = fetch(`https://your-domain.com/api/usage`, {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer user_token',
    'Content-Type': 'application/json'
  },
})
```
响应体：
```json
{
  "email": "user@example.com",
  "username": "User Name",
  "user_plan": "basic",
  "expire_at": "2023-12-31T23:59:59Z"
}
```

### 3. 为用户提供服务

用户访问您的产品时：
1. 引导用户注册/登录
2. 展示套餐选择页面
3. 用户选择套餐并支付
4. 立即开始使用AI功能

## 💰 商业模式

### 用户套餐示例

**翻译应用套餐**
| 套餐 | 月费 | 包含服务 | 适用场景 |
|------|------|----------|----------|
| 🌱 **入门版** | ¥99/月 | 10万字符/月 | 个人项目、小型应用 |
| 🚀 **专业版** | ¥299/月 | 50万字符/月 + 优先支持 | 中小企业、成长期产品 |
| 🏢 **企业版** | ¥999/月 | 200万字符/月 + 专属客服 | 大型企业、高并发应用 |
| 🎯 **定制版** | 面议 | 无限制 + 私有部署 | 特殊需求、合规要求 |

**写作应用套餐**
| 套餐 | 月费 | 包含服务 | 适用场景 |
|------|------|----------|----------|
| 🌱 **入门版** | ¥99/月 | 1000次/月 | 个人项目、小型应用 |
| 🚀 **专业版** | ¥299/月 | 5000次/月 + 优先支持 | 中小企业、成长期产品 |
| 🏢 **企业版** | ¥999/月 | 20000次/月 + 专属客服 | 大型企业、高并发应用 |
| 🎯 **定制版** | 面议 | 无限制 + 私有部署 | 特殊需求、合规要求 |

---

**让AI能力触手可及，让您专注于核心业务价值创造。**
