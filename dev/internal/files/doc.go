// Package files provides file system operations and discovery utilities
// for the ISX Daily Reports Scrapper application.
//
// This package contains two main components:
//
// Discovery: Provides file discovery operations such as finding Excel files,
// CSV files, and files matching specific patterns. It also includes utilities
// for filtering files by date range and finding the latest file.
//
// Manager: Provides basic file management operations such as copying, moving,
// deleting files, and ensuring directories exist. All operations are relative
// to a base path to maintain portability.
//
// Example usage:
//
//	// Create a discovery instance
//	discovery := files.NewDiscovery("/path/to/base")
//	
//	// Find all Excel files
//	excelFiles, err := discovery.FindExcelFiles("downloads")
//	
//	// Create a manager instance
//	manager := files.NewManager("/path/to/base")
//	
//	// Check if file exists
//	if manager.FileExists("data/report.csv") {
//	    // Process file
//	}
package files