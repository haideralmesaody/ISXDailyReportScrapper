package common

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogWriter handles file-based logging with rotation
type LogWriter struct {
	mu              sync.Mutex
	baseDir         string
	fileName        string
	currentFile     *os.File
	currentDate     string
	writer          *bufio.Writer
	maxSize         int64
	currentSize     int64
	retentionDays   int
	flushInterval   time.Duration
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

// NewLogWriter creates a new log writer with rotation support
func NewLogWriter(baseDir, fileName string) (*LogWriter, error) {
	lw := &LogWriter{
		baseDir:       baseDir,
		fileName:      fileName,
		maxSize:       10 * 1024 * 1024, // 10MB default
		retentionDays: 7,
		flushInterval: time.Second,
		stopCh:        make(chan struct{}),
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open initial log file
	if err := lw.rotate(); err != nil {
		return nil, err
	}

	// Start flush goroutine
	lw.wg.Add(1)
	go lw.flushLoop()

	// Start cleanup goroutine
	lw.wg.Add(1)
	go lw.cleanupLoop()

	return lw, nil
}

// Write implements io.Writer interface
func (lw *LogWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	// Check if we need to rotate based on date
	currentDate := time.Now().Format("2006-01-02")
	if currentDate != lw.currentDate {
		if err := lw.rotate(); err != nil {
			return 0, err
		}
	}

	// Check if we need to rotate based on size
	if lw.currentSize+int64(len(p)) > lw.maxSize {
		if err := lw.rotateSize(); err != nil {
			return 0, err
		}
	}

	// Write to buffer
	n, err = lw.writer.Write(p)
	lw.currentSize += int64(n)
	return n, err
}

// rotate creates a new log file for the current date
func (lw *LogWriter) rotate() error {
	// Close existing file
	if lw.currentFile != nil {
		lw.writer.Flush()
		lw.currentFile.Close()
	}

	// Create date directory
	currentDate := time.Now().Format("2006-01-02")
	dateDir := filepath.Join(lw.baseDir, currentDate)
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return fmt.Errorf("failed to create date directory: %w", err)
	}

	// Open new file
	filePath := filepath.Join(dateDir, lw.fileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	lw.currentFile = file
	lw.currentDate = currentDate
	lw.currentSize = info.Size()
	lw.writer = bufio.NewWriterSize(file, 4096)

	return nil
}

// rotateSize creates a new log file when size limit is reached
func (lw *LogWriter) rotateSize() error {
	// Flush current buffer
	lw.writer.Flush()
	lw.currentFile.Close()

	// Rename current file with timestamp
	currentPath := filepath.Join(lw.baseDir, lw.currentDate, lw.fileName)
	timestamp := time.Now().Format("150405")
	newName := fmt.Sprintf("%s.%s", lw.fileName, timestamp)
	newPath := filepath.Join(lw.baseDir, lw.currentDate, newName)

	if err := os.Rename(currentPath, newPath); err != nil {
		return fmt.Errorf("failed to rename log file: %w", err)
	}

	// Create new file
	return lw.rotate()
}

// flushLoop periodically flushes the buffer
func (lw *LogWriter) flushLoop() {
	defer lw.wg.Done()
	ticker := time.NewTicker(lw.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lw.mu.Lock()
			if lw.writer != nil {
				lw.writer.Flush()
			}
			lw.mu.Unlock()
		case <-lw.stopCh:
			return
		}
	}
}

// cleanupLoop removes old log files
func (lw *LogWriter) cleanupLoop() {
	defer lw.wg.Done()
	
	// Run cleanup once at startup
	lw.cleanup()
	
	// Then run daily
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lw.cleanup()
		case <-lw.stopCh:
			return
		}
	}
}

// cleanup removes log directories older than retention period
func (lw *LogWriter) cleanup() {
	cutoff := time.Now().AddDate(0, 0, -lw.retentionDays)

	entries, err := os.ReadDir(lw.baseDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse date from directory name
		date, err := time.Parse("2006-01-02", entry.Name())
		if err != nil {
			continue
		}

		// Remove if older than cutoff
		if date.Before(cutoff) {
			dirPath := filepath.Join(lw.baseDir, entry.Name())
			os.RemoveAll(dirPath)
		}
	}
}

// Close flushes and closes the log writer
func (lw *LogWriter) Close() error {
	close(lw.stopCh)
	lw.wg.Wait()

	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.writer != nil {
		lw.writer.Flush()
	}
	if lw.currentFile != nil {
		return lw.currentFile.Close()
	}
	return nil
}

// MultiWriter combines multiple writers
type MultiWriter struct {
	writers []io.Writer
}

// NewMultiWriter creates a writer that duplicates writes to all provided writers
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

// Write implements io.Writer
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return len(p), nil
}