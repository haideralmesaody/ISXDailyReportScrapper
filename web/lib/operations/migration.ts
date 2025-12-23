/**
 * Operations Migration Utilities
 *
 * Enhanced migration functions for operation data structures that handle legacy payloads
 * and provide detailed error reporting.
 */

import { OperationSnapshotDTO, validateOperationSnapshot } from './validation'

export interface MigrationResult {
  success: boolean
  data?: OperationSnapshotDTO
  errors: string[]
  warnings: string[]
  migrated_from?: string
  migration_details?: {
    original_format?: string
    fields_added?: string[]
    fields_normalized?: string[]
    version?: string
  }
}

/**
 * Enhanced migration that handles both current format and legacy payloads.
 * Provides detailed error reporting and tracks what changes were made.
 */
export function migrateToOperationSnapshotDTO(data: any, options?: any): MigrationResult {
  // Basic validation - check if it looks like operation data
  if (!data || typeof data !== 'object') {
    return {
      success: false,
      errors: ['Invalid operation data structure: not an object'],
      warnings: [],
      migration_details: {
        original_format: 'unknown'
      }
    }
  }

  // Check for required operation_id field
  if (!data.operation_id) {
    return {
      success: false,
      errors: ['Missing required field: operation_id'],
      warnings: [],
      migration_details: {
        original_format: 'malformed'
      }
    }
  }

  const result: MigrationResult = {
    success: true,
    errors: [],
    warnings: [],
    migration_details: {
      fields_added: [],
      fields_normalized: [],
      version: data.metadata?.version || 'unknown'
    }
  }

  // Detect original format
  let originalFormat = 'current'
  if (!data.metadata) {
    originalFormat = 'legacy_v1'
  } else if (!data.metadata.version) {
    originalFormat = 'legacy_v2'
  }

  // Perform migration and normalization
  const migrated: any = { ...data }

  // Ensure all required fields exist
  const requiredFields = ['status', 'progress', 'current_stage', 'started_at', 'updated_at']
  for (const field of requiredFields) {
    if (!migrated[field]) {
      switch (field) {
        case 'status':
          migrated[field] = 'unknown'
          result.migration_details!.fields_added!.push(field + ' (defaulted)')
          break
        case 'progress':
          migrated[field] = 0
          result.migration_details!.fields_added!.push(field + ' (defaulted)')
          break
        case 'current_stage':
          migrated[field] = migrated.current_step || ''
          if (migrated.current_step) {
            result.migration_details!.fields_normalized!.push('current_step -> current_stage')
          } else {
            result.migration_details!.fields_added!.push(field + ' (defaulted)')
          }
          break
        case 'started_at':
        case 'updated_at':
          migrated[field] = new Date().toISOString()
          result.migration_details!.fields_added!.push(field + ' (defaulted)')
          break
      }
    }
  }

  // Normalize data types
  if (typeof migrated.progress !== 'number') {
    migrated.progress = parseFloat(migrated.progress) || 0
    result.migration_details!.fields_normalized!.push('progress (coerced to number)')
  }

  if (typeof migrated.operation_id !== 'string') {
    migrated.operation_id = String(migrated.operation_id)
    result.migration_details!.fields_normalized!.push('operation_id (coerced to string)')
  }

  // Ensure metadata exists
  if (!migrated.metadata) {
    migrated.metadata = {}
    result.migration_details!.fields_added!.push('metadata (empty object)')
  }

  // Add migration metadata
  migrated.metadata.migrated_at = new Date().toISOString()
  migrated.metadata.migration_from = originalFormat

  // Validate the migrated data
  const validation = validateOperationSnapshot(migrated)
  if (!validation.isValid) {
    result.errors = validation.errors.map(err => `${err.field}: ${err.message}`)
    result.success = false
  }

  result.warnings = validation.warnings
  result.data = migrated as OperationSnapshotDTO
  result.migrated_from = originalFormat

  // Add migration summary if there were changes
  if (result.migration_details!.fields_added!.length > 0 ||
      result.migration_details!.fields_normalized!.length > 0) {
    result.warnings.push(`Migration performed: ${result.migration_details!.fields_added!.length} fields added, ${result.migration_details!.fields_normalized!.length} fields normalized`)
  }

  return result
}