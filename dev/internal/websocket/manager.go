package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Additional Message types not in types.go
const (
	// Keep only unique constants not already in types.go
)

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	Command   string                 `json:"command,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Level     string                 `json:"level,omitempty"`
	Stage     string                 `json:"stage,omitempty"`
	Progress  int                    `json:"progress,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Manager handles all WebSocket connections and message routing
type Manager struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan Message
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
	connMutex  map[*websocket.Conn]*sync.Mutex // Mutex per connection for thread-safe writes
}

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
	return &Manager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan Message, 100),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		connMutex:  make(map[*websocket.Conn]*sync.Mutex),
	}
}

// Run starts the WebSocket manager
func (m *Manager) Run() {
	for {
		select {
		case client := <-m.register:
			m.mutex.Lock()
			m.clients[client] = true
			m.connMutex[client] = &sync.Mutex{}
			m.mutex.Unlock()
			log.Printf("[WebSocket Manager] Client connected. Total clients: %d", m.ClientCount())
			
			// Send connection success message with consistent format
			// This is the only place where connection status should be sent
			connMsg := Message{
				Type:      TypeConnection,
				Message:   "Connected to ISX WebSocket",
				Data:      map[string]interface{}{"status": "connected"},
				Timestamp: time.Now(),
			}
			log.Printf("[WebSocket Manager] Sending connection message: %+v", connMsg)
			m.SendToClient(client, connMsg)

		case client := <-m.unregister:
			m.mutex.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				delete(m.connMutex, client)
				client.Close()
			}
			m.mutex.Unlock()
			log.Printf("[WebSocket Manager] Client disconnected. Total clients: %d", m.ClientCount())

		case message := <-m.broadcast:
			m.mutex.RLock()
			clients := make([]*websocket.Conn, 0, len(m.clients))
			for client := range m.clients {
				clients = append(clients, client)
			}
			m.mutex.RUnlock()

			// Send to all clients
			for _, client := range clients {
				if err := m.SendToClient(client, message); err != nil {
					m.unregister <- client
				}
			}
		}
	}
}

// RegisterClient registers a new WebSocket client
func (m *Manager) RegisterClient(conn *websocket.Conn) {
	m.register <- conn
}

// UnregisterClient unregisters a WebSocket client
func (m *Manager) UnregisterClient(conn *websocket.Conn) {
	m.unregister <- conn
}

// Broadcast sends a message to all connected clients
func (m *Manager) Broadcast(message Message) {
	message.Timestamp = time.Now()
	select {
	case m.broadcast <- message:
	default:
		log.Printf("[WebSocket Manager] Broadcast channel full, dropping message: %+v", message)
	}
}

// SendToClient sends a message to a specific client
func (m *Manager) SendToClient(client *websocket.Conn, message Message) error {
	// Get the mutex for this connection
	m.mutex.RLock()
	mu, exists := m.connMutex[client]
	m.mutex.RUnlock()
	
	if !exists {
		return nil // Connection not registered
	}
	
	// Lock the connection-specific mutex to prevent concurrent writes
	mu.Lock()
	defer mu.Unlock()
	
	// Set write deadline
	client.SetWriteDeadline(time.Now().Add(10 * time.Second))
	
	if err := client.WriteJSON(message); err != nil {
		log.Printf("[WebSocket Manager] Write error: %v", err)
		return err
	}
	return nil
}

// ClientCount returns the number of connected clients
func (m *Manager) ClientCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.clients)
}

// Helper methods for common message types

// SendInfo sends an info message
func (m *Manager) SendInfo(command, message string) {
	m.Broadcast(Message{
		Type:    LevelInfo,
		Command: command,
		Message: message,
		Level:   LevelInfo,
	})
}

// SendSuccess sends a success message
func (m *Manager) SendSuccess(command, message string) {
	m.Broadcast(Message{
		Type:    LevelSuccess,
		Command: command,
		Message: message,
		Level:   LevelSuccess,
	})
}

// SendWarning sends a warning message
func (m *Manager) SendWarning(command, message string) {
	m.Broadcast(Message{
		Type:    LevelWarning,
		Command: command,
		Message: message,
		Level:   LevelWarning,
	})
}

// SendError sends an error message
func (m *Manager) SendError(command, message string) {
	m.Broadcast(Message{
		Type:    LevelError,
		Command: command,
		Message: message,
		Level:   LevelError,
	})
}

// SendProgress sends a progress update
func (m *Manager) SendProgress(stage, message string, progress int) {
	m.Broadcast(Message{
		Type:     TypeProgress,
		Stage:    stage,
		Message:  message,
		Progress: progress,
		Data: map[string]interface{}{
			"stage":    stage,
			"message":  message,
			"progress": progress,
		},
	})
}

// SendPipelineStatus sends pipeline status update
func (m *Manager) SendPipelineStatus(stage, status string, details map[string]interface{}) {
	m.Broadcast(Message{
		Type:    TypePipelineStatus,
		Stage:   stage,
		Message: status,
		Data: map[string]interface{}{
			"stage":   stage,
			"status":  status,
			"details": details,
		},
	})
}

// SendDataUpdate sends data update notification
func (m *Manager) SendDataUpdate(dataType string, details map[string]interface{}) {
	m.Broadcast(Message{
		Type:    TypeDataUpdate,
		Message: "Data updated: " + dataType,
		Data: map[string]interface{}{
			"type":    dataType,
			"details": details,
		},
	})
}

// SendOutput sends command output
func (m *Manager) SendOutput(command, output string) {
	m.Broadcast(Message{
		Type:    TypeOutput,
		Command: command,
		Message: output,
	})
}

// SendLog sends a log message
func (m *Manager) SendLog(level, message string, details map[string]interface{}) {
	m.Broadcast(Message{
		Type:    TypeLog,
		Level:   level,
		Message: message,
		Details: details,
	})
}