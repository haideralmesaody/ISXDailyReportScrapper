import '@testing-library/jest-dom'
import { render, screen, within } from '@testing-library/react'
import React from 'react'
import UnifiedOperationProgress from '@/components/operations/UnifiedOperationProgress'

type StageMetadata = {
  stage_id: string
  phase: string
  phase_message: string
  progress_percent?: number
  files_processed?: number
  total_files?: number
  telemetry_missing?: boolean
}

const buildStageOperation = (
  stageId: string,
  status: string,
  metadata: StageMetadata,
  stageNumber: number,
  totalStages: number
) => ({
  operation: {
    operation_id: `op-${stageId}`,
    name: `Stage ${stageId}`,
    status,
    progress: metadata.progress_percent ?? 0,
    message: metadata.phase_message,
    error: null,
    metadata,
    steps: [
      {
        id: stageId,
        name: stageId,
        status,
        progress: metadata.progress_percent ?? 0,
        metadata
      }
    ],
    stageContext: {
      allStages: [],
      pipelineMetadata: {},
      completedPipeline: false
    },
    pipelineMessage: '',
    stepData: {
      id: stageId,
      status,
      progress: metadata.progress_percent ?? 0,
      metadata
    },
    updated_at: new Date().toISOString()
  },
  stageNumber,
  totalStages
})

const PipelineHarness = ({
  stages
}: {
  stages: ReturnType<typeof buildStageOperation>[]
}) => (
  <>
    {stages.map(({ operation, stageNumber, totalStages }) => (
      <UnifiedOperationProgress
        key={operation.operation_id}
        operation={operation}
        pipelineMode
        stageNumber={stageNumber}
        totalStages={totalStages}
        showConnector={stageNumber < totalStages}
      />
    ))}
  </>
)

describe('Pipeline telemetry contract', () => {
  it('keeps future stages ready until telemetry arrives', () => {
    const scraping = buildStageOperation(
      'scraping',
      'running',
      {
        stage_id: 'scraping',
        phase: 'downloading',
        phase_message: 'Downloading files',
        progress_percent: 65,
        files_processed: 5,
        total_files: 10
      },
      1,
      2
    )

    const processingPending = buildStageOperation(
      'processing',
      'pending',
      {
        stage_id: 'processing',
        phase: 'pending',
        phase_message: 'Awaiting start',
        telemetry_missing: true
      },
      2,
      2
    )

    const { rerender } = render(
      <PipelineHarness stages={[scraping, processingPending]} />
    )

    const runningStages = screen.getAllByText(/Stage running/i)
    expect(runningStages).toHaveLength(1)

    const processingCard = screen.getByTestId('stage-card-processing')
    expect(within(processingCard).getByText(/Stage ready/i)).toBeInTheDocument()
    expect(
      within(processingCard).getByText(/Awaiting telemetry/i)
    ).toBeInTheDocument()

    const processingRunning = buildStageOperation(
      'processing',
      'running',
      {
        stage_id: 'processing',
        phase: 'reading',
        phase_message: 'Working on processing',
        progress_percent: 25,
        files_processed: 2,
        total_files: 10
      },
      2,
      2
    )

    rerender(<PipelineHarness stages={[scraping, processingRunning]} />)

    expect(screen.getAllByText(/Stage running/i)).toHaveLength(2)
    expect(
      within(processingCard).queryByText(/Awaiting telemetry/i)
    ).not.toBeInTheDocument()
  })
})
