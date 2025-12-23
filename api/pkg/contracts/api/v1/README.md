# API v1 Contracts

Version 1 API contracts for HTTP requests and responses.

## Structure

- **requests.go** - All API request types including operations requests
- **responses.go** - All API response types including operations responses
- **errors.go** - RFC 7807 Problem Details implementation
- **pagination.go** - Pagination request/response structures

## Request Types

### Operations Requests
```go
type StartOperationRequest struct {
    Type      string    `json:"type" validate:"required,oneof=scraping processing export"`
    Mode      string    `json:"mode" validate:"required,oneof=daily accumulative"`
    DateRange DateRange `json:"date_range,omitempty"`
}

type CancelOperationRequest struct {
    Reason string `json:"reason,omitempty"`
}
```

### License Requests
```go
type ActivateLicenseRequest struct {
    Key string `json:"key" validate:"required,len=20"`
}

type TransferLicenseRequest struct {
    NewHardwareID string `json:"new_hardware_id" validate:"required"`
}
```

## Response Types

### Operations Responses
```go
type OperationResponse struct {
    ID          string          `json:"id"`
    Type        string          `json:"type"`
    Status      string          `json:"status"`
    Progress    float64         `json:"progress"`
    CurrentStep *OperationStep  `json:"current_step,omitempty"`
    StartedAt   time.Time       `json:"started_at"`
    CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

type OperationListResponse struct {
    Operations []OperationSummary `json:"operations"`
    Total      int                `json:"total"`
}
```

### License Responses
```go
type LicenseStatusResponse struct {
    Valid     bool      `json:"valid"`
    ExpiresAt time.Time `json:"expires_at"`
    DaysLeft  int       `json:"days_left"`
    Features  []string  `json:"features"`
}
```

## RFC 7807 Problem Details

All errors follow RFC 7807 standard:

```go
type Problem struct {
    Type     string                 `json:"type"`
    Title    string                 `json:"title"`
    Status   int                    `json:"status"`
    Detail   string                 `json:"detail,omitempty"`
    Instance string                 `json:"instance,omitempty"`
    Trace    string                 `json:"trace_id,omitempty"`
    Extensions map[string]interface{} `json:",inline"`
}
```

### Common Problem Types
- `validation_failed` - Request validation errors
- `not_found` - Resource not found
- `conflict` - Resource conflict
- `rate_limit_exceeded` - Too many requests
- `internal_error` - Server errors

## Validation

All requests include validation tags:
```go
type Request struct {
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"min=0,max=150"`
    Status   string `json:"status" validate:"oneof=active inactive"`
}
```

## Pagination

Standard pagination for list endpoints:
```go
type PaginationRequest struct {
    Page     int    `json:"page" validate:"min=1"`
    PageSize int    `json:"page_size" validate:"min=1,max=100"`
    Sort     string `json:"sort,omitempty"`
    Order    string `json:"order,omitempty" validate:"omitempty,oneof=asc desc"`
}

type PaginationResponse struct {
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}
```

## Versioning

When breaking changes are needed:
1. Create a new version directory (e.g., v2)
2. Copy and modify contracts
3. Support both versions during migration period
4. Document migration path
5. Add deprecation notices to old version

## Change Log
- 2025-07-30: Added operations request/response types for Phase 4
- 2025-07-30: Enhanced error types with RFC 7807 extensions
- 2025-07-30: Added validation examples and common problem types
- 2025-07-30: Updated documentation with request/response examples