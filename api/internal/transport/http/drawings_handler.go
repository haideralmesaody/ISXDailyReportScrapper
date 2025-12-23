package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	apierrors "github.com/isxcli/isxcli/internal/errors"
)

type DrawingsHandler struct {
	dataDir      string
	logger       *slog.Logger
	errorHandler *apierrors.ErrorHandler
	locks        sync.Map // ticker -> *sync.Mutex
}

func NewDrawingsHandler(dataDir string, logger *slog.Logger, errorHandler *apierrors.ErrorHandler) *DrawingsHandler {
	return &DrawingsHandler{
		dataDir:      dataDir,
		logger:       logger.With(slog.String("component", "drawings_handler")),
		errorHandler: errorHandler,
	}
}

func (h *DrawingsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/v1/drawings", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.Route("/{ticker}", func(r chi.Router) {
			r.Use(h.tickerCtx)
			r.Get("/", h.GetDrawings)
			r.Put("/", h.PutDrawings)
		})
	})
}

var tickerPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,10}$`)

func (h *DrawingsHandler) tickerCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ticker := chi.URLParam(r, "ticker")
		ticker = strings.TrimSpace(ticker)
		if ticker == "" {
			h.errorHandler.HandleError(w, r, apierrors.ErrValidation("ticker", "Ticker symbol is required"))
			return
		}
		if !tickerPattern.MatchString(ticker) {
			h.errorHandler.HandleError(w, r, apierrors.ErrValidation("ticker", "Invalid ticker symbol format"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

type drawingPoint struct {
	Time  float64 `json:"time"`
	Price float64 `json:"price"`
}

type drawingShape struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Points    []drawingPoint         `json:"points"`
	Pane      int                    `json:"pane"`
	Style     map[string]any         `json:"style,omitempty"`
	Meta      map[string]any         `json:"meta,omitempty"`
	Locked    bool                   `json:"locked"`
	CreatedAt int64                  `json:"createdAt"`
	UpdatedAt int64                  `json:"updatedAt"`
	Text      string                 `json:"text,omitempty"`
}

type drawingsEnvelope struct {
	Version   int           `json:"version"`
	Ticker    string        `json:"ticker"`
	UpdatedAt int64         `json:"updatedAt"`
	Shapes    []drawingShape `json:"shapes"`
}

func (h *DrawingsHandler) getLock(ticker string) *sync.Mutex {
	value, _ := h.locks.LoadOrStore(ticker, &sync.Mutex{})
	return value.(*sync.Mutex)
}

func (h *DrawingsHandler) filePath(ticker string) string {
	ticker = strings.ToUpper(ticker)
	return filepath.Join(h.dataDir, "indicators", fmt.Sprintf("%s.json", ticker))
}

func (h *DrawingsHandler) GetDrawings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticker := strings.ToUpper(chi.URLParam(r, "ticker"))

	lock := h.getLock(ticker)
	lock.Lock()
	defer lock.Unlock()

	path := h.filePath(ticker)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			render.JSON(w, r, drawingsEnvelope{
				Version:   1,
				Ticker:    ticker,
				UpdatedAt: 0,
				Shapes:    []drawingShape{},
			})
			return
		}
		h.logger.ErrorContext(ctx, "failed reading drawings file", slog.String("ticker", ticker), slog.String("path", path), slog.String("error", err.Error()))
		h.errorHandler.HandleError(w, r, apierrors.NewInternalError("failed to read drawings"))
		return
	}

	var envelope drawingsEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		// Backward-compatible: allow raw shape array.
		var shapes []drawingShape
		if err2 := json.Unmarshal(data, &shapes); err2 != nil {
			h.logger.ErrorContext(ctx, "failed decoding drawings json", slog.String("ticker", ticker), slog.String("path", path), slog.String("error", err.Error()))
			h.errorHandler.HandleError(w, r, apierrors.NewInternalError("invalid drawings file"))
			return
		}
		envelope = drawingsEnvelope{
			Version:   1,
			Ticker:    ticker,
			UpdatedAt: time.Now().UnixMilli(),
			Shapes:    shapes,
		}
	}

	// Always normalize ticker in response.
	envelope.Ticker = ticker
	if envelope.Version == 0 {
		envelope.Version = 1
	}
	if envelope.Shapes == nil {
		envelope.Shapes = []drawingShape{}
	}

	render.JSON(w, r, envelope)
}

func (h *DrawingsHandler) PutDrawings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticker := strings.ToUpper(chi.URLParam(r, "ticker"))

	lock := h.getLock(ticker)
	lock.Lock()
	defer lock.Unlock()

	body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20)) // 2MB
	if err != nil {
		h.errorHandler.HandleError(w, r, apierrors.NewValidationError("invalid request body"))
		return
	}

	var envelope drawingsEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		// Also allow raw shape array for simplicity.
		var shapes []drawingShape
		if err2 := json.Unmarshal(body, &shapes); err2 != nil {
			h.errorHandler.HandleError(w, r, apierrors.NewValidationError("invalid json payload"))
			return
		}
		envelope = drawingsEnvelope{
			Version:   1,
			Ticker:    ticker,
			UpdatedAt: time.Now().UnixMilli(),
			Shapes:    shapes,
		}
	}

	if envelope.Version == 0 {
		envelope.Version = 1
	}
	envelope.Ticker = ticker
	envelope.UpdatedAt = time.Now().UnixMilli()

	if err := validateShapes(envelope.Shapes); err != nil {
		h.errorHandler.HandleError(w, r, apierrors.NewValidationError(err.Error()))
		return
	}

	path := h.filePath(ticker)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		h.logger.ErrorContext(ctx, "failed creating drawings directory", slog.String("ticker", ticker), slog.String("path", path), slog.String("error", err.Error()))
		h.errorHandler.HandleError(w, r, apierrors.NewInternalError("failed to prepare storage"))
		return
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), fmt.Sprintf("%s.*.tmp", ticker))
	if err != nil {
		h.logger.ErrorContext(ctx, "failed creating temp drawings file", slog.String("ticker", ticker), slog.String("error", err.Error()))
		h.errorHandler.HandleError(w, r, apierrors.NewInternalError("failed to write drawings"))
		return
	}

	tmpPath := tmp.Name()
	encoder := json.NewEncoder(tmp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&envelope); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		h.errorHandler.HandleError(w, r, apierrors.NewInternalError("failed to write drawings"))
		return
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		h.errorHandler.HandleError(w, r, apierrors.NewInternalError("failed to write drawings"))
		return
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		h.logger.ErrorContext(ctx, "failed finalizing drawings file", slog.String("ticker", ticker), slog.String("error", err.Error()))
		h.errorHandler.HandleError(w, r, apierrors.NewInternalError("failed to write drawings"))
		return
	}

	render.JSON(w, r, map[string]any{
		"status":     "success",
		"ticker":     ticker,
		"updatedAt":  envelope.UpdatedAt,
		"shapeCount": len(envelope.Shapes),
	})
}

var allowedShapeTypes = map[string]bool{
	"trendLine":       true,
	"ray":             true,
	"parallelChannel": true,
	"horizontalLine":  true,
	"verticalLine":    true,
	"rectangle":       true,
	"text":            true,
	"arrow":           true,
	"measure":         true,
}

func validateShapes(shapes []drawingShape) error {
	for i := range shapes {
		shape := shapes[i]
		if strings.TrimSpace(shape.ID) == "" {
			return fmt.Errorf("shape[%d].id is required", i)
		}
		if strings.TrimSpace(shape.Type) == "" {
			return fmt.Errorf("shape[%d].type is required", i)
		}
		if !allowedShapeTypes[shape.Type] {
			return fmt.Errorf("shape[%d].type is not supported", i)
		}
		if shape.Pane < 0 {
			return fmt.Errorf("shape[%d].pane must be >= 0", i)
		}
		if len(shape.Points) == 0 {
			return fmt.Errorf("shape[%d].points is required", i)
		}
		if shape.Type == "parallelChannel" && len(shape.Points) < 3 {
			return fmt.Errorf("shape[%d].points must have at least 3 points for parallelChannel", i)
		}
		for j := range shape.Points {
			p := shape.Points[j]
			if p.Time != p.Time || p.Price != p.Price { // NaN check
				return fmt.Errorf("shape[%d].points[%d] contains invalid values", i, j)
			}
		}
		if shape.Type == "text" && shape.Text == "" {
			// Allow empty text but keep payload stable.
			continue
		}
	}
	return nil
}
