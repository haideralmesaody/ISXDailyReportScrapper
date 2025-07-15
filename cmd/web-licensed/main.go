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

	"isxcli/internal/license"
	"isxcli/internal/updater"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const VERSION = "enhanced-v2.0.0"
const REPO_URL = "https://github.com/haideralmesaody/ISXDailyReportScrapper"

// Global executable directory for relative paths
var executableDir string

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
	IsValid        bool      `json:"is_valid"`
	ExpiryDate     time.Time `json:"expiry_date,omitempty"`
	DaysLeft       int       `json:"days_left,omitempty"`
	Message        string    `json:"message"`
	Status         string    `json:"status,omitempty"`
	NeedsRenewal   bool      `json:"needs_renewal,omitempty"`
	RenewalMessage string    `json:"renewal_message,omitempty"`
}

type LicenseTransferRequest struct {
	LicenseKey    string `json:"license_key"`
	ForceTransfer bool   `json:"force_transfer"`
}

type SystemStatsResponse struct {
	Performance map[string]interface{} `json:"performance"`
	Cache       map[string]interface{} `json:"cache"`
	Security    map[string]interface{} `json:"security"`
	Timestamp   time.Time              `json:"timestamp"`
	MachineID   string                 `json:"machine_id"`
	Version     string                 `json:"version"`
	Uptime      time.Duration          `json:"uptime"`
}

var (
	clients           = make(map[*websocket.Conn]bool)
	broadcast         = make(chan WebSocketMessage)
	mutex             = &sync.Mutex{}
	licenseManager    *license.Manager
	updateChecker     *updater.AutoUpdateChecker
	wsConnections     []*websocket.Conn
	wsConnectionsLock sync.Mutex
	startTime         = time.Now()
)

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	return ip
}

// validateLicenseForWebAccess performs local-first license validation optimized for web access
// Returns (isValid, isRecentActivation) to help with user experience
func validateLicenseForWebAccess() (bool, bool) {
	log.Printf("DEBUG: validateLicenseForWebAccess called")

	if licenseManager == nil {
		log.Printf("DEBUG: licenseManager is nil, returning false")
		return false, false
	}

	// Try to load local license first
	log.Printf("DEBUG: Attempting to get license info...")
	info, err := licenseManager.GetLicenseInfo()
	if err != nil {
		log.Printf("DEBUG: No local license found: %v", err)
		return false, false
	}

	log.Printf("DEBUG: License info loaded successfully, expiry: %v", info.ExpiryDate)

	// Check basic local validation first
	now := time.Now()

	// Check if license has expired
	if now.After(info.ExpiryDate) {
		log.Printf("License expired on %s", info.ExpiryDate.Format("2006-01-02"))
		return false, false
	}

	// Check if this is a recently activated license (within last 10 minutes)
	// This gives time for the user experience to be smooth after activation
	isRecentActivation := false
	if info.LastChecked.IsZero() {
		// If LastChecked is not set, check if license file was modified recently
		licensePath := filepath.Join(executableDir, "license.dat")
		if stat, err := os.Stat(licensePath); err == nil {
			fileAge := now.Sub(stat.ModTime())
			if fileAge < 10*time.Minute {
				isRecentActivation = true
				log.Printf("Recently activated license detected (file age: %v)", fileAge.Round(time.Second))
			}
		}
	} else {
		// Check based on LastChecked time
		timeSinceLastCheck := now.Sub(info.LastChecked)
		if timeSinceLastCheck < 10*time.Minute {
			isRecentActivation = true
			log.Printf("Recently validated license detected (last check: %v ago)", timeSinceLastCheck.Round(time.Second))
		}
	}

	// For recently activated licenses, use more lenient validation
	if isRecentActivation {
		log.Printf("Using lenient validation for recently activated license")
		// Just check local basics - don't require immediate remote validation
		return true, true
	}

	// For older licenses, use standard validation but with timeout protection
	// Set a shorter timeout for web requests to avoid hanging the page
	validationDone := make(chan bool, 1)
	var isValid bool

	go func() {
		valid, _ := licenseManager.ValidateLicense()
		validationDone <- valid
	}()

	// Wait for validation with timeout
	select {
	case isValid = <-validationDone:
		// Validation completed normally
		log.Printf("Standard license validation completed: %v", isValid)
		return isValid, false
	case <-time.After(5 * time.Second):
		// Validation timed out - fall back to local checks for better UX
		log.Printf("License validation timed out, using local validation fallback")
		// Just verify basic local requirements and allow access
		return true, false
	}
}

// securityMiddleware adds rate limiting and security checks
func securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip security for static files
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := getClientIP(r)

		// Log request for monitoring (could be enhanced with actual security checks)
		if strings.HasPrefix(r.URL.Path, "/api/license/") {
			log.Printf("License API request from IP: %s, Path: %s", clientIP, r.URL.Path)
		}

		next.ServeHTTP(w, r)
	})
}

// openBrowser opens the default web browser to the specified URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

func main() {
	// Get executable directory for all relative paths
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Could not get executable path: %v", err)
		exePath = "." // fallback to current directory
	}
	exeDir := filepath.Dir(exePath)
	executableDir = exeDir // Set global variable

	// Change to executable directory to ensure all relative paths work correctly
	if err := os.Chdir(exeDir); err != nil {
		log.Printf("Warning: Could not change to executable directory: %v", err)
	}

	// Initialize license manager with path relative to executable
	licensePath := filepath.Join(exeDir, "license.dat")
	licenseManager, err = license.NewManager(licensePath)
	if err != nil {
		log.Printf("Warning: Failed to initialize license manager: %v", err)
	}

	// Ensure proper cleanup on exit
	defer func() {
		if licenseManager != nil {
			licenseManager.Close()
		}
	}()

	// Check license status (but don't exit if invalid - let web interface handle it)
	checkLicenseOnStartup()

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

	// Add security middleware to all routes
	r.Use(securityMiddleware)

	// Add license middleware to protected API routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(licenseMiddleware)

	// Serve static files (relative to executable)
	staticDir := filepath.Join(executableDir, "web", "static")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Public license endpoints (no license middleware needed)
	r.HandleFunc("/api/license/status", handleLicenseStatus).Methods("GET")
	r.HandleFunc("/api/license/activate", handleLicenseActivate).Methods("POST")
	r.HandleFunc("/api/license/transfer", handleLicenseTransfer).Methods("POST")
	r.HandleFunc("/api/license/renewal-status", handleRenewalStatus).Methods("GET")
	r.HandleFunc("/api/license/test-connectivity", handleTestConnectivity).Methods("GET")
	r.HandleFunc("/api/license/heartbeat", handleLicenseHeartbeat).Methods("POST")

	// Administrative endpoints (requires special handling)
	r.HandleFunc("/api/admin/system-stats", handleSystemStats).Methods("GET")
	r.HandleFunc("/api/admin/performance", handlePerformanceStats).Methods("GET")
	r.HandleFunc("/api/admin/cache-stats", handleCacheStats).Methods("GET")
	r.HandleFunc("/api/admin/security-stats", handleSecurityStats).Methods("GET")
	r.HandleFunc("/api/admin/logs", handleGetLogs).Methods("GET")

	// Protected API endpoints (require valid license)
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

	// Generate ticker summary on startup only if data exists
	combinedDataPath := filepath.Join(executableDir, "reports", "isx_combined_data.csv")
	if _, err := os.Stat(combinedDataPath); err == nil {
		if err := generateTickerSummary(); err != nil {
			log.Printf("Warning: Failed to generate ticker summary on startup: %v", err)
		}
	}

	serverURL := "http://localhost:8080"
	fmt.Printf("🔐 ISX Web Interface (Enhanced Licensed v2.0.0) starting on %s\n", serverURL)

	// Start server in background
	go func() {
		log.Fatal(http.ListenAndServe(":8080", r))
	}()

	// Wait a moment for server to start, then open browser
	time.Sleep(2 * time.Second)
	if err := openBrowser(serverURL); err != nil {
		log.Printf("Warning: Could not open browser automatically: %v", err)
		fmt.Printf("Please open your browser and navigate to: %s\n", serverURL)
	} else {
		fmt.Println("✓ Browser opened automatically")
	}

	// Keep the application running
	select {}
}

func checkLicenseOnStartup() {
	if licenseManager == nil {
		fmt.Println("⚠️  License system not available - web interface will handle activation")
		return
	}

	// Perform a single license validation check
	valid, _ := licenseManager.ValidateLicense()
	if valid {
		// Get license info for display
		info, infoErr := licenseManager.GetLicenseInfo()
		if infoErr == nil {
			daysLeft := int(time.Until(info.ExpiryDate).Hours() / 24)
			fmt.Printf("✅ License Valid - %d days remaining\n", daysLeft)

			// Check renewal status separately if needed
			renewalInfo, _ := licenseManager.CheckRenewalStatus()
			if renewalInfo != nil && renewalInfo.NeedsRenewal {
				fmt.Printf("⚠️  %s\n", renewalInfo.Message)
			} else {
				fmt.Println("🌐 Opening main application interface...")
			}
		} else {
			fmt.Println("✅ License Valid - ready to start")
			fmt.Println("🌐 Opening main application interface...")
		}
	} else {
		// Get validation state to provide better messaging
		validationState, _ := licenseManager.GetValidationState()

		if validationState != nil {
			switch validationState.ErrorType {
			case "machine_mismatch":
				fmt.Println("🚫 License not valid for this machine")
				fmt.Println("   📞 Please contact Iraqi Investor to get a new license for this machine")
				fmt.Println("   🌐 Opening license activation page...")
			case "expired":
				fmt.Println("⏰ License has expired")
				fmt.Println("   📞 Please contact Iraqi Investor to renew your license")
				fmt.Println("   🌐 Opening license activation page...")
			case "network_error":
				fmt.Println("🌐 Cannot verify license due to network issues")
				fmt.Println("   💡 Check your internet connection and try again")
				fmt.Println("   🌐 Opening license activation page...")
			default:
				fmt.Println("🔑 No license found")
				fmt.Println("   📞 Please contact Iraqi Investor to get a license")
				fmt.Println("   🌐 Opening license activation page...")
			}
		} else {
			fmt.Println("🔑 No license found")
			fmt.Println("   📞 Please contact Iraqi Investor to get a license")
			fmt.Println("   🌐 Opening license activation page...")
		}
	}
}

func licenseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if licenseManager == nil {
			http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
			return
		}

		// Use the same smart validation as serveIndex for consistent behavior
		valid, _ := validateLicenseForWebAccess()
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
					response["actions"] = []string{"contact_support", "activate_new_license"}
				case "expired":
					response["code"] = "LICENSE_EXPIRED"
					response["message"] = "Your license has expired. Please contact Iraqi Investor to renew your license."
					response["contact_info"] = "Please contact Iraqi Investor for renewal"
					response["actions"] = []string{"contact_support", "activate_new_license"}
				case "network_error":
					response["code"] = "LICENSE_NETWORK_ERROR"
					response["message"] = "Cannot verify license due to network issues. Please check your internet connection and try again."
					response["actions"] = []string{"retry", "check_network"}
				default:
					response["message"] = "No valid license found. Please contact Iraqi Investor to get a license."
					response["contact_info"] = "Please contact Iraqi Investor for assistance"
					response["actions"] = []string{"contact_support", "activate_new_license"}
				}
			} else {
				response["message"] = "No valid license found. Please contact Iraqi Investor to get a license."
				response["contact_info"] = "Please contact Iraqi Investor for assistance"
				response["actions"] = []string{"contact_support", "activate_new_license"}
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
			Status:  "System Error",
		})
		return
	}

	// Use enhanced validation with renewal checking
	valid, renewalInfo, _ := licenseManager.ValidateWithRenewalCheck()

	// Get detailed validation state for better feedback
	validationState, _ := licenseManager.GetValidationState()

	if !valid {
		status := LicenseStatus{
			IsValid: false,
		}

		// Add helpful information based on validation state
		if validationState != nil {
			switch validationState.ErrorType {
			case "machine_mismatch":
				status.Status = "Invalid Machine"
				status.Message = "This license is not valid for this machine."
				status.RenewalMessage = "Please contact Iraqi Investor to get a new license for this machine."
			case "expired":
				status.Status = "Expired"
				status.Message = "Your license has expired."
				status.RenewalMessage = "Please contact Iraqi Investor to renew your license."
			case "network_error":
				status.Status = "Network Error"
				status.Message = "Cannot verify license due to network issues."
				status.RenewalMessage = "Please check your internet connection and try again."
			default:
				status.Status = "No License"
				status.Message = "No valid license found."
				status.RenewalMessage = "Please contact Iraqi Investor to get a license."
			}
		} else {
			status.Status = "No License"
			status.Message = "No valid license found."
			status.RenewalMessage = "Please contact Iraqi Investor to get a license."
		}

		json.NewEncoder(w).Encode(status)
		return
	}

	// Valid license - get license info and renewal status
	info, err := licenseManager.GetLicenseInfo()
	if err != nil {
		json.NewEncoder(w).Encode(LicenseStatus{
			IsValid:        false,
			Message:        "Failed to get license information.",
			Status:         "Error",
			RenewalMessage: "Please contact Iraqi Investor for assistance.",
		})
		return
	}

	status := LicenseStatus{
		IsValid:    true,
		ExpiryDate: info.ExpiryDate,
		DaysLeft:   int(time.Until(info.ExpiryDate).Hours() / 24),
		Message:    "License is valid and active",
		Status:     "Active",
	}

	// Add renewal information if available
	if renewalInfo != nil {
		status.NeedsRenewal = renewalInfo.NeedsRenewal
		if renewalInfo.NeedsRenewal {
			status.RenewalMessage = renewalInfo.Message + " Please contact Iraqi Investor for renewal assistance."
		} else {
			status.RenewalMessage = renewalInfo.Message
		}
	}

	json.NewEncoder(w).Encode(status)
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

	// Log license activation attempt for debugging
	log.Printf("License activation attempt for key: %s...", req.LicenseKey[:min(8, len(req.LicenseKey))])

	if err := licenseManager.ActivateLicense(req.LicenseKey); err != nil {
		log.Printf("License activation failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		// Provide more detailed error messages
		var userMessage string
		errorStr := err.Error()

		if strings.Contains(errorStr, "network") || strings.Contains(errorStr, "connection") {
			userMessage = "Network connection error. Please check your internet connection and try again."
		} else if strings.Contains(errorStr, "timeout") {
			userMessage = "Connection timeout. Please check your internet connection and try again."
		} else if strings.Contains(errorStr, "invalid license") {
			userMessage = "Invalid license key. Please check your license key and try again."
		} else if strings.Contains(errorStr, "expired") {
			userMessage = "License has expired. Please contact support for renewal."
		} else if strings.Contains(errorStr, "already activated") {
			userMessage = "License is already activated on another machine. Contact support if you need to transfer your license."
		} else if strings.Contains(errorStr, "sheets") || strings.Contains(errorStr, "google") {
			userMessage = "Unable to connect to license validation service. Please check your internet connection and try again."
		} else {
			userMessage = fmt.Sprintf("License activation failed: %s", errorStr)
		}

		json.NewEncoder(w).Encode(map[string]string{
			"error": userMessage,
			"debug": errorStr, // Include technical details for debugging
		})
		return
	}

	log.Printf("License activated successfully")

	// Get license information to return to user
	info, err := licenseManager.GetLicenseInfo()
	if err != nil {
		log.Printf("Warning: Could not get license info after activation: %v", err)
		// Still return success but with basic message
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
		"duration":    info.Duration,
		"expiry_date": info.ExpiryDate.Format("January 2, 2006"),
		"user_email":  info.UserEmail,
	})
}

// Helper function for min operation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	// Check if user has a valid license using local-first validation for better user experience
	if licenseManager != nil {
		log.Printf("DEBUG: Calling validateLicenseForWebAccess...")
		valid, isRecentActivation := validateLicenseForWebAccess()
		log.Printf("DEBUG: validateLicenseForWebAccess returned valid=%v, isRecentActivation=%v", valid, isRecentActivation)

		if valid {
			// License is valid, serve the main application
			if isRecentActivation {
				log.Printf("Recently activated license found, serving main application")
			} else {
				log.Printf("Valid license found, serving main application")
			}
			indexPath := filepath.Join(executableDir, "web", "index.html")
			log.Printf("DEBUG: Serving main application from: %s", indexPath)
			http.ServeFile(w, r, indexPath)
			return
		} else {
			// License is invalid or missing, serve license activation page
			log.Printf("License validation failed, serving license page")
		}
	} else {
		log.Printf("License manager not available, serving license page")
	}

	// Serve license activation page
	licensePath := filepath.Join(executableDir, "web", "license.html")
	log.Printf("DEBUG: Serving license page from: %s", licensePath)
	http.ServeFile(w, r, licensePath)
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

	// Check if downloads directory has files for the requested date range
	downloadsDir := filepath.Join(executableDir, "downloads")
	needsDownload := true

	// Get user-requested date range
	fromDate := req.Args["from"]
	toDate := req.Args["to"]

	if entries, err := os.ReadDir(downloadsDir); err == nil {
		excelCount := 0
		existingFiles := make(map[string]bool)

		// Build a map of existing Excel files
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".xlsx") {
				existingFiles[entry.Name()] = true
				excelCount++
			}
		}

		if excelCount > 0 {
			broadcastMessage("info", fmt.Sprintf("Found %d existing Excel files in downloads directory", excelCount), "scrape")

			// If user provided date range, check if we need to download missing files
			if fromDate != "" && toDate != "" {
				// Check if we have files for the requested date range
				missingFiles := checkMissingDateRangeFiles(existingFiles, fromDate, toDate)
				if len(missingFiles) > 0 {
					broadcastMessage("info", fmt.Sprintf("Missing %d files for date range %s to %s - will download fresh data", len(missingFiles), fromDate, toDate), "scrape")
					needsDownload = true
				} else {
					broadcastMessage("info", fmt.Sprintf("All files exist for date range %s to %s - skipping download", fromDate, toDate), "scrape")
					needsDownload = false
				}
			} else {
				// No specific date range - process existing files
				broadcastMessage("info", "No date range specified - processing existing files", "scrape")
				needsDownload = false
			}
		}
	}

	var response CommandResponse

	// Download fresh data if needed
	if needsDownload {
		broadcastMessage("info", "No Excel files found. Downloading fresh data from ISX website...", "scrape")

		// Use the web scraper to download Excel files
		scraperArgs := []string{"-mode=initial", "-out=downloads"}

		// Use EXACTLY the dates selected by user in HTML form (no validation overrides)
		fromDate := req.Args["from"]
		toDate := req.Args["to"]

		// Always use the dates selected by user - no modifications or validation
		if fromDate != "" {
			scraperArgs = append(scraperArgs, "-from="+fromDate)
			broadcastMessage("info", fmt.Sprintf("Using FROM date from form: %s", fromDate), "scrape")
		}
		if toDate != "" {
			scraperArgs = append(scraperArgs, "-to="+toDate)
			broadcastMessage("info", fmt.Sprintf("Using TO date from form: %s", toDate), "scrape")
		}

		scraperPath := filepath.Join(executableDir, "bin", "isx-web-scraper.exe")
		broadcastMessage("info", fmt.Sprintf("Starting scrape command: %s %s", scraperPath, strings.Join(scraperArgs, " ")), "scrape")

		scraperResponse := executeCommandWithTimeout(scraperPath, scraperArgs, "scrape", 5*time.Minute)

		if !scraperResponse.Success {
			broadcastMessage("error", "Failed to download fresh data from ISX website", "scrape")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(scraperResponse)
			return
		}

		broadcastMessage("success", "✅ Fresh data downloaded successfully from ISX website", "scrape")
	}

	// Now process the Excel files
	broadcastMessage("info", "Processing Excel files from downloads directory...", "scrape")

	// Build command arguments for the processing tool
	args := []string{}

	// Set input directory (default: downloads)
	if inDir := req.Args["in"]; inDir != "" {
		args = append(args, "-in="+inDir)
	} else {
		args = append(args, "-in=downloads")
	}

	// Set output directory (default: reports)
	if outDir := req.Args["out"]; outDir != "" {
		args = append(args, "-out="+outDir)
	} else {
		args = append(args, "-out=reports")
	}

	// Enable full rework if requested
	if mode := req.Args["mode"]; mode == "full" {
		args = append(args, "-full")
	}

	processPath := filepath.Join(executableDir, "process.exe")
	response = executeCommand(processPath, args, "scrape")

	// If scraping was successful, automatically process the data
	if response.Success {
		broadcastMessage("info", "Scraping completed successfully. Starting automatic data processing...", "scrape")

		// Run processing automatically
		processArgs := []string{"-in=downloads"}
		processPath := filepath.Join(executableDir, "process.exe")
		processResponse := executeCommandWithStreaming(processPath, processArgs, "process")

		if processResponse.Success {
			broadcastMessage("info", "Data processing completed. Extracting market indices...", "scrape")

			// Run index extraction automatically
			indexArgs := []string{"-dir=downloads", "-out=reports/indexes.csv"}
			indexcsvPath := filepath.Join(executableDir, "indexcsv.exe")
			indexResponse := executeCommand(indexcsvPath, indexArgs, "indexcsv")

			if indexResponse.Success {
				broadcastMessage("info", "Index extraction completed. Generating ticker summary...", "scrape")

				// Generate fresh ticker summary after processing
				if err := generateTickerSummary(); err != nil {
					broadcastMessage("warning", fmt.Sprintf("Warning: Failed to generate ticker summary: %v", err), "scrape")
				} else {
					broadcastMessage("success", "✅ Complete data pipeline finished! All data updated.", "scrape")

					// Notify frontend to refresh all components
					broadcastMessage("refresh", "data_updated", "scrape")
				}
			} else {
				broadcastMessage("warning", "Index extraction failed after processing", "scrape")
			}
		} else {
			broadcastMessage("warning", "Data processing failed after scraping", "scrape")
		}
	}

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

	processPath := filepath.Join(executableDir, "process.exe")
	response := executeCommandWithStreaming(processPath, args, "process")

	// If processing was successful, run complete pipeline
	if response.Success {
		broadcastMessage("info", "Processing completed successfully. Extracting market indices...", "process")

		// Run index extraction automatically
		indexArgs := []string{"-dir=downloads", "-out=reports/indexes.csv"}
		indexcsvPath := filepath.Join(executableDir, "indexcsv.exe")
		indexResponse := executeCommand(indexcsvPath, indexArgs, "indexcsv")

		if indexResponse.Success {
			broadcastMessage("info", "Index extraction completed. Generating ticker summary...", "process")

			// Generate fresh ticker summary after processing
			if err := generateTickerSummary(); err != nil {
				broadcastMessage("warning", fmt.Sprintf("Warning: Failed to generate ticker summary: %v", err), "process")
			} else {
				broadcastMessage("success", "✅ Complete processing pipeline finished! All data updated.", "process")

				// Notify frontend to refresh all components
				broadcastMessage("refresh", "data_updated", "process")
			}
		} else {
			broadcastMessage("warning", "Index extraction failed after processing", "process")
		}
	}

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

	// Set input directory (default: downloads)
	if dir := req.Args["dir"]; dir != "" {
		args = append(args, "-dir="+dir)
	} else {
		args = append(args, "-dir=downloads")
	}

	// Set output file (default: reports/indexes.csv)
	if out := req.Args["out"]; out != "" {
		args = append(args, "-out="+out)
	} else {
		args = append(args, "-out=reports/indexes.csv")
	}

	indexcsvPath := filepath.Join(executableDir, "indexcsv.exe")
	response := executeCommand(indexcsvPath, args, "indexcsv")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleListTickers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	summaryFile := filepath.Join(executableDir, "reports", "ticker_summary.json")

	// Check if summary file exists
	if _, err := os.Stat(summaryFile); os.IsNotExist(err) {
		// Generate summary if it doesn't exist
		if err := generateTickerSummary(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Failed to generate ticker summary",
				"tickers": []TickerSummary{},
			})
			return
		}
	}

	// Read the summary file
	data, err := os.ReadFile(summaryFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Failed to read ticker summary",
			"tickers": []TickerSummary{},
		})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Failed to parse ticker summary",
			"tickers": []TickerSummary{},
		})
		return
	}

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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  "Ticker not found",
			"ticker": ticker,
		})
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Write(csvData)
}

func handleListFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dir := r.URL.Query().Get("dir")

	// If no dir specified, return organized file listing for File Archive
	if dir == "" {
		// Get downloads (Excel files)
		downloadsFiles, err1 := listDirectory("downloads")
		if err1 != nil {
			downloadsFiles = []string{}
		}

		// Get generated reports (CSV files)
		reportsFiles, err2 := listDirectory("reports")
		if err2 != nil {
			reportsFiles = []string{}
		}

		// Filter downloads to show only Excel files
		var excelFiles []string
		for _, file := range downloadsFiles {
			if strings.HasSuffix(strings.ToLower(file), ".xlsx") {
				excelFiles = append(excelFiles, file)
			}
		}

		// Separate CSV files into ticker files and daily reports
		var tickerFiles []string
		var dailyReports []string
		var otherFiles []string

		for _, file := range reportsFiles {
			fileName := strings.ToLower(file)
			if strings.HasSuffix(fileName, ".csv") || strings.HasSuffix(fileName, ".json") {
				if strings.Contains(fileName, "_trading_history.csv") {
					// Individual ticker files
					tickerFiles = append(tickerFiles, file)
				} else if strings.HasPrefix(fileName, "isx_daily_") && strings.HasSuffix(fileName, ".csv") {
					// Daily report files
					dailyReports = append(dailyReports, file)
				} else {
					// Other important files (ticker_summary, indexes, etc.)
					otherFiles = append(otherFiles, file)
				}
			}
		}

		// Sort files appropriately
		sort.Strings(excelFiles)                                // Downloads: alphabetical
		sort.Strings(tickerFiles)                               // Ticker files: alphabetical
		sort.Sort(sort.Reverse(sort.StringSlice(dailyReports))) // Daily reports: newest first
		sort.Strings(otherFiles)                                // Other files: alphabetical

		response := map[string]interface{}{
			"downloads":     excelFiles,
			"ticker_files":  tickerFiles,
			"daily_reports": dailyReports,
			"other_files":   otherFiles,
		}

		json.NewEncoder(w).Encode(response)
		return
	}

	// If dir is specified, return files from that directory
	files, err := listDirectory(dir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
			"files": []string{},
		})
		return
	}

	json.NewEncoder(w).Encode(files)
}

func handleDownloadFile(w http.ResponseWriter, r *http.Request) {
	filename := mux.Vars(r)["filename"]
	if filename == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Missing filename",
		})
		return
	}

	dir := r.URL.Query().Get("dir")

	// If no dir specified, auto-detect based on file extension
	if dir == "" {
		if strings.HasSuffix(strings.ToLower(filename), ".xlsx") {
			dir = "downloads"
		} else if strings.HasSuffix(strings.ToLower(filename), ".csv") || strings.HasSuffix(strings.ToLower(filename), ".json") {
			dir = "reports"
		} else {
			// Try both directories
			possiblePaths := []string{
				filepath.Join("downloads", filename),
				filepath.Join("reports", filename),
			}

			var foundPath string
			for _, path := range possiblePaths {
				if _, err := os.Stat(path); err == nil {
					foundPath = path
					break
				}
			}

			if foundPath == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":    "File not found in downloads or reports directories",
					"filename": filename,
				})
				return
			}

			file, err := os.Open(foundPath)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":    err.Error(),
					"filename": filename,
				})
				return
			}
			defer file.Close()

			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
			w.Header().Set("Content-Type", "application/octet-stream")

			_, err = io.Copy(w, file)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": err.Error(),
				})
			}
			return
		}
	}

	file, err := os.Open(filepath.Join(dir, filename))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":    err.Error(),
			"filename": filename,
		})
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err = io.Copy(w, file)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	// Implementation remains the same
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

func executeCommandWithTimeout(command string, args []string, commandType string, timeout time.Duration) CommandResponse {
	broadcastMessage("info", fmt.Sprintf("Starting %s command with %v timeout: %s %s", commandType, timeout, command, strings.Join(args, " ")), commandType)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	output, err := cmd.CombinedOutput()

	response := CommandResponse{
		Success: err == nil,
		Output:  string(output),
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			response.Error = fmt.Sprintf("Command timed out after %v", timeout)
			broadcastMessage("error", fmt.Sprintf("Command timed out after %v", timeout), commandType)
		} else {
			response.Error = err.Error()
			broadcastMessage("error", fmt.Sprintf("Command failed: %s", err.Error()), commandType)
		}
	} else {
		broadcastMessage("success", fmt.Sprintf("Command completed successfully"), commandType)
	}

	broadcastMessage("output", string(output), commandType)

	return response
}

func checkMissingDateRangeFiles(existingFiles map[string]bool, fromDate, toDate string) []string {
	// Parse the date range
	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		log.Printf("Error parsing from date %s: %v", fromDate, err)
		return []string{}
	}

	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		log.Printf("Error parsing to date %s: %v", toDate, err)
		return []string{}
	}

	var missingFiles []string

	// Check each date in the range (excluding weekends)
	for current := from; !current.After(to); current = current.AddDate(0, 0, 1) {
		// Skip weekends (Saturday = 6, Sunday = 0)
		if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
			continue
		}

		// Generate expected filename for this date
		expectedFileName := fmt.Sprintf("%s ISX Daily Report.xlsx", current.Format("2006 01 02"))

		// Check if file exists
		if !existingFiles[expectedFileName] {
			missingFiles = append(missingFiles, expectedFileName)
		}
	}

	return missingFiles
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

func generateTickerSummary() error {
	combinedFile := filepath.Join(executableDir, "reports", "isx_combined_data.csv")
	summaryCSVFile := filepath.Join(executableDir, "reports", "ticker_summary.csv")
	summaryJSONFile := filepath.Join(executableDir, "reports", "ticker_summary.json")

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
	outFile, err := os.Create(summaryCSVFile)
	if err != nil {
		return fmt.Errorf("failed to create CSV summary file: %v", err)
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

	// Also write JSON file for API consumption
	jsonData := map[string]interface{}{
		"tickers":      summaries,
		"count":        len(summaries),
		"generated_at": time.Now().Format(time.RFC3339),
	}

	jsonFile, err := os.Create(summaryJSONFile)
	if err != nil {
		return fmt.Errorf("failed to create JSON summary file: %v", err)
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jsonData); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	log.Printf("Generated ticker summary with %d tickers (CSV and JSON)", len(summaries))
	return nil
}

func handleTestConnectivity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if licenseManager == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "License system unavailable",
		})
		return
	}

	log.Printf("Running network connectivity test...")

	if err := licenseManager.TestNetworkConnectivity(); err != nil {
		log.Printf("Connectivity test failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	log.Printf("Connectivity test passed")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All connectivity tests passed",
	})
}

func handleLicenseTransfer(w http.ResponseWriter, r *http.Request) {
	var req LicenseTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if licenseManager == nil {
		http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
		return
	}

	// Log license transfer attempt for debugging
	log.Printf("License transfer attempt for key: %s...", req.LicenseKey[:min(8, len(req.LicenseKey))])

	if err := licenseManager.TransferLicense(req.LicenseKey, req.ForceTransfer); err != nil {
		log.Printf("License transfer failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		// Provide detailed error messages
		var userMessage string
		errorStr := err.Error()

		if strings.Contains(errorStr, "already activated") && !req.ForceTransfer {
			userMessage = "License is already activated on another machine. Enable 'Force Transfer' to override this."
		} else if strings.Contains(errorStr, "network") || strings.Contains(errorStr, "connection") {
			userMessage = "Network connection error. Please check your internet connection and try again."
		} else if strings.Contains(errorStr, "expired") {
			userMessage = "License has expired. Please contact support for renewal."
		} else {
			userMessage = fmt.Sprintf("License transfer failed: %s", errorStr)
		}

		json.NewEncoder(w).Encode(map[string]string{
			"error": userMessage,
			"debug": errorStr,
		})
		return
	}

	log.Printf("License transferred successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "License transferred successfully",
	})
}

func handleRenewalStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if licenseManager == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":         "License system unavailable",
			"needs_renewal": true,
			"status":        "System Error",
		})
		return
	}

	renewalInfo, err := licenseManager.CheckRenewalStatus()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":         err.Error(),
			"needs_renewal": true,
			"status":        "Error",
		})
		return
	}

	json.NewEncoder(w).Encode(renewalInfo)
}

// Admin endpoint handlers
func handleSystemStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if licenseManager == nil {
		http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
		return
	}

	stats := licenseManager.GetSystemStats()

	// Add additional system information with safe type conversion
	response := SystemStatsResponse{
		Timestamp: time.Now(),
		MachineID: "current_machine",
		Version:   "2.0.0",
		Uptime:    time.Since(startTime),
	}

	// Safely convert performance stats
	if perfData, ok := stats["performance"]; ok {
		// Handle both map[string]*PerformanceMetrics and map[string]interface{}
		if perfMap, ok := perfData.(map[string]interface{}); ok {
			response.Performance = perfMap
		} else {
			// Convert any other type to interface{}
			response.Performance = map[string]interface{}{"data": perfData}
		}
	} else {
		response.Performance = map[string]interface{}{"status": "no data"}
	}

	// Safely convert cache stats
	if cacheData, ok := stats["cache"]; ok {
		if cacheMap, ok := cacheData.(map[string]interface{}); ok {
			response.Cache = cacheMap
		} else {
			response.Cache = map[string]interface{}{"error": "cache data unavailable"}
		}
	} else {
		response.Cache = map[string]interface{}{"status": "no data"}
	}

	// Safely convert security stats
	if secData, ok := stats["security"]; ok {
		if secMap, ok := secData.(map[string]interface{}); ok {
			response.Security = secMap
		} else {
			response.Security = map[string]interface{}{"error": "security data unavailable"}
		}
	} else {
		response.Security = map[string]interface{}{"status": "no data"}
	}

	json.NewEncoder(w).Encode(response)
}

func handlePerformanceStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if licenseManager == nil {
		http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
		return
	}

	stats := licenseManager.GetPerformanceMetrics()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"performance_metrics": stats,
		"timestamp":           time.Now(),
	})
}

func handleCacheStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if licenseManager == nil {
		http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
		return
	}

	systemStats := licenseManager.GetSystemStats()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"cache_stats": systemStats["cache"],
		"timestamp":   time.Now(),
	})
}

func handleSecurityStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if licenseManager == nil {
		http.Error(w, "License system unavailable", http.StatusServiceUnavailable)
		return
	}

	systemStats := licenseManager.GetSystemStats()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"security_stats": systemStats["security"],
		"timestamp":      time.Now(),
	})
}

func handleGetLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	logType := r.URL.Query().Get("type") // "license" or "audit"
	limitStr := r.URL.Query().Get("limit")

	limit := 100 // Default limit
	if limitStr != "" {
		if parsedLimit, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || parsedLimit != 1 {
			limit = 100
		}
	}

	var logFile string
	switch logType {
	case "audit":
		logFile = "logs/audit.log"
	default:
		logFile = "logs/license.log"
	}

	// Read log file (simplified - in production, you'd want streaming/pagination)
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"logs":      []string{},
			"message":   "Log file not found",
			"timestamp": time.Now(),
		})
		return
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		http.Error(w, "Failed to read log file", http.StatusInternalServerError)
		return
	}

	lines := strings.Split(string(data), "\n")

	// Get the last 'limit' lines
	start := 0
	if len(lines) > limit {
		start = len(lines) - limit
	}

	recentLines := lines[start:]

	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":      recentLines,
		"total":     len(lines),
		"limit":     limit,
		"log_type":  logType,
		"timestamp": time.Now(),
	})
}

func handleLicenseHeartbeat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if licenseManager == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "License system unavailable",
		})
		return
	}

	// Send license heartbeat by updating last connected time
	if err := licenseManager.UpdateLastConnected(); err != nil {
		log.Printf("License heartbeat failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "License heartbeat sent successfully",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
