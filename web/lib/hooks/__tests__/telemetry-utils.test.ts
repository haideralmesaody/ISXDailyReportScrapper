import { normalizeOperationMetadata, normalizeStageTelemetry } from '../../operations/telemetry-utils'

describe('normalizeStageTelemetry', () => {
  it('fills canonical defaults when fields are missing', () => {
    const step = {
      id: 'processing',
      metadata: {
        files_processed: undefined,
        progress_percent: undefined
      }
    }

    const normalized = normalizeStageTelemetry(step)

    expect(normalized.metadata.stage_id).toBe('processing')
    expect(normalized.metadata.phase).toBe('running')
    expect(normalized.metadata.phase_message).toContain('processing')
    expect(normalized.metadata.progress_percent).toBe(0)
    expect(normalized.metadata.files_processed).toBe(0)
    expect(normalized.metadata.total_files).toBe(0)
    expect(typeof normalized.metadata.stage_started_at).toBe('string')
    expect(typeof normalized.metadata.stage_updated_at).toBe('string')
  })

  it('preserves previous canonical data when available', () => {
    const previous = {
      stage_started_at: '2024-01-01T00:00:00.000Z',
      files_processed: 3
    }

    const step = {
      id: 'processing',
      metadata: {
        files_processed: 5,
        progress_percent: 50,
        phase: 'transforming'
      }
    }

    const normalized = normalizeStageTelemetry(step, previous)

    expect(normalized.metadata.stage_started_at).toBe(previous.stage_started_at)
    expect(normalized.metadata.files_processed).toBe(5)
    expect(normalized.metadata.progress_percent).toBe(50)
    expect(normalized.metadata.phase).toBe('transforming')
  })
})

describe('normalizeOperationMetadata', () => {
  it('promotes stage telemetry to operation metadata', () => {
    const stageMetadata = {
      phase: 'downloading',
      phase_message: 'Downloading files',
      progress_percent: 42
    }

    const normalized = normalizeOperationMetadata({}, stageMetadata)

    expect(normalized.phase).toBe('downloading')
    expect(normalized.current_phase).toBe('downloading')
    expect(normalized.phase_message).toBe('Downloading files')
    expect(normalized.stage_message).toBe('Downloading files')
    expect(normalized.progress_percent).toBe(42)
  })
})
