/**
 * Type definitions for ISX Daily Reports Scrapper
 * Central type definitions for the application
 */

// ============================================================================
// API Error Types (RFC 7807 Problem Details)
// ============================================================================

export interface ApiError {
  type: string
  title: string
  status: number
  detail?: string
  instance?: string
  trace_id?: string
  errors?: Record<string, string>
}

// ============================================================================
// License Types
// ============================================================================

export interface LicenseActivationRequest {
  license_key: string
  device_fingerprint?: DeviceFingerprint
  activation_id?: string
}

export interface LicenseResponse {
  status: 'valid' | 'invalid' | 'expired' | 'pending' | 'reactivated'
  message: string
  expiry_date?: string
  days_remaining?: number
  activation_id?: string
  device_info?: DeviceInfo
  features?: string[]
  reactivation_count?: number
  reactivation_limit?: number
  similarity_score?: number
  remaining_attempts?: number
}

export interface LicenseApiResponse {
  license_status?: 'active' | 'inactive' | 'expired' | 'invalid' | 'error' | 'not_activated' | 'warning' | 'critical' | 'reactivated'
  status: 'valid' | 'invalid' | 'expired' | 'pending' | 'reactivated'
  message: string
  expiry_date?: string
  days_remaining?: number
  days_left?: number
  activation_id?: string
  device_info?: DeviceInfo
  features?: string[]
  last_check?: string
  license_info?: {
    expiry_date?: string
    activation_date?: string
    device_info?: DeviceInfo
  }
  reactivation_count?: number
  reactivation_limit?: number
  similarity_score?: number
  remaining_attempts?: number
  trace_id?: string
  timestamp?: string
}

// ============================================================================
// Scratch Card Types
// ============================================================================

export interface ScratchCardData {
  code: string
  format: 'standard' | 'scratch' // ISX1M02LYE1F9QJHR9D7Z vs ISX-XXXX-XXXX-XXXX
  revealed: boolean
  activationId?: string
}

export interface ScratchCardActivationRequest {
  scratch_code: string
  device_fingerprint: DeviceFingerprint
  activation_id: string
}

export interface ScratchCardActivationResponse {
  success: boolean
  license_key: string
  status: 'activated' | 'already_used' | 'invalid' | 'expired'
  activation_id: string
  device_info: DeviceInfo
  expiry_date: string
  features: string[]
  message: string
}

// ============================================================================
// Device Fingerprint Types
// ============================================================================

export interface DeviceFingerprint {
  browser: string
  browserVersion: string
  os: string
  osVersion: string
  platform: string
  screenResolution: string
  timezone: string
  language: string
  userAgent: string
  hash: string
  timestamp: string
}

export interface DeviceInfo {
  fingerprint: string
  browser: string
  os: string
  platform: string
  first_activation: string
  last_seen: string
  trusted: boolean
}

// ============================================================================
// License Status Types
// ============================================================================

export interface LicenseStatus {
  isActive: boolean
  daysRemaining: number
  expiryDate: string
  status: 'active' | 'warning' | 'critical' | 'expired' | 'invalid'
  activationHistory: LicenseActivationHistory[]
  deviceInfo: DeviceInfo
  features: string[]
}

export interface LicenseActivationHistory {
  id: string
  date: string
  action: 'activated' | 'renewed' | 'transferred' | 'deactivated'
  device: string
  ip_address?: string
  success: boolean
  message?: string
}

// ============================================================================
// Operation Types
// ============================================================================

export interface Operation {
  id: string
  type: string
  name: string
  description: string
  status: 'idle' | 'running' | 'completed' | 'failed' | 'cancelled'
  progress: number
  created_at: string
  updated_at: string
  started_at?: string
  completed_at?: string
  config: OperationConfig
  stages: OperationStage[]
  metadata?: Record<string, any>
  results?: OperationResult[]
  error?: string
}

// ============================================================================
// Enhanced Operation Types (Phase 1)
// ============================================================================

// Base operation status types for backward compatibility
export type OperationStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
export type StageStatus = 'pending' | 'running' | 'completed' | 'failed' | 'skipped'
export type TriggerType = 'user' | 'schedule' | 'api' | 'system' | 'recovery'

// Core operation snapshot structure (matches backend OperationSnapshot)
export interface OperationSnapshot {
  operation_id: string
  status: OperationStatus
  progress: number
  current_step: string
  steps: StepSnapshot[]
  started_at: string
  updated_at: string
  completed_at?: string
  error?: string
  message?: string
  metadata?: Record<string, any>
}

// Step snapshot structure
export interface StepSnapshot {
  id: string
  name: string
  status: StageStatus
  progress: number
  message?: string
  error?: string
  metadata?: Record<string, any>
}

// Enhanced operation snapshot with rich metadata (matches backend OperationSnapshotDTO)
export interface OperationSnapshotDTO {
  // Base OperationSnapshot fields (embedded for compatibility)
  operation_id: string
  status: OperationStatus
  progress: number
  current_step: string
  steps: StepSnapshot[]
  started_at: string
  updated_at: string
  completed_at?: string
  error?: string
  message?: string
  metadata?: Record<string, any>

  // Enhanced fields
  pipeline_summary?: PipelineSummary
  stage_timeline?: StageTimeline[]
  derived_fields?: DerivedFields
}

// Pipeline summary with high-level metrics
export interface PipelineSummary {
  total_steps: number
  completed_steps: number
  failed_steps: number
  skipped_steps: number
  running_steps: number
  pending_steps: number
  overall_progress: number // 0-100, weighted by step importance
  estimated_time_remaining: string // ISO duration format
  triggered_by: TriggerType
  triggered_by_user?: string
  active_stage_index: number
  active_stage_name?: string
  operation_mode?: string
  operation_type?: string
}

// Detailed stage timeline with metrics
export interface StageTimeline {
  stage_id: string
  stage_name: string
  stage_type: string
  status: StageStatus
  progress: number // 0-100
  start_time?: string
  end_time?: string
  duration: string // ISO duration format
  estimated_eta: string // ISO duration format
  dependencies?: string[]
  stage_metrics?: StageMetricsInterface
  messages?: StageMessage[]
}

// Stage messages for detailed logging
export interface StageMessage {
  timestamp: string
  level: 'info' | 'warn' | 'error' | 'debug'
  message: string
  details?: string
}

// Derived fields with calculated values
export interface DerivedFields {
  execution_time: string // ISO duration format
  average_step_time: string // ISO duration format
  throughput_per_min: number
  success_rate: number // percentage
  health_score: number // 0-100
  is_recoverable: boolean
  next_step_eta: string // ISO duration format
}

// ============================================================================
// Stage Metrics Interfaces (matches backend StageMetricsInterface)
// ============================================================================

// Base interface for all stage metrics
export interface StageMetricsInterface {
  // Type discriminator for runtime type checking
  type: 'scraping' | 'processing' | 'indices' | 'liquidity'
  // Common validation method signature
  validate?(): ValidationError[]
  // Summary method for UI display
  getSummary?(): Record<string, any>
}

// Scraping stage metrics with calendar information
export interface ScrapingStageMetrics extends StageMetricsInterface {
  type: 'scraping'
  trading_days: number
  weekend_days: number
  holiday_days: number
  download_velocity: number // files per minute
  expected_files: number
  processed_files: number
  skipped_files: number
  failed_files: number
  total_size: number // bytes
  downloaded_size: number // bytes
  calendar_segments: CalendarSegment[]
  last_processed_file?: string
  current_file?: string
  download_source: string
  retry_count: number
  network_errors: number
}

// Processing stage metrics for data transformation
export interface ProcessingStageMetrics extends StageMetricsInterface {
  type: 'processing'
  files_processed: number
  validation_errors: number
  parsing_errors: number
  transformation_errors: number
  records_processed: number
  records_skipped: number
  processing_speed: number // records per second
  data_quality_score: number // 0-100
  output_files: number
  output_size: number // bytes
  processing_time: string // ISO duration format
}

// Indices stage metrics for index extraction
export interface IndicesStageMetrics extends StageMetricsInterface {
  type: 'indices'
  indices_extracted: number
  market_data_points: number
  index_calculated: boolean
  index_value?: number
  index_change?: number
  index_change_percent?: number
  constituents: number
  processing_time: string // ISO duration format
  data_sources: string[]
  validation_passed: boolean
}

// Liquidity stage metrics for liquidity analysis
export interface LiquidityStageMetrics extends StageMetricsInterface {
  type: 'liquidity'
  tickers_analyzed: number
  liquidity_scores: number
  liquidity_buckets: Record<string, number> // high|medium|low -> count
  average_volume: number
  average_value: number // average daily value traded
  processing_time: string // ISO duration format
  market_cap_total: number
  illiquid_count: number
  liquid_count: number
}


// Calendar segment for Iraq market trading days
export interface CalendarSegment {
  date: string // ISO date format (YYYY-MM-DD)
  date_string: string // YYYY-MM-DD format (same as date for consistency)
  day_of_week: number // 0=Sunday, 6=Saturday
  status: CalendarSegmentStatus
  file_name?: string
  file_size?: number
  download_time?: string // ISO duration format
  retry_count: number
  last_error?: string
}

// Calendar segment status types
export type CalendarSegmentStatus =
  | 'downloaded'
  | 'pending'
  | 'weekend'
  | 'holiday'
  | 'downloading'
  | 'failed'
  | 'skipped'

// Union type for all stage metrics implementations
export type StageMetrics =
  | ScrapingStageMetrics
  | ProcessingStageMetrics
  | IndicesStageMetrics
  | LiquidityStageMetrics

// Validation error interface
export interface ValidationError {
  field: string
  message: string
  value?: any
}

export interface OperationConfig {
  auto_start: boolean
  retry_attempts: number
  timeout_seconds: number
  steps: string[]
  notification_email?: string
  parameters?: Record<string, any>
}

export interface OperationStage {
  id: string
  name: string
  description: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped'
  progress: number
  started_at?: string
  completed_at?: string
  duration?: number
  error?: string
  results?: Record<string, any>
}

export interface OperationResult {
  stage_id: string
  stage_name: string
  success: boolean
  data?: any
  error?: string
  timestamp: string
}

export interface OperationTypeDefinition {
  id: string
  name: string
  description: string
  category: string
  config_schema: Record<string, any>
  required_features?: string[]
  estimated_duration?: string
}

export interface CreateOperationRequest {
  operation: string
  stage_id?: string
  parameters?: Record<string, any>
}

export interface CreateOperationResponse {
  operation?: Operation
  job_id?: string
  operation_id?: string
  status?: string
  message?: string
  poll_url?: string
  websocket_url?: string
}

// ============================================================================
// Job Types
// ============================================================================

export interface JobStatus {
  job_id?: string
  id?: string
  operation_id: string
  stage_id?: string
  stage_name?: string
  status: 'pending' | 'queued' | 'running' | 'completed' | 'failed' | 'cancelled'
  progress?: number
  message?: string
  error?: string
  created_at?: string
  started_at?: string
  completed_at?: string
  metadata?: Record<string, any>
  result?: any
}

export interface JobListResponse {
  jobs: JobStatus[]
  count?: number
  stats?: Record<string, any>
  total?: number
  page?: number
  page_size?: number
}

// ============================================================================
// Market Data Types
// ============================================================================

export interface Ticker {
  symbol: string
  company_name: string
  last_price: number
  change: number
  change_percent: number
  volume: number
  value: number
  high: number
  low: number
  open: number
  trades: number
  last_update: string
}

export interface Report {
  id: string
  ticker: string
  type: 'daily' | 'summary' | 'liquidity'
  date: string
  file_path: string
  file_size: number
  created_at: string
  metadata?: Record<string, any>
}

export interface MarketSummary {
  date: string
  total_volume: number
  total_value: number
  total_trades: number
  advancing: number
  declining: number
  unchanged: number
  top_gainers: Ticker[]
  top_losers: Ticker[]
  most_active: Ticker[]
  indices: MarketIndex[]
}

export interface MarketIndex {
  name: string
  value: number
  change: number
  change_percent: number
  date: string
}

// ============================================================================
// UI State Types
// ============================================================================

export interface LoadingState {
  isLoading: boolean
  message?: string
  progress?: number
}

export interface ErrorState {
  hasError: boolean
  error?: ApiError | Error | string
  timestamp?: string
}

export interface PaginationState {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

export interface SortState {
  column: string
  direction: 'asc' | 'desc'
}

export interface FilterState {
  [key: string]: any
}

// ============================================================================
// WebSocket Types (Enhanced for Phase 1)
// ============================================================================

export interface WebSocketMessage {
  type: string
  data: any
  timestamp: string
  id?: string
}

// Enhanced operation snapshot message (primary event type)
export interface OperationSnapshotMessage extends WebSocketMessage {
  type: 'operation:snapshot'
  data: OperationSnapshotDTO | OperationSnapshot // Support both formats
}

// Legacy operation status message (for backward compatibility)
export interface OperationStatusMessage extends WebSocketMessage {
  type: 'operation_status'
  data: {
    operation_id: string
    status: Operation['status']
    progress: number
    stage?: OperationStage
    error?: string
  }
}

// License status message
export interface LicenseStatusMessage extends WebSocketMessage {
  type: 'license_status'
  data: {
    status: LicenseApiResponse['status']
    days_remaining?: number
    expiry_date?: string
    message?: string
  }
}

// Union type for all operation-related WebSocket messages
export type OperationWebSocketMessage =
  | OperationSnapshotMessage
  | OperationStatusMessage

// Union type for all WebSocket messages
export type TypedWebSocketMessage =
  | OperationSnapshotMessage
  | OperationStatusMessage
  | LicenseStatusMessage
  | (WebSocketMessage & { type: 'market_update' })
  | (WebSocketMessage & { type: 'system_status' })
  | (WebSocketMessage & { type: 'connection_status' })

// ============================================================================
// Chart and Analysis Types
// ============================================================================

export interface ChartDataPoint {
  x: number
  y: number
  [key: string]: any
}

export interface TimeSeriesData {
  timestamp: number
  value: number
  volume?: number
  open?: number
  high?: number
  low?: number
  close?: number
}

// ============================================================================
// Utility Types
// ============================================================================

export type Status = 'idle' | 'loading' | 'success' | 'error'

export type Theme = 'light' | 'dark' | 'system'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface ToastMessage {
  id: string
  type: ToastType
  title: string
  description?: string
  duration?: number
  action?: {
    label: string
    onClick: () => void
  }
}

// ============================================================================
// Form Types
// ============================================================================

export interface FormField {
  name: string
  label: string
  type: 'text' | 'email' | 'password' | 'number' | 'select' | 'checkbox' | 'textarea'
  placeholder?: string
  required?: boolean
  options?: { label: string; value: string }[]
  validation?: any
}

export interface FormState {
  values: Record<string, any>
  errors: Record<string, string>
  touched: Record<string, boolean>
  isSubmitting: boolean
  isDirty: boolean
  isValid: boolean
}

// ============================================================================
// Component Props Types
// ============================================================================

export interface BaseComponentProps {
  className?: string
  children?: React.ReactNode
}

export interface PageProps {
  params?: Record<string, string>
  searchParams?: Record<string, string>
}

// ============================================================================
// API Response Wrappers
// ============================================================================

export interface ApiResponse<T = any> {
  data: T
  message?: string
  timestamp: string
  request_id?: string
}

export interface PaginatedResponse<T = any> {
  data: T[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
  message?: string
  timestamp: string
}

// ============================================================================
// Configuration Types
// ============================================================================

export interface AppConfig {
  api_base_url: string
  websocket_url: string
  version: string
  features: string[]
  debug: boolean
}

export interface UserPreferences {
  theme: Theme
  language: string
  timezone: string
  notifications: {
    email: boolean
    browser: boolean
    sounds: boolean
  }
  dashboard: {
    auto_refresh: boolean
    refresh_interval: number
    default_view: string
  }
}

// ============================================================================
// Strategy Types
// ============================================================================

export interface StrategyInfo {
  id: string
  name: string
  description: string
}

export type SignalAction = 'BUY' | 'SELL' | 'HOLD'

export interface StrategySignal {
  id: string
  strategy_id: string
  symbol: string
  action: SignalAction
  strength: number
  price: number
  quantity: number
  timestamp: string
  reasoning: string
  metadata?: Record<string, unknown> | null
  valid_until: string
}

export interface ExecuteBatchRequest {
  data_points?: number
  symbols?: string[]
  include_backtest?: boolean
  backtest_start_date?: string
  backtest_end_date?: string
  transaction_fee?: number
}

export interface ExecuteBatchError {
  symbol: string
  error: string
}

export interface ExecuteBatchResponse {
  run_id: string
  strategy_id: string
  started_at: string
  completed_at: string
  total: number
  buy_count: number
  sell_count: number
  hold_count: number
  signals: StrategySignal[]
  errors?: ExecuteBatchError[]
  backtest?: BatchBacktestSummary
}

export interface StrategyRunInfo {
  run_id: string
  strategy_id: string
  started_at: string
  completed_at: string
  total: number
  buy_count: number
  sell_count: number
  hold_count: number
  error_count: number
  has_backtest: boolean
}

export interface BatchBacktestSummary {
  start_date: string
  end_date: string
  transaction_fee: number
  by_ticker: BacktestTickerSummary[]
  aggregate?: BatchBacktestAggregate
}

export interface BatchBacktestAggregate {
  total_tickers: number
  successful_tickers: number
  error_tickers: number
  open_positions: number
  total_completed_trades: number
  total_winning_trades: number
  total_losing_trades: number
  avg_net_profit_pct: number
  median_net_profit_pct: number
  top_tickers?: BacktestAggregateTicker[]
  bottom_tickers?: BacktestAggregateTicker[]
}

export interface BacktestAggregateTicker {
  symbol: string
  net_profit_pct: number
  gross_profit_pct: number
  completed_trades: number
  winning_trades: number
  losing_trades: number
  open_position: boolean
}

export interface BacktestTickerSummary {
  symbol: string
  completed_trades: number
  winning_trades: number
  losing_trades: number
  gross_profit_pct: number
  net_profit_pct: number
  open_position: boolean
  error?: string
}

export interface BacktestTrade {
  symbol: string
  status: 'CLOSED' | 'OPEN'
  signal_buy_date?: string
  buy_date?: string
  buy_price?: number
  signal_sell_date?: string
  sell_date?: string
  sell_price?: number
  gross_return_pct: number
  net_return_pct: number
  transaction_fee: number
}

export interface BacktestTickerDetails {
  symbol: string
  summary: BacktestTickerSummary
  trades: BacktestTrade[]
}

// ============================================================================
// Export All Types
// ============================================================================

// Re-export reports types
export * from './reports'

// Note: React doesn't export named 'React' type
// Import React directly where needed: import React from 'react'

