package websocket

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHubCreation tests hub initialization
func TestHubCreation(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T, hub *Hub)
	}{
		{
			name: "new hub has empty client map",
			test: func(t *testing.T, hub *Hub) {
				assert.Equal(t, 0, hub.ClientCount())
			},
		},
		{
			name: "new hub has initialized channels",
			test: func(t *testing.T, hub *Hub) {
				assert.NotNil(t, hub.clients)
				assert.NotNil(t, hub.broadcast)
				assert.NotNil(t, hub.register)
				assert.NotNil(t, hub.unregister)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := NewHub()
			tt.test(t, hub)
		})
	}
}

// TestClientRegistration tests client registration and unregistration
func TestClientRegistration(t *testing.T) {
	hub := NewHub()
	
	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				hub.Run()
			}
		}
	}()

	// Create mock WebSocket connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		client := NewClient(hub, conn)
		hub.Register(client)

		// Wait for registration
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 1, hub.ClientCount())

		// Unregister
		hub.unregister <- client
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 0, hub.ClientCount())
	}))
	defer server.Close()

	// Connect to the server
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()
}

// TestBroadcastMessages tests various broadcast methods
func TestBroadcastMessages(t *testing.T) {
	tests := []struct {
		name      string
		broadcast func(hub *Hub)
		validate  func(t *testing.T, msg map[string]interface{})
	}{
		{
			name: "broadcast progress",
			broadcast: func(hub *Hub) {
				hub.BroadcastProgress("scraping", 50, "Processing files")
			},
			validate: func(t *testing.T, msg map[string]interface{}) {
				assert.Equal(t, TypeProgress, msg["type"])
				data := msg["data"].(map[string]interface{})
				assert.Equal(t, "scraping", data["stage"])
				assert.Equal(t, float64(50), data["progress"])
				assert.Equal(t, "Processing files", data["message"])
			},
		},
		{
			name: "broadcast status",
			broadcast: func(hub *Hub) {
				hub.BroadcastStatus("active", "Pipeline running")
			},
			validate: func(t *testing.T, msg map[string]interface{}) {
				assert.Equal(t, "status", msg["type"])
				data := msg["data"].(map[string]interface{})
				assert.Equal(t, "active", data["status"])
				assert.Equal(t, "Pipeline running", data["message"])
			},
		},
		{
			name: "broadcast error",
			broadcast: func(hub *Hub) {
				hub.BroadcastError(ErrScrapingTimeout, "Connection timeout", "Network error", "scraping", true)
			},
			validate: func(t *testing.T, msg map[string]interface{}) {
				assert.Equal(t, TypeError, msg["type"])
				data := msg["data"].(map[string]interface{})
				assert.Equal(t, ErrScrapingTimeout, data["code"])
				assert.Equal(t, "Connection timeout", data["message"])
				assert.Equal(t, true, data["recoverable"])
				assert.NotEmpty(t, data["hint"])
			},
		},
		{
			name: "broadcast connection",
			broadcast: func(hub *Hub) {
				hub.BroadcastConnection("connected", map[string]interface{}{
					"valid": true,
					"type":  "pro",
				})
			},
			validate: func(t *testing.T, msg map[string]interface{}) {
				assert.Equal(t, TypeConnection, msg["type"])
				data := msg["data"].(map[string]interface{})
				assert.Equal(t, "connected", data["status"])
				license := data["license"].(map[string]interface{})
				assert.Equal(t, true, license["valid"])
			},
		},
		{
			name: "broadcast output",
			broadcast: func(hub *Hub) {
				hub.BroadcastOutput("Starting process", LevelInfo)
			},
			validate: func(t *testing.T, msg map[string]interface{}) {
				assert.Equal(t, TypeOutput, msg["type"])
				data := msg["data"].(map[string]interface{})
				assert.Equal(t, "Starting process", data["message"])
				assert.Equal(t, LevelInfo, data["level"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := NewHub()
			msgChan := make(chan []byte, 1)

			// Start hub in background
			go hub.Run()

			// Register a mock client to receive broadcasts
			client := &Client{
				hub:  hub,
				conn: nil,
				send: make(chan []byte, 256),
			}
			hub.Register(client)

			// Capture messages sent to client
			go func() {
				// Skip the first connection message
				<-client.send
				
				// Now capture the test message
				for msg := range client.send {
					select {
					case msgChan <- msg:
					default:
						// Channel full, drop message
					}
				}
			}()

			// Give hub time to register client
			time.Sleep(50 * time.Millisecond)

			// Execute broadcast
			tt.broadcast(hub)

			// Wait for message with timeout
			select {
			case msgBytes := <-msgChan:
				var msg map[string]interface{}
				err := json.Unmarshal(msgBytes, &msg)
				require.NoError(t, err)
				
				// Validate timestamp if present
				if timestamp, ok := msg["timestamp"]; ok && timestamp != nil {
					_, err = time.Parse(time.RFC3339, timestamp.(string))
					assert.NoError(t, err)
				}
				
				// Run specific validation
				tt.validate(t, msg)
			case <-time.After(1 * time.Second):
				t.Fatal("timeout waiting for broadcast message")
			}
		})
	}
}

// TestBroadcastProgressWithDetails tests detailed progress updates
func TestBroadcastProgressWithDetails(t *testing.T) {
	hub := NewHub()
	msgChan := make(chan []byte, 1)

	// Start hub
	go hub.Run()

	// Register a mock client
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 256),
	}
	hub.Register(client)

	// Capture messages
	go func() {
		// Skip the first connection message
		<-client.send
		
		// Now capture the test message
		for msg := range client.send {
			select {
			case msgChan <- msg:
			default:
			}
		}
	}()

	// Give hub time to register
	time.Sleep(50 * time.Millisecond)

	details := map[string]interface{}{
		"filesProcessed": 10,
		"totalFiles":     100,
		"currentFile":    "report_2024.xlsx",
	}

	hub.BroadcastProgressWithDetails(
		"processing", 
		10, 
		100, 
		10.0, 
		"Processing files", 
		"9m 30s",
		details,
	)

	select {
	case msgBytes := <-msgChan:
		var msg map[string]interface{}
		err := json.Unmarshal(msgBytes, &msg)
		require.NoError(t, err)

		assert.Equal(t, TypeProgress, msg["type"])
		data := msg["data"].(map[string]interface{})
		assert.Equal(t, "processing", data["stage"])
		assert.Equal(t, float64(10), data["current"])
		assert.Equal(t, float64(100), data["total"])
		assert.Equal(t, float64(10.0), data["percentage"])
		assert.Equal(t, "9m 30s", data["eta"])
		
		receivedDetails := data["details"].(map[string]interface{})
		assert.Equal(t, float64(10), receivedDetails["filesProcessed"])
		assert.Equal(t, "report_2024.xlsx", receivedDetails["currentFile"])
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for broadcast message")
	}
}

// TestClientMessageHandling tests client read/write pumps
func TestClientMessageHandling(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		
		client := NewClient(hub, conn)
		
		// Test write pump with multiple messages
		messages := []string{
			`{"type":"test","data":"message1"}`,
			`{"type":"test","data":"message2"}`,
			`{"type":"test","data":"message3"}`,
		}
		
		// Send messages
		for _, msg := range messages {
			client.send <- []byte(msg)
		}
		
		// Start pumps
		var wg sync.WaitGroup
		wg.Add(2)
		
		go func() {
			defer wg.Done()
			client.WritePump()
		}()
		
		go func() {
			defer wg.Done()
			client.ReadPump()
		}()
		
		// Keep connection alive briefly
		time.Sleep(100 * time.Millisecond)
		conn.Close()
		wg.Wait()
	}))
	defer server.Close()

	// Connect client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Read messages
	received := 0
	done := make(chan bool)
	
	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				break
			}
			
			var msg map[string]interface{}
			json.Unmarshal(message, &msg)
			if msg["type"] == "test" {
				received++
			}
			
			if received >= 3 {
				done <- true
				return
			}
		}
	}()

	select {
	case <-done:
		assert.GreaterOrEqual(t, received, 3)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for messages")
	}
}

// TestHeartbeatHandling tests heartbeat message processing
func TestHeartbeatHandling(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()
		
		ServeWS(hub, conn)
		
		// Keep connection alive
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	// Connect and send heartbeat
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	// Send heartbeat
	err = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"heartbeat"}`))
	assert.NoError(t, err)

	// Connection should remain open
	time.Sleep(100 * time.Millisecond)
	
	// Try to send another message to verify connection is still alive
	err = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"test"}`))
	assert.NoError(t, err)
}

// TestConcurrentBroadcasts tests thread safety
func TestConcurrentBroadcasts(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Track message counts
	msgCount := 0
	var mu sync.Mutex

	// Register multiple mock clients by directly adding to hub
	for i := 0; i < 5; i++ {
		client := &Client{
			hub:  hub,
			conn: nil, // Not used in broadcast
			send: make(chan []byte, 256),
		}
		hub.clients[client] = true
		
		// Consume messages from client
		go func(c *Client) {
			for range c.send {
				mu.Lock()
				msgCount++
				mu.Unlock()
			}
		}(client)
	}

	// Send concurrent broadcasts
	broadcasts := 100
	var wg sync.WaitGroup
	for i := 0; i < broadcasts; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			hub.BroadcastProgress("test", n, "Concurrent test")
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	
	// Verify messages were sent
	mu.Lock()
	assert.Greater(t, msgCount, 0, "should have received messages")
	mu.Unlock()
}

// TestErrorRecoveryHints tests error hint mapping
func TestErrorRecoveryHints(t *testing.T) {
	tests := []struct {
		code string
		hint string
	}{
		{ErrScrapingTimeout, "Check your internet connection and try again"},
		{ErrScrapingNoData, "No data available for the specified date range"},
		{ErrProcessingInvalidFile, "File may be corrupted, try re-downloading"},
		{ErrSystemDiskFull, "Not enough disk space, please free up some space"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.hint, ErrorRecoveryHints[tt.code])
		})
	}
}

// TestHubStop tests graceful shutdown
func TestHubStop(t *testing.T) {
	hub := NewHub()
	
	// Create mock client
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 256),
	}
	hub.clients[client] = true
	
	assert.Equal(t, 1, hub.ClientCount())
	
	// Stop hub
	hub.Stop()
	
	assert.Equal(t, 0, hub.ClientCount())
}

// mockConn implements a minimal websocket.Conn for testing
type mockConn struct {
	writeFunc func(messageType int, data []byte) error
}

func (m *mockConn) WriteMessage(messageType int, data []byte) error {
	if m.writeFunc != nil {
		return m.writeFunc(messageType, data)
	}
	return nil
}

func (m *mockConn) Close() error                                      { return nil }
func (m *mockConn) LocalAddr() net.Addr                               { return nil }
func (m *mockConn) RemoteAddr() net.Addr                              { return nil }
func (m *mockConn) WriteControl(messageType int, data []byte, deadline time.Time) error { return nil }
func (m *mockConn) NextWriter(messageType int) (io.WriteCloser, error) { return nil, nil }
func (m *mockConn) WritePreparedMessage(pm *websocket.PreparedMessage) error { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error                { return nil }
func (m *mockConn) NextReader() (messageType int, r io.Reader, err error) { return 0, nil, nil }
func (m *mockConn) ReadMessage() (messageType int, p []byte, err error) { return 0, nil, nil }
func (m *mockConn) SetReadDeadline(t time.Time) error                 { return nil }
func (m *mockConn) SetReadLimit(limit int64)                          {}
func (m *mockConn) CloseHandler() func(code int, text string) error   { return nil }
func (m *mockConn) SetCloseHandler(h func(code int, text string) error) {}
func (m *mockConn) PingHandler() func(appData string) error           { return nil }
func (m *mockConn) SetPingHandler(h func(appData string) error)       {}
func (m *mockConn) PongHandler() func(appData string) error           { return nil }
func (m *mockConn) SetPongHandler(h func(appData string) error)       {}
func (m *mockConn) UnderlyingConn() net.Conn                          { return nil }
func (m *mockConn) EnableWriteCompression(enable bool)                {}
func (m *mockConn) SetCompressionLevel(level int) error               { return nil }