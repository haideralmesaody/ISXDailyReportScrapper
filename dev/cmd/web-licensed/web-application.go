package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"isxcli/internal/analytics"
	"isxcli/internal/common"
	"isxcli/internal/files"
	"isxcli/internal/pipeline"
	"isxcli/internal/watcher"
	"isxcli/internal/websocket"

	"isxcli/internal/license"
	"isxcli/internal/updater"

	"isxcli/cmd/web-licensed/stages"

	"github.com/gorilla/mux"
	gorillaws "github.com/gorilla/websocket"
)

// handleDownloadFile handles file downloads
func handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Security check - prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Determine file location
	var filePath string

	// Check downloads directory first
	downloadsPath := filepath.Join(getProjectPath("data/downloads"), filename)
	if _, err := os.Stat(downloadsPath); err == nil {
		filePath = downloadsPath
	} else {
		// Check reports directory
		reportsPath := filepath.Join(getProjectPath("data/reports"), filename)
		if _, err := os.Stat(reportsPath); err == nil {
			filePath = reportsPath
		} else {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
	}

	// Set appropriate headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Set content type based on file extension
	contentType := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		contentType = "text/csv"
	} else if strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	} else if strings.HasSuffix(strings.ToLower(filename), ".json") {
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)

	// Serve the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to send file", http.StatusInternalServerError)
		return
	}
}

// handleGetTickerChart returns chart data for a specific ticker
func handleGetTickerChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	ticker := vars["ticker"]

	// Read combined data CSV
	csvPath := filepath.Join(executableDir, "data", "reports", "isx_combined_data.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ticker": ticker,
			"dates":  []string{},
			"open":   []float64{},
			"high":   []float64{},
			"low":    []float64{},
			"close":  []float64{},
			"volume": []int64{},
		})
		return
	}

	// Open CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		http.Error(w, "Failed to open data file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Parse CSV
	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		http.Error(w, "Failed to read CSV headers", http.StatusInternalServerError)
		return
	}

	// Find column indices
	dateIdx, tickerIdx := -1, -1
	openIdx, highIdx, lowIdx, closeIdx, volumeIdx := -1, -1, -1, -1, -1

	for i, header := range headers {
		switch header {
		case "Date":
			dateIdx = i
		case "Ticker", "Symbol":
			tickerIdx = i
		case "Open", "OpenPrice":
			openIdx = i
		case "High", "HighPrice":
			highIdx = i
		case "Low", "LowPrice":
			lowIdx = i
		case "Close", "ClosePrice":
			closeIdx = i
		case "Volume":
			volumeIdx = i
		}
	}

	// Collect data for ticker
	var dates []string
	var open, high, low, close []float64
	var volume []int64

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		if tickerIdx >= 0 && record[tickerIdx] == ticker {
			if dateIdx >= 0 {
				dates = append(dates, record[dateIdx])
			}
			if openIdx >= 0 {
				if val, err := strconv.ParseFloat(record[openIdx], 64); err == nil {
					open = append(open, val)
				} else {
					open = append(open, 0)
				}
			}
			if highIdx >= 0 {
				if val, err := strconv.ParseFloat(record[highIdx], 64); err == nil {
					high = append(high, val)
				} else {
					high = append(high, 0)
				}
			}
			if lowIdx >= 0 {
				if val, err := strconv.ParseFloat(record[lowIdx], 64); err == nil {
					low = append(low, val)
				} else {
					low = append(low, 0)
				}
			}
			if closeIdx >= 0 {
				if val, err := strconv.ParseFloat(record[closeIdx], 64); err == nil {
					close = append(close, val)
				} else {
					close = append(close, 0)
				}
			}
			if volumeIdx >= 0 {
				if val, err := strconv.ParseInt(record[volumeIdx], 10, 64); err == nil {
					volume = append(volume, val)
				} else {
					volume = append(volume, 0)
				}
			}
		}
	}

	// Return chart data
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ticker": ticker,
		"dates":  dates,
		"open":   open,
		"high":   high,
		"low":    low,
		"close":  close,
		"volume": volume,
	})
}

// Operation API handlers

func handleStartScrape(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] handleStartScrape called from %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	// Check license first
	valid, _ := licenseManager.ValidateLicense()
	if !valid {
		log.Printf("[ERROR] License validation failed for scrape request")
		http.Error(w, "Invalid license", http.StatusForbidden)
		return
	}

	// Parse request body
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("[ERROR] Failed to parse scrape request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Scrape request params: %+v", params)

	// Extract parameters
	mode, _ := params["mode"].(string)
	from, _ := params["from"].(string)
	to, _ := params["to"].(string)
	headless, _ := params["headless"].(bool)

	// Start scraping via pipeline manager
	go func() {
		pipelineID := time.Now().Format("pipeline-20060102-150405")
		pipelineReq := pipeline.PipelineRequest{
			ID:       pipelineID,
			Mode:     mode,
			FromDate: from,
			ToDate:   to,
			Parameters: map[string]interface{}{
				pipeline.ContextKeyFromDate:    from,
				pipeline.ContextKeyToDate:      to,
				pipeline.ContextKeyMode:        mode,
				pipeline.ContextKeyDownloadDir: filepath.Join(executableDir, "data", "downloads"),
				pipeline.ContextKeyReportDir:   filepath.Join(executableDir, "data", "reports"),
				"headless":                     headless,
			},
		}

		log.Printf("[INFO] Starting scrape pipeline: %s", pipelineID)

		// Start monitoring the pipeline
		go MonitorPipelineProgress(pipelineID, pipelineManager)

		ctx := context.Background()
		if _, err := pipelineManager.Execute(ctx, pipelineReq); err != nil {
			log.Printf("Pipeline failed: %v", err)
			// Send error to WebSocket
			wsHub.BroadcastError("PIPELINE_ERROR", fmt.Sprintf("Pipeline failed: %v", err), err.Error(), "scraping", false)
		}
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Scraping started",
	})
}

func handleStartProcess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check license first
	valid, _ := licenseManager.ValidateLicense()
	if !valid {
		http.Error(w, "Invalid license", http.StatusForbidden)
		return
	}

	// Start processing stage only
	go func() {
		pipelineID := time.Now().Format("process-20060102-150405")
		pipelineReq := pipeline.PipelineRequest{
			ID:   pipelineID,
			Mode: "process",
			Parameters: map[string]interface{}{
				pipeline.ContextKeyDownloadDir: filepath.Join(executableDir, "data", "downloads"),
				pipeline.ContextKeyReportDir:   filepath.Join(executableDir, "data", "reports"),
			},
		}

		ctx := context.Background()
		if _, err := pipelineManager.Execute(ctx, pipelineReq); err != nil {
			log.Printf("Processing failed: %v", err)
		}
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Processing started",
	})
}

func handleStartIndexExtraction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check license first
	valid, _ := licenseManager.ValidateLicense()
	if !valid {
		http.Error(w, "Invalid license", http.StatusForbidden)
		return
	}

	// Start indices stage only
	go func() {
		pipelineID := time.Now().Format("indices-20060102-150405")
		pipelineReq := pipeline.PipelineRequest{
			ID:   pipelineID,
			Mode: "indices",
			Parameters: map[string]interface{}{
				pipeline.ContextKeyDownloadDir: filepath.Join(executableDir, "data", "downloads"),
				pipeline.ContextKeyReportDir:   filepath.Join(executableDir, "data", "reports"),
			},
		}

		ctx := context.Background()
		if _, err := pipelineManager.Execute(ctx, pipelineReq); err != nil {
			log.Printf("Indices extraction failed: %v", err)
		}
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Index extraction started",
	})
}

// handleDownloadFile handles file downloads
func handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Security check - prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Determine file location
	var filePath string

	// Check downloads directory first
	downloadsPath := filepath.Join(executableDir, "data", "downloads", filename)
	if _, err := os.Stat(downloadsPath); err == nil {
		filePath = downloadsPath
	} else {
		// Check reports directory
		reportsPath := filepath.Join(executableDir, "data", "reports", filename)
		if _, err := os.Stat(reportsPath); err == nil {
			filePath = reportsPath
		} else {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
	}

	// Set appropriate headers
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Set content type based on file extension
	contentType := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		contentType = "text/csv"
	} else if strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	} else if strings.HasSuffix(strings.ToLower(filename), ".json") {
		contentType = "application/json"
	}
	w.Header().Set("Content-Type", contentType)

	// Serve the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to send file", http.StatusInternalServerError)
		return
	}
}
