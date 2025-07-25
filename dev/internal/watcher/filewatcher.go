package watcher

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"isxcli/internal/websocket"
)

// FileWatcher monitors file system changes and broadcasts updates
type FileWatcher struct {
	watchPath string
	hub       *websocket.Hub
	watcher   *fsnotify.Watcher
	debounce  map[string]time.Time
}

// NewFileWatcher creates a new file watcher instance
func NewFileWatcher(watchPath string, hub *websocket.Hub) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &FileWatcher{
		watchPath: watchPath,
		hub:       hub,
		watcher:   watcher,
		debounce:  make(map[string]time.Time),
	}, nil
}

// Start begins watching for file changes
func (fw *FileWatcher) Start() error {
	// Add the watch path
	err := fw.watcher.Add(fw.watchPath)
	if err != nil {
		return fmt.Errorf("failed to add watch path: %w", err)
	}

	log.Printf("File watcher started, monitoring: %s", fw.watchPath)

	// Start the event processing loop
	go fw.processEvents()

	return nil
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() error {
	return fw.watcher.Close()
}

// processEvents handles file system events
func (fw *FileWatcher) processEvents() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Debounce rapid changes (100ms)
			if fw.shouldDebounce(event.Name) {
				continue
			}

			// Handle the event based on operation
			if event.Op&fsnotify.Write == fsnotify.Write {
				fw.handleFileChange(event.Name, "updated")
			} else if event.Op&fsnotify.Create == fsnotify.Create {
				fw.handleFileChange(event.Name, "created")
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				fw.handleFileChange(event.Name, "deleted")
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		}
	}
}

// shouldDebounce checks if we should skip this event due to debouncing
func (fw *FileWatcher) shouldDebounce(filename string) bool {
	lastTime, exists := fw.debounce[filename]
	now := time.Now()

	// If file was modified in last 100ms, skip
	if exists && now.Sub(lastTime) < 100*time.Millisecond {
		return true
	}

	fw.debounce[filename] = now
	return false
}

// handleFileChange processes a file change event
func (fw *FileWatcher) handleFileChange(filename, action string) {
	// Get relative path and file extension
	relPath, _ := filepath.Rel(fw.watchPath, filename)
	ext := strings.ToLower(filepath.Ext(filename))
	base := filepath.Base(filename)

	// Skip non-data files
	if ext != ".csv" && ext != ".json" {
		return
	}

	// Skip temporary files
	if strings.HasPrefix(base, ".") || strings.HasPrefix(base, "~") {
		return
	}

	log.Printf("File %s: %s", action, relPath)

	// Determine the data type based on filename
	var subtype string
	switch {
	case base == "ticker_summary.json":
		subtype = websocket.SubtypeTickerSummary
	case base == "ticker_summary.csv":
		subtype = websocket.SubtypeTickerSummary
	case base == "isx_combined_data.csv":
		subtype = websocket.SubtypeCombinedData
	case base == "indexes.csv":
		subtype = websocket.SubtypeIndexes
	case strings.HasPrefix(base, "isx_daily_"):
		subtype = websocket.SubtypeDailyReport
	case strings.HasSuffix(base, "_trading_history.csv"):
		subtype = websocket.SubtypeTickerHistory
	default:
		// Unknown file type, skip
		return
	}

	// Broadcast the update
	fw.hub.BroadcastUpdate(websocket.TypeDataUpdate, subtype, action, map[string]interface{}{
		"filename": base,
		"path":     relPath,
	})

	// Also send a specific output message
	level := websocket.LevelInfo
	if action == "deleted" {
		level = websocket.LevelWarning
	}
	fw.hub.BroadcastOutput(fmt.Sprintf("File %s: %s", action, base), level)
}