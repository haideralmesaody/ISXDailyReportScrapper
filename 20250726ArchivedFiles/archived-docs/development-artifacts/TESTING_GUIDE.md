# ISX Daily Reports Scrapper - Testing Guide

> This guide covers testing standards, setup, and best practices for the ISX Daily Reports Scrapper project.

## Table of Contents
- [Overview](#overview)
- [Testing Requirements](#testing-requirements)
- [Go Testing](#go-testing)
- [JavaScript Testing](#javascript-testing)
- [Running Tests](#running-tests)
- [Coverage Requirements](#coverage-requirements)
- [Best Practices](#best-practices)

## Overview

This project follows test-driven development practices with comprehensive test coverage for both backend (Go) and frontend (JavaScript) code.

### Testing Stack
- **Go Testing**: Built-in `testing` package with race detection
- **JavaScript Testing**: Jest with jsdom for unit tests
- **E2E Testing**: Playwright for end-to-end tests

## Testing Requirements

Per CLAUDE.md standards:
- Table-driven tests for Go code
- Race detector enabled for concurrent code
- ≥ 90% coverage for critical packages
- ≥ 80% coverage for other packages
- All tests must pass before merging

## Go Testing

### Setup

1. **Install Go 1.24+**
   ```bash
   go version  # Should show go1.24 or higher
   ```

2. **Enable Race Detector (Windows)**
   - Install MinGW-w64 GCC compiler
   - Set `CGO_ENABLED=1` environment variable
   - See `RACE_DETECTOR_SETUP.md` for detailed instructions

### Writing Go Tests

Follow the table-driven test pattern:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("Function() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Running Go Tests

```bash
# Run all tests with race detection
cd dev
go test -race ./...

# Run with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test -race -v ./internal/handlers

# Run specific test
go test -race -v -run TestHealthHandler ./internal/handlers
```

### Critical Go Packages (≥90% coverage required)
- `internal/pipeline`
- `internal/handlers`
- `internal/websocket`
- `internal/license`

## JavaScript Testing

### Setup

1. **Install Node.js dependencies**
   ```bash
   npm install
   ```

2. **Jest Configuration**
   - See `jest.config.js` for test configuration
   - See `jest.setup.js` for global mocks

### Writing JavaScript Tests

Follow Jest conventions:

```javascript
describe('ComponentName', () => {
    let component;

    beforeEach(() => {
        // Setup
        component = new Component();
    });

    afterEach(() => {
        // Cleanup
        jest.clearAllMocks();
    });

    describe('methodName', () => {
        test('should handle normal case', () => {
            const result = component.method('input');
            expect(result).toBe('expected');
        });

        test('should handle error case', () => {
            expect(() => component.method(null)).toThrow();
        });
    });
});
```

### Running JavaScript Tests

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run in watch mode
npm run test:watch

# Run specific test file
npm test logger.test.js
```

### Critical JavaScript Modules (≥80% coverage required)
- `core/logger.js`
- `core/websocket.js`
- `core/eventBus.js`
- `services/api.js`

## Running Tests

### Quick Test Commands

```bash
# Run all Go tests
cd dev && go test -race ./...

# Run all JavaScript tests
npm test

# Run all tests with coverage
cd dev && go test -race -coverprofile=coverage.out ./... && cd .. && npm run test:coverage

# Run E2E tests
npm run test:e2e
```

### Continuous Integration

Tests run automatically on:
- Pull request creation
- Push to main branch
- Pre-commit hooks (if configured)

## Coverage Requirements

### Go Coverage Targets

| Package Type | Minimum Coverage |
|-------------|------------------|
| Critical (pipeline, licensing, handlers) | ≥ 90% |
| Other packages | ≥ 80% |

### JavaScript Coverage Targets

| Module Type | Minimum Coverage |
|------------|------------------|
| Core modules | ≥ 80% |
| UI components | ≥ 70% |
| Utilities | ≥ 90% |

### Viewing Coverage Reports

```bash
# Go coverage
cd dev
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# JavaScript coverage
npm run test:coverage
# Open coverage/lcov-report/index.html
```

## Best Practices

### General Testing Principles

1. **Test Behavior, Not Implementation**
   - Focus on what the code does, not how it does it
   - Tests should survive refactoring

2. **Use Descriptive Test Names**
   - Test names should describe the scenario and expected outcome
   - Example: `TestWebSocketManager_Connect_WithInvalidURL_ReturnsError`

3. **Keep Tests Independent**
   - Each test should be able to run in isolation
   - Use setup/teardown functions appropriately

4. **Mock External Dependencies**
   - Database connections
   - File system operations
   - Network requests
   - Time-based operations

### Go-Specific Best Practices

1. **Use Table-Driven Tests**
   - Reduces code duplication
   - Makes it easy to add test cases
   - Improves test readability

2. **Always Enable Race Detector**
   - Use `-race` flag for all tests
   - Fix any race conditions immediately

3. **Test Error Cases**
   - Test both success and failure paths
   - Verify error messages are meaningful

4. **Use Test Helpers**
   ```go
   func setupTest(t *testing.T) (*Service, func()) {
       t.Helper()
       // Setup code
       return service, func() {
           // Cleanup code
       }
   }
   ```

### JavaScript-Specific Best Practices

1. **Mock Browser APIs**
   - WebSocket, fetch, localStorage
   - Use jest.fn() for function mocks

2. **Test Async Code Properly**
   ```javascript
   test('async operation', async () => {
       const result = await asyncFunction();
       expect(result).toBe('expected');
   });
   ```

3. **Use beforeEach/afterEach**
   - Set up clean state for each test
   - Clear all mocks between tests

4. **Test Event Handlers**
   ```javascript
   const handler = jest.fn();
   component.on('event', handler);
   component.emit('event', data);
   expect(handler).toHaveBeenCalledWith(data);
   ```

## Troubleshooting

### Common Go Test Issues

1. **Race detector not working**
   - Ensure CGO_ENABLED=1
   - Install GCC compiler (MinGW on Windows)
   - See RACE_DETECTOR_SETUP.md

2. **Import cycle errors**
   - Move test helpers to separate package
   - Use interfaces to break dependencies

3. **Flaky tests**
   - Remove time.Sleep calls
   - Use proper synchronization
   - Mock time-dependent operations

### Common JavaScript Test Issues

1. **Module not found errors**
   - Check jest.config.js moduleNameMapper
   - Ensure correct file paths

2. **Async test timeouts**
   - Increase timeout: `jest.setTimeout(10000)`
   - Check for unresolved promises

3. **DOM-related errors**
   - Ensure jsdom environment is configured
   - Mock browser-specific APIs

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [Playwright Documentation](https://playwright.dev/)
- [Project CLAUDE.md](../CLAUDE.md) - Coding standards
- [RACE_DETECTOR_SETUP.md](../archived-docs/RACE_DETECTOR_SETUP.md) - Race detector setup guide