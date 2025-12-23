/**
 * Tests for Operation Request Builder
 * Tests the simplified 2-operation system (run_stage and full_pipeline)
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { OperationRequestBuilder, OperationRequest, OperationRequestSchema, OPERATIONS_CONFIG } from '@/lib/api/operation-request-builder'

describe('OperationRequestBuilder', () => {
  beforeEach(() => {
    // Reset any module state if needed
  })

  describe('buildRunStage', () => {
    it('should create a valid run_stage request for scraping stage', () => {
      const request = OperationRequestBuilder.buildRunStage('scraping', '2025-01-01', '2025-01-31')

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('scraping')
      expect(request.parameters).toEqual({
        from: '2025-01-01',
        to: '2025-01-31'
      })

      // Validate against schema
      expect(() => OperationRequestSchema.parse(request)).not.toThrow()
    })

    it('should create a valid run_stage request for processing stage', () => {
      const request = OperationRequestBuilder.buildRunStage('processing')

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('processing')
      expect(request.parameters).toEqual({})

      // Validate against schema
      expect(() => OperationRequestSchema.parse(request)).not.toThrow()
    })

    it('should create a valid run_stage request for indices stage', () => {
      const request = OperationRequestBuilder.buildRunStage('indices')

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('indices')
      expect(request.parameters).toEqual({})

      // Validate against schema
      expect(() => OperationRequestSchema.parse(request)).not.toThrow()
    })

    it('should create a valid run_stage request for liquidity stage', () => {
      const request = OperationRequestBuilder.buildRunStage('liquidity')

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('liquidity')
      expect(request.parameters).toEqual({})

      // Validate against schema
      expect(() => OperationRequestSchema.parse(request)).not.toThrow()
    })

    it('should create a valid run_stage request for analysis stage', () => {
      const request = OperationRequestBuilder.buildRunStage('analysis')

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('analysis')
      expect(request.parameters).toEqual({})

      // Validate against schema
      expect(() => OperationRequestSchema.parse(request)).not.toThrow()
    })

    it('should create a valid run_stage request for indicators stage', () => {
      const request = OperationRequestBuilder.buildRunStage('indicators')

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('indicators')
      expect(request.parameters).toEqual({})

      // Validate against schema
      expect(() => OperationRequestSchema.parse(request)).not.toThrow()
    })

    it('should include date parameters only for scraping stage', () => {
      const scrapingRequest = OperationRequestBuilder.buildRunStage('scraping', '2025-01-01', '2025-01-31')
      const processingRequest = OperationRequestBuilder.buildRunStage('processing')

      expect(scrapingRequest.parameters).toHaveProperty('from', '2025-01-01')
      expect(scrapingRequest.parameters).toHaveProperty('to', '2025-01-31')

      expect(processingRequest.parameters).not.toHaveProperty('from')
      expect(processingRequest.parameters).not.toHaveProperty('to')
    })

    it('should use default dates when not provided for scraping', () => {
      const request = OperationRequestBuilder.buildRunStage('scraping')

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('scraping')
      expect(request.parameters).toHaveProperty('from')
      expect(request.parameters).toHaveProperty('to')

      // Should use defaults from OPERATION_DATE_DEFAULTS
      expect(typeof request.parameters.from).toBe('string')
      expect(typeof request.parameters.to).toBe('string')
    })
  })

  describe('buildFullPipeline', () => {
    it('should create a valid full_pipeline request', () => {
      const request = OperationRequestBuilder.buildFullPipeline('2025-01-01', '2025-01-31')

      expect(request.operation).toBe('full_pipeline')
      expect(request.stage_id).toBeUndefined()
      expect(request.parameters).toEqual({
        from: '2025-01-01',
        to: '2025-01-31'
      })

      // Validate against schema
      expect(() => OperationRequestSchema.parse(request)).not.toThrow()
    })

    it('should create a full_pipeline request with default dates', () => {
      const request = OperationRequestBuilder.buildFullPipeline()

      expect(request.operation).toBe('full_pipeline')
      expect(request.stage_id).toBeUndefined()
      expect(request.parameters).toHaveProperty('from')
      expect(request.parameters).toHaveProperty('to')

      // Should use defaults from OPERATION_DATE_DEFAULTS
      expect(typeof request.parameters.from).toBe('string')
      expect(typeof request.parameters.to).toBe('string')

      // Validate against schema
      expect(() => OperationRequestSchema.parse(request)).not.toThrow()
    })

    it('should not include stage_id for full_pipeline requests', () => {
      const request = OperationRequestBuilder.buildFullPipeline()

      expect(request.operation).toBe('full_pipeline')
      expect(request.stage_id).toBeUndefined()
    })
  })

  describe('Legacy Compatibility Methods', () => {
    describe('buildQuickStart', () => {
      it('should map to run_stage operation', () => {
        const request = OperationRequestBuilder.buildQuickStart('scraping')

        expect(request.operation).toBe('run_stage')
        expect(request.stage_id).toBe('scraping')
        expect(request.parameters).toEqual({})
      })

      it('should work with different stage types', () => {
        const stages = ['scraping', 'processing', 'indices', 'liquidity', 'analysis', 'indicators']

        stages.forEach(stage => {
          const request = OperationRequestBuilder.buildQuickStart(stage)
          expect(request.operation).toBe('run_stage')
          expect(request.stage_id).toBe(stage)
        })
      })
    })

    describe('buildWithDates', () => {
      it('should map to run_stage operation with dates', () => {
        const request = OperationRequestBuilder.buildWithDates('scraping', '2025-01-01', '2025-01-31')

        expect(request.operation).toBe('run_stage')
        expect(request.stage_id).toBe('scraping')
        expect(request.parameters).toEqual({
          from: '2025-01-01',
          to: '2025-01-31'
        })
      })

      it('should work without dates for non-scraping stages', () => {
        const request = OperationRequestBuilder.buildWithDates('processing')

        expect(request.operation).toBe('run_stage')
        expect(request.stage_id).toBe('processing')
        expect(request.parameters).toEqual({})
      })
    })
  })

  describe('buildFromParams', () => {
    it('should handle new format operation requests', () => {
      const params = {
        operation: 'run_stage',
        stage_id: 'scraping',
        parameters: {
          from: '2025-01-01',
          to: '2025-01-31'
        }
      }

      const request = OperationRequestBuilder.buildFromParams(params)

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('scraping')
      expect(request.parameters).toEqual({
        from: '2025-01-01',
        to: '2025-01-31'
      })
    })

    it('should handle full_pipeline requests', () => {
      const params = {
        operation: 'full_pipeline',
        parameters: {
          from: '2025-01-01',
          to: '2025-01-31'
        }
      }

      const request = OperationRequestBuilder.buildFromParams(params)

      expect(request.operation).toBe('full_pipeline')
      expect(request.stage_id).toBeUndefined()
      expect(request.parameters).toEqual({
        from: '2025-01-01',
        to: '2025-01-31'
      })
    })

    it('should convert legacy format to new format', () => {
      const params = {
        step: 'scraping',
        from: '2025-01-01',
        to: '2025-01-31'
      }

      const request = OperationRequestBuilder.buildFromParams(params)

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('scraping')
      expect(request.parameters).toEqual({
        from: '2025-01-01',
        to: '2025-01-31'
      })
    })

    it('should convert legacy full_pipeline to new format', () => {
      const params = {
        step: 'full_pipeline',
        from: '2025-01-01',
        to: '2025-01-31'
      }

      const request = OperationRequestBuilder.buildFromParams(params)

      expect(request.operation).toBe('full_pipeline')
      expect(request.stage_id).toBeUndefined()
      expect(request.parameters).toEqual({
        from: '2025-01-01',
        to: '2025-01-31'
      })
    })

    it('should handle legacy operation parameter', () => {
      const params = {
        operation: 'scraping',
        from: '2025-01-01',
        to: '2025-01-31'
      }

      const request = OperationRequestBuilder.buildFromParams(params)

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('scraping')
      expect(request.parameters).toEqual({
        from: '2025-01-01',
        to: '2025-01-31'
      })
    })

    it('should default to processing for unknown stages', () => {
      const params = {
        unknown_param: 'unknown_value'
      }

      const request = OperationRequestBuilder.buildFromParams(params)

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('processing')
      expect(request.parameters).toEqual({})
    })
  })

  describe('validate', () => {
    it('should validate a correct run_stage request', () => {
      const request = {
        operation: 'run_stage' as const,
        stage_id: 'scraping',
        parameters: { from: '2025-01-01', to: '2025-01-31' }
      }

      const validated = OperationRequestBuilder.validate(request)
      expect(validated).toEqual(request)
    })

    it('should validate a correct full_pipeline request', () => {
      const request = {
        operation: 'full_pipeline' as const,
        parameters: { from: '2025-01-01', to: '2025-01-31' }
      }

      const validated = OperationRequestBuilder.validate(request)
      expect(validated).toEqual(request)
    })

    it('should throw validation error for invalid operation', () => {
      const request = {
        operation: 'invalid_operation' as any,
        stage_id: 'scraping'
      }

      expect(() => OperationRequestBuilder.validate(request)).toThrow()
    })

    it('should throw validation error for missing stage_id in run_stage', () => {
      const request = {
        operation: 'run_stage' as const,
        parameters: {}
      }

      expect(() => OperationRequestBuilder.validate(request)).toThrow()
    })

    it('should throw validation error for invalid parameters', () => {
      const request = {
        operation: 'run_stage' as const,
        stage_id: 'scraping',
        parameters: 123 // Should be an object
      }

      expect(() => OperationRequestBuilder.validate(request)).toThrow()
    })
  })

  describe('isValid', () => {
    it('should return true for valid run_stage request', () => {
      const request = {
        operation: 'run_stage' as const,
        stage_id: 'scraping',
        parameters: { from: '2025-01-01' }
      }

      expect(OperationRequestBuilder.isValid(request)).toBe(true)
    })

    it('should return true for valid full_pipeline request', () => {
      const request = {
        operation: 'full_pipeline' as const,
        parameters: { from: '2025-01-01' }
      }

      expect(OperationRequestBuilder.isValid(request)).toBe(true)
    })

    it('should return false for invalid operation', () => {
      const request = {
        operation: 'invalid_operation' as any,
        stage_id: 'scraping'
      }

      expect(OperationRequestBuilder.isValid(request)).toBe(false)
    })

    it('should return false for missing stage_id in run_stage', () => {
      const request = {
        operation: 'run_stage' as const,
        parameters: {}
      }

      expect(OperationRequestBuilder.isValid(request)).toBe(false)
    })

    it('should return false for malformed request', () => {
      const request = null

      expect(OperationRequestBuilder.isValid(request)).toBe(false)
    })
  })

  describe('Schema Validation', () => {
    it('should allow only run_stage and full_pipeline operations', () => {
      const validOperations = ['run_stage', 'full_pipeline']

      validOperations.forEach(operation => {
        const request = {
          operation,
          ...(operation === 'run_stage' ? { stage_id: 'scraping' } : {}),
          parameters: {}
        }

        expect(() => OperationRequestSchema.parse(request)).not.toThrow()
      })
    })

    it('should reject invalid operations', () => {
      const invalidOperations = ['invalid', 'scraping', 'processing', 'custom']

      invalidOperations.forEach(operation => {
        const request = {
          operation: operation as any,
          stage_id: 'scraping',
          parameters: {}
        }

        expect(() => OperationRequestSchema.parse(request)).toThrow()
      })
    })

    it('should require stage_id for run_stage operation', () => {
      const requestWithoutStageId = {
        operation: 'run_stage' as const,
        parameters: {}
      }

      expect(() => OperationRequestSchema.parse(requestWithoutStageId)).toThrow()
    })

    it('should not require stage_id for full_pipeline operation', () => {
      const request = {
        operation: 'full_pipeline' as const,
        parameters: {}
      }

      expect(() => OperationRequestSchema.parse(request)).not.toThrow()
    })
  })

  describe('Type Safety', () => {
    it('should have correct type definitions', () => {
      // These tests verify TypeScript types are working correctly
      const runStageRequest: OperationRequest = OperationRequestBuilder.buildRunStage('scraping')
      const fullPipelineRequest: OperationRequest = OperationRequestBuilder.buildFullPipeline()

      expect(runStageRequest.operation).toBe('run_stage')
      expect(fullPipelineRequest.operation).toBe('full_pipeline')

      // Type assertions would fail at compile time if types were incorrect
      expect(typeof runStageRequest.stage_id).toBe('string')
      expect(fullPipelineRequest.stage_id).toBeUndefined()
    })
  })

  describe('Edge Cases', () => {
    it('should handle empty parameters gracefully', () => {
      const request = OperationRequestBuilder.buildFromParams({})

      expect(request.operation).toBe('run_stage')
      expect(request.stage_id).toBe('processing') // Default stage
      expect(request.parameters).toEqual({})
    })

    it('should handle null/undefined parameters gracefully', () => {
      const request1 = OperationRequestBuilder.buildFromParams(null as any)
      const request2 = OperationRequestBuilder.buildFromParams(undefined as any)

      expect(request1.operation).toBe('run_stage')
      expect(request1.stage_id).toBe('processing')
      expect(request1.parameters).toEqual({})

      expect(request2.operation).toBe('run_stage')
      expect(request2.stage_id).toBe('processing')
      expect(request2.parameters).toEqual({})
    })

    it('should handle stringified dates', () => {
      const request = OperationRequestBuilder.buildRunStage('scraping', '2025-01-01', '2025-01-31')

      expect(typeof request.parameters.from).toBe('string')
      expect(typeof request.parameters.to).toBe('string')
      expect(request.parameters.from).toBe('2025-01-01')
      expect(request.parameters.to).toBe('2025-01-31')
    })
  })
})

describe('OPERATIONS_CONFIG', () => {
  it('should have configuration for all stages', () => {
    const expectedStages = [
      'scraping',
      'processing',
      'indices',
      'liquidity',
      'analysis',
      'indicators',
      'full_pipeline'
    ]

    expectedStages.forEach(stage => {
      expect(OPERATIONS_CONFIG).toHaveProperty(stage)
      expect(OPERATIONS_CONFIG[stage]).toHaveProperty('name')
      expect(OPERATIONS_CONFIG[stage]).toHaveProperty('description')
      expect(OPERATIONS_CONFIG[stage]).toHaveProperty('icon')
      expect(OPERATIONS_CONFIG[stage]).toHaveProperty('requiresDates')
      expect(OPERATIONS_CONFIG[stage]).toHaveProperty('quickStart')
    })
  })

  it('should have correct date requirements', () => {
    expect(OPERATIONS_CONFIG.scraping.requiresDates).toBe(true)
    expect(OPERATIONS_CONFIG.full_pipeline.requiresDates).toBe(true)

    const nonDateStages = ['processing', 'indices', 'liquidity', 'analysis', 'indicators']
    nonDateStages.forEach(stage => {
      expect(OPERATIONS_CONFIG[stage].requiresDates).toBe(false)
    })
  })

  it('should have quick start settings', () => {
    expect(OPERATIONS_CONFIG.scraping.quickStart).toBe(false)
    expect(OPERATIONS_CONFIG.full_pipeline.quickStart).toBe(false)

    const quickStartStages = ['processing', 'indices', 'liquidity', 'analysis', 'indicators']
    quickStartStages.forEach(stage => {
      expect(OPERATIONS_CONFIG[stage].quickStart).toBe(true)
    })
  })
})