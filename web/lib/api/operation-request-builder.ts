/**
 * Operation Request Builder
 * Centralized request construction for ISX operations
 * Updated to support new simplified backend API structure
 * - run_stage: Execute individual stages with stage_id parameter
 * - full_pipeline: Execute all stages in sequence
 */

import { z } from 'zod'
import { OPERATION_DATE_DEFAULTS } from '@/lib/constants'

// Schema matching new backend expectations
export const OperationRequestSchema = z.object({
  operation: z.enum(['run_stage', 'full_pipeline']),
  stage_id: z.string().optional(), // Required for run_stage, optional for full_pipeline
  parameters: z.record(z.any()).optional()
})

export type OperationRequest = z.infer<typeof OperationRequestSchema>

/**
 * Builder class for creating properly structured operation requests
 * Updated to support new simplified backend API structure
 */
export class OperationRequestBuilder {

  /**
   * Build request for individual stage execution
   * Maps to "run_stage" operation with stage_id parameter
   */
  static buildRunStage(stageId: string, fromDate?: string, toDate?: string): OperationRequest {
    const parameters: any = {}

    // Add date parameters for scraping stage
    if (stageId === 'scraping') {
      const effectiveFromDate = fromDate || OPERATION_DATE_DEFAULTS.getFromDateString()
      const effectiveToDate = toDate || OPERATION_DATE_DEFAULTS.getToDateString()
      parameters.from = effectiveFromDate
      parameters.to = effectiveToDate
    }

    const request = {
      operation: 'run_stage' as const,
      stage_id: stageId,
      parameters
    }

    // Validate before returning
    return OperationRequestSchema.parse(request)
  }

  /**
   * Build request for full pipeline execution
   * Maps to "full_pipeline" operation
   */
  static buildFullPipeline(fromDate?: string, toDate?: string): OperationRequest {
    const effectiveFromDate = fromDate || OPERATION_DATE_DEFAULTS.getFromDateString()
    const effectiveToDate = toDate || OPERATION_DATE_DEFAULTS.getToDateString()

    const request = {
      operation: 'full_pipeline' as const,
      parameters: {
        from: effectiveFromDate,
        to: effectiveToDate
      }
    }

    return OperationRequestSchema.parse(request)
  }
  
  /**
   * Legacy compatibility method - buildQuickStart
   * Maps to new run_stage structure
   */
  static buildQuickStart(stageId: string): OperationRequest {
    return this.buildRunStage(stageId)
  }

  /**
   * Legacy compatibility method - buildWithDates
   * Maps to new run_stage structure with date parameters
   */
  static buildWithDates(
    stageId: string,
    fromDate?: string,
    toDate?: string
  ): OperationRequest {
    return this.buildRunStage(stageId, fromDate, toDate)
  }

  /**
   * Build request from generic parameters (used by FloatingConfigPanel)
   * Handles both individual stages and full pipeline
   */
  static buildFromParams(params: any): OperationRequest {
    // If this looks like new format
    if (params.operation) {
      return OperationRequestSchema.parse(params)
    }

    // Legacy format - convert to new format
    const stageId = params.step || params.stage_id || params.operation || 'processing'

    // Check if this is a full pipeline request
    if (stageId === 'full_pipeline' || params.operation === 'full_pipeline') {
      return this.buildFullPipeline(params.from, params.to)
    }

    // Individual stage request
    if (params.from && params.to) {
      return this.buildRunStage(stageId, params.from, params.to)
    } else {
      return this.buildRunStage(stageId)
    }
  }

  /**
   * Validate a request structure without building
   */
  static validate(request: unknown): OperationRequest {
    return OperationRequestSchema.parse(request)
  }

  /**
   * Check if a request is valid
   */
  static isValid(request: unknown): boolean {
    try {
      OperationRequestSchema.parse(request)
      return true
    } catch {
      return false
    }
  }
}

/**
 * Operation configuration defining which operations need user input
 * Updated to match new backend API structure
 */
export const OPERATIONS_CONFIG = {
  scraping: {
    requiresDates: true,
    quickStart: false,
    name: "Data Collection",
    description: "Download ISX daily reports for specified date range",
    icon: "Download"
  },
  processing: {
    requiresDates: false,
    quickStart: true,
    name: "Data Processing",
    description: "Convert Excel files to CSV format",
    icon: "FileSpreadsheet"
  },
  indices: {
    requiresDates: false,
    quickStart: true,
    name: "Index Extraction",
    description: "Extract ISX60 and ISX15 indices",
    icon: "BarChart3"
  },
  liquidity: {
    requiresDates: false,
    quickStart: true,
    name: "Liquidity Analysis",
    description: "Calculate ISX Hybrid Liquidity Metrics and scoring",
    icon: "Zap"
  },
  full_pipeline: {
    requiresDates: true,
    quickStart: false,
    name: "Full Pipeline",
    description: "Run complete data processing workflow",
    icon: "Workflow"
  }
} as const

export type OperationType = keyof typeof OPERATIONS_CONFIG
