# MCP 使用示例

这个目录包含了展示如何在 llmack 框架中使用 MCP (Model Context Protocol) 功能的示例代码。

## 文件说明

- `main.go` - 基础MCP功能使用示例
- `agent_example.go` - 在Agent中使用MCP工具的高级示例

## 前置要求

1. **Node.js** - 用于运行MCP服务器
   ```bash
   # 安装Node.js (如果尚未安装)
   # macOS
   brew install node
   
   # Ubuntu/Debian
   sudo apt install nodejs npm
   ```

2. **MCP服务器包** - 通过npm安装需要的MCP服务器
   ```bash
   # 文件系统服务器
   npx -y @modelcontextprotocol/server-filesystem
   
   # SQLite服务器
   npx -y @modelcontextprotocol/server-sqlite
   
   # GitHub服务器 (需要API token)
   npx -y @modelcontextprotocol/server-github
   ```

## 运行示例

### 1. 基础功能示例

```bash
cd example/mcp
go run main.go
```

这个示例会演示：
- 列出可用工具
- 连接到文件系统和SQLite MCP服务器
- 列出连接的服务器和可用工具
- 调用MCP工具
- 创建和使用配置文件
- 断开服务器连接

### 2. Agent使用示例

创建一个运行Agent示例的文件：

```bash
# 创建 run_agent.go
cat > run_agent.go << 'EOF'
package main

import _ "github.com/showntop/llmack/tool/mcp"

func main() {
    RunAgentExample()
}
EOF

# 运行Agent示例
go run agent_example.go run_agent.go
```

Agent示例会演示：
- 在Agent中设置MCP服务器连接
- 执行文件分析任务
- 执行系统信息收集任务
- 执行内容搜索任务
- 自动清理资源

## 示例输出

### 基础功能示例输出

```
=== llmack MCP (Model Context Protocol) 使用示例 ===

1. 列出所有可用工具:
  - mcp_manage - MCP服务器管理工具
  - mcp_invoke - 调用MCP工具
  - weather - 天气查询工具
  - search - 搜索工具

2. 查看当前MCP服务器状态:
  服务器列表: {"servers":[],"count":0}

3. 连接到文件系统MCP服务器:
  连接结果: {"status":"connected","server_name":"filesystem","tools_count":1,"tools":[...]}

4. 连接到SQLite MCP服务器:
  连接结果: {"status":"connected","server_name":"sqlite","tools_count":1,"tools":[...]}

...
```

### Agent示例输出

```
=== MCP Agent 使用示例 ===
🔧 设置MCP服务器连接...
✅ 成功连接到 filesystem 服务器: {...}
⚠️  连接GitHub服务器失败 (可能需要配置API Token): ...

🚀 执行任务: file_analysis
📁 执行文件分析任务...
📂 目录内容: {...}
📄 文件内容: {...}
📊 分析报告:
=== 文件分析报告 ===
...

🚀 执行任务: system_info
💻 执行系统信息收集任务...
🔍 MCP服务器信息: {...}
🛠️ 可用工具信息: {...}

...
```

## 配置说明

### 真实 vs 模拟连接

默认情况下，MCP工具使用模拟连接来演示功能。要使用真实的MCP连接：

1. **全局配置**：在代码中设置
   ```go
   import "github.com/showntop/llmack/tool/mcp"
   
   func main() {
       // 启用真实MCP连接
       mcp.EnableRealMCP = true
       
       // ... 你的代码
   }
   ```

2. **按连接配置**：在连接参数中指定
   ```json
   {
     "action": "connect",
     "server_name": "filesystem",
     "server_config": "{...}",
     "use_real_connection": true
   }
   ```

### 环境变量配置

某些MCP服务器需要API密钥或认证信息：

```bash
# GitHub服务器
export GITHUB_PERSONAL_ACCESS_TOKEN="your_github_token"

# Brave搜索服务器
export BRAVE_API_KEY="your_brave_api_key"

# Google Maps服务器
export GOOGLE_MAPS_API_KEY="your_google_maps_api_key"
```

### 配置文件示例

示例会自动创建配置文件 `./mcp_config/servers.json`：

```json
{
  "servers": {
    "filesystem": {
      "transport": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
      "description": "文件系统访问服务器",
      "enabled": true
    },
    "github": {
      "transport": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "your-token-here"
      },
      "description": "GitHub仓库访问服务器",
      "enabled": false
    }
  }
}
```

## 常见问题

### 1. Node.js命令找不到

确保Node.js已正确安装并在PATH中：
```bash
node --version
npm --version
```

### 2. MCP服务器包安装失败

尝试清理npm缓存：
```bash
npm cache clean --force
```

### 3. 权限错误

确保有足够权限访问指定目录（如`/tmp`）。

### 4. GitHub API限制

如果没有配置GitHub token，GitHub相关功能会失败，这是正常的。

## 扩展示例

你可以基于这些示例创建自己的MCP应用：

1. **添加新的MCP服务器**
2. **创建自定义工具组合**
3. **集成到现有的Agent工作流**
4. **实现配置文件热重载**

## 参考链接

- [MCP官方文档](https://modelcontextprotocol.io/)
- [MCP服务器列表](https://github.com/modelcontextprotocol)
- [llmack框架文档](../README.md) 