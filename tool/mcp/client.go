package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/showntop/llmack/log"
)

// MCPProtocolClient handles the actual MCP protocol communication
type MCPProtocolClient struct {
	serverName string
	transport  string
	cmd        *exec.Cmd
	connected  bool
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	mu         sync.RWMutex

	// For JSON-RPC communication
	requestID       int
	pendingRequests map[int]chan *MCPMessage
	requestMu       sync.RWMutex
}

// MCPMessage represents a JSON-RPC message for MCP protocol
type MCPMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error in MCP protocol
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// InitializeParams represents MCP initialization parameters
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// InitializeResult represents MCP initialization result
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ListToolsResult represents the result of listing tools
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolParams represents parameters for calling a tool
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResult represents the result of calling a tool
type CallToolResult struct {
	Content []interface{}          `json:"content"`
	IsError bool                   `json:"isError,omitempty"`
	Meta    map[string]interface{} `json:"_meta,omitempty"`
}

// NewMCPClient creates a new MCP client for the specified server
func NewMCPClient(serverName string, config *MCPServerConnection) (*MCPProtocolClient, error) {
	client := &MCPProtocolClient{
		serverName:      serverName,
		transport:       config.Transport,
		connected:       false,
		pendingRequests: make(map[int]chan *MCPMessage),
	}

	switch config.Transport {
	case "stdio":
		if config.Command == "" {
			return nil, fmt.Errorf("command is required for stdio transport")
		}
		return client, nil
	case "sse":
		return nil, fmt.Errorf("SSE transport not yet implemented")
	case "http":
		return nil, fmt.Errorf("HTTP transport not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported transport: %s", config.Transport)
	}
}

// Connect establishes connection to the MCP server
func (c *MCPProtocolClient) Connect(ctx context.Context, config *MCPServerConnection) error {
	if c.connected {
		return nil
	}

	switch c.transport {
	case "stdio":
		return c.connectStdio(ctx, config)
	default:
		return fmt.Errorf("unsupported transport: %s", c.transport)
	}
}

// connectStdio connects using stdio transport
func (c *MCPProtocolClient) connectStdio(ctx context.Context, config *MCPServerConnection) error {
	log.InfoContextf(ctx, "[DEBUG] Starting stdio connection for server %s", c.serverName)

	// Start the MCP server process
	args := config.Args
	if args == nil {
		args = []string{}
	}

	log.InfoContextf(ctx, "[DEBUG] Creating command: %s with args: %v", config.Command, args)
	c.cmd = exec.CommandContext(ctx, config.Command, args...)

	// Set environment variables if provided
	if config.Env != nil {
		env := make([]string, 0, len(config.Env))
		for key, value := range config.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		c.cmd.Env = append(c.cmd.Env, env...)
		log.InfoContextf(ctx, "[DEBUG] Set environment variables: %v", env)
	}

	// Set up pipes for communication
	var err error
	log.InfoContextf(ctx, "[DEBUG] Setting up stdin pipe...")
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	log.InfoContextf(ctx, "[DEBUG] Setting up stdout pipe...")
	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	log.InfoContextf(ctx, "[DEBUG] Setting up stderr pipe...")
	c.stderr, err = c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start the process
	log.InfoContextf(ctx, "[DEBUG] Starting MCP server process...")
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start MCP server process: %v", err)
	}
	log.InfoContextf(ctx, "[DEBUG] MCP server process started with PID: %d", c.cmd.Process.Pid)

	// Start reading from stdout in a goroutine
	log.InfoContextf(ctx, "[DEBUG] Starting stdout reader goroutine...")
	go c.readMessages(ctx)

	// Start reading from stderr in a goroutine
	log.InfoContextf(ctx, "[DEBUG] Starting stderr reader goroutine...")
	go c.readErrors(ctx)

	c.connected = true

	log.InfoContextf(ctx, "Connected to MCP server %s with command: %s %v",
		c.serverName, config.Command, args)

	return nil
}

// readMessages reads JSON-RPC messages from stdout
func (c *MCPProtocolClient) readMessages(ctx context.Context) {
	log.InfoContextf(ctx, "[DEBUG] Starting message reader for server %s", c.serverName)
	scanner := bufio.NewScanner(c.stdout)
	for scanner.Scan() {
		line := scanner.Text()
		log.InfoContextf(ctx, "[DEBUG] Received raw message: %s", line)
		if line == "" {
			continue
		}

		var message MCPMessage
		if err := json.Unmarshal([]byte(line), &message); err != nil {
			log.ErrorContextf(ctx, "Failed to parse MCP message: %v, raw: %s", err, line)
			continue
		}

		log.InfoContextf(ctx, "[DEBUG] Parsed message: ID=%v, Method=%s, Error=%v", message.ID, message.Method, message.Error)
		c.handleMessage(ctx, &message)
	}

	if err := scanner.Err(); err != nil {
		log.ErrorContextf(ctx, "Error reading from MCP server stdout: %v", err)
	}
	log.InfoContextf(ctx, "[DEBUG] Message reader finished for server %s", c.serverName)
}

// readErrors reads error messages from stderr
func (c *MCPProtocolClient) readErrors(ctx context.Context) {
	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			log.WarnContextf(ctx, "MCP server stderr: %s", line)
		}
	}
}

// handleMessage handles incoming JSON-RPC messages
func (c *MCPProtocolClient) handleMessage(ctx context.Context, message *MCPMessage) {
	// Handle responses to our requests
	if message.ID != nil {
		var id int
		switch v := message.ID.(type) {
		case float64:
			id = int(v)
		case int:
			id = v
		default:
			log.ErrorContextf(ctx, "Invalid message ID type: %T", message.ID)
			return
		}

		c.requestMu.RLock()
		ch, exists := c.pendingRequests[id]
		c.requestMu.RUnlock()

		if exists {
			select {
			case ch <- message:
			case <-time.After(5 * time.Second):
				log.ErrorContextf(ctx, "Timeout sending response to channel for request %d", id)
			}
		} else {
			log.WarnContextf(ctx, "Received response for unknown request ID: %d", id)
		}
	} else {
		// Handle notifications or other messages
		log.InfoContextf(ctx, "Received notification: %s", message.Method)
	}
}

// sendRequest sends a JSON-RPC request and waits for response
func (c *MCPProtocolClient) sendRequest(ctx context.Context, method string, params interface{}) (*MCPMessage, error) {
	log.InfoContextf(ctx, "[DEBUG] Preparing to send request: method=%s", method)

	c.requestMu.Lock()
	c.requestID++
	id := c.requestID
	responseCh := make(chan *MCPMessage, 1)
	c.pendingRequests[id] = responseCh
	c.requestMu.Unlock()

	log.InfoContextf(ctx, "[DEBUG] Request ID assigned: %d", id)

	// Clean up the pending request when done
	defer func() {
		c.requestMu.Lock()
		delete(c.pendingRequests, id)
		c.requestMu.Unlock()
		log.InfoContextf(ctx, "[DEBUG] Cleaned up pending request %d", id)
	}()

	message := MCPMessage{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	log.InfoContextf(ctx, "[DEBUG] Sending message: %s", string(messageBytes))

	// Send the message
	if _, err := c.stdin.Write(append(messageBytes, '\n')); err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	log.InfoContextf(ctx, "[DEBUG] Message sent, waiting for response...")

	// Wait for response
	select {
	case response := <-responseCh:
		log.InfoContextf(ctx, "[DEBUG] Received response for request %d", id)
		if response.Error != nil {
			return nil, fmt.Errorf("MCP error: %s (code %d)", response.Error.Message, response.Error.Code)
		}
		return response, nil
	case <-time.After(30 * time.Second):
		log.ErrorContextf(ctx, "[DEBUG] Timeout waiting for response to %s (request %d)", method, id)
		return nil, fmt.Errorf("timeout waiting for response to %s", method)
	case <-ctx.Done():
		log.ErrorContextf(ctx, "[DEBUG] Context cancelled while waiting for response to %s (request %d)", method, id)
		return nil, ctx.Err()
	}
}

// Disconnect closes the connection to the MCP server
func (c *MCPProtocolClient) Disconnect(ctx context.Context) error {
	if !c.connected {
		return nil
	}

	// Close pipes
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}

	// Terminate the process
	if c.cmd != nil && c.cmd.Process != nil {
		if err := c.cmd.Process.Kill(); err != nil {
			log.ErrorContextf(ctx, "Failed to kill MCP server process: %v", err)
		}

		// Wait for the process to exit
		c.cmd.Wait()
	}

	c.connected = false
	c.cmd = nil

	log.InfoContextf(ctx, "Disconnected from MCP server %s", c.serverName)
	return nil
}

// Initialize sends the initialize message to the MCP server
func (c *MCPProtocolClient) Initialize(ctx context.Context) (*InitializeResult, error) {
	if !c.connected {
		return nil, fmt.Errorf("client is not connected")
	}

	log.InfoContextf(ctx, "[DEBUG] Sending initialize request to MCP server %s", c.serverName)

	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		ClientInfo: ClientInfo{
			Name:    "llmack",
			Version: "1.0.0",
		},
	}

	log.InfoContextf(ctx, "[DEBUG] Initialize params: %+v", params)

	response, err := c.sendRequest(ctx, "initialize", params)
	if err != nil {
		return nil, fmt.Errorf("initialize request failed: %v", err)
	}

	log.InfoContextf(ctx, "[DEBUG] Received initialize response: %+v", response)

	var result InitializeResult
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", response.Result)), &result); err != nil {
		// Try to parse as JSON
		resultBytes, _ := json.Marshal(response.Result)
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to parse initialize result: %v", err)
		}
	}

	log.InfoContextf(ctx, "Successfully initialized MCP server %s", c.serverName)
	return &result, nil
}

// ListTools requests the list of available tools from the MCP server
func (c *MCPProtocolClient) ListTools(ctx context.Context) ([]Tool, error) {
	if !c.connected {
		return nil, fmt.Errorf("client is not connected")
	}

	response, err := c.sendRequest(ctx, "tools/list", map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("tools/list request failed: %v", err)
	}

	var result ListToolsResult
	resultBytes, _ := json.Marshal(response.Result)
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tools list: %v", err)
	}

	log.InfoContextf(ctx, "Retrieved %d tools from MCP server %s", len(result.Tools), c.serverName)
	return result.Tools, nil
}

// CallTool invokes a specific tool on the MCP server
func (c *MCPProtocolClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*CallToolResult, error) {
	if !c.connected {
		return nil, fmt.Errorf("client is not connected")
	}

	params := CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	}

	response, err := c.sendRequest(ctx, "tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("tools/call request failed: %v", err)
	}

	var result CallToolResult
	resultBytes, _ := json.Marshal(response.Result)
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tool call result: %v", err)
	}

	log.InfoContextf(ctx, "Successfully called tool %s on MCP server %s", toolName, c.serverName)
	return &result, nil
}

// IsConnected returns whether the client is connected to the server
func (c *MCPProtocolClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetServerName returns the name of the connected server
func (c *MCPProtocolClient) GetServerName() string {
	return c.serverName
}

// RealConnect performs the actual connection to MCP server
func RealConnect(ctx context.Context, serverName string, config *MCPServerConnection) error {
	log.InfoContextf(ctx, "[DEBUG] RealConnect started for server %s", serverName)

	client, err := NewMCPClient(serverName, config)
	if err != nil {
		return fmt.Errorf("failed to create MCP client: %v", err)
	}
	log.InfoContextf(ctx, "[DEBUG] MCP client created successfully")

	if err := client.Connect(ctx, config); err != nil {
		return fmt.Errorf("failed to connect to MCP server: %v", err)
	}
	log.InfoContextf(ctx, "[DEBUG] MCP client connected successfully")

	result, err := client.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP connection: %v", err)
	}
	log.InfoContextf(ctx, "[DEBUG] MCP client initialized successfully")

	log.InfoContextf(ctx, "MCP server info: %s v%s", result.ServerInfo.Name, result.ServerInfo.Version)

	// Get the list of tools from the server
	log.InfoContextf(ctx, "[DEBUG] Requesting tools list...")
	tools, err := client.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %v", err)
	}
	log.InfoContextf(ctx, "[DEBUG] Received %d tools from server", len(tools))

	// Convert tools to our internal format
	// Note: The caller should already hold the appropriate lock
	server := mcpClient.servers[serverName]
	if server == nil {
		return fmt.Errorf("server configuration not found for %s", serverName)
	}

	server.Tools = make([]MCPTool, len(tools))
	for i, tool := range tools {
		server.Tools[i] = MCPTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
			ServerName:  serverName,
		}
	}
	server.Connected = true

	// Store the client for later use
	log.InfoContextf(ctx, "[DEBUG] Storing client connection...")
	mcpClient.mu.Lock()
	if mcpClient.clients == nil {
		mcpClient.clients = make(map[string]*MCPProtocolClient)
	}
	mcpClient.clients[serverName] = client
	mcpClient.mu.Unlock()

	log.InfoContextf(ctx, "Successfully connected to MCP server %s with %d tools",
		serverName, len(tools))

	return nil
}

// RealInvoke performs the actual tool invocation on MCP server
func RealInvoke(ctx context.Context, serverName, toolName string, arguments map[string]interface{}) (*CallToolResult, error) {
	mcpClient.mu.RLock()
	client, exists := mcpClient.clients[serverName]
	server, serverExists := mcpClient.servers[serverName]
	mcpClient.mu.RUnlock()

	if !serverExists || !server.Connected {
		return nil, fmt.Errorf("server %s is not connected", serverName)
	}

	if !exists {
		return nil, fmt.Errorf("MCP client for server %s not found", serverName)
	}

	result, err := client.CallTool(ctx, toolName, arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %v", err)
	}

	return result, nil
}
