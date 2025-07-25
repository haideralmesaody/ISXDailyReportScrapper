package api

import (
	"fmt"
	"net/http"
	"github.com/go-chi/chi/v5"
)

// RegisterHandlers registers all API routes
func RegisterHandlers(r *chi.Mux) {
	r.With(middleware).Get("/api/reports", getReports)
	r.With(middleware).Get("/api/license", getLicenseStatus)
	r.With(middleware).Post("/api/settings", updateSettings)
}

// getReports handles report data requests
func getReports(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{\"error\":\"Not implemented\"}")
	w.WriteHeader(http.StatusNotImplemented)
}

// getLicenseStatus checks current license status
func getLicenseStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{\"error\":\"Not implemented\"}")
	w.WriteHeader(http.StatusNotImplemented)
}

// updateSettings handles settings updates
func updateSettings(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{\"error\":\"Not implemented\"}")
	w.WriteHeader(http.StatusNotImplemented)
}

// middleware adds common request handling
func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}