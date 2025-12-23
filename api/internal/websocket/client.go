package websocket

import (
	"bytes"
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/isxcli/isxcli/internal/infrastructure"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Client metadata
	id          string
	traceID     string
	remoteAddr  string
	connectedAt time.Time

	// Logger
	logger *slog.Logger
}

// NewClient creates a new Client with dependency injection
func NewClient(hub *Hub, conn *websocket.Conn, logger *slog.Logger) *Client {
	if logger == nil {
		logger = infrastructure.GetLogger()
	}

	id := uuid.New().String()
	logger = logger.With(
		slog.String("component", "websocket.client"),
		slog.String("client_id", id),
	)

	remoteAddr := ""
	if conn != nil && conn.RemoteAddr() != nil {
		remoteAddr = conn.RemoteAddr().String()
	}

	return &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		id:          id,
		remoteAddr:  remoteAddr,
		connectedAt: time.Now(),
		logger:      logger,
	}
}

// NewClientWithTrace creates a new Client with trace ID
func NewClientWithTrace(hub *Hub, conn *websocket.Conn, traceID string, logger *slog.Logger) *Client {
	client := NewClient(hub, conn, logger)
	client.traceID = traceID
	client.logger = client.logger.With(slog.String("trace_id", traceID))
	return client
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		ctx := c.getContext()
		c.logger.InfoContext(ctx, "WebSocket client disconnected",
			slog.Duration("connection_duration", time.Since(c.connectedAt)))
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			ctx := c.getContext()
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.ErrorContext(ctx, "Unexpected WebSocket close error", slog.String("error", err.Error()))
			} else {
				c.logger.InfoContext(ctx, "WebSocket read ended", slog.String("reason", err.Error()))
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// Handle heartbeat messages
		if string(message) == `{"type":"heartbeat"}` {
			c.logger.Debug("Heartbeat received")
			continue
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.logger.ErrorContext(c.getContext(), "Error writing message", slog.String("error", err.Error()))
				return
			}

			// Send queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				select {
				case msg := <-c.send:
					c.conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
						c.logger.ErrorContext(c.getContext(), "Error writing queued message", slog.String("error", err.Error()))
						return
					}
				default:
					break
				}
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.DebugContext(c.getContext(), "Failed to send ping", slog.String("error", err.Error()))
				return
			}
		}
	}
}

// ServeWS handles websocket requests from the peer
func ServeWS(hub *Hub, conn *websocket.Conn) {
	client := NewClient(hub, conn, nil)

	// Register client immediately
	hub.register <- client
	client.logger.Info("WebSocket client connected",
		slog.String("client_id", client.id),
		slog.String("remote_addr", client.remoteAddr))

	// Start pumps
	go client.WritePump()
	go client.ReadPump()
}

// getContext returns context with trace ID if available
func (c *Client) getContext() context.Context {
	ctx := context.Background()
	if c.traceID != "" {
		ctx = infrastructure.WithTraceID(ctx, c.traceID)
	}
	return ctx
}
