package stages

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"isxcli/internal/pipeline"
	"isxcli/internal/websocket"
)

// Global WebSocket hub accessor (set by web-application.go)
var getWebSocketHub func() *websocket.Hub

// SetWebSocketHubAccessor sets the function to access the WebSocket hub
func SetWebSocketHubAccessor(f func() *websocket.Hub) {
	getWebSocketHub = f
}

// ProgressParser parses command output and updates stage progress
type ProgressParser struct {
	stageState      *pipeline.StageState
	buffer          bytes.Buffer
	filesDownloaded int
	totalFiles      int
	currentFile     int
}

// NewProgressParser creates a new progress parser
func NewProgressParser(stageState *pipeline.StageState) *ProgressParser {
	return &ProgressParser{
		stageState: stageState,
	}
}

// Write implements io.Writer interface
func (p *ProgressParser) Write(data []byte) (int, error) {
	// Log raw output for debugging
	if len(data) > 0 {
		rawOutput := string(data)
		// Store raw output in metadata for debugging
		if p.stageState.Metadata == nil {
			p.stageState.Metadata = make(map[string]interface{})
		}
		p.stageState.Metadata["last_raw_output"] = rawOutput
		
		// Log to help debug
		fmt.Printf("[SCRAPER RAW OUTPUT] %s", rawOutput)
		
		// Also send raw output as WebSocket message for debugging
		if wsHub := getWebSocketHub(); wsHub != nil {
			wsHub.BroadcastUpdate("log", "", "", map[string]interface{}{
				"message": fmt.Sprintf("[SCRAPER OUTPUT] %s", strings.TrimSpace(rawOutput)),
				"level":   "info",
			})
		}
	}
	
	n, err := p.buffer.Write(data)
	if err != nil {
		return n, err
	}
	
	// Process complete lines
	scanner := bufio.NewScanner(&p.buffer)
	for scanner.Scan() {
		line := scanner.Text()
		p.parseLine(line)
	}
	
	// Keep any incomplete line in buffer
	if scanner.Err() == nil {
		p.buffer.Reset()
		remaining := scanner.Bytes()
		if len(remaining) > 0 {
			p.buffer.Write(remaining)
		}
	}
	
	return n, nil
}

// parseLine parses a single line of output
func (p *ProgressParser) parseLine(line string) {
	// Skip empty lines
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	
	// Parse different types of messages
	switch {
	case strings.Contains(line, "[WEBSOCKET_PROGRESS]"):
		// Parse structured WebSocket progress message
		jsonStart := strings.Index(line, "{")
		if jsonStart != -1 {
			jsonData := line[jsonStart:]
			var progress map[string]interface{}
			if err := json.Unmarshal([]byte(jsonData), &progress); err == nil {
				// Extract progress data
				if pct, ok := progress["progress"].(float64); ok {
					p.stageState.Progress = pct
				}
				if msg, ok := progress["message"].(string); ok {
					p.stageState.UpdateProgress(p.stageState.Progress, msg)
				}
				if meta, ok := progress["metadata"].(map[string]interface{}); ok {
					for k, v := range meta {
						p.stageState.Metadata[k] = v
					}
				}
			}
		}
		
	case strings.Contains(line, "[WEBSOCKET_STATUS]"):
		// Parse structured WebSocket status message
		jsonStart := strings.Index(line, "{")
		if jsonStart != -1 {
			jsonData := line[jsonStart:]
			var status map[string]interface{}
			if err := json.Unmarshal([]byte(jsonData), &status); err == nil {
				if msg, ok := status["message"].(string); ok {
					p.stageState.UpdateProgress(p.stageState.Progress, msg)
				}
			}
		}
		
	case strings.Contains(line, "[INIT] Files to Download:"):
		// Extract total files to download
		if match := regexp.MustCompile(`Files to Download: (\d+)`).FindStringSubmatch(line); len(match) > 1 {
			p.totalFiles, _ = strconv.Atoi(match[1])
			p.stageState.Metadata["total_files"] = p.totalFiles
		}
		
	case strings.Contains(line, "[DOWNLOAD] File"):
		// Extract current file number
		if match := regexp.MustCompile(`File (\d+)/(\d+):`).FindStringSubmatch(line); len(match) > 2 {
			p.currentFile, _ = strconv.Atoi(match[1])
			fileTotal, _ := strconv.Atoi(match[2])
			
			// Calculate progress based on current file
			if p.totalFiles > 0 {
				progress := float64(p.filesDownloaded) / float64(p.totalFiles) * 100
				// Add some progress for the current file being downloaded
				progress += (float64(p.currentFile) / float64(fileTotal)) * (100.0 / float64(p.totalFiles))
				
				// Keep progress between 20-90 during download
				if progress < 20 {
					progress = 20
				} else if progress > 90 {
					progress = 90
				}
				
				filename := extractFilename(line)
				p.stageState.UpdateProgress(progress, fmt.Sprintf("Downloading %s", filename))
				p.stageState.Metadata["current_file"] = filename
			}
		}
		
	case strings.Contains(line, "[SUCCESS] Downloaded"):
		// Increment downloaded files count
		p.filesDownloaded++
		p.stageState.Metadata["files_downloaded"] = p.filesDownloaded
		
		if p.totalFiles > 0 {
			progress := float64(p.filesDownloaded) / float64(p.totalFiles) * 100
			if progress > 90 {
				progress = 90 // Keep some room for completion
			}
			p.stageState.UpdateProgress(progress, fmt.Sprintf("Downloaded %d of %d files", p.filesDownloaded, p.totalFiles))
		}
		
	case strings.Contains(line, "[NAVIGATE] Moving to page"):
		// Extract page number
		if match := regexp.MustCompile(`page (\d+)`).FindStringSubmatch(line); len(match) > 1 {
			page, _ := strconv.Atoi(match[1])
			p.stageState.UpdateProgress(p.stageState.Progress, fmt.Sprintf("Scanning page %d for reports...", page))
			p.stageState.Metadata["current_page"] = page
		}
		
	case strings.Contains(line, "[COMPLETE]") || strings.Contains(line, "[SUMMARY] ====== Download Complete"):
		// Download completed
		p.stageState.UpdateProgress(95, "Finalizing download...")
		
	case strings.Contains(line, "[STATUS]"):
		// General status update
		message := strings.TrimPrefix(line, "[STATUS]")
		message = strings.TrimSpace(message)
		p.stageState.UpdateProgress(p.stageState.Progress, message)
		
	case strings.Contains(line, "[ERROR]") || strings.Contains(line, "ERROR:"):
		// Store error in metadata
		p.stageState.Metadata["last_error"] = line
		
	case strings.Contains(line, "License validation failed"):
		// Handle license error specifically
		p.stageState.UpdateProgress(0, "License validation failed")
		p.stageState.Metadata["license_error"] = true
		p.stageState.Metadata["error_message"] = "License validation failed. Please check your license.dat file."
		
	case strings.Contains(line, "License system initialization failed"):
		// Handle license initialization error
		p.stageState.UpdateProgress(0, "License system initialization failed")
		p.stageState.Metadata["license_error"] = true
		p.stageState.Metadata["error_message"] = "Failed to initialize license system. Check license.dat file."
		
	case strings.Contains(line, "Contact The Iraqi Investor Group"):
		// Capture license renewal message
		p.stageState.Metadata["license_renewal_needed"] = true
	}
}

// GetFilesDownloaded returns the number of files downloaded
func (p *ProgressParser) GetFilesDownloaded() int {
	return p.filesDownloaded
}

// extractFilename extracts the filename from a download message
func extractFilename(line string) string {
	// Look for pattern like "2025 07 17 ISX Daily Report.xlsx"
	if match := regexp.MustCompile(`(\d{4} \d{2} \d{2} ISX Daily Report\.xlsx)`).FindStringSubmatch(line); len(match) > 1 {
		return match[1]
	}
	
	// Fallback to extracting after last colon
	parts := strings.Split(line, ":")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	
	return "file"
}

// ProcessingProgressParser handles progress for process.exe
type ProcessingProgressParser struct {
	*ProgressParser
	totalRecords    int
	processedRecords int
}

// NewProcessingProgressParser creates a parser for processing stage
func NewProcessingProgressParser(stageState *pipeline.StageState) *ProcessingProgressParser {
	return &ProcessingProgressParser{
		ProgressParser: NewProgressParser(stageState),
	}
}

// Write implements io.Writer interface for ProcessingProgressParser
func (p *ProcessingProgressParser) Write(data []byte) (int, error) {
	// Log raw output for debugging
	if len(data) > 0 {
		rawOutput := string(data)
		p.stageState.Metadata["last_output"] = rawOutput
		// Also log to console for debugging
		if strings.TrimSpace(rawOutput) != "" {
			p.stageState.Metadata["debug_output"] = fmt.Sprintf("[PROCESS OUTPUT] %s", rawOutput)
		}
	}
	
	n, err := p.buffer.Write(data)
	if err != nil {
		return n, err
	}
	
	// Process complete lines
	scanner := bufio.NewScanner(&p.buffer)
	for scanner.Scan() {
		line := scanner.Text()
		p.parseLine(line) // This will call ProcessingProgressParser's parseLine
	}
	
	// Keep any incomplete line in buffer
	if scanner.Err() == nil {
		p.buffer.Reset()
		remaining := scanner.Bytes()
		if len(remaining) > 0 {
			p.buffer.Write(remaining)
		}
	}
	
	return n, nil
}

// parseLine parses processing-specific output
func (p *ProcessingProgressParser) parseLine(line string) {
	// Let base parser handle common patterns
	p.ProgressParser.parseLine(line)
	
	// Handle processing-specific patterns
	switch {
	case strings.Contains(line, "Processing file"):
		if match := regexp.MustCompile(`Processing file (\d+) of (\d+):`).FindStringSubmatch(line); len(match) > 2 {
			current, _ := strconv.Atoi(match[1])
			total, _ := strconv.Atoi(match[2])
			
			progress := float64(current) / float64(total) * 100
			filename := extractFilename(line)
			p.stageState.UpdateProgress(progress, fmt.Sprintf("Processing %s (%d/%d)", filename, current, total))
			p.stageState.Metadata["current_file"] = filename
			p.stageState.Metadata["files_processed"] = current
			p.stageState.Metadata["total_files"] = total
		}
		
	case strings.Contains(line, "records processed"):
		if match := regexp.MustCompile(`(\d+) records processed`).FindStringSubmatch(line); len(match) > 1 {
			p.processedRecords, _ = strconv.Atoi(match[1])
			p.stageState.Metadata["records_processed"] = p.processedRecords
		}
		
	case strings.Contains(line, "Processing complete"):
		p.stageState.UpdateProgress(95, "Finalizing data processing...")
	}
}