# UTF-8 BOM Handling Guide

## Overview
This guide documents how to properly handle UTF-8 Byte Order Mark (BOM) in CSV files throughout the ISX Daily Reports Scrapper application. The BOM (bytes 0xEF, 0xBB, 0xBF) can cause parsing issues if not handled correctly.

## What is UTF-8 BOM?

The UTF-8 BOM is a 3-byte sequence at the beginning of a file:
- Hex: `EF BB BF`
- Appears as: `﻿` (invisible character)
- Purpose: Indicates file encoding as UTF-8

## The Problem

When CSV files contain a BOM, it becomes part of the first field name:
- Expected header: `Date`
- With BOM: `﻿Date`

This causes:
- Column matching failures
- JSON parsing errors
- Data validation issues

## Solution: Always Strip BOM When Reading CSV

### Go Implementation

```go
// readCSVWithBOM reads a CSV file and automatically strips UTF-8 BOM if present
func readCSVWithBOM(filePath string) ([][]string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Read first 3 bytes to check for BOM
    buf := make([]byte, 3)
    n, err := file.Read(buf)
    if err != nil && err != io.EOF {
        return nil, err
    }

    // If not BOM, seek back to start
    if n < 3 || buf[0] != 0xEF || buf[1] != 0xBB || buf[2] != 0xBF {
        file.Seek(0, 0)
    }
    // If BOM present, we're already past it

    reader := csv.NewReader(file)
    return reader.ReadAll()
}
```

### Alternative: Using bufio.Reader

```go
func stripBOM(reader io.Reader) io.Reader {
    buffered := bufio.NewReader(reader)
    
    // Peek at first 3 bytes
    peek, err := buffered.Peek(3)
    if err == nil && len(peek) >= 3 {
        if peek[0] == 0xEF && peek[1] == 0xBB && peek[2] == 0xBF {
            // Discard BOM
            buffered.Discard(3)
        }
    }
    
    return buffered
}

// Usage
file, _ := os.Open("data.csv")
defer file.Close()

reader := csv.NewReader(stripBOM(file))
records, _ := reader.ReadAll()
```

## Where BOM Handling is Required

### 1. Ticker Summary Generation
File: `dev/internal/analytics/summary.go`
```go
// Always handle BOM when reading the combined CSV
bufferedReader := bufio.NewReader(file)
// Check and skip BOM...
reader := csv.NewReader(bufferedReader)
```

### 2. Data Processing
Any function that reads CSV files should handle BOM:
- Combined data CSV reader
- Individual ticker CSV readers
- Index data CSV readers

### 3. Web API Endpoints
When serving CSV data via API, ensure BOM is stripped before processing.

## Testing for BOM Issues

### Create a Test File with BOM
```bash
# Windows PowerShell
[byte[]]@(0xEF,0xBB,0xBF) + [System.Text.Encoding]::UTF8.GetBytes("Date,Value`ndata") | Set-Content test.csv -Encoding Byte

# Linux/Mac
printf '\xEF\xBB\xBF' > test.csv && echo "Date,Value" >> test.csv
```

### Detect BOM in Files
```go
func hasBOM(filePath string) (bool, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return false, err
    }
    defer file.Close()

    buf := make([]byte, 3)
    n, err := file.Read(buf)
    if err != nil && err != io.EOF {
        return false, err
    }

    return n == 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF, nil
}
```

## Best Practices

1. **Always assume CSV files might have BOM**
   - External tools often add BOM
   - Excel exports may include BOM

2. **Handle BOM at the lowest level**
   - Create utility functions for CSV reading
   - Don't duplicate BOM handling code

3. **Test with BOM files**
   - Include BOM test files in test suite
   - Verify all CSV readers handle BOM

4. **Document BOM handling**
   - Comment when BOM is being handled
   - Note in function documentation

## Common Symptoms of BOM Issues

1. **Column not found errors**:
   ```
   Error: column "Date" not found (found "﻿Date")
   ```

2. **JSON field name mismatches**:
   ```
   Cannot find field "date" in {"\ufeffdate": "2024-01-01"}
   ```

3. **First row parsing failures**:
   - Only the first row is affected
   - Subsequent rows parse correctly

## Code References

- BOM handling implementation: `dev/internal/analytics/summary.go:80-98`
- CSV reading utilities: Throughout codebase where CSV files are read

## Validation Checklist

Before committing code that reads CSV files:

- [ ] Does the code handle UTF-8 BOM?
- [ ] Is BOM handling tested?
- [ ] Are error messages clear if BOM causes issues?
- [ ] Is the BOM handling documented in comments?