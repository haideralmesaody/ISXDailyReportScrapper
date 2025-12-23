package websocket

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func setupHubConnection(t *testing.T) (*Hub, *websocket.Conn, func()) {
	t.Helper()

	hub := NewHub(newTestLogger())
	hub.Start()

	// Ensure the hub is stopped after the test
	t.Cleanup(hub.Stop)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		ServeWS(hub, conn)
	}))

	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	cleanup := func() {
		_ = conn.Close()
	}

	return hub, conn, cleanup
}

func TestHubRegistersAndUnregistersClients(t *testing.T) {
	hub, conn, cleanup := setupHubConnection(t)
	defer cleanup()

	require.Eventually(t, func() bool {
		return hub.ClientCount() == 1
	}, 3*time.Second, 10*time.Millisecond)

	// Closing the connection should unregister the client.
	require.NoError(t, conn.Close())

	require.Eventually(t, func() bool {
		return hub.ClientCount() == 0
	}, 3*time.Second, 10*time.Millisecond)
}

func TestHubBroadcastStatus(t *testing.T) {
	hub, conn, cleanup := setupHubConnection(t)
	defer cleanup()

	require.Eventually(t, func() bool {
		return hub.ClientCount() == 1
	}, 3*time.Second, 10*time.Millisecond)

	hub.BroadcastStatus("ready", "pipeline initialized")

	var message map[string]interface{}
	for attempts := 0; attempts < 3; attempts++ {
		require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))
		_, payload, err := conn.ReadMessage()
		require.NoError(t, err)

		require.NoError(t, json.Unmarshal(payload, &message))
		if message["type"] == "status" {
			break
		}
	}
	defer conn.SetReadDeadline(time.Time{})

	require.Equal(t, "status", message["type"])

	data, ok := message["data"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "ready", data["status"])
	require.Equal(t, "pipeline initialized", data["message"])
}
