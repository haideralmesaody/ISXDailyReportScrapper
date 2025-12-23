/**
 * Stage Mapping Utility Tests
 *
 * Comprehensive tests for stage identification, mapping, and helper functions.
 * Ensures reliable stage attribution and proper pipeline ordering.
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
  type StageMappingResult
} from '../stage-mapping'

describe('Stage Mapping Utilities', () => {
  describe('identifyStage', () => {
    it('identifies stage from explicit stage_id', () => {
      const operation = {
        stage_id: 'scraping',
        name: 'Data Collection Operation'
      }

      const result = identifyStage(operation)
      expect(result.stageId).toBe('scraping')
      expect(result.confidence).toBe('high')
      expect(result.source).toBe('explicit')
      expect(result.metadata?.origin).toBe('stage_id')
    })

    it('identifies stage from metadata.stage_id', () => {
      const operation = {
        metadata: {
          stage_id: 'processing',
          operation_type: 'run_stage'
        },
        name: 'Process CSV Files'
      }

      const result = identifyStage(operation)
      expect(result.stageId).toBe('processing')
      expect(result.confidence).toBe('high')
      expect(result.source).toBe('explicit')
    })

    it('identifies stage from operation_type mapping', () => {
      const operation = {
        metadata: {
          operation_type: 'scraping'
        },
        name: 'Download ISX Reports'
      }

      const result = identifyStage(operation)
      expect(result.stageId).toBe('scraping')
      expect(result.confidence).toBe('high')
      expect(result.source).toBe('operation_type')
    })

    it('identifies stage from run_stage operation', () => {
      const operation = {
        metadata: {
          operation: 'run_stage',
          target_stage_id: 'indices'
        },
        name: 'Extract ISX60 Indices'
      }

      const result = identifyStage(operation)
      expect(result.stageId).toBe('indices')
      expect(result.confidence).toBe('high')
      expect(result.source).toBe('operation_type')
    })

    it('identifies stage from name pattern matching', () => {
      const operation = {
        name: 'Liquidity Analysis for Market Data'
      }

      const result = identifyStage(operation)
      expect(result.stageId).toBe('liquidity')
      expect(result.confidence).toBe('medium')
      expect(result.source).toBe('pattern_match')
    })

    it('handles steps array with stage metadata', () => {
      const operation = {
        steps: [
          {
            id: 'step-1',
            metadata: {
              stage_id: 'indicators'
            }
          }
        ]
      }

      const result = identifyStage(operation)
      expect(result.stageId).toBe('indicators')
      expect(result.confidence).toBe('high')
      expect(result.metadata?.stepIndex).toBe(0)
    })

    it('returns fallback for unknown operations', () => {
      const operation = {
        name: 'Unknown Operation Type'
      }

      const result = identifyStage(operation)
      expect(result.stageId).toBe('scraping')
      expect(result.confidence).toBe('low')
      expect(result.source).toBe('fallback')
    })

    it('handles malformed input gracefully', () => {
      const result1 = identifyStage(null)
      const result2 = identifyStage(undefined)
      const result3 = identifyStage({})

      expect(result1.stageId).toBe('scraping')
      expect(result2.stageId).toBe('scraping')
      expect(result3.stageId).toBe('scraping')
    })
  })

  describe('getStageInfo', () => {
    it('returns correct stage information', () => {
      const info = getStageInfo('scraping')
      expect(info.name).toBe('Data Collection')
      expect(info.description).toBe('Download ISX daily reports')
      expect(info.category).toBe('data_acquisition')
      expect(info.color).toBe('blue')
      expect(info.order).toBe(1)
    })

    it('returns default stage for unknown ID', () => {
      const info = getStageInfo('unknown' as any)
      expect(info.name).toBe('Data Collection') // Default to scraping
      expect(info.description).toBe('Download ISX daily reports')
    })
  })

  describe('isStageReliable', () => {
    it('returns true for high confidence mappings', () => {
      const mapping: StageMappingResult = {
        stageId: 'scraping',
        confidence: 'high',
        source: 'explicit'
      }

      expect(isStageReliable(mapping)).toBe(true)
    })

    it('returns true for medium confidence pattern matches', () => {
      const mapping: StageMappingResult = {
        stageId: 'processing',
        confidence: 'medium',
        source: 'pattern_match'
      }

      expect(isStageReliable(mapping)).toBe(true)
    })

    it('returns false for low confidence mappings', () => {
      const mapping: StageMappingResult = {
        stageId: 'scraping',
        confidence: 'low',
        source: 'fallback'
      }

      expect(isStageReliable(mapping)).toBe(false)
    })

    it('returns false for medium confidence non-pattern matches', () => {
      const mapping: StageMappingResult = {
        stageId: 'scraping',
        confidence: 'medium',
        source: 'name_analysis'
      }

      expect(isStageReliable(mapping)).toBe(false)
    })
  })

  describe('getAllStages', () => {
    it('returns all available stages with complete information', () => {
      const stages = getAllStages()

      expect(stages).toHaveLength(9) // All defined stages
      expect(stages[0]).toHaveProperty('id')
      expect(stages[0]).toHaveProperty('name')
      expect(stages[0]).toHaveProperty('description')
      expect(stages[0]).toHaveProperty('category')
      expect(stages[0]).toHaveProperty('color')
      expect(stages[0]).toHaveProperty('icon')
      expect(stages[0]).toHaveProperty('order')

      // Check specific stages
      const scrapingStage = stages.find(s => s.id === 'scraping')
      expect(scrapingStage?.name).toBe('Data Collection')

      const analysisStage = stages.find(s => s.id === 'analysis')
      expect(analysisStage?.name).toBe('Market & Technical Analysis')
    })
  })

  describe('getStageOrder', () => {
    it('returns correct order number for stages', () => {
      expect(getStageOrder('scraping')).toBe(1)
      expect(getStageOrder('data_processing')).toBe(1.5)
      expect(getStageOrder('processing')).toBe(2)
      expect(getStageOrder('indices')).toBe(3)
      expect(getStageOrder('liquidity')).toBe(4)
      expect(getStageOrder('indicators')).toBe(5)
      expect(getStageOrder('analysis')).toBe(6)
    })

    it('returns high number for unknown stages', () => {
      expect(getStageOrder('unknown' as any)).toBe(999)
    })
  })

  describe('getPipelineStageOrder', () => {
    it('returns stages in correct pipeline order', () => {
      const order = getPipelineStageOrder()

      expect(order).toContain('scraping')
      expect(order).toContain('data_processing')
      expect(order).toContain('processing')
      expect(order).toContain('indices')
      expect(order).toContain('liquidity')
      expect(order).toContain('indicators')
      expect(order).toContain('analysis')

      // Should not include wrapper operations
      expect(order).not.toContain('full_pipeline')

      // Should be in correct order
      expect(order.indexOf('scraping')).toBeLessThan(order.indexOf('processing'))
      expect(order.indexOf('processing')).toBeLessThan(order.indexOf('indices'))
      expect(order.indexOf('indices')).toBeLessThan(order.indexOf('liquidity'))
      expect(order.indexOf('liquidity')).toBeLessThan(order.indexOf('indicators'))
      expect(order.indexOf('indicators')).toBeLessThan(order.indexOf('analysis'))
    })
  })

  describe('isStageActive', () => {
    it('returns true for first stage always', () => {
      const allStages = []
      expect(isStageActive('scraping', allStages)).toBe(true)
    })

    it('returns false for stages with incomplete dependencies', () => {
      const allStages = [
        {
          steps: [{ metadata: { stage_id: 'scraping' } }],
          status: 'completed'
        }
      ]

      // processing depends on scraping, but indices depends on processing
      expect(isStageActive('processing', allStages)).toBe(true)
      expect(isStageActive('indices', allStages)).toBe(false)
    })

    it('returns true when previous stage is completed', () => {
      const allStages = [
        {
          steps: [{ metadata: { stage_id: 'scraping' } }],
          status: 'completed'
        },
        {
          steps: [{ metadata: { stage_id: 'processing' } }],
          status: 'completed'
        }
      ]

      expect(isStageActive('indices', allStages)).toBe(true)
    })
  })

  describe('canActivateStage', () => {
    it('returns false for malformed stage data', () => {
      expect(canActivateStage(null, [])).toBe(false)
      expect(canActivateStage(undefined, [])).toBe(false)
      expect(canActivateStage({}, [])).toBe(false)
    })

    it('uses isStageActive internally', () => {
      const stageData = {
        steps: [{ metadata: { stage_id: 'processing' } }]
      }

      const allStages = [
        {
          steps: [{ metadata: { stage_id: 'scraping' } }],
          status: 'completed'
        }
      ]

      expect(canActivateStage(stageData, allStages)).toBe(true)
    })
  })

  describe('extractStageFileCount', () => {
    it('extracts file count from metadata.files_processed', () => {
      const operation = {
        metadata: {
          files_processed: 25
        }
      }

      expect(extractStageFileCount(operation, 'scraping')).toBe(25)
    })

    it('extracts from stage-specific keys', () => {
      const operation = {
        metadata: {
          files_downloaded: 15
        }
      }

      expect(extractStageFileCount(operation, 'scraping')).toBe(15)
    })

    it('extracts from completed steps', () => {
      const operation = {
        steps: [
          {
            status: 'completed',
            metadata: {
              files_processed: 10
            }
          }
        ]
      }

      expect(extractStageFileCount(operation, 'processing')).toBe(10)
    })

    it('returns defaults for no data', () => {
      expect(extractStageFileCount({}, 'scraping')).toBe(15)
      expect(extractStageFileCount({}, 'processing')).toBe(15)
      expect(extractStageFileCount({}, 'indices')).toBe(1)
    })
  })

  describe('getStageMetrics', () => {
    it('returns correct metrics configuration', () => {
      const scrapingMetrics = getStageMetrics('scraping')
      expect(scrapingMetrics.primaryMetric).toBe('files_processed')
      expect(scrapingMetrics.primaryLabel).toBe('Files Downloaded')
      expect(scrapingMetrics.secondaryMetric).toBe('duration')
      expect(scrapingMetrics.secondaryLabel).toBe('Duration')

      const indicesMetrics = getStageMetrics('indices')
      expect(indicesMetrics.primaryMetric).toBe('indices_extracted')
      expect(indicesMetrics.primaryLabel).toBe('Indices Extracted')
    })

    it('returns default metrics for unknown stage', () => {
      const metrics = getStageMetrics('unknown' as any)
      expect(metrics.primaryMetric).toBe('files_processed')
      expect(metrics.primaryLabel).toBe('Files Downloaded')
    })
  })

  describe('STAGE_DEFINITIONS', () => {
    it('includes all required stages', () => {
      expect(STAGE_DEFINITIONS).toHaveProperty('scraping')
      expect(STAGE_DEFINITIONS).toHaveProperty('data_processing')
      expect(STAGE_DEFINITIONS).toHaveProperty('processing')
      expect(STAGE_DEFINITIONS).toHaveProperty('indices')
      expect(STAGE_DEFINITIONS).toHaveProperty('liquidity')
      expect(STAGE_DEFINITIONS).toHaveProperty('indicators')
      expect(STAGE_DEFINITIONS).toHaveProperty('analysis')
      expect(STAGE_DEFINITIONS).toHaveProperty('full_pipeline')
    })

    it('has consistent structure for all stages', () => {
      Object.entries(STAGE_DEFINITIONS).forEach(([stageId, definition]) => {
        expect(definition).toHaveProperty('name')
        expect(definition).toHaveProperty('description')
        expect(definition).toHaveProperty('category')
        expect(definition).toHaveProperty('color')
        expect(definition).toHaveProperty('icon')
        expect(definition).toHaveProperty('order')
        expect(typeof definition.order).toBe('number')
      })
    })
  })

  test('edge cases and error handling', () => {
    // Test with empty strings
    expect(identifyStage({ stage_id: '' }).stageId).toBe('scraping')

    // Test with special characters
    expect(identifyStage({ name: 'Special@#$% Operation' }).confidence).toBe('low')

    // Test with extremely long strings
    const longName = 'A'.repeat(1000)
    expect(() => identifyStage({ name: longName })).not.toThrow()

    // Test with circular references (should not hang)
    const circular: any = {}
    circular.self = circular
    expect(() => identifyStage(circular)).not.toThrow()
  })
})