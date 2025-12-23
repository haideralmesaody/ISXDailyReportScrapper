package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Paths contains all the application paths
// This is the single source of truth for ALL file paths in the application
type Paths struct {
	ExecutableDir string
	WebDir        string
	StaticDir     string
	DataDir       string
	DownloadsDir  string
	ReportsDir    string
	CacheDir      string
	LogsDir       string
	LicenseFile   string
	
	// Config files
	CredentialsFile   string
	SheetsConfigFile  string
	
	// Report subdirectories for organized structure (legacy support)
	DailyReportsDir     string
	TickerReportsDir    string
	LiquidityReportsDir string
	SummaryReportsDir   string
	CombinedReportsDir  string
	IndexesReportsDir   string
	
	// Well-known report files (simplified paths in output directory)
	IndexCSV          string
	TickerSummaryJSON string
	TickerSummaryCSV  string
	CombinedDataCSV   string
}

// GetPaths returns the application paths relative to the executable location
// All paths are ALWAYS relative to the executable directory, never the current working directory
func GetPaths() (*Paths, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}
	
	// Resolve symlinks to get the actual executable location
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable symlinks: %v", err)
	}
	
	// Get the directory containing the executable
	exeDir := filepath.Dir(exe)
	
	// Log the resolved executable directory for debugging
	if logger := slog.Default(); logger != nil {
		logger.Info("Resolved executable directory",
			slog.String("exe_path", exe),
			slog.String("exe_dir", exeDir))
	}
	
	// All paths are relative to the executable directory
	// This ensures the application works correctly whether run from dev/ or dist/
	// Directory structure:
	// dist/
	//   ├── license.dat
	//   ├── credentials.json
	//   ├── sheets-config.json
	//   ├── data/
	//   │   ├── downloads/     (Excel files from scraper)
	//   │   ├── reports/       (Generated CSV reports)
	//   │   └── cache/         (Temporary files)
	//   ├── logs/              (Application logs)
	//   └── web/               (Frontend assets)
	
	dataDir := filepath.Join(exeDir, "data")
	reportsDir := filepath.Join(dataDir, "reports")
	
	// Define report subdirectories (kept for legacy compatibility)
	dailyReportsDir := filepath.Join(reportsDir, "daily")
	tickerReportsDir := filepath.Join(reportsDir, "ticker")
	liquidityReportsDir := filepath.Join(reportsDir, "liquidity_reports")
	summaryReportsDir := filepath.Join(reportsDir, "summary")
	combinedReportsDir := filepath.Join(reportsDir, "combined")
	indexesReportsDir := filepath.Join(reportsDir, "indexes")
	
	paths := &Paths{
		ExecutableDir: exeDir,
		DataDir:       dataDir,
		WebDir:        filepath.Join(exeDir, "web"),
		StaticDir:     filepath.Join(exeDir, "web", "static"),
		DownloadsDir:  filepath.Join(dataDir, "downloads"),
		ReportsDir:    reportsDir,
		CacheDir:      filepath.Join(dataDir, "cache"),
		LogsDir:       filepath.Join(exeDir, "logs"),
		
		// Configuration files (root of executable directory)
		LicenseFile:      filepath.Join(exeDir, "license.dat"),
		CredentialsFile:  filepath.Join(exeDir, "credentials.json"),
		SheetsConfigFile: filepath.Join(exeDir, "sheets-config.json"),
		
		// Report subdirectories (legacy compatibility)
		DailyReportsDir:     dailyReportsDir,
		TickerReportsDir:    tickerReportsDir,
		LiquidityReportsDir: liquidityReportsDir,
		SummaryReportsDir:   summaryReportsDir,
		CombinedReportsDir:  combinedReportsDir,
		IndexesReportsDir:   indexesReportsDir,
		
		// Well-known report files (in proper subdirectories)
		IndexCSV:          filepath.Join(indexesReportsDir, "indexes.csv"),
		TickerSummaryJSON: filepath.Join(summaryReportsDir, "ticker_summary.json"),
		TickerSummaryCSV:  filepath.Join(summaryReportsDir, "ticker_summary.csv"),
		CombinedDataCSV:   filepath.Join(combinedReportsDir, "isx_combined_data.csv"),
	}
	
	return paths, nil
}

// EnsureDirectories creates all required directories if they don't exist
func (p *Paths) EnsureDirectories() error {
	// List of all directories to create
	// Note: Each process (processor, indexcsv, etc.) creates its own subdirectories
	// This only creates the base directories needed by all processes
	directories := []string{
		p.DataDir,
		p.DownloadsDir,
		p.ReportsDir,    // Base reports directory only
		p.CacheDir,
		p.LogsDir,
		p.WebDir,
		p.StaticDir,
		// Report subdirectories are created by their respective processes:
		// - processor.exe creates: combined/, daily/, ticker/
		// - indexcsv.exe creates: indexes/
		// - Other processes create their own directories as needed
	}
	
	// Log directory creation
	logger := slog.Default()
	
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
		
		// Log successful directory creation
		if logger != nil {
			logger.Debug("Ensured directory exists",
				slog.String("directory", dir))
		}
	}
	
	return nil
}

// GetRelativePath returns a path relative to the executable directory
func (p *Paths) GetRelativePath(subpath string) string {
	return filepath.Join(p.ExecutableDir, subpath)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetTickerSector determines the sector for a given ticker symbol
func GetTickerSector(ticker string) string {
	// Banking tickers
	banks := []string{"BBOB", "BMNS", "BNOI", "BCOI", "BIME", "BIIB", "BKUI", "BROI", 
		"BASH", "BEFI", "BGUC", "BIBI", "BIDB", "BINT", "BLAD", "BMFI", "BMUI", 
		"BNAI", "BSUC", "BTIB", "BTRI", "BTRU", "BUND", "BUOI", "BAIB", "BELF", "BCIH"}
	
	// Telecom tickers
	telecom := []string{"TASC", "TZNI"}
	
	// Industry tickers
	industry := []string{"IBSD", "IFCM", "IITC", "IMAP", "IMCM", "IMIB", "INCP", 
		"IRMC", "ITLI", "IELI", "IHFI", "IHLI", "IICM", "IIDP", "IIEW", "IKHC", 
		"IKLV", "IMCI", "IMOS", "IBPM"}
	
	// Check ticker against categories
	tickerUpper := strings.ToUpper(ticker)
	
	for _, bank := range banks {
		if tickerUpper == bank {
			return "banks"
		}
	}
	
	for _, tel := range telecom {
		if tickerUpper == tel {
			return "telecom"
		}
	}
	
	for _, ind := range industry {
		if tickerUpper == ind {
			return "industry"
		}
	}
	
	// Default to "other" for uncategorized tickers
	return "other"
}

// GetLicensePath returns the license file path
// This ONLY uses the executable directory path - no current working directory fallback
func GetLicensePath() (string, error) {
	paths, err := GetPaths()
	if err != nil {
		return "", fmt.Errorf("failed to get paths: %w", err)
	}
	
	logger := slog.Default()
	if logger != nil {
		// Get current working directory and absolute path for enhanced logging
		wd, _ := os.Getwd()
		absPath, _ := filepath.Abs(paths.LicenseFile)
		
		logger.Info("License path resolution - Complete Details",
			slog.Group("paths",
				slog.String("configured", paths.LicenseFile),
				slog.String("absolute", absPath),
				slog.String("executable_dir", paths.ExecutableDir),
			),
			slog.Group("environment",
				slog.String("working_dir", wd),
				slog.String("exe_path", paths.ExecutableDir),
			),
			slog.Group("status",
				slog.Bool("file_exists", FileExists(paths.LicenseFile)),
				slog.String("method", "executable-relative"),
			),
		)
	}
	
	// Always return the executable-relative path
	// This ensures consistency across all components
	return paths.LicenseFile, nil
}

// GetWebFilePath returns the path to a web file
func (p *Paths) GetWebFilePath(filename string) string {
	return filepath.Join(p.WebDir, filename)
}

// GetStaticFilePath returns the path to a static file
func (p *Paths) GetStaticFilePath(filename string) string {
	return filepath.Join(p.StaticDir, filename)
}

// GetDownloadPath returns the path for a downloaded file
func (p *Paths) GetDownloadPath(filename string) string {
	return filepath.Join(p.DownloadsDir, filename)
}

// GetReportPath returns the path for a report file
func (p *Paths) GetReportPath(filename string) string {
	return filepath.Join(p.ReportsDir, filename)
}

// GetLogPath returns the path for a log file
func (p *Paths) GetLogPath(filename string) string {
	return filepath.Join(p.LogsDir, filename)
}

// GetCachePath returns the path for a cache file
func (p *Paths) GetCachePath(filename string) string {
	return filepath.Join(p.CacheDir, filename)
}

// GetCredentialsPath returns the path for the Google Sheets credentials file
func (p *Paths) GetCredentialsPath() string {
	path := p.CredentialsFile
	logger := slog.Default()
	if logger != nil {
		logger.Debug("Credentials path resolved",
			slog.String("path", path),
			slog.Bool("exists", FileExists(path)))
	}
	return path
}

// GetSheetsConfigPath returns the path for the sheets configuration file
func (p *Paths) GetSheetsConfigPath() string {
	path := p.SheetsConfigFile
	logger := slog.Default()
	if logger != nil {
		logger.Debug("Sheets config path resolved",
			slog.String("path", path),
			slog.Bool("exists", FileExists(path)))
	}
	return path
}

// GetIndexCSVPath returns the path for the indexes.csv file
func (p *Paths) GetIndexCSVPath() string {
	return p.IndexCSV
}

// GetTickerSummaryJSONPath returns the path for the ticker_summary.json file
func (p *Paths) GetTickerSummaryJSONPath() string {
	return p.TickerSummaryJSON
}

// GetTickerSummaryCSVPath returns the path for the ticker_summary.csv file
func (p *Paths) GetTickerSummaryCSVPath() string {
	return p.TickerSummaryCSV
}

// GetCombinedDataCSVPath returns the path for the isx_combined_data.csv file
func (p *Paths) GetCombinedDataCSVPath() string {
	return p.CombinedDataCSV
}

// GetDailyCSVPath returns the path for a daily CSV file (e.g., isx_daily_20240115.csv)
func (p *Paths) GetDailyCSVPath(date time.Time) string {
	filename := fmt.Sprintf("isx_daily_%s.csv", date.Format("20060102"))
	return filepath.Join(p.ReportsDir, filename)
}

// GetTickerDailyCSVPath returns the path for a per-ticker daily CSV file (e.g., BBOB_daily.csv)
func (p *Paths) GetTickerDailyCSVPath(ticker string) string {
	filename := fmt.Sprintf("%s_daily.csv", ticker)
	return filepath.Join(p.ReportsDir, filename)
}

// GetExcelPath returns the path for a downloaded Excel file
func (p *Paths) GetExcelPath(filename string) string {
	return filepath.Join(p.DownloadsDir, filename)
}

// GetExcelPathForDate returns the expected path for an Excel file for a specific date
func (p *Paths) GetExcelPathForDate(date time.Time) string {
	// Expected format: "YYYY MM DD ISX Daily Report.xlsx"
	filename := fmt.Sprintf("%s ISX Daily Report.xlsx", date.Format("2006 01 02"))
	return filepath.Join(p.DownloadsDir, filename)
}

// LogPathResolution logs detailed path resolution information for debugging
func (p *Paths) LogPathResolution() {
	logger := slog.Default()
	if logger == nil {
		return
	}
	
	logger.Info("Path resolution summary",
		slog.Group("directories",
			slog.String("executable", p.ExecutableDir),
			slog.String("data", p.DataDir),
			slog.String("downloads", p.DownloadsDir),
			slog.String("reports", p.ReportsDir),
			slog.String("cache", p.CacheDir),
			slog.String("logs", p.LogsDir),
			slog.String("web", p.WebDir),
		),
		slog.Group("config_files",
			slog.String("license", p.LicenseFile),
			slog.String("credentials", p.CredentialsFile),
			slog.String("sheets_config", p.SheetsConfigFile),
		),
		slog.Group("report_files",
			slog.String("index_csv", p.IndexCSV),
			slog.String("ticker_summary_json", p.TickerSummaryJSON),
			slog.String("ticker_summary_csv", p.TickerSummaryCSV),
			slog.String("combined_data_csv", p.CombinedDataCSV),
		))
}

// ValidateRequiredFiles checks if critical files exist and returns detailed error information
func (p *Paths) ValidateRequiredFiles() error {
	requiredFiles := map[string]string{
		"License":     p.LicenseFile,
		"Credentials": p.CredentialsFile,
	}
	
	var missingFiles []string
	for name, path := range requiredFiles {
		if !FileExists(path) {
			missingFiles = append(missingFiles, fmt.Sprintf("%s (%s)", name, path))
		}
	}
	
	if len(missingFiles) > 0 {
		return fmt.Errorf("required files missing: %s", strings.Join(missingFiles, ", "))
	}
	
	return nil
}