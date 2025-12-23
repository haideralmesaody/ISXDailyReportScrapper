package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	// Read the file
	content, err := ioutil.ReadFile("api/internal/operations/stages.go")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	contentStr := string(content)

	// Find and replace the CanRun function
	oldFunc := `// Also check the actual reports directory for CSV files
	reportsDir := filepath.Join(i.executableDir, "data", "reports")
	files, _ := filepath.Glob(filepath.Join(reportsDir, "*.csv"))
	return len(files) > 0`

	newFunc := `// Also check the actual reports directory for CSV files recursively
	// This handles the case where processing stage outputs to subdirectories
	reportsDir := filepath.Join(i.executableDir, "data", "reports")
	var csvFiles []string
	err := filepath.Walk(reportsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file has .csv extension
		if strings.EqualFold(filepath.Ext(info.Name()), ".csv") {
			csvFiles = append(csvFiles, path)
		}

		return nil
	})

	if err == nil {
		return len(csvFiles) > 0
	}

	// Fallback to top-level glob (original behavior)
	files, _ := filepath.Glob(filepath.Join(reportsDir, "*.csv"))
	return len(files) > 0`

	// Replace the function
	newContent := strings.ReplaceAll(contentStr, oldFunc, newFunc)

	// Write back to file
	err = ioutil.WriteFile("api/internal/operations/stages.go", []byte(newContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully updated IndicesStage.CanRun() method")
}