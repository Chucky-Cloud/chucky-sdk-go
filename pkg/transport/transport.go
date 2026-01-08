// Package transport provides the transport layer abstraction for the Chucky SDK.
package transport

import (
	"github.com/chucky-cloud/chucky-sdk-go/pkg/types"
)

// ConnectionStatus represents the current connection state.
type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusConnected    ConnectionStatus = "connected"
	StatusReconnecting ConnectionStatus = "reconnecting"
	StatusError        ConnectionStatus = "error"
)

// TransportEvents contains callbacks for transport events.
type TransportEvents struct {
	OnMessage      func(msg types.IncomingMessage)
	OnClose        func(code int, reason string)
	OnStatusChange func(status ConnectionStatus)
	OnError        func(err error)
}

// Transport defines the interface for SDK message transport.
type Transport interface {
	// Status returns the current connection status.
	Status() ConnectionStatus

	// Connect establishes the connection.
	Connect() error

	// Disconnect closes the connection.
	Disconnect() error

	// Send sends a message through the transport.
	Send(msg types.OutgoingMessage) error

	// SetEventHandlers sets the callbacks for transport events.
	SetEventHandlers(handlers TransportEvents)

	// WaitForReady blocks until the connection is ready.
	WaitForReady() error
}
