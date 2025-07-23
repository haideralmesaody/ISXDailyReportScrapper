# Chi Framework Migration Plan

## Overview

This document outlines the migration from gorilla/mux to Chi framework to resolve critical HTTP response issues and improve the overall web application architecture.

## Problem Statement

### Current Issues
1. **HTTP 206 Partial Content Responses**: Custom LoggingMiddleware causing incomplete HTML delivery
2. **Response Hanging**: Browser receives HTML but doesn't parse it due to middleware interference
3. **WebSocket Interference**: Middleware affecting WebSocket upgrade process
4. **Complex Middleware Implementation**: Custom responseWriter wrapper with interface compatibility issues

### Root Cause
The custom LoggingMiddleware in `dev/cmd/web-licensed/main.go` (lines 107-161) is interfering with HTTP response completion, causing browsers to receive incomplete responses despite the correct content being sent.

## Chi Framework Selection

### Why Chi?
1. **Google-Aligned**: Follows Go standard library patterns and idioms
2. **Lightweight**: Minimal overhead with fast routing performance
3. **Composable**: Middleware can be applied selectively to route groups
4. **Battle-Tested**: Used in production by many Go applications
5. **Standard Library Compatible**: Works seamlessly with net/http interfaces
6. **Active Maintenance**: Regular updates and security fixes

### Chi vs Alternatives
| Framework | Pros | Cons | Decision |
|-----------|------|------|----------|
| **Chi** | Lightweight, composable, Google-aligned | Learning curve | ✅ **Selected** |
| Gin | Fast, popular | Not standard library aligned | ❌ |
| Echo | Feature-rich | Heavy, non-standard patterns | ❌ |
| Fiber | Very fast | Express.js-like (not Go idiomatic) | ❌ |

## Migration Architecture

### Current Architecture (gorilla/mux)
```
HTTP Request → LoggingMiddleware (PROBLEMATIC) → gorilla/mux Router → Handler
```

### Target Architecture (Chi)
```
HTTP Request → Chi Router → Middleware Stack → Route Groups → Handler
                           ├── RequestID
                           ├── RealIP  
                           ├── Logger (Chi's)
                           ├── Recoverer
                           └── Timeout
```

## Implementation Plan

### Phase 1: Documentation & Preparation
1. ✅ Create this migration plan
2. ✅ Update development roadmap
3. ✅ Create UAT test document
4. ✅ Update CHANGELOG.md
5. ✅ Push documentation to Git

### Phase 2: Dependencies
```go
// Add to dev/go.mod
github.com/go-chi/chi/v5 v5.1.0
github.com/go-chi/middleware v5.1.0
```

### Phase 3: Router Migration
Replace gorilla/mux implementation with Chi-based routing:

```go
func setupRoutes() *chi.Mux {
    r := chi.NewRouter()
    
    // Global middleware stack
    r.Use(middleware.RequestID)     // Add unique request IDs
    r.Use(middleware.RealIP)        // Extract real client IP
    r.Use(ChiLoggingMiddleware)     // Custom logging (Chi-compatible)
    r.Use(middleware.Recoverer)     // Panic recovery
    r.Use(middleware.Timeout(60 * time.Second)) // Request timeouts
    
    // WebSocket endpoint (no middleware interference)
    r.HandleFunc("/ws", handleWebSocket)
    
    // Static files with compression
    r.Route("/static", func(r chi.Router) {
        r.Use(middleware.Compress(5))
        workDir, _ := os.Getwd()
        filesDir := http.Dir(filepath.Join(workDir, "web", "static"))
        FileServer(r, "/", filesDir)
    })
    
    // API routes with JSON content type
    r.Route("/api", func(r chi.Router) {
        r.Use(middleware.SetHeader("Content-Type", "application/json"))
        
        // License management
        r.Route("/license", func(r chi.Router) {
            r.Get("/check", handleLicenseCheck)
            r.Get("/status", handleLicenseCheck) // Alias
            r.Post("/activate", handleLicenseActivate)
        })
        
        // Pipeline control
        r.Route("/pipeline", func(r chi.Router) {
            r.Post("/start", handlePipelineStart)
            r.Post("/stop", handlePipelineStop)
            r.Get("/status", handlePipelineStatus)
        })
        
        // Data endpoints
        r.Route("/data", func(r chi.Router) {
            r.Get("/reports", handleDataReports)
            r.Get("/tickers", handleDataTickers)
            r.Get("/indices", handleDataIndices)
        })
    })
    
    // HTML pages
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/license", http.StatusTemporaryRedirect)
    })
    r.Get("/license", serveLicensePage)
    r.Get("/app", serveAppPage)
    
    return r
}
```

### Phase 4: Middleware Implementation
Replace problematic custom middleware with Chi's built-in solutions:

```go
// Chi-compatible logging middleware
func ChiLoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip WebSocket upgrades
        if r.Header.Get("Upgrade") == "websocket" {
            next.ServeHTTP(w, r)
            return
        }
        
        start := time.Now()
        requestID := middleware.GetReqID(r.Context())
        clientIP := middleware.GetReqIP(r)
        
        // Use Chi's WrapWriter for proper response capture
        ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
        
        logger.InfoCategory(common.CategoryHTTP, 
            "Request: %s %s [%s] from %s", 
            r.Method, r.URL.Path, requestID, clientIP)
        
        next.ServeHTTP(ww, r)
        
        duration := time.Since(start)
        logger.InfoCategory(common.CategoryHTTP,
            "Response: %s %s [%s] - %d (%v)",
            r.Method, r.URL.Path, requestID, ww.Status(), duration)
    })
}
```

### Phase 5: Static File Serving
Implement efficient static file serving with Chi:

```go
// FileServer conveniently sets up a http.FileServer handler
func FileServer(r chi.Router, path string, root http.FileSystem) {
    if strings.ContainsAny(path, "{}*") {
        panic("FileServer does not permit any URL parameters.")
    }

    if path != "/" && path[len(path)-1] != '/' {
        r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
        path += "/"
    }
    path += "*"

    r.Get(path, func(w http.ResponseWriter, r *http.Request) {
        rctx := chi.RouteContext(r.Context())
        pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
        fs := http.StripPrefix(pathPrefix, http.FileServer(root))
        fs.ServeHTTP(w, r)
    })
}
```

## Testing Strategy

### 1. Functional Testing
- ✅ All HTTP endpoints return 200 (not 206)
- ✅ HTML pages load completely without hanging
- ✅ WebSocket connections establish properly
- ✅ License validation flow works end-to-end
- ✅ All API endpoints respond correctly
- ✅ Static assets load with proper caching

### 2. Performance Testing
- Benchmark request handling before/after migration
- Compare response times and memory usage
- Test concurrent request handling
- Validate middleware overhead

### 3. Integration Testing
- Full pipeline execution with WebSocket updates
- Client logging endpoint functionality
- Error handling and recovery
- Request ID tracing through logs

## Risk Mitigation

### 1. Development Safety
- **Feature Branch**: `feature/chi-framework-migration`
- **Incremental Commits**: Each component migrated separately
- **Rollback Plan**: Keep gorilla/mux imports commented for quick revert

### 2. Testing Approach
- **Local Testing**: Comprehensive local validation before deployment
- **Gradual Migration**: Routes migrated in logical groups
- **Backward Compatibility**: Ensure all existing endpoints work

### 3. Monitoring
- **Request Tracing**: Chi's RequestID for better debugging
- **Performance Metrics**: Before/after comparison
- **Error Tracking**: Structured error logging

## Success Criteria

### Technical Metrics
- [ ] All HTTP responses return proper status codes (200, not 206)
- [ ] HTML page loading time < 500ms
- [ ] WebSocket connection success rate 100%
- [ ] Zero middleware-related errors in logs
- [ ] Request processing time improvement > 10%

### Functional Metrics
- [ ] License page loads and validates correctly
- [ ] Main application navigates properly after license check
- [ ] All pipeline operations work through web interface
- [ ] Data endpoints return proper JSON responses
- [ ] Static assets load with appropriate caching headers

### Quality Metrics
- [ ] Code coverage maintained > 80%
- [ ] All existing tests pass
- [ ] New middleware tests added
- [ ] Documentation updated completely
- [ ] Zero security vulnerabilities introduced

## Timeline

| Phase | Duration | Tasks |
|-------|----------|-------|
| Documentation | 1 hour | Create docs, update roadmap, UAT |
| Dependencies | 15 min | Update go.mod, run go mod tidy |
| Router Migration | 2 hours | Replace mux with Chi, organize routes |
| Middleware | 1 hour | Implement Chi-compatible middleware |
| Testing | 2 hours | Functional, performance, integration tests |
| Documentation Updates | 30 min | Update guides and API docs |
| **Total** | **~6.5 hours** | Complete migration |

## Post-Migration Benefits

### 1. Immediate Fixes
- ✅ Resolves HTTP 206 partial content issue
- ✅ Eliminates response hanging problems
- ✅ Fixes WebSocket middleware interference
- ✅ Provides proper request tracing

### 2. Long-term Improvements
- 🔄 Better performance with lightweight routing
- 🔄 Improved debugging with request IDs
- 🔄 More maintainable middleware stack
- 🔄 Google-aligned Go practices
- 🔄 Easier testing and development

### 3. Architecture Benefits
- 🔄 Cleaner separation of concerns
- 🔄 Composable middleware design
- 🔄 Better error handling and recovery
- 🔄 Standard library compatibility

## Rollback Plan

If issues arise during migration:

1. **Immediate Rollback**: Uncomment gorilla/mux imports
2. **Revert Commits**: Use Git to revert to previous working state
3. **Emergency Fix**: Disable LoggingMiddleware temporarily
4. **Assessment**: Analyze what went wrong and plan fixes

## Next Steps

1. ✅ Complete this documentation
2. ✅ Push documentation to Git
3. 🔄 Create feature branch
4. 🔄 Update dependencies
5. 🔄 Implement Chi router
6. 🔄 Test thoroughly
7. 🔄 Create pull request

---

**Author**: Claude AI Assistant  
**Date**: 2025-07-23  
**Version**: 1.0  
**Epic**: INFRA-019 Chi Framework Migration  
**Related Issues**: HTTP 206 partial content, response hanging, middleware problems