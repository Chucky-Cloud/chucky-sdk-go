package tools

import (
	"github.com/chucky-cloud/chucky-sdk-go/pkg/types"
)

// McpServerBuilder helps build MCP server configurations.
type McpServerBuilder struct {
	name    string
	version string
	tools   []types.ToolDefinition
}

// NewMcpServer creates a new MCP server builder.
func NewMcpServer(name string) *McpServerBuilder {
	return &McpServerBuilder{
		name:    name,
		version: "1.0.0",
		tools:   make([]types.ToolDefinition, 0),
	}
}

// Version sets the server version.
func (b *McpServerBuilder) Version(version string) *McpServerBuilder {
	b.version = version
	return b
}

// AddTool adds a tool using options.
func (b *McpServerBuilder) AddTool(opts CreateToolOptions) *McpServerBuilder {
	b.tools = append(b.tools, CreateTool(opts))
	return b
}

// Add adds an existing tool definition.
func (b *McpServerBuilder) Add(tool types.ToolDefinition) *McpServerBuilder {
	b.tools = append(b.tools, tool)
	return b
}

// AddTools adds multiple tool definitions.
func (b *McpServerBuilder) AddTools(tools ...types.ToolDefinition) *McpServerBuilder {
	b.tools = append(b.tools, tools...)
	return b
}

// Build returns the completed MCP server definition.
func (b *McpServerBuilder) Build() types.McpClientToolsServer {
	return types.McpClientToolsServer{
		Name:    b.name,
		Version: b.version,
		Tools:   b.tools,
	}
}

// CreateSdkMcpServer creates an MCP server with tools.
// This is the main helper function for creating tool servers.
func CreateSdkMcpServer(name string, tools ...types.ToolDefinition) types.McpClientToolsServer {
	return types.McpClientToolsServer{
		Name:    name,
		Version: "1.0.0",
		Tools:   tools,
	}
}

// StdioServer creates an MCP stdio server configuration.
func StdioServer(name, command string, args ...string) types.McpStdioServerConfig {
	return types.McpStdioServerConfig{
		Name:    name,
		Type:    types.McpServerTypeStdio,
		Command: command,
		Args:    args,
	}
}

// StdioServerWithEnv creates an MCP stdio server with environment variables.
func StdioServerWithEnv(name, command string, args []string, env map[string]string) types.McpStdioServerConfig {
	return types.McpStdioServerConfig{
		Name:    name,
		Type:    types.McpServerTypeStdio,
		Command: command,
		Args:    args,
		Env:     env,
	}
}

// SSEServer creates an MCP SSE server configuration.
func SSEServer(name, url string, headers ...map[string]string) types.McpSSEServerConfig {
	var h map[string]string
	if len(headers) > 0 {
		h = headers[0]
	}
	return types.McpSSEServerConfig{
		Name:    name,
		Type:    types.McpServerTypeSSE,
		URL:     url,
		Headers: h,
	}
}

// HTTPServer creates an MCP HTTP server configuration.
func HTTPServer(name, url string, headers ...map[string]string) types.McpHTTPServerConfig {
	var h map[string]string
	if len(headers) > 0 {
		h = headers[0]
	}
	return types.McpHTTPServerConfig{
		Name:    name,
		Type:    types.McpServerTypeHTTP,
		URL:     url,
		Headers: h,
	}
}
