// Package utils provides utility functions for the Chucky SDK.
package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chucky-cloud/chucky-sdk-go/pkg/types"
)

// CreateToken creates a new JWT token for authentication.
func CreateToken(opts types.CreateTokenOptions) (string, error) {
	expiresIn := opts.ExpiresIn
	if expiresIn == 0 {
		expiresIn = time.Hour
	}

	now := time.Now()
	payload := types.BudgetTokenPayload{
		Subject:     opts.UserID,
		Issuer:      opts.ProjectID,
		IssuedAt:    now.Unix(),
		ExpiresAt:   now.Add(expiresIn).Unix(),
		Budget:      opts.Budget,
		Permissions: opts.Permissions,
		SdkConfig:   opts.SdkConfig,
	}

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Base64URL encode
	headerB64 := base64URLEncode(headerJSON)
	payloadB64 := base64URLEncode(payloadJSON)

	// Create signature
	signingInput := headerB64 + "." + payloadB64
	signature := signHS256(signingInput, opts.Secret)
	signatureB64 := base64URLEncode(signature)

	return signingInput + "." + signatureB64, nil
}

// CreateBudget creates a budget from human-readable values.
func CreateBudget(opts types.CreateBudgetOptions) types.TokenBudget {
	windowStart := opts.WindowStart
	if windowStart.IsZero() {
		windowStart = time.Now()
	}

	return types.TokenBudget{
		AI:          types.MicroDollars(opts.AIDollars),
		Compute:     types.ComputeSeconds(opts.ComputeHours),
		Window:      opts.Window,
		WindowStart: windowStart.Format(time.RFC3339),
	}
}

// DecodeToken decodes a JWT token without verification.
func DecodeToken(token string) (*types.DecodedToken, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerJSON, err := base64URLDecode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var header map[string]any
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	var payload types.BudgetTokenPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return &types.DecodedToken{
		Header:  header,
		Payload: payload,
	}, nil
}

// VerifyToken verifies a JWT token signature.
func VerifyToken(token, secret string) (bool, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false, fmt.Errorf("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := base64URLEncode(signHS256(signingInput, secret))

	return parts[2] == expectedSignature, nil
}

// IsTokenExpired checks if a token has expired.
func IsTokenExpired(token string) (bool, error) {
	decoded, err := DecodeToken(token)
	if err != nil {
		return true, err
	}

	return time.Now().Unix() > decoded.Payload.ExpiresAt, nil
}

// ExtractProjectID extracts the project ID from an HMAC key.
// The format is expected to be: "hmac_<project_id>_<secret>"
func ExtractProjectID(hmacKey string) string {
	parts := strings.Split(hmacKey, "_")
	if len(parts) >= 2 && parts[0] == "hmac" {
		return parts[1]
	}
	return ""
}

// GetTokenExpiration returns the expiration time of a token.
func GetTokenExpiration(token string) (time.Time, error) {
	decoded, err := DecodeToken(token)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(decoded.Payload.ExpiresAt, 0), nil
}

// GetTokenBudget returns the budget from a token.
func GetTokenBudget(token string) (*types.TokenBudget, error) {
	decoded, err := DecodeToken(token)
	if err != nil {
		return nil, err
	}
	return &decoded.Payload.Budget, nil
}

// Helper functions

func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func base64URLDecode(s string) ([]byte, error) {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

func signHS256(data, secret string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return h.Sum(nil)
}
