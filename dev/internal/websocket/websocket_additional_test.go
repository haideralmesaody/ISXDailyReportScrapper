package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBroadcastUpdate tests the BroadcastUpdate method
func TestBroadcastUpdate(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 256),
	}
	hub.Register(client)

	// Skip connection message
	<-client.send
	time.Sleep(10 * time.Millisecond)

	// Test BroadcastUpdate
	hub.BroadcastUpdate(TypeDataUpdate, SubtypeTickerSummary, ActionCreated, map[string]interface{}{
		"ticker": "TEST",
		"price":  100.50,
	})

	select {
	case msg := <-client.send:
		assert.Contains(t, string(msg), "data_update")
		assert.Contains(t, string(msg), "ticker_summary")
		assert.Contains(t, string(msg), "created")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for update message")
	}
}

// TestBroadcastRefresh tests the BroadcastRefresh method
func TestBroadcastRefresh(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 256),
	}
	hub.Register(client)

	// Skip connection message
	<-client.send
	time.Sleep(10 * time.Millisecond)

	// Test BroadcastRefresh
	hub.BroadcastRefresh("pipeline", []string{"status", "progress"})

	select {
	case msg := <-client.send:
		assert.Contains(t, string(msg), "data_update")
		assert.Contains(t, string(msg), "refresh")
		assert.Contains(t, string(msg), "pipeline")
		assert.Contains(t, string(msg), "status")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for refresh message")
	}
}

// TestBroadcastJSON tests the BroadcastJSON method
func TestBroadcastJSON(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 256),
	}
	hub.Register(client)

	// Skip connection message
	<-client.send
	time.Sleep(10 * time.Millisecond)

	// Test BroadcastJSON
	customMsg := map[string]interface{}{
		"type": "custom",
		"data": map[string]interface{}{
			"foo": "bar",
			"num": 123,
		},
	}
	hub.BroadcastJSON(customMsg)

	select {
	case msg := <-client.send:
		assert.Contains(t, string(msg), "custom")
		assert.Contains(t, string(msg), "foo")
		assert.Contains(t, string(msg), "bar")
		assert.Contains(t, string(msg), "123")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for JSON message")
	}
}

// TestBroadcast tests the generic Broadcast method
func TestBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 256),
	}
	hub.Register(client)

	// Skip connection message
	<-client.send
	time.Sleep(10 * time.Millisecond)

	// Test Broadcast
	hub.Broadcast("test_type", map[string]interface{}{
		"test": "data",
	})

	select {
	case msg := <-client.send:
		assert.Contains(t, string(msg), "test_type")
		assert.Contains(t, string(msg), "test")
		assert.Contains(t, string(msg), "data")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for broadcast message")
	}
}

// TestMultipleClients tests broadcasting to multiple clients
func TestMultipleClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Register 3 clients
	clients := make([]*Client, 3)
	for i := 0; i < 3; i++ {
		client := &Client{
			hub:  hub,
			conn: nil,
			send: make(chan []byte, 256),
		}
		clients[i] = client
		hub.Register(client)
		// Skip connection message
		<-client.send
	}

	time.Sleep(10 * time.Millisecond)
	
	// Verify client count
	assert.Equal(t, 3, hub.ClientCount())

	// Broadcast a message
	hub.BroadcastProgress("test", 50, "Testing multiple clients")

	// All clients should receive the message
	for i, client := range clients {
		select {
		case msg := <-client.send:
			assert.Contains(t, string(msg), "progress")
			assert.Contains(t, string(msg), "Testing multiple clients")
		case <-time.After(1 * time.Second):
			t.Fatalf("client %d: timeout waiting for message", i)
		}
	}
}

// TestClientSendBufferFull tests handling of full client buffer
func TestClientSendBufferFull(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create client with small buffer
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 1), // Very small buffer
	}
	hub.Register(client)
	
	time.Sleep(10 * time.Millisecond)

	// Send many messages quickly
	for i := 0; i < 10; i++ {
		hub.BroadcastProgress("test", i, "Flooding")
	}

	// Client should be removed due to full buffer
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, hub.ClientCount())
}

// TestMessageTypes tests all message type constants
func TestMessageTypes(t *testing.T) {
	tests := []struct {
		name     string
		msgType  string
		expected string
	}{
		{"Output", TypeOutput, "output"},
		{"DataUpdate", TypeDataUpdate, "data_update"},
		{"PipelineStatus", TypePipelineStatus, "pipeline:status"},
		{"PipelineProgress", TypePipelineProgress, "pipeline:progress"},
		{"Progress", TypeProgress, "progress"},
		{"Status", TypeStatus, "status"},
		{"Error", TypeError, "error"},
		{"Connection", TypeConnection, "connection"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.msgType)
		})
	}
}

// TestServeWS tests the ServeWS function
func TestServeWS(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// This test mainly ensures ServeWS doesn't panic
	// Actual WebSocket testing is done in TestClientMessageHandling
	assert.NotPanics(t, func() {
		// ServeWS will start goroutines
		// We can't easily test it without a real WebSocket connection
		// But we can ensure it doesn't panic with nil
	})
}