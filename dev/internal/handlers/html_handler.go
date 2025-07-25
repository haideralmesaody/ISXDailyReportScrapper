package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ServeLicensePage serves the license activation page
func ServeLicensePage(webDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		licensePath := filepath.Join(webDir, "license.html")
		
		// Check if license file exists
		if _, err := os.Stat(licensePath); os.IsNotExist(err) {
			http.Error(w, "License page not found", http.StatusNotFound)
			return
		}

		serveHTML(w, r, licensePath)
	}
}

// ServeMainApp serves the main application page
func ServeMainApp(webDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexPath := filepath.Join(webDir, "index.html")
		
		// Check if index file exists
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			http.Error(w, "Main application page not found", http.StatusNotFound)
			return
		}

		serveHTML(w, r, indexPath)
	}
}

// RedirectToLicense redirects root requests to license page
func RedirectToLicense(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/license", http.StatusTemporaryRedirect)
}

// ServeTestPage serves a simple test page for debugging
func ServeTestPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>ISX Daily Reports Scrapper - Test Page</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .status { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .success { background-color: #d4edda; color: #155724; }
        .info { background-color: #d1ecf1; color: #0c5460; }
        .warning { background-color: #fff3cd; color: #856404; }
        .error { background-color: #f8d7da; color: #721c24; }
    </style>
</head>
<body>
    <h1>ISX Daily Reports Scrapper - Test Page</h1>
    <div class="status info">
        <strong>Status:</strong> Application is running
        <br><strong>Time:</strong> %s
    </div>
    <h2>Quick Links</h2>
    <ul>
        <li><a href="/license">License Activation</a></li>
        <li><a href="/app">Main Application</a></li>
        <li><a href="/api/health">Health Check</a></li>
        <li><a href="/api/version">Version Info</a></li>
        <li><a href="/ws">WebSocket Endpoint</a></li>
    </ul>
</body>
</html>
		`, time.Now().Format("2006-01-02 15:04:05"))
	}
}

// serveHTML serves an HTML file with proper headers
func serveHTML(w http.ResponseWriter, r *http.Request, filePath string) {
	// Set security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Parse and execute template
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}