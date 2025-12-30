/**
 * Operations Validation Utilities
 *
 * Enhanced validation functions for operation data that handles versioned envelope structure
 * with comprehensive schema registry, field-level validation, and automated migration.
 */

import { OperationSnapshotDTO, TypedWebSocketMessage } from '@/types'

export interface OperationValidationResult {
  isValid: boolean
  errors: string[]
}

export interface ValidationError {
  field: string
  message: string
  value?: any
  code: string
  severity: 'error' | 'warning' | 'info'
}

export interface ValidationDetails {
  isValid: boolean
  errors: ValidationError[]
  warnings: string[]
  version?: string
  source?: string
  migrationPerformed?: boolean
  migrationDetails?: string[]
}

export interface SchemaDefinition {
  version: string
  fields: Record<string, SchemaField>
  required: string[]
  deprecated?: Record<string, string>
  migrationRules?: Record<string, MigrationRule>
}

export interface SchemaField {
  type: 'string' | 'number' | 'boolean' | 'date' | 'array' | 'object'
  required: boolean
  nullable?: boolean
  validation?: FieldValidation
  defaultValue?: any
}

export interface FieldValidation {
  min?: number
  max?: number
  pattern?: string
  enum?: any[]
  custom?: (value: any) => string | null
}

export interface MigrationRule {
  fromVersion: string
  toVersion: string
  transform: (data: any) => any
  description: string
}

export interface SchemaRegistry {
  register: (schema: SchemaDefinition) => void
  get: (version: string) => SchemaDefinition | null
  getLatest: () => SchemaDefinition
  migrate: (data: any, targetVersion?: string) => { data: any; migrations: string[] }
  validate: (data: any, version?: string) => ValidationDetails
}

/**
 * Schema Registry Implementation
 *
 * Provides comprehensive schema management for operation DTOs with versioning,
 * validation, and automatic migration capabilities.
 */
class OperationSchemaRegistry implements SchemaRegistry {
  private schemas: Map<string, SchemaDefinition> = new Map()
  private latestVersion: string = '1.0'

  constructor() {
    this.initializeSchemas()
  }

  private initializeSchemas(): void {
    // Register v1.0 schema (current version)
    this.register({
      version: '1.0',
      fields: {
        operation_id: {
          type: 'string',
          required: true,
          validation: {
            pattern: '^[a-zA-Z0-9_-]+$',
            custom: (value: string) => {
              if (!value || value.trim().length === 0) {
                return 'Operation ID cannot be empty'
              }
              if (value.length > 100) {
                return 'Operation ID too long (max 100 characters)'
              }
              return null
            }
          }
        },
        status: {
          type: 'string',
          required: true,
          defaultValue: 'unknown',
          validation: {
            enum: ['pending', 'running', 'completed', 'failed', 'cancelled', 'unknown']
          }
        },
        progress: {
          type: 'number',
          required: true,
          defaultValue: 0,
          validation: {
            min: 0,
            max: 100,
            custom: (value: number) => {
              if (!Number.isInteger(value)) {
                return 'Progress must be an integer'
              }
              return null
            }
          }
        },
        current_step: {
          type: 'string',
          required: false,
          defaultValue: ''
        },
        current_stage: {
          type: 'string',
          required: false,
          defaultValue: '',
          validation: {
            custom: (value: string) => {
              if (value && value.length > 50) {
                return 'Stage name too long (max 50 characters)'
              }
              return null
            }
          }
        },
        steps: {
          type: 'array',
          required: false,
          defaultValue: []
        },
        started_at: {
          type: 'date',
          required: true,
          defaultValue: () => new Date().toISOString()
        },
        updated_at: {
          type: 'date',
          required: true,
          defaultValue: () => new Date().toISOString()
        },
        completed_at: {
          type: 'date',
          required: false,
          nullable: true
        },
        error: {
          type: 'string',
          required: false,
          nullable: true,
          validation: {
            custom: (value: string) => {
              if (value && value.length > 1000) {
                return 'Error message too long (max 1000 characters)'
              }
              return null
            }
          }
        },
        message: {
          type: 'string',
          required: false,
          validation: {
            custom: (value: string) => {
              if (value && value.length > 500) {
                return 'Message too long (max 500 characters)'
              }
              return null
            }
          }
        },
        metadata: {
          type: 'object',
          required: false,
          defaultValue: {}
        }
      },
      required: ['operation_id', 'status', 'progress', 'started_at', 'updated_at'],
      deprecated: {
        current_step: 'Use current_stage instead',
        progress: 'Still valid but may be enhanced in future versions'
      }
    })

    // Register v0.9 schema (legacy version)
    this.register({
      version: '0.9',
      fields: {
        operation_id: {
          type: 'string',
          required: true,
          validation: {
            pattern: '^[a-zA-Z0-9_-]+$'
          }
        },
        status: {
          type: 'string',
          required: true,
          defaultValue: 'unknown'
        },
        progress: {
          type: 'number',
          required: true,
          defaultValue: 0,
          validation: {
            min: 0,
            max: 100
          }
        },
        current_stage: {
          type: 'string',
          required: false,
          defaultValue: ''
        },
        started_at: {
          type: 'date',
          required: true,
          defaultValue: () => new Date().toISOString()
        },
        updated_at: {
          type: 'date',
          required: true,
          defaultValue: () => new Date().toISOString()
        },
        completed_at: {
          type: 'date',
          required: false,
          nullable: true
        },
        error: {
          type: 'string',
          required: false,
          nullable: true
        },
        message: {
          type: 'string',
          required: false
        }
      },
      required: ['operation_id', 'status', 'progress', 'started_at', 'updated_at'],
      migrationRules: {
        '1.0': {
          fromVersion: '0.9',
          toVersion: '1.0',
          transform: (data: any) => {
            return {
              ...data,
              current_step: data.current_step || '',
              steps: data.steps || [],
              metadata: data.metadata || {},
              // Ensure timestamp format consistency
              started_at: this.normalizeTimestamp(data.started_at),
              updated_at: this.normalizeTimestamp(data.updated_at),
              completed_at: data.completed_at ? this.normalizeTimestamp(data.completed_at) : null
            }
          },
          description: 'Migrate from v0.9 to v1.0: Add missing fields and normalize timestamps'
        }
      }
    })
  }

  private normalizeTimestamp(timestamp: any): string {
    if (!timestamp) return new Date().toISOString()

    // If it's already in ISO format, return as-is
    if (typeof timestamp === 'string' && timestamp.includes('T')) {
      return timestamp
    }

    // Convert to ISO string
    const date = new Date(timestamp)
    return isNaN(date.getTime()) ? new Date().toISOString() : date.toISOString()
  }

  register(schema: SchemaDefinition): void {
    this.schemas.set(schema.version, schema)

    // Update latest version if this is newer
    if (parseFloat(schema.version) > parseFloat(this.latestVersion)) {
      this.latestVersion = schema.version
    }
  }

  get(version: string): SchemaDefinition | null {
    return this.schemas.get(version) || null
  }

  getLatest(): SchemaDefinition {
    return this.schemas.get(this.latestVersion)!
  }

  migrate(data: any, targetVersion?: string): { data: any; migrations: string[] } {
    const sourceVersion = data.metadata?.version || '0.9' // Default to legacy version
    const target = targetVersion || this.latestVersion

    if (sourceVersion === target) {
      return { data, migrations: [] }
    }

    const migrations: string[] = []
    let currentData = { ...data }

    // Find migration path
    const path = this.findMigrationPath(sourceVersion, target)

    if (path.length === 0) {
      throw new Error(`No migration path found from ${sourceVersion} to ${target}`)
    }

    // Apply migrations in sequence
    for (const version of path) {
      const schema = this.get(version)
      const migrationRule = schema?.migrationRules?.[this.getNextVersion(version)]

      if (migrationRule) {
        currentData = migrationRule.transform(currentData)
        migrations.push(migrationRule.description)
        currentData.metadata = {
          ...currentData.metadata,
          version: migrationRule.toVersion,
          migratedAt: new Date().toISOString()
        }
      }
    }

    return { data: currentData, migrations }
  }

  findMigrationPath(fromVersion: string, toVersion: string): string[] {
    // Simple implementation for now - can be enhanced with graph algorithms
    const path: string[] = []

    if (fromVersion === '0.9' && toVersion === '1.0') {
      path.push('0.9', '1.0')
    }

    return path
  }

  private getNextVersion(version: string): string {
    const versions = Array.from(this.schemas.keys()).sort((a, b) =>
      parseFloat(a) - parseFloat(b)
    )

    const currentIndex = versions.indexOf(version)
    const next = currentIndex < versions.length - 1 ? versions[currentIndex + 1] : undefined
    return next ?? version
  }

  validate(data: any, version?: string): ValidationDetails {
    const targetVersion = version || this.latestVersion
    const schema = this.get(targetVersion)

    if (!schema) {
      return {
        isValid: false,
        errors: [{
          field: 'schema',
          message: `Unknown schema version: ${targetVersion}`,
          code: 'UNKNOWN_SCHEMA_VERSION',
          severity: 'error'
        }],
        warnings: [],
        version: targetVersion
      }
    }

    const details: ValidationDetails = {
      isValid: false,
      errors: [],
      warnings: [],
      version: targetVersion,
      ...(data.metadata?.source !== undefined ? { source: data.metadata.source } : {})
    }

    // Check required fields
    for (const requiredField of schema.required) {
      if (!(requiredField in data) || data[requiredField] === null || data[requiredField] === undefined) {
        details.errors.push({
          field: requiredField,
          message: `Required field '${requiredField}' is missing`,
          value: data[requiredField],
          code: 'MISSING_REQUIRED_FIELD',
          severity: 'error'
        })
      }
    }

    // Validate field types and constraints
    for (const [fieldName, fieldDef] of Object.entries(schema.fields)) {
      if (!(fieldName in data) && fieldDef.defaultValue !== undefined) {
        continue // Skip validation for missing fields with defaults
      }

      const value = data[fieldName]
      const fieldErrors = this.validateField(fieldName, value, fieldDef)
      details.errors.push(...fieldErrors)

      // Check for deprecated fields
      if (schema.deprecated?.[fieldName]) {
        details.warnings.push(`Field '${fieldName}' is deprecated: ${schema.deprecated[fieldName]}`)
      }
    }

    // Check for unknown fields
    for (const fieldName of Object.keys(data)) {
      if (fieldName !== 'metadata' && !schema.fields[fieldName] && fieldName !== 'operation_id') {
        details.warnings.push(`Unknown field '${fieldName}' - may be from newer version`)
      }
    }

    details.isValid = details.errors.filter(e => e.severity === 'error').length === 0
    return details
  }

  validateField(fieldName: string, value: any, fieldDef: SchemaField): ValidationError[] {
    const errors: ValidationError[] = []

    // Skip validation if field is nullable and value is null
    if (fieldDef.nullable && (value === null || value === undefined)) {
      return errors
    }

    // Type validation
    const typeError = this.validateFieldType(fieldName, value, fieldDef.type)
    if (typeError) {
      errors.push(typeError)
      return errors // Skip further validation if type is wrong
    }

    // Custom validation rules
    if (fieldDef.validation) {
      const validationError = this.validateFieldConstraints(fieldName, value, fieldDef.validation)
      if (validationError) {
        errors.push(validationError)
      }
    }

    return errors
  }

  private validateFieldType(fieldName: string, value: any, expectedType: string): ValidationError | null {
    if (expectedType === 'string' && typeof value !== 'string') {
      return {
        field: fieldName,
        message: `Expected string, got ${typeof value}`,
        value,
        code: 'INVALID_TYPE',
        severity: 'error'
      }
    }

    if (expectedType === 'number' && typeof value !== 'number') {
      return {
        field: fieldName,
        message: `Expected number, got ${typeof value}`,
        value,
        code: 'INVALID_TYPE',
        severity: 'error'
      }
    }

    if (expectedType === 'boolean' && typeof value !== 'boolean') {
      return {
        field: fieldName,
        message: `Expected boolean, got ${typeof value}`,
        value,
        code: 'INVALID_TYPE',
        severity: 'error'
      }
    }

    if (expectedType === 'date') {
      const date = new Date(value)
      if (isNaN(date.getTime())) {
        return {
          field: fieldName,
          message: 'Invalid date format',
          value,
          code: 'INVALID_DATE',
          severity: 'error'
        }
      }
    }

    if (expectedType === 'array' && !Array.isArray(value)) {
      return {
        field: fieldName,
        message: `Expected array, got ${typeof value}`,
        value,
        code: 'INVALID_TYPE',
        severity: 'error'
      }
    }

    if (expectedType === 'object' && (typeof value !== 'object' || Array.isArray(value))) {
      return {
        field: fieldName,
        message: `Expected object, got ${Array.isArray(value) ? 'array' : typeof value}`,
        value,
        code: 'INVALID_TYPE',
        severity: 'error'
      }
    }

    return null
  }

  private validateFieldConstraints(fieldName: string, value: any, validation: FieldValidation): ValidationError | null {
    // Numeric constraints
    if (typeof value === 'number') {
      if (validation.min !== undefined && value < validation.min) {
        return {
          field: fieldName,
          message: `Value ${value} is below minimum ${validation.min}`,
          value,
          code: 'MIN_VALUE_VIOLATION',
          severity: 'error'
        }
      }

      if (validation.max !== undefined && value > validation.max) {
        return {
          field: fieldName,
          message: `Value ${value} is above maximum ${validation.max}`,
          value,
          code: 'MAX_VALUE_VIOLATION',
          severity: 'error'
        }
      }
    }

    // String constraints
    if (typeof value === 'string') {
      if (validation.pattern && !new RegExp(validation.pattern).test(value)) {
        return {
          field: fieldName,
          message: `Value '${value}' does not match required pattern`,
          value,
          code: 'PATTERN_MISMATCH',
          severity: 'error'
        }
      }
    }

    // Enum constraints
    if (validation.enum && !validation.enum.includes(value)) {
      return {
        field: fieldName,
        message: `Value '${value}' is not in allowed values: [${validation.enum.join(', ')}]`,
        value,
        code: 'INVALID_ENUM_VALUE',
        severity: 'error'
      }
    }

    // Custom validation
    if (validation.custom) {
      const customError = validation.custom(value)
      if (customError) {
        return {
          field: fieldName,
          message: customError,
          value,
          code: 'CUSTOM_VALIDATION_FAILED',
          severity: 'error'
        }
      }
    }

    return null
  }
}

// Global schema registry instance
export const schemaRegistry = new OperationSchemaRegistry()

/**
 * Enhanced parser that accepts versioned envelope structure and returns OperationSnapshotDTO
 * with proper validation, normalization, and automatic migration. Handles both current and legacy payload formats.
 */
export function parseOperationSnapshotDTO(data: any): OperationSnapshotDTO | null {
  if (!data || typeof data !== 'object') {
    return null
  }

  // Check for absolutely required fields
  if (!data.operation_id) {
    return null
  }

  try {
    // Detect version and migrate if necessary
    const migrationResult = schemaRegistry.migrate(data)
    const validatedData = migrationResult.data

    // Validate against latest schema
    const validationDetails = schemaRegistry.validate(validatedData)

    if (!validationDetails.isValid) {
      // Log validation errors but still return normalized data for backward compatibility
      console.warn('[parseOperationSnapshotDTO] Validation errors:', validationDetails.errors)
    }

    // Apply schema defaults and normalize
    const schema = schemaRegistry.getLatest()
    const normalized = { ...validatedData }

    // Apply default values for missing optional fields
    for (const [fieldName, fieldDef] of Object.entries(schema.fields)) {
      if (!(fieldName in normalized) && fieldDef.defaultValue !== undefined) {
        normalized[fieldName] = typeof fieldDef.defaultValue === 'function'
          ? fieldDef.defaultValue()
          : fieldDef.defaultValue
      }
    }

    // Ensure field type consistency
    normalized.operation_id = String(normalized.operation_id)
    normalized.progress = Number(normalized.progress) || 0
    normalized.status = normalized.status || 'unknown'
    normalized.current_step = normalized.current_step || normalized.current_stage || ''
    normalized.current_stage = normalized.current_stage || normalized.current_step || ''
    normalized.steps = Array.isArray(normalized.steps) ? normalized.steps : []
    normalized.started_at = normalized.started_at || new Date().toISOString()
    normalized.updated_at = normalized.updated_at || new Date().toISOString()
    normalized.metadata = normalized.metadata || {}

    // Store migration information
    if (migrationResult.migrations.length > 0) {
      normalized.metadata.migrationDetails = migrationResult.migrations
      normalized.metadata.migrationPerformed = true
    }

    return normalized as OperationSnapshotDTO

  } catch (error) {
    console.error('[parseOperationSnapshotDTO] Migration/validation failed:', error)

    // Fallback to basic normalization for backward compatibility
    const fallback = {
      operation_id: String(data.operation_id),
      status: data.status || 'unknown',
      progress: typeof data.progress === 'number' ? data.progress : 0,
      current_step: data.current_step || data.current_stage || '',
      current_stage: data.current_stage || data.current_step || '',
      steps: Array.isArray(data.steps) ? data.steps : [],
      started_at: data.started_at || new Date().toISOString(),
      updated_at: data.updated_at || new Date().toISOString(),
      completed_at: data.completed_at,
      error: data.error,
      message: data.message,
      metadata: {
        ...data.metadata,
        validationFailed: true,
        validationError: error instanceof Error ? error.message : 'Unknown error'
      },
    }

    return fallback as OperationSnapshotDTO
  }
}

/**
 * Validates operation snapshot data and provides detailed feedback using schema registry
 */
export function validateOperationSnapshot(data: any): ValidationDetails {
  // Use schema registry for comprehensive validation
  try {
    // Detect version and migrate if necessary
    const migrationResult = schemaRegistry.migrate(data)
    const migratedData = migrationResult.data

    // Validate against latest schema
    const details = schemaRegistry.validate(migratedData)

    // Add migration information to validation details
    if (migrationResult.migrations.length > 0) {
      details.migrationPerformed = true
      details.migrationDetails = migrationResult.migrations
    }

    // Add legacy warnings for backward compatibility
    if (data.current_step && data.current_stage && data.current_step !== data.current_stage) {
      details.warnings.push('Both current_step and current_stage present - using current_stage')
    }

    return details

  } catch (error) {
    // Fallback validation if schema registry fails
    const details: ValidationDetails = {
      isValid: false,
      errors: [{
        field: 'validation',
        message: `Schema validation failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
        code: 'VALIDATION_ERROR',
        severity: 'error'
      }],
      warnings: [],
      source: data.metadata?.source
    }

    // Basic fallback validation
    if (!data || typeof data !== 'object') {
      details.errors.unshift({
        field: 'root',
        message: 'Data must be an object',
        code: 'INVALID_TYPE',
        severity: 'error'
      })
    } else if (!data.operation_id) {
      details.errors.unshift({
        field: 'operation_id',
        message: 'Operation ID is required',
        code: 'MISSING_REQUIRED_FIELD',
        severity: 'error'
      })
    }

    return details
  }
}

export function parseWebSocketMessage(data: any): TypedWebSocketMessage | null {
  // Basic validation - ensure we have an object
  if (!data || typeof data !== 'object') {
    return null
  }

  // Check for required type field
  if (!data.type || typeof data.type !== 'string') {
    return null
  }

  // Check for required data field
  if (!data.data || typeof data.data !== 'object') {
    return null
  }

  // Ensure we have the correct event type for operation snapshots
  if (data.type !== 'operation:snapshot') {
    return null
  }

  // Return properly structured TypedWebSocketMessage
  return {
    type: data.type as 'operation:snapshot',
    data: data.data,
    timestamp: data.timestamp || new Date().toISOString(),
    id: data.id
  }
}

export function isOperationSnapshotMessage(message: any): boolean {
  // Validate proper structure for operation snapshot messages
  return message &&
         message.type === 'operation:snapshot' &&
         typeof message.data === 'object' &&
         message.data !== null
}

// ============================================================================
// Enhanced Validation Utilities
// ============================================================================

/**
 * Enhanced WebSocket message parser with validation and migration support
 */
export function parseWebSocketMessageEnhanced(data: any): TypedWebSocketMessage | null {
  // Basic validation - ensure we have an object
  if (!data || typeof data !== 'object') {
    return null
  }

  // Check for required type field
  if (!data.type || typeof data.type !== 'string') {
    return null
  }

  // Check for required data field
  if (!data.data || typeof data.data !== 'object') {
    return null
  }

  // Ensure we have the correct event type for operation snapshots
  if (data.type !== 'operation:snapshot') {
    return null
  }

  // Validate and migrate the operation data
  const operationData = parseOperationSnapshotDTO(data.data)
  if (!operationData) {
    console.warn('[parseWebSocketMessageEnhanced] Invalid operation data in message')
    return null
  }

  // Return properly structured TypedWebSocketMessage with enhanced validation
  return {
    type: data.type as 'operation:snapshot',
    data: operationData,
    timestamp: data.timestamp || new Date().toISOString(),
    id: data.id
  }
}

/**
 * Comprehensive validation report with actionable insights
 */
export interface ValidationReport {
  isValid: boolean
  errors: ValidationError[]
  warnings: string[]
  migrationPerformed: boolean
  migrationDetails: string[]
  version: string
  recommendations: string[]
  performanceMetrics: {
    validationTime: number
    migrationTime: number
  }
}

/**
 * Comprehensive validation with performance metrics and recommendations
 */
export function validateOperationSnapshotComprehensive(data: any): ValidationReport {
  const startTime = performance.now()

  try {
    // Detect version and migrate if necessary
    const migrationStart = performance.now()
    const migrationResult = schemaRegistry.migrate(data)
    const migrationTime = performance.now() - migrationStart

    const migratedData = migrationResult.data

    // Validate against latest schema
    const validationStart = performance.now()
    const details = schemaRegistry.validate(migratedData)
    const validationTime = performance.now() - validationStart

    // Generate recommendations based on validation results
    const recommendations: string[] = []

    if (!details.isValid) {
      recommendations.push('Fix validation errors before processing')

      // Specific recommendations based on error types
      const missingFields = details.errors.filter(e => e.code === 'MISSING_REQUIRED_FIELD')
      if (missingFields.length > 0) {
        recommendations.push(`Add missing required fields: ${missingFields.map(e => e.field).join(', ')}`)
      }

      const typeErrors = details.errors.filter(e => e.code === 'INVALID_TYPE')
      if (typeErrors.length > 0) {
        recommendations.push(`Correct data types for fields: ${typeErrors.map(e => e.field).join(', ')}`)
      }
    }

    if (details.warnings.length > 0) {
      recommendations.push('Review and address warnings for optimal performance')
    }

    if (migrationResult.migrations.length > 0) {
      recommendations.push('Update data source to use latest schema format')
    }

    return {
      isValid: details.isValid,
      errors: details.errors,
      warnings: details.warnings,
      migrationPerformed: migrationResult.migrations.length > 0,
      migrationDetails: migrationResult.migrations,
      version: details.version || '1.0',
      recommendations,
      performanceMetrics: {
        validationTime,
        migrationTime
      }
    }

  } catch (error) {
    const endTime = performance.now()

    return {
      isValid: false,
      errors: [{
        field: 'validation',
        message: `Comprehensive validation failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
        code: 'COMPREHENSIVE_VALIDATION_ERROR',
        severity: 'error'
      }],
      warnings: [],
      migrationPerformed: false,
      migrationDetails: [],
      version: 'unknown',
      recommendations: ['Check data structure and schema compatibility'],
      performanceMetrics: {
        validationTime: endTime - startTime,
        migrationTime: 0
      }
    }
  }
}

/**
 * Batch validation for multiple operation snapshots
 */
export interface BatchValidationResult {
  totalItems: number
  validItems: number
  invalidItems: number
  itemsWithWarnings: number
  migrationsPerformed: number
  errors: ValidationError[]
  warnings: string[]
  reports: ValidationReport[]
  summary: {
    averageValidationTime: number
    mostCommonErrors: Array<{ error: string; count: number }>
    recommendations: string[]
  }
}

export function validateOperationSnapshotsBatch(dataArray: any[]): BatchValidationResult {
  const reports: ValidationReport[] = []
  const allErrors: ValidationError[] = []
  const allWarnings: string[] = []
  let totalValidationTime = 0

  for (const data of dataArray) {
    const report = validateOperationSnapshotComprehensive(data)
    reports.push(report)

    allErrors.push(...report.errors)
    allWarnings.push(...report.warnings)
    totalValidationTime += report.performanceMetrics.validationTime
  }

  const totalItems = dataArray.length
  const validItems = reports.filter(r => r.isValid).length
  const invalidItems = totalItems - validItems
  const itemsWithWarnings = reports.filter(r => r.warnings.length > 0).length
  const migrationsPerformed = reports.filter(r => r.migrationPerformed).length

  // Analyze most common errors
  const errorCounts = new Map<string, number>()
  allErrors.forEach(error => {
    const key = `${error.field}:${error.code}`
    errorCounts.set(key, (errorCounts.get(key) || 0) + 1)
  })

  const mostCommonErrors = Array.from(errorCounts.entries())
    .map(([error, count]) => ({ error, count }))
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)

  // Generate batch recommendations
  const recommendations: string[] = []

  if (invalidItems > 0) {
    recommendations.push(`${invalidItems} items have validation errors that need attention`)
  }

  if (migrationsPerformed > 0) {
    recommendations.push(`${migrationsPerformed} items required migration - consider updating data sources`)
  }

  if (itemsWithWarnings > 0) {
    recommendations.push(`${itemsWithWarnings} items have warnings that should be reviewed`)
  }

  return {
    totalItems,
    validItems,
    invalidItems,
    itemsWithWarnings,
    migrationsPerformed,
    errors: allErrors,
    warnings: allWarnings,
    reports,
    summary: {
      averageValidationTime: totalValidationTime / totalItems,
      mostCommonErrors,
      recommendations
    }
  }
}

/**
 * Validation utility functions for common scenarios
 */
export const ValidationUtils = {
  /**
   * Validate a single field against schema definition
   */
  validateField: (fieldName: string, value: any, schema?: SchemaDefinition): ValidationError[] => {
    if (!schema) {
      schema = schemaRegistry.getLatest()
    }

    const fieldDef = schema.fields[fieldName]
    if (!fieldDef) {
      return [{
        field: fieldName,
        message: `Unknown field: ${fieldName}`,
        value,
        code: 'UNKNOWN_FIELD',
        severity: 'warning'
      }]
    }

    const registry = new OperationSchemaRegistry()
    return registry.validateField(fieldName, value, fieldDef)
  },

  /**
   * Check if data requires migration
   */
  requiresMigration: (data: any): boolean => {
    const sourceVersion = data.metadata?.version || '0.9'
    const latestVersion = schemaRegistry.getLatest().version
    return sourceVersion !== latestVersion
  },

  /**
   * Get migration path information
   */
  getMigrationPath: (fromVersion: string, toVersion?: string): string[] => {
    const target = toVersion || schemaRegistry.getLatest().version
    const registry = new OperationSchemaRegistry()
    return registry.findMigrationPath(fromVersion, target)
  },

  /**
   * Validate and format operation ID
   */
  validateOperationId: (id: string): ValidationError | null => {
    if (!id || typeof id !== 'string') {
      return {
        field: 'operation_id',
        message: 'Operation ID must be a non-empty string',
        value: id,
        code: 'INVALID_TYPE',
        severity: 'error'
      }
    }

    if (id.trim().length === 0) {
      return {
        field: 'operation_id',
        message: 'Operation ID cannot be empty',
        value: id,
        code: 'INVALID_VALUE',
        severity: 'error'
      }
    }

    if (id.length > 100) {
      return {
        field: 'operation_id',
        message: 'Operation ID too long (max 100 characters)',
        value: id,
        code: 'INVALID_LENGTH',
        severity: 'error'
      }
    }

    const pattern = /^[a-zA-Z0-9_-]+$/
    if (!pattern.test(id)) {
      return {
        field: 'operation_id',
        message: 'Operation ID can only contain letters, numbers, underscores, and hyphens',
        value: id,
        code: 'PATTERN_MISMATCH',
        severity: 'error'
      }
    }

    return null
  },

  /**
   * Validate progress value
   */
  validateProgress: (progress: number): ValidationError | null => {
    if (typeof progress !== 'number') {
      return {
        field: 'progress',
        message: 'Progress must be a number',
        value: progress,
        code: 'INVALID_TYPE',
        severity: 'error'
      }
    }

    if (!Number.isInteger(progress)) {
      return {
        field: 'progress',
        message: 'Progress must be an integer',
        value: progress,
        code: 'INVALID_VALUE',
        severity: 'error'
      }
    }

    if (progress < 0 || progress > 100) {
      return {
        field: 'progress',
        message: 'Progress must be between 0 and 100',
        value: progress,
        code: 'OUT_OF_RANGE',
        severity: 'warning'
      }
    }

    return null
  },

  /**
   * Validate status value
   */
  validateStatus: (status: string): ValidationError | null => {
    if (!status || typeof status !== 'string') {
      return {
        field: 'status',
        message: 'Status must be a non-empty string',
        value: status,
        code: 'INVALID_TYPE',
        severity: 'error'
      }
    }

    const validStatuses = ['pending', 'running', 'completed', 'failed', 'cancelled', 'unknown']
    if (!validStatuses.includes(status)) {
      return {
        field: 'status',
        message: `Invalid status: ${status}. Valid values: ${validStatuses.join(', ')}`,
        value: status,
        code: 'INVALID_ENUM_VALUE',
        severity: 'error'
      }
    }

    return null
  }
}
