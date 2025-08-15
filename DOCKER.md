# Docker 部署指南

本文档介绍如何使用 Docker 部署 LLM 会员管理系统。

## 文件说明

- `Dockerfile`: 多阶段构建的 Docker 镜像定义文件
- `.dockerignore`: Docker 构建时忽略的文件列表
- `docker-compose.yml`: Docker Compose 配置文件，用于简化部署

## 快速开始

### 方式一：使用 Docker Compose（推荐）

1. **准备环境变量**
   ```bash
   # 复制环境变量模板
   cp .env.example .env
   
   # 编辑 .env 文件，填入实际的 API 密钥
   vim .env
   ```

2. **启动服务**
   ```bash
   # 构建并启动服务
   docker-compose up -d
   
   # 查看服务状态
   docker-compose ps
   
   # 查看日志
   docker-compose logs -f
   ```

3. **访问应用**
   - 应用地址: http://localhost:8080
   - 管理后台: http://localhost:8080/admin.html
   - 默认管理员账号: admin / admin123

### 方式二：使用 Docker 命令

1. **构建镜像**
   ```bash
   docker build -t llm-member:latest .
   ```

2. **运行容器**
   ```bash
   docker run -d \
     --name llm-member-app \
     -p 8080:8080 \
     -v llm_storage:/app/storage \
     -e APP_PORT=8080 \
     -e GIN_MODE=release \
     -e ADMIN_USERNAME=admin \
     -e ADMIN_PASSWORD=admin123 \
     -e OPENAI_API_KEY=your_openai_api_key_here \
     llm-member:latest
   ```

## 配置说明

### 环境变量

主要环境变量说明：

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| APP_PORT | 8080 | 应用端口 |
| GIN_MODE | release | Gin 运行模式 |
| ADMIN_USERNAME | admin | 管理员用户名 |
| ADMIN_PASSWORD | admin123 | 管理员密码 |
| OPENAI_API_KEY | - | OpenAI API 密钥 |
| CLAUDE_API_KEY | - | Claude API 密钥 |

### 数据持久化

- 数据库文件: `/app/storage/app.db`
- 日志文件: `/app/storage/app.log`
- 使用 Docker Volume `llm_storage` 进行数据持久化

## 常用命令

### Docker Compose 命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f

# 进入容器
docker-compose exec llm-member sh

# 更新服务
docker-compose pull
docker-compose up -d
```

### Docker 命令

```bash
# 查看容器状态
docker ps

# 查看日志
docker logs -f llm-member-app

# 进入容器
docker exec -it llm-member-app sh

# 停止容器
docker stop llm-member-app

# 删除容器
docker rm llm-member-app
```

## 安全建议

1. **修改默认密码**: 部署前请修改 `ADMIN_PASSWORD` 环境变量
2. **API 密钥安全**: 不要在代码中硬编码 API 密钥，使用环境变量
3. **网络安全**: 生产环境建议使用反向代理（如 Nginx）
4. **数据备份**: 定期备份 `/app/storage` 目录中的数据

## 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 检查端口占用
   lsof -i :8080
   
   # 修改 docker-compose.yml 中的端口映射
   ports:
     - "8081:8080"  # 将主机端口改为 8081
   ```

2. **权限问题**
   ```bash
   # 检查存储目录权限
   docker-compose exec llm-member ls -la /app/storage
   ```

3. **健康检查失败**
   ```bash
   # 查看健康检查状态
   docker inspect llm-member-app | grep Health -A 10
   ```

### 日志查看

```bash
# 查看应用日志
docker-compose logs llm-member

# 实时查看日志
docker-compose logs -f llm-member

# 查看最近 100 行日志
docker-compose logs --tail=100 llm-member
```

## 生产环境部署

生产环境部署建议：

1. 使用 `GIN_MODE=release`
2. 配置反向代理（Nginx/Traefik）
3. 启用 HTTPS
4. 设置资源限制
5. 配置日志轮转
6. 定期备份数据

示例 Nginx 配置：

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```