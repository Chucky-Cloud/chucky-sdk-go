package types

import "encoding/json"

// MessageType represents the type of SDK message.
type MessageType string

const (
	MessageTypeInit        MessageType = "init"
	MessageTypeUser        MessageType = "user"
	MessageTypeAssistant   MessageType = "assistant"
	MessageTypeSystem      MessageType = "system"
	MessageTypeResult      MessageType = "result"
	MessageTypeStreamEvent MessageType = "stream_event"
	MessageTypeControl     MessageType = "control"
	MessageTypeError       MessageType = "error"
	MessageTypePing        MessageType = "ping"
	MessageTypePong        MessageType = "pong"
	MessageTypeToolCall    MessageType = "tool_call"
	MessageTypeToolResult  MessageType = "tool_result"
)

// ResultSubtype represents the subtype of a result message.
type ResultSubtype string

const (
	ResultSubtypeSuccess             ResultSubtype = "success"
	ResultSubtypeErrorMaxTurns       ResultSubtype = "error_max_turns"
	ResultSubtypeErrorDuringExec     ResultSubtype = "error_during_execution"
	ResultSubtypeErrorBudget         ResultSubtype = "error_budget"
	ResultSubtypeErrorConcurrency    ResultSubtype = "error_concurrency"
	ResultSubtypeErrorAuthentication ResultSubtype = "error_authentication"
)

// SystemSubtype represents the subtype of a system message.
type SystemSubtype string

const (
	SystemSubtypeInit            SystemSubtype = "init"
	SystemSubtypeCompactBoundary SystemSubtype = "compact_boundary"
)

// ControlAction represents the action in a control message.
type ControlAction string

const (
	ControlActionReady       ControlAction = "ready"
	ControlActionSessionInfo ControlAction = "session_info"
	ControlActionEndInput    ControlAction = "end_input"
	ControlActionClose       ControlAction = "close"
)

// Role represents the role of a message sender.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// ContentBlockType represents the type of content block.
type ContentBlockType string

const (
	ContentBlockTypeText       ContentBlockType = "text"
	ContentBlockTypeImage      ContentBlockType = "image"
	ContentBlockTypeToolUse    ContentBlockType = "tool_use"
	ContentBlockTypeToolResult ContentBlockType = "tool_result"
)

// ContentBlock represents a content block in a message.
type ContentBlock struct {
	Type       ContentBlockType `json:"type"`
	Text       string           `json:"text,omitempty"`
	ID         string           `json:"id,omitempty"`
	Name       string           `json:"name,omitempty"`
	Input      any              `json:"input,omitempty"`
	ToolUseID  string           `json:"tool_use_id,omitempty"`
	Content    any              `json:"content,omitempty"`
	IsError    bool             `json:"is_error,omitempty"`
	Source     *ImageSource     `json:"source,omitempty"`
}

// ImageSource represents the source of an image.
type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// Message represents a message with role and content.
type Message struct {
	Role    Role   `json:"role"`
	Content any    `json:"content"` // string or []ContentBlock
}

// Usage represents token usage statistics.
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

// IncomingMessage is the interface for all incoming messages.
type IncomingMessage interface {
	GetType() MessageType
}

// OutgoingMessage is the interface for all outgoing messages.
type OutgoingMessage interface {
	GetType() MessageType
}

// InitPayload contains the initialization configuration.
type InitPayload struct {
	Model                 Model               `json:"model,omitempty"`
	FallbackModel         string              `json:"fallbackModel,omitempty"`
	SystemPrompt          any                 `json:"systemPrompt,omitempty"`
	MaxTurns              int                 `json:"maxTurns,omitempty"`
	MaxBudgetUsd          float64             `json:"maxBudgetUsd,omitempty"`
	MaxThinkingTokens     int                 `json:"maxThinkingTokens,omitempty"`
	Tools                 any                 `json:"tools,omitempty"`
	McpServers            any                 `json:"mcpServers,omitempty"`
	PermissionMode        PermissionMode      `json:"permissionMode,omitempty"`
	OutputFormat          *OutputFormat       `json:"outputFormat,omitempty"`
	IncludePartialMessages bool               `json:"includePartialMessages,omitempty"`
	Env                   map[string]string   `json:"env,omitempty"`
	SessionID             string              `json:"sessionId,omitempty"`
	ForkSession           bool                `json:"forkSession,omitempty"`
	ResumeSessionAt       string              `json:"resumeSessionAt,omitempty"`
	Continue              bool                `json:"continue,omitempty"`
}

// InitEnvelope is the init message sent to start a session.
type InitEnvelope struct {
	Type    MessageType `json:"type"`
	Payload InitPayload `json:"payload"`
}

func (InitEnvelope) GetType() MessageType { return MessageTypeInit }

// SDKUserMessage is a user message sent to Claude.
type SDKUserMessage struct {
	Type             MessageType `json:"type"`
	UUID             string      `json:"uuid,omitempty"`
	SessionID        string      `json:"session_id"`
	Message          Message     `json:"message"`
	ParentToolUseID  *string     `json:"parent_tool_use_id"`
}

func (SDKUserMessage) GetType() MessageType { return MessageTypeUser }

// ControlPayload contains control message data.
type ControlPayload struct {
	Action ControlAction `json:"action"`
	Data   any           `json:"data,omitempty"`
}

// ControlEnvelope is a control message for session management.
type ControlEnvelope struct {
	Type    MessageType    `json:"type"`
	Payload ControlPayload `json:"payload"`
}

func (ControlEnvelope) GetType() MessageType { return MessageTypeControl }

// PingPayload contains ping message data.
type PingPayload struct {
	Timestamp int64 `json:"timestamp"`
}

// PingEnvelope is a keep-alive ping message.
type PingEnvelope struct {
	Type    MessageType `json:"type"`
	Payload PingPayload `json:"payload"`
}

func (PingEnvelope) GetType() MessageType { return MessageTypePing }

// ToolResultPayload contains tool result data.
type ToolResultPayload struct {
	CallID string      `json:"callId"`
	Result *ToolResult `json:"result"`
}

// ToolResultEnvelope sends a tool execution result.
type ToolResultEnvelope struct {
	Type    MessageType       `json:"type"`
	Payload ToolResultPayload `json:"payload"`
}

func (ToolResultEnvelope) GetType() MessageType { return MessageTypeToolResult }

// SDKAssistantMessage is an assistant response from Claude.
type SDKAssistantMessage struct {
	Type            MessageType `json:"type"`
	UUID            string      `json:"uuid"`
	SessionID       string      `json:"session_id"`
	Message         Message     `json:"message"`
	ParentToolUseID *string     `json:"parent_tool_use_id"`
}

func (SDKAssistantMessage) GetType() MessageType { return MessageTypeAssistant }

// GetTextContent extracts text content from the message.
func (m SDKAssistantMessage) GetTextContent() string {
	switch content := m.Message.Content.(type) {
	case string:
		return content
	case []any:
		var text string
		for _, block := range content {
			if blockMap, ok := block.(map[string]any); ok {
				if blockMap["type"] == "text" {
					if t, ok := blockMap["text"].(string); ok {
						text += t
					}
				}
			}
		}
		return text
	}
	return ""
}

// SDKResultMessage is the final result of a session.
type SDKResultMessage struct {
	Type          MessageType   `json:"type"`
	Subtype       ResultSubtype `json:"subtype"`
	UUID          string        `json:"uuid"`
	SessionID     string        `json:"session_id"`
	DurationMs    int           `json:"duration_ms"`
	DurationApiMs int           `json:"duration_api_ms"`
	IsError       bool          `json:"is_error"`
	NumTurns      int           `json:"num_turns"`
	Result        string        `json:"result,omitempty"`
	TotalCostUsd  float64       `json:"total_cost_usd"`
	Usage         Usage         `json:"usage"`
	Errors        []string      `json:"errors,omitempty"`
}

func (SDKResultMessage) GetType() MessageType { return MessageTypeResult }

// SystemInitData contains data for system init messages.
type SystemInitData struct {
	CWD            string   `json:"cwd,omitempty"`
	Tools          []string `json:"tools,omitempty"`
	McpServers     []string `json:"mcp_servers,omitempty"`
	Model          string   `json:"model,omitempty"`
	PermissionMode string   `json:"permissionMode,omitempty"`
}

// SDKSystemMessage is a system message from the server.
type SDKSystemMessage struct {
	Type      MessageType   `json:"type"`
	Subtype   SystemSubtype `json:"subtype"`
	UUID      string        `json:"uuid"`
	SessionID string        `json:"session_id"`
	Data      any           `json:"data,omitempty"` // SystemInitData or compact metadata
}

func (SDKSystemMessage) GetType() MessageType { return MessageTypeSystem }

// SDKPartialAssistantMessage is a streaming event.
type SDKPartialAssistantMessage struct {
	Type            MessageType `json:"type"`
	Event           any         `json:"event"` // Raw streaming event
	UUID            string      `json:"uuid"`
	SessionID       string      `json:"session_id"`
	ParentToolUseID *string     `json:"parent_tool_use_id"`
}

func (SDKPartialAssistantMessage) GetType() MessageType { return MessageTypeStreamEvent }

// ErrorPayload contains error message data.
type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

// ErrorEnvelope is an error message from the server.
type ErrorEnvelope struct {
	Type    MessageType  `json:"type"`
	Payload ErrorPayload `json:"payload"`
}

func (ErrorEnvelope) GetType() MessageType { return MessageTypeError }

// PongPayload contains pong message data.
type PongPayload struct {
	Timestamp int64 `json:"timestamp"`
}

// PongEnvelope is a keep-alive pong response.
type PongEnvelope struct {
	Type    MessageType `json:"type"`
	Payload PongPayload `json:"payload"`
}

func (PongEnvelope) GetType() MessageType { return MessageTypePong }

// ToolCallPayload contains tool call data.
type ToolCallPayload struct {
	CallID   string `json:"callId"`
	ToolName string `json:"toolName"`
	Input    any    `json:"input"`
}

// ToolCallEnvelope requests execution of a tool.
type ToolCallEnvelope struct {
	Type    MessageType     `json:"type"`
	Payload ToolCallPayload `json:"payload"`
}

func (ToolCallEnvelope) GetType() MessageType { return MessageTypeToolCall }

// ParseIncomingMessage parses a JSON message into the appropriate type.
func ParseIncomingMessage(data []byte) (IncomingMessage, error) {
	var base struct {
		Type MessageType `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	var msg IncomingMessage
	switch base.Type {
	case MessageTypeAssistant:
		msg = &SDKAssistantMessage{}
	case MessageTypeResult:
		msg = &SDKResultMessage{}
	case MessageTypeSystem:
		msg = &SDKSystemMessage{}
	case MessageTypeStreamEvent:
		msg = &SDKPartialAssistantMessage{}
	case MessageTypeControl:
		msg = &ControlEnvelope{}
	case MessageTypeError:
		msg = &ErrorEnvelope{}
	case MessageTypePong:
		msg = &PongEnvelope{}
	case MessageTypeToolCall:
		msg = &ToolCallEnvelope{}
	default:
		// Return a generic structure for unknown types
		var generic map[string]any
		if err := json.Unmarshal(data, &generic); err != nil {
			return nil, err
		}
		return &GenericMessage{RawType: base.Type, Data: generic}, nil
	}

	if err := json.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// GenericMessage holds unknown message types.
type GenericMessage struct {
	RawType MessageType
	Data    map[string]any
}

func (g *GenericMessage) GetType() MessageType { return g.RawType }
