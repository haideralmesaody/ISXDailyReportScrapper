/**
 * Report types for ISX Pulse
 * Following CLAUDE.md TypeScript strict mode standards
 */

// Import types from schemas
import type {
  ReportType,
  ReportMetadata,
  ReportFile,
  CSVData,
  ParsedCSVData,
  ReportApiResponse,
  ReportDownloadOptions,
  ReportError,
  GetReportsParams,
  DownloadReportParams
} from '@/lib/schemas/reports'

// Re-export for convenience
export type {
  ReportType,
  ReportMetadata,
  ReportFile,
  CSVData,
  ParsedCSVData,
  ReportApiResponse,
  ReportDownloadOptions,
  ReportError,
  GetReportsParams,
  DownloadReportParams
} from '@/lib/schemas/reports'

// Define component prop types
export interface CSVViewerProps {
  report: ReportMetadata | null
  csvData: ParsedCSVData | null  // Allow null for no-data state
  isLoading: boolean
  error: Error | null
}

export interface ReportListProps {
  reports: ReportMetadata[]
  selectedReport: ReportMetadata | null
  onSelectReport: (report: ReportMetadata) => void
  onDownloadReport: (report: ReportMetadata) => void
  isLoading?: boolean
}

export interface ReportFiltersProps {
  reports: ReportMetadata[]
  onFiltersChange: (filtered: ReportMetadata[]) => void
  showAdvanced?: boolean
}

export interface ReportTypeSelectorProps {
  selectedType: ReportType
  onTypeChange: (type: ReportType) => void
  counts: Record<ReportType, number>
}