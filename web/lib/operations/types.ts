export interface ScrapingTelemetry {
  operation_id: string
  status: 'running' | 'completed' | 'failed' | 'pending'
  progress_percent: number
  stage_message?: string
  from_date?: string
  to_date?: string
  downloaded_files: string[]
  skipped_files: string[]
  trading_days_total: number
  trading_days_completed: number
  trading_days_remaining?: number
  scraped_files_count?: number
  total_downloaded_mb?: number
  current_file?: string
  watchdog_status?: string
  holidays_detected?: number
  telemetry_missing?: boolean
  [key: string]: any
}

export interface OperationDeltaEvent {
  operation_id: string
  stage_id?: string
  status?: string
  progress?: number
  message?: string
  updated_at?: string
  metadata?: Record<string, any>
  sequence?: number
}
