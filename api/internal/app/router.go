package app

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	apperrors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/middleware"
	handlers "github.com/isxcli/isxcli/internal/transport/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log/slog"
)

// setupRouter configures the HTTP router with all routes
func (a *Application) setupRouter() {
	r := chi.NewRouter()

	// Apply MINIMAL middleware that won't interfere with WebSocket
	// These are safe because they don't wrap the ResponseWriter
	r.Use(middleware.RequestID) // Use our CLAUDE.md compliant RequestID
	r.Use(middleware.RealIP)

	// Add a debugging middleware to log all requests
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log WebSocket requests with high detail
			if r.URL.Path == "/ws" {
				reqID := r.Header.Get("X-Request-ID")
				if reqID == "" {
					reqID = fmt.Sprintf("debug-%d", time.Now().UnixNano())
				}
				ctx := infrastructure.WithTraceID(r.Context(), reqID)
				a.Logger.InfoContext(ctx, "🔍 DEBUG: WebSocket request received",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("query", r.URL.RawQuery),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("host", r.Host),
					slog.String("origin", r.Header.Get("Origin")),
					slog.String("user_agent", r.UserAgent()),
					slog.String("connection_header", r.Header.Get("Connection")),
					slog.String("upgrade_header", r.Header.Get("Upgrade")),
					slog.String("accept_header", r.Header.Get("Accept")))
			}
			next.ServeHTTP(w, r)
		})
	})

	// WebSocket route with NO middleware to prevent upgrade interference
	// CRITICAL: WebSocket upgrades fail when middleware wraps ResponseWriter
	routeStart := time.Now()
	r.HandleFunc("/ws", a.handleWebSocket)
	routeTime := time.Since(routeStart)

	a.Logger.Info("🔧 [WS-SETUP] WebSocket route registered",
		slog.Duration("route_registration_time", routeTime),
		slog.String("websocket_route", "/ws"),
		slog.Bool("middleware_bypassed", true),
		slog.String("route_handler", "handleWebSocket"),
		slog.Time("route_registered_at", time.Now()))

	// Serve static assets OUTSIDE middleware group to avoid license validation
	if a.FrontendFS != nil {
		a.setupStaticAssetsOnly(r)
	}

	// Create a route group for everything else with FULL middleware
	r.Group(func(r chi.Router) {
		// Apply remaining middleware only to this group
		// Follow CLAUDE.md ordering: RequestID → RealIP → OTel → Logger → Recoverer → Timeout

		// TODO: Add OpenTelemetry middleware when available
		// Business metrics can be added later when needed

		r.Use(middleware.StructuredLogger(a.Logger)) // Use infrastructure logger
		r.Use(middleware.Recoverer(a.Logger))        // Use our CLAUDE.md compliant recoverer
		// NOTE: Timeout middleware moved to specific route groups below to allow different timeouts for operations
		r.Use(middleware.SecurityHeaders)

		// CORS middleware - configured for embedded frontend and development
		corsConfig := a.getCORSConfig()
		r.Use(middleware.CORS(corsConfig))

		// Rate limiting
		if a.Config.Security.RateLimit.Enabled {
			r.Use(middleware.NewRateLimiter(
				a.Config.Security.RateLimit.RPS,
				a.Config.Security.RateLimit.Burst,
				a.Logger, // Pass infrastructure logger
			).Handler)
		}

		// License validation
		validationSvc := license.NewValidationService(a.LicenseManager)
		licenseValidator := middleware.NewLicenseValidatorWithValidation(validationSvc.Validate, a.Logger)
		r.Use(licenseValidator.Handler)

		// Now register all other routes within this group
		a.setupAPIRoutes(r)
		a.setupHTMLRoutes(r)
	})

	// Add Prometheus metrics endpoint (outside the middleware group for performance)
	if a.OTelProviders.PrometheusHTTP != nil {
		r.Handle("/metrics", a.OTelProviders.PrometheusHTTP)
	}

	a.Router = r
}

// setupAPIRoutes configures API endpoints
func (a *Application) setupAPIRoutes(r chi.Router) {
	// API routes with common middleware
	r.Route("/api", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))

		// Apply standard timeout to most API endpoints
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(a.Config.Server.ReadTimeout, a.Logger))

			// Health handler with expanded responsibilities (includes diagnostics from operations handler)
			healthHandler := handlers.NewHealthHandler(a.HealthService, a.OperationService.GetManager(), a.JobQueue, a.Logger, apperrors.NewErrorHandler(a.Logger, false))
			r.Mount("/", healthHandler.Routes())

			// Metrics and observability handler
			metricsHandler, err := handlers.NewOperationsMetricsHandler(a.OperationService.GetManager(), a.JobQueue, a.Logger)
			if err != nil {
				a.Logger.Error("failed to create operations metrics handler",
					slog.String("error", err.Error()))
				return
			}
			r.Mount("/metrics", metricsHandler.Routes())

			// License endpoints
			licenseHandler := handlers.NewLicenseHandler(a.Services.LicenseService, a.Logger)
			r.Mount("/license", licenseHandler.Routes())

			// Create error handler
			errorHandler := apperrors.NewErrorHandler(a.Logger, false)

			// Data handler
			dataHandler := handlers.NewDataHandler(a.DataService, a.Logger, errorHandler)
			r.Mount("/data", dataHandler.Routes())

			// Operations handler (minimal single-stage endpoints)
			if a.OperationService != nil {
				opsHandler := handlers.NewOperationsHandler(a.OperationService, a.Logger, errorHandler)
				r.Mount("/operations", opsHandler.Routes())
				r.Mount("/stages", opsHandler.StageRoutes())
			} else {
				a.Logger.Warn("OperationService is nil - skipping /api/operations routes")
			}

			// Liquidity handler
			liquidityHandler := handlers.NewLiquidityHandler(a.Services.Liquidity, a.Logger)
			liquidityHandler.RegisterRoutes(r)

			// Strategy handler
			strategyHandler := handlers.NewStrategyHandler(a.Services.Strategy, a.Logger)
			strategyHandler.RegisterRoutes(r)

	
			
		})

		// Client logging endpoint with standard timeout
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(a.Config.Server.ReadTimeout, a.Logger))
			r.Post("/logs", handlers.NewClientLogHandler(a.Logger).Handle)
		})
	})
}

// setupHTMLRoutes configures HTML page routes and embedded Next.js frontend
func (a *Application) setupHTMLRoutes(r chi.Router) {
	// Serve embedded Next.js frontend
	a.setupEmbeddedFrontend(r)

	// Legacy routes for backward compatibility (served from filesystem if available)
	r.Get("/legacy/license", handlers.ServeLicensePage(a.Config.GetWebDir()))
	r.Get("/legacy/app", handlers.ServeMainApp(a.Config.GetWebDir()))
	r.Get("/test", handlers.ServeTestPage())
}

// setupStaticAssetsOnly configures ONLY static assets without middleware
func (a *Application) setupStaticAssetsOnly(r chi.Router) {
	frontendFS := a.FrontendFS

	// Serve static assets (JS, CSS, images, etc.) with cache-busting headers
	r.Route("/_next", func(r chi.Router) {
		// Use no-cache for development, aggressive caching for production with versioned files
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if file has hash in name (cache-busted)
				if strings.Contains(r.URL.Path, ".") && (strings.Contains(r.URL.Path, "-") || strings.Contains(r.URL.Path, "_")) {
					// Versioned assets can be cached aggressively
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					// Non-versioned assets should not be cached
					w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
					w.Header().Set("Pragma", "no-cache")
					w.Header().Set("Expires", "0")
				}
				next.ServeHTTP(w, r)
			})
		})
		r.HandleFunc("/*", a.serveStaticWithMIME(frontendFS, "/_next").ServeHTTP)
	})

	// Serve other frontend static assets (from public folder)
	r.Route("/assets", func(r chi.Router) {
		// TODO: Add cache control header middleware when needed
		r.HandleFunc("/*", a.serveStaticWithMIME(frontendFS, "/assets").ServeHTTP)
	})

	// Serve favicon and other root assets
	r.Get("/favicon.ico", a.serveFrontendFile(frontendFS, "favicon.ico"))
	r.Get("/favicon-16x16.png", a.serveFrontendFile(frontendFS, "favicon-16x16.png"))
	r.Get("/favicon-32x32.png", a.serveFrontendFile(frontendFS, "favicon-32x32.png"))
	r.Get("/apple-touch-icon.png", a.serveFrontendFile(frontendFS, "apple-touch-icon.png"))
	r.Get("/android-chrome-192x192.png", a.serveFrontendFile(frontendFS, "android-chrome-192x192.png"))
	r.Get("/android-chrome-512x512.png", a.serveFrontendFile(frontendFS, "android-chrome-512x512.png"))
	r.Get("/site.webmanifest", a.serveFrontendFile(frontendFS, "site.webmanifest"))
	r.Get("/robots.txt", a.serveFrontendFile(frontendFS, "robots.txt"))
	r.Get("/iraqi-investor-logo.svg", a.serveFrontendFile(frontendFS, "iraqi-investor-logo.svg"))
	r.Get("/index.txt", a.serveFrontendFile(frontendFS, "index.txt"))

	// Serve highcharts icons (if any)
	r.Route("/highcharts-icons", func(r chi.Router) {
		// TODO: Add cache control header middleware when needed
		r.HandleFunc("/*", a.serveStaticWithMIME(frontendFS, "/highcharts-icons").ServeHTTP)
	})
}

// setupEmbeddedFrontend configures the embedded Next.js frontend serving (SPA routes only)
func (a *Application) setupEmbeddedFrontend(r chi.Router) {
	// Check if frontend filesystem is available
	if a.FrontendFS == nil {
		a.Logger.Warn("Frontend filesystem not available, falling back to legacy handlers")
		r.Get("/", handlers.RedirectToLicense)
		return
	}

	frontendFS := a.FrontendFS

	// NOTE: Static assets are served by setupStaticAssetsOnly outside middleware group

	// SPA routing - serve index.html for frontend routes (non-API)
	// Use specific patterns to avoid catching API routes
	registerSPARoute := func(route string, includeWildcard bool) {
		handler := a.serveSPAHandler(frontendFS)
		r.Get(route, handler)
		if includeWildcard {
			if !strings.HasSuffix(route, "/*") && route != "/" {
				wildcard := strings.TrimSuffix(route, "/") + "/*"
				r.Get(wildcard, handler)
			}
		}
	}

	registerSPARoute("/", false)
	registerSPARoute("/license", true)
	registerSPARoute("/market-overview", true)
	registerSPARoute("/operations", true)
	registerSPARoute("/liquidity", true)
	registerSPARoute("/strategy", true)
	registerSPARoute("/reports", true)
	registerSPARoute("/guide", true)

	// Handle other frontend routes with a more specific pattern
	// Exclude /api, /_next, /health, /metrics, etc.
	r.Get("/app/*", a.serveSPAHandler(frontendFS))
}

// serveFrontendFile serves a specific file from the embedded frontend
func (a *Application) serveFrontendFile(frontendFS fs.FS, filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle Next.js RSC requests by stripping query parameters
		// RSC requests come with parameters like ?_rsc=1sdqn
		cleanFilename := strings.Split(r.URL.Path, "?")[0]
		cleanFilename = strings.TrimPrefix(cleanFilename, "/")

		// Use the requested path if it differs from the expected filename
		// This handles cases where the route pattern doesn't match the actual request
		if cleanFilename != "" && cleanFilename != filename {
			filename = cleanFilename
		}

		file, err := frontendFS.Open(filename)
		if err != nil {
			// For RSC requests, try removing any remaining query-like patterns
			if strings.Contains(filename, "_rsc") || strings.Contains(filename, "?") {
				// Extract base filename before any RSC parameters
				parts := strings.Split(filename, "?")
				if len(parts) > 0 {
					baseFilename := strings.Split(parts[0], "_rsc")[0]
					file, err = frontendFS.Open(baseFilename)
				}
			}

			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		defer file.Close()

		// Set appropriate content type
		switch filepath.Ext(filename) {
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		case ".txt":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		}

		// Set caching headers
		w.Header().Set("Cache-Control", "public, max-age=86400")

		io.Copy(w, file)
	}
}

// serveStaticWithMIME creates a file server that properly sets MIME types for embedded files
func (a *Application) serveStaticWithMIME(frontendFS fs.FS, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For embedded filesystem, we need to map URL paths to the correct embedded structure
		// URLs like /_next/static/... should map to _next/static/... in embedded FS
		path := r.URL.Path

		// Remove leading slash for fs.Open, but preserve the prefix structure
		if strings.HasPrefix(path, "/") {
			path = path[1:]
		}

		// Debug logging
		a.Logger.InfoContext(r.Context(), "Static file request",
			"original_url", r.URL.Path,
			"prefix", prefix,
			"resolved_path", path)

		// Try to open the file
		file, err := frontendFS.Open(path)
		if err != nil {
			a.Logger.WarnContext(r.Context(), "Static file not found",
				"path", path,
				"error", err.Error())
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		// Set content type based on extension
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		case ".txt":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".gif":
			w.Header().Set("Content-Type", "image/gif")
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		case ".woff2":
			w.Header().Set("Content-Type", "font/woff2")
		case ".woff":
			w.Header().Set("Content-Type", "font/woff")
		case ".ttf":
			w.Header().Set("Content-Type", "font/ttf")
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		default:
			w.Header().Set("Content-Type", "application/octet-stream")
		}

		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Copy file content
		io.Copy(w, file)
	})
}

// serveSPAHandler serves the Next.js SPA with license-first routing
func (a *Application) serveSPAHandler(frontendFS fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract request ID for logging
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = fmt.Sprintf("spa-%d", time.Now().UnixNano())
		}

		ctx := infrastructure.WithTraceID(r.Context(), reqID)

		// Log the SPA request
		a.Logger.InfoContext(ctx, "SPA route request",
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
			slog.String("user_agent", r.Header.Get("User-Agent")))

		// SMART REDIRECT: Route based on license status
		if r.URL.Path == "/" {
			// Check license status
			if a.Services != nil && a.Services.LicenseService != nil {
				status, err := a.Services.LicenseService.GetStatus(ctx)
				// If license is valid (active or warning), go to market overview
				if err == nil && (status.LicenseStatus == "active" || status.LicenseStatus == "warning") {
					a.Logger.InfoContext(ctx, "Redirecting to operations (valid license)",
						slog.String("license_status", status.LicenseStatus))
					http.Redirect(w, r, "/operations", http.StatusTemporaryRedirect)
					return
				}
			}
			// Otherwise, go to license page
			a.Logger.InfoContext(ctx, "Redirecting to license page (invalid/missing license)")
			http.Redirect(w, r, "/license", http.StatusTemporaryRedirect)
			return
		}

		urlPath := path.Clean(r.URL.Path)

		// Check license status for protected routes (redirect to /license if invalid)
		// Skip license check for license routes and API routes
		isLicenseRoute := strings.HasPrefix(urlPath, "/license")
		if !isLicenseRoute && !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/_next/") {
			if a.Services != nil && a.Services.LicenseService != nil {
				status, err := a.Services.LicenseService.GetStatus(ctx)
				// Allow both "active" and "warning" states (warning means <30 days left)
				if err != nil || (status.LicenseStatus != "active" && status.LicenseStatus != "warning") {
					a.Logger.InfoContext(ctx, "Redirecting to license page - invalid license",
						slog.String("path", r.URL.Path),
						slog.String("license_status", status.LicenseStatus),
						slog.Bool("has_error", err != nil))
					http.Redirect(w, r, "/license", http.StatusTemporaryRedirect)
					return
				}
			}
		}

		// Try to serve the specific file first
		if urlPath != "/" {
			exactPath := strings.TrimPrefix(urlPath, "/")
			a.Logger.InfoContext(ctx, "Trying exact path",
				slog.String("url_path", urlPath),
				slog.String("exact_path", exactPath))
			file, err := frontendFS.Open(exactPath)
			if err == nil {
				defer file.Close()

				// Check if this is a directory (which we don't want to serve directly)
				if stat, statErr := file.Stat(); statErr == nil && stat.IsDir() {
					a.Logger.InfoContext(ctx, "Exact path is directory, skipping",
						slog.String("exact_path", exactPath))
					// Continue to Next.js route check
				} else {
					a.Logger.InfoContext(ctx, "Serving exact file",
						slog.String("exact_path", exactPath))

					// Set content type based on extension
					ext := filepath.Ext(urlPath)
					switch ext {
					case ".html":
						w.Header().Set("Content-Type", "text/html; charset=utf-8")
					case ".js":
						w.Header().Set("Content-Type", "application/javascript")
					case ".css":
						w.Header().Set("Content-Type", "text/css")
					case ".json":
						w.Header().Set("Content-Type", "application/json")
					case ".txt":
						w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					case ".svg":
						w.Header().Set("Content-Type", "image/svg+xml")
					case ".png":
						w.Header().Set("Content-Type", "image/png")
					case ".jpg", ".jpeg":
						w.Header().Set("Content-Type", "image/jpeg")
					case ".gif":
						w.Header().Set("Content-Type", "image/gif")
					case ".ico":
						w.Header().Set("Content-Type", "image/x-icon")
					}

					// Set security headers
					w.Header().Set("X-Content-Type-Options", "nosniff")
					w.Header().Set("X-Frame-Options", "DENY")
					w.Header().Set("X-XSS-Protection", "1; mode=block")

					io.Copy(w, file)
					return
				}
			}

			// For Next.js routes like /license, try /license/index.html
			indexPath := strings.TrimPrefix(urlPath, "/") + "/index.html"
			a.Logger.InfoContext(ctx, "Trying Next.js route index.html",
				slog.String("url_path", urlPath),
				slog.String("index_path", indexPath))
			indexFile, err := frontendFS.Open(indexPath)
			if err == nil {
				defer indexFile.Close()

				a.Logger.InfoContext(ctx, "Serving Next.js route index.html",
					slog.String("route", urlPath),
					slog.String("file_path", indexPath))

				// Set headers for HTML content
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("X-Frame-Options", "DENY")
				w.Header().Set("X-XSS-Protection", "1; mode=block")
				w.Header().Set("X-Build-Time", time.Now().Format(time.RFC3339))

				io.Copy(w, indexFile)
				return
			} else {
				a.Logger.WarnContext(ctx, "Next.js route index.html not found",
					slog.String("url_path", urlPath),
					slog.String("index_path", indexPath),
					slog.String("error", err.Error()))
			}
		}

		// Fallback to index.html for SPA routing
		indexFile, err := frontendFS.Open("index.html")
		if err != nil {
			a.Logger.ErrorContext(ctx, "Failed to open index.html",
				slog.String("error", err.Error()),
				slog.String("path", urlPath))
			http.Error(w, "Frontend not available", http.StatusServiceUnavailable)
			return
		}
		defer indexFile.Close()

		// Set headers for HTML content
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Serve index.html for client-side routing
		io.Copy(w, indexFile)

		a.Logger.DebugContext(ctx, "Served SPA index.html",
			slog.String("original_path", urlPath))
	}
}

// getCORSConfig returns CORS configuration based on environment
func (a *Application) getCORSConfig() middleware.CORSConfig {
	// Detect environment
	isDevelopment := a.isDevelopmentMode()

	config := middleware.CORSConfig{
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Request-ID",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           300,
		Logger:           a.Logger,
	}

	if isDevelopment {
		// Development mode: Allow Next.js dev server
		config.AllowedOrigins = []string{
			"http://localhost:3000", // Next.js dev server
			"http://127.0.0.1:3000",
			"http://localhost:8080", // Go server
			"http://127.0.0.1:8080",
		}
		a.Logger.Info("CORS configured for development mode",
			slog.Any("allowed_origins", config.AllowedOrigins))
	} else {
		// Production mode: Only allow same origin
		config.AllowedOrigins = []string{
			"http://localhost:8080",
			"http://127.0.0.1:8080",
		}

		// Add any configured origins
		if a.Config.Security.EnableCORS && len(a.Config.Security.AllowedOrigins) > 0 {
			config.AllowedOrigins = append(config.AllowedOrigins, a.Config.Security.AllowedOrigins...)
		}

		a.Logger.Info("CORS configured for production mode",
			slog.Any("allowed_origins", config.AllowedOrigins))
	}

	return config
}

// isDevelopmentMode detects if we're running in development mode
func (a *Application) isDevelopmentMode() bool {
	// Check environment variable
	if env := os.Getenv("NODE_ENV"); env == "development" {
		return true
	}
	if env := os.Getenv("GO_ENV"); env == "development" {
		return true
	}

	// Check if Next.js dev files exist (indicates development)
	if _, err := os.Stat("frontend/package.json"); err == nil {
		if _, err := os.Stat("frontend/.next"); err == nil {
			return true
		}
	}

	// Check if running from dev directory
	if wd, err := os.Getwd(); err == nil {
		if strings.Contains(wd, "dev") || strings.Contains(wd, "development") {
			return true
		}
	}

	return false
}
