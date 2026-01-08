package chucky

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"

	"github.com/chucky-cloud/chucky-sdk-go/pkg/transport"
	"github.com/chucky-cloud/chucky-sdk-go/pkg/types"
)

// SessionState represents the current state of a session.
type SessionState string

const (
	SessionStateIdle        SessionState = "idle"
	SessionStateInitializing SessionState = "initializing"
	SessionStateReady       SessionState = "ready"
	SessionStateProcessing  SessionState = "processing"
	SessionStateWaitingTool SessionState = "waiting_tool"
	SessionStateCompleted   SessionState = "completed"
	SessionStateError       SessionState = "error"
)

// SessionEventHandlers contains callbacks for session events.
type SessionEventHandlers struct {
	OnMessage func(msg types.IncomingMessage)
	OnError   func(err error)
	OnClose   func()
}

// Session manages a multi-turn conversation with Claude.
type Session struct {
	client    *Client
	transport transport.Transport
	options   types.SessionOptions
	sessionID string
	state     SessionState
	stateMu   sync.RWMutex
	handlers  SessionEventHandlers

	connected    bool
	connectedMu  sync.RWMutex

	msgCh        chan types.IncomingMessage
	errCh        chan error
	closeCh      chan struct{}
	closeOnce    sync.Once

	// For waiting on server ready
	readyCh      chan struct{}
	readyOnce    sync.Once
	initErr      error

	toolHandlers map[string]types.ToolHandler
	toolsMu      sync.RWMutex
}

func newSession(client *Client, t transport.Transport, opts types.SessionOptions) *Session {
	// Don't generate sessionID - server will assign it
	s := &Session{
		client:       client,
		transport:    t,
		options:      opts,
		sessionID:    "", // Will be assigned by server in system:init
		state:        SessionStateIdle,
		msgCh:        make(chan types.IncomingMessage, 100),
		errCh:        make(chan error, 10),
		closeCh:      make(chan struct{}),
		readyCh:      make(chan struct{}),
		toolHandlers: make(map[string]types.ToolHandler),
	}

	// Extract tool handlers from MCP servers
	s.extractToolHandlers()

	// Set up transport handlers
	t.SetEventHandlers(transport.TransportEvents{
		OnMessage:      s.handleMessage,
		OnClose:        s.handleClose,
		OnStatusChange: s.handleStatusChange,
		OnError:        s.handleError,
	})

	return s
}

// ID returns the session ID.
func (s *Session) ID() string {
	return s.sessionID
}

// State returns the current session state.
func (s *Session) State() SessionState {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.state
}

func (s *Session) setState(state SessionState) {
	s.stateMu.Lock()
	s.state = state
	s.stateMu.Unlock()
}

// Connect establishes the connection and initializes the session.
func (s *Session) Connect(ctx context.Context) error {
	s.connectedMu.Lock()
	if s.connected {
		s.connectedMu.Unlock()
		return nil
	}
	s.connectedMu.Unlock()

	s.setState(SessionStateInitializing)

	if err := s.transport.Connect(); err != nil {
		s.setState(SessionStateError)
		return err
	}

	if err := s.transport.WaitForReady(); err != nil {
		s.setState(SessionStateError)
		return err
	}

	// Send init message
	if err := s.sendInit(); err != nil {
		s.setState(SessionStateError)
		return err
	}

	// Wait for server to be ready (control:ready or system:init)
	select {
	case <-s.readyCh:
		if s.initErr != nil {
			s.setState(SessionStateError)
			return s.initErr
		}
	case <-ctx.Done():
		s.setState(SessionStateError)
		return ctx.Err()
	case <-s.closeCh:
		s.setState(SessionStateError)
		return types.SessionError("session closed during initialization")
	}

	s.connectedMu.Lock()
	s.connected = true
	s.connectedMu.Unlock()

	s.setState(SessionStateReady)
	s.client.notifySessionStart(s.sessionID)

	return nil
}

func (s *Session) sendInit() error {
	// Convert MCP servers to serializable format
	var mcpServers any
	if len(s.options.McpServers) > 0 {
		servers := make([]map[string]any, 0, len(s.options.McpServers))
		for _, server := range s.options.McpServers {
			serverMap := s.mcpServerToMap(server)
			if serverMap != nil {
				servers = append(servers, serverMap)
			}
		}
		mcpServers = servers
	}

	init := types.InitEnvelope{
		Type: types.MessageTypeInit,
		Payload: types.InitPayload{
			Model:                  s.options.Model,
			FallbackModel:          s.options.FallbackModel,
			SystemPrompt:           s.options.SystemPrompt,
			MaxTurns:               s.options.MaxTurns,
			MaxBudgetUsd:           s.options.MaxBudgetUsd,
			MaxThinkingTokens:      s.options.MaxThinkingTokens,
			Tools:                  s.options.Tools,
			McpServers:             mcpServers,
			PermissionMode:         s.options.PermissionMode,
			OutputFormat:           s.options.OutputFormat,
			IncludePartialMessages: s.options.IncludePartialMessages,
			Env:                    s.options.Env,
			// Note: SessionID is NOT sent - server assigns it in system:init
			ForkSession:            s.options.ForkSession,
			ResumeSessionAt:        s.options.ResumeSessionAt,
			Continue:               s.options.Continue,
		},
	}

	return s.transport.Send(init)
}

func (s *Session) mcpServerToMap(server types.McpServerDefinition) map[string]any {
	switch srv := server.(type) {
	case types.McpClientToolsServer:
		// Convert tools to serializable format
		tools := make([]map[string]any, 0, len(srv.Tools))
		for _, tool := range srv.Tools {
			toolMap := map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": tool.InputSchema,
			}
			// If tool has a handler, mark it for client-side execution
			// This tells the server to send tool_call messages back to us
			if tool.Handler != nil {
				toolMap["executeIn"] = "client"
			}
			tools = append(tools, toolMap)
		}
		return map[string]any{
			"name":    srv.Name,
			"version": srv.Version,
			"tools":   tools,
		}
	case types.McpStdioServerConfig:
		return map[string]any{
			"name":    srv.Name,
			"type":    "stdio",
			"command": srv.Command,
			"args":    srv.Args,
			"env":     srv.Env,
		}
	case types.McpSSEServerConfig:
		return map[string]any{
			"name":    srv.Name,
			"type":    "sse",
			"url":     srv.URL,
			"headers": srv.Headers,
		}
	case types.McpHTTPServerConfig:
		return map[string]any{
			"name":    srv.Name,
			"type":    "http",
			"url":     srv.URL,
			"headers": srv.Headers,
		}
	}
	return nil
}

// Send sends a user message to Claude.
func (s *Session) Send(ctx context.Context, message string) error {
	// Auto-connect if needed
	s.connectedMu.RLock()
	connected := s.connected
	s.connectedMu.RUnlock()

	if !connected {
		if err := s.Connect(ctx); err != nil {
			return err
		}
	}

	s.setState(SessionStateProcessing)

	// Use server-assigned session ID, or "unknown" if not yet received
	sessionID := s.sessionID
	if sessionID == "" {
		sessionID = "unknown"
	}

	msg := types.SDKUserMessage{
		Type:      types.MessageTypeUser,
		UUID:      uuid.New().String(),
		SessionID: sessionID,
		Message: types.Message{
			Role:    types.RoleUser,
			Content: message,
		},
		ParentToolUseID: nil,
	}

	return s.transport.Send(msg)
}

// Stream returns a channel that yields incoming messages.
func (s *Session) Stream(ctx context.Context) <-chan types.IncomingMessage {
	out := make(chan types.IncomingMessage)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.closeCh:
				return
			case msg, ok := <-s.msgCh:
				if !ok {
					return
				}
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				case <-s.closeCh:
					return
				}

				// Check if session is complete
				if _, ok := msg.(*types.SDKResultMessage); ok {
					s.setState(SessionStateCompleted)
					return
				}
			}
		}
	}()

	return out
}

// Receive is an alias for Stream.
func (s *Session) Receive(ctx context.Context) <-chan types.IncomingMessage {
	return s.Stream(ctx)
}

// Close closes the session.
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		close(s.closeCh)

		// Send close control message
		closeMsg := types.ControlEnvelope{
			Type: types.MessageTypeControl,
			Payload: types.ControlPayload{
				Action: types.ControlActionClose,
			},
		}
		_ = s.transport.Send(closeMsg)

		_ = s.transport.Disconnect()

		s.client.removeSession(s.sessionID)

		if s.handlers.OnClose != nil {
			s.handlers.OnClose()
		}
	})
}

// On sets the session event handlers.
func (s *Session) On(handlers SessionEventHandlers) *Session {
	s.handlers = handlers
	return s
}

func (s *Session) handleMessage(msg types.IncomingMessage) {
	// Check if this is a ready signal (before session is fully connected)
	s.connectedMu.RLock()
	connected := s.connected
	s.connectedMu.RUnlock()

	if !connected {
		// Check for ready signals during initialization
		switch m := msg.(type) {
		case *types.ControlEnvelope:
			// Signal ready on control:ready - we can send user message before system:init
			if m.Payload.Action == types.ControlActionReady || m.Payload.Action == types.ControlActionSessionInfo {
				s.readyOnce.Do(func() {
					close(s.readyCh)
				})
				return
			}
		case *types.SDKSystemMessage:
			if m.Subtype == types.SystemSubtypeInit {
				// Update session ID from server when it arrives (may come after first user message)
				if m.SessionID != "" {
					s.sessionID = m.SessionID
				}
				// Also signal ready in case control:ready didn't come first
				s.readyOnce.Do(func() {
					close(s.readyCh)
				})
				// Don't return - also forward to message channel
			}
		case *types.ErrorEnvelope:
			s.initErr = types.SessionError(m.Payload.Message)
			s.readyOnce.Do(func() {
				close(s.readyCh)
			})
			// Forward error to channel too
		}
	}

	// Handle tool calls internally
	if toolCall, ok := msg.(*types.ToolCallEnvelope); ok {
		s.handleToolCall(toolCall)
		return
	}

	// Forward message to channel
	select {
	case s.msgCh <- msg:
	case <-s.closeCh:
	}

	if s.handlers.OnMessage != nil {
		s.handlers.OnMessage(msg)
	}
}

func (s *Session) handleToolCall(call *types.ToolCallEnvelope) {
	s.setState(SessionStateWaitingTool)

	s.toolsMu.RLock()
	handler, ok := s.toolHandlers[call.Payload.ToolName]
	s.toolsMu.RUnlock()

	var result *types.ToolResult
	if !ok {
		result = &types.ToolResult{
			Content: []any{
				types.TextToolContent{
					Type: "text",
					Text: "Tool not found: " + call.Payload.ToolName,
				},
			},
			IsError: true,
		}
	} else {
		// Convert input to map
		var input map[string]any
		switch v := call.Payload.Input.(type) {
		case map[string]any:
			input = v
		default:
			// Try to marshal and unmarshal to get a map
			data, _ := json.Marshal(call.Payload.Input)
			_ = json.Unmarshal(data, &input)
		}

		var err error
		result, err = handler(context.Background(), input)
		if err != nil {
			result = &types.ToolResult{
				Content: []any{
					types.TextToolContent{
						Type: "text",
						Text: "Tool execution error: " + err.Error(),
					},
				},
				IsError: true,
			}
		}
	}

	// Send tool result
	resultMsg := types.ToolResultEnvelope{
		Type: types.MessageTypeToolResult,
		Payload: types.ToolResultPayload{
			CallID: call.Payload.CallID,
			Result: result,
		},
	}

	if err := s.transport.Send(resultMsg); err != nil {
		s.handleError(err)
	}

	s.setState(SessionStateProcessing)
}

func (s *Session) handleClose(code int, reason string) {
	s.Close()
}

func (s *Session) handleStatusChange(status transport.ConnectionStatus) {
	// Could map transport status to session state
}

func (s *Session) handleError(err error) {
	select {
	case s.errCh <- err:
	default:
	}

	s.client.notifyError(err)

	if s.handlers.OnError != nil {
		s.handlers.OnError(err)
	}
}

func (s *Session) extractToolHandlers() {
	for _, server := range s.options.McpServers {
		if clientTools, ok := server.(types.McpClientToolsServer); ok {
			for _, tool := range clientTools.Tools {
				if tool.Handler != nil {
					s.toolsMu.Lock()
					s.toolHandlers[tool.Name] = tool.Handler
					s.toolsMu.Unlock()
				}
			}
		}
	}
}
