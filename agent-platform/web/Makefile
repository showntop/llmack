# AI Agent Platform - Web Frontend Makefile

.PHONY: help install dev build clean lint test

# 默认目标
help:
	@echo "AI Agent Platform - 前端应用管理"
	@echo ""
	@echo "可用命令:"
	@echo "  install     - 安装依赖"
	@echo "  dev         - 启动开发服务器"
	@echo "  build       - 构建生产版本"
	@echo "  preview     - 预览生产版本"
	@echo "  lint        - 代码检查"
	@echo "  test        - 运行测试"
	@echo "  clean       - 清理构建文件"

# 安装依赖
install:
	@echo "📦 安装依赖..."
	npm install

# 启动开发服务器
dev:
	@echo "🚀 启动开发服务器..."
	npm run dev

# 构建生产版本
build:
	@echo "🔨 构建生产版本..."
	npm run build

# 预览生产版本
preview:
	@echo "👀 预览生产版本..."
	npm run preview

# 代码检查
lint:
	@echo "🔍 代码检查..."
	npm run lint

# 运行测试
test:
	@echo "🧪 运行测试..."
	npm test

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -rf node_modules
	rm -rf dist
	rm -f package-lock.json 