package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apierrors "github.com/isxcli/isxcli/internal/errors"
)

func TestDrawingsHandler_GetDrawings_ReturnsEmptyEnvelopeWhenMissing(t *testing.T) {
	dataDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	errorHandler := apierrors.NewErrorHandler(logger, false)

	handler := NewDrawingsHandler(dataDir, logger, errorHandler)

	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.RegisterRoutes(r)
	})

	req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/drawings/ABC/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	require.Equal(t, 200, res.StatusCode)

	var envelope map[string]any
	require.NoError(t, json.NewDecoder(res.Body).Decode(&envelope))

	assert.Equal(t, float64(1), envelope["version"])
	assert.Equal(t, "ABC", envelope["ticker"])

	shapes, ok := envelope["shapes"].([]any)
	require.True(t, ok)
	assert.Len(t, shapes, 0)
}

func TestDrawingsHandler_PutThenGet_RoundTripsShapes(t *testing.T) {
	dataDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	errorHandler := apierrors.NewErrorHandler(logger, false)

	handler := NewDrawingsHandler(dataDir, logger, errorHandler)

	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.RegisterRoutes(r)
	})

	payload := map[string]any{
		"version": 1,
		"ticker":  "abc",
		"shapes": []map[string]any{
			{
				"id":   "shape-1",
				"type": "trendLine",
				"points": []map[string]any{
					{"time": 1, "price": 10.5},
					{"time": 2, "price": 11.25},
				},
				"pane":      0,
				"locked":    false,
				"createdAt": 1,
				"updatedAt": 1,
				"style": map[string]any{
					"strokeColor": "#22c55e",
					"strokeWidth": 2,
					"strokeStyle": "solid",
				},
			},
			{
				"id":   "shape-2",
				"type": "parallelChannel",
				"points": []map[string]any{
					{"time": 1, "price": 1},
					{"time": 2, "price": 1},
					{"time": 1, "price": 2},
				},
				"pane":      0,
				"locked":    false,
				"createdAt": 1,
				"updatedAt": 1,
			},
		},
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	putReq := httptest.NewRequest(nethttp.MethodPut, "/api/v1/drawings/abc/", bytes.NewReader(body))
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)

	putRes := putRec.Result()
	defer putRes.Body.Close()
	require.Equal(t, 200, putRes.StatusCode)

	getReq := httptest.NewRequest(nethttp.MethodGet, "/api/v1/drawings/ABC/", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	getRes := getRec.Result()
	defer getRes.Body.Close()
	require.Equal(t, 200, getRes.StatusCode)

	var envelope map[string]any
	require.NoError(t, json.NewDecoder(getRes.Body).Decode(&envelope))

	assert.Equal(t, "ABC", envelope["ticker"])
	shapes := envelope["shapes"].([]any)
	require.Len(t, shapes, 2)

	first := shapes[0].(map[string]any)
	assert.Equal(t, "shape-1", first["id"])
	assert.Equal(t, "trendLine", first["type"])
}

func TestDrawingsHandler_GetDrawings_BackwardCompatibleArrayFile(t *testing.T) {
	dataDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	errorHandler := apierrors.NewErrorHandler(logger, false)

	handler := NewDrawingsHandler(dataDir, logger, errorHandler)

	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.RegisterRoutes(r)
	})

	path := filepath.Join(dataDir, "indicators", "AAA.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))

	rawShapes := []map[string]any{
		{
			"id":   "shape-1",
			"type": "horizontalLine",
			"points": []map[string]any{
				{"time": 1, "price": 100},
			},
			"pane":      0,
			"locked":    false,
			"createdAt": 1,
			"updatedAt": 1,
		},
	}
	rawBody, err := json.Marshal(rawShapes)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, rawBody, 0o644))

	req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/drawings/AAA/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)

	var envelope map[string]any
	require.NoError(t, json.NewDecoder(res.Body).Decode(&envelope))

	assert.Equal(t, float64(1), envelope["version"])
	assert.Equal(t, "AAA", envelope["ticker"])
	shapes := envelope["shapes"].([]any)
	require.Len(t, shapes, 1)
}

func TestDrawingsHandler_PutDrawings_Validation(t *testing.T) {
	dataDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	errorHandler := apierrors.NewErrorHandler(logger, false)

	handler := NewDrawingsHandler(dataDir, logger, errorHandler)

	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.RegisterRoutes(r)
	})

	body := []byte(`{"version":1,"ticker":"AAA","shapes":[{"id":"1","type":"not-supported","points":[{"time":1,"price":2}],"pane":0,"locked":false,"createdAt":1,"updatedAt":1}]}`)

	req := httptest.NewRequest(nethttp.MethodPut, "/api/v1/drawings/AAA/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, 400, res.StatusCode)
}
