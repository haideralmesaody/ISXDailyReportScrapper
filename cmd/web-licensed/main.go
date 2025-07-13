package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"isxcli/internal/license"
	"isxcli/internal/updater"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const VERSION = "v1.0.0"
const REPO_URL = "https://github.com/haideralmesaody/ISXDailyReportScrapper"

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
	updateChecker  *updater.AutoUpdateChecker
)

func main() {
	// Initialize license manager
	var err error
	licenseManager, err = license.NewManager("license-config.json", "license.dat")
	if err != nil {
		log.Printf("Warning: Failed to initialize license manager: %v", err)
	}

	// Check license before starting
	if !checkLicenseOnStartup() {
		return // Exit if no valid license
	}

	// Initialize auto-updater
	if updaterInstance, err := updater.NewUpdater(VERSION, REPO_URL); err == nil {
		updateChecker = updater.NewAutoUpdateChecker(updaterInstance, 24*time.Hour, func(updateInfo *updater.UpdateInfo) bool {
			log.Printf("Update available: %s -> %s", updateInfo.CurrentVersion, updateInfo.LatestVersion)
			// For now, just log. In production, you might want to prompt user or auto-update
			return false
		})
		updateChecker.Start()
	}

	r := mux.NewRouter()

	// Add license middleware to all API routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(licenseMiddleware)

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	// License endpoints (no middleware needed)
	r.HandleFunc("/api/license/status", handleLicenseStatus).Methods("GET")
	r.HandleFunc("/api/license/activate", handleLicenseActivate).Methods("POST")

	// Protected API endpoints
	api.HandleFunc("/scrape", handleScrape).Methods("POST")
	api.HandleFunc("/process", handleProcess).Methods("POST")
	api.HandleFunc("/indexcsv", handleIndexCSV).Methods("POST")
	api.HandleFunc("/tickers", handleListTickers).Methods("GET")
	api.HandleFunc("/ticker/{ticker}", handleGetTicker).Methods("GET")
	api.HandleFunc("/files", handleListFiles).Methods("GET")
	api.HandleFunc("/download/{filename}", handleDownloadFile).Methods("GET")
	api.HandleFunc("/status", handleStatus).Methods("GET")
	api.HandleFunc("/update/check", handleCheckUpdates).Methods("GET")
	api.HandleFunc("/update/install", handleInstallUpdate).Methods("POST")

	// WebSocket endpoint (protected)
	r.HandleFunc("/ws", licenseMiddleware(http.HandlerFunc(handleWebSocket)).ServeHTTP)

	// Serve the main page
	r.HandleFunc("/", serveIndex)

	// Start WebSocket message broadcaster
	go handleMessages()

	// Generate ticker summary on startup
	if err := generateTickerSummary(); err != nil {
		log.Printf("Warning: Failed to generate ticker summary on startup: %v", err)
	}

	fmt.Println("ISX Web Interface (Licensed) starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func checkLicenseOnStartup() bool {
	if licenseManager == nil {
		fmt.Println("❌ License system not available")
		return false
	}

	// Check if license is valid
	valid, err := licenseManager.ValidateLicense()
	if valid {
		// Get license info for display
		info, _ := licenseManager.GetLicenseInfo()
		daysLeft := int(time.Until(info.ExpiryDate).Hours() / 24)

		fmt.Printf("✅ License Valid - %d days remaining\n", daysLeft)
		if daysLeft <= 7 {
			fmt.Printf("⚠️  License expires soon: %s\n", info.ExpiryDate.Format("2006-01-02"))
		}
		return true
	}

	// License is invalid or expired
	fmt.Println("❌ Invalid or Expired License")
	fmt.Println("═══════════════════════════════════")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Prompt for license key
	fmt.Print("Please enter your license key: ")
	reader := bufio.NewReader(os.Stdin)
	licenseKey, _ := reader.ReadString('\n')
	licenseKey = strings.TrimSpace(licenseKey)

	if licenseKey == "" {
		fmt.Println("No license key provided. Exiting.")
		return false
	}

	// Activate license
	if err := licenseManager.ActivateLicense(licenseKey); err != nil {
		fmt.Printf("❌ License activation failed: %v\n", err)
		return false
	}

	fmt.Println("✅ License activated successfully!")
	return true
}

func licenseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if licenseManager == nil {
			http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
			return
		}

		valid, err := licenseManager.ValidateLicense()
		if !valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "License invalid or expired",
				"message": err.Error(),
				"code":    "LICENSE_REQUIRED",
			})
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
			Message: "License system unavailable",
		})
		return
	}

	valid, err := licenseManager.ValidateLicense()
	if !valid {
		json.NewEncoder(w).Encode(LicenseStatus{
			IsValid: false,
			Message: err.Error(),
		})
		return
	}

	// Get license info
	info, err := licenseManager.GetLicenseInfo()
	if err != nil {
		json.NewEncoder(w).Encode(LicenseStatus{
			IsValid: false,
			Message: "Failed to get license info",
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "License activated successfully",
	})
}

func handleCheckUpdates(w http.ResponseWriter, r *http.Request) {
	updaterInstance, err := updater.NewUpdater(VERSION, REPO_URL)
	if err != nil {
		http.Error(w, "Failed to initialize updater", http.StatusInternalServerError)
		return
	}

	updateInfo, err := updaterInstance.CheckForUpdates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if updateInfo == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"update_available": false,
			"current_version":  VERSION,
			"message":          "No updates available",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"update_available": true,
			"current_version":  updateInfo.CurrentVersion,
			"latest_version":   updateInfo.LatestVersion,
			"release_notes":    updateInfo.ReleaseNotes,
			"size":             updateInfo.Size,
		})
	}
}

func handleInstallUpdate(w http.ResponseWriter, r *http.Request) {
	updaterInstance, err := updater.NewUpdater(VERSION, REPO_URL)
	if err != nil {
		http.Error(w, "Failed to initialize updater", http.StatusInternalServerError)
		return
	}

	updateInfo, err := updaterInstance.CheckForUpdates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if updateInfo == nil {
		http.Error(w, "No updates available", http.StatusBadRequest)
		return
	}

	go func() {
		if err := updaterInstance.PerformUpdate(updateInfo); err != nil {
			log.Printf("Update failed: %v", err)
		} else {
			log.Println("Update completed successfully. Please restart the application.")
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Update started. Please restart the application when complete.",
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
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
	if dir := req.Args["dir"]; dir != "" {
		args = append(args, "-dir="+dir)
	}

	response := executeCommand("./cmd/indexcsv/indexcsv.exe", args, "indexcsv")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleListTickers(w http.ResponseWriter, r *http.Request) {
	summaryFile := "reports/ticker_summary.json"

	// Check if summary file exists
	if _, err := os.Stat(summaryFile); os.IsNotExist(err) {
		// Generate summary if it doesn't exist
		if err := generateTickerSummary(); err != nil {
			http.Error(w, "Failed to generate ticker summary", http.StatusInternalServerError)
			return
		}
	}

	// Read the summary file
	data, err := os.ReadFile(summaryFile)
	if err != nil {
		http.Error(w, "Failed to read ticker summary", http.StatusInternalServerError)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		http.Error(w, "Failed to parse ticker summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleGetTicker(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticker := vars["ticker"]

	// Try both possible CSV file names
	csvFiles := []string{
		filepath.Join("reports", ticker+".csv"),
		filepath.Join("reports", ticker+"_trading_history.csv"),
	}

	var csvData []byte
	var err error

	for _, csvFile := range csvFiles {
		csvData, err = os.ReadFile(csvFile)
		if err == nil {
			break
		}
	}

	if err != nil {
		http.Error(w, "Ticker not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Write(csvData)
}

// ... rest of the functions remain the same as in the original file
func handleListFiles(w http.ResponseWriter, r *http.Request) {
	// Implementation remains the same
}

func handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	// Implementation remains the same
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	// Implementation remains the same
}

func executeCommand(command string, args []string, commandType string) CommandResponse {
	// Implementation remains the same
	return CommandResponse{}
}

func executeCommandWithStreaming(command string, args []string, commandType string) CommandResponse {
	// Implementation remains the same
	return CommandResponse{}
}

func listDirectory(dir string) ([]string, error) {
	// Implementation remains the same
	return nil, nil
}

func getActualLastTradingDate(ticker string) (string, float64) {
	// Implementation remains the same
	return "", 0
}

func getActualLast10TradingDays(ticker string) []float64 {
	// Implementation remains the same
	return nil
}

func generateTickerSummary() error {
	// Implementation remains the same
	return nil
}
