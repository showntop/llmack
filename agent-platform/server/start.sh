#!/bin/bash

# Android Agent Platform Server 启动脚本

echo "🚀 启动 Android Agent Platform Server..."

# 检查环境文件
if [ ! -f ".env" ]; then
    echo "⚠️  未找到 .env 文件，请创建并配置以下环境变量："
    echo "   doubao_api_key=your_doubao_api_key_here"
    echo "   claude_api_key=your_claude_api_key_here"
    echo "   qwen_api_key=your_qwen_api_key_here"
    echo "   ANDROID_DEVICE_SERIAL=47.96.179.122:1000"
    echo ""
fi

# 检查 Go 版本
echo "📋 检查 Go 版本..."
go version

# 检查 ADB 连接（如果需要 Android 功能）
echo "📱 检查 Android 设备连接..."
if command -v adb &> /dev/null; then
    adb devices
    echo ""
else
    echo "⚠️  ADB 未安装，Android 功能将不可用"
    echo "   请安装 Android SDK Platform Tools"
    echo ""
fi

# 初始化 Go 模块
echo "📦 初始化 Go 模块..."
go mod tidy

# 构建项目
echo "🔨 构建项目..."
go build -o server .

# 启动服务器
echo "🌟 启动服务器..."
echo "   访问地址: http://localhost:8080"
echo "   健康检查: http://localhost:8080/api/v1/health"
echo "   文档说明: 查看 README_ANDROID.md"
echo ""
echo "✨ 支持的功能："
echo "   - 智能任务分析"
echo "   - Android 设备控制 (@android)"
echo "   - 移动应用自动化"
echo "   - 实时状态跟踪"
echo ""

./server 