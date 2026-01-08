package types

import "context"

// ExecuteLocation specifies where a tool executes.
type ExecuteLocation string

const (
	ExecuteInServer  ExecuteLocation = "server"
	ExecuteInBrowser ExecuteLocation = "browser"
)

// ToolInputSchema represents a JSON Schema for tool input validation.
type ToolInputSchema struct {
	Type                 string                     `json:"type"`
	Properties           map[string]JsonSchemaProperty `json:"properties,omitempty"`
	Required             []string                   `json:"required,omitempty"`
	AdditionalProperties *bool                      `json:"additionalProperties,omitempty"`
}

// JsonSchemaProperty represents a property in a JSON Schema.
type JsonSchemaProperty struct {
	Type        string              `json:"type,omitempty"`
	Description string              `json:"description,omitempty"`
	Enum        []any               `json:"enum,omitempty"`
	Default     any                 `json:"default,omitempty"`
	MinLength   *int                `json:"minLength,omitempty"`
	MaxLength   *int                `json:"maxLength,omitempty"`
	Pattern     string              `json:"pattern,omitempty"`
	Minimum     *float64            `json:"minimum,omitempty"`
	Maximum     *float64            `json:"maximum,omitempty"`
	Items       *JsonSchemaProperty `json:"items,omitempty"`
	Properties  map[string]JsonSchemaProperty `json:"properties,omitempty"`
	Required    []string            `json:"required,omitempty"`
}

// ToolContent represents content returned by a tool.
type ToolContent interface {
	toolContent()
}

// TextToolContent represents text content from a tool.
type TextToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (TextToolContent) toolContent() {}

// ImageToolContent represents image content from a tool.
type ImageToolContent struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}

func (ImageToolContent) toolContent() {}

// ResourceToolContent represents resource content from a tool.
type ResourceToolContent struct {
	Type     string `json:"type"`
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

func (ResourceToolContent) toolContent() {}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Content []any `json:"content"` // []ToolContent as any for JSON marshaling
	IsError bool  `json:"isError,omitempty"`
}

// ToolHandler is the function signature for tool handlers.
type ToolHandler func(ctx context.Context, input map[string]any) (*ToolResult, error)

// ToolDefinition defines a tool that can be used by Claude.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema ToolInputSchema `json:"inputSchema"`
	ExecuteIn   ExecuteLocation `json:"executeIn,omitempty"`
	Handler     ToolHandler     `json:"-"` // Not serialized
}

// McpServerType represents the type of MCP server.
type McpServerType string

const (
	McpServerTypeStdio McpServerType = "stdio"
	McpServerTypeSSE   McpServerType = "sse"
	McpServerTypeHTTP  McpServerType = "http"
)

// McpServerDefinition is an interface for MCP server configurations.
type McpServerDefinition interface {
	mcpServer()
	GetName() string
}

// McpClientToolsServer represents an MCP server with client-side tools.
type McpClientToolsServer struct {
	Name    string           `json:"name"`
	Version string           `json:"version,omitempty"`
	Tools   []ToolDefinition `json:"tools"`
}

func (McpClientToolsServer) mcpServer() {}
func (s McpClientToolsServer) GetName() string { return s.Name }

// McpStdioServerConfig represents an MCP server running via stdio.
type McpStdioServerConfig struct {
	Name    string            `json:"name"`
	Type    McpServerType     `json:"type,omitempty"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func (McpStdioServerConfig) mcpServer() {}
func (s McpStdioServerConfig) GetName() string { return s.Name }

// McpSSEServerConfig represents an MCP server using SSE transport.
type McpSSEServerConfig struct {
	Name    string            `json:"name"`
	Type    McpServerType     `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (McpSSEServerConfig) mcpServer() {}
func (s McpSSEServerConfig) GetName() string { return s.Name }

// McpHTTPServerConfig represents an MCP server using HTTP transport.
type McpHTTPServerConfig struct {
	Name    string            `json:"name"`
	Type    McpServerType     `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (McpHTTPServerConfig) mcpServer() {}
func (s McpHTTPServerConfig) GetName() string { return s.Name }
