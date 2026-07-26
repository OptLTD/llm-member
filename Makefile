.DEFAULT_GOAL := help

APP      := member
IMAGE    ?= optltd/member
TAG      ?= latest
PORT     ?= 8080
BIN      := bin/$(APP)
UI_DIR   := ui-admin
GO       ?= go
DOCKER   ?= docker
COMPOSE  ?= docker compose

.PHONY: help setup deps dev dev-api dev-admin build build-admin run test \
        image image-push compose-up compose-down compose-logs clean

help: ## 显示可用命令
	@echo "用法: make <target>"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

setup: .env deps ## 首次初始化（复制 .env + 安装依赖）

.env:
	@test -f .env || cp .env.example .env
	@echo "已创建 .env，请按需修改配置"

deps: ## 安装 Go / 前端依赖
	$(GO) mod download
	cd $(UI_DIR) && npm ci

dev: ## 开发模式：API :8080 + 管理端 Vite :5173
	@trap 'kill 0' INT TERM; \
	$(GO) run . & \
	(cd $(UI_DIR) && npm run dev) & \
	wait

dev-api: ## 仅启动 Go API（:8080）
	$(GO) run .

dev-admin: ## 仅启动管理端 Vite 开发服务器（需 API 已运行）
	cd $(UI_DIR) && npm run dev

build-admin: ## 构建管理端到 ui-front/admin/dist
	cd $(UI_DIR) && npm run build

build: build-admin ## 构建二进制（含 embed 静态资源）
	@mkdir -p bin
	CGO_ENABLED=1 $(GO) build -o $(BIN) .

run: ## 直接 go run（不预编译管理端）
	$(GO) run .

test: ## 运行 Go 测试
	$(GO) test ./...

image: ## 构建 Docker 镜像（默认 optltd/member:latest）
	$(DOCKER) build -t $(IMAGE):$(TAG) .

image-push: image ## 构建并推送 Docker 镜像
	$(DOCKER) push $(IMAGE):$(TAG)

compose-up: ## docker compose 启动（postgres + redis + app）
	$(COMPOSE) up -d

compose-down: ## 停止 docker compose
	$(COMPOSE) down

compose-logs: ## 查看 compose 日志
	$(COMPOSE) logs -f

clean: ## 清理构建产物
	rm -rf bin/
	rm -f main $(APP)
