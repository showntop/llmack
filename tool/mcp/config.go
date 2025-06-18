package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCPConfig represents the configuration for MCP servers
type MCPConfig struct {
	Servers map[string]MCPServerConfig `json:"servers"`
}

// MCPServerConfig represents configuration for a single MCP server
type MCPServerConfig struct {
	Transport   string            `json:"transport"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	URL         string            `json:"url,omitempty"`
	Description string            `json:"description,omitempty"`
	Enabled     bool              `json:"enabled"`
}

// LoadMCPConfig loads MCP configuration from a file
func LoadMCPConfig(configPath string) (*MCPConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// SaveMCPConfig saves MCP configuration to a file
func SaveMCPConfig(config *MCPConfig, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GetDefaultMCPConfig returns a default MCP configuration with example servers
func GetDefaultMCPConfig() *MCPConfig {
	return &MCPConfig{
		Servers: map[string]MCPServerConfig{
			"filesystem": {
				Transport:   "stdio",
				Command:     "npx",
				Args:        []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
				Description: "File system access MCP server",
				Enabled:     false,
			},
			"sqlite": {
				Transport:   "stdio",
				Command:     "npx",
				Args:        []string{"-y", "@modelcontextprotocol/server-sqlite", "--db-path", "./database.db"},
				Description: "SQLite database MCP server",
				Enabled:     false,
			},
			"brave-search": {
				Transport: "stdio",
				Command:   "npx",
				Args:      []string{"-y", "@modelcontextprotocol/server-brave-search"},
				Env: map[string]string{
					"BRAVE_API_KEY": "your-brave-api-key-here",
				},
				Description: "Brave search MCP server",
				Enabled:     false,
			},
			"github": {
				Transport: "stdio",
				Command:   "npx",
				Args:      []string{"-y", "@modelcontextprotocol/server-github"},
				Env: map[string]string{
					"GITHUB_PERSONAL_ACCESS_TOKEN": "your-github-token-here",
				},
				Description: "GitHub repository access MCP server",
				Enabled:     false,
			},
			"google-maps": {
				Transport: "stdio",
				Command:   "npx",
				Args:      []string{"-y", "@modelcontextprotocol/server-google-maps"},
				Env: map[string]string{
					"GOOGLE_MAPS_API_KEY": "your-google-maps-api-key-here",
				},
				Description: "Google Maps MCP server",
				Enabled:     false,
			},
			"postgres": {
				Transport: "stdio",
				Command:   "npx",
				Args:      []string{"-y", "@modelcontextprotocol/server-postgres"},
				Env: map[string]string{
					"POSTGRES_CONNECTION_STRING": "postgresql://user:password@localhost:5432/database",
				},
				Description: "PostgreSQL database MCP server",
				Enabled:     false,
			},
			"puppeteer": {
				Transport:   "stdio",
				Command:     "npx",
				Args:        []string{"-y", "@modelcontextprotocol/server-puppeteer"},
				Description: "Web scraping with Puppeteer MCP server",
				Enabled:     false,
			},
			"sequential-thinking": {
				Transport:   "stdio",
				Command:     "npx",
				Args:        []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
				Description: "Sequential thinking MCP server",
				Enabled:     false,
			},
		},
	}
}

// ValidateConfig validates the MCP configuration
func ValidateConfig(config *MCPConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	for name, serverConfig := range config.Servers {
		if err := ValidateServerConfig(name, &serverConfig); err != nil {
			return fmt.Errorf("invalid config for server %s: %v", name, err)
		}
	}

	return nil
}

// ValidateServerConfig validates a single server configuration
func ValidateServerConfig(name string, config *MCPServerConfig) error {
	if name == "" {
		return fmt.Errorf("server name is required")
	}

	switch config.Transport {
	case "stdio":
		if config.Command == "" {
			return fmt.Errorf("command is required for stdio transport")
		}
	case "sse":
		if config.URL == "" {
			return fmt.Errorf("URL is required for SSE transport")
		}
	case "http":
		if config.URL == "" {
			return fmt.Errorf("URL is required for HTTP transport")
		}
	case "":
		return fmt.Errorf("transport is required")
	default:
		return fmt.Errorf("unsupported transport: %s", config.Transport)
	}

	return nil
}

// ToMCPServerConnection converts MCPServerConfig to MCPServerConnection
func (c *MCPServerConfig) ToMCPServerConnection(name string) *MCPServerConnection {
	return &MCPServerConnection{
		Name:      name,
		URL:       c.URL,
		Transport: c.Transport,
		Command:   c.Command,
		Args:      c.Args,
		Env:       c.Env,
		Connected: false,
		Tools:     []MCPTool{},
	}
}
