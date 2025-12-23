import '@testing-library/jest-dom'
import { render, screen } from '@testing-library/react'
import UnifiedOperationProgress from '../UnifiedOperationProgress'

const baseOperation = {
  operation_id: 'op-test',
  name: 'Data Collection',
  status: 'running',
  progress: 0,
  message: '',
  error: null,
  metadata: {},
  steps: [
    {
      id: 'scraping',
      status: 'pending',
      progress: 0,
      metadata: {}
    }
  ],
  stepData: {
    id: 'scraping',
    status: 'pending',
    progress: 0,
    metadata: {}
  },
  stageContext: {
    allStages: [],
    pipelineMetadata: {},
    completedPipeline: false
  }
}

describe('UnifiedOperationProgress', () => {
  it('shows awaiting telemetry for pending stages without telemetry', () => {
    render(
      <UnifiedOperationProgress
        operation={baseOperation}
        pipelineMode={true}
        stageNumber={1}
        totalStages={6}
        showConnector={false}
      />
    )

    expect(screen.getByText(/Awaiting telemetry/i)).toBeInTheDocument()
    expect(screen.getByText(/Stage ready/i)).toBeInTheDocument()
  })

  it('shows running status and hides telemetry warning when data is present', () => {
    const runningOperation = {
      ...baseOperation,
      status: 'running',
      stepData: {
        id: 'scraping',
        status: 'running',
        progress: 0,
        metadata: {
          phase: 'downloading',
          phase_message: 'Downloading files',
          progress_percent: 42,
          trading_days_total: 10,
          trading_days_completed: 5,
          holidays_detected: 0,
          downloaded_files: ['2025-01-01', '2025-01-02', '2025-01-03', '2025-01-04', '2025-01-05']
        }
      },
      steps: [
        {
          id: 'scraping',
          status: 'running',
          progress: 0,
          metadata: {
            phase: 'downloading',
            phase_message: 'Downloading files',
            progress_percent: 42,
            trading_days_total: 10,
            trading_days_completed: 5,
            holidays_detected: 0,
            downloaded_files: ['2025-01-01', '2025-01-02', '2025-01-03', '2025-01-04', '2025-01-05']
          }
        }
      ]
    }

    render(
      <UnifiedOperationProgress
        operation={runningOperation}
        pipelineMode={false}
        stageNumber={1}
        totalStages={6}
        showConnector={false}
      />
    )

    expect(screen.queryByText(/Awaiting telemetry/i)).not.toBeInTheDocument()
    expect(screen.getByText(/Stage running/i)).toBeInTheDocument()
  })

  it('shows skipped state when pipeline completes without telemetry', () => {
    const skippedOperation = {
      ...baseOperation,
      status: 'completed',
      stepData: {
        id: 'scraping',
        status: 'pending',
        progress: 0,
        metadata: {
          stage_skipped: true,
          phase: 'skipped',
          phase_message: 'Stage not executed in this run',
          trading_days_total: 1,
          trading_days_completed: 1,
          holidays_detected: 0,
          downloaded_files: ['2025-01-01']
        }
      },
      steps: [
        {
          id: 'scraping',
          status: 'pending',
          progress: 0,
          metadata: {
            stage_skipped: true,
            phase: 'skipped',
            phase_message: 'Stage not executed in this run',
            trading_days_total: 1,
            trading_days_completed: 1,
            holidays_detected: 0,
            downloaded_files: ['2025-01-01']
          }
        }
      ]
    }

    render(
      <UnifiedOperationProgress
        operation={skippedOperation}
        pipelineMode={true}
        stageNumber={1}
        totalStages={6}
        showConnector={false}
      />
    )

    expect(screen.getAllByText(/Stage skipped/i).length).toBeGreaterThan(0)
    expect(screen.getByText(/not executed in this run/i)).toBeInTheDocument()
    expect(screen.queryByText(/Awaiting telemetry/i)).not.toBeInTheDocument()
  })

  it('renders liquidity telemetry metrics when provided', () => {
    const liquidityOperation = {
      ...baseOperation,
      status: 'running',
      steps: [
        {
          id: 'liquidity',
          status: 'running',
          progress: 80,
          metadata: {
            stage_id: 'liquidity',
            phase: 'scaling',
            phase_message: 'Scaling results',
            progress_percent: 80,
            tickers_analyzed: 120,
            liquidity_buckets: { high: 40, medium: 60, low: 20 },
            average_value: 1500000
          }
        }
      ],
      stepData: {
        id: 'liquidity',
        status: 'running',
        progress: 80,
        metadata: {
          stage_id: 'liquidity',
          phase: 'scaling',
          phase_message: 'Scaling results',
          progress_percent: 80,
          tickers_analyzed: 120,
          liquidity_buckets: { high: 40, medium: 60, low: 20 },
          average_value: 1500000
        }
      }
    }

    render(
      <UnifiedOperationProgress
        operation={liquidityOperation}
        pipelineMode
        stageNumber={4}
        totalStages={6}
        showConnector={false}
      />
    )

    expect(screen.getByText(/Tickers analyzed/i)).toBeInTheDocument()
    expect(screen.getByText(/High: 40/)).toBeInTheDocument()
    expect(screen.getByText(/Medium: 60/)).toBeInTheDocument()
  })

  it('renders analysis telemetry metrics when provided', () => {
    const analysisOperation = {
      ...baseOperation,
      status: 'running',
      steps: [
        {
          id: 'analysis',
          status: 'running',
          progress: 90,
          metadata: {
            stage_id: 'analysis',
            phase: 'generating',
            phase_message: 'Generating alerts',
            progress_percent: 90,
            price_alerts: 8,
            rsi_alerts: 4,
            traded_stocks: 12
          }
        }
      ],
      stepData: {
        id: 'analysis',
        status: 'running',
        progress: 90,
        metadata: {
          stage_id: 'analysis',
          phase: 'generating',
          phase_message: 'Generating alerts',
          progress_percent: 90,
          price_alerts: 8,
          rsi_alerts: 4,
          traded_stocks: 12
        }
      }
    }

    render(
      <UnifiedOperationProgress
        operation={analysisOperation}
        pipelineMode
        stageNumber={5}
        totalStages={6}
        showConnector={false}
      />
    )

    expect(screen.getByText(/Price alerts/i)).toBeInTheDocument()
    expect(screen.getByText(/RSI alerts/i)).toBeInTheDocument()
    expect(screen.getByText(/Traded stocks processed/i)).toBeInTheDocument()
  })
})
