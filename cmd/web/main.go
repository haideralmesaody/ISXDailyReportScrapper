package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"isxcli/internal/license"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var startTime time.Time

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type CommandRequest struct {
	Command string            `json:"command"`
	Args    map[string]string `json:"args"`
}

type CommandResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

type WebSocketMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Command string `json:"command"`
}

type TickerSummary struct {
	Ticker      string    `json:"ticker"`
	CompanyName string    `json:"company_name"`
	LastPrice   float64   `json:"last_price"`
	LastDate    string    `json:"last_date"`
	TradingDays int       `json:"trading_days"`
	Last10Days  []float64 `json:"last_10_days"`
}

type LicenseRequest struct {
	LicenseKey string `json:"license_key"`
}

type LicenseStatus struct {
	IsValid    bool      `json:"is_valid"`
	ExpiryDate time.Time `json:"expiry_date,omitempty"`
	DaysLeft   int       `json:"days_left,omitempty"`
	Message    string    `json:"message"`
}

var (
	clients        = make(map[*websocket.Conn]bool)
	broadcast      = make(chan WebSocketMessage)
	mutex          = &sync.Mutex{}
	licenseManager *license.Manager
)

func main() {
	startTime = time.Now()
	// Initialize license manager
	var err error
	licenseManager, err = license.NewManager("license.dat")
	if err != nil {
		log.Printf("Warning: Failed to initialize license manager: %v", err)
	}

	r := mux.NewRouter()

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	// License endpoints (no middleware needed)
	r.HandleFunc("/api/license/status", handleLicenseStatus).Methods("GET")
	r.HandleFunc("/api/license/activate", handleLicenseActivate).Methods("POST")
	r.HandleFunc("/api/license/heartbeat", handleLicenseHeartbeat).Methods("POST")

	// Protected API endpoints - require valid license
	api := r.PathPrefix("/api").Subrouter()
	api.Use(licenseMiddleware)
	api.HandleFunc("/scrape", handleScrape).Methods("POST")
	api.HandleFunc("/process", handleProcess).Methods("POST")
	api.HandleFunc("/indexcsv", handleIndexCSV).Methods("POST")
	api.HandleFunc("/tickers", handleListTickers).Methods("GET")
	api.HandleFunc("/ticker/{ticker}", handleGetTicker).Methods("GET")
	api.HandleFunc("/files", handleListFiles).Methods("GET")
	api.HandleFunc("/download/{filename}", handleDownloadFile).Methods("GET")
	api.HandleFunc("/status", handleStatus).Methods("GET")

	// WebSocket endpoint (license check handled in handleWebSocket)
	r.HandleFunc("/ws", handleWebSocket)

	// Serve the main page
	r.HandleFunc("/", serveIndex)

	// Start WebSocket message broadcaster
	go handleMessages()

	// Generate ticker summary on startup
	if err := generateTickerSummary(); err != nil {
		log.Printf("Warning: Failed to generate ticker summary on startup: %v", err)
	}

	fmt.Println("🔐 ISX Web Interface (Licensed) starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func licenseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if licenseManager == nil {
			http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
			return
		}

		valid, _ := licenseManager.ValidateLicense()
		if !valid {
			// Get detailed validation state for better error messages
			validationState, _ := licenseManager.GetValidationState()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)

			response := map[string]interface{}{
				"error":    "License required",
				"code":     "LICENSE_REQUIRED",
				"redirect": "/license.html",
			}

			// Add specific guidance based on error type
			if validationState != nil {
				response["error_type"] = validationState.ErrorType

				switch validationState.ErrorType {
				case "machine_mismatch":
					response["code"] = "LICENSE_MACHINE_MISMATCH"
					response["message"] = "This license is not valid for this machine. Please contact Iraqi Investor to get a new license for this machine."
					response["contact_info"] = "Please contact Iraqi Investor for assistance"
				case "expired":
					response["code"] = "LICENSE_EXPIRED"
					response["message"] = "Your license has expired. Please contact Iraqi Investor to renew your license."
					response["contact_info"] = "Please contact Iraqi Investor for renewal"
				case "network_error":
					response["code"] = "LICENSE_NETWORK_ERROR"
					response["message"] = "Cannot verify license due to network issues. Please check your internet connection and try again."
				default:
					response["message"] = "No valid license found. Please contact Iraqi Investor to get a license."
					response["contact_info"] = "Please contact Iraqi Investor for assistance"
				}
			} else {
				response["message"] = "No valid license found. Please contact Iraqi Investor to get a license."
				response["contact_info"] = "Please contact Iraqi Investor for assistance"
			}

			json.NewEncoder(w).Encode(response)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleLicenseStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if licenseManager == nil {
		json.NewEncoder(w).Encode(LicenseStatus{
			IsValid: false,
			Message: "License system unavailable. Please contact Iraqi Investor for assistance.",
		})
		return
	}

	valid, _ := licenseManager.ValidateLicense()
	if !valid {
		// Get detailed validation state for better feedback
		validationState, _ := licenseManager.GetValidationState()

		status := LicenseStatus{
			IsValid: false,
		}

		// Add helpful information based on validation state
		if validationState != nil {
			switch validationState.ErrorType {
			case "machine_mismatch":
				status.Message = "This license is not valid for this machine. Please contact Iraqi Investor to get a new license for this machine."
			case "expired":
				status.Message = "Your license has expired. Please contact Iraqi Investor to renew your license."
			case "network_error":
				status.Message = "Cannot verify license due to network issues. Please check your internet connection and try again."
			default:
				status.Message = "No valid license found. Please contact Iraqi Investor to get a license."
			}
		} else {
			status.Message = "No valid license found. Please contact Iraqi Investor to get a license."
		}

		json.NewEncoder(w).Encode(status)
		return
	}

	// Valid license - get license info
	info, err := licenseManager.GetLicenseInfo()
	if err != nil {
		json.NewEncoder(w).Encode(LicenseStatus{
			IsValid: false,
			Message: "Failed to get license information. Please contact Iraqi Investor for assistance.",
		})
		return
	}

	daysLeft := int(time.Until(info.ExpiryDate).Hours() / 24)

	json.NewEncoder(w).Encode(LicenseStatus{
		IsValid:    true,
		ExpiryDate: info.ExpiryDate,
		DaysLeft:   daysLeft,
		Message:    fmt.Sprintf("License valid for %d days", daysLeft),
	})
}

func handleLicenseActivate(w http.ResponseWriter, r *http.Request) {
	var req LicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if licenseManager == nil {
		http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
		return
	}

	if err := licenseManager.ActivateLicense(req.LicenseKey); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Get license info after activation
	info, err := licenseManager.GetLicenseInfo()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "License activated successfully",
			"success":  true,
			"redirect": true,
		})
		return
	}

	daysLeft := int(time.Until(info.ExpiryDate).Hours() / 24)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "License activated successfully",
		"success":     true,
		"redirect":    true,
		"days_left":   daysLeft,
		"expiry_date": info.ExpiryDate.Format("January 2, 2006"),
		"duration":    info.Duration,
		"user_email":  info.UserEmail,
	})
}

func handleLicenseHeartbeat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if licenseManager == nil {
		http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
		return
	}

	// First check if license is valid
	valid, err := licenseManager.ValidateLicense()
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "License invalid or expired",
			"message": err.Error(),
		})
		return
	}

	// Update last connected time in Google Sheets
	if err := licenseManager.UpdateLastConnected(); err != nil {
		// Don't fail the request if Google Sheets update fails
		// Just log it and continue
		log.Printf("Warning: Failed to update last connected time: %v", err)
	}

	// Return success response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Heartbeat recorded",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	// Check license status first
	if licenseManager != nil {
		valid, err := licenseManager.ValidateLicense()
		if !valid {
			// Log the reason for debugging
			log.Printf("License invalid, serving license page. Error: %v", err)
			// Serve license activation page
			http.ServeFile(w, r, "./web/license.html")
			return
		}

		// License is valid, get license info and log success
		if info, err := licenseManager.GetLicenseInfo(); err == nil {
			daysLeft := int(time.Until(info.ExpiryDate).Hours() / 24)
			log.Printf("License valid, serving main application. License expires in %d days (%s)", daysLeft, info.ExpiryDate.Format("2006-01-02"))
		} else {
			log.Printf("License valid, serving main application")
		}
	} else {
		// No license manager, serve license page
		log.Printf("License manager not available, serving license page")
		http.ServeFile(w, r, "./web/license.html")
		return
	}

	// Serve main application
	http.ServeFile(w, r, "./web/index.html")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check license before allowing WebSocket connection
	if licenseManager != nil {
		if valid, _ := licenseManager.ValidateLicense(); !valid {
			http.Error(w, "License required", http.StatusUnauthorized)
			return
		}
	} else {
		http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	// Send welcome message with license info
	if licenseManager != nil {
		if info, err := licenseManager.GetLicenseInfo(); err == nil {
			daysLeft := int(time.Until(info.ExpiryDate).Hours() / 24)
			conn.WriteJSON(WebSocketMessage{
				Type:    "info",
				Message: fmt.Sprintf("Connected to ISX CLI Web Interface (Licensed - %d days remaining)", daysLeft),
			})
		}
	} else {
		conn.WriteJSON(WebSocketMessage{
			Type:    "info",
			Message: "Connected to ISX CLI Web Interface",
		})
	}

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func broadcastMessage(msgType, message, command string) {
	broadcast <- WebSocketMessage{
		Type:    msgType,
		Message: message,
		Command: command,
	}
}

func handleScrape(w http.ResponseWriter, r *http.Request) {
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build command arguments
	args := []string{}
	if mode := req.Args["mode"]; mode != "" {
		args = append(args, "--mode="+mode)
	}
	if from := req.Args["from"]; from != "" {
		args = append(args, "--from="+from)
	}
	if to := req.Args["to"]; to != "" {
		args = append(args, "--to="+to)
	}
	if headless := req.Args["headless"]; headless != "" {
		args = append(args, "--headless="+headless)
	}

	response := executeCommand("./isxcli.exe", args, "scrape")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	args := []string{}
	if inDir := req.Args["in"]; inDir != "" {
		args = append(args, "-in="+inDir)
	}
	if mode := req.Args["mode"]; mode == "full" {
		args = append(args, "-full")
	}

	response := executeCommandWithStreaming("./cmd/process/process.exe", args, "process")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleIndexCSV(w http.ResponseWriter, r *http.Request) {
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	args := []string{}
	if mode := req.Args["mode"]; mode != "" {
		args = append(args, "-mode="+mode)
	}
	if dir := req.Args["dir"]; dir != "" {
		args = append(args, "-dir="+dir)
	}
	if out := req.Args["out"]; out != "" {
		args = append(args, "-out="+out)
	}

	response := executeCommandWithStreaming("./cmd/indexcsv/indexcsv.exe", args, "indexcsv")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleListTickers(w http.ResponseWriter, r *http.Request) {
	summaryFile := "reports/ticker_summary.csv"

	// Check if ticker summary exists
	if _, err := os.Stat(summaryFile); os.IsNotExist(err) {
		// Try to generate it
		if genErr := generateTickerSummary(); genErr != nil {
			http.Error(w, fmt.Sprintf("Ticker summary not available: %v", genErr), http.StatusInternalServerError)
			return
		}
	}

	// Read ticker summary CSV
	file, err := os.Open(summaryFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open ticker summary: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read ticker summary: %v", err), http.StatusInternalServerError)
		return
	}

	if len(records) < 2 {
		response := map[string]interface{}{
			"tickers": []TickerSummary{},
			"count":   0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse ticker summaries
	var summaries []TickerSummary
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 6 {
			continue
		}

		lastPrice, _ := strconv.ParseFloat(record[2], 64)
		tradingDays, _ := strconv.Atoi(record[4])

		// Parse last 10 days
		var last10Days []float64
		if record[5] != "" {
			priceStrs := strings.Split(record[5], ",")
			for _, priceStr := range priceStrs {
				price, _ := strconv.ParseFloat(strings.TrimSpace(priceStr), 64)
				last10Days = append(last10Days, price)
			}
		}

		summary := TickerSummary{
			Ticker:      record[0],
			CompanyName: record[1],
			LastPrice:   lastPrice,
			LastDate:    record[3],
			TradingDays: tradingDays,
			Last10Days:  last10Days,
		}

		summaries = append(summaries, summary)
	}

	response := map[string]interface{}{
		"tickers": summaries,
		"count":   len(summaries),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetTicker(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticker := vars["ticker"]

	if ticker == "" {
		http.Error(w, "Ticker parameter is required", http.StatusBadRequest)
		return
	}

	// Construct file path - try both formats
	filePath := filepath.Join("reports", ticker+".csv")

	// Check if direct file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Try with _trading_history suffix
		filePath = filepath.Join("reports", ticker+"_trading_history.csv")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Ticker file not found: %s", ticker), http.StatusNotFound)
			return
		}
	}

	// Serve the CSV file
	http.ServeFile(w, r, filePath)
}

func handleListFiles(w http.ResponseWriter, r *http.Request) {
	files := make(map[string][]string)

	// List downloads
	if downloadFiles, err := listDirectory("downloads"); err == nil {
		files["downloads"] = downloadFiles
	}

	// List generated files
	if csvFiles, err := listDirectory("."); err == nil {
		var filtered []string
		for _, file := range csvFiles {
			if strings.HasSuffix(file, ".csv") || strings.HasSuffix(file, ".json") {
				filtered = append(filtered, file)
			}
		}
		files["generated"] = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// Security check - prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	// Check if file exists in downloads or current directory
	var filePath string
	if _, err := os.Stat(filepath.Join("downloads", filename)); err == nil {
		filePath = filepath.Join("downloads", filename)
	} else if _, err := os.Stat(filename); err == nil {
		filePath = filename
	} else {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, filePath)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
		"commands":  []string{"scrape", "process", "indexcsv"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func executeCommand(command string, args []string, commandType string) CommandResponse {
	broadcastMessage("info", fmt.Sprintf("Starting %s command: %s %s", commandType, command, strings.Join(args, " ")), commandType)

	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()

	response := CommandResponse{
		Success: err == nil,
		Output:  string(output),
	}

	if err != nil {
		response.Error = err.Error()
		broadcastMessage("error", fmt.Sprintf("Command failed: %s", err.Error()), commandType)
	} else {
		broadcastMessage("success", fmt.Sprintf("Command completed successfully"), commandType)
	}

	broadcastMessage("output", string(output), commandType)

	return response
}

func executeCommandWithStreaming(command string, args []string, commandType string) CommandResponse {
	broadcastMessage("info", fmt.Sprintf("Starting %s command: %s %s", commandType, command, strings.Join(args, " ")), commandType)

	cmd := exec.Command(command, args...)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		response := CommandResponse{
			Success: false,
			Error:   err.Error(),
		}
		broadcastMessage("error", fmt.Sprintf("Failed to create stdout pipe: %s", err.Error()), commandType)
		return response
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		response := CommandResponse{
			Success: false,
			Error:   err.Error(),
		}
		broadcastMessage("error", fmt.Sprintf("Failed to create stderr pipe: %s", err.Error()), commandType)
		return response
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		response := CommandResponse{
			Success: false,
			Error:   err.Error(),
		}
		broadcastMessage("error", fmt.Sprintf("Failed to start command: %s", err.Error()), commandType)
		return response
	}

	// Create a goroutine to read and broadcast stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			broadcastMessage("output", line, commandType)
		}
	}()

	// Create a goroutine to read and broadcast stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			broadcastMessage("error", line, commandType)
		}
	}()

	// Wait for the command to complete
	err = cmd.Wait()

	response := CommandResponse{
		Success: err == nil,
		Output:  "Command output streamed via WebSocket",
	}

	if err != nil {
		response.Error = err.Error()
		broadcastMessage("error", fmt.Sprintf("Command failed: %s", err.Error()), commandType)
	} else {
		broadcastMessage("success", fmt.Sprintf("Command completed successfully"), commandType)
	}

	return response
}

func listDirectory(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// getActualLastTradingDate reads individual ticker file to get actual last trading date
func getActualLastTradingDate(ticker string) (string, float64) {
	// Try both possible file formats
	possibleFiles := []string{
		filepath.Join("reports", ticker+".csv"),
		filepath.Join("reports", ticker+"_trading_history.csv"),
	}

	for _, filePath := range possibleFiles {
		if _, err := os.Stat(filePath); err == nil {
			// File exists, read it
			file, err := os.Open(filePath)
			if err != nil {
				continue
			}
			defer file.Close()

			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil || len(records) < 2 {
				continue
			}

			// Find the last trading day (tradingStatus = true)
			// CSV format: Date,CompanyName,Symbol,OpenPrice,HighPrice,LowPrice,AveragePrice,PrevAveragePrice,ClosePrice,PrevClosePrice,Change,ChangePercent,NumTrades,Volume,Value,TradingStatus
			for i := len(records) - 1; i >= 1; i-- {
				record := records[i]
				if len(record) >= 16 {
					date := strings.TrimSpace(record[0])
					closePrice := strings.TrimSpace(record[8])
					tradingStatus := strings.TrimSpace(record[15])

					// Return the last actual trading day
					if tradingStatus == "true" && date != "" && closePrice != "" {
						price, _ := strconv.ParseFloat(closePrice, 64)
						return date, price
					}
				}
			}
		}
	}

	return "", 0
}

// getActualLast10TradingDays reads individual ticker file to get last 10 trading days' close prices
func getActualLast10TradingDays(ticker string) []float64 {
	// Try both possible file formats
	possibleFiles := []string{
		filepath.Join("reports", ticker+".csv"),
		filepath.Join("reports", ticker+"_trading_history.csv"),
	}

	for _, filePath := range possibleFiles {
		if _, err := os.Stat(filePath); err == nil {
			// File exists, read it
			file, err := os.Open(filePath)
			if err != nil {
				continue
			}
			defer file.Close()

			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil || len(records) < 2 {
				continue
			}

			// Collect trading days (tradingStatus = true) from most recent backwards
			var tradingDayPrices []float64

			// CSV format: Date,CompanyName,Symbol,OpenPrice,HighPrice,LowPrice,AveragePrice,PrevAveragePrice,ClosePrice,PrevClosePrice,Change,ChangePercent,NumTrades,Volume,Value,TradingStatus
			for i := len(records) - 1; i >= 1 && len(tradingDayPrices) < 10; i-- {
				record := records[i]
				if len(record) >= 16 {
					closePrice := strings.TrimSpace(record[8])
					tradingStatus := strings.TrimSpace(record[15])

					// Only include actual trading days
					if tradingStatus == "true" && closePrice != "" {
						price, err := strconv.ParseFloat(closePrice, 64)
						if err == nil {
							tradingDayPrices = append(tradingDayPrices, price)
						}
					}
				}
			}

			// Reverse the array to get chronological order (oldest to newest)
			for i, j := 0, len(tradingDayPrices)-1; i < j; i, j = i+1, j-1 {
				tradingDayPrices[i], tradingDayPrices[j] = tradingDayPrices[j], tradingDayPrices[i]
			}

			return tradingDayPrices
		}
	}

	return []float64{}
}

// generateTickerSummary creates a ticker summary CSV from the combined CSV file
func generateTickerSummary() error {
	combinedFile := "reports/isx_combined_data.csv"
	summaryFile := "reports/ticker_summary.csv"

	// Check if combined file exists
	if _, err := os.Stat(combinedFile); os.IsNotExist(err) {
		return fmt.Errorf("combined CSV file not found: %s", combinedFile)
	}

	// Read combined CSV
	file, err := os.Open(combinedFile)
	if err != nil {
		return fmt.Errorf("failed to open combined file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read combined CSV: %v", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("combined CSV has no data rows")
	}

	// Parse header to find column indices
	header := records[0]
	tickerCol := -1
	companyCol := -1
	dateCol := -1
	closeCol := -1

	for i, col := range header {
		switch strings.ToLower(col) {
		case "ticker", "company_symbol", "symbol":
			tickerCol = i
		case "company_name", "companyname", "company", "name":
			companyCol = i
		case "date":
			dateCol = i
		case "close_price", "closeprice", "close":
			closeCol = i
		}
	}

	if tickerCol == -1 || companyCol == -1 || dateCol == -1 || closeCol == -1 {
		return fmt.Errorf("required columns not found in combined CSV. Found: %v", header)
	}

	// Group data by ticker
	tickerData := make(map[string][]map[string]string)

	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) <= tickerCol || len(record) <= companyCol || len(record) <= dateCol || len(record) <= closeCol {
			continue
		}

		ticker := strings.TrimSpace(record[tickerCol])
		if ticker == "" {
			continue
		}

		rowData := map[string]string{
			"ticker":       ticker,
			"company_name": strings.TrimSpace(record[companyCol]),
			"date":         strings.TrimSpace(record[dateCol]),
			"close_price":  strings.TrimSpace(record[closeCol]),
		}

		tickerData[ticker] = append(tickerData[ticker], rowData)
	}

	// Create ticker summaries with actual last trading dates from individual files
	var summaries []TickerSummary

	for ticker, data := range tickerData {
		if len(data) == 0 {
			continue
		}

		// Sort by date
		sort.Slice(data, func(i, j int) bool {
			return data[i]["date"] < data[j]["date"]
		})

		lastRecord := data[len(data)-1]
		lastPrice, _ := strconv.ParseFloat(lastRecord["close_price"], 64)

		// Get actual last trading date from individual ticker file
		actualLastDate, actualLastPrice := getActualLastTradingDate(ticker)
		if actualLastDate != "" {
			lastRecord["date"] = actualLastDate
			if actualLastPrice > 0 {
				lastPrice = actualLastPrice
			}
		}

		// Get actual last 10 trading days from individual ticker file
		last10Days := getActualLast10TradingDays(ticker)

		// Fallback to combined data if individual file data is not available
		if len(last10Days) == 0 {
			start := len(data) - 10
			if start < 0 {
				start = 0
			}

			for i := start; i < len(data); i++ {
				price, _ := strconv.ParseFloat(data[i]["close_price"], 64)
				last10Days = append(last10Days, price)
			}
		}

		summary := TickerSummary{
			Ticker:      ticker,
			CompanyName: lastRecord["company_name"],
			LastPrice:   lastPrice,
			LastDate:    lastRecord["date"],
			TradingDays: len(data),
			Last10Days:  last10Days,
		}

		summaries = append(summaries, summary)
	}

	// Sort summaries by ticker
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Ticker < summaries[j].Ticker
	})

	// Write ticker summary CSV
	outFile, err := os.Create(summaryFile)
	if err != nil {
		return fmt.Errorf("failed to create summary file: %v", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Ticker", "CompanyName", "LastPrice", "LastDate", "TradingDays", "Last10Days"})

	// Write data
	for _, summary := range summaries {
		last10DaysStr := ""
		for i, price := range summary.Last10Days {
			if i > 0 {
				last10DaysStr += ","
			}
			last10DaysStr += fmt.Sprintf("%.3f", price)
		}

		writer.Write([]string{
			summary.Ticker,
			summary.CompanyName,
			fmt.Sprintf("%.3f", summary.LastPrice),
			summary.LastDate,
			fmt.Sprintf("%d", summary.TradingDays),
			last10DaysStr,
		})
	}

	log.Printf("Generated ticker summary with %d tickers", len(summaries))
	return nil
}
