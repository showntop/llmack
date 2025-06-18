package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/showntop/llmack/tool"
	_ "github.com/showntop/llmack/tool/mcp" // 导入MCP工具包
)

// MCPAgent 演示如何在Agent中使用MCP工具
type MCPAgent struct {
	ctx     context.Context
	servers map[string]bool // 跟踪已连接的服务器
}

// NewMCPAgent 创建新的MCP Agent示例
func NewMCPAgent(ctx context.Context) *MCPAgent {
	return &MCPAgent{
		ctx:     ctx,
		servers: make(map[string]bool),
	}
}

// SetupMCPServers 设置MCP服务器连接
func (a *MCPAgent) SetupMCPServers() error {
	fmt.Println("🔧 设置MCP服务器连接...")

	// 连接到文件系统服务器
	if err := a.connectToServer("filesystem", map[string]interface{}{
		"transport":   "stdio",
		"command":     "npx",
		"args":        []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
		"description": "文件系统访问服务器",
	}); err != nil {
		return fmt.Errorf("连接文件系统服务器失败: %v", err)
	}

	// 连接到GitHub服务器（需要配置API Token）
	if err := a.connectToServer("github", map[string]interface{}{
		"transport": "stdio",
		"command":   "npx",
		"args":      []string{"-y", "@modelcontextprotocol/server-github"},
		"env": map[string]string{
			"GITHUB_PERSONAL_ACCESS_TOKEN": "your-token-here", // 实际使用时需要真实token
		},
		"description": "GitHub仓库访问服务器",
	}); err != nil {
		log.Printf("⚠️  连接GitHub服务器失败 (可能需要配置API Token): %v", err)
	}

	return nil
}

// connectToServer 连接到指定的MCP服务器
func (a *MCPAgent) connectToServer(serverName string, config map[string]interface{}) error {
	mcpManageTool := tool.Spawn("mcp_manage")
	if mcpManageTool == nil {
		return fmt.Errorf("无法找到 mcp_manage 工具")
	}

	configJSON, _ := json.Marshal(config)
	params := map[string]interface{}{
		"action":        "connect",
		"server_name":   serverName,
		"server_config": string(configJSON),
	}

	paramsJSON, _ := json.Marshal(params)
	result, err := mcpManageTool.Invoke(a.ctx, string(paramsJSON))
	if err != nil {
		return err
	}

	a.servers[serverName] = true
	fmt.Printf("✅ 成功连接到 %s 服务器: %s\n", serverName, result)
	return nil
}

// ExecuteTask 执行一个具体的任务，展示如何使用MCP工具
func (a *MCPAgent) ExecuteTask(taskName string) error {
	fmt.Printf("\n🚀 执行任务: %s\n", taskName)

	switch taskName {
	case "file_analysis":
		return a.executeFileAnalysisTask()
	case "system_info":
		return a.executeSystemInfoTask()
	case "content_search":
		return a.executeContentSearchTask()
	default:
		return fmt.Errorf("未知任务: %s", taskName)
	}
}

// executeFileAnalysisTask 文件分析任务
func (a *MCPAgent) executeFileAnalysisTask() error {
	fmt.Println("📁 执行文件分析任务...")

	if !a.servers["filesystem"] {
		return fmt.Errorf("文件系统服务器未连接")
	}

	// 1. 列出目录内容
	listResult, err := a.invokeMCPTool("filesystem", "list_directory", map[string]interface{}{
		"path": "/tmp",
	})
	if err != nil {
		log.Printf("列出目录失败: %v", err)
	} else {
		fmt.Printf("📂 目录内容: %s\n", listResult)
	}

	// 2. 读取文件内容（如果存在）
	readResult, err := a.invokeMCPTool("filesystem", "file_read", map[string]interface{}{
		"path": "/tmp/test.txt",
	})
	if err != nil {
		log.Printf("读取文件失败: %v", err)
	} else {
		fmt.Printf("📄 文件内容: %s\n", readResult)
	}

	// 3. 创建分析报告
	analysisReport := a.generateAnalysisReport(listResult, readResult)
	fmt.Printf("📊 分析报告:\n%s\n", analysisReport)

	return nil
}

// executeSystemInfoTask 系统信息收集任务
func (a *MCPAgent) executeSystemInfoTask() error {
	fmt.Println("💻 执行系统信息收集任务...")

	// 收集可用的MCP服务器信息
	serverInfo, err := a.getMCPServerInfo()
	if err != nil {
		return fmt.Errorf("获取服务器信息失败: %v", err)
	}

	fmt.Printf("🔍 MCP服务器信息:\n%s\n", serverInfo)

	// 收集可用工具信息
	toolsInfo, err := a.getMCPToolsInfo()
	if err != nil {
		return fmt.Errorf("获取工具信息失败: %v", err)
	}

	fmt.Printf("🛠️ 可用工具信息:\n%s\n", toolsInfo)

	return nil
}

// executeContentSearchTask 内容搜索任务
func (a *MCPAgent) executeContentSearchTask() error {
	fmt.Println("🔍 执行内容搜索任务...")

	// 在文件系统中搜索特定内容
	if a.servers["filesystem"] {
		searchResult, err := a.invokeMCPTool("filesystem", "search_files", map[string]interface{}{
			"pattern": "*.txt",
			"path":    "/tmp",
		})
		if err != nil {
			log.Printf("文件搜索失败: %v", err)
		} else {
			fmt.Printf("🔎 搜索结果: %s\n", searchResult)
		}
	}

	// 如果GitHub服务器可用，搜索代码
	if a.servers["github"] {
		codeSearchResult, err := a.invokeMCPTool("github", "search_code", map[string]interface{}{
			"query": "function main",
			"repo":  "owner/repository",
		})
		if err != nil {
			log.Printf("代码搜索失败: %v", err)
		} else {
			fmt.Printf("💻 代码搜索结果: %s\n", codeSearchResult)
		}
	}

	return nil
}

// invokeMCPTool 调用MCP工具的便利方法
func (a *MCPAgent) invokeMCPTool(serverName, toolName string, arguments map[string]interface{}) (string, error) {
	mcpInvokeTool := tool.Spawn("mcp_invoke")
	if mcpInvokeTool == nil {
		return "", fmt.Errorf("无法找到 mcp_invoke 工具")
	}

	argumentsJSON, _ := json.Marshal(arguments)
	params := map[string]interface{}{
		"server_name": serverName,
		"tool_name":   toolName,
		"arguments":   string(argumentsJSON),
	}

	paramsJSON, _ := json.Marshal(params)
	return mcpInvokeTool.Invoke(a.ctx, string(paramsJSON))
}

// getMCPServerInfo 获取MCP服务器信息
func (a *MCPAgent) getMCPServerInfo() (string, error) {
	mcpManageTool := tool.Spawn("mcp_manage")
	if mcpManageTool == nil {
		return "", fmt.Errorf("无法找到 mcp_manage 工具")
	}

	params := map[string]interface{}{
		"action": "list_servers",
	}

	paramsJSON, _ := json.Marshal(params)
	return mcpManageTool.Invoke(a.ctx, string(paramsJSON))
}

// getMCPToolsInfo 获取MCP工具信息
func (a *MCPAgent) getMCPToolsInfo() (string, error) {
	mcpManageTool := tool.Spawn("mcp_manage")
	if mcpManageTool == nil {
		return "", fmt.Errorf("无法找到 mcp_manage 工具")
	}

	params := map[string]interface{}{
		"action": "list_tools",
	}

	paramsJSON, _ := json.Marshal(params)
	return mcpManageTool.Invoke(a.ctx, string(paramsJSON))
}

// generateAnalysisReport 生成分析报告
func (a *MCPAgent) generateAnalysisReport(listResult, readResult string) string {
	report := "=== 文件分析报告 ===\n"
	report += fmt.Sprintf("目录列表结果: %s\n", listResult)
	report += fmt.Sprintf("文件读取结果: %s\n", readResult)
	report += "建议: 基于MCP工具的分析结果，系统运行正常\n"
	return report
}

// Cleanup 清理资源
func (a *MCPAgent) Cleanup() {
	fmt.Println("\n🧹 清理MCP连接...")

	mcpManageTool := tool.Spawn("mcp_manage")
	if mcpManageTool == nil {
		log.Println("无法找到 mcp_manage 工具进行清理")
		return
	}

	for serverName := range a.servers {
		params := map[string]interface{}{
			"action":      "disconnect",
			"server_name": serverName,
		}

		paramsJSON, _ := json.Marshal(params)
		result, err := mcpManageTool.Invoke(a.ctx, string(paramsJSON))
		if err != nil {
			log.Printf("断开 %s 连接失败: %v", serverName, err)
		} else {
			fmt.Printf("✅ 成功断开 %s: %s\n", serverName, result)
		}
	}
}

// RunAgentExample 运行Agent示例
func RunAgentExample() {
	fmt.Println("=== MCP Agent 使用示例 ===")

	ctx := context.Background()
	agent := NewMCPAgent(ctx)

	// 设置MCP服务器
	if err := agent.SetupMCPServers(); err != nil {
		log.Fatalf("设置MCP服务器失败: %v", err)
	}

	// 执行各种任务
	tasks := []string{"file_analysis", "system_info", "content_search"}

	for _, task := range tasks {
		if err := agent.ExecuteTask(task); err != nil {
			log.Printf("执行任务 %s 失败: %v", task, err)
		}
		fmt.Println()
	}

	// 清理资源
	agent.Cleanup()

	fmt.Println("=== Agent示例完成 ===")
}
