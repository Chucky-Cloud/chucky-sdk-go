package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/chucky-cloud/chucky-sdk-go/pkg/types"
)

// WebSocketTransport implements Transport using WebSocket.
type WebSocketTransport struct {
	baseURL           string
	token             string
	timeout           time.Duration
	keepAliveInterval time.Duration
	debug             bool

	conn    *websocket.Conn
	status  ConnectionStatus
	mu      sync.RWMutex
	handlers TransportEvents

	readyCh    chan struct{}
	readyOnce  sync.Once
	closeCh    chan struct{}
	closeOnce  sync.Once

	msgQueue   []types.OutgoingMessage
	queueMu    sync.Mutex
}

// WebSocketTransportOptions contains options for creating a WebSocket transport.
type WebSocketTransportOptions struct {
	BaseURL           string
	Token             string
	Timeout           time.Duration
	KeepAliveInterval time.Duration
	Debug             bool
}

// NewWebSocketTransport creates a new WebSocket transport.
func NewWebSocketTransport(opts WebSocketTransportOptions) *WebSocketTransport {
	if opts.Timeout == 0 {
		opts.Timeout = 60 * time.Second
	}
	if opts.KeepAliveInterval == 0 {
		opts.KeepAliveInterval = 5 * time.Minute
	}

	return &WebSocketTransport{
		baseURL:           opts.BaseURL,
		token:             opts.Token,
		timeout:           opts.Timeout,
		keepAliveInterval: opts.KeepAliveInterval,
		debug:             opts.Debug,
		status:            StatusDisconnected,
		readyCh:           make(chan struct{}),
		closeCh:           make(chan struct{}),
	}
}

// Status returns the current connection status.
func (t *WebSocketTransport) Status() ConnectionStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *WebSocketTransport) setStatus(status ConnectionStatus) {
	t.mu.Lock()
	oldStatus := t.status
	t.status = status
	t.mu.Unlock()

	if oldStatus != status && t.handlers.OnStatusChange != nil {
		t.handlers.OnStatusChange(status)
	}
}

// Connect establishes the WebSocket connection.
func (t *WebSocketTransport) Connect() error {
	t.setStatus(StatusConnecting)

	// Build URL with token
	u, err := url.Parse(t.baseURL)
	if err != nil {
		t.setStatus(StatusError)
		return types.ConnectionError("invalid URL").Wrap(err)
	}

	q := u.Query()
	q.Set("token", t.token)
	q.Set("type", "prompt")
	u.RawQuery = q.Encode()

	if t.debug {
		fmt.Printf("[WebSocket] Connecting to %s\n", u.String())
	}

	// Connect with timeout
	dialer := websocket.Dialer{
		HandshakeTimeout: t.timeout,
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		t.setStatus(StatusError)
		return types.ConnectionError("failed to connect").Wrap(err)
	}

	t.mu.Lock()
	t.conn = conn
	t.mu.Unlock()

	t.setStatus(StatusConnected)

	// Mark as ready
	t.readyOnce.Do(func() {
		close(t.readyCh)
	})

	// Flush queued messages
	t.flushQueue()

	// Start read loop
	go t.readLoop()

	// Start keep-alive
	go t.keepAliveLoop()

	return nil
}

// Disconnect closes the WebSocket connection.
func (t *WebSocketTransport) Disconnect() error {
	t.closeOnce.Do(func() {
		close(t.closeCh)
	})

	t.mu.Lock()
	conn := t.conn
	t.conn = nil
	t.mu.Unlock()

	if conn != nil {
		// Send close control message
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		_ = conn.Close()
	}

	t.setStatus(StatusDisconnected)
	return nil
}

// Send sends a message through the WebSocket.
func (t *WebSocketTransport) Send(msg types.OutgoingMessage) error {
	t.mu.RLock()
	conn := t.conn
	status := t.status
	t.mu.RUnlock()

	// Queue message if not connected yet
	if status == StatusConnecting || conn == nil {
		t.queueMu.Lock()
		t.msgQueue = append(t.msgQueue, msg)
		t.queueMu.Unlock()
		return nil
	}

	return t.sendImmediate(msg)
}

func (t *WebSocketTransport) sendImmediate(msg types.OutgoingMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return types.ProtocolError("failed to marshal message").Wrap(err)
	}

	if t.debug {
		fmt.Printf("[WebSocket] Sending: %s\n", string(data))
	}

	t.mu.RLock()
	conn := t.conn
	t.mu.RUnlock()

	if conn == nil {
		return types.ConnectionError("not connected")
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return types.ConnectionError("failed to send message").Wrap(err)
	}

	return nil
}

func (t *WebSocketTransport) flushQueue() {
	t.queueMu.Lock()
	queue := t.msgQueue
	t.msgQueue = nil
	t.queueMu.Unlock()

	for _, msg := range queue {
		if err := t.sendImmediate(msg); err != nil {
			if t.handlers.OnError != nil {
				t.handlers.OnError(err)
			}
		}
	}
}

func (t *WebSocketTransport) readLoop() {
	for {
		select {
		case <-t.closeCh:
			return
		default:
		}

		t.mu.RLock()
		conn := t.conn
		t.mu.RUnlock()

		if conn == nil {
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			select {
			case <-t.closeCh:
				return
			default:
			}

			if t.handlers.OnError != nil {
				t.handlers.OnError(types.ConnectionError("read error").Wrap(err))
			}
			if t.handlers.OnClose != nil {
				t.handlers.OnClose(websocket.CloseAbnormalClosure, err.Error())
			}
			return
		}

		if t.debug {
			fmt.Printf("[WebSocket] Received: %s\n", string(data))
		}

		msg, err := types.ParseIncomingMessage(data)
		if err != nil {
			if t.handlers.OnError != nil {
				t.handlers.OnError(types.ProtocolError("failed to parse message").Wrap(err))
			}
			continue
		}

		if t.handlers.OnMessage != nil {
			t.handlers.OnMessage(msg)
		}
	}
}

func (t *WebSocketTransport) keepAliveLoop() {
	ticker := time.NewTicker(t.keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.closeCh:
			return
		case <-ticker.C:
			ping := types.PingEnvelope{
				Type: types.MessageTypePing,
				Payload: types.PingPayload{
					Timestamp: time.Now().UnixMilli(),
				},
			}
			if err := t.Send(ping); err != nil {
				if t.handlers.OnError != nil {
					t.handlers.OnError(err)
				}
			}
		}
	}
}

// SetEventHandlers sets the callbacks for transport events.
func (t *WebSocketTransport) SetEventHandlers(handlers TransportEvents) {
	t.handlers = handlers
}

// WaitForReady blocks until the connection is ready.
func (t *WebSocketTransport) WaitForReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	select {
	case <-t.readyCh:
		return nil
	case <-ctx.Done():
		return types.TimeoutError("connection timeout")
	}
}
