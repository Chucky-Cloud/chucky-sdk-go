// Package chuckysdk provides the Go SDK for interacting with Chucky (Claude Code sandbox).
//
// This SDK allows you to create sessions with Claude, send messages, and receive
// responses including tool calls.
//
// Basic usage:
//
//	client := chuckysdk.NewClient(types.ClientOptions{
//	    Token: "your-token",
//	})
//
//	session := client.CreateSession(&types.SessionOptions{
//	    Model: types.ModelClaudeSonnet,
//	})
//
//	err := session.Send(ctx, "Hello, Claude!")
//	for msg := range session.Stream(ctx) {
//	    // Handle messages
//	}
package chuckysdk

import (
	"github.com/chucky-cloud/chucky-sdk-go/pkg/chucky"
	"github.com/chucky-cloud/chucky-sdk-go/pkg/tools"
	"github.com/chucky-cloud/chucky-sdk-go/pkg/types"
	"github.com/chucky-cloud/chucky-sdk-go/pkg/utils"
)

// Re-export client types
type (
	Client              = chucky.Client
	Session             = chucky.Session
	ClientEventHandlers = chucky.ClientEventHandlers
	SessionEventHandlers = chucky.SessionEventHandlers
	SessionState        = chucky.SessionState
)

// Re-export session states
const (
	SessionStateIdle        = chucky.SessionStateIdle
	SessionStateInitializing = chucky.SessionStateInitializing
	SessionStateReady       = chucky.SessionStateReady
	SessionStateProcessing  = chucky.SessionStateProcessing
	SessionStateWaitingTool = chucky.SessionStateWaitingTool
	SessionStateCompleted   = chucky.SessionStateCompleted
	SessionStateError       = chucky.SessionStateError
)

// NewClient creates a new Chucky client.
var NewClient = chucky.NewClient

// Re-export type definitions
type (
	// Options
	ClientOptions  = types.ClientOptions
	SessionOptions = types.SessionOptions
	BaseOptions    = types.BaseOptions
	Model          = types.Model
	PermissionMode = types.PermissionMode
	OutputFormat   = types.OutputFormat

	// Messages
	IncomingMessage            = types.IncomingMessage
	OutgoingMessage            = types.OutgoingMessage
	SDKAssistantMessage        = types.SDKAssistantMessage
	SDKResultMessage           = types.SDKResultMessage
	SDKSystemMessage           = types.SDKSystemMessage
	SDKPartialAssistantMessage = types.SDKPartialAssistantMessage
	SDKUserMessage             = types.SDKUserMessage
	ControlEnvelope            = types.ControlEnvelope
	ErrorEnvelope              = types.ErrorEnvelope
	ToolCallEnvelope           = types.ToolCallEnvelope
	Message                    = types.Message
	ContentBlock               = types.ContentBlock
	Usage                      = types.Usage

	// Tools
	ToolDefinition       = types.ToolDefinition
	ToolResult           = types.ToolResult
	ToolHandler          = types.ToolHandler
	ToolInputSchema      = types.ToolInputSchema
	JsonSchemaProperty   = types.JsonSchemaProperty
	TextToolContent      = types.TextToolContent
	ImageToolContent     = types.ImageToolContent
	ResourceToolContent  = types.ResourceToolContent
	ExecuteLocation      = types.ExecuteLocation
	McpServerDefinition  = types.McpServerDefinition
	McpClientToolsServer = types.McpClientToolsServer
	McpStdioServerConfig = types.McpStdioServerConfig
	McpSSEServerConfig   = types.McpSSEServerConfig
	McpHTTPServerConfig  = types.McpHTTPServerConfig

	// Results
	SessionResult = types.SessionResult
	PromptResult  = types.PromptResult

	// Token
	TokenBudget        = types.TokenBudget
	TokenPermissions   = types.TokenPermissions
	TokenSdkConfig     = types.TokenSdkConfig
	BudgetTokenPayload = types.BudgetTokenPayload
	CreateTokenOptions = types.CreateTokenOptions
	CreateBudgetOptions = types.CreateBudgetOptions
	DecodedToken       = types.DecodedToken
	BudgetWindow       = types.BudgetWindow

	// Errors
	ChuckyError = types.ChuckyError
	ErrorCode   = types.ErrorCode
)

// Re-export constants
const (
	// Models
	ModelClaudeSonnet = types.ModelClaudeSonnet
	ModelClaudeOpus   = types.ModelClaudeOpus

	// Permission modes
	PermissionModeDefault           = types.PermissionModeDefault
	PermissionModePlan              = types.PermissionModePlan
	PermissionModeBypassPermissions = types.PermissionModeBypassPermissions

	// Execute locations
	ExecuteInServer  = types.ExecuteInServer
	ExecuteInBrowser = types.ExecuteInBrowser

	// Budget windows
	BudgetWindowHour  = types.BudgetWindowHour
	BudgetWindowDay   = types.BudgetWindowDay
	BudgetWindowWeek  = types.BudgetWindowWeek
	BudgetWindowMonth = types.BudgetWindowMonth
)

// Tool helpers
var (
	// CreateTool creates a new tool definition.
	CreateTool = tools.CreateTool

	// Tool is a shorthand for creating a tool.
	Tool = tools.Tool

	// BrowserTool creates a browser-side tool.
	BrowserTool = tools.BrowserTool

	// ServerTool creates a server-side tool.
	ServerTool = tools.ServerTool

	// TextResult creates a text tool result.
	TextResult = tools.TextResult

	// ErrorResult creates an error tool result.
	ErrorResult = tools.ErrorResult

	// ImageResult creates an image tool result.
	ImageResult = tools.ImageResult

	// ResourceResult creates a resource tool result.
	ResourceResult = tools.ResourceResult

	// NewSchema creates a new schema builder.
	NewSchema = tools.NewSchema

	// SimpleHandler wraps a simple function as a tool handler.
	SimpleHandler = tools.SimpleHandler
)

// MCP server helpers
var (
	// NewMcpServer creates a new MCP server builder.
	NewMcpServer = tools.NewMcpServer

	// CreateSdkMcpServer creates an MCP server with tools.
	CreateSdkMcpServer = tools.CreateSdkMcpServer

	// StdioServer creates an MCP stdio server.
	StdioServer = tools.StdioServer

	// StdioServerWithEnv creates an MCP stdio server with env vars.
	StdioServerWithEnv = tools.StdioServerWithEnv

	// SSEServer creates an MCP SSE server.
	SSEServer = tools.SSEServer

	// HTTPServer creates an MCP HTTP server.
	HTTPServer = tools.HTTPServer
)

// Token utilities
var (
	// CreateToken creates a new JWT token.
	CreateToken = utils.CreateToken

	// CreateBudget creates a budget from human-readable values.
	CreateBudget = utils.CreateBudget

	// DecodeToken decodes a JWT token without verification.
	DecodeToken = utils.DecodeToken

	// VerifyToken verifies a JWT token signature.
	VerifyToken = utils.VerifyToken

	// IsTokenExpired checks if a token has expired.
	IsTokenExpired = utils.IsTokenExpired

	// ExtractProjectID extracts the project ID from an HMAC key.
	ExtractProjectID = utils.ExtractProjectID

	// GetTokenExpiration returns the expiration time of a token.
	GetTokenExpiration = utils.GetTokenExpiration

	// GetTokenBudget returns the budget from a token.
	GetTokenBudget = utils.GetTokenBudget
)

// Budget helpers
var (
	// MicroDollars converts dollars to microdollars.
	MicroDollars = types.MicroDollars

	// ComputeSeconds converts hours to seconds.
	ComputeSeconds = types.ComputeSeconds
)

// Result helpers
var (
	// GetResultText extracts result text from a message.
	GetResultText = types.GetResultText

	// GetAssistantText extracts assistant text from a message.
	GetAssistantText = types.GetAssistantText

	// FromResultMessage converts an SDKResultMessage to SessionResult.
	FromResultMessage = types.FromResultMessage
)

// Error constructors
var (
	ConnectionError      = types.ConnectionError
	AuthenticationError  = types.AuthenticationError
	BudgetExceededError  = types.BudgetExceededError
	ConcurrencyLimitError = types.ConcurrencyLimitError
	RateLimitError       = types.RateLimitError
	SessionError         = types.SessionError
	ToolExecutionError   = types.ToolExecutionError
	TimeoutError         = types.TimeoutError
	ValidationError      = types.ValidationError
	ProtocolError        = types.ProtocolError
)

// CreateToolOptions is the options for creating a tool.
type CreateToolOptions = tools.CreateToolOptions

// McpServerBuilder helps build MCP server configurations.
type McpServerBuilder = tools.McpServerBuilder

// SchemaBuilder helps build JSON schemas.
type SchemaBuilder = tools.SchemaBuilder
