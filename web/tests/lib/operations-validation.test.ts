/**
 * Operations Validation Tests
 *
 * Comprehensive tests for the enhanced validation system including:
 * - Schema registry functionality
 * - Field-level validation
 * - Migration rules
 * - Error handling
 * - Performance metrics
 */

import {
  schemaRegistry,
  parseOperationSnapshotDTO,
  validateOperationSnapshot,
  validateOperationSnapshotComprehensive,
  validateOperationSnapshotsBatch,
  parseWebSocketMessageEnhanced,
  ValidationUtils,
  ValidationError,
  ValidationDetails,
  ValidationReport,
  BatchValidationResult
} from '@/lib/operations/validation'

describe('Schema Registry', () => {
  describe('Schema Registration and Retrieval', () => {
    test('should register and retrieve schemas correctly', () => {
      const v1Schema = schemaRegistry.get('1.0')
      expect(v1Schema).toBeDefined()
      expect(v1Schema?.version).toBe('1.0')
      expect(v1Schema?.fields).toHaveProperty('operation_id')
      expect(v1Schema?.fields).toHaveProperty('status')
      expect(v1Schema?.fields).toHaveProperty('progress')
    })

    test('should get latest schema version', () => {
      const latestSchema = schemaRegistry.getLatest()
      expect(latestSchema.version).toBe('1.0')
    })

    test('should return null for unknown schema version', () => {
      const unknownSchema = schemaRegistry.get('99.99')
      expect(unknownSchema).toBeNull()
    })
  })

  describe('Migration System', () => {
    test('should migrate v0.9 to v1.0 correctly', () => {
      const v0_9Data = {
        operation_id: 'test-op-123',
        status: 'running',
        progress: 50,
        current_stage: 'processing',
        started_at: '2025-01-01T10:00:00Z',
        updated_at: '2025-01-01T10:30:00Z'
      }

      const result = schemaRegistry.migrate(v0_9Data)

      expect(result.data).toHaveProperty('current_step', '')
      expect(result.data).toHaveProperty('steps', [])
      expect(result.data).toHaveProperty('metadata')
      expect(result.data.metadata.version).toBe('1.0')
      expect(result.migrations.length).toBeGreaterThan(0)
      expect(result.migrations[0]).toContain('Migrate from v0.9 to v1.0')
    })

    test('should not migrate if already at latest version', () => {
      const v1Data = {
        operation_id: 'test-op-456',
        status: 'completed',
        progress: 100,
        current_stage: 'finished',
        started_at: '2025-01-01T10:00:00Z',
        updated_at: '2025-01-01T11:00:00Z',
        completed_at: '2025-01-01T11:00:00Z',
        metadata: { version: '1.0' }
      }

      const result = schemaRegistry.migrate(v1Data)

      expect(result.migrations).toHaveLength(0)
      expect(result.data).toEqual(v1Data)
    })

    test('should handle migration errors gracefully', () => {
      const invalidData = {
        // Missing required operation_id
        status: 'running'
      }

      expect(() => {
        schemaRegistry.migrate(invalidData)
      }).not.toThrow() // Should not throw, but may log warnings
    })
  })

  describe('Validation', () => {
    test('should validate correct data successfully', () => {
      const validData = {
        operation_id: 'test-op-789',
        status: 'running',
        progress: 75,
        current_stage: 'processing',
        started_at: '2025-01-01T10:00:00Z',
        updated_at: '2025-01-01T10:45:00Z',
        metadata: { version: '1.0' }
      }

      const result = schemaRegistry.validate(validData)

      expect(result.isValid).toBe(true)
      expect(result.errors).toHaveLength(0)
      expect(result.version).toBe('1.0')
    })

    test('should detect missing required fields', () => {
      const invalidData = {
        status: 'running',
        progress: 50
        // Missing operation_id, started_at, updated_at
      }

      const result = schemaRegistry.validate(invalidData)

      expect(result.isValid).toBe(false)
      expect(result.errors.length).toBeGreaterThan(0)

      const missingFieldErrors = result.errors.filter(e => e.code === 'MISSING_REQUIRED_FIELD')
      expect(missingFieldErrors.length).toBeGreaterThan(0)
    })

    test('should validate field types and constraints', () => {
      const invalidData = {
        operation_id: 123, // Should be string
        status: 'invalid_status', // Invalid enum value
        progress: 150, // Out of range
        started_at: 'invalid-date', // Invalid date format
        updated_at: '2025-01-01T10:00:00Z'
      }

      const result = schemaRegistry.validate(invalidData)

      expect(result.isValid).toBe(false)
      expect(result.errors.length).toBeGreaterThan(2) // At least type and enum errors
    })

    test('should provide warnings for deprecated fields', () => {
      const dataWithDeprecatedFields = {
        operation_id: 'test-op-123',
        status: 'running',
        progress: 50,
        current_step: 'processing', // Deprecated field
        current_stage: 'processing',
        started_at: '2025-01-01T10:00:00Z',
        updated_at: '2025-01-01T10:30:00Z'
      }

      const result = schemaRegistry.validate(dataWithDeprecatedFields)

      expect(result.warnings.length).toBeGreaterThan(0)
      expect(result.warnings.some(w => w.includes('current_step'))).toBe(true)
    })
  })
})

describe('parseOperationSnapshotDTO', () => {
  test('should parse valid operation data correctly', () => {
    const validData = {
      operation_id: 'test-op-123',
      status: 'running',
      progress: 75,
      current_stage: 'processing'
    }

    const result = parseOperationSnapshotDTO(validData)

    expect(result).toBeDefined()
    expect(result!.operation_id).toBe('test-op-123')
    expect(result!.status).toBe('running')
    expect(result!.progress).toBe(75)
    expect(result!.metadata).toBeDefined()
  })

  test('should return null for invalid data', () => {
    const invalidData = {
      // Missing operation_id
      status: 'running'
    }

    const result = parseOperationSnapshotDTO(invalidData)
    expect(result).toBeNull()
  })

  test('should handle migration during parsing', () => {
    const legacyData = {
      operation_id: 'legacy-op-456',
      status: 'completed',
      progress: 100,
      current_stage: 'finished'
      // Missing started_at, updated_at - should get defaults
    }

    const result = parseOperationSnapshotDTO(legacyData)

    expect(result).toBeDefined()
    expect(result!.operation_id).toBe('legacy-op-456')
    expect(result!.started_at).toBeDefined()
    expect(result!.updated_at).toBeDefined()
    expect(result!.metadata.migrationPerformed).toBe(true)
  })

  test('should normalize data types correctly', () => {
    const dataWithWrongTypes = {
      operation_id: 123, // Should be converted to string
      status: 'running',
      progress: '75', // Should be converted to number
      current_stage: 'processing',
      steps: null // Should be converted to empty array
    }

    const result = parseOperationSnapshotDTO(dataWithWrongTypes)

    expect(result).toBeDefined()
    expect(typeof result!.operation_id).toBe('string')
    expect(typeof result!.progress).toBe('number')
    expect(Array.isArray(result!.steps)).toBe(true)
  })

  test('should handle parsing errors gracefully', () => {
    const problematicData = {
      operation_id: 'test-op-error',
      status: 'running',
      progress: 50
    }

    // Mock console.warn to verify error logging
    const originalWarn = console.warn
    console.warn = jest.fn()

    const result = parseOperationSnapshotDTO(problematicData)

    expect(result).toBeDefined() // Should return fallback data
    expect(console.warn).toHaveBeenCalled()

    console.warn = originalWarn
  })
})

describe('validateOperationSnapshot', () => {
  test('should provide comprehensive validation details', () => {
    const testData = {
      operation_id: 'test-validation-123',
      status: 'running',
      progress: 60,
      current_stage: 'processing'
    }

    const result = validateOperationSnapshot(testData)

    expect(result).toHaveProperty('isValid')
    expect(result).toHaveProperty('errors')
    expect(result).toHaveProperty('warnings')
    expect(result).toHaveProperty('version')
    expect(result).toHaveProperty('migrationPerformed')
    expect(result).toHaveProperty('migrationDetails')
  })

  test('should include migration information in results', () => {
    const v0_9Data = {
      operation_id: 'migration-test-456',
      status: 'completed',
      progress: 100,
      current_stage: 'finished'
    }

    const result = validateOperationSnapshot(v0_9Data)

    expect(result.migrationPerformed).toBe(true)
    expect(result.migrationDetails).toBeDefined()
    expect(result.migrationDetails!.length).toBeGreaterThan(0)
  })

  test('should provide legacy warnings for compatibility', () => {
    const dataWithBothStepFields = {
      operation_id: 'step-test-789',
      status: 'running',
      progress: 50,
      current_step: 'processing-step',
      current_stage: 'processing-stage'
    }

    const result = validateOperationSnapshot(dataWithBothStepFields)

    expect(result.warnings).toContain('Both current_step and current_stage present - using current_stage')
  })
})

describe('validateOperationSnapshotComprehensive', () => {
  test('should provide performance metrics', () => {
    const testData = {
      operation_id: 'perf-test-123',
      status: 'running',
      progress: 75,
      current_stage: 'processing'
    }

    const result = validateOperationSnapshotComprehensive(testData)

    expect(result).toHaveProperty('performanceMetrics')
    expect(result.performanceMetrics).toHaveProperty('validationTime')
    expect(result.performanceMetrics).toHaveProperty('migrationTime')
    expect(typeof result.performanceMetrics.validationTime).toBe('number')
    expect(typeof result.performanceMetrics.migrationTime).toBe('number')
  })

  test('should provide actionable recommendations', () => {
    const invalidData = {
      operation_id: '', // Invalid
      status: 'invalid_status', // Invalid
      progress: 150 // Out of range
    }

    const result = validateOperationSnapshotComprehensive(invalidData)

    expect(result.recommendations.length).toBeGreaterThan(0)
    expect(result.recommendations.some(r => r.includes('Fix validation errors'))).toBe(true)
  })

  test('should handle comprehensive validation errors gracefully', () => {
    const result = validateOperationSnapshotComprehensive(null)

    expect(result.isValid).toBe(false)
    expect(result.errors.length).toBeGreaterThan(0)
    expect(result.recommendations).toContain('Check data structure and schema compatibility')
  })
})

describe('validateOperationSnapshotsBatch', () => {
  test('should validate multiple operation snapshots', () => {
    const testDataArray = [
      {
        operation_id: 'batch-test-1',
        status: 'running',
        progress: 25,
        current_stage: 'processing'
      },
      {
        operation_id: 'batch-test-2',
        status: 'completed',
        progress: 100,
        current_stage: 'finished'
      },
      {
        operation_id: '', // Invalid
        status: 'invalid_status' // Invalid
      }
    ]

    const result = validateOperationSnapshotsBatch(testDataArray)

    expect(result.totalItems).toBe(3)
    expect(result.validItems).toBe(2)
    expect(result.invalidItems).toBe(1)
    expect(result.reports).toHaveLength(3)
  })

  test('should provide batch-level statistics', () => {
    const testDataArray = Array.from({ length: 10 }, (_, i) => ({
      operation_id: `batch-item-${i}`,
      status: i % 2 === 0 ? 'running' : 'completed',
      progress: i * 10,
      current_stage: 'processing'
    }))

    const result = validateOperationSnapshotsBatch(testDataArray)

    expect(result.summary).toHaveProperty('averageValidationTime')
    expect(result.summary).toHaveProperty('mostCommonErrors')
    expect(result.summary).toHaveProperty('recommendations')
    expect(typeof result.summary.averageValidationTime).toBe('number')
  })

  test('should analyze most common errors', () => {
    const testDataArray = Array.from({ length: 5 }, (_, i) => ({
      // All missing required fields
      status: 'running',
      progress: 50
    }))

    const result = validateOperationSnapshotsBatch(testDataArray)

    expect(result.invalidItems).toBe(5)
    expect(result.summary.mostCommonErrors.length).toBeGreaterThan(0)
    expect(result.summary.mostCommonErrors[0].error).toContain('operation_id')
  })

  test('should provide batch recommendations', () => {
    const testDataArray = [
      { operation_id: 'valid-1', status: 'running', progress: 50 },
      { operation_id: '', status: 'running', progress: 50 }, // Invalid
      { operation_id: 'valid-3', status: 'running', progress: 50 }
    ]

    const result = validateOperationSnapshotsBatch(testDataArray)

    expect(result.summary.recommendations.length).toBeGreaterThan(0)
    expect(result.summary.recommendations.some(r => r.includes('validation errors'))).toBe(true)
  })
})

describe('parseWebSocketMessageEnhanced', () => {
  test('should parse valid WebSocket messages with validation', () => {
    const validMessage = {
      type: 'operation:snapshot',
      data: {
        operation_id: 'ws-test-123',
        status: 'running',
        progress: 75,
        current_stage: 'processing'
      },
      timestamp: '2025-01-01T10:30:00Z',
      id: 'msg-123'
    }

    const result = parseWebSocketMessageEnhanced(validMessage)

    expect(result).toBeDefined()
    expect(result!.type).toBe('operation:snapshot')
    expect(result!.data.operation_id).toBe('ws-test-123')
    expect(result!.id).toBe('msg-123')
  })

  test('should return null for invalid WebSocket messages', () => {
    const invalidMessage = {
      type: 'invalid-type',
      data: { operation_id: 'test' }
    }

    const result = parseWebSocketMessageEnhanced(invalidMessage)
    expect(result).toBeNull()
  })

  test('should handle invalid operation data in messages', () => {
    const messageWithInvalidData = {
      type: 'operation:snapshot',
      data: {
        // Missing operation_id
        status: 'running'
      }
    }

    const result = parseWebSocketMessageEnhanced(messageWithInvalidData)
    expect(result).toBeNull()
  })
})

describe('ValidationUtils', () => {
  describe('validateOperationId', () => {
    test('should validate correct operation IDs', () => {
      const validIds = ['test-op-123', 'Operation_456', 'OP-789_ABC']

      validIds.forEach(id => {
        const result = ValidationUtils.validateOperationId(id)
        expect(result).toBeNull()
      })
    })

    test('should reject invalid operation IDs', () => {
      const invalidCases = [
        { id: '', expectedError: 'Operation ID cannot be empty' },
        { id: '   ', expectedError: 'Operation ID cannot be empty' },
        { id: 'a'.repeat(101), expectedError: 'Operation ID too long' },
        { id: 'invalid@id', expectedError: 'can only contain' },
        { id: 123, expectedError: 'must be a non-empty string' }
      ]

      invalidCases.forEach(({ id, expectedError }) => {
        const result = ValidationUtils.validateOperationId(id as any)
        expect(result).toBeDefined()
        expect(result!.message).toContain(expectedError)
      })
    })
  })

  describe('validateProgress', () => {
    test('should validate correct progress values', () => {
      const validProgress = [0, 25, 50, 75, 100]

      validProgress.forEach(progress => {
        const result = ValidationUtils.validateProgress(progress)
        expect(result).toBeNull()
      })
    })

    test('should reject invalid progress values', () => {
      const invalidProgress = [-10, 150, 75.5, '50', null, undefined]

      invalidProgress.forEach(progress => {
        const result = ValidationUtils.validateProgress(progress as any)
        expect(result).toBeDefined()
      })
    })

    test('should warn about out-of-range values', () => {
      const result = ValidationUtils.validateProgress(150)
      expect(result).toBeDefined()
      expect(result!.severity).toBe('warning')
      expect(result!.code).toBe('OUT_OF_RANGE')
    })
  })

  describe('validateStatus', () => {
    test('should validate correct status values', () => {
      const validStatuses = ['pending', 'running', 'completed', 'failed', 'cancelled', 'unknown']

      validStatuses.forEach(status => {
        const result = ValidationUtils.validateStatus(status)
        expect(result).toBeNull()
      })
    })

    test('should reject invalid status values', () => {
      const invalidStatuses = ['invalid', '', null, undefined, 123]

      invalidStatuses.forEach(status => {
        const result = ValidationUtils.validateStatus(status as any)
        expect(result).toBeDefined()
        expect(result!.code).toBe('INVALID_ENUM_VALUE')
      })
    })
  })

  describe('requiresMigration', () => {
    test('should detect when migration is needed', () => {
      const legacyData = { operation_id: 'test', status: 'running' }
      expect(ValidationUtils.requiresMigration(legacyData)).toBe(true)
    })

    test('should detect when migration is not needed', () => {
      const currentData = {
        operation_id: 'test',
        status: 'running',
        metadata: { version: '1.0' }
      }
      expect(ValidationUtils.requiresMigration(currentData)).toBe(false)
    })
  })

  describe('validateField', () => {
    test('should validate individual fields', () => {
      const result = ValidationUtils.validateField('operation_id', 'test-op-123')
      expect(result).toHaveLength(0)
    })

    test('should detect unknown fields', () => {
      const result = ValidationUtils.validateField('unknown_field', 'value')
      expect(result).toHaveLength(1)
      expect(result[0].code).toBe('UNKNOWN_FIELD')
      expect(result[0].severity).toBe('warning')
    })
  })
})

describe('Performance Tests', () => {
  test('schema validation should be performant', () => {
    const testData = {
      operation_id: 'perf-test-123',
      status: 'running',
      progress: 75,
      current_stage: 'processing'
    }

    const startTime = performance.now()

    // Run validation 100 times
    for (let i = 0; i < 100; i++) {
      schemaRegistry.validate(testData)
    }

    const endTime = performance.now()
    const averageTime = (endTime - startTime) / 100

    // Should complete in less than 1ms per validation on average
    expect(averageTime).toBeLessThan(1)
  })

  test('batch validation should scale efficiently', () => {
    const testDataArray = Array.from({ length: 100 }, (_, i) => ({
      operation_id: `perf-batch-${i}`,
      status: i % 2 === 0 ? 'running' : 'completed',
      progress: (i % 11) * 10,
      current_stage: 'processing'
    }))

    const startTime = performance.now()
    const result = validateOperationSnapshotsBatch(testDataArray)
    const endTime = performance.now()

    expect(result.totalItems).toBe(100)
    expect(result.validItems).toBe(100)

    // Should complete in reasonable time (less than 100ms for 100 items)
    expect(endTime - startTime).toBeLessThan(100)
  })
})

describe('Error Handling and Edge Cases', () => {
  test('should handle null and undefined inputs gracefully', () => {
    expect(parseOperationSnapshotDTO(null)).toBeNull()
    expect(parseOperationSnapshotDTO(undefined)).toBeNull()

    const validationResult = validateOperationSnapshot(null)
    expect(validationResult.isValid).toBe(false)
    expect(validationResult.errors.length).toBeGreaterThan(0)
  })

  test('should handle circular object references', () => {
    const circularData: any = { operation_id: 'test' }
    circularData.self = circularData

    // Should not throw and should handle gracefully
    expect(() => {
      validateOperationSnapshot(circularData)
    }).not.toThrow()
  })

  test('should handle extremely large strings', () => {
    const largeString = 'a'.repeat(10000)
    const largeData = {
      operation_id: largeString,
      status: 'running',
      progress: 50
    }

    const result = validateOperationSnapshot(largeData)
    expect(result.isValid).toBe(false)
    expect(result.errors.some(e => e.message.includes('too long'))).toBe(true)
  })

  test('should handle special characters in strings', () => {
    const specialCharData = {
      operation_id: 'test-op-123-特殊字符-🚀',
      status: 'running',
      progress: 50,
      current_stage: 'processing-特殊',
      error: 'Error with special chars: éàöä@#$%^&*()'
    }

    const result = parseOperationSnapshotDTO(specialCharData)
    expect(result).toBeDefined()
    expect(result!.operation_id).toBe('test-op-123-特殊字符-🚀')
  })
})