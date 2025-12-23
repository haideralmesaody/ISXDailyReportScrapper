package app

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/infrastructure"
	ws "github.com/isxcli/isxcli/internal/websocket"

	"github.com/gorilla/websocket"
	"log/slog"
)

// handleWebSocket handles WebSocket connections with comprehensive diagnostics
func (a *Application) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Enhanced request timing and diagnostics
	startTime := time.Now()

	// Extract any available request ID (might not have middleware)
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = fmt.Sprintf("ws-%d", time.Now().UnixNano())
	}

	// Structured logging per CLAUDE.md with enhanced diagnostics
	ctx := infrastructure.WithTraceID(r.Context(), reqID)

	// Log detailed request information for debugging
	a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] WebSocket upgrade request received",
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("origin", r.Header.Get("Origin")),
		slog.String("host", r.Host),
		slog.String("user_agent", r.UserAgent()),
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("query", r.URL.RawQuery),
		slog.String("connection_header", r.Header.Get("Connection")),
		slog.String("upgrade_header", r.Header.Get("Upgrade")),
		slog.String("sec_websocket_key", r.Header.Get("Sec-WebSocket-Key")),
		slog.String("sec_websocket_version", r.Header.Get("Sec-WebSocket-Version")),
		slog.String("sec_websocket_protocol", r.Header.Get("Sec-WebSocket-Protocol")),
		slog.String("accept_header", r.Header.Get("Accept")),
		slog.String("accept_language", r.Header.Get("Accept-Language")),
		slog.String("accept_encoding", r.Header.Get("Accept-Encoding")),
		slog.String("referer", r.Header.Get("Referer")),
		slog.String("forwarded_for", r.Header.Get("X-Forwarded-For")),
		slog.String("real_ip", r.Header.Get("X-Real-IP")),
		slog.Time("request_time", startTime),
		slog.String("server_time", time.Now().Format(time.RFC3339Nano)))

	// Log server state for diagnostics
	a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] Server state at connection request",
		slog.Int("hub_clients", a.WebSocketHub.ClientCount()),
		slog.Bool("hub_running", a.WebSocketHub.IsRunning()),
		slog.String("server_addr", a.Server.Addr),
		slog.String("server_uptime", time.Since(startTime).String()),
		slog.Bool("middleware_applied", false), // WebSocket route bypasses middleware
		slog.String("app_mode", func() string {
			if a.isDevelopmentMode() {
				return "development"
			}
			return "production"
		}()))

	// Set CORS headers explicitly for WebSocket upgrade with debugging
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Handle cases where Origin header is missing (e.g., file:// protocol)
		origin = fmt.Sprintf("http://%s", r.Host)
		a.Logger.WarnContext(ctx, "🔍 [WS-DEBUG] No Origin header in WebSocket request, using host",
			slog.String("host", r.Host),
			slog.String("inferred_origin", origin))
	}

	// Log CORS headers being set
	a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] Setting CORS headers for WebSocket",
		slog.String("origin", origin),
		slog.String("access_control_allow_origin", origin),
		slog.String("access_control_allow_credentials", "true"),
		slog.String("access_control_allow_methods", "GET, OPTIONS"),
		slog.String("access_control_allow_headers", "Content-Type, Authorization, X-Request-ID"))

	// Set WebSocket-specific CORS headers
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

	// Enhanced upgrader with detailed logging
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			checkStart := time.Now()
			origin := r.Header.Get("Origin")
			host := r.Host
			remoteAddr := r.RemoteAddr

			a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] Origin check started",
				slog.String("origin", origin),
				slog.String("host", host),
				slog.String("remote_addr", remoteAddr),
				slog.String("user_agent", r.UserAgent()))

			// SIMPLIFIED ORIGIN CHECK: Always allow localhost and development origins
			// Allow if no origin (local file or same-origin request)
			if origin == "" {
				a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] WebSocket origin check - no origin header, allowing",
					slog.String("host", host),
					slog.String("remote_addr", remoteAddr),
					slog.Duration("check_duration", time.Since(checkStart)))
				return true
			}

			// Always allow localhost origins for development
			if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") ||
				strings.Contains(host, "::1") || strings.Contains(remoteAddr, "127.0.0.1") ||
				strings.Contains(remoteAddr, "::1") {
				a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] WebSocket origin check - localhost/development origin allowed",
					slog.String("origin", origin),
					slog.String("host", host),
					slog.String("remote_addr", remoteAddr),
					slog.Duration("check_duration", time.Since(checkStart)))
				return true
			}

			// Allow same-host origins regardless of scheme
			if parsedOrigin, err := url.Parse(origin); err == nil {
				originHost := parsedOrigin.Host
				originHostname := parsedOrigin.Hostname()

				reqHost := r.Host
				reqHostname := reqHost
				if h, _, err := net.SplitHostPort(reqHost); err == nil && h != "" {
					reqHostname = h
				}

				if strings.EqualFold(originHost, reqHost) || strings.EqualFold(originHostname, reqHostname) {
					a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] WebSocket origin check - allowing same host origin",
						slog.String("origin", origin),
						slog.String("request_host", r.Host),
						slog.String("origin_host", originHost),
						slog.String("origin_hostname", originHostname),
						slog.String("req_hostname", reqHostname),
						slog.Duration("check_duration", time.Since(checkStart)))
					return true
				}
			}

			// In development mode, be very permissive
			if a.isDevelopmentMode() {
				a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] WebSocket origin check - development mode, allowing all origins",
					slog.String("origin", origin),
					slog.String("host", host),
					slog.Duration("check_duration", time.Since(checkStart)))
				return true
			}

			// In production, validate against allowed origins
			corsConfig := a.getCORSConfig()
			for _, allowed := range corsConfig.AllowedOrigins {
				if origin == allowed {
					a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] WebSocket origin check - origin allowed",
						slog.String("origin", origin),
						slog.Duration("check_duration", time.Since(checkStart)))
					return true
				}
			}

			a.Logger.WarnContext(ctx, "🔍 [WS-DEBUG] WebSocket origin check - origin not allowed",
				slog.String("origin", origin),
				slog.String("host", host),
				slog.String("remote_addr", remoteAddr),
				slog.Any("allowed_origins", corsConfig.AllowedOrigins),
				slog.Duration("check_duration", time.Since(checkStart)))
			return false
		},
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 10 * time.Second, // Add handshake timeout
	}

	// Log upgrade attempt
	upgradeStart := time.Now()
	a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] Starting WebSocket upgrade",
		slog.Time("upgrade_start", upgradeStart),
		slog.String("read_buffer_size", "1024"),
		slog.String("write_buffer_size", "1024"),
		slog.String("handshake_timeout", "10s"))

	conn, err := upgrader.Upgrade(w, r, nil)
	upgradeDuration := time.Since(upgradeStart)

	if err != nil {
		// Enhanced error logging with specific error types
		errorType := "unknown"
		errorDetails := ""

		if websocket.IsCloseError(err) {
			if closeErr, ok := err.(*websocket.CloseError); ok {
				errorType = fmt.Sprintf("close_error_%d", closeErr.Code)
				errorDetails = closeErr.Text
			}
		} else if strings.Contains(err.Error(), "not a websocket handshake") {
			errorType = "invalid_handshake"
		} else if strings.Contains(err.Error(), "bad request") {
			errorType = "bad_request"
		} else if strings.Contains(err.Error(), "connection upgrade") {
			errorType = "upgrade_failed"
		} else if strings.Contains(err.Error(), "timeout") {
			errorType = "timeout"
		}

		a.Logger.ErrorContext(ctx, "❌ [WS-ERROR] WebSocket upgrade failed",
			slog.String("error", err.Error()),
			slog.String("error_type", errorType),
			slog.String("error_details", errorDetails),
			slog.String("details", fmt.Sprintf("%+v", err)),
			slog.String("origin", origin),
			slog.Duration("upgrade_duration", upgradeDuration),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("request_id", reqID),
			slog.String("user_agent", r.UserAgent()),
			slog.String("connection_header", r.Header.Get("Connection")),
			slog.String("upgrade_header", r.Header.Get("Upgrade")),
			slog.String("sec_websocket_key", r.Header.Get("Sec-WebSocket-Key")),
			slog.String("sec_websocket_version", r.Header.Get("Sec-WebSocket-Version")),
			slog.String("server_time", time.Now().Format(time.RFC3339Nano)),
			slog.Int("hub_clients", a.WebSocketHub.ClientCount()))

		// Log response headers that were set
		a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] Response headers at error",
			slog.String("content_type", w.Header().Get("Content-Type")),
			slog.String("access_control_allow_origin", w.Header().Get("Access-Control-Allow-Origin")),
			slog.String("access_control_allow_credentials", w.Header().Get("Access-Control-Allow-Credentials")))

		return
	}

	connectionTime := time.Since(startTime)

	a.Logger.InfoContext(ctx, "✅ [WS-SUCCESS] WebSocket connection established, starting client setup",
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("request_id", reqID),
		slog.String("origin", origin),
		slog.Duration("total_connection_time", connectionTime),
		slog.Duration("upgrade_duration", upgradeDuration),
		slog.String("local_addr", conn.LocalAddr().String()),
		slog.String("remote_addr_ws", conn.RemoteAddr().String()),
		slog.Bool("websocket_subprotocol", len(conn.Subprotocol()) > 0),
		slog.String("websocket_subprotocol", conn.Subprotocol()),
		slog.String("server_time", time.Now().Format(time.RFC3339Nano)),
		slog.Int("hub_clients_before", a.WebSocketHub.ClientCount()))

	// Verify hub is running before proceeding
	if !a.WebSocketHub.IsRunning() {
		a.Logger.ErrorContext(ctx, "❌ [WS-ERROR] WebSocket hub is not running, cannot accept connection",
			slog.String("request_id", reqID),
			slog.String("remote_addr", r.RemoteAddr))
		conn.Close()
		return
	}

	// Log hub status after connection
	hubClientsAfter := a.WebSocketHub.ClientCount()

	a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] Hub status before client registration",
		slog.Int("clients_before", a.WebSocketHub.ClientCount()),
		slog.String("request_id", reqID),
		slog.String("remote_addr", r.RemoteAddr))

	// Use the improved ServeWS function that handles connection state properly
	ws.ServeWS(a.WebSocketHub, conn)

	// Log hub status after registration (will be logged asynchronously from hub)
	a.Logger.InfoContext(ctx, "🔍 [WS-DEBUG] Client handed to hub for registration",
		slog.Int("clients_before_handoff", hubClientsAfter),
		slog.String("request_id", reqID),
		slog.String("remote_addr", r.RemoteAddr))
}
