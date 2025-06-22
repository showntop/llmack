#!/bin/bash

# AI Agent Platform 快速启动脚本
echo "🤖 AI Agent Platform 启动脚本"
echo "================================="

# 检查依赖
echo "🔍 检查环境依赖..."
if ! command -v node >/dev/null 2>&1; then
    echo "❌ 需要安装 Node.js"
    exit 1
fi

if ! command -v go >/dev/null 2>&1; then
    echo "❌ 需要安装 Go"
    exit 1
fi

echo "✅ 环境检查通过"

# 安装依赖
echo ""
echo "📦 安装依赖..."
if [ ! -d "web/node_modules" ]; then
    echo "安装前端依赖..."
    cd web && npm install && cd ..
fi

if [ ! -f "server/go.sum" ]; then
    echo "安装后端依赖..."
    cd server && go mod tidy && cd ..
fi

echo "✅ 依赖安装完成"

# 启动服务
echo ""
echo "🚀 启动服务..."

# 检查是否安装了 concurrently
if command -v concurrently >/dev/null 2>&1; then
    echo "使用 concurrently 同时启动前后端..."
    concurrently \
        --names "🌐WEB,🔧SERVER" \
        --prefix-colors "cyan,yellow" \
        "cd web && npm run dev" \
        "cd server && go run main.go"
else
    echo "❗ 未安装 concurrently，将分别启动服务"
    echo "请在另一个终端运行以下命令:"
    echo "  cd agent-platform/web && npm run dev"
    echo ""
    echo "现在启动后端服务器..."
    cd server && go run main.go
fi 