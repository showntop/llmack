# AI Agent Platform

一个现代化的AI代理平台，提供智能对话和任务执行功能。

## 项目结构

```
agent-platform/
├── web/                    # 前端应用 (React + TypeScript + Vite)
│   ├── src/
│   │   ├── components/
│   │   │   └── ChatInterface.tsx
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   └── index.css
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   └── index.html
├── server/                 # 后端服务 (Go)
│   ├── main.go
│   └── go.mod
└── README.md
```

## 功能特性

### 前端 (Web)
- 🎨 **现代化UI**: 基于Tailwind CSS的美观界面
- 💬 **实时聊天**: 类似ChatGPT的对话体验
- 📊 **进度跟踪**: 实时显示AI任务执行步骤
- 📱 **响应式设计**: 支持桌面和移动设备
- ⚡ **实时更新**: Server-Sent Events流式通信

### 后端 (Server)
- 🚀 **高性能**: Go语言原生HTTP服务器
- 🔄 **流式响应**: 支持Server-Sent Events
- 🌐 **CORS支持**: 跨域资源共享
- 📝 **RESTful API**: 标准化API接口
- 🔍 **健康检查**: 服务状态监控

## 快速开始

### 1. 启动后端服务

```bash
cd server
go run main.go
```

服务器将在 `http://localhost:8080` 启动

### 2. 启动前端应用

```bash
cd web
npm install
npm run dev
```

前端应用将在 `http://localhost:3000` 启动

### 3. 访问应用

打开浏览器访问: `http://localhost:3000`

## API 接口

### 健康检查
```bash
GET /health
```

### 聊天接口
```bash
POST /api/v1/chat
Content-Type: application/json

{
  "message": "用户消息",
  "session_id": "会话ID（可选）",
  "stream": true
}
```

### 流式响应
```bash
GET /api/v1/chat/stream/{sessionId}
Accept: text/event-stream
```

### 会话管理
```bash
GET /api/v1/sessions
```

## 技术栈

### 前端
- **React 17** - UI框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Tailwind CSS** - 样式框架

### 后端
- **Go** - 服务器语言
- **net/http** - HTTP服务器
- **encoding/json** - JSON处理

## 开发指南

### 前端开发

1. 安装依赖：
```bash
cd web
npm install
```

2. 启动开发服务器：
```bash
npm run dev
```

3. 构建生产版本：
```bash
npm run build
```

### 后端开发

1. 运行服务器：
```bash
cd server
go run main.go
```

2. 构建可执行文件：
```bash
go build -o agent-server main.go
```

## 部署

### Docker部署（推荐）

```bash
# 构建镜像
docker build -t agent-platform .

# 运行容器
docker run -p 3000:3000 -p 8080:8080 agent-platform
```

### 传统部署

1. 构建前端：
```bash
cd web && npm run build
```

2. 构建后端：
```bash
cd server && go build -o agent-server main.go
```

3. 部署到服务器并配置反向代理

## 环境变量

### 前端
- `VITE_API_URL`: 后端API地址（默认: http://localhost:8080）

### 后端
- `PORT`: 服务端口（默认: 8080）
- `ENV`: 环境模式（development/production）

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 联系方式

- 项目地址: [GitHub](https://github.com/your-username/agent-platform)
- 问题反馈: [Issues](https://github.com/your-username/agent-platform/issues)

---

⭐ 如果这个项目对您有帮助，请给我们一个星星！ 