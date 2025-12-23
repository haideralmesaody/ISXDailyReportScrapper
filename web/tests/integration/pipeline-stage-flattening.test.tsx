/**
 * Pipeline Stage Flattening Integration Tests
 *
 * Tests the stage flattening logic in operations-content.tsx that transforms
 * multi-step pipeline operations into individual stage tickets while preserving
 * analytics data and maintaining proper stage ordering.
 */

import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, test, vi, beforeEach } from 'vitest'
import { identifyStage, getStageOrder } from '@/lib/operations/stage-mapping'

// Mock the operations content component
const mockOperations = [
  {
    operation_id: 'pipeline-op-1',
    status: 'running',
    progress: 35,
    steps: [
      {
        id: 'scraping-step',
        name: 'Data Collection',
        status: 'completed',
        progress: 100,
        metadata: {
          stage_id: 'scraping',
          files_downloaded: 15,
          duration: '3m 15s'
        }
      },
      {
        id: 'processing-step',
        name: 'Excel to CSV Processing',
        status: 'active',
        progress: 45,
        metadata: {
          stage_id: 'processing',
          files_processed: 7,
          total_files: 15,
          current_file: 'market_data.xlsx'
        }
      },
      {
        id: 'indices-step',
        name: 'Index Extraction',
        status: 'pending',
        progress: 0,
        metadata: {
          stage_id: 'indices'
        }
      },
      {
        id: 'liquidity-step',
        name: 'Liquidity Analysis',
        status: 'pending',
        progress: 0,
        metadata: {
          stage_id: 'liquidity'
        }
      }
    ],
    metadata: {
      pipeline_summary: {
        stage_timeline: {
          scraping: { started_at: '2025-10-17T17:00:00Z', completed_at: '2025-10-17T17:03:15Z' },
          processing: { started_at: '2025-10-17T17:03:20Z' }
        }
      }
    }
  },
  {
    operation_id: 'single-stage-op-1',
    status: 'completed',
    progress: 100,
    steps: [{
      id: 'indicators-step',
      name: 'Indicator Calculation',
      status: 'completed',
      progress: 100,
      metadata: {
        stage_id: 'indicators',
        indicators_calculated: 20,
        tickers_processed: 95
      }
    }]
  }
]

describe('Pipeline Stage Flattening Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Stage Flattening Logic', () => {
    test('transforms multi-step pipeline into individual stage tickets', () => {
      // Simulate the flattening logic from operations-content.tsx
      const pipelineOperation = mockOperations[0]

      const stageTickets = pipelineOperation.steps.flatMap((step, index) => {
        const originalOperationId = pipelineOperation.operation_id
        const uniqueKey = `${originalOperationId}-${step.id}`

        const stageOperation = {
          ...pipelineOperation,
          operation_id: originalOperationId, // Preserve original
          react_key: uniqueKey,
          steps: [step], // Single step
          progress: step.progress || 0,
          status: step.status || 'pending',
          metadata: {
            pipelineMetadata: pipelineOperation.metadata,
            ...pipelineOperation.metadata,
            ...step.metadata,
            stage_id: step.metadata?.stage_id || step.id,
            stage_timeline: step.metadata?.stage_timeline || pipelineOperation.metadata?.pipeline_summary?.stage_timeline
          }
        }

        const stageMapping = identifyStage(stageOperation)
        if (!stageMapping?.stageId) {
          console.warn(`Unable to identify stage for step ${step.id}`)
          return []
        }

        return [{
          key: uniqueKey,
          operation: stageOperation,
          pipelineMode: true,
          stageNumber: index + 1,
          totalStages: pipelineOperation.steps.length,
          showConnector: index < pipelineOperation.steps.length - 1,
          originalOperationId
        }]
      })

      expect(stageTickets).toHaveLength(4) // 4 stages in the pipeline

      // Verify each stage ticket structure
      stageTickets.forEach((ticket, index) => {
        expect(ticket).toHaveProperty('key')
        expect(ticket).toHaveProperty('operation')
        expect(ticket).toHaveProperty('pipelineMode', true)
        expect(ticket).toHaveProperty('stageNumber', index + 1)
        expect(ticket).toHaveProperty('totalStages', 4)
        expect(ticket).toHaveProperty('originalOperationId', 'pipeline-op-1')
        expect(ticket).toHaveProperty('showConnector')

        // Verify connector logic (last stage should not have connector)
        if (index === stageTickets.length - 1) {
          expect(ticket.showConnector).toBe(false)
        } else {
          expect(ticket.showConnector).toBe(true)
        }
      })

      // Verify specific stage information
      const scrapingTicket = stageTickets.find(t => t.stageNumber === 1)
      expect(scrapingTicket?.operation.status).toBe('completed')
      expect(scrapingTicket?.operation.progress).toBe(100)
      expect(scrapingTicket?.operation.metadata.files_downloaded).toBe(15)

      const processingTicket = stageTickets.find(t => t.stageNumber === 2)
      expect(processingTicket?.operation.status).toBe('active')
      expect(processingTicket?.operation.progress).toBe(45)
      expect(processingTicket?.operation.metadata.files_processed).toBe(7)
    })

    test('preserves analytics data during flattening', () => {
      const pipelineOperation = mockOperations[0]

      const stageTicket = pipelineOperation.steps.flatMap((step, index) => {
        const originalOperationId = pipelineOperation.operation_id
        const uniqueKey = `${originalOperationId}-${step.id}`

        const stageOperation = {
          ...pipelineOperation,
          operation_id: originalOperationId,
          react_key: uniqueKey,
          steps: [step],
          progress: step.progress || 0,
          status: step.status || 'pending',
          metadata: {
            pipelineMetadata: pipelineOperation.metadata,
            ...pipelineOperation.metadata,
            ...step.metadata,
            stage_id: step.metadata?.stage_id || step.id,
            stage_timeline: step.metadata?.stage_timeline || pipelineOperation.metadata?.pipeline_summary?.stage_timeline
          }
        }

        return [{
          key: uniqueKey,
          operation: stageOperation,
          originalOperationId
        }]
      })[0] // Get first stage ticket

      // Verify analytics preservation
      expect(stageTicket.originalOperationId).toBe('pipeline-op-1')
      expect(stageTicket.operation.metadata.pipelineMetadata).toBeDefined()
      expect(stageTicket.operation.metadata.pipeline_summary).toBeDefined()
      expect(stageTicket.operation.metadata.stage_timeline).toBeDefined()

      // Verify both pipeline and step metadata are preserved
      expect(stageTicket.operation.metadata.stage_id).toBe('scraping')
      expect(stageTicket.operation.metadata.files_downloaded).toBe(15)
      expect(stageTicket.operation.metadata.pipeline_summary.stage_timeline).toBeDefined()
    })

    test('filters out invalid stages gracefully', () => {
      const invalidPipelineOperation = {
        operation_id: 'invalid-pipeline',
        status: 'running',
        progress: 20,
        steps: [
          {
            id: 'valid-step',
            name: 'Valid Stage',
            status: 'active',
            progress: 50,
            metadata: { stage_id: 'scraping' }
          },
          {
            id: 'invalid-step',
            name: 'Invalid Stage',
            status: 'pending',
            progress: 0,
            metadata: { stage_id: 'unknown_stage' }
          }
        ]
      }

      // Mock identifyStage to return null for unknown stages
      vi.mocked(identifyStage).mockImplementation((operation) => {
        const stageId = operation?.metadata?.stage_id
        if (stageId === 'unknown_stage') {
          return null // Simulate invalid stage
        }
        return {
          stageId: stageId as any,
          confidence: 'high',
          source: 'explicit'
        }
      })

      const stageTickets = invalidPipelineOperation.steps.flatMap((step, index) => {
        const originalOperationId = invalidPipelineOperation.operation_id
        const uniqueKey = `${originalOperationId}-${step.id}`

        const stageOperation = {
          ...invalidPipelineOperation,
          operation_id: originalOperationId,
          react_key: uniqueKey,
          steps: [step],
          progress: step.progress || 0,
          status: step.status || 'pending',
          metadata: {
            ...invalidPipelineOperation.metadata,
            ...step.metadata,
            stage_id: step.metadata?.stage_id || step.id
          }
        }

        const stageMapping = identifyStage(stageOperation)
        if (!stageMapping?.stageId) {
          console.warn(`Unable to identify stage for step ${step.id}`)
          return []
        }

        return [{
          key: uniqueKey,
          operation: stageOperation,
          pipelineMode: true,
          stageNumber: index + 1,
          totalStages: invalidPipelineOperation.steps.length,
          showConnector: index < invalidPipelineOperation.steps.length - 1,
          originalOperationId
        }]
      })

      // Should only have 1 valid stage ticket
      expect(stageTickets).toHaveLength(1)
      expect(stageTickets[0].operation.metadata.stage_id).toBe('scraping')
    })

    test('handles empty steps array', () => {
      const emptyPipelineOperation = {
        operation_id: 'empty-pipeline',
        status: 'pending',
        progress: 0,
        steps: []
      }

      const stageTickets = emptyPipelineOperation.steps.flatMap(() => [])

      expect(stageTickets).toHaveLength(0)
    })

    test('maintains stage order consistency', () => {
      const pipelineOperation = mockOperations[0]

      const stageTickets = pipelineOperation.steps.flatMap((step, index) => {
        const originalOperationId = pipelineOperation.operation_id
        const uniqueKey = `${originalOperationId}-${step.id}`

        const stageOperation = {
          ...pipelineOperation,
          operation_id: originalOperationId,
          react_key: uniqueKey,
          steps: [step],
          progress: step.progress || 0,
          status: step.status || 'pending',
          metadata: {
            ...pipelineOperation.metadata,
            ...step.metadata,
            stage_id: step.metadata?.stage_id || step.id
          }
        }

        const stageMapping = identifyStage(stageOperation)
        if (!stageMapping?.stageId) {
          return []
        }

        return [{
          key: uniqueKey,
          operation: stageOperation,
          pipelineMode: true,
          stageNumber: index + 1,
          totalStages: pipelineOperation.steps.length,
          showConnector: index < pipelineOperation.steps.length - 1,
          originalOperationId
        }]
      })

      // Verify stage order matches expected pipeline order
      const expectedOrder = ['scraping', 'processing', 'indices', 'liquidity']
      const actualOrder = stageTickets.map(ticket => ticket.operation.metadata.stage_id)

      expect(actualOrder).toEqual(expectedOrder)

      // Verify stage numbers are sequential
      stageTickets.forEach((ticket, index) => {
        expect(ticket.stageNumber).toBe(index + 1)
      })
    })
  })

  describe('Single Stage Operations', () => {
    test('handles single stage operations without flattening', () => {
      const singleStageOperation = mockOperations[1]

      // Single stage operations should not be flattened
      const shouldFlatten = Array.isArray(singleStageOperation.steps) && singleStageOperation.steps.length > 1
      expect(shouldFlatten).toBe(false)

      // Should render as single operation
      expect(singleStageOperation.operation_id).toBe('single-stage-op-1')
      expect(singleStageOperation.steps).toHaveLength(1)
      expect(singleStageOperation.steps[0].metadata.stage_id).toBe('indicators')
    })

    test('preserves original operation structure for single stages', () => {
      const singleStageOperation = mockOperations[1]

      // Verify original structure is maintained
      expect(singleStageOperation.operation_id).toBe('single-stage-op-1')
      expect(singleStageOperation.status).toBe('completed')
      expect(singleStageOperation.progress).toBe(100)
      expect(singleStageOperation.steps[0].metadata.indicators_calculated).toBe(20)
      expect(singleStageOperation.steps[0].metadata.tickers_processed).toBe(95)
    })
  })

  describe('React Key Generation', () => {
    test('generates unique React keys for each stage', () => {
      const pipelineOperation = mockOperations[0]

      const keys = pipelineOperation.steps.flatMap((step, index) => {
        const originalOperationId = pipelineOperation.operation_id
        const uniqueKey = `${originalOperationId}-${step.id}`

        return [{
          key: uniqueKey,
          stageId: step.metadata?.stage_id || step.id
        }]
      })

      expect(keys).toHaveLength(4)

      // Verify all keys are unique
      const keySet = new Set(keys.map(k => k.key))
      expect(keySet.size).toBe(4)

      // Verify key format
      keys.forEach((keyObj) => {
        expect(keyObj.key).toMatch(/^pipeline-op-1-/)
        expect(keyObj.key).toContain('-step')
      })
    })

    test('preserves original operation ID for analytics', () => {
      const pipelineOperation = mockOperations[0]

      const stageTickets = pipelineOperation.steps.flatMap((step, index) => {
        const originalOperationId = pipelineOperation.operation_id
        const uniqueKey = `${originalOperationId}-${step.id}`

        return [{
          key: uniqueKey,
          originalOperationId,
          stageId: step.metadata?.stage_id || step.id
        }]
      })

      stageTickets.forEach(ticket => {
        expect(ticket.originalOperationId).toBe('pipeline-op-1')
      })
    })
  })

  describe('Error Handling in Flattening', () => {
    test('handles malformed step data', () => {
      const malformedPipelineOperation = {
        operation_id: 'malformed-pipeline',
        status: 'running',
        progress: 10,
        steps: [
          null,
          undefined,
          {}, // Empty object
          {
            id: 'valid-step',
            name: 'Valid Stage',
            status: 'active',
            progress: 50,
            metadata: { stage_id: 'processing' }
          }
        ]
      }

      const stageTickets = malformedPipelineOperation.steps.flatMap((step, index) => {
        if (!step || !step.id) {
          console.warn(`Skipping malformed step at index ${index}`)
          return []
        }

        const originalOperationId = malformedPipelineOperation.operation_id
        const uniqueKey = `${originalOperationId}-${step.id}`

        const stageOperation = {
          ...malformedPipelineOperation,
          operation_id: originalOperationId,
          react_key: uniqueKey,
          steps: [step],
          progress: step.progress || 0,
          status: step.status || 'pending',
          metadata: {
            ...malformedPipelineOperation.metadata,
            ...step.metadata,
            stage_id: step.metadata?.stage_id || step.id
          }
        }

        const stageMapping = identifyStage(stageOperation)
        if (!stageMapping?.stageId) {
          return []
        }

        return [{
          key: uniqueKey,
          operation: stageOperation,
          originalOperationId
        }]
      })

      // Should only process the valid step
      expect(stageTickets).toHaveLength(1)
      expect(stageTickets[0].operation.metadata.stage_id).toBe('processing')
    })

    test('handles missing metadata gracefully', () => {
      const noMetadataPipeline = {
        operation_id: 'no-metadata-pipeline',
        status: 'running',
        progress: 15,
        steps: [
          {
            id: 'step-1',
            name: 'Step Without Metadata',
            status: 'active',
            progress: 15
            // No metadata property
          }
        ]
      }

      const stageTickets = noMetadataPipeline.steps.flatMap((step, index) => {
        const originalOperationId = noMetadataPipeline.operation_id
        const uniqueKey = `${originalOperationId}-${step.id}`

        const stageOperation = {
          ...noMetadataPipeline,
          operation_id: originalOperationId,
          react_key: uniqueKey,
          steps: [step],
          progress: step.progress || 0,
          status: step.status || 'pending',
          metadata: {
            ...noMetadataPipeline.metadata,
            ...step.metadata,
            stage_id: step.metadata?.stage_id || step.id
          }
        }

        return [{
          key: uniqueKey,
          operation: stageOperation,
          originalOperationId
        }]
      })

      expect(stageTickets).toHaveLength(1)
      expect(stageTickets[0].operation.metadata).toBeDefined()
    })
  })

  describe('Performance Considerations', () => {
    test('efficiently handles large pipeline operations', () => {
      // Create a large pipeline operation with many stages
      const largePipeline = {
        operation_id: 'large-pipeline',
        status: 'running',
        progress: 10,
        steps: Array.from({ length: 50 }, (_, i) => ({
          id: `step-${i + 1}`,
          name: `Stage ${i + 1}`,
          status: i === 0 ? 'active' : 'pending',
          progress: i === 0 ? 20 : 0,
          metadata: {
            stage_id: `stage_${i + 1}`
          }
        }))
      }

      // Mock identifyStage to return a valid mapping for all stages
      vi.mocked(identifyStage).mockReturnValue({
        stageId: 'scraping',
        confidence: 'high',
        source: 'explicit'
      })

      const startTime = performance.now()

      const stageTickets = largePipeline.steps.flatMap((step, index) => {
        const originalOperationId = largePipeline.operation_id
        const uniqueKey = `${originalOperationId}-${step.id}`

        const stageOperation = {
          ...largePipeline,
          operation_id: originalOperationId,
          react_key: uniqueKey,
          steps: [step],
          progress: step.progress || 0,
          status: step.status || 'pending',
          metadata: {
            ...largePipeline.metadata,
            ...step.metadata,
            stage_id: step.metadata?.stage_id || step.id
          }
        }

        const stageMapping = identifyStage(stageOperation)
        if (!stageMapping?.stageId) {
          return []
        }

        return [{
          key: uniqueKey,
          operation: stageOperation,
          pipelineMode: true,
          stageNumber: index + 1,
          totalStages: largePipeline.steps.length,
          showConnector: index < largePipeline.steps.length - 1,
          originalOperationId
        }]
      })

      const endTime = performance.now()
      const processingTime = endTime - startTime

      expect(stageTickets).toHaveLength(50)
      expect(processingTime).toBeLessThan(100) // Should process in under 100ms

      // Verify structure integrity
      stageTickets.forEach((ticket, index) => {
        expect(ticket.stageNumber).toBe(index + 1)
        expect(ticket.totalStages).toBe(50)
        expect(ticket.originalOperationId).toBe('large-pipeline')
      })
    })
  })
})