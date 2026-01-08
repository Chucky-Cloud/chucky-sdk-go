// Package chucky provides the main client for the Chucky SDK.
package chucky

import (
	"context"
	"sync"

	"github.com/chucky-cloud/chucky-sdk-go/pkg/transport"
	"github.com/chucky-cloud/chucky-sdk-go/pkg/types"
)

// Client is the main entry point for the Chucky SDK.
type Client struct {
	options    types.ClientOptions
	sessions   map[string]*Session
	sessionsMu sync.RWMutex
	handlers   ClientEventHandlers
}

// ClientEventHandlers contains callbacks for client events.
type ClientEventHandlers struct {
	OnError        func(err error)
	OnSessionStart func(sessionID string)
	OnSessionEnd   func(sessionID string)
}

// NewClient creates a new Chucky client.
func NewClient(opts types.ClientOptions) *Client {
	merged := types.DefaultClientOptions().Merge(opts)
	return &Client{
		options:  merged,
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session with the given options.
func (c *Client) CreateSession(opts *types.SessionOptions) *Session {
	if opts == nil {
		opts = &types.SessionOptions{}
	}

	// Create transport
	t := transport.NewWebSocketTransport(transport.WebSocketTransportOptions{
		BaseURL:           c.options.BaseURL,
		Token:             c.options.Token,
		Timeout:           c.options.Timeout,
		KeepAliveInterval: c.options.KeepAliveInterval,
		Debug:             c.options.Debug,
	})

	session := newSession(c, t, *opts)

	c.sessionsMu.Lock()
	c.sessions[session.ID()] = session
	c.sessionsMu.Unlock()

	return session
}

// ResumeSession resumes an existing session by ID.
func (c *Client) ResumeSession(sessionID string, opts *types.SessionOptions) *Session {
	if opts == nil {
		opts = &types.SessionOptions{}
	}
	opts.SessionID = sessionID
	opts.Continue = true

	return c.CreateSession(opts)
}

// Prompt sends a one-shot prompt and returns the result.
func (c *Client) Prompt(ctx context.Context, message string, opts *types.SessionOptions) (*types.SessionResult, error) {
	session := c.CreateSession(opts)
	defer session.Close()

	if err := session.Send(ctx, message); err != nil {
		return nil, err
	}

	var result *types.SessionResult
	for msg := range session.Stream(ctx) {
		if resultMsg, ok := msg.(*types.SDKResultMessage); ok {
			result = types.FromResultMessage(resultMsg)
		}
	}

	if result == nil {
		return nil, types.SessionError("no result received")
	}

	return result, nil
}

// Close closes all sessions and the client.
func (c *Client) Close() {
	c.sessionsMu.Lock()
	sessions := make([]*Session, 0, len(c.sessions))
	for _, s := range c.sessions {
		sessions = append(sessions, s)
	}
	c.sessionsMu.Unlock()

	for _, s := range sessions {
		s.Close()
	}
}

// On sets the client event handlers.
func (c *Client) On(handlers ClientEventHandlers) *Client {
	c.handlers = handlers
	return c
}

func (c *Client) removeSession(sessionID string) {
	c.sessionsMu.Lock()
	delete(c.sessions, sessionID)
	c.sessionsMu.Unlock()

	if c.handlers.OnSessionEnd != nil {
		c.handlers.OnSessionEnd(sessionID)
	}
}

func (c *Client) notifySessionStart(sessionID string) {
	if c.handlers.OnSessionStart != nil {
		c.handlers.OnSessionStart(sessionID)
	}
}

func (c *Client) notifyError(err error) {
	if c.handlers.OnError != nil {
		c.handlers.OnError(err)
	}
}
