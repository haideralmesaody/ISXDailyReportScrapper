package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// CSVWriter provides CSV export functionality
type CSVWriter struct {
	basePath string
}

// NewCSVWriter creates a new CSV writer instance
func NewCSVWriter(basePath string) *CSVWriter {
	return &CSVWriter{basePath: basePath}
}

// WriteOptions configures CSV writing behavior
type WriteOptions struct {
	Headers    []string
	Records    [][]string
	Append     bool
	BOMPrefix  bool // Add UTF-8 BOM for Excel compatibility
}

// WriteCSV writes data to a CSV file with the given options
func (w *CSVWriter) WriteCSV(filePath string, options WriteOptions) error {
	fullPath := filepath.Join(w.basePath, filePath)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Open file with appropriate flags
	flags := os.O_CREATE | os.O_WRONLY
	if options.Append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	
	file, err := os.OpenFile(fullPath, flags, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// Write BOM if requested (helps Excel recognize UTF-8)
	if options.BOMPrefix && !options.Append {
		if _, err := file.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
			return fmt.Errorf("failed to write BOM: %w", err)
		}
	}
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write headers if not appending
	if !options.Append && len(options.Headers) > 0 {
		if err := writer.Write(options.Headers); err != nil {
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}
	
	// Write records
	for i, record := range options.Records {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}
	
	return writer.Error()
}

// WriteSimpleCSV writes a simple CSV file with headers and records
func (w *CSVWriter) WriteSimpleCSV(filePath string, headers []string, records [][]string) error {
	return w.WriteCSV(filePath, WriteOptions{
		Headers: headers,
		Records: records,
		Append:  false,
		BOMPrefix: true,
	})
}

// AppendToCSV appends records to an existing CSV file
func (w *CSVWriter) AppendToCSV(filePath string, records [][]string) error {
	return w.WriteCSV(filePath, WriteOptions{
		Records: records,
		Append:  true,
	})
}

// StreamWriter provides streaming CSV writing for large datasets
type StreamWriter struct {
	file   *os.File
	writer *csv.Writer
}

// CreateStreamWriter creates a new streaming CSV writer
func (w *CSVWriter) CreateStreamWriter(filePath string, headers []string) (*StreamWriter, error) {
	fullPath := filepath.Join(w.basePath, filePath)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	
	// Write BOM for Excel compatibility
	if _, err := file.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write BOM: %w", err)
	}
	
	writer := csv.NewWriter(file)
	
	// Write headers
	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write headers: %w", err)
		}
	}
	
	return &StreamWriter{
		file:   file,
		writer: writer,
	}, nil
}

// WriteRecord writes a single record to the stream
func (s *StreamWriter) WriteRecord(record []string) error {
	return s.writer.Write(record)
}

// Close flushes and closes the stream writer
func (s *StreamWriter) Close() error {
	s.writer.Flush()
	if err := s.writer.Error(); err != nil {
		s.file.Close()
		return err
	}
	return s.file.Close()
}