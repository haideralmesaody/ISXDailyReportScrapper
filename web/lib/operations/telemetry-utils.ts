/**
 * Canonical telemetry helpers shared between websocket normalization and UI components.
 * Ensures every stage snapshot exposes the SSOT fields defined in docs/ops/OPERATIONS_DATA_FLOW.md.
 */

type StageLike = {
  id?: string
  step_id?: string
  stage_id?: string
  status?: string
  progress?: number
  message?: string
  started_at?: string
  start_time?: string
  updated_at?: string
  metadata?: Record<string, any>
}

const isDev = process.env.NODE_ENV === 'development'

const canonicalKeys = [
  'stage_id',
  'stage_alias',
  'phase',
  'phase_message',
  'progress_percent',
  'files_processed',
  'total_files',
  'current_file',
  'stage_started_at',
  'stage_updated_at'
]

const shouldLogTelemetryWarnings = (): boolean => {
  if (!isDev || typeof window === 'undefined') return false
  try {
    return window.localStorage?.getItem('debug_telemetry') === 'true'
  } catch {
    return false
  }
}

const toFiniteNumber = (value: unknown): number | null => {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim().length) {
    const parsed = Number(value)
    if (Number.isFinite(parsed)) return parsed
  }
  return null
}

const clampNumber = (value: number, min = 0, max = Number.POSITIVE_INFINITY) =>
  Math.min(max, Math.max(min, value))

const coerceNumber = (value: unknown, fallback = 0, min = 0, max = Number.POSITIVE_INFINITY) => {
  const parsed = toFiniteNumber(value)
  if (parsed === null) return clampNumber(fallback, min, max)
  return clampNumber(parsed, min, max)
}

const coerceIsoTimestamp = (value: unknown, fallbackToNow = true): string | undefined => {
  if (value instanceof Date && !Number.isNaN(value.getTime())) {
    return value.toISOString()
  }
  if (typeof value === 'number' && Number.isFinite(value)) {
    const date = new Date(value)
    if (!Number.isNaN(date.getTime())) return date.toISOString()
  }
  if (typeof value === 'string' && value.length > 0) {
    const parsed = Date.parse(value)
    if (!Number.isNaN(parsed)) {
      return new Date(parsed).toISOString()
    }
  }
  if (!fallbackToNow) {
    return undefined
  }
  const now = new Date()
  return now.toISOString()
}

const normalizeFileStatus = (raw: any) => {
  if (!raw || typeof raw !== 'object') {
    return raw
  }

  const filename =
    raw.filename ||
    raw.file_name ||
    raw.FileName ||
    raw.name ||
    'Unknown file'

  const status = typeof raw.status === 'string' ? raw.status : 'pending'

  return {
    ...raw,
    filename,
    status,
    size_mb: toFiniteNumber(raw.size_mb) ?? toFiniteNumber(raw.sizeMB) ?? raw.size_mb,
    progress: toFiniteNumber(raw.progress) ?? undefined,
    error_message: raw.error_message || raw.errorMessage,
    processing_time_ms: toFiniteNumber(raw.processing_time_ms) ?? toFiniteNumber(raw.processingMs) ?? raw.processing_time_ms
  }
}

const emitTelemetryWarning = (stageId: string, missing: string[]) => {
  if (!shouldLogTelemetryWarnings() || missing.length === 0) return

  console.warn('[telemetry-utils] Stage telemetry missing fields', {
    stageId,
    missing
  })
}

export const CANONICAL_STAGE_KEYS = canonicalKeys

/**
 * Normalize step metadata so that every canonical field is available to the UI.
 */
export function normalizeStageTelemetry(step: StageLike, previousMetadata?: Record<string, any>) {
  const baseMetadata = typeof previousMetadata === 'object' ? { ...previousMetadata } : {}
  const incomingMetadata = step?.metadata && typeof step.metadata === 'object' ? step.metadata : {}
  const metadata = { ...baseMetadata, ...incomingMetadata }

  const resolvedStageId =
    metadata.stage_id ||
    metadata.target_stage_id ||
    step.stage_id ||
    step.step_id ||
    step.id ||
    baseMetadata.stage_id ||
    'stage'

  metadata.stage_id = resolvedStageId
  if (!metadata.stage_alias && metadata.target_stage_id) {
    metadata.stage_alias = metadata.target_stage_id
  }

  const phaseCandidate = typeof metadata.phase === 'string'
    ? metadata.phase
    : (typeof metadata.current_phase === 'string' ? metadata.current_phase : undefined)

  metadata.phase = (phaseCandidate || 'running').toLowerCase()
  metadata.current_phase = metadata.phase

  const phaseMessage =
    metadata.phase_message ||
    metadata.stage_message ||
    step.message ||
    baseMetadata.phase_message ||
    `Working on ${resolvedStageId}`

  metadata.phase_message = phaseMessage
  metadata.stage_message = phaseMessage

  const progressPercent = metadata.progress_percent ?? step.progress ?? baseMetadata.progress_percent
  metadata.progress_percent = coerceNumber(progressPercent, 0, 0, 100)

  // Normalize trading day counters (primarily for scraping)
  const tradingDaysCompleted = toFiniteNumber(metadata.trading_days_completed)
  if (tradingDaysCompleted !== null) {
    metadata.trading_days_completed = tradingDaysCompleted
  }
  const tradingDaysTotal = toFiniteNumber(metadata.trading_days_total)
  if (tradingDaysTotal !== null) {
    metadata.trading_days_total = tradingDaysTotal
  }

  metadata.files_processed = coerceNumber(
    metadata.files_processed ?? metadata.processed_files ?? baseMetadata.files_processed,
    0
  )

  metadata.total_files = coerceNumber(
    metadata.total_files ?? metadata.expected_files ?? baseMetadata.total_files,
    0
  )

  // Outputs generation (processing stage)
  metadata.files_generated = coerceNumber(
    metadata.files_generated ?? metadata.outputs_generated ?? baseMetadata.files_generated,
    0
  )
  metadata.total_outputs = coerceNumber(
    metadata.total_outputs ?? metadata.expected_outputs ?? baseMetadata.total_outputs,
    0
  )
  metadata.generation_progress_percent = coerceNumber(
    metadata.generation_progress_percent ?? metadata.output_progress ?? baseMetadata.generation_progress_percent,
    0,
    0,
    100
  )

  // Normalize processing stage metrics
  if (metadata.stage_id === 'processing') {
    const filesProcessed = coerceNumber(metadata.files_processed ?? metadata.processed_files, 0)
    const totalFiles = coerceNumber(metadata.total_files ?? metadata.expected_files, 0)
    const failedFiles = coerceNumber(metadata.failed_files, 0)
    metadata.files_processed = filesProcessed
    metadata.total_files = totalFiles
    metadata.failed_files = failedFiles
    if (metadata.current_file) {
      metadata.current_file = String(metadata.current_file)
    }
    if (Array.isArray(metadata.file_statuses)) {
      metadata.file_statuses = metadata.file_statuses.map(normalizeFileStatus)
    }
    const errors: string[] = []
    if (totalFiles <= 0) errors.push('total_files missing or invalid')
    if (filesProcessed < 0) errors.push('files_processed missing or invalid')
    if (errors.length) {
      metadata.telemetry_errors = errors
      metadata.telemetry_error = errors.join('; ')
    }
  }

  // Normalize indices stage metrics
  if (metadata.stage_id === 'indices' || metadata.stage_id === 'index') {
    const recordsProcessed = coerceNumber(metadata.records_processed, 0)
    const indicesExtracted = coerceNumber(metadata.indices_extracted ?? recordsProcessed, recordsProcessed)
    const failedFiles = coerceNumber(metadata.failed_files, 0)
    metadata.records_processed = recordsProcessed
    metadata.indices_extracted = indicesExtracted
    metadata.failed_files = failedFiles
    if (metadata.current_file) {
      metadata.current_file = String(metadata.current_file)
    }
    const errors: string[] = []
    if (recordsProcessed < 0) errors.push('records_processed missing or invalid')
    if (errors.length) {
      metadata.telemetry_errors = errors
      metadata.telemetry_error = errors.join('; ')
    }
  }

  // Clamp scraping file counts to realistic bounds to avoid UI over-reporting (e.g., 590 instead of 59)
  if (metadata.stage_id === 'scraping') {
    const skippedCount = Array.isArray(metadata.skipped_files) ? metadata.skipped_files.length : null
    const holidaysDetected = toFiniteNumber(metadata.holidays_detected)
    const tradingDaysTotal = toFiniteNumber(metadata.trading_days_total)
    const tradingDaysCompleted = toFiniteNumber(metadata.trading_days_completed)

    const errors: string[] = []
    if (tradingDaysTotal === null) errors.push('trading_days_total missing or invalid')
    if (holidaysDetected === null && skippedCount === null) errors.push('holidays_detected missing and skipped_files absent')
    if (tradingDaysCompleted === null) errors.push('trading_days_completed missing or invalid')

    // Preserve backend-provided numbers; do not derive or clamp
    if (tradingDaysTotal !== null) {
      metadata.trading_days_total = tradingDaysTotal
    }
    if (tradingDaysCompleted !== null) {
      metadata.trading_days_completed = tradingDaysCompleted
      metadata.files_processed = tradingDaysCompleted
    }
    if (holidaysDetected !== null) {
      metadata.holidays_detected = holidaysDetected
    } else if (skippedCount !== null) {
      metadata.holidays_detected = skippedCount
    }

    if (errors.length > 0) {
      metadata.telemetry_errors = errors
      metadata.telemetry_error = errors.join('; ')
    }
  }

  if (!metadata.current_file && metadata.last_processed_file) {
    metadata.current_file = metadata.last_processed_file
  }

  const stageStartedSource =
    metadata.stage_started_at ||
    metadata.start_time ||
    step.started_at ||
    step.start_time ||
    baseMetadata.stage_started_at

  metadata.stage_started_at = coerceIsoTimestamp(stageStartedSource, true)

  const stageUpdatedSource =
    metadata.stage_updated_at ||
    step.updated_at ||
    baseMetadata.stage_updated_at ||
    Date.now()

  metadata.stage_updated_at = coerceIsoTimestamp(stageUpdatedSource, true)

  if (Array.isArray(metadata.file_statuses)) {
    metadata.file_statuses = metadata.file_statuses.map(normalizeFileStatus)
  }

  const missingKeys = canonicalKeys.filter((key) => metadata[key] === undefined || metadata[key] === null || metadata[key] === '')
  emitTelemetryWarning(resolvedStageId, missingKeys)

  return {
    ...step,
    metadata
  }
}

/**
 * Promote canonical stage metadata to the operation-level metadata bag so components
 * always have a consistent fallback even when a single stage is not focused.
 */
export function normalizeOperationMetadata(
  metadata: Record<string, any> | undefined,
  stageMetadata?: Record<string, any>
) {
  const result: Record<string, any> = metadata ? { ...metadata } : {}

  if (stageMetadata) {
    canonicalKeys.forEach((key) => {
      if (result[key] === undefined && stageMetadata[key] !== undefined) {
        result[key] = stageMetadata[key]
      }
    })
  }

  if (result.phase && !result.current_phase) {
    result.current_phase = result.phase
  } else if (!result.phase && result.current_phase) {
    result.phase = result.current_phase
  }

  if (result.phase_message && !result.stage_message) {
    result.stage_message = result.phase_message
  } else if (!result.phase_message && result.stage_message) {
    result.phase_message = result.stage_message
  }

  if (typeof result.progress_percent !== 'number' && typeof result.progress === 'number') {
    result.progress_percent = result.progress
  }

  return result
}
