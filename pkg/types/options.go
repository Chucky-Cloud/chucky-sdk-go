package types

import "time"

// Model represents the Claude model to use.
type Model string

const (
	ModelClaudeSonnet Model = "claude-sonnet-4-5-20250929"
	ModelClaudeOpus   Model = "claude-opus-4-5-20251101"
)

// PermissionMode represents the permission mode for tool execution.
type PermissionMode string

const (
	PermissionModeDefault          PermissionMode = "default"
	PermissionModePlan             PermissionMode = "plan"
	PermissionModeBypassPermissions PermissionMode = "bypassPermissions"
)

// SystemPromptPreset represents a preset system prompt.
type SystemPromptPreset struct {
	Type   string `json:"type"`
	Preset string `json:"preset"`
	Append string `json:"append,omitempty"`
}

// OutputFormat represents the output format configuration.
type OutputFormat struct {
	Type   string `json:"type"`
	Schema any    `json:"schema"`
}

// BaseOptions contains common options for sessions.
type BaseOptions struct {
	// Model selection
	Model         Model  `json:"model,omitempty"`
	FallbackModel string `json:"fallbackModel,omitempty"`

	// Prompting
	SystemPrompt      any `json:"systemPrompt,omitempty"` // string or SystemPromptPreset
	MaxTurns          int `json:"maxTurns,omitempty"`
	MaxBudgetUsd      float64 `json:"maxBudgetUsd,omitempty"`
	MaxThinkingTokens int `json:"maxThinkingTokens,omitempty"`

	// Tools
	Tools      any                   `json:"tools,omitempty"` // []string or ToolsPreset
	McpServers []McpServerDefinition `json:"mcpServers,omitempty"`

	// Other
	PermissionMode        PermissionMode `json:"permissionMode,omitempty"`
	OutputFormat          *OutputFormat  `json:"outputFormat,omitempty"`
	IncludePartialMessages bool          `json:"includePartialMessages,omitempty"`
	Env                   map[string]string `json:"env,omitempty"`
}

// SessionOptions extends BaseOptions with session-specific options.
type SessionOptions struct {
	BaseOptions

	// Session-specific
	SessionID       string `json:"sessionId,omitempty"`
	ForkSession     bool   `json:"forkSession,omitempty"`
	ResumeSessionAt string `json:"resumeSessionAt,omitempty"`
	Continue        bool   `json:"continue,omitempty"`

	// Setting sources
	SettingSources []string `json:"settingSources,omitempty"`
}

// ClientOptions contains options for creating a Chucky client.
type ClientOptions struct {
	// Connection
	BaseURL string `json:"baseUrl,omitempty"`
	Token   string `json:"token"`

	// Behavior
	Debug                 bool          `json:"debug,omitempty"`
	Timeout               time.Duration `json:"timeout,omitempty"`
	KeepAliveInterval     time.Duration `json:"keepAliveInterval,omitempty"`
	AutoReconnect         bool          `json:"autoReconnect,omitempty"`
	MaxReconnectAttempts  int           `json:"maxReconnectAttempts,omitempty"`
}

// DefaultClientOptions returns the default client options.
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		BaseURL:           "wss://conjure.chucky.cloud/ws",
		Timeout:           60 * time.Second,
		KeepAliveInterval: 5 * time.Minute,
		AutoReconnect:     false,
		MaxReconnectAttempts: 0,
	}
}

// Merge merges the provided options with the default options.
func (o ClientOptions) Merge(other ClientOptions) ClientOptions {
	if other.BaseURL != "" {
		o.BaseURL = other.BaseURL
	}
	if other.Token != "" {
		o.Token = other.Token
	}
	if other.Debug {
		o.Debug = true
	}
	if other.Timeout > 0 {
		o.Timeout = other.Timeout
	}
	if other.KeepAliveInterval > 0 {
		o.KeepAliveInterval = other.KeepAliveInterval
	}
	if other.AutoReconnect {
		o.AutoReconnect = true
	}
	if other.MaxReconnectAttempts > 0 {
		o.MaxReconnectAttempts = other.MaxReconnectAttempts
	}
	return o
}
