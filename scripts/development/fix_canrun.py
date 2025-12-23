#!/usr/bin/env python3
import re

# Read the stages.go file
with open('api/internal/operations/stages.go', 'r') as f:
    content = f.read()

# Define the old CanRun function
old_function = '''// CanRun checks if CSV files are available for index extraction
func (i *IndicesStage) CanRun(manifest *PipelineManifest) bool {
	// Check if CSV files are available (processed data from Processing stage)
	if data, exists := manifest.GetData("csv_files"); exists {
		return data.FileCount >= 1
	}
	// Also check the actual reports directory for CSV files
	reportsDir := filepath.Join(i.executableDir, "data", "reports")
	files, _ := filepath.Glob(filepath.Join(reportsDir, "*.csv"))
	return len(files) > 0
}'''

# Define the new CanRun function
new_function = '''// CanRun checks if CSV files are available for index extraction
func (i *IndicesStage) CanRun(manifest *PipelineManifest) bool {
	// Check if CSV files are available (processed data from Processing stage)
	if data, exists := manifest.GetData("csv_files"); exists {
		return data.FileCount >= 1
	}

	// Also check the actual reports directory for CSV files recursively
	// This handles the case where processing stage outputs to subdirectories
	reportsDir := filepath.Join(i.executableDir, "data", "reports")
	if _, err := os.Stat(reportsDir); err == nil {
		// Use recursive search to find CSV files in subdirectories
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
	}

	// Fallback to top-level glob (original behavior)
	files, _ := filepath.Glob(filepath.Join(reportsDir, "*.csv"))
	return len(files) > 0
}'''

# Replace the function
new_content = content.replace(old_function, new_function)

# Write back to file
with open('api/internal/operations/stages.go', 'w') as f:
    f.write(new_content)

print("Successfully updated IndicesStage.CanRun() method")