# ISX Daily Reports Scrapper – Development Standards

> **Purpose** — This CLAUDE.md sets the coding, testing, documentation, and governance standards that Claude Code (claude.ai/code) must follow when generating, reviewing, or proposing changes for this repository.

---

## Core Principles  

| Principle | What It Means in This Repo |
|-----------|----------------------------|
| **Idiomatic Go** | Follow *Effective Go*, *Go Code Review Comments*, and **Uber Go Style**. Prefer small, cohesive functions; avoid premature abstraction. |
| **Clear Architecture** | Thin HTTP handlers (Chi) → business logic in services → isolated data‑access layer. No cross‑layer imports. |
| **Everything Observable** | Structured logs (slog), contextual errors, **OpenTelemetry** traces & metrics; every request traceable by ID. |
| **Test What You Write** | Table‑driven tests, race detector, ≥ 90 % coverage for critical pkgs, ≥ 80 % for others. |
| **Docs‑as‑Code** | Update docs, changelog, and build scripts **in the same PR** as the code change. |
| **No Blind Sleeps** | Replace `time.Sleep`/busy‑waits with context, channels, timers, back‑off, or `errgroup`. |

---

## Build & Run Commands
```bash
# Build all executables
build.bat           # Windows
task build          # Taskfile (cross‑platform)

# Start the web server locally
start-server.bat    # Windows
task run            # Taskfile (cross‑platform)

# Direct run from release
cd release && ./web-licensed.exe
```

---

## Development Commands
```bash
# Lint, vet, static analysis
task lint           # wraps go vet + golangci-lint

# Run all tests with race + coverage
task test           # go test -race -coverprofile=coverage.out ./...
task coverhtml      # opens coverage.html

# Module hygiene
task tidy           # go mod tidy && go mod download
```

---

## Architecture Overview
### Core Components
| Component | Purpose |
|-----------|---------|
| **Pipeline System** | Multi‑stage data processing with retries, progress tracking, and WebSocket updates. |
| **WebSocket Manager** | Bi‑directional JSON protocol for live progress and events. |
| **License Manager** | File‑based, AES‑GCM‑encrypted licenses; checked early in request flow. |
| **Service Layer** | Pure business logic; injected with interfaces. |
| **Data Layer** | Interfaces for DB / filesystem / external APIs; no HTTP details leak inward. |

### Patterns & Practices
- Constructor‑based **dependency injection**.
- **RFC 7807** error payloads (see § Error Handling).
- **OpenTelemetry** middleware for tracing & metrics.
- Graceful shutdown with `server.Shutdown(ctx)`.

---

## HTTP API & Middleware (Chi)
```go
r := chi.NewRouter()

// Core middleware (order matters)
r.Use(middleware.RequestID)          // must be first
r.Use(middleware.RealIP)
r.Use(LoggerMiddleware)              // slog + trace id
r.Use(middleware.Recoverer)
r.Use(middleware.Timeout(60 * time.Second))

// Security / governance
r.Use(AuthMiddleware)                // JWT or API key
r.Use(RateLimitMiddleware)           // token bucket
r.Use(otelchi.Middleware("isx-api")) // OpenTelemetry

// Optional
r.Use(middleware.Compress(5))
r.Use(middleware.StripSlashes)
```

### Handler Skeleton
```go
func (h *ResourceHandler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Get("/", h.List)
    r.Post("/", h.Create)
    r.Route("/{id}", func(r chi.Router) {
        r.Get("/", h.Get)
        r.Put("/", h.Update)
        r.Delete("/", h.Delete)
    })
    return r
}

// Example handler
func (h *ResourceHandler) Get(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    res, err := h.svc.Get(r.Context(), id)
    if err != nil {
        render.Render(w, r, httperr.From(err)) // RFC 7807 mapper
        return
    }
    render.JSON(w, r, res)
}
```

---

## Error Handling – **RFC 7807**
### Go Struct
```go
type Problem struct {
    Type   string `json:"type"`             // URI reference
    Title  string `json:"title"`            // short, human‑readable
    Status int    `json:"status"`           // HTTP status
    Detail string `json:"detail,omitempty"` // human‑readable detail
    Trace  string `json:"trace_id,omitempty"`
}

func (p Problem) Render(w http.ResponseWriter, r *http.Request) error {
    render.Status(r, p.Status)
    return nil
}
```

### Mapper
```go
func From(err error) *Problem {
    switch {
    case errors.Is(err, service.ErrNotFound):
        return &Problem{
            Type:   "/problems/not-found",
            Title:  "Resource Not Found",
            Status: http.StatusNotFound,
            Detail: err.Error(),
        }
    // more mappings …
    default:
        return &Problem{
            Type:   "/problems/internal",
            Title:  "Internal Server Error",
            Status: http.StatusInternalServerError,
            Detail: "unexpected error",
        }
    }
}
```

---

## Logging & Observability
### slog Setup
```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level:     slog.LevelInfo,
    AddSource: true,
}))
slog.SetDefault(logger)
```

### OpenTelemetry
```go
tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter),
    sdktrace.WithResource(resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceName("isx-api"),
    )),
)
otel.SetTracerProvider(tp)
```
- **Correlation**: `trace_id` added to every log line via slog handler.
- **Metrics**: Prometheus exporter exposed at `/metrics`.

---

## Configuration & Paths
| Rule | Implementation |
|------|----------------|
| **12‑Factor** | `CONFIG_PATH=/etc/isx/config.yaml` or env‑vars only. |
| Validation | Parse & validate config at startup; abort on error. |
| Paths | Use `path/filepath`; embed static assets with `//go:embed`. |

---

## Testing & Continuous Integration
### Local
```bash
task test         # race + coverage
task lint         # vet + golangci-lint
task coverhtml    # open coverage report
```

### Coverage Targets
| Package Group | Min Coverage |
|---------------|--------------|
| Critical (pipeline, licensing, handlers) | **≥ 90 %** |
| Other | **≥ 80 %** |

### GitHub Actions Workflow (excerpt)
```yaml
jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - run: task lint
      - run: task test
      - name: Docs-check
        run: ./scripts/ensure-docs-updated.sh
```
`ensure-docs-updated.sh` fails if code changes touch public APIs but docs/build scripts or `CHANGELOG.md` aren’t updated.

---

## Docs & Build‑Script Standards
1. **When behavior / flags change**: update `README.md`, API docs, ADRs, and `build.bat` / `Taskfile.yml`.
2. PR template reminds authors to tick “docs updated” & “build scripts updated”.
3. Exceptions logged in `/docs/exceptions.md` and reviewed quarterly.

---

## Security & Compliance
- **Standards**: OWASP ASVS, NIST SSDF.
- **Secrets**: fetched via vault (Azure Key‑Vault / AWS SSM).
- **Transport**: TLS 1.2+; security headers (`CSP`, `HSTS`, `X-Content-Type-Options`).
- **SBOM**: generated via CycloneDX.
- **Scanning**: `gitleaks`, `trivy`, Dependabot.
- **Signing**: release artifacts signed with Sigstore.

---

## Reliability & Performance
| Area | Practice |
|------|----------|
| **Resilience** | Retries + exp. back‑off (`backoff/v4`), circuit breakers (`sony/gobreaker`). |
| **Health** | `/healthz` (fast) & `/readyz` (dependency checks). |
| **Scaling** | Stateless services + HPA. |
| **Long Jobs** | Off‑load to queue (NATS JetStream) and notify via WS. |
| **Deploy** | Immutable images, IaC (Terraform), blue‑green or canary, auto‑rollback on failed SLOs. |

---

## Governance
- **Conventional Commits** + `CHANGELOG.md` (keep‑a‑changelog).
- **Architectural Decision Records** in `/docs/adr/`.
- **Quarterly** review of this standard; deviations captured in `/docs/exceptions.md`.

---

## Common Tasks Cheat‑sheet
| Task | Steps |
|------|-------|
| **Add Pipeline Stage** | 1) Implement `Stage` in `internal/pipeline/`.<br>2) Register in stage registry.<br>3) Add OTEL span + WS msg.<br>4) Write unit + integration tests. |
| **Add API Endpoint** | 1) Add handler in `internal/handlers/`.<br>2) Wire route.<br>3) Define request/response models + RFC 7807 errors.<br>4) Update OpenAPI spec & docs. |
| **Modify WebSocket Types** | 1) Update `internal/websocket/types.go`.<br>2) Maintain backward compatibility.<br>3) Update frontend listeners.<br>4) Bump API version if breaking. |

---

*End of file.*

