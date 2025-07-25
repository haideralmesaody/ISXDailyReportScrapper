package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"isxcli/internal/config"
)

// DataService provides data access functionality
type DataService struct {
	config *config.Config
	logger Logger
}

// NewDataService creates a new data service
func NewDataService(cfg *config.Config, logger Logger) (*DataService, error) {
	return &DataService{
		config: cfg,
		logger: logger,
	}, nil
}

// GetReports returns a list of available reports
func (ds *DataService) GetReports() ([]map[string]interface{}, error) {
	reportsDir := filepath.Join(ds.config.GetDataDir(), "reports")
	
	files, err := os.ReadDir(reportsDir)
	if err != nil {
		if os.IsNotExist(err) {
				return []map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("failed to read reports directory: %w", err)
	}

	var reports []map[string]interface{}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".csv") {
			info, err := file.Info()
			if err != nil {
				continue
			}

			reports = append(reports, map[string]interface{}{
				"name":     file.Name(),
				"size":     info.Size(),
				"modified": info.ModTime(),
			})
		}
	}

	// Sort by modification time (newest first)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i]["modified"].(time.Time).After(reports[j]["modified"].(time.Time))
	})

	return reports, nil
}

// GetTickers returns ticker information
func (ds *DataService) GetTickers() (interface{}, error) {
	tickerFile := filepath.Join(ds.config.GetDataDir(), "reports", "ticker_summary.json")
	
	data, err := os.ReadFile(tickerFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []interface{}{}, nil
		}
		return nil, fmt.Errorf("failed to read ticker summary: %w", err)
	}

	var tickerData interface{}
	if err := json.Unmarshal(data, &tickerData); err != nil {
		return nil, fmt.Errorf("failed to parse ticker summary: %w", err)
	}

	return tickerData, nil
}

// GetIndices returns market indices data
func (ds *DataService) GetIndices() (map[string]interface{}, error) {
	indicesFile := filepath.Join(ds.config.GetDataDir(), "reports", "indexes.csv")
	
	file, err := os.Open(indicesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{
				"dates": []string{},
				"isx60": []float64{},
				"isx15": []float64{},
			}, nil
		}
		return nil, fmt.Errorf("failed to open indices file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}
	
	// Validate header
	if len(header) < 2 || header[0] != "Date" || header[1] != "ISX60" {
		return nil, fmt.Errorf("invalid CSV header format")
	}
	
	var dates []string
	var isx60Values []float64
	var isx15Values []float64
	
	// Read data rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}
		
		if len(record) < 2 {
			continue // Skip invalid rows
		}
		
		// Parse date
		dates = append(dates, record[0])
		
		// Parse ISX60
		isx60, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			ds.logger.Error("Failed to parse ISX60 value", "value", record[1], "error", err)
			isx60 = 0
		}
		isx60Values = append(isx60Values, isx60)
		
		// Parse ISX15 if present
		if len(record) > 2 && record[2] != "" {
			isx15, err := strconv.ParseFloat(record[2], 64)
			if err != nil {
				ds.logger.Error("Failed to parse ISX15 value", "value", record[2], "error", err)
				isx15 = 0
			}
			isx15Values = append(isx15Values, isx15)
		} else {
			isx15Values = append(isx15Values, 0)
		}
	}
	
	return map[string]interface{}{
		"dates": dates,
		"isx60": isx60Values,
		"isx15": isx15Values,
	}, nil
}

// GetFiles returns file listings from different directories
func (ds *DataService) GetFiles() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"downloads":     []interface{}{},
		"reports":       []interface{}{},
		"csvFiles":      []interface{}{},
		"total_size":    int64(0),
		"last_modified": time.Time{},
	}

	// List downloaded Excel files
	if err := ds.listFiles("downloads", ".xlsx", result); err != nil {
		ds.logger.Error("Failed to list downloads", "error", err)
	}

	// List report files
	if err := ds.listFiles("reports", ".csv", result); err != nil {
		ds.logger.Error("Failed to list reports", "error", err)
	}

	return result, nil
}

// GetMarketMovers returns market movers data
func (ds *DataService) GetMarketMovers(period, limit, minVolume string) (map[string]interface{}, error) {
	// Default values
	if period == "" {
		period = "1d"
	}
	if limit == "" {
		limit = "10"
	}
	if minVolume == "" {
		minVolume = "0"
	}

	return map[string]interface{}{
		"gainers":    []interface{}{},
		"losers":     []interface{}{},
		"mostActive": []interface{}{},
		"period":     period,
		"updated":    time.Now().Format(time.RFC3339),
	}, nil
}

// GetTickerChart returns chart data for a specific ticker
func (ds *DataService) GetTickerChart(ticker string) (map[string]interface{}, error) {
	if ticker == "" {
		return nil, fmt.Errorf("ticker parameter required")
	}

	tickerFile := filepath.Join(ds.config.GetDataDir(), "reports", ticker+"_daily.csv")
	
	_, err := os.Stat(tickerFile)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{
				"ticker": ticker,
				"data":   []interface{}{},
			}, nil
		}
		return nil, fmt.Errorf("failed to check ticker file: %w", err)
	}

	// For now, return empty structure - implement CSV parsing later
	return map[string]interface{}{
		"ticker": ticker,
		"data":   []interface{}{},
	}, nil
}

// DownloadFile serves a file for download
func (ds *DataService) DownloadFile(w http.ResponseWriter, r *http.Request, fileType, filename string) error {
	var dir string
	switch fileType {
	case "downloads":
		dir = filepath.Join(ds.config.GetDataDir(), "downloads")
	case "reports":
		dir = filepath.Join(ds.config.GetDataDir(), "reports")
	default:
		return fmt.Errorf("invalid file type")
	}

	// Security check - ensure the file is within the expected directory
	filePath := filepath.Join(dir, filename)
	if !strings.HasPrefix(filePath, dir) {
		return fmt.Errorf("invalid file path")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	// Set headers for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Serve the file
	http.ServeFile(w, r, filePath)
	return nil
}

// listFiles lists files in a directory with filtering
func (ds *DataService) listFiles(dirName, extension string, result map[string]interface{}) error {
	dir := filepath.Join(ds.config.GetDataDir(), dirName)
	
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var fileList []map[string]interface{}
	var totalSize int64
	var lastModified time.Time

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), extension) {
			info, err := file.Info()
			if err != nil {
				continue
			}

			fileInfo := map[string]interface{}{
				"name":     file.Name(),
				"size":     info.Size(),
				"modified": info.ModTime().Format(time.RFC3339),
			}

			fileList = append(fileList, fileInfo)
			totalSize += info.Size()

			if info.ModTime().After(lastModified) {
				lastModified = info.ModTime()
			}
		}
	}

	// Sort files by modification time (newest first)
	sort.Slice(fileList, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, fileList[i]["modified"].(string))
		timeJ, _ := time.Parse(time.RFC3339, fileList[j]["modified"].(string))
		return timeI.After(timeJ)
	})

	// Update result based on directory
	switch dirName {
	case "downloads":
		result["downloads"] = fileList
	case "reports":
		// Separate ticker CSV files from other reports
		var reports []interface{}
		var csvFiles []interface{}

		for _, file := range fileList {
			name := file["name"].(string)
			if strings.Contains(name, "_trading_history.csv") || strings.Contains(name, "isx_daily_") {
				csvFiles = append(csvFiles, file)
			} else {
				reports = append(reports, file)
			}
		}

		result["reports"] = reports
		result["csvFiles"] = csvFiles
	}

	result["total_size"] = totalSize
	result["last_modified"] = lastModified

	return nil
}