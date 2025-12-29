export type FileStatus = 'pending' | 'processing' | 'completed' | 'failed'

export type FileStatusLike = {
  status?: unknown
  progress?: unknown
}

const clamp = (value: number, min: number, max: number) =>
  Math.min(max, Math.max(min, value))

const toFiniteNumber = (value: unknown): number | null => {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim().length > 0) {
    const parsed = Number(value)
    if (Number.isFinite(parsed)) return parsed
  }
  return null
}

const normalizeStatus = (value: unknown): FileStatus => {
  if (typeof value !== 'string') return 'pending'
  const normalized = value.toLowerCase().trim()
  switch (normalized) {
    case 'completed':
    case 'complete':
    case 'success':
    case 'done':
      return 'completed'
    case 'processing':
    case 'running':
    case 'active':
    case 'in_progress':
      return 'processing'
    case 'failed':
    case 'error':
    case 'aborted':
      return 'failed'
    default:
      return 'pending'
  }
}

/**
 * Compute a "true" aggregate progress percent based on file statuses.
 *
 * - completed/failed files contribute 1.0
 * - processing file contributes `progress/100` when provided, otherwise 0
 * - pending contributes 0
 *
 * This avoids misleading 0% when a file is actively processing and the backend
 * provides per-file progress (e.g., indices stage sets the current file to 50).
 */
export function calculateFileAggregateProgressPercent(input: {
  totalFiles?: number | null
  fileStatuses?: FileStatusLike[] | null
}): number | undefined {
  const fileStatuses = Array.isArray(input.fileStatuses) ? input.fileStatuses : []
  const totalFromArg = toFiniteNumber(input.totalFiles)
  const total = totalFromArg && totalFromArg > 0 ? totalFromArg : fileStatuses.length
  if (!total || total <= 0) return undefined

  let completedUnits = 0
  for (const raw of fileStatuses) {
    const status = normalizeStatus(raw?.status)
    if (status === 'completed' || status === 'failed') {
      completedUnits += 1
      continue
    }
    if (status === 'processing') {
      const progress = toFiniteNumber(raw?.progress)
      if (progress !== null) {
        completedUnits += clamp(progress, 0, 100) / 100
      }
    }
  }

  return clamp((completedUnits / total) * 100, 0, 100)
}

export function formatProgressPercentLabel(percent: number | undefined): string {
  if (percent === undefined || !Number.isFinite(percent)) return '—'
  const clamped = clamp(percent, 0, 100)
  if (clamped > 0 && clamped < 1) return '<1%'
  if (clamped < 10) return `${clamped.toFixed(1)}%`
  return `${Math.round(clamped)}%`
}

