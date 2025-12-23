/**
 * Stage Definitions Constants
 *
 * Single source of truth for stage information
 * No side effects or complex logic - pure constants only
 */

// Stage definitions with display information - SINGLE SOURCE OF TRUTH
// Names must exactly match backend StageName constants in api/internal/operations/types.go
export const STAGE_DEFINITIONS = {
  scraping: {
    name: 'Data Collection',
    description: 'Download ISX daily reports',
    category: 'data_acquisition',
    color: 'blue',
    icon: 'Download',
    order: 1
  },
  processing: {
    name: 'Data Processing',
    description: 'Process and transform market data',
    category: 'data_transformation',
    color: 'purple',
    icon: 'Database',
    order: 2
  },
  indices: {
    name: 'Index Extraction',
    description: 'Extract ISX60 and ISX15 indices',
    category: 'data_extraction',
    color: 'purple',
    icon: 'BarChart3',
    order: 3
  },
  liquidity: {
    name: 'Liquidity Calculation',
    description: 'Calculate liquidity metrics',
    category: 'analysis',
    color: 'orange',
    icon: 'Zap',
    order: 4
  },
  indicators: {
    name: 'Indicators Pre-calculation',
    description: 'Pre-compute indicators for SSOT',
    category: 'calculation',
    color: 'pink',
    icon: 'Activity',
    order: 6
  },
  analysis: {
    name: 'Technical Analysis',
    description: 'Comprehensive technical and market analysis',
    category: 'analysis',
    color: 'indigo',
    icon: 'TrendingUp',
    order: 5
  },
  full_pipeline: {
    name: 'Full Pipeline',
    description: 'Run all stages in sequence',
    category: 'pipeline',
    color: 'emerald',
    icon: 'Workflow',
    order: 0
  }
} as const

// Operation type to stage mapping - SINGLE SOURCE OF TRUTH
// Maps operation types to canonical stage IDs
export const OPERATION_TYPE_TO_STAGE: Record<string, keyof typeof STAGE_DEFINITIONS> = {
  'scraping': 'scraping',
  'processing': 'processing',
  'indices': 'indices',
  'liquidity': 'liquidity',
  'analysis': 'analysis',
  'indicators': 'indicators',
  'full_pipeline': 'full_pipeline'
}

// Stage identification patterns
export const STAGE_PATTERNS = {
  scraping: [
    /scraping/i,
    /scrape/i,
    /download/i,
    /isx.*reports?/i,
    /daily.*reports?/i,
    /market.*reports?/i
  ],
  processing: [
    /processing/i,
    /process/i,
    /data.*processing/i,
    /data.*process/i,
    /excel.*csv/i,
    /csv.*conversion/i,
    /file.*conversion/i,
    /market.*data/i,
    /raw.*data/i,
    /data.*transform/i
  ],
  indices: [
    /indices/i,
    /index/i,
    /isx60/i,
    /isx15/i,
    /market.*index/i
  ],
  liquidity: [
    /liquidity/i,
    /liquidity.*analysis/i,
    /liquidity.*metrics/i,
    /market.*liquidity/i
  ],
  analysis: [
    /analysis/i,
    /technical/i,
    /technical.*analysis/i,
    /market.*analysis/i
  ],
  indicators: [
    /indicators/i,
    /indicator/i,
    /technical.*indicators/i,
    /ssot/i,
    /pre.*compute/i
  ],
  full_pipeline: [
    /pipeline/i,
    /full.*pipeline/i,
    /complete/i,
    /all.*stages/i,
    /end.*to.*end/i
  ]
} as const

// Stage-specific metrics configuration
export const STAGE_METRICS = {
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
  analysis: {
    primaryMetric: 'analysis_complete',
    primaryLabel: 'Analysis Complete',
    secondaryMetric: 'insights_generated',
    secondaryLabel: 'Insights Generated'
  },
  indicators: {
    primaryMetric: 'indicators_calculated',
    primaryLabel: 'Indicators Calculated',
    secondaryMetric: 'tickers_processed',
    secondaryLabel: 'Tickers Processed'
  },
  full_pipeline: {
    primaryMetric: 'stages_completed',
    primaryLabel: 'Stages Completed',
    secondaryMetric: 'total_duration',
    secondaryLabel: 'Total Duration'
  }
} as const

export type StageId = keyof typeof STAGE_DEFINITIONS
export type ConfidenceLevel = 'high' | 'medium' | 'low' | 'unknown'

// Union types for cache values to maintain type safety
export type StageDefinition = (typeof STAGE_DEFINITIONS)[StageId]
export type StageMetrics = (typeof STAGE_METRICS)[StageId]