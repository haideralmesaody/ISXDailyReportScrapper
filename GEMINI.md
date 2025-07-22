# Gemini Project Configuration

## Project Overview

This project, "ISXDailyReportsScrapper," is a Go application designed to scrape daily reports from the Iraqi Stock Exchange (ISX). It includes a command-line interface (CLI), a web interface, and various tools for data processing and license generation.

## Technical Details

- **Language:** Go
- **Build System:** Go modules (`go.mod`, `go.sum`) and custom `.bat` scripts.
- **Main Application:** The main entry point is `main.go`.
- **Web Interface:** Located in the `web/` directory, with its own `main.go` in `cmd/web/`.
- **Testing:** Tests are written using the standard Go `testing` package (e.g., `main_test.go`, `internal/parser/parser_test.go`). Use `go test ./...` to run all tests.

## Development Guidelines

- **Dependencies:** Manage dependencies using `go get` and `go mod tidy`.
- **Code Style:** Adhere to standard Go formatting (`gofmt`).
- **Commits:** Follow the conventions outlined in `GITHUB_COMMIT_GUIDE.md`.
- **Building:** Use the provided `build-*.bat` scripts for creating releases and installers. For general development, use `go build`.
