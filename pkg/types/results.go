package types

// SessionResult represents the result of a completed session.
type SessionResult struct {
	Type       string  `json:"type"`
	Subtype    string  `json:"subtype"`
	SessionID  string  `json:"session_id"`
	Result     string  `json:"result"`
	IsError    bool    `json:"is_error"`
	DurationMs int     `json:"duration_ms"`
	NumTurns   int     `json:"num_turns"`
	TotalCostUsd float64 `json:"total_cost_usd"`
	Usage      Usage   `json:"usage"`
	Errors     []string `json:"errors,omitempty"`
}

// PromptResult is an alias for SessionResult for one-shot prompts.
type PromptResult = SessionResult

// FromResultMessage converts an SDKResultMessage to SessionResult.
func FromResultMessage(msg *SDKResultMessage) *SessionResult {
	return &SessionResult{
		Type:       string(msg.Type),
		Subtype:    string(msg.Subtype),
		SessionID:  msg.SessionID,
		Result:     msg.Result,
		IsError:    msg.IsError,
		DurationMs: msg.DurationMs,
		NumTurns:   msg.NumTurns,
		TotalCostUsd: msg.TotalCostUsd,
		Usage:      msg.Usage,
		Errors:     msg.Errors,
	}
}

// GetResultText extracts the result text from various message types.
func GetResultText(msg any) string {
	switch m := msg.(type) {
	case *SDKResultMessage:
		return m.Result
	case SDKResultMessage:
		return m.Result
	case *SessionResult:
		return m.Result
	case SessionResult:
		return m.Result
	}
	return ""
}

// GetAssistantText extracts text content from an assistant message.
func GetAssistantText(msg any) string {
	switch m := msg.(type) {
	case *SDKAssistantMessage:
		return m.GetTextContent()
	case SDKAssistantMessage:
		return m.GetTextContent()
	}
	return ""
}
