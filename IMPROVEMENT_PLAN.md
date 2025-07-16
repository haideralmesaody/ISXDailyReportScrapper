# Minimal Fix Plan

The following steps outline how to address the issues found in the project with minimal code changes. Apply each change carefully and adapt any affected files.

## 1. Correct time measurements

### Scraper (`main.go`)
- Store the start time before each page is processed, e.g. `start := time.Now()`.
- After the page actions complete, use `time.Since(start)` instead of `time.Since(time.Now())`.
- No other files depend on this variable, so no further changes should be necessary.

### Web server (`cmd/web/main.go`)
- Add a global variable `startTime` initialised at the beginning of `main()`.
- In `handleStatus`, replace `time.Since(time.Now())` with `time.Since(startTime).String()`.
- Ensure other files do not reference `startTime`; if additional status handlers are added, reuse the same variable.

## 2. Replace deprecated `io/ioutil` usage
- Search for `ioutil` calls in `cmd/process/main.go` and replace them with the modern `os` or `io` equivalents (e.g. `os.ReadFile` and `os.WriteFile`).
- Confirm behaviour through existing tests to ensure no regressions.

## 3. Reduce debug output
- The parser and scraper emit many `fmt.Printf` statements. Wrap these calls behind a debug flag or remove them once functionality is verified.
- Check that tests (`internal/parser/parser_test.go`) continue to pass after cleaning output.

## 4. Fix parser header detection
- The parser test fails because `ParseFile` cannot find a header row. Review the header‑matching logic and ensure it recognises minimal workbooks.
- Update `internal/parser/parser.go` accordingly and extend `parser_test.go` if needed.

## 5. Clean repository
- Remove generated CSV `indexes` from version control and keep it excluded via `.gitignore`.
- Delete example license files under `license-generator-app/` (pattern `licenses_*.txt`).
- Remove built web assets under `release/web/` if they can be generated during the build.
- After deleting the files, verify that `.gitignore` already covers these paths.

## 6. Testing
- Run `go vet ./...` and `go test ./...` to ensure all packages build and tests succeed after the changes.
- If any new dependencies are introduced, update `go.mod` and `go.sum` with `go mod tidy`.

Following this plan will resolve the discovered issues while keeping modifications minimal and contained.
