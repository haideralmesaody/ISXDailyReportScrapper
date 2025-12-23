/**
 * Zod validation schemas for Reports feature
 * Following CLAUDE.md standards for input validation
 */

import { z } from 'zod'

// Report type enum schema
export const reportTypeSchema = z.enum(['daily', 'ticker', 'index', 'indexes', 'summary', 'liquidity', 'combined', 'all'])

// Report metadata schema
export const reportMetadataSchema = z.object({
  name: z.string().min(1),
  size: z.number().min(0),
  modified: z.string().datetime(), // ISO date string
  type: reportTypeSchema,
  displayName: z.string().min(1),
  path: z.string().min(1)
})

// Report file schema
export const reportFileSchema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  type: reportTypeSchema,
  size: z.number().min(0),
  modifiedDate: z.date(),
  downloadUrl: z.string().url()
})

// CSV data schema
export const csvDataSchema = z.object({
  headers: z.array(z.string()),
  rows: z.array(z.array(z.string())),
  totalRows: z.number().min(0)
})

// Parsed CSV data schema
export const parsedCSVDataSchema = z.object({
  columns: z.array(z.object({
    key: z.string(),
    header: z.string(),
    accessor: z.string()
  })),
  data: z.array(z.record(z.unknown()))
})

// API response schema - updated to match backend response
export const reportApiResponseSchema = z.object({
  status: z.string().optional(),
  data: z.array(z.object({
    name: z.string(),
    size: z.number(),
    modified: z.string(),
    path: z.string().optional(), // Nested path like "daily/2024/01/report.csv"
    category: z.string().optional(), // Category from backend
    fullPath: z.string().optional() // Full system path (for internal use)
  })),
  count: z.number()
})

// Download options schema
export const reportDownloadOptionsSchema = z.object({
  type: z.enum(['reports', 'downloads']),
  filename: z.string().min(1).max(255)
})

// RFC 7807 Problem Details schema
export const reportErrorSchema = z.object({
  type: z.string(),
  title: z.string(),
  status: z.number().min(100).max(599),
  detail: z.string(),
  instance: z.string(),
  trace_id: z.string()
})

// Request parameter schemas
export const getReportsParamsSchema = z.object({
  type: reportTypeSchema.optional(),
  search: z.string().optional(),
  sortBy: z.enum(['name', 'size', 'modified']).optional(),
  sortOrder: z.enum(['asc', 'desc']).optional()
})

export const downloadReportParamsSchema = z.object({
  type: z.enum(['reports', 'downloads']),
  filename: z.string()
    .min(1)
    .max(500) // Increased to support nested paths
    .regex(/^[a-zA-Z0-9_\-\.\/]+$/, 'Invalid path format') // Allow forward slashes for paths
})

// Type exports from schemas
export type ReportType = z.infer<typeof reportTypeSchema>
export type ReportMetadata = z.infer<typeof reportMetadataSchema>
export type ReportFile = z.infer<typeof reportFileSchema>
export type CSVData = z.infer<typeof csvDataSchema>
export type ParsedCSVData = z.infer<typeof parsedCSVDataSchema>
export type ReportApiResponse = z.infer<typeof reportApiResponseSchema>
export type ReportDownloadOptions = z.infer<typeof reportDownloadOptionsSchema>
export type ReportError = z.infer<typeof reportErrorSchema>
export type GetReportsParams = z.infer<typeof getReportsParamsSchema>
export type DownloadReportParams = z.infer<typeof downloadReportParamsSchema>