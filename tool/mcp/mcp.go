package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/showntop/llmack/tool"
)

// EnableRealMCP controls whether to use real MCP connections or simulated ones
var EnableRealMCP = false

const Name = "mcp"

// MCPClient represents an MCP client connection
type MCPClient struct {
	mu          sync.RWMutex
	servers     map[string]*MCPServerConnection
	clients     map[string]*MCPProtocolClient // 存储活跃的客户端连接
	clientTools map[string]*tool.Tool
}

// MCPServerConnection represents a connection to an MCP server
type MCPServerConnection struct {
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Transport string            `json:"transport"` // "stdio", "sse", "http"
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Connected bool              `json:"connected"`
	Tools     []MCPTool         `json:"tools,omitempty"`
}

// MCPTool represents a tool available on an MCP server
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	ServerName  string                 `json:"serverName"`
}

// MCPRequest represents a request to call an MCP tool
type MCPRequest struct {
	ServerName string                 `json:"server_name"`
	ToolName   string                 `json:"tool_name"`
	Arguments  map[string]interface{} `json:"arguments"`
}

var mcpClient *MCPClient

func init() {
	mcpClient = &MCPClient{
		servers:     make(map[string]*MCPServerConnection),
		clients:     make(map[string]*MCPProtocolClient),
		clientTools: make(map[string]*tool.Tool),
	}

	// Register the main MCP management tool
	manageTool := tool.New(
		tool.WithName("mcp_manage"),
		tool.WithDescription("Manage MCP server connections - connect, disconnect, list servers and tools"),
		tool.WithParameters(
			tool.Parameter{
				Name:          "action",
				Type:          tool.String,
				Required:      true,
				LLMDescrition: "Action to perform: 'connect', 'disconnect', 'list_servers', 'list_tools'",
			},
			tool.Parameter{
				Name:          "server_name",
				Type:          tool.String,
				Required:      false,
				LLMDescrition: "Name of the MCP server (required for connect/disconnect)",
			},
			tool.Parameter{
				Name:          "server_config",
				Type:          tool.String,
				Required:      false,
				LLMDescrition: "JSON configuration for MCP server connection (required for connect)",
			},
		),
		tool.WithFunction(handleMCPManage),
	)
	tool.Register(manageTool)

	// Register the MCP tool invocation tool
	invokeTool := tool.New(
		tool.WithName("mcp_invoke"),
		tool.WithDescription("Invoke a tool on a connected MCP server"),
		tool.WithParameters(
			tool.Parameter{
				Name:          "server_name",
				Type:          tool.String,
				Required:      true,
				LLMDescrition: "Name of the MCP server",
			},
			tool.Parameter{
				Name:          "tool_name",
				Type:          tool.String,
				Required:      true,
				LLMDescrition: "Name of the tool to invoke",
			},
			tool.Parameter{
				Name:          "arguments",
				Type:          tool.String,
				Required:      false,
				LLMDescrition: "JSON string of arguments for the tool",
			},
		),
		tool.WithFunction(handleMCPInvoke),
	)
	tool.Register(invokeTool)
}

func handleMCPManage(ctx context.Context, args string) (string, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	action, ok := params["action"].(string)
	if !ok {
		return "", fmt.Errorf("action parameter is required")
	}

	switch action {
	case "connect":
		return handleConnect(params)
	case "disconnect":
		return handleDisconnect(params)
	case "list_servers":
		return handleListServers()
	case "list_tools":
		return handleListTools(params)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func handleConnect(params map[string]interface{}) (string, error) {
	serverName, ok := params["server_name"].(string)
	if !ok {
		return "", fmt.Errorf("server_name parameter is required for connect action")
	}

	serverConfigStr, ok := params["server_config"].(string)
	if !ok {
		return "", fmt.Errorf("server_config parameter is required for connect action")
	}

	var serverConfig MCPServerConnection
	if err := json.Unmarshal([]byte(serverConfigStr), &serverConfig); err != nil {
		return "", fmt.Errorf("failed to parse server config: %v", err)
	}

	serverConfig.Name = serverName

	// Check if we should use real connection or simulation
	useRealConnection, explicitSet := params["use_real_connection"].(bool)
	if !explicitSet {
		// If not explicitly set in params, use global configuration
		useRealConnection = EnableRealMCP
	}

	if useRealConnection {

		ctx := context.Background() // You might want to pass this from the caller
		err := RealConnect(ctx, serverName, &serverConfig)
		if err != nil {
			return "", fmt.Errorf("failed to establish real MCP connection: %v", err)
		}

		// Use real MCP connection
		mcpClient.mu.Lock() // TODO: bugfix 大锁，内部函数获取锁deadlock
		// Store the server configuration first
		mcpClient.servers[serverName] = &serverConfig

		// Register dynamic tools for this server
		finalServerConfig := mcpClient.servers[serverName] // Access while lock is held
		for _, mcpTool := range finalServerConfig.Tools {
			toolName := fmt.Sprintf("mcp_%s_%s", serverName, mcpTool.Name)
			dynamicTool := tool.New(
				tool.WithName(toolName),
				tool.WithDescription(fmt.Sprintf("[MCP:%s] %s", serverName, mcpTool.Description)),
				tool.WithParameters(
					tool.Parameter{
						Name:          "arguments",
						Type:          tool.String,
						Required:      false,
						LLMDescrition: "JSON arguments for the MCP tool",
					},
				),
				tool.WithFunction(func(ctx context.Context, args string) (string, error) {
					return invokeMCPTool(ctx, serverName, mcpTool.Name, args)
				}),
			)
			tool.Register(dynamicTool)
			mcpClient.clientTools[toolName] = dynamicTool
		}
		mcpClient.mu.Unlock()

		// Prepare response using the config we already accessed
		response := map[string]interface{}{
			"status":          "connected",
			"server_name":     serverName,
			"tools_count":     len(finalServerConfig.Tools),
			"tools":           finalServerConfig.Tools,
			"real_connection": useRealConnection,
		}

		responseJSON, _ := json.MarshalIndent(response, "", "  ")
		return string(responseJSON), nil
	} else {
		// Use simulated connection (default behavior)
		mcpClient.mu.Lock()

		// Store the server configuration
		mcpClient.servers[serverName] = &serverConfig

		// Simulate a successful connection
		serverConfig.Connected = true

		// Simulate some example tools being available
		serverConfig.Tools = []MCPTool{
			{
				Name:        "example_tool",
				Description: "An example tool from the MCP server",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type":        "string",
							"description": "Message to process",
						},
					},
					"required": []string{"message"},
				},
				ServerName: serverName,
			},
		}

		// Register dynamic tools for this server
		for _, mcpTool := range serverConfig.Tools {
			toolName := fmt.Sprintf("mcp_%s_%s", serverName, mcpTool.Name)
			dynamicTool := tool.New(
				tool.WithName(toolName),
				tool.WithDescription(fmt.Sprintf("[MCP:%s] %s", serverName, mcpTool.Description)),
				tool.WithParameters(
					tool.Parameter{
						Name:          "arguments",
						Type:          tool.String,
						Required:      false,
						LLMDescrition: "JSON arguments for the MCP tool",
					},
				),
				tool.WithFunction(func(ctx context.Context, args string) (string, error) {
					return invokeMCPTool(ctx, serverName, mcpTool.Name, args)
				}),
			)
			tool.Register(dynamicTool)
			mcpClient.clientTools[toolName] = dynamicTool
		}

		// Prepare response while we still hold the lock
		response := map[string]interface{}{
			"status":          "connected",
			"server_name":     serverName,
			"tools_count":     len(serverConfig.Tools),
			"tools":           serverConfig.Tools,
			"real_connection": useRealConnection,
		}

		mcpClient.mu.Unlock()

		responseJSON, _ := json.MarshalIndent(response, "", "  ")
		return string(responseJSON), nil
	}
}

func handleDisconnect(params map[string]interface{}) (string, error) {
	serverName, ok := params["server_name"].(string)
	if !ok {
		return "", fmt.Errorf("server_name parameter is required for disconnect action")
	}

	mcpClient.mu.Lock()
	defer mcpClient.mu.Unlock()

	server, exists := mcpClient.servers[serverName]
	if !exists {
		return "", fmt.Errorf("server %s is not connected", serverName)
	}

	// Remove dynamic tools for this server
	for toolName := range mcpClient.clientTools {
		if strings.HasPrefix(toolName, fmt.Sprintf("mcp_%s_", serverName)) {
			delete(mcpClient.clientTools, toolName)
		}
	}

	// TODO: Implement actual disconnection logic here
	server.Connected = false
	delete(mcpClient.servers, serverName)

	return fmt.Sprintf("Disconnected from MCP server: %s", serverName), nil
}

func handleListServers() (string, error) {
	mcpClient.mu.RLock()
	defer mcpClient.mu.RUnlock()

	servers := make([]map[string]interface{}, 0)
	for name, server := range mcpClient.servers {
		servers = append(servers, map[string]interface{}{
			"name":       name,
			"url":        server.URL,
			"transport":  server.Transport,
			"connected":  server.Connected,
			"tool_count": len(server.Tools),
		})
	}

	response := map[string]interface{}{
		"servers": servers,
		"count":   len(servers),
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")
	return string(responseJSON), nil
}

func handleListTools(params map[string]interface{}) (string, error) {
	mcpClient.mu.RLock()
	defer mcpClient.mu.RUnlock()

	serverName, ok := params["server_name"].(string)
	if ok {
		// List tools for a specific server
		server, exists := mcpClient.servers[serverName]
		if !exists {
			return "", fmt.Errorf("server %s is not connected", serverName)
		}

		response := map[string]interface{}{
			"server_name": serverName,
			"tools":       server.Tools,
			"count":       len(server.Tools),
		}

		responseJSON, _ := json.MarshalIndent(response, "", "  ")
		return string(responseJSON), nil
	}

	// List all tools from all servers
	allTools := make([]MCPTool, 0)
	for _, server := range mcpClient.servers {
		allTools = append(allTools, server.Tools...)
	}

	response := map[string]interface{}{
		"tools": allTools,
		"count": len(allTools),
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")
	return string(responseJSON), nil
}

func handleMCPInvoke(ctx context.Context, args string) (string, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	serverName, ok := params["server_name"].(string)
	if !ok {
		return "", fmt.Errorf("server_name parameter is required")
	}

	toolName, ok := params["tool_name"].(string)
	if !ok {
		return "", fmt.Errorf("tool_name parameter is required")
	}

	argumentsStr, _ := params["arguments"].(string)
	return invokeMCPTool(ctx, serverName, toolName, argumentsStr)
}

func invokeMCPTool(ctx context.Context, serverName, toolName, argumentsStr string) (string, error) {
	mcpClient.mu.RLock()
	server, exists := mcpClient.servers[serverName]
	mcpClient.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("server %s is not connected", serverName)
	}

	if !server.Connected {
		return "", fmt.Errorf("server %s is not connected", serverName)
	}

	// Find the tool
	var targetTool *MCPTool
	for _, tool := range server.Tools {
		if tool.Name == toolName {
			targetTool = &tool
			break
		}
	}

	if targetTool == nil {
		return "", fmt.Errorf("tool %s not found on server %s", toolName, serverName)
	}

	// Parse arguments
	var arguments map[string]interface{}
	if argumentsStr != "" {
		if err := json.Unmarshal([]byte(argumentsStr), &arguments); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %v", err)
		}
	}

	// Check if we should use real invocation
	if EnableRealMCP {
		// Use real MCP invocation
		result, err := RealInvoke(ctx, serverName, toolName, arguments)
		if err != nil {
			return "", fmt.Errorf("real MCP invocation failed: %v", err)
		}

		// Convert result to JSON string
		resultJSON, _ := json.MarshalIndent(map[string]interface{}{
			"server":    serverName,
			"tool":      toolName,
			"arguments": arguments,
			"result":    result,
			"status":    "success",
			"real_call": true,
		}, "", "  ")
		return string(resultJSON), nil
	}

	// Use simulated response (default behavior)
	response := map[string]interface{}{
		"server":    serverName,
		"tool":      toolName,
		"arguments": arguments,
		"result":    fmt.Sprintf("Simulated response from %s.%s", serverName, toolName),
		"status":    "success",
		"real_call": false,
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")
	return string(responseJSON), nil
}
