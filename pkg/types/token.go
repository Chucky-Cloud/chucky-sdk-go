package types

import "time"

// BudgetWindow represents the time window for budget tracking.
type BudgetWindow string

const (
	BudgetWindowHour  BudgetWindow = "hour"
	BudgetWindowDay   BudgetWindow = "day"
	BudgetWindowWeek  BudgetWindow = "week"
	BudgetWindowMonth BudgetWindow = "month"
)

// TokenBudget represents the budget configuration for a token.
type TokenBudget struct {
	AI          int64        `json:"ai"`          // Microdollars (1 USD = 1,000,000)
	Compute     int64        `json:"compute"`     // Seconds
	Window      BudgetWindow `json:"window"`
	WindowStart string       `json:"windowStart"` // ISO 8601
}

// TokenPermissions represents optional permission restrictions.
type TokenPermissions struct {
	AllowedModels   []string `json:"allowedModels,omitempty"`
	AllowedTools    []string `json:"allowedTools,omitempty"`
	MaxTurnsPerSession int  `json:"maxTurnsPerSession,omitempty"`
}

// TokenSdkConfig represents optional SDK configuration overrides.
type TokenSdkConfig struct {
	DefaultModel  string `json:"defaultModel,omitempty"`
	SystemPrompt  string `json:"systemPrompt,omitempty"`
}

// BudgetTokenPayload represents the JWT payload for a budget token.
type BudgetTokenPayload struct {
	// Standard JWT claims
	Subject   string `json:"sub"`           // User ID
	Issuer    string `json:"iss"`           // Project ID
	IssuedAt  int64  `json:"iat"`           // Unix timestamp
	ExpiresAt int64  `json:"exp"`           // Unix timestamp

	// Custom claims
	Budget      TokenBudget       `json:"budget"`
	Permissions *TokenPermissions `json:"permissions,omitempty"`
	SdkConfig   *TokenSdkConfig   `json:"sdkConfig,omitempty"`
}

// CreateTokenOptions contains options for creating a token.
type CreateTokenOptions struct {
	UserID      string
	ProjectID   string
	Secret      string
	Budget      TokenBudget
	ExpiresIn   time.Duration // Default: 1 hour
	Permissions *TokenPermissions
	SdkConfig   *TokenSdkConfig
}

// CreateBudgetOptions contains options for creating a budget.
type CreateBudgetOptions struct {
	AIDollars     float64
	ComputeHours  float64
	Window        BudgetWindow
	WindowStart   time.Time
}

// DecodedToken represents a decoded (but not verified) token.
type DecodedToken struct {
	Header  map[string]any
	Payload BudgetTokenPayload
}

// MicroDollars converts dollars to microdollars.
func MicroDollars(dollars float64) int64 {
	return int64(dollars * 1_000_000)
}

// ComputeSeconds converts hours to seconds.
func ComputeSeconds(hours float64) int64 {
	return int64(hours * 3600)
}
