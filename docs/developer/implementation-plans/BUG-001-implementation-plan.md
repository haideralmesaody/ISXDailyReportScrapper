# Comprehensive Plan for BUG-001: Create Bug Report Backend API

## Overview
BUG-001 involves creating a bug report backend API that integrates with Google Sheets for storage and Google Drive for image uploads. This will follow the existing license system patterns and implement comprehensive test-driven development.

## Current Situation
- The web-licensed application is missing its main function and route initialization
- We need to create the main.go file for web-licensed
- The bug report system will follow the existing pattern from the license manager

## Implementation Plan

### Phase 1: Test Planning & Setup

#### 1.1 Create Test Plan Document
**File**: `docs/developer/test-plans/BUG-001-test-plan.md`
- Define all test scenarios for bug report API
- Security test cases (injection, XSS, path traversal)
- Integration test cases with Google APIs
- Performance benchmarks

#### 1.2 Create Test Files First (TDD)
1. **Unit Tests**: `internal/bugreport/manager_test.go`
   - Mock Google Sheets API
   - Mock Google Drive API
   - Test all validation scenarios
   - Test error handling

2. **Integration Tests**: `internal/bugreport/integration_test.go`
   - Test with real Google APIs (using test account)
   - Test file upload scenarios
   - Test data persistence

3. **Security Tests**: `internal/bugreport/security_test.go`
   - Input validation tests
   - File upload security tests
   - API authentication tests

### Phase 2: Create Missing Web Application Main

#### 2.1 Create `cmd/web-licensed/main.go`
```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
    
    "github.com/gorilla/mux"
    "isxcli/internal/license"
    "isxcli/internal/pipeline"
    "isxcli/internal/websocket"
    "isxcli/internal/bugreport"
)

var (
    wsHub           *websocket.Hub
    licenseManager  *license.Manager
    pipelineManager *pipeline.Manager
    bugManager      *bugreport.Manager
    executableDir   string
)

func main() {
    // Initialize components
    setupDirectories()
    initializeManagers()
    
    // Setup routes
    router := setupRoutes()
    
    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Starting web server on port %s", port)
    if err := http.ListenAndServe(":"+port, router); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}

func setupRoutes() *mux.Router {
    router := mux.NewRouter()
    
    // Static files
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", 
        http.FileServer(http.Dir(filepath.Join(executableDir, "web/static")))))
    
    // API routes
    api := router.PathPrefix("/api").Subrouter()
    
    // Existing routes
    api.HandleFunc("/start-scrape", handleStartScrape).Methods("POST")
    api.HandleFunc("/start-process", handleStartProcess).Methods("POST")
    api.HandleFunc("/start-index", handleStartIndexExtraction).Methods("POST")
    api.HandleFunc("/download/{filename}", handleDownloadFile).Methods("GET")
    
    // Bug report routes
    api.HandleFunc("/bug-report", handleSubmitBugReport).Methods("POST")
    api.HandleFunc("/bug-report/upload", handleImageUpload).Methods("POST")
    api.HandleFunc("/bug-reports", handleListBugReports).Methods("GET")
    
    // WebSocket
    router.HandleFunc("/ws", handleWebSocket)
    
    // Serve index.html for all other routes
    router.PathPrefix("/").HandlerFunc(serveIndex)
    
    return router
}
```

### Phase 3: Implement Bug Report Package

#### 3.1 Create Package Structure
```
internal/bugreport/
├── manager.go          # Main bug report manager
├── google_sheets.go    # Google Sheets integration
├── google_drive.go     # Google Drive integration  
├── validation.go       # Input validation
├── types.go           # Data structures
└── doc.go             # Package documentation
```

#### 3.2 Core Components

**types.go** - Data structures:
```go
type BugReport struct {
    ID          string    `json:"id"`
    Timestamp   time.Time `json:"timestamp"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    ImageURL    string    `json:"image_url,omitempty"`
    Status      string    `json:"status"`
    Priority    string    `json:"priority"`
    Version     string    `json:"version"`
    UserAgent   string    `json:"user_agent"`
    SessionID   string    `json:"session_id"`
}

type Manager struct {
    sheetsService *sheets.Service
    driveService  *drive.Service
    config        Config
    logger        *common.Logger
    validator     *Validator
}
```

**manager.go** - Core functionality:
- NewManager() - Initialize with Google services
- SubmitBugReport() - Main submission endpoint
- ValidateReport() - Input validation
- SaveToSheets() - Google Sheets integration
- UploadImage() - Google Drive integration

### Phase 4: API Endpoints Implementation

#### 4.1 Bug Report Submission
```go
func handleSubmitBugReport(w http.ResponseWriter, r *http.Request) {
    // 1. Parse multipart form
    // 2. Validate inputs
    // 3. Upload image if present
    // 4. Save to Google Sheets
    // 5. Send WebSocket notification
    // 6. Return response
}
```

#### 4.2 Image Upload Handler
```go
func handleImageUpload(w http.ResponseWriter, r *http.Request) {
    // 1. Parse multipart form
    // 2. Validate file type and size
    // 3. Compress if needed
    // 4. Upload to Google Drive
    // 5. Return shareable link
}
```

### Phase 5: Google Integration

#### 5.1 Google Sheets Setup
- Create service account credentials
- Set up Google Sheets API client
- Implement batch writing for performance
- Add retry logic for reliability

#### 5.2 Google Drive Setup
- Configure Drive API client
- Set up folder structure
- Implement file upload with progress
- Generate shareable links

### Phase 6: Security Implementation

#### 6.1 Input Validation
- Title: Max 100 chars, sanitize HTML
- Description: Max 1000 chars, sanitize HTML
- Image: Max 5MB, validate MIME types
- Rate limiting per session

#### 6.2 File Upload Security
- Validate file extensions (.jpg, .jpeg, .png, .gif)
- Check file headers (magic bytes)
- Scan for malicious content
- Generate unique filenames

### Phase 7: WebSocket Integration

#### 7.1 Progress Messages
```go
type BugReportProgress struct {
    Type     string  `json:"type"`     // "uploading", "saving", "complete"
    Progress float64 `json:"progress"` // 0-100
    Message  string  `json:"message"`
    Error    string  `json:"error,omitempty"`
}
```

#### 7.2 Real-time Updates
- Send progress during image upload
- Notify on successful submission
- Handle error notifications

### Phase 8: Testing Implementation

#### 8.1 Unit Tests (100% coverage required)
- Test all manager methods
- Test validation logic
- Test error scenarios
- Mock all external dependencies

#### 8.2 Integration Tests
- Test Google Sheets integration
- Test Google Drive uploads
- Test complete submission flow
- Test with various file sizes

#### 8.3 Security Tests
- SQL injection attempts
- XSS in title/description
- Path traversal in filenames
- Large file DoS attempts

#### 8.4 Performance Tests
- Benchmark image compression
- Test concurrent submissions
- Memory usage monitoring
- Response time requirements

### Phase 9: Documentation

#### 9.1 API Documentation
- OpenAPI specification
- Example requests/responses
- Error codes and meanings

#### 9.2 Setup Guide
- Google Cloud setup instructions
- Service account configuration
- Environment variables required

## File Structure Summary

```
dev/
├── cmd/
│   └── web-licensed/
│       ├── main.go (NEW)
│       ├── routes.go (NEW)
│       ├── handlers.go (NEW)
│       └── bug_handlers.go (NEW)
├── internal/
│   └── bugreport/ (NEW)
│       ├── manager.go
│       ├── manager_test.go
│       ├── google_sheets.go
│       ├── google_sheets_test.go
│       ├── google_drive.go
│       ├── google_drive_test.go
│       ├── validation.go
│       ├── validation_test.go
│       ├── security_test.go
│       ├── integration_test.go
│       ├── types.go
│       └── doc.go
└── docs/
    └── developer/
        └── test-plans/
            └── BUG-001-test-plan.md (NEW)
```

## Implementation Order

1. Create test plan document
2. Write all test files (TDD approach)
3. Create main.go for web-licensed
4. Implement bugreport package structure
5. Implement core manager functionality
6. Add Google Sheets integration
7. Add Google Drive integration
8. Implement API handlers
9. Add WebSocket notifications
10. Run all tests and ensure 100% coverage
11. Update documentation

## Success Criteria

- All tests pass with 100% coverage for new code
- Security tests validate all inputs
- Integration tests work with Google APIs
- Performance benchmarks meet requirements
- WebSocket notifications work correctly
- Complete API documentation
- No regression in existing functionality

## Test Checklist (from TEST_CHECKLIST_TEMPLATE.md)

### Pre-Development
- [ ] Test plan document created
- [ ] All test scenarios identified
- [ ] Security considerations documented
- [ ] Test files created before implementation

### Unit Tests (100% Coverage)
- [ ] Manager methods tested
- [ ] Validation logic tested
- [ ] Error handling tested
- [ ] All edge cases covered
- [ ] Mocks for external dependencies

### Integration Tests
- [ ] Google Sheets API integration
- [ ] Google Drive API integration
- [ ] Complete submission flow
- [ ] File upload scenarios

### Security Tests
- [ ] Input validation (SQL injection, XSS)
- [ ] File upload security
- [ ] Path traversal prevention
- [ ] Rate limiting

### Performance Tests
- [ ] Image compression benchmarks
- [ ] Concurrent submission handling
- [ ] Memory usage within limits
- [ ] Response time < 2 seconds

### End-to-End Tests
- [ ] Complete bug report submission
- [ ] Image upload and preview
- [ ] Error handling UI
- [ ] WebSocket notifications

## Notes

- The web-licensed application needs its main.go file created first
- Follow the existing license manager pattern for Google Sheets integration
- Use the embedded service account credentials approach
- Ensure all WebSocket messages follow the existing format
- Rate limiting should be implemented to prevent abuse

This comprehensive plan ensures BUG-001 is implemented following test-driven development with full security, performance, and integration testing.