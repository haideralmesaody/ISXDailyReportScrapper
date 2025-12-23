# Indices Stage Pipeline Stall Fix

## Root Cause Confirmed ✅

**Issue**: The indices stage is stuck in "Preparing" and gets skipped because:
- Processing stage outputs 377 CSV files in subdirectories (`combined/`, `daily/`, `summary/`, `ticker/`)
- Pipeline scanning pattern `"*.csv"` only checks top-level `data/reports/` directory
- Top-level scan finds **0 files**, causing `IndicesStage.CanRun()` to return `false`
- Stage gets skipped with "requirements not met"

## Evidence
```bash
# CSV files in subdirectories (where processing actually outputs)
$ find data/reports -name "*.csv" | wc -l
377

# CSV files at top level (what current pattern looks for)
$ ls data/reports/*.csv | wc -l
0
```

## Solution Implemented

### ✅ COMPLETED: Manifest Recursive Scanning
**File**: `api/internal/operations/manifest.go`
- Added `strings` import
- Updated `ScanDataDirectory()` to support `**/*.csv` recursive patterns
- Added `scanDirectoryRecursive()` method for recursive file discovery
- Maintains backward compatibility with existing `*.csv` patterns

### ⏳ PENDING: Processing Stage Pattern Update
**File**: `api/internal/operations/stages.go` (line ~1552)
**Change needed**:
```go
// BEFORE:
Pattern: "*.csv",

// AFTER:
Pattern: "**/*.csv", // Use recursive pattern to detect CSVs in subdirectories
```

### ⏳ PENDING: Indices Stage Fallback Logic
**File**: `api/internal/operations/stages.go` (lines ~1930-1940)
**Change needed**: Update `CanRun()` method to use recursive detection as fallback:
```go
// BEFORE:
func (i *IndicesStage) CanRun(manifest *PipelineManifest) bool {
    if data, exists := manifest.GetData("csv_files"); exists {
        return data.FileCount >= 1
    }
    reportsDir := filepath.Join(i.executableDir, "data", "reports")
    files, _ := filepath.Glob(filepath.Join(reportsDir, "*.csv"))
    return len(files) > 0
}

// AFTER:
func (i *IndicesStage) CanRun(manifest *PipelineManifest) bool {
    if data, exists := manifest.GetData("csv_files"); exists {
        return data.FileCount >= 1
    }

    // Recursive fallback for subdirectory detection
    reportsDir := filepath.Join(i.executableDir, "data", "reports")
    if _, err := os.Stat(reportsDir); err == nil {
        var csvFiles []string
        filepath.Walk(reportsDir, func(path string, info os.FileInfo, err error) error {
            if err != nil || info.IsDir() {
                return nil
            }
            if strings.EqualFold(filepath.Ext(info.Name()), ".csv") {
                csvFiles = append(csvFiles, path)
            }
            return nil
        })
        if len(csvFiles) > 0 {
            return true
        }
    }

    // Original fallback
    files, _ := filepath.Glob(filepath.Join(reportsDir, "*.csv"))
    return len(files) > 0
}
```

## Manual Fix Instructions

Since file locking prevented editing stages.go, apply these changes manually:

### Step 1: Update Processing Stage Pattern
1. Open `api/internal/operations/stages.go`
2. Find line ~1552 with `Pattern: "*.csv",`
3. Change to: `Pattern: "**/*.csv", // Use recursive pattern to detect CSVs in subdirectories`

### Step 2: Update Indices Stage CanRun Method
1. Find the `CanRun` method around line 1930
2. Replace the entire method with the recursive version shown above
3. Add necessary imports if not already present (`os`, `strings`)

## Expected Results After Fix

### Before Fix:
- Manifest shows: `csv_files.FileCount = 0`
- Indices stage gets skipped
- UI shows "Preparing" indefinitely
- Pipeline never completes

### After Fix:
- Manifest shows: `csv_files.FileCount = 377`
- Indices stage starts executing
- UI shows actual progress for indices stage
- Full pipeline completes successfully

## Testing the Fix

1. **Build the server**: `./build.bat`
2. **Check logs**: Look for indices stage starting (not being skipped)
3. **Verify manifest**: Check that `csv_files.FileCount > 0` after processing stage
4. **Test pipeline**: Run full pipeline and confirm indices stage executes

## Verification Commands

```bash
# Check if CSV files are found correctly after processing stage
grep "csv_files" dist/logs/app.log | tail -5

# Look for indices stage execution (not skipping)
grep "Indices.*started" dist/logs/app.log

# Verify indices stage is not skipped
grep "indices.*skip" dist/logs/app.log | wc -l  # Should be 0
```

## Technical Details

### New Manifest Scanning Logic
- Detects `**/*.csv` patterns
- Uses `filepath.Walk()` for recursive directory traversal
- Matches files by name pattern in all subdirectories
- Maintains compatibility with existing `*.csv` patterns

### Dependencies Updated
- `manifest.go`: Added `strings` import for pattern detection
- `stages.go`: Needs `os` and `strings` imports for recursive logic

### Performance Impact
- Minimal overhead for recursive scanning (377 files)
- Caching in manifest prevents repeated scanning
- Only used when `**/` pattern is detected

---

**Status**: ✅ Core fix implemented (manifest scanning), ⏳ Two small changes pending due to file locking
**Impact**: 🚀 Will resolve indices stage pipeline stall completely
**Risk**: 🟢 Low - backward compatible, only adds functionality