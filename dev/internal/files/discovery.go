package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileInfo represents information about a discovered file
type FileInfo struct {
	Path     string
	Name     string
	Size     int64
	ModTime  time.Time
	IsDir    bool
}

// Discovery provides file discovery operations
type Discovery struct {
	basePath string
}

// NewDiscovery creates a new file discovery instance
func NewDiscovery(basePath string) *Discovery {
	return &Discovery{basePath: basePath}
}

// FindExcelFiles finds all Excel files in the specified directory
func (d *Discovery) FindExcelFiles(dir string) ([]FileInfo, error) {
	// If dir is already absolute, use it directly
	fullPath := dir
	if !filepath.IsAbs(dir) {
		fullPath = filepath.Join(d.basePath, dir)
	}
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}

	var files []FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".xlsx") || 
		   strings.HasSuffix(strings.ToLower(name), ".xls") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			files = append(files, FileInfo{
				Path:    filepath.Join(fullPath, name),
				Name:    name,
				Size:    info.Size(),
				ModTime: info.ModTime(),
				IsDir:   false,
			})
		}
	}
	
	// Sort by modification time (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	})
	
	return files, nil
}

// FindCSVFiles finds all CSV files in the specified directory
func (d *Discovery) FindCSVFiles(dir string) ([]FileInfo, error) {
	// If dir is already absolute, use it directly
	fullPath := dir
	if !filepath.IsAbs(dir) {
		fullPath = filepath.Join(d.basePath, dir)
	}
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}

	var files []FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".csv") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			files = append(files, FileInfo{
				Path:    filepath.Join(fullPath, name),
				Name:    name,
				Size:    info.Size(),
				ModTime: info.ModTime(),
				IsDir:   false,
			})
		}
	}
	
	return files, nil
}

// FindFilesByPattern finds files matching a glob pattern
func (d *Discovery) FindFilesByPattern(dir string, pattern string) ([]FileInfo, error) {
	// If dir is already absolute, use it directly
	fullPath := dir
	if !filepath.IsAbs(dir) {
		fullPath = filepath.Join(d.basePath, dir)
	}
	searchPattern := filepath.Join(fullPath, pattern)
	
	matches, err := filepath.Glob(searchPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %s: %w", pattern, err)
	}
	
	var files []FileInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		
		if !info.IsDir() {
			files = append(files, FileInfo{
				Path:    match,
				Name:    filepath.Base(match),
				Size:    info.Size(),
				ModTime: info.ModTime(),
				IsDir:   false,
			})
		}
	}
	
	return files, nil
}

// FindDailyCSVFiles finds daily CSV files (isx_daily_YYYY_MM_DD.csv)
func (d *Discovery) FindDailyCSVFiles(dir string) (map[string]FileInfo, error) {
	files, err := d.FindCSVFiles(dir)
	if err != nil {
		return nil, err
	}
	
	dailyFiles := make(map[string]FileInfo)
	for _, file := range files {
		if strings.HasPrefix(file.Name, "isx_daily_") && strings.HasSuffix(file.Name, ".csv") {
			// Extract date from filename: isx_daily_YYYY_MM_DD.csv
			dateStr := strings.TrimPrefix(file.Name, "isx_daily_")
			dateStr = strings.TrimSuffix(dateStr, ".csv")
			dailyFiles[dateStr] = file
		}
	}
	
	return dailyFiles, nil
}

// ListDirectories lists all subdirectories in the specified directory
func (d *Discovery) ListDirectories(dir string) ([]FileInfo, error) {
	// If dir is already absolute, use it directly
	fullPath := dir
	if !filepath.IsAbs(dir) {
		fullPath = filepath.Join(d.basePath, dir)
	}
	
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}

	var dirs []FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			dirs = append(dirs, FileInfo{
				Path:    filepath.Join(fullPath, entry.Name()),
				Name:    entry.Name(),
				Size:    0,
				ModTime: info.ModTime(),
				IsDir:   true,
			})
		}
	}
	
	return dirs, nil
}

// GetLatestFile returns the most recently modified file from a list
func GetLatestFile(files []FileInfo) (FileInfo, bool) {
	if len(files) == 0 {
		return FileInfo{}, false
	}
	
	latest := files[0]
	for _, file := range files[1:] {
		if file.ModTime.After(latest.ModTime) {
			latest = file
		}
	}
	
	return latest, true
}

// FilterFilesByDateRange filters files based on modification time
func FilterFilesByDateRange(files []FileInfo, startDate, endDate time.Time) []FileInfo {
	var filtered []FileInfo
	for _, file := range files {
		if file.ModTime.After(startDate) && file.ModTime.Before(endDate) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}