/**
 * UnifiedOperationProgress Integration Tests
 *
 * Comprehensive integration tests for the unified operation progress component
 * with pipeline mode, stage activation, and KPI display functionality.
 */

import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { describe, it, expect, test, beforeEach, jest } from '@jest/globals'
import UnifiedOperationProgress from '../UnifiedOperationProgress'
import { identifyStage, getStageInfo, isStageReliable } from '@/lib/operations/stage-mapping'

// Mock the stage mapping utilities
jest.mock('@/lib/operations/stage-mapping', () => ({
  identifyStage: jest.fn(),
  getStageInfo: jest.fn(),
  isStageReliable: jest.fn()
}))

const mockIdentifyStage = identifyStage as unknown as jest.Mock
const mockGetStageInfo = getStageInfo as unknown as jest.Mock
const mockIsStageReliable = isStageReliable as unknown as jest.Mock

describe('UnifiedOperationProgress Integration', () => {
  beforeEach(() => {
    mockIdentifyStage.mockReset()
    mockGetStageInfo.mockReset()
    mockIsStageReliable.mockReset()
  })

  describe('Single Stage Operations', () => {
    it('renders single stage operation correctly', () => {
      const operation = {
        operation_id: 'test-op-1',
        status: 'running',
        progress: 45,
        steps: [{
          id: 'scraping-step',
          name: 'Data Collection',
          status: 'active',
          progress: 45
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'scraping',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Data Collection',
        description: 'Download ISX daily reports',
        category: 'data_acquisition',
        color: 'blue',
        icon: 'Download',
        order: 1
      })

      mockIsStageReliable.mockReturnValue(true)

      render(<UnifiedOperationProgress operation={operation} />)

      expect(screen.getByText('Data Collection')).toBeInTheDocument()
      expect(screen.getByText('45%')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /next operation/i })).toBeInTheDocument()
    })

    it('handles completed single stage operation', () => {
      const operation = {
        operation_id: 'test-op-2',
        status: 'completed',
        progress: 100,
        steps: [{
          id: 'processing-step',
          name: 'Excel to CSV Processing',
          status: 'completed',
          progress: 100,
          metadata: {
            files_processed: 15,
            duration: '2m 30s'
          }
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'processing',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Excel to CSV Processing',
        description: 'Convert Excel files to CSV format',
        category: 'data_transformation',
        color: 'green',
        icon: 'FileSpreadsheet',
        order: 2
      })

      mockIsStageReliable.mockReturnValue(true)

      render(<UnifiedOperationProgress operation={operation} />)

      expect(screen.getByText('Excel to CSV Processing')).toBeInTheDocument()
      expect(screen.getByText(/15 files processed/i)).toBeInTheDocument()
      expect(screen.getByText(/2m 30s/i)).toBeInTheDocument()
    })

    it('handles failed operation with error display', () => {
      const operation = {
        operation_id: 'test-op-3',
        status: 'failed',
        progress: 30,
        error: 'Connection timeout to ISX website',
        steps: [{
          id: 'scraping-step',
          name: 'Data Collection',
          status: 'failed',
          progress: 30,
          error: 'Connection timeout'
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'scraping',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Data Collection',
        description: 'Download ISX daily reports',
        category: 'data_acquisition',
        color: 'blue',
        icon: 'Download',
        order: 1
      })

      mockIsStageReliable.mockReturnValue(true)

      render(<UnifiedOperationProgress operation={operation} />)

      expect(screen.getByText('Connection timeout')).toBeInTheDocument()
      expect(screen.getByText(/failed/i)).toBeInTheDocument()
    })
  })

  describe('Pipeline Mode Operations', () => {
    it('renders pipeline stage with stage number and context', () => {
      const operation = {
        operation_id: 'pipeline-op-1',
        react_key: 'pipeline-op-1-processing',
        status: 'running',
        progress: 60,
        steps: [{
          id: 'processing-step',
          name: 'Excel to CSV Processing',
          status: 'active',
          progress: 60,
          metadata: {
            stage_id: 'processing',
            files_processed: 12,
            total_files: 15
          }
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'processing',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Excel to CSV Processing',
        description: 'Convert Excel files to CSV format',
        category: 'data_transformation',
        color: 'green',
        icon: 'FileSpreadsheet',
        order: 2
      })

      mockIsStageReliable.mockReturnValue(true)

      render(
        <UnifiedOperationProgress
          operation={operation}
          pipelineMode={true}
          stageNumber={2}
          totalStages={4}
          showConnector={true}
        />
      )

      expect(screen.getByText('Stage 2 of 4')).toBeInTheDocument()
      expect(screen.getByText('Excel to CSV Processing')).toBeInTheDocument()
      expect(screen.getByText(/12.*15.*files/i)).toBeInTheDocument()
      expect(screen.getByRole('separator')).toBeInTheDocument() // Pipeline connector
    })

    it('shows pipeline connector only when showConnector is true', () => {
      const operation = {
        operation_id: 'pipeline-op-2',
        react_key: 'pipeline-op-2-indices',
        status: 'pending',
        progress: 0,
        steps: [{
          id: 'indices-step',
          name: 'Index Extraction',
          status: 'pending',
          progress: 0
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'indices',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Index Extraction',
        description: 'Extract ISX60 and ISX15 indices',
        category: 'data_extraction',
        color: 'purple',
        icon: 'BarChart3',
        order: 3
      })

      mockIsStageReliable.mockReturnValue(true)

      const { rerender } = render(
        <UnifiedOperationProgress
          operation={operation}
          pipelineMode={true}
          stageNumber={3}
          totalStages={5}
          showConnector={true}
        />
      )

      expect(screen.getByRole('separator')).toBeInTheDocument()

      rerender(
        <UnifiedOperationProgress
          operation={operation}
          pipelineMode={true}
          stageNumber={3}
          totalStages={5}
          showConnector={false}
        />
      )

      expect(screen.queryByRole('separator')).not.toBeInTheDocument()
    })

    it('activates next steps only for final completed stage', () => {
      const operation = {
        operation_id: 'pipeline-op-final',
        react_key: 'pipeline-op-final-analysis',
        status: 'completed',
        progress: 100,
        steps: [{
          id: 'analysis-step',
          name: 'Market & Technical Analysis',
          status: 'completed',
          progress: 100
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'analysis',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Market & Technical Analysis',
        description: 'Comprehensive analysis including market trends and technical indicators',
        category: 'analysis',
        color: 'indigo',
        icon: 'TrendingUp',
        order: 6
      })

      mockIsStageReliable.mockReturnValue(true)

      const mockOnNextOperation = vi.fn()

      render(
        <UnifiedOperationProgress
          operation={operation}
          pipelineMode={true}
          stageNumber={5}
          totalStages={5} // Final stage
          showConnector={false}
          onNextOperation={mockOnNextOperation}
        />
      )

      const nextButton = screen.getByRole('button', { name: /next operation/i })
      expect(nextButton).toBeInTheDocument()
      expect(nextButton).not.toBeDisabled()

      fireEvent.click(nextButton)
      expect(mockOnNextOperation).toHaveBeenCalledWith('scraping')
    })

    it('disables next steps for non-final completed stages', () => {
      const operation = {
        operation_id: 'pipeline-op-middle',
        react_key: 'pipeline-op-middle-indices',
        status: 'completed',
        progress: 100,
        steps: [{
          id: 'indices-step',
          name: 'Index Extraction',
          status: 'completed',
          progress: 100
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'indices',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Index Extraction',
        description: 'Extract ISX60 and ISX15 indices',
        category: 'data_extraction',
        color: 'purple',
        icon: 'BarChart3',
        order: 3
      })

      mockIsStageReliable.mockReturnValue(true)

      render(
        <UnifiedOperationProgress
          operation={operation}
          pipelineMode={true}
          stageNumber={3}
          totalStages={5} // Not final stage
          showConnector={true}
        />
      )

      expect(screen.queryByRole('button', { name: /next operation/i })).not.toBeInTheDocument()
    })
  })

  describe('KPI Display', () => {
    it('displays rich KPI information for live data', () => {
      const operation = {
        operation_id: 'kpi-op-1',
        status: 'running',
        progress: 75,
        steps: [{
          id: 'indicators-step',
          name: 'Indicator Calculation',
          status: 'active',
          progress: 75,
          metadata: {
            stage_id: 'indicators',
            indicators_calculated: 18,
            total_indicators: 20,
            tickers_processed: 85,
            total_tickers: 95,
            current_indicator: 'RSI',
            estimated_completion: '2m 15s'
          }
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'indicators',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Indicator Calculation',
        description: 'Pre-compute indicators for SSOT',
        category: 'calculation',
        color: 'pink',
        icon: 'Activity',
        order: 5
      })

      mockIsStageReliable.mockReturnValue(true)

      render(<UnifiedOperationProgress operation={operation} />)

      expect(screen.getByText(/18.*20.*indicators/i)).toBeInTheDocument()
      expect(screen.getByText(/85.*95.*tickers/i)).toBeInTheDocument()
      expect(screen.getByText(/calculating RSI/i)).toBeInTheDocument()
      expect(screen.getByText(/2m 15s/i)).toBeInTheDocument()
    })

    it('falls back to configuration-based KPIs for missing metadata', () => {
      const operation = {
        operation_id: 'kpi-op-2',
        status: 'pending',
        progress: 0,
        steps: [{
          id: 'liquidity-step',
          name: 'Liquidity Analysis',
          status: 'pending',
          progress: 0
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'liquidity',
        confidence: 'medium',
        source: 'pattern_match'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Liquidity Analysis',
        description: 'Calculate liquidity metrics',
        category: 'analysis',
        color: 'orange',
        icon: 'Zap',
        order: 4
      })

      mockIsStageReliable.mockReturnValue(true)

      render(<UnifiedOperationProgress operation={operation} />)

      expect(screen.getByText(/ready to analyze/i)).toBeInTheDocument()
      expect(screen.getByText(/pending/i)).toBeInTheDocument()
    })

    it('displays different KPI patterns for different operation types', () => {
      const testCases = [
        {
          stageId: 'scraping',
          metadata: {
            files_downloaded: 12,
            total_files: 15,
            current_file: 'market_summary_2025-10-17.xlsx'
          },
          expectedTexts: [/12.*15.*files/i, /downloading market_summary/i]
        },
        {
          stageId: 'processing',
          metadata: {
            files_processed: 10,
            total_files: 15,
            processing_rate: '3 files/min'
          },
          expectedTexts: [/10.*15.*files/i, /3 files.*min/i]
        },
        {
          stageId: 'analysis',
          metadata: {
            analysis_complete: true,
            insights_generated: 8,
            processing_time: '45s'
          },
          expectedTexts: [/analysis complete/i, /8 insights/i]
        }
      ]

      testCases.forEach(({ stageId, metadata, expectedTexts }) => {
        const operation = {
          operation_id: `kpi-test-${stageId}`,
          status: 'running',
          progress: 50,
          steps: [{
            id: `${stageId}-step`,
            name: `Test ${stageId}`,
            status: 'active',
            progress: 50,
            metadata: { stage_id: stageId, ...metadata }
          }]
        }

        mockIdentifyStage.mockReturnValue({
          stageId,
          confidence: 'high',
          source: 'explicit'
        })

        mockGetStageInfo.mockReturnValue({
          name: `Test ${stageId}`,
          description: `Test description for ${stageId}`,
          category: 'test',
          color: 'blue',
          icon: 'Test',
          order: 1
        })

        mockIsStageReliable.mockReturnValue(true)

        const { unmount } = render(<UnifiedOperationProgress operation={operation} />)

        expectedTexts.forEach(expectedText => {
          expect(screen.getByText(expectedText)).toBeInTheDocument()
        })

        unmount()
      })
    })
  })

  describe('Error Handling and Edge Cases', () => {
    it('handles operations without steps gracefully', () => {
      const operation = {
        operation_id: 'no-steps-op',
        status: 'running',
        progress: 25
        // No steps array
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'scraping',
        confidence: 'low',
        source: 'fallback'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Data Collection',
        description: 'Download ISX daily reports',
        category: 'data_acquisition',
        color: 'blue',
        icon: 'Download',
        order: 1
      })

      mockIsStageReliable.mockReturnValue(false)

      render(<UnifiedOperationProgress operation={operation} />)

      // Should not crash and should show basic information
      expect(screen.getByText('25%')).toBeInTheDocument()
    })

    it('handles malformed step data', () => {
      const operation = {
        operation_id: 'malformed-op',
        status: 'running',
        progress: 30,
        steps: [
          null,
          undefined,
          {}, // Empty object
          { id: 'valid-step', name: 'Valid Step', status: 'active', progress: 30 }
        ]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'scraping',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Data Collection',
        description: 'Download ISX daily reports',
        category: 'data_acquisition',
        color: 'blue',
        icon: 'Download',
        order: 1
      })

      mockIsStageReliable.mockReturnValue(true)

      render(<UnifiedOperationProgress operation={operation} />)

      // Should handle malformed steps and still render the valid one
      expect(screen.getByText('Valid Step')).toBeInTheDocument()
      expect(screen.getByText('30%')).toBeInTheDocument()
    })

    it('handles stage mapping failures gracefully', () => {
      const operation = {
        operation_id: 'mapping-fail-op',
        status: 'running',
        progress: 40,
        steps: [{
          id: 'unknown-step',
          name: 'Unknown Operation',
          status: 'active',
          progress: 40
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'scraping',
        confidence: 'low',
        source: 'fallback'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Data Collection',
        description: 'Download ISX daily reports',
        category: 'data_acquisition',
        color: 'blue',
        icon: 'Download',
        order: 1
      })

      mockIsStageReliable.mockReturnValue(false)

      render(<UnifiedOperationProgress operation={operation} />)

      // Should still render with fallback information
      expect(screen.getByText('Unknown Operation')).toBeInTheDocument()
      expect(screen.getByText('40%')).toBeInTheDocument()
    })

    test('handles rapid prop changes', async () => {
      const operation = {
        operation_id: 'rapid-change-op',
        status: 'pending',
        progress: 0,
        steps: [{
          id: 'test-step',
          name: 'Test Operation',
          status: 'pending',
          progress: 0
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'scraping',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Test Operation',
        description: 'Test description',
        category: 'test',
        color: 'blue',
        icon: 'Test',
        order: 1
      })

      mockIsStageReliable.mockReturnValue(true)

      const { rerender } = render(<UnifiedOperationProgress operation={operation} />)

      // Rapidly change states
      const activeOperation = {
        ...operation,
        status: 'running',
        progress: 50,
        steps: [{ ...operation.steps[0], status: 'active', progress: 50 }]
      }

      const completedOperation = {
        ...operation,
        status: 'completed',
        progress: 100,
        steps: [{ ...operation.steps[0], status: 'completed', progress: 100 }]
      }

      rerender(<UnifiedOperationProgress operation={activeOperation} />)
      expect(screen.getByText('50%')).toBeInTheDocument()

      rerender(<UnifiedOperationProgress operation={completedOperation} />)
      expect(screen.getByText('100%')).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('provides proper ARIA labels and roles', () => {
      const operation = {
        operation_id: 'a11y-op',
        status: 'running',
        progress: 60,
        steps: [{
          id: 'test-step',
          name: 'Test Operation',
          status: 'active',
          progress: 60
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'scraping',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Test Operation',
        description: 'Test description',
        category: 'test',
        color: 'blue',
        icon: 'Test',
        order: 1
      })

      mockIsStageReliable.mockReturnValue(true)

      render(<UnifiedOperationProgress operation={operation} />)

      // Check for proper ARIA labels
      expect(screen.getByRole('progressbar')).toBeInTheDocument()
      expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '60')
      expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuemin', '0')
      expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuemax', '100')
    })

    it('announces status changes to screen readers', () => {
      const operation = {
        operation_id: 'announce-op',
        status: 'completed',
        progress: 100,
        steps: [{
          id: 'test-step',
          name: 'Test Operation',
          status: 'completed',
          progress: 100
        }]
      }

      mockIdentifyStage.mockReturnValue({
        stageId: 'scraping',
        confidence: 'high',
        source: 'explicit'
      })

      mockGetStageInfo.mockReturnValue({
        name: 'Test Operation',
        description: 'Test description',
        category: 'test',
        color: 'blue',
        icon: 'Test',
        order: 1
      })

      mockIsStageReliable.mockReturnValue(true)

      render(<UnifiedOperationProgress operation={operation} />)

      // Check for live region announcements
      expect(screen.getByText(/completed/i)).toBeInTheDocument()
    })
  })
})
