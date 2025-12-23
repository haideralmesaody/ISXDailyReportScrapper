package main

import (
	"log/slog"
	"testing"
)

type broadcastCall struct {
	eventType string
	step      string
	status    string
	data      interface{}
}

type mockHub struct {
	calls []broadcastCall
}

func (m *mockHub) BroadcastUpdate(eventType, step, status string, data interface{}) {
	m.calls = append(m.calls, broadcastCall{
		eventType: eventType,
		step:      step,
		status:    status,
		data:      data,
	})
}

func TestWebSocketAdapter_Passthrough(t *testing.T) {
	t.Parallel()

	hub := &mockHub{}
	adapter := NewWebSocketAdapter(hub, slog.Default())

	payload := map[string]interface{}{"custom": "value"}
	adapter.BroadcastUpdate("operation:snapshot", "scraping", "running", payload)

	if len(hub.calls) != 1 {
		t.Fatalf("expected 1 broadcast call, got %d", len(hub.calls))
	}

	call := hub.calls[0]
	if call.eventType != "operation:snapshot" {
		t.Errorf("expected event type operation:snapshot, got %s", call.eventType)
	}
	if call.step != "scrape" {
		t.Errorf("expected step scrape, got %s", call.step)
	}
	if call.status != "running" {
		t.Errorf("expected status running, got %s", call.status)
	}

	data, ok := call.data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected metadata map, got %T", call.data)
	}
	if data["custom"] != "value" {
		t.Errorf("expected custom metadata to be preserved")
	}
	if data["step"] != "scrape" {
		t.Errorf("expected step in metadata to be scrape, got %v", data["step"])
	}
	if data["status"] != "running" {
		t.Errorf("expected status in metadata to be running, got %v", data["status"])
	}
}

func TestWebSocketAdapter_NormalizesLegacyEvent(t *testing.T) {
	t.Parallel()

	hub := &mockHub{}
	adapter := NewWebSocketAdapter(hub, slog.Default())

	adapter.BroadcastUpdate("operation_snapshot", "processing", "active", nil)

	if len(hub.calls) != 1 {
		t.Fatalf("expected 1 broadcast call, got %d", len(hub.calls))
	}

	call := hub.calls[0]
	if call.eventType != "operation:snapshot" {
		t.Errorf("expected legacy event to normalize to operation:snapshot, got %s", call.eventType)
	}
	if call.step != "process" {
		t.Errorf("expected step to normalize to process, got %s", call.step)
	}
	if call.status != "active" {
		t.Errorf("expected status active, got %s", call.status)
	}

	data, ok := call.data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected metadata map, got %T", call.data)
	}
	if data["step"] != "process" {
		t.Errorf("expected metadata step process, got %v", data["step"])
	}
	if data["status"] != "active" {
		t.Errorf("expected metadata status active, got %v", data["status"])
	}
}

func TestWebSocketAdapter_WrapsNonMapMetadata(t *testing.T) {
	t.Parallel()

	hub := &mockHub{}
	adapter := NewWebSocketAdapter(hub, slog.Default())

	adapter.BroadcastUpdate("custom_event", "indices", "queued", "raw-string")

	if len(hub.calls) != 1 {
		t.Fatalf("expected 1 broadcast call, got %d", len(hub.calls))
	}

	data, ok := hub.calls[0].data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected wrapped metadata map, got %T", hub.calls[0].data)
	}
	if data["value"] != "raw-string" {
		t.Errorf("expected metadata to wrap original value, got %v", data["value"])
	}
	if data["step"] != "index" {
		t.Errorf("expected metadata step to be index, got %v", data["step"])
	}
	if data["status"] != "queued" {
		t.Errorf("expected metadata status to be queued, got %v", data["status"])
	}
}
