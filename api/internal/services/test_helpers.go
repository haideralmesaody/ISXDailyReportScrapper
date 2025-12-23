package services

import (
	"github.com/stretchr/testify/mock"
)

// MockWebSocketHub is a mock for WebSocketHub interface
type MockWebSocketHub struct {
	mock.Mock
}

func (m *MockWebSocketHub) Broadcast(messageType string, data interface{}) {
	m.Called(messageType, data)
}