# MCP (Model Context Protocol) 工具

这个工具包实现了对 MCP (Model Context Protocol) 的支持，允许 llmack 框架连接到外部 MCP 服务器并使用其提供的工具和资源。

## 功能特性

- 连接到多个 MCP 服务器
- 支持 stdio、SSE 和 HTTP 传输协议
- 动态发现和注册 MCP 服务器提供的工具
- 工具调用和结果处理
- 配置文件管理

## 已注册的工具

### 1. `mcp_manage` - MCP 服务器管理
管理 MCP 服务器连接的主要工具。

**参数：**
- `action` (必需): 要执行的操作
  - `connect`: 连接到 MCP 服务器
  - `disconnect`: 断开 MCP 服务器连接
  - `list_servers`: 列出所有服务器
  - `list_tools`: 列出可用工具
- `server_name` (可选): MCP 服务器名称（connect/disconnect 操作必需）
- `server_config` (可选): JSON 格式的服务器配置（connect 操作必需）

### 2. `mcp_invoke` - 调用 MCP 工具
调用连接的 MCP 服务器上的工具。

**参数：**
- `server_name` (必需): MCP 服务器名称
- `tool_name` (必需): 要调用的工具名称
- `arguments` (可选): JSON 格式的工具参数

## 使用示例

### 1. 连接到文件系统 MCP 服务器

```json
{
  "action": "connect",
  "server_name": "filesystem",
  "server_config": "{\"transport\":\"stdio\",\"command\":\"npx\",\"args\":[\"-y\",\"@modelcontextprotocol/server-filesystem\",\"/tmp\"],\"description\":\"File system access MCP server\"}"
}
```

### 2. 列出所有连接的服务器

```json
{
  "action": "list_servers"
}
```

### 3. 列出特定服务器的工具

```json
{
  "action": "list_tools",
  "server_name": "filesystem"
}
```

### 4. 调用文件读取工具

```json
{
  "server_name": "filesystem",
  "tool_name": "file_read",
  "arguments": "{\"path\":\"/tmp/example.txt\"}"
}
```

## 服务器配置格式

MCP 服务器配置是一个 JSON 对象，包含以下字段：

```json
{
  "transport": "stdio|sse|http",
  "command": "命令路径（stdio 传输必需）",
  "args": ["命令参数列表"],
  "env": {"环境变量": "值"},
  "url": "服务器 URL（sse/http 传输必需）",
  "description": "服务器描述"
}
```

## 支持的 MCP 服务器示例

### 1. 文件系统服务器
```json
{
  "transport": "stdio",
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/directory"],
  "description": "提供文件系统访问功能"
}
```

### 2. SQLite 数据库服务器
```json
{
  "transport": "stdio",
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-sqlite", "--db-path", "./database.db"],
  "description": "SQLite 数据库访问"
}
```

### 3. GitHub 服务器
```json
{
  "transport": "stdio",
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-github"],
  "env": {
    "GITHUB_PERSONAL_ACCESS_TOKEN": "your-token-here"
  },
  "description": "GitHub 仓库访问"
}
```

### 4. Brave 搜索服务器
```json
{
  "transport": "stdio",
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-brave-search"],
  "env": {
    "BRAVE_API_KEY": "your-api-key-here"
  },
  "description": "Brave 搜索引擎"
}
```

## 配置文件

系统支持通过配置文件管理多个 MCP 服务器。配置文件格式：

```json
{
  "servers": {
    "filesystem": {
      "transport": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
      "description": "File system access MCP server",
      "enabled": true
    },
    "sqlite": {
      "transport": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite", "--db-path", "./database.db"],
      "description": "SQLite database MCP server",
      "enabled": false
    }
  }
}
```

## 动态工具注册

当连接到 MCP 服务器时，系统会自动：

1. 发现服务器提供的所有工具
2. 为每个工具创建对应的 llmack 工具
3. 使用命名格式：`mcp_{server_name}_{tool_name}`
4. 在工具描述中添加 `[MCP:server_name]` 前缀

例如，连接到名为 "filesystem" 的服务器后，其 "read_file" 工具会被注册为：
- 工具名称：`mcp_filesystem_read_file`
- 描述：`[MCP:filesystem] Read the contents of a file`

## 错误处理

系统提供详细的错误信息，包括：

- 连接失败原因
- 工具调用错误
- 配置验证错误
- 服务器通信问题

## 传输协议支持

目前实现包括：

1. **stdio**: 通过标准输入/输出与 MCP 服务器通信（已实现基础框架）
2. **sse**: Server-Sent Events 协议（待实现）
3. **http**: HTTP REST API（待实现）

## 依赖要求

- Node.js (用于运行 npm/npx 命令)
- 相应的 MCP 服务器包（通过 npm 安装）

## 开发说明

当前实现包含了完整的框架结构，但 MCP 协议的实际通信部分使用了模拟实现。要完成真实的 MCP 通信，需要：

1. 实现 JSON-RPC 2.0 消息传输
2. 处理 MCP 协议的初始化握手
3. 实现工具列表获取和调用
4. 添加资源和提示符支持
5. 实现错误处理和重连机制

相关代码位置：
- `client.go`: MCP 客户端协议实现
- `mcp.go`: 主要工具注册和管理
- `config.go`: 配置文件处理

这为将来集成真实的 MCP 库（如 mark3labs/mcp-go）提供了清晰的集成点。 