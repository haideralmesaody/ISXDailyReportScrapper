/**
 * Zod validation schemas for ISX Daily Reports Scrapper
 * Professional form validation with TypeScript integration
 */

import { z } from 'zod'

// ============================================================================
// License Validation Schemas
// ============================================================================

export const licenseActivationSchema = z.object({
  license_key: z
    .string()
    .min(1, 'License key is required')
    .transform((val) => {
      const trimmed = val.trim().toUpperCase()
      // Keep dashes if it's a scratch card format
      if (trimmed.match(/^ISX-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/)) {
        return trimmed
      }
      // Otherwise clean for standard format
      return trimmed.replace(/[^A-Z0-9]/g, '')
    })
    .pipe(
      z.string()
        .min(10, 'License key must be at least 10 characters')
        .refine(
          (val) => {
            // Check scratch card format with dashes
            if (val.match(/^ISX-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/)) {
              return true
            }
            // Check standard format (no dashes)
            return val.match(/^ISX[A-Z0-9]+$/) !== null
          },
          'Invalid license key format'
        )
    ),
})

export type LicenseActivationForm = z.infer<typeof licenseActivationSchema>

// ============================================================================
// operation Configuration Schemas
// ============================================================================

export const OperationConfigSchema = z.object({
  auto_start: z.boolean().default(false),
  retry_attempts: z
    .number()
    .min(0, 'Retry attempts cannot be negative')
    .max(10, 'Maximum 10 retry attempts allowed')
    .default(3),
  timeout_seconds: z
    .number()
    .min(10, 'Minimum timeout is 10 seconds')
    .max(3600, 'Maximum timeout is 1 hour')
    .default(300),
  steps: z
    .array(z.string().min(1, 'step name cannot be empty'))
    .min(1, 'At least one step is required')
    .max(20, 'Maximum 20 steps allowed'),
  notification_email: z
    .string()
    .email('Must be a valid email address')
    .optional()
    .or(z.literal('')),
})

export type OperationConfigForm = z.infer<typeof OperationConfigSchema>

// ============================================================================
// Operation Request Schema (for API calls)
// ============================================================================

export const OperationStepSchema = z.object({
  id: z.string().min(1, 'Step ID is required'),
  type: z.string().min(1, 'Step type is required'),
  parameters: z.record(z.any()).optional()
})

export const OperationRequestSchema = z.object({
  mode: z.enum(['full', 'partial'], {
    errorMap: () => ({ message: 'Mode must be full or partial' })
  }),
  steps: z
    .array(OperationStepSchema)
    .min(1, 'At least one step is required'),
  parameters: z.record(z.any()).optional()
})

export type OperationRequest = z.infer<typeof OperationRequestSchema>
export type OperationStep = z.infer<typeof OperationStepSchema>

// ============================================================================
// File Upload Schemas
// ============================================================================

export const fileUploadSchema = z.object({
  file: z
    .instanceof(File)
    .refine((file) => file.size <= 10 * 1024 * 1024, 'File size must be less than 10MB')
    .refine(
      (file) => ['text/csv', 'application/json', 'text/plain'].includes(file.type),
      'Only CSV, JSON, and text files are allowed'
    ),
  type: z.enum(['data', 'config'], {
    required_error: 'File type is required',
  }),
  description: z
    .string()
    .min(5, 'Description must be at least 5 characters')
    .max(200, 'Description must be less than 200 characters')
    .optional()
    .or(z.literal('')),
})

export type FileUploadForm = z.infer<typeof fileUploadSchema>

// ============================================================================
// User Profile Schemas
// ============================================================================

export const userProfileSchema = z.object({
  name: z
    .string()
    .min(2, 'Name must be at least 2 characters')
    .max(50, 'Name must be less than 50 characters'),
  email: z
    .string()
    .email('Must be a valid email address'),
  organization: z
    .string()
    .min(2, 'Organization name must be at least 2 characters')
    .max(100, 'Organization name must be less than 100 characters')
    .optional()
    .or(z.literal('')),
  notifications: z.object({
    pipeline_complete: z.boolean().default(true),
    system_alerts: z.boolean().default(true),
    weekly_summary: z.boolean().default(false),
  }),
})

export type UserProfileForm = z.infer<typeof userProfileSchema>

// ============================================================================
// Report Filter Schemas
// ============================================================================

export const reportFilterSchema = z.object({
  ticker: z
    .string()
    .min(1, 'Ticker symbol must be at least 1 character')
    .max(10, 'Ticker symbol must be less than 10 characters')
    .regex(/^[A-Z0-9]+$/, 'Ticker must contain only uppercase letters and numbers')
    .optional()
    .or(z.literal('')),
  type: z
    .enum(['daily', 'summary', 'liquidity'], {
      errorMap: () => ({ message: 'Please select a valid report type' }),
    })
    .optional(),
  date_from: z
    .string()
    .regex(/^\d{4}-\d{2}-\d{2}$/, 'Date must be in YYYY-MM-DD format')
    .optional()
    .or(z.literal('')),
  date_to: z
    .string()
    .regex(/^\d{4}-\d{2}-\d{2}$/, 'Date must be in YYYY-MM-DD format')
    .optional()
    .or(z.literal('')),
})
.refine(
  (data) => {
    if (data.date_from && data.date_to) {
      return new Date(data.date_from) <= new Date(data.date_to)
    }
    return true
  },
  {
    message: 'Start date must be before or equal to end date',
    path: ['date_to'],
  }
)

export type ReportFilterForm = z.infer<typeof reportFilterSchema>

// ============================================================================
// System Settings Schemas
// ============================================================================

export const systemSettingsSchema = z.object({
  api_timeout: z
    .number()
    .min(5, 'Minimum API timeout is 5 seconds')
    .max(300, 'Maximum API timeout is 5 minutes')
    .default(30),
  max_concurrent_pipelines: z
    .number()
    .min(1, 'At least 1 concurrent operation required')
    .max(10, 'Maximum 10 concurrent operations allowed')
    .default(3),
  log_level: z
    .enum(['debug', 'info', 'warn', 'error'], {
      errorMap: () => ({ message: 'Please select a valid log level' }),
    })
    .default('info'),
  auto_cleanup_days: z
    .number()
    .min(1, 'Minimum cleanup period is 1 day')
    .max(365, 'Maximum cleanup period is 1 year')
    .default(30),
  websocket_reconnect: z.boolean().default(true),
})

export type SystemSettingsForm = z.infer<typeof systemSettingsSchema>

// ============================================================================
// Common Validation Helpers
// ============================================================================

/**
 * Validates Iraqi Investor license key format
 */
export function isValidLicenseKey(key: string): boolean {
  // Simple validation - just check if it starts with ISX and has reasonable length
  return /^ISX[A-Z0-9]{7,}$/.test(key) // ISX + at least 7 more characters = 10 total minimum
}

/**
 * Validates ticker symbol format
 */
export function isValidTickerSymbol(symbol: string): boolean {
  return /^[A-Z0-9]{1,10}$/.test(symbol)
}

/**
 * Validates date string in YYYY-MM-DD format
 */
export function isValidDateString(dateStr: string): boolean {
  return /^\d{4}-\d{2}-\d{2}$/.test(dateStr) && !isNaN(Date.parse(dateStr))
}

/**
 * Creates a date range validation schema
 */
export function createDateRangeSchema(maxDays = 365) {
  return z.object({
    start_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
    end_date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
  }).refine(
    (data) => {
      const start = new Date(data.start_date)
      const end = new Date(data.end_date)
      const diffInDays = (end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)
      
      return start <= end && diffInDays <= maxDays
    },
    {
      message: `Date range cannot exceed ${maxDays} days and start date must be before end date`,
      path: ['end_date'],
    }
  )
}

// ============================================================================
// Error Handling Schema
// ============================================================================

export const apiErrorSchema = z.object({
  type: z.string(),
  title: z.string(),
  status: z.number(),
  detail: z.string().optional(),
  trace_id: z.string().optional(),
})

export type ApiErrorResponse = z.infer<typeof apiErrorSchema>