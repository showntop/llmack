package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/mcp"   // 修改为具名导入
	_ "github.com/showntop/llmack/tool/mcp" // 保持匿名导入用于工具注册
)

func main() {
	log.SetLogger(&log.WrapLogger{})

	fmt.Println("=== llmack MCP (Model Context Protocol) 使用示例 ===")

	// 可选：启用真实MCP连接（默认为false，使用模拟连接）
	// mcp.EnableRealMCP = true
	fmt.Printf("MCP真实连接模式: %v\n", mcp.EnableRealMCP)

	ctx := context.Background()

	// 示例1: 列出所有可用的工具（包括MCP工具）
	fmt.Println("\n1. 列出所有可用工具:")
	showAvailableTools()

	// 示例2: 列出当前连接的MCP服务器
	fmt.Println("\n2. 查看当前MCP服务器状态:")
	listMCPServers(ctx)

	// 示例3: 连接到文件系统MCP服务器
	fmt.Println("\n3. 连接到文件系统MCP服务器:")
	connectFileSystemServer(ctx)
	return
	// 示例4: 连接到SQLite MCP服务器
	fmt.Println("\n4. 连接到SQLite MCP服务器:")
	connectSQLiteServer(ctx)

	// 示例5: 列出连接后的服务器和工具
	fmt.Println("\n5. 查看连接后的服务器状态:")
	listMCPServers(ctx)
	listMCPTools(ctx)

	// 示例6: 调用MCP工具
	fmt.Println("\n6. 调用MCP工具:")
	invokeMCPTools(ctx)

	// 示例7: 使用配置文件管理MCP服务器
	fmt.Println("\n7. 配置文件管理示例:")
	configFileExample(ctx)

	// 示例8: 断开服务器连接
	fmt.Println("\n8. 断开服务器连接:")
	disconnectServers(ctx)

	fmt.Println("\n=== 示例完成 ===")
}

// showAvailableTools 显示所有可用的工具
func showAvailableTools() {
	// 这里我们模拟列出工具，在实际使用中你可能需要通过工具注册表来获取
	availableTools := []string{
		"mcp_manage - MCP服务器管理工具",
		"mcp_invoke - 调用MCP工具",
		"weather - 天气查询工具",
		"search - 搜索工具",
		// ... 其他工具
	}

	for _, toolName := range availableTools {
		fmt.Printf("  - %s\n", toolName)
	}
}

// listMCPServers 列出当前MCP服务器状态
func listMCPServers(ctx context.Context) {
	mcpManageTool := tool.Spawn("mcp_manage")
	if mcpManageTool == nil {
		fmt.Println("  错误: 无法找到 mcp_manage 工具")
		return
	}

	params := map[string]interface{}{
		"action": "list_servers",
	}

	paramsJSON, _ := json.Marshal(params)
	result, err := mcpManageTool.Invoke(ctx, string(paramsJSON))
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
		return
	}

	fmt.Printf("  服务器列表: %s\n", result)
}

// connectFileSystemServer 连接到文件系统MCP服务器
func connectFileSystemServer(ctx context.Context) {
	mcpManageTool := tool.Spawn("mcp_manage")
	if mcpManageTool == nil {
		fmt.Println("  错误: 无法找到 mcp_manage 工具")
		return
	}

	// 准备文件系统服务器配置
	serverConfig := map[string]interface{}{
		"transport":   "stdio",
		"command":     "npx",
		"args":        []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
		"description": "文件系统访问MCP服务器",
	}

	serverConfigJSON, _ := json.Marshal(serverConfig)

	params := map[string]interface{}{
		"action":              "connect",
		"server_name":         "filesystem",
		"server_config":       string(serverConfigJSON),
		"use_real_connection": true,
	}

	paramsJSON, _ := json.Marshal(params)
	result, err := mcpManageTool.Invoke(ctx, string(paramsJSON))
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
		return
	}

	fmt.Printf("  连接结果: %s\n", result)
	time.Sleep(1 * time.Second) // 等待连接完成
}

// connectSQLiteServer 连接到SQLite MCP服务器
func connectSQLiteServer(ctx context.Context) {
	mcpManageTool := tool.Spawn("mcp_manage")
	if mcpManageTool == nil {
		fmt.Println("  错误: 无法找到 mcp_manage 工具")
		return
	}

	// 准备SQLite服务器配置
	serverConfig := map[string]interface{}{
		"transport":   "stdio",
		"command":     "npx",
		"args":        []string{"-y", "@modelcontextprotocol/server-sqlite", "--db-path", "./example.db"},
		"description": "SQLite数据库MCP服务器",
	}

	serverConfigJSON, _ := json.Marshal(serverConfig)

	params := map[string]interface{}{
		"action":        "connect",
		"server_name":   "sqlite",
		"server_config": string(serverConfigJSON),
	}

	paramsJSON, _ := json.Marshal(params)
	result, err := mcpManageTool.Invoke(ctx, string(paramsJSON))
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
		return
	}

	fmt.Printf("  连接结果: %s\n", result)
	time.Sleep(1 * time.Second) // 等待连接完成
}

// listMCPTools 列出MCP工具
func listMCPTools(ctx context.Context) {
	mcpManageTool := tool.Spawn("mcp_manage")
	if mcpManageTool == nil {
		fmt.Println("  错误: 无法找到 mcp_manage 工具")
		return
	}

	params := map[string]interface{}{
		"action": "list_tools",
	}

	paramsJSON, _ := json.Marshal(params)
	result, err := mcpManageTool.Invoke(ctx, string(paramsJSON))
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
		return
	}

	fmt.Printf("  所有工具: %s\n", result)
}

// invokeMCPTools 调用MCP工具
func invokeMCPTools(ctx context.Context) {
	mcpInvokeTool := tool.Spawn("mcp_invoke")
	if mcpInvokeTool == nil {
		fmt.Println("  错误: 无法找到 mcp_invoke 工具")
		return
	}

	// 示例1: 调用文件系统工具
	fmt.Println("  调用文件系统工具:")
	fileParams := map[string]interface{}{
		"server_name": "filesystem",
		"tool_name":   "file_read",
		"arguments":   `{"path": "/tmp/test.txt"}`,
	}

	paramsJSON, _ := json.Marshal(fileParams)
	result, err := mcpInvokeTool.Invoke(ctx, string(paramsJSON))
	if err != nil {
		fmt.Printf("    错误: %v\n", err)
	} else {
		fmt.Printf("    结果: %s\n", result)
	}

	// 示例2: 调用SQLite工具
	fmt.Println("  调用SQLite工具:")
	sqlParams := map[string]interface{}{
		"server_name": "sqlite",
		"tool_name":   "query",
		"arguments":   `{"sql": "SELECT 1 as test"}`,
	}

	paramsJSON, _ = json.Marshal(sqlParams)
	result, err = mcpInvokeTool.Invoke(ctx, string(paramsJSON))
	if err != nil {
		fmt.Printf("    错误: %v\n", err)
	} else {
		fmt.Printf("    结果: %s\n", result)
	}
}

// configFileExample 配置文件管理示例
func configFileExample(ctx context.Context) {
	// 创建示例配置文件
	configDir := "./mcp_config"
	configFile := filepath.Join(configDir, "servers.json")

	// 确保目录存在
	os.MkdirAll(configDir, 0755)

	// 示例配置
	config := map[string]interface{}{
		"servers": map[string]interface{}{
			"filesystem": map[string]interface{}{
				"transport":   "stdio",
				"command":     "npx",
				"args":        []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
				"description": "文件系统访问服务器",
				"enabled":     true,
			},
			"github": map[string]interface{}{
				"transport": "stdio",
				"command":   "npx",
				"args":      []string{"-y", "@modelcontextprotocol/server-github"},
				"env": map[string]string{
					"GITHUB_PERSONAL_ACCESS_TOKEN": "your-token-here",
				},
				"description": "GitHub仓库访问服务器",
				"enabled":     false,
			},
		},
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("  错误: 无法序列化配置: %v\n", err)
		return
	}

	err = os.WriteFile(configFile, configJSON, 0644)
	if err != nil {
		fmt.Printf("  错误: 无法写入配置文件: %v\n", err)
		return
	}

	fmt.Printf("  配置文件已创建: %s\n", configFile)
	fmt.Printf("  配置内容:\n%s\n", string(configJSON))
}

// disconnectServers 断开服务器连接
func disconnectServers(ctx context.Context) {
	mcpManageTool := tool.Spawn("mcp_manage")
	if mcpManageTool == nil {
		fmt.Println("  错误: 无法找到 mcp_manage 工具")
		return
	}

	servers := []string{"filesystem", "sqlite"}

	for _, serverName := range servers {
		params := map[string]interface{}{
			"action":      "disconnect",
			"server_name": serverName,
		}

		paramsJSON, _ := json.Marshal(params)
		result, err := mcpManageTool.Invoke(ctx, string(paramsJSON))
		if err != nil {
			fmt.Printf("  断开 %s 错误: %v\n", serverName, err)
		} else {
			fmt.Printf("  断开 %s: %s\n", serverName, result)
		}
	}
}
