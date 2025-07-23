package websocket

import (
	"encoding/json"
	"log"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client registered. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered. Total clients: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send channel is full, close it
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastUpdate sends a data update message to all connected clients
func (h *Hub) BroadcastUpdate(updateType, subtype, action string, data interface{}) {
	// Use the provided updateType if it's a special type, otherwise default to data_update
	messageType := updateType
	if updateType == TypeDataUpdate || updateType == "" {
		messageType = "data_update"
	}
	
	message := map[string]interface{}{
		"type":      messageType,
		"subtype":   subtype,
		"action":    action,
		"data":      data,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling update message: %v", err)
		return
	}

	h.broadcast <- jsonData
}

// BroadcastProgress sends a progress update message
func (h *Hub) BroadcastProgress(stage string, progress int, message string) {
	update := map[string]interface{}{
		"type": TypeProgress,
		"data": map[string]interface{}{
			"stage":    stage,
			"progress": progress,
			"message":  message,
		},
	}

	jsonData, err := json.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling progress message: %v", err)
		return
	}

	h.broadcast <- jsonData
}

// BroadcastProgressWithDetails sends a detailed progress update
func (h *Hub) BroadcastProgressWithDetails(stage string, current, total int, percentage float64, message, eta string, details map[string]interface{}) {
	update := map[string]interface{}{
		"type": TypeProgress,
		"data": map[string]interface{}{
			"stage":      stage,
			"current":    current,
			"total":      total,
			"percentage": percentage,
			"message":    message,
			"eta":        eta,
			"details":    details,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling detailed progress message: %v", err)
		return
	}

	h.broadcast <- jsonData
}

// BroadcastStatus sends a status update message
func (h *Hub) BroadcastStatus(status, message string) {
	update := map[string]interface{}{
		"type": "status",
		"data": map[string]interface{}{
			"status":  status,
			"message": message,
		},
	}

	jsonData, err := json.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling status message: %v", err)
		return
	}

	h.broadcast <- jsonData
}

// BroadcastOutput sends an output message
func (h *Hub) BroadcastOutput(message, level string) {
	update := map[string]interface{}{
		"type": TypeOutput,
		"data": map[string]interface{}{
			"message": message,
			"level":   level,
		},
	}

	jsonData, err := json.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling output message: %v", err)
		return
	}

	h.broadcast <- jsonData
}

// BroadcastConnection sends a connection status message
func (h *Hub) BroadcastConnection(status string, licenseInfo interface{}) {
	message := map[string]interface{}{
		"type": TypeConnection,
		"data": map[string]interface{}{
			"status": status,
			"message": "Connected to ISX CLI Web Interface",
			"license": licenseInfo,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling connection message: %v", err)
		return
	}

	h.broadcast <- jsonData
}

// BroadcastError sends a structured error message
func (h *Hub) BroadcastError(code, message, details, stage string, recoverable bool) {
	hint := ErrorRecoveryHints[code]
	if hint == "" {
		hint = "Please try again or contact support"
	}

	errorMsg := map[string]interface{}{
		"type": TypeError,
		"data": map[string]interface{}{
			"code":        code,
			"message":     message,
			"details":     details,
			"stage":       stage,
			"recoverable": recoverable,
			"hint":        hint,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(errorMsg)
	if err != nil {
		log.Printf("Error marshaling error message: %v", err)
		return
	}

	h.broadcast <- jsonData
}

// BroadcastRefresh sends a data refresh notification (for UI updates)
func (h *Hub) BroadcastRefresh(source string, components []string) {
	h.BroadcastUpdate(TypeDataUpdate, SubtypeAll, ActionRefresh, map[string]interface{}{
		"source":     source,
		"components": components,
	})
}

// BroadcastJSON sends a pre-formatted JSON message directly
func (h *Hub) BroadcastJSON(message map[string]interface{}) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling JSON message: %v", err)
		return
	}
	
	h.broadcast <- jsonData
}