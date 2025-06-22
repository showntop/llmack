# AI Agent Platform - 主项目管理 Makefile

.PHONY: help install dev build clean test lint check

# 默认目标
help:
	@echo "🤖 AI Agent Platform - 项目管理"
	@echo "================================="
	@echo ""
	@echo "可用命令:"
	@echo "  install     - 安装所有依赖"
	@echo "  dev         - 启动开发环境 (前后端)"
	@echo "  dev-web     - 仅启动前端开发服务器"
	@echo "  dev-server  - 仅启动后端开发服务器"
	@echo "  build       - 构建所有项目"
	@echo "  test        - 运行所有测试"
	@echo "  lint        - 代码检查和格式化"
	@echo "  clean       - 清理所有构建文件"
	@echo "  check       - 检查环境和依赖"

# 安装所有依赖
install:
	@echo "📦 安装项目依赖..."
	@echo "正在安装前端依赖..."
	cd web && $(MAKE) install
	@echo "正在安装后端依赖..."
	cd server && $(MAKE) install
	@echo "✅ 所有依赖安装完成!"

# 启动开发环境 (需要 concurrently)
dev:
	@echo "🚀 启动完整开发环境..."
	@if command -v concurrently >/dev/null 2>&1; then \
		concurrently \
			--names "🌐WEB,🔧SERVER" \
			--prefix-colors "cyan,yellow" \
			"cd web && $(MAKE) dev" \
			"cd server && $(MAKE) run"; \
	else \
		echo "需要安装 concurrently: npm install -g concurrently"; \
		echo "或者分别运行: make dev-web 和 make dev-server"; \
	fi

# 仅启动前端
dev-web:
	@echo "🌐 启动前端开发服务器..."
	cd web && $(MAKE) dev

# 仅启动后端
dev-server:
	@echo "🔧 启动后端开发服务器..."
	cd server && $(MAKE) run

# 构建所有项目
build:
	@echo "🔨 构建所有项目..."
	@echo "构建前端..."
	cd web && $(MAKE) build
	@echo "构建后端..."
	cd server && $(MAKE) build
	@echo "✅ 构建完成!"

# 运行所有测试
test:
	@echo "🧪 运行所有测试..."
	@echo "运行前端测试..."
	cd web && $(MAKE) test
	@echo "运行后端测试..."
	cd server && $(MAKE) test
	@echo "✅ 测试完成!"

# 代码检查和格式化
lint:
	@echo "🔍 代码检查和格式化..."
	@echo "检查前端代码..."
	cd web && $(MAKE) lint
	@echo "检查后端代码..."
	cd server && $(MAKE) lint
	@echo "✅ 代码检查完成!"

# 清理所有构建文件
clean:
	@echo "🧹 清理所有构建文件..."
	cd web && $(MAKE) clean
	cd server && $(MAKE) clean
	@echo "✅ 清理完成!"

# 检查环境和依赖
check:
	@echo "🔍 检查开发环境..."
	@echo "检查 Node.js..."
	@command -v node >/dev/null 2>&1 || { echo "❌ 需要安装 Node.js"; exit 1; }
	@echo "✅ Node.js: $(shell node --version)"
	@echo "检查 npm..."
	@command -v npm >/dev/null 2>&1 || { echo "❌ 需要安装 npm"; exit 1; }
	@echo "✅ npm: $(shell npm --version)"
	@echo "检查 Go..."
	@command -v go >/dev/null 2>&1 || { echo "❌ 需要安装 Go"; exit 1; }
	@echo "✅ Go: $(shell go version | cut -d' ' -f3)"
	@echo "检查 make..."
	@command -v make >/dev/null 2>&1 || { echo "❌ 需要安装 make"; exit 1; }
	@echo "✅ make: $(shell make --version | head -1)"
	@echo ""
	@echo "🎉 开发环境检查完成! 所有依赖都已安装"

# 项目信息
info:
	@echo "📊 AI Agent Platform 项目信息"
	@echo "================================="
	@echo "前端技术栈:"
	@echo "  - React 17 + TypeScript"
	@echo "  - Vite (构建工具)"
	@echo "  - Tailwind CSS (样式框架)"
	@echo ""
	@echo "后端技术栈:"
	@echo "  - Go (服务器语言)"
	@echo "  - net/http (HTTP服务器)"
	@echo "  - Server-Sent Events (实时通信)"
	@echo ""
	@echo "端口配置:"
	@echo "  - 前端: http://localhost:3000"
	@echo "  - 后端: http://localhost:8080"
	@echo ""
	@echo "开发流程:"
	@echo "  1. make install  # 安装依赖"
	@echo "  2. make dev      # 启动开发环境"
	@echo "  3. 打开浏览器访问 http://localhost:3000"

# 安装全局依赖
setup:
	@echo "🛠️ 安装全局开发依赖..."
	npm install -g concurrently
	@echo "✅ 全局依赖安装完成!" 