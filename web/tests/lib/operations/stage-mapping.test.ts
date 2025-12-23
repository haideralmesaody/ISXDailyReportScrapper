/**
 * Stage Mapping Library Tests
 *
 * Tests for the stage mapping library functions including stage identification,
 * ordering, activation logic, and helper utilities.
 */

import { describe, it, expect, test } from 'vitest'
import {
  identifyStage,
  getStageInfo,
  isStageReliable,
  getAllStages,
  getStageOrder,
  getPipelineStageOrder,
  isStageActive,
  canActivateStage,
  extractStageFileCount,
  getStageMetrics,
  STAGE_DEFINITIONS,
  type StageId
} from '@/lib/operations/stage-mapping'

describe('Stage Mapping Library', () => {
  describe('Stage Identification Edge Cases', () => {
    it('handles case-insensitive stage identification', () => {
      const testCases = [
        { stage_id: 'SCRAPING', expected: 'scraping' },
        { stage_id: 'Scraping', expected: 'scraping' },
        { stage_id: 'PROCESSING', expected: 'processing' },
        { stage_id: 'Processing', expected: 'processing' }
      ]

      testCases.forEach(({ stage_id, expected }) => {
        const result = identifyStage({ stage_id })
        expect(result.stageId).toBe(expected)
        expect(result.confidence).toBe('high')
        expect(result.source).toBe('explicit')
      })
    })

    it('handles whitespace trimming in stage identification', () => {
      const testCases = [
        { stage_id: '  scraping  ', expected: 'scraping' },
        { stage_id: '\tprocessing\n', expected: 'processing' },
        { stage_id: '  indices  ', expected: 'indices' }
      ]

      testCases.forEach(({ stage_id, expected }) => {
        const result = identifyStage({ stage_id })
        expect(result.stageId).toBe(expected)
        expect(result.confidence).toBe('high')
      })
    })

    it('identifies stage from nested metadata structures', () => {
      const complexOperation = {
        data: {
          target_stage_id: 'liquidity',
          operation_type: 'run_stage'
        },
        metadata: {
          pipeline_summary: {
            current_stage: 'liquidity'
          }
        },
        name: 'Complex Nested Operation'
      }

      const result = identifyStage(complexOperation)
      expect(result.stageId).toBe('liquidity')
      expect(result.confidence).toBe('high')
      expect(result.source).toBe('operation_type')
    })

    it('handles multiple possible stage identifiers with priority', () => {
      const operation = {
        stage_id: 'processing', // Highest priority
        metadata: {
          stage_id: 'scraping',
          operation_type: 'scraping'
        },
        name: 'Data Processing Operation'
      }

      const result = identifyStage(operation)
      expect(result.stageId).toBe('processing') // Should pick the first match
      expect(result.metadata?.origin).toBe('stage_id')
    })
  })

  describe('Pattern Matching', () => {
    it('matches complex operation names correctly', () => {
      const testCases = [
        {
          name: 'ISX Daily Market Data Scraping Operation',
          expected: 'scraping',
          confidence: 'medium'
        },
        {
          name: 'Excel to CSV Conversion and Processing',
          expected: 'processing',
          confidence: 'medium'
        },
        {
          name: 'Technical Analysis with RSI and MACD Indicators',
          expected: 'analysis',
          confidence: 'medium'
        },
        {
          name: 'Complete End-to-End Pipeline Execution',
          expected: 'full_pipeline',
          confidence: 'medium'
        },
        {
          name: 'Liquidity Metrics Calculation and Analysis',
          expected: 'liquidity',
          confidence: 'medium'
        }
      ]

      testCases.forEach(({ name, expected, confidence }) => {
        const result = identifyStage({ name })
        expect(result.stageId).toBe(expected)
        expect(result.confidence).toBe(confidence)
        expect(result.source).toBe('pattern_match')
      })
    })

    it('handles non-matching operation names', () => {
      const nonMatchingNames = [
        'Unknown Operation Type',
        'Custom Business Process',
        'System Maintenance Task',
        'Database Backup Operation'
      ]

      nonMatchingNames.forEach(name => {
        const result = identifyStage({ name })
        expect(result.stageId).toBe('scraping') // Fallback
        expect(result.confidence).toBe('low')
        expect(result.source).toBe('fallback')
      })
    })
  })

  describe('Stage Activation Logic', () => {
    it('correctly determines stage activation based on dependencies', () => {
      const mockStages = [
        {
          steps: [{ metadata: { stage_id: 'scraping' } }],
          status: 'completed'
        },
        {
          steps: [{ metadata: { stage_id: 'processing' } }],
          status: 'completed'
        },
        {
          steps: [{ metadata: { stage_id: 'indices' } }],
          status: 'running'
        }
      ]

      expect(isStageActive('scraping', mockStages)).toBe(true) // First stage always active
      expect(isStageActive('processing', mockStages)).toBe(true) // Previous stage completed
      expect(isStageActive('indices', mockStages)).toBe(true) // Currently running
      expect(isStageActive('liquidity', mockStages)).toBe(false) // Previous stage not completed
      expect(isStageActive('indicators', mockStages)).toBe(false) // Previous stage not completed
    })

    it('handles activation with different step structures', () => {
      const mockStages = [
        {
          steps: [{ id: 'scraping-step', metadata: { stage_id: 'scraping' } }],
          status: 'completed'
        }
      ]

      expect(isStageActive('processing', mockStages)).toBe(true)
      expect(isStageActive('processing', [])).toBe(false) // No previous stages
    })

    test('canActivateStage handles malformed input', () => {
      expect(canActivateStage(null, [])).toBe(false)
      expect(canActivateStage(undefined, [])).toBe(false)
      expect(canActivateStage({}, [])).toBe(false)
      expect(canActivateStage({ steps: [] }, [])).toBe(false)

      // Valid stage data
      const validStage = {
        steps: [{ metadata: { stage_id: 'processing' } }]
      }
      expect(canActivateStage(validStage, [])).toBe(false) // No previous completed stages
    })
  })

  describe('File Count Extraction', () => {
    it('extracts file counts from various metadata locations', () => {
      const testCases = [
        {
          operation: { metadata: { files_processed: 25 } },
          stageId: 'scraping' as StageId,
          expected: 25
        },
        {
          operation: { metadata: { files_downloaded: 15 } },
          stageId: 'scraping' as StageId,
          expected: 15
        },
        {
          operation: { metadata: { files_converted: 12 } },
          stageId: 'processing' as StageId,
          expected: 12
        },
        {
          operation: { metadata: { indices_extracted: 2 } },
          stageId: 'indices' as StageId,
          expected: 2
        },
        {
          operation: { metadata: { analysis_files: 3 } },
          stageId: 'analysis' as StageId,
          expected: 3
        }
      ]

      testCases.forEach(({ operation, stageId, expected }) => {
        expect(extractStageFileCount(operation, stageId)).toBe(expected)
      })
    })

    it('provides sensible defaults for different stages', () => {
      const defaults = {
        scraping: 15,
        data_processing: 15,
        processing: 15,
        indices: 1,
        liquidity: 1,
        indicators: 1,
        analysis: 1,
        full_pipeline: 30
      }

      Object.entries(defaults).forEach(([stageId, expectedDefault]) => {
        expect(extractStageFileCount({}, stageId as StageId)).toBe(expectedDefault)
      })
    })

    it('handles completed steps with file counts', () => {
      const operation = {
        steps: [
          {
            status: 'completed',
            metadata: { files_processed: 18 }
          },
          {
            status: 'failed',
            metadata: { files_processed: 5 }
          }
        ]
      }

      expect(extractStageFileCount(operation, 'processing')).toBe(18) // Should use completed step
    })

    it('handles zero and negative values', () => {
      const operation = {
        metadata: { files_processed: 0 }
      }

      expect(extractStageFileCount(operation, 'scraping')).toBe(15) // Should use default for 0
    })
  })

  describe('Stage Metrics Configuration', () => {
    it('provides appropriate metrics for each stage type', () => {
      const expectedMetrics = {
        scraping: {
          primaryMetric: 'files_processed',
          primaryLabel: 'Files Downloaded',
          secondaryMetric: 'duration',
          secondaryLabel: 'Duration'
        },
        processing: {
          primaryMetric: 'files_processed',
          primaryLabel: 'Files Processed',
          secondaryMetric: 'duration',
          secondaryLabel: 'Duration'
        },
        indices: {
          primaryMetric: 'indices_extracted',
          primaryLabel: 'Indices Extracted',
          secondaryMetric: 'records_processed',
          secondaryLabel: 'Records Processed'
        },
        liquidity: {
          primaryMetric: 'analysis_files',
          primaryLabel: 'Analysis Files',
          secondaryMetric: 'companies_analyzed',
          secondaryLabel: 'Companies Analyzed'
        },
        indicators: {
          primaryMetric: 'indicators_calculated',
          primaryLabel: 'Indicators Calculated',
          secondaryMetric: 'tickers_processed',
          secondaryLabel: 'Tickers Processed'
        },
        analysis: {
          primaryMetric: 'analysis_complete',
          primaryLabel: 'Analysis Complete',
          secondaryMetric: 'insights_generated',
          secondaryLabel: 'Insights Generated'
        }
      }

      Object.entries(expectedMetrics).forEach(([stageId, expected]) => {
        const metrics = getStageMetrics(stageId as StageId)
        expect(metrics).toEqual(expected)
      })
    })

    it('returns default metrics for unknown stages', () => {
      const metrics = getStageMetrics('unknown' as StageId)
      expect(metrics.primaryMetric).toBe('files_processed')
      expect(metrics.primaryLabel).toBe('Files Downloaded')
    })
  })

  describe('Pipeline Stage Ordering', () => {
    it('maintains correct pipeline order consistency', () => {
      const pipelineOrder = getPipelineStageOrder()
      const expectedOrder = ['scraping', 'data_processing', 'processing', 'indices', 'liquidity', 'indicators', 'analysis']

      expect(pipelineOrder).toEqual(expectedOrder)

      // Verify numerical ordering through getStageOrder
      let previousOrder = -1
      pipelineOrder.forEach(stageId => {
        const currentOrder = getStageOrder(stageId as StageId)
        expect(currentOrder).toBeGreaterThan(previousOrder)
        previousOrder = currentOrder
      })
    })

    it('excludes wrapper operations from pipeline order', () => {
      const pipelineOrder = getPipelineStageOrder()
      expect(pipelineOrder).not.toContain('full_pipeline')
    })

    it('provides correct order numbers for all stages', () => {
      const expectedOrders = {
        scraping: 1,
        data_processing: 1.5,
        processing: 2,
        indices: 3,
        liquidity: 4,
        indicators: 5,
        analysis: 6
      }

      Object.entries(expectedOrders).forEach(([stageId, expectedOrder]) => {
        expect(getStageOrder(stageId as StageId)).toBe(expectedOrder)
      })
    })
  })

  describe('Stage Information Consistency', () => {
    it('provides complete information for all stages', () => {
      const allStages = getAllStages()

      allStages.forEach(stage => {
        expect(stage).toHaveProperty('id')
        expect(stage).toHaveProperty('name')
        expect(stage).toHaveProperty('description')
        expect(stage).toHaveProperty('category')
        expect(stage).toHaveProperty('color')
        expect(stage).toHaveProperty('icon')
        expect(stage).toHaveProperty('order')

        expect(typeof stage.id).toBe('string')
        expect(typeof stage.name).toBe('string')
        expect(typeof stage.description).toBe('string')
        expect(typeof stage.category).toBe('string')
        expect(typeof stage.color).toBe('string')
        expect(typeof stage.icon).toBe('string')
        expect(typeof stage.order).toBe('number')

        expect(stage.order).toBeGreaterThan(0)
        expect(stage.name.length).toBeGreaterThan(0)
        expect(stage.description.length).toBeGreaterThan(0)
      })
    })

    it('returns consistent information between getStageInfo and getAllStages', () => {
      const allStages = getAllStages()

      allStages.forEach(stage => {
        const stageInfo = getStageInfo(stage.id as StageId)
        expect(stageInfo.name).toBe(stage.name)
        expect(stageInfo.description).toBe(stage.description)
        expect(stageInfo.category).toBe(stage.category)
        expect(stageInfo.color).toBe(stage.color)
        expect(stageInfo.icon).toBe(stage.icon)
        expect(stageInfo.order).toBe(stage.order)
      })
    })
  })

  describe('Stage Reliability Assessment', () => {
    it('correctly assesses reliability based on confidence and source', () => {
      const testCases = [
        {
          confidence: 'high' as const,
          source: 'explicit' as const,
          expected: true
        },
        {
          confidence: 'medium' as const,
          source: 'pattern_match' as const,
          expected: true
        },
        {
          confidence: 'medium' as const,
          source: 'name_analysis' as const,
          expected: false
        },
        {
          confidence: 'low' as const,
          source: 'fallback' as const,
          expected: false
        }
      ]

      testCases.forEach(({ confidence, source, expected }) => {
        const mapping = { stageId: 'scraping' as StageId, confidence, source }
        expect(isStageReliable(mapping)).toBe(expected)
      })
    })
  })

  describe('Error Handling and Robustness', () => {
    it('handles null and undefined inputs gracefully', () => {
      expect(() => identifyStage(null)).not.toThrow()
      expect(() => identifyStage(undefined)).not.toThrow()
      expect(() => getStageInfo(null as any)).not.toThrow()
      expect(() => getStageOrder(null as any)).not.toThrow()
      expect(() => extractStageFileCount(null as any, 'scraping')).not.toThrow()
    })

    it('handles circular object references', () => {
      const circular: any = { name: 'circular operation' }
      circular.self = circular

      expect(() => identifyStage(circular)).not.toThrow()
      const result = identifyStage(circular)
      expect(result.stageId).toBe('scraping') // Should fallback gracefully
    })

    it('handles extremely large strings', () => {
      const largeString = 'a'.repeat(10000)
      const operation = { name: largeString }

      expect(() => identifyStage(operation)).not.toThrow()
      const result = identifyStage(operation)
      expect(result.confidence).toBe('low') // Should not match any patterns
    })

    it('handles special characters and unicode', () => {
      const specialNames = [
        'Data Collection 🏗️',
        'Обработка данных', // Russian
        'معالجة البيانات', // Arabic
        '数据处理', // Chinese
        'Procesamiento\ndatos\ncon\nsaltos\ndelínea'
      ]

      specialNames.forEach(name => {
        expect(() => identifyStage({ name })).not.toThrow()
      })
    })
  })

  describe('Performance Characteristics', () => {
    it('performs efficiently with large operation objects', () => {
      const largeOperation = {
        operation_id: 'large-op',
        name: 'Large Operation',
        metadata: Object.fromEntries(
          Array.from({ length: 1000 }, (_, i) => [`key_${i}`, `value_${i}`])
        ),
        steps: Array.from({ length: 100 }, (_, i) => ({
          id: `step_${i}`,
          name: `Step ${i}`,
          metadata: {
            stage_id: 'scraping',
            step_data: Array.from({ length: 100 }, (_, j) => (`step_data_${i}_${j}`))
          }
        }))
      }

      const startTime = performance.now()
      const result = identifyStage(largeOperation)
      const endTime = performance.now()

      expect(endTime - startTime).toBeLessThan(50) // Should process in under 50ms
      expect(result.stageId).toBe('scraping')
      expect(result.confidence).toBe('high')
    })

    it('efficiently processes many stage identification calls', () => {
      const operations = Array.from({ length: 1000 }, (_, i) => ({
        operation_id: `op_${i}`,
        name: `Operation ${i}`,
        metadata: {
          stage_id: 'processing'
        }
      }))

      const startTime = performance.now()
      const results = operations.map(op => identifyStage(op))
      const endTime = performance.now()

      expect(endTime - startTime).toBeLessThan(100) // Should process 1000 operations in under 100ms
      expect(results).toHaveLength(1000)
      results.forEach(result => {
        expect(result.stageId).toBe('processing')
        expect(result.confidence).toBe('high')
      })
    })
  })
})