#!/bin/bash

# 测试实时状态更新的脚本

echo "🧪 测试 Android Agent Platform 实时更新功能"

# 启动服务器（后台运行）
echo "🚀 启动服务器..."
cd "$(dirname "$0")"
go run . &
SERVER_PID=$!

# 等待服务器启动
sleep 3

echo "📡 测试健康检查..."
curl -s http://localhost:8080/health | jq '.'

echo ""
echo "📨 发送测试消息..."

# 发送测试消息
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "我想购买一辆儿童自行车",
    "stream": false
  }')

SESSION_ID=$(echo $RESPONSE | jq -r '.session_id')
echo "✅ 会话创建成功，ID: $SESSION_ID"

echo ""
echo "🔄 监听实时状态更新 (10秒)..."

# 监听流式更新
timeout 10s curl -s http://localhost:8080/api/v1/chat/stream/$SESSION_ID | \
while IFS= read -r line; do
  if [[ $line =~ ^data:\ (.*)$ ]]; then
    echo "📥 收到更新: ${BASH_REMATCH[1]}" | jq '.'
  fi
done

echo ""
echo "🛑 停止服务器..."
kill $SERVER_PID

echo "✅ 测试完成！" 