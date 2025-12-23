package testutil

import (
	"log/slog"
	"testing"
)

func TestBufferedSlogHandler(t *testing.T) {
	t.Run("captures log records", func(t *testing.T) {
		logger, handler := NewTestLogger(t)
		
		logger.Info("test message", slog.String("key", "value"))
		logger.Error("error message", slog.Int("code", 500))
		
		records := handler.GetRecords()
		if len(records) != 2 {
			t.Errorf("Expected 2 records, got %d", len(records))
		}
		
		if !handler.ContainsMessage("test message") {
			t.Error("Expected to find 'test message'")
		}
		
		if !handler.ContainsAttr("key", "value") {
			t.Error("Expected to find attribute key=value")
		}
	})
	
	t.Run("filters by level", func(t *testing.T) {
		logger, handler := NewTestLogger(t)
		
		logger.Debug("debug msg")
		logger.Info("info msg")
		logger.Warn("warn msg")
		logger.Error("error msg")
		
		infoRecords := handler.GetRecordsByLevel(slog.LevelInfo)
		if len(infoRecords) != 1 {
			t.Errorf("Expected 1 info record, got %d", len(infoRecords))
		}
		
		errorRecords := handler.GetRecordsByLevel(slog.LevelError)
		if len(errorRecords) != 1 {
			t.Errorf("Expected 1 error record, got %d", len(errorRecords))
		}
	})
	
	t.Run("clear functionality", func(t *testing.T) {
		logger, handler := NewTestLogger(t)
		
		logger.Info("message 1")
		logger.Info("message 2")
		
		if handler.Count() != 2 {
			t.Errorf("Expected 2 records, got %d", handler.Count())
		}
		
		handler.Clear()
		
		if handler.Count() != 0 {
			t.Errorf("Expected 0 records after clear, got %d", handler.Count())
		}
	})
	
	t.Run("assertion helpers", func(t *testing.T) {
		logger, handler := NewTestLogger(t)
		
		logger.Info("important message", slog.String("component", "test"))
		logger.Warn("warning message", slog.Int("retry", 3))
		
		// These should pass
		AssertLogContains(t, handler, slog.LevelInfo, "important")
		AssertLogAttr(t, handler, "component", "test")
		AssertNoErrors(t, handler)
		
		// Test error case
		logger.Error("something went wrong")
		
		// This would fail if we called AssertNoErrors now
		errors := handler.GetRecordsByLevel(slog.LevelError)
		if len(errors) != 1 {
			t.Error("Expected to capture error log")
		}
	})
	
	t.Run("thread safety", func(t *testing.T) {
		logger, handler := NewTestLogger(t)
		
		// Run concurrent logging
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(n int) {
				logger.Info("concurrent log", slog.Int("goroutine", n))
				done <- true
			}(i)
		}
		
		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
		
		// Should have captured all logs
		if handler.Count() != 10 {
			t.Errorf("Expected 10 records from concurrent logging, got %d", handler.Count())
		}
	})
}