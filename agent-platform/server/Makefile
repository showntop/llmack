# AI Agent Platform - Server Makefile

.PHONY: help run build test clean install

# 默认目标
help:
	@echo "AI Agent Platform - 后端服务管理"
	@echo ""
	@echo "可用命令:"
	@echo "  run         - 运行开发服务器"
	@echo "  build       - 构建可执行文件"
	@echo "  test        - 运行测试"
	@echo "  clean       - 清理构建文件"
	@echo "  install     - 安装依赖"

# 运行开发服务器
run:
	@echo "🚀 启动后端服务器..."
	go run main.go

# 构建可执行文件
build:
	@echo "🔨 构建可执行文件..."
	go build -o agent-server main.go

# 安装依赖
install:
	@echo "📦 安装依赖..."
	go mod tidy

# 运行测试
test:
	@echo "🧪 运行测试..."
	go test ./...

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -f agent-server
	go clean -cache

# 运行静态分析
lint:
	@echo "🔍 代码检查..."
	go vet ./...
	go fmt ./...

# 显示项目信息
info:
	@echo "📊 项目信息:"
	@echo "Go版本: $(shell go version)"
	@echo "模块: $(shell go list -m)"
	@echo "依赖: $(shell go list -m all | wc -l) 个包" 