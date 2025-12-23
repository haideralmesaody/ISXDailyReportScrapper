package operations

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	processingFilePattern   = regexp.MustCompile(`(?i)processing file\s+(\d+)\s+of\s+(\d+):\s*(.+)`)
	startingParsePattern    = regexp.MustCompile(`(?i)starting to parse\s+(.+)`)
	completedFilePattern    = regexp.MustCompile(`(?i)(?:completed file|converted)\s*:?\s*(.+)`)
	failedFilePattern       = regexp.MustCompile(`(?i)failed file:\s*(.+?)\s*-\s*(.+)`)
	foundExcelFilesPattern  = regexp.MustCompile(`(?i)found\s+(\d+)\s+excel files`)
	processingCompleteMatch = regexp.MustCompile(`(?i)processing complete`)
	totalOutputsPattern     = regexp.MustCompile(`(?i)total outputs to generate:\s*(\d+)`)
	generatedOutputPattern  = regexp.MustCompile(`(?i)generated output\s+(\d+)\s+of\s+(\d+):\s*(.+)`)
)

// ProcessingProgressSnapshot captures the current processing counters for broadcasting.
type ProcessingProgressSnapshot struct {
	FilesProcessed     int
	TotalFiles         int
	FailedFiles        int
	CurrentFile        string
	Message            string
	Progress           int
	OutputsGenerated   int
	TotalOutputs       int
	GenerationProgress int
	OutputStatuses     []*FileProcessingStatus
}

// ProcessingParser reduces processor.exe output to simple counters.
type ProcessingParser struct {
	totalFiles  int
	processed   int
	failed      int
	currentFile string

	// expectedTotal, if >0, is treated as SSOT and prevents downstream
	// stdout lines from inflating the total file count.
	expectedTotal int

	totalOutputs     int
	outputsGenerated int
	outputStatuses   []*FileProcessingStatus
}

// NewProcessingParser creates a parser seeded with the expected total file count.
func NewProcessingParser(totalFiles int) *ProcessingParser {
	if totalFiles < 0 {
		totalFiles = 0
	}
	return &ProcessingParser{
		totalFiles:    totalFiles,
		expectedTotal: totalFiles,
	}
}

// ParseLine updates counters based on a single processor.exe output line.
func (p *ProcessingParser) ParseLine(line string) (ProcessingProgressSnapshot, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return p.snapshot(""), false
	}

	changed := false
	message := ""

	if matches := processingFilePattern.FindStringSubmatch(trimmed); len(matches) == 4 {
		if current, err := strconv.Atoi(matches[1]); err == nil && current > 0 {
			// Update processed count to current-1 (files before this one)
			if p.processed != current-1 {
				p.processed = current - 1
				changed = true
			}
		}
		if total, err := strconv.Atoi(matches[2]); err == nil && total > 0 && p.expectedTotal == 0 && total != p.totalFiles {
			p.totalFiles = total
			changed = true
		}
		if file := strings.TrimSpace(matches[3]); file != "" {
			p.currentFile = filepath.Base(file)
			changed = true
		}
		message = p.defaultMessage(p.processed, p.effectiveTotal())
	} else if matches := startingParsePattern.FindStringSubmatch(trimmed); len(matches) == 2 {
		// This is just a status update - no counters change
		if file := strings.TrimSpace(matches[1]); file != "" {
			p.currentFile = filepath.Base(file)
			changed = true
			message = fmt.Sprintf("Parsing %s...", filepath.Base(file))
		}
	} else if matches := completedFilePattern.FindStringSubmatch(trimmed); len(matches) == 2 {
		p.processed++
		p.currentFile = ""
		changed = true
		base := filepath.Base(strings.TrimSpace(matches[1]))
		if base != "" {
			message = fmt.Sprintf("Completed %s (%d/%d files)", base, p.processed, p.effectiveTotal())
		}
	} else if matches := failedFilePattern.FindStringSubmatch(trimmed); len(matches) == 3 {
		p.failed++
		p.processed++
		p.currentFile = ""
		changed = true
		base := filepath.Base(strings.TrimSpace(matches[1]))
		reason := strings.TrimSpace(matches[2])
		message = fmt.Sprintf("Failed %s: %s", base, reason)
	} else if matches := foundExcelFilesPattern.FindStringSubmatch(trimmed); len(matches) == 2 {
		if total, err := strconv.Atoi(matches[1]); err == nil && total > 0 && p.expectedTotal == 0 && total != p.totalFiles {
			p.totalFiles = total
			changed = true
			message = fmt.Sprintf("Found %d files to process", total)
		}
	} else if processingCompleteMatch.MatchString(trimmed) {
		// REMOVED: Do NOT force p.processed = p.totalFiles here.
		// Progress should be driven by actual "Completed file:" messages, not this summary line.
		// This pattern can fire early (e.g., after batch detection finishes, before individual files complete).
		message = "Processing complete"
		changed = true // Still emit the message update
	} else if matches := totalOutputsPattern.FindStringSubmatch(trimmed); len(matches) == 2 {
		if total, err := strconv.Atoi(matches[1]); err == nil && total > 0 && total != p.totalOutputs {
			p.totalOutputs = total
			changed = true
			message = fmt.Sprintf("Preparing to generate %d outputs", total)
			// initialize pending statuses
			p.outputStatuses = make([]*FileProcessingStatus, total)
			for i := 0; i < total; i++ {
				p.outputStatuses[i] = &FileProcessingStatus{
					FileName: fmt.Sprintf("output_%d", i+1),
					Status:   "pending",
					Progress: 0,
				}
			}
		}
		// Debug
		fmt.Printf("[PARSER] Total outputs detected: %d\n", p.totalOutputs)
	} else if matches := generatedOutputPattern.FindStringSubmatch(trimmed); len(matches) == 4 {
		if total, err := strconv.Atoi(matches[2]); err == nil && total > 0 && total != p.totalOutputs {
			p.totalOutputs = total
			changed = true
			if len(p.outputStatuses) == 0 {
				p.outputStatuses = make([]*FileProcessingStatus, total)
				for i := 0; i < total; i++ {
					p.outputStatuses[i] = &FileProcessingStatus{
						FileName: fmt.Sprintf("output_%d", i+1),
						Status:   "pending",
						Progress: 0,
					}
				}
			}
		}
		p.outputsGenerated++
		if p.outputsGenerated < 0 {
			p.outputsGenerated = 0
		}
		changed = true
		base := filepath.Base(strings.TrimSpace(matches[3]))
		idx := p.outputsGenerated - 1
		if idx >= 0 && idx < len(p.outputStatuses) {
			p.outputStatuses[idx] = &FileProcessingStatus{
				FileName: base,
				Status:   "completed",
				Progress: 100,
			}
		}
		if base != "" {
			message = fmt.Sprintf("Generated %s (%d/%d outputs)", base, p.outputsGenerated, p.effectiveTotalOutputs())
		}
		// Debug
		fmt.Printf("[PARSER] Generated output: %d/%d\n", p.outputsGenerated, p.effectiveTotalOutputs())
	}

	return p.snapshot(message), changed
}

// Snapshot returns the latest counters without mutating state.
func (p *ProcessingParser) Snapshot() ProcessingProgressSnapshot {
	return p.snapshot("")
}

func (p *ProcessingParser) snapshot(message string) ProcessingProgressSnapshot {
	total := p.effectiveTotal()

	// Bound counts so they never exceed total files.
	boundedProcessed := p.processed
	boundedFailed := p.failed
	if total > 0 {
		if boundedProcessed > total {
			boundedProcessed = total
		}
		if boundedFailed > total {
			boundedFailed = total
		}
		if overflow := (boundedProcessed + boundedFailed) - total; overflow > 0 {
			if boundedFailed >= overflow {
				boundedFailed -= overflow
			} else {
				boundedProcessed -= overflow - boundedFailed
				boundedFailed = 0
			}
		}
	}

	progress := 0
	if total > 0 {
		progress = clampProgress(int(math.Round((float64(boundedProcessed+boundedFailed) / float64(total)) * 100)))
	}

	genTotal := p.effectiveTotalOutputs()
	genProgress := 0
	if genTotal > 0 {
		generated := p.outputsGenerated
		if generated > genTotal {
			generated = genTotal
		}
		genProgress = clampProgress(int(math.Round((float64(generated) / float64(genTotal)) * 100)))
	}

	if message == "" {
		message = p.defaultMessage(boundedProcessed, total)
	}

	return ProcessingProgressSnapshot{
		FilesProcessed:     boundedProcessed,
		TotalFiles:         p.totalFiles,
		FailedFiles:        boundedFailed,
		CurrentFile:        p.currentFile,
		Message:            message,
		Progress:           progress,
		OutputsGenerated:   p.outputsGenerated,
		TotalOutputs:       genTotal,
		GenerationProgress: genProgress,
		OutputStatuses:     p.outputStatuses,
	}
}

func (p *ProcessingParser) defaultMessage(boundedProcessed, total int) string {
	effectiveTotal := total
	if effectiveTotal == 0 {
		effectiveTotal = p.effectiveTotal()
	}

	if effectiveTotal > 0 {
		processed := boundedProcessed
		if processed > effectiveTotal {
			processed = effectiveTotal
		}
		return fmt.Sprintf("Processing files... (%d/%d files)", processed, effectiveTotal)
	}
	if p.currentFile != "" {
		return fmt.Sprintf("Processing %s", filepath.Base(p.currentFile))
	}
	return "Processing files..."
}

func (p *ProcessingParser) effectiveTotal() int {
	if p.totalFiles > 0 {
		return p.totalFiles
	}
	if p.processed > 0 {
		return p.processed
	}
	return 0
}

func (p *ProcessingParser) effectiveTotalOutputs() int {
	if p.totalOutputs > 0 {
		return p.totalOutputs
	}
	return 0
}
