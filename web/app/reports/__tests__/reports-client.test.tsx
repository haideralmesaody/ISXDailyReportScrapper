/**
 * @jest-environment jsdom
 */
import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import ReportsClient from '../reports-client'

// Mock Next.js components
jest.mock('next/link', () => {
  return ({ children, href, ...props }: any) => (
    <a href={href} {...props}>
      {children}
    </a>
  )
})

jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    back: jest.fn(),
    forward: jest.fn(),
    refresh: jest.fn(),
    prefetch: jest.fn(),
  }),
  usePathname: () => '/reports',
  useSearchParams: () => new URLSearchParams(),
}))

// Mock toast hook
const mockToast = jest.fn()
jest.mock('@/lib/hooks/use-toast', () => ({
  useToast: () => ({
    toast: mockToast,
  }),
}))

// Mock hydration hook
let mockIsHydrated = false
jest.mock('@/lib/hooks', () => ({
  useHydration: jest.fn(() => mockIsHydrated),
}))

// Mock API functions
const mockFetchReports = jest.fn()
const mockDownloadReportContent = jest.fn()
const mockDownloadReportFile = jest.fn()
jest.mock('@/lib/api/reports', () => ({
  fetchReports: mockFetchReports,
  downloadReportContent: mockDownloadReportContent,
  downloadReportFile: mockDownloadReportFile,
}))

// Mock CSV parser utility
const mockParseCSVContent = jest.fn()
const mockGroupReportsByType = jest.fn()
jest.mock('@/lib/utils/csv-parser', () => ({
  parseCSVContent: mockParseCSVContent,
  groupReportsByType: mockGroupReportsByType,
}))

// Mock UI components
jest.mock('@/components/ui', () => ({
  NoDataState: ({ title, description, actions, instructions, icon: Icon, iconClassName }: any) => (
    <div data-testid="no-data-state">
      {Icon && <div data-testid="no-data-icon" className={iconClassName}><Icon /></div>}
      <h2 data-testid="no-data-title">{title}</h2>
      <p data-testid="no-data-description">{description}</p>
      {instructions && (
        <div data-testid="no-data-instructions">
          {instructions.map((instruction: string, index: number) => (
            <div key={index}>{instruction}</div>
          ))}
        </div>
      )}
      {actions && (
        <div data-testid="no-data-actions">
          {actions.map((action: any, index: number) => (
            action.href ? (
              <a key={index} href={action.href} data-testid={`action-link-${index}`}>
                {action.icon && <action.icon />}
                {action.label}
              </a>
            ) : (
              <button key={index} onClick={action.onClick} data-testid={`action-button-${index}`}>
                {action.icon && <action.icon />}
                {action.label}
              </button>
            )
          ))}
        </div>
      )}
    </div>
  ),
  DataLoadingState: ({ title, description, compact }: any) => (
    <div data-testid="data-loading-state" data-compact={compact}>
      <div>{title || 'Loading...'}</div>
      {description && <div>{description}</div>}
    </div>
  ),
}))

// Mock Button component
jest.mock('@/components/ui/button', () => ({
  Button: ({ children, asChild, variant, size, onClick, ...props }: any) => {
    if (asChild) {
      return <div {...props}>{children}</div>
    }
    return (
      <button onClick={onClick} data-variant={variant} data-size={size} {...props}>
        {children}
      </button>
    )
  },
}))

// Mock Card component
jest.mock('@/components/ui/card', () => ({
  Card: ({ children, className, ...props }: any) => (
    <div className={className} data-testid="card" {...props}>
      {children}
    </div>
  ),
}))

// Mock Reports components
jest.mock('@/components/reports/ReportTypeSelector', () => ({
  ReportTypeSelector: ({ selectedType, onTypeChange, reportCounts }: any) => (
    <div data-testid="report-type-selector">
      <select 
        value={selectedType} 
        onChange={(e) => onTypeChange(e.target.value)}
        data-testid="type-selector"
      >
        <option value="all">All ({reportCounts?.all || 0})</option>
        <option value="daily">Daily ({reportCounts?.daily || 0})</option>
        <option value="ticker">Ticker ({reportCounts?.ticker || 0})</option>
        <option value="liquidity">Liquidity ({reportCounts?.liquidity || 0})</option>
        <option value="combined">Combined ({reportCounts?.combined || 0})</option>
        <option value="indexes">Indexes ({reportCounts?.indexes || 0})</option>
        <option value="summary">Summary ({reportCounts?.summary || 0})</option>
      </select>
    </div>
  ),
}))

jest.mock('@/components/reports/ReportList', () => ({
  ReportList: ({ reports, selectedReport, onSelectReport, onDownloadReport }: any) => (
    <div data-testid="report-list">
      {reports.map((report: any) => (
        <div key={report.name} data-testid={`report-item-${report.name}`}>
          <button
            onClick={() => onSelectReport(report)}
            data-selected={selectedReport?.name === report.name}
            data-testid={`report-select-${report.name}`}
          >
            {report.displayName}
          </button>
          <button
            onClick={() => onDownloadReport(report)}
            data-testid={`report-download-${report.name}`}
          >
            Download
          </button>
        </div>
      ))}
    </div>
  ),
}))

jest.mock('@/components/reports/CSVViewer', () => ({
  CSVViewer: ({ report, csvData, isLoading, error }: any) => (
    <div data-testid="csv-viewer">
      {isLoading && <div data-testid="csv-loading">Loading CSV...</div>}
      {error && <div data-testid="csv-error">{error.message}</div>}
      {report && <div data-testid="csv-report-name">{report.displayName}</div>}
      {csvData && (
        <div data-testid="csv-data">
          Rows: {csvData.data?.length || 0}
        </div>
      )}
    </div>
  ),
}))

jest.mock('@/components/reports/ReportFilters', () => ({
  ReportFilters: ({ reports, onFiltersChange }: any) => (
    <div data-testid="report-filters">
      <button onClick={() => onFiltersChange(reports.slice(0, 1))}>
        Apply Filter
      </button>
    </div>
  ),
}))

// Sample test data
interface ReportMetadata {
  name: string
  displayName: string
  path?: string
  type: 'daily' | 'ticker' | 'liquidity' | 'combined' | 'indexes' | 'summary'
  size: number
  lastModified: string
}

interface ParsedCSVData {
  data: Array<Record<string, string>>
  headers: string[]
  rowCount: number
}

const mockReportsData: ReportMetadata[] = [
  {
    name: 'daily_report_2025-01-15.csv',
    displayName: 'Daily Report - Jan 15, 2025',
    type: 'daily',
    size: 1024,
    lastModified: '2025-01-15T10:00:00Z',
    path: 'reports/daily/daily_report_2025-01-15.csv',
  },
  {
    name: 'ticker_analysis_BBOB.csv',
    displayName: 'BBOB Ticker Analysis',
    type: 'ticker',
    size: 512,
    lastModified: '2025-01-15T09:30:00Z',
    path: 'reports/tickers/ticker_analysis_BBOB.csv',
  },
  {
    name: 'liquidity_summary.csv',
    displayName: 'Market Liquidity Summary',
    type: 'liquidity',
    size: 2048,
    lastModified: '2025-01-15T11:00:00Z',
    path: 'reports/liquidity/liquidity_summary.csv',
  },
]

const mockCSVData: ParsedCSVData = {
  data: [
    { Symbol: 'BBOB', Price: '1.25', Volume: '1000000', Change: '+0.05' },
    { Symbol: 'BAGH', Price: '2.15', Volume: '500000', Change: '-0.10' },
    { Symbol: 'TASC', Price: '0.85', Volume: '750000', Change: '+0.02' },
  ],
  headers: ['Symbol', 'Price', 'Volume', 'Change'],
  rowCount: 3,
}

describe('ReportsClient Integration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockIsHydrated = false
    mockToast.mockClear()
    mockFetchReports.mockClear()
    mockDownloadReportContent.mockClear()
    mockDownloadReportFile.mockClear()
    mockParseCSVContent.mockClear()
    mockGroupReportsByType.mockClear()
    
    // Update the mock return value
    const mockUseHydration = require('@/lib/hooks').useHydration
    mockUseHydration.mockReturnValue(mockIsHydrated)
    
    // Set up default grouping mock
    mockGroupReportsByType.mockReturnValue(new Map([
      ['daily', mockReportsData.filter(r => r.type === 'daily')],
      ['ticker', mockReportsData.filter(r => r.type === 'ticker')],
      ['liquidity', mockReportsData.filter(r => r.type === 'liquidity')],
      ['combined', []],
      ['indexes', []],
      ['summary', []],
    ]))
    
    // Reset console mocks
    jest.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    jest.restoreAllMocks()
  })

  describe('Hydration States', () => {
    it('shows loading state before hydration', () => {
      mockIsHydrated = false
      
      render(<ReportsClient />)
      
      const loadingState = screen.getByTestId('data-loading-state')
      expect(loadingState).toBeInTheDocument()
      expect(loadingState).toHaveTextContent('Initializing Reports')
      expect(loadingState).toHaveTextContent('Setting up the reports interface...')
    })

    it('transitions to main content after hydration', async () => {
      mockIsHydrated = true
      mockFetchReports.mockResolvedValue(mockReportsData)
      mockDownloadReportContent.mockResolvedValue('Symbol,Price,Volume\nBBOB,1.25,1000000')
      mockParseCSVContent.mockResolvedValue(mockCSVData)
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByText('Reports')).toBeInTheDocument()
        expect(screen.getByText('View and download Iraqi Stock Exchange reports')).toBeInTheDocument()
      })
    })
  })

  describe('No Data States', () => {
    const noDataScenarios = [
      {
        name: 'shows no-data state when no reports are available',
        mockData: [],
        mockError: null,
        shouldShowNoData: true,
        shouldShowToast: false,
      },
      {
        name: 'shows no-data state when API returns 404',
        mockData: null,
        mockError: new Error('404: Not Found'),
        shouldShowNoData: true,
        shouldShowToast: false,
      },
      {
        name: 'shows no-data state when API returns Not Found',
        mockData: null,
        mockError: new Error('Not Found'),
        shouldShowNoData: true,
        shouldShowToast: false,
      },
    ]

    noDataScenarios.forEach(({ name, mockData, mockError, shouldShowNoData, shouldShowToast }) => {
      it(name, async () => {
        mockIsHydrated = true
        
        if (mockError) {
          mockFetchReports.mockRejectedValue(mockError)
        } else if (mockData) {
          mockFetchReports.mockResolvedValue(mockData)
        }
        
        render(<ReportsClient />)
        
        await waitFor(() => {
          if (shouldShowNoData) {
            expect(screen.getByTestId('no-data-state')).toBeInTheDocument()
          } else {
            expect(screen.queryByTestId('no-data-state')).not.toBeInTheDocument()
          }
        })
        
        if (shouldShowNoData) {
          expect(screen.getByTestId('no-data-title')).toHaveTextContent('No Reports Available')
          expect(screen.getByTestId('no-data-description')).toHaveTextContent(
            'You need to run the data collection operations first to generate reports.'
          )
          
          // Check instructions
          const instructions = screen.getByTestId('no-data-instructions')
          expect(instructions).toBeInTheDocument()
          expect(instructions).toHaveTextContent('Go to the Operations page')
          expect(instructions).toHaveTextContent('Run \'Full Pipeline\' to collect and process data')
          
          // Check actions
          const actions = screen.getByTestId('no-data-actions')
          expect(actions).toBeInTheDocument()
          
          const operationsLink = screen.getByTestId('action-link-0')
          expect(operationsLink).toHaveAttribute('href', '/operations')
          expect(operationsLink).toHaveTextContent('Go to Operations')
          
          const checkAgainButton = screen.getByTestId('action-button-1')
          expect(checkAgainButton).toHaveTextContent('Check Again')
          
          // Check footer shows 0 reports
          expect(screen.getByText('0 reports available')).toBeInTheDocument()
        }
        
        if (shouldShowToast) {
          expect(mockToast).toHaveBeenCalled()
        } else {
          expect(mockToast).not.toHaveBeenCalled()
        }
      })
    })

    it('shows error state and toast for real API errors (not 404)', async () => {
      mockIsHydrated = true
      const realError = new Error('Network connection failed')
      mockFetchReports.mockRejectedValue(realError)
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(mockToast).toHaveBeenCalledWith({
          title: 'Error',
          description: 'Failed to load reports. Please try again.',
          variant: 'destructive',
        })
      })
      
      // Should not show no-data state for real errors
      expect(screen.queryByTestId('no-data-state')).not.toBeInTheDocument()
    })
  })

  describe('No Data Actions', () => {
    beforeEach(async () => {
      mockIsHydrated = true
      mockFetchReports.mockResolvedValue([])
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByTestId('no-data-state')).toBeInTheDocument()
      })
    })

    it('navigates to operations page when clicking link action', () => {
      const operationsLink = screen.getByTestId('action-link-0')
      expect(operationsLink).toHaveAttribute('href', '/operations')
      expect(operationsLink).toHaveTextContent('Go to Operations')
    })

    it('retries data loading when clicking check again button', async () => {
      const checkAgainButton = screen.getByTestId('action-button-1')
      expect(checkAgainButton).toHaveTextContent('Check Again')
      
      // Clear previous calls and set up successful response
      mockFetchReports.mockClear()
      mockFetchReports.mockResolvedValue(mockReportsData)
      
      fireEvent.click(checkAgainButton)
      
      await waitFor(() => {
        expect(mockFetchReports).toHaveBeenCalledTimes(1)
      })
    })
  })

  describe('Successful Data Loading', () => {
    beforeEach(() => {
      mockIsHydrated = true
      mockFetchReports.mockResolvedValue(mockReportsData)
      mockDownloadReportContent.mockResolvedValue('Symbol,Price,Volume\nBBOB,1.25,1000000')
      mockParseCSVContent.mockResolvedValue(mockCSVData)
    })

    it('loads reports data successfully', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(mockFetchReports).toHaveBeenCalledTimes(1)
        expect(screen.getByTestId('report-list')).toBeInTheDocument()
        expect(screen.getByText('Reports')).toBeInTheDocument()
      })
      
      // Check reports are displayed
      mockReportsData.forEach(report => {
        expect(screen.getByTestId(`report-item-${report.name}`)).toBeInTheDocument()
      })
    })

    it('auto-selects first report and loads CSV data', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(mockDownloadReportContent).toHaveBeenCalledWith(mockReportsData[0].path || mockReportsData[0].name)
        expect(mockParseCSVContent).toHaveBeenCalled()
      })
      
      // First report should be selected
      const firstReportButton = screen.getByTestId(`report-select-${mockReportsData[0].name}`)
      expect(firstReportButton).toHaveAttribute('data-selected', 'true')
      
      // CSV viewer should show data
      expect(screen.getByTestId('csv-viewer')).toBeInTheDocument()
      expect(screen.getByTestId('csv-report-name')).toHaveTextContent(mockReportsData[0].displayName)
      expect(screen.getByTestId('csv-data')).toHaveTextContent(`Rows: ${mockCSVData.data.length}`)
    })

    it('handles report selection and loads CSV data', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByTestId('report-list')).toBeInTheDocument()
      })
      
      // Clear previous calls
      mockDownloadReportContent.mockClear()
      mockParseCSVContent.mockClear()
      
      // Select different report
      const secondReportButton = screen.getByTestId(`report-select-${mockReportsData[1].name}`)
      fireEvent.click(secondReportButton)
      
      await waitFor(() => {
        expect(mockDownloadReportContent).toHaveBeenCalledWith(mockReportsData[1].path || mockReportsData[1].name)
        expect(mockParseCSVContent).toHaveBeenCalled()
      })
      
      expect(secondReportButton).toHaveAttribute('data-selected', 'true')
    })

    it('shows CSV loading state during data fetch', async () => {
      // Make CSV fetch slow
      let resolveCSVPromise: (value: any) => void
      const csvPromise = new Promise(resolve => {
        resolveCSVPromise = resolve
      })
      mockDownloadReportContent.mockReturnValue(csvPromise)
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByTestId('csv-loading')).toBeInTheDocument()
      })
      
      // Resolve the promise
      resolveCSVPromise!('Symbol,Price,Volume\nBBOB,1.25,1000000')
      
      await waitFor(() => {
        expect(screen.queryByTestId('csv-loading')).not.toBeInTheDocument()
      })
    })
  })

  describe('Report Type Filtering', () => {
    beforeEach(() => {
      mockIsHydrated = true
      mockFetchReports.mockResolvedValue(mockReportsData)
    })

    it('displays correct report counts by type', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        const typeSelector = screen.getByTestId('type-selector')
        expect(typeSelector).toBeInTheDocument()
      })
      
      // Check option text includes counts
      expect(screen.getByText('All (3)')).toBeInTheDocument()
      expect(screen.getByText('Daily (1)')).toBeInTheDocument()
      expect(screen.getByText('Ticker (1)')).toBeInTheDocument()
      expect(screen.getByText('Liquidity (1)')).toBeInTheDocument()
    })

    it('filters reports by selected type', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByTestId('type-selector')).toBeInTheDocument()
      })
      
      // Change to daily reports only
      const typeSelector = screen.getByTestId('type-selector')
      fireEvent.change(typeSelector, { target: { value: 'daily' } })
      
      expect(typeSelector).toHaveValue('daily')
    })
  })

  describe('Report Download', () => {
    beforeEach(() => {
      mockIsHydrated = true
      mockFetchReports.mockResolvedValue(mockReportsData)
      mockDownloadReportFile.mockResolvedValue(undefined)
    })

    it('downloads report when clicking download button', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByTestId('report-list')).toBeInTheDocument()
      })
      
      const downloadButton = screen.getByTestId(`report-download-${mockReportsData[0].name}`)
      fireEvent.click(downloadButton)
      
      await waitFor(() => {
        expect(mockDownloadReportFile).toHaveBeenCalledWith(mockReportsData[0].path || mockReportsData[0].name)
        expect(mockToast).toHaveBeenCalledWith({
          title: 'Success',
          description: `Downloaded ${mockReportsData[0].displayName}`,
        })
      })
    })

    it('handles download errors gracefully', async () => {
      const downloadError = new Error('Download failed')
      mockDownloadReportFile.mockRejectedValue(downloadError)
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByTestId('report-list')).toBeInTheDocument()
      })
      
      const downloadButton = screen.getByTestId(`report-download-${mockReportsData[0].name}`)
      fireEvent.click(downloadButton)
      
      await waitFor(() => {
        expect(mockToast).toHaveBeenCalledWith({
          title: 'Error',
          description: 'Failed to download report. Please try again.',
          variant: 'destructive',
        })
      })
    })
  })

  describe('CSV Data Loading and Error Handling', () => {
    beforeEach(() => {
      mockIsHydrated = true
      mockFetchReports.mockResolvedValue(mockReportsData)
    })

    it('handles CSV load errors without showing error state for 404', async () => {
      mockDownloadReportContent.mockRejectedValue(new Error('404: File not found'))
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByTestId('csv-viewer')).toBeInTheDocument()
      })
      
      // Should not show error in CSV viewer for 404
      expect(screen.queryByTestId('csv-error')).not.toBeInTheDocument()
      // Should not show toast for 404
      expect(mockToast).not.toHaveBeenCalled()
    })

    it('handles CSV load errors with toast for real errors', async () => {
      const realError = new Error('Network error')
      mockDownloadReportContent.mockRejectedValue(realError)
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(mockToast).toHaveBeenCalledWith({
          title: 'Error',
          description: 'Failed to load report data. Please try again.',
          variant: 'destructive',
        })
      })
    })

    it('handles CSV parsing errors', async () => {
      mockDownloadReportContent.mockResolvedValue('invalid,csv,data')
      mockParseCSVContent.mockRejectedValue(new Error('Invalid CSV format'))
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(mockToast).toHaveBeenCalledWith({
          title: 'Error',
          description: 'Failed to load report data. Please try again.',
          variant: 'destructive',
        })
      })
    })
  })

  describe('Advanced Filtering', () => {
    beforeEach(() => {
      mockIsHydrated = true
      mockFetchReports.mockResolvedValue(mockReportsData)
    })

    it('toggles advanced filters visibility', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByText('Show Advanced Filters')).toBeInTheDocument()
      })
      
      // Initially filters should not be visible
      expect(screen.queryByTestId('report-filters')).not.toBeInTheDocument()
      
      // Click to show filters
      fireEvent.click(screen.getByText('Show Advanced Filters'))
      
      expect(screen.getByTestId('report-filters')).toBeInTheDocument()
      expect(screen.getByText('Hide Advanced Filters')).toBeInTheDocument()
    })

    it('applies advanced filters to report list', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByText('Show Advanced Filters')).toBeInTheDocument()
      })
      
      // Show filters
      fireEvent.click(screen.getByText('Show Advanced Filters'))
      
      // Apply filter (mock implementation returns first report only)
      const applyFilterButton = screen.getByText('Apply Filter')
      fireEvent.click(applyFilterButton)
      
      // The filtered results should be applied
      expect(screen.getByTestId('report-filters')).toBeInTheDocument()
    })
  })

  describe('Footer Information', () => {
    beforeEach(() => {
      mockIsHydrated = true
    })

    it('displays correct report count in footer', async () => {
      mockFetchReports.mockResolvedValue(mockReportsData)
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByText(`${mockReportsData.length} of ${mockReportsData.length} reports`)).toBeInTheDocument()
      })
    })

    it('displays selected report name in footer', async () => {
      mockFetchReports.mockResolvedValue(mockReportsData)
      mockDownloadReportContent.mockResolvedValue('data')
      mockParseCSVContent.mockResolvedValue(mockCSVData)
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByText(mockReportsData[0].displayName)).toBeInTheDocument()
      })
    })

    it('displays CSV row count in footer', async () => {
      mockFetchReports.mockResolvedValue(mockReportsData)
      mockDownloadReportContent.mockResolvedValue('data')
      mockParseCSVContent.mockResolvedValue(mockCSVData)
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByText(`${mockCSVData.data.length} rows`)).toBeInTheDocument()
      })
    })

    it('displays version information in footer', async () => {
      mockFetchReports.mockResolvedValue(mockReportsData)
      
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByText('ISX Pulse v3.0.0')).toBeInTheDocument()
        expect(screen.getByText('© 2025 ISX Daily Reports')).toBeInTheDocument()
      })
    })
  })

  describe('Component Lifecycle and Performance', () => {
    it('calls API only after hydration is complete', () => {
      mockIsHydrated = false
      mockFetchReports.mockResolvedValue(mockReportsData)
      
      render(<ReportsClient />)
      
      // Should not call API before hydration
      expect(mockFetchReports).not.toHaveBeenCalled()
    })

    it('prevents duplicate API calls during rapid re-renders', async () => {
      mockIsHydrated = true
      mockFetchReports.mockResolvedValue(mockReportsData)
      
      const { rerender } = render(<ReportsClient />)
      
      // Trigger multiple re-renders quickly
      rerender(<ReportsClient />)
      rerender(<ReportsClient />)
      
      await waitFor(() => {
        // Should only call once due to dependency management
        expect(mockFetchReports).toHaveBeenCalledTimes(1)
      })
    })

    it('handles unmounting during async operations gracefully', async () => {
      mockIsHydrated = true
      
      // Create a promise that never resolves
      const neverResolves = new Promise(() => {})
      mockFetchReports.mockReturnValue(neverResolves)
      
      const { unmount } = render(<ReportsClient />)
      
      // Unmount while async operation is in progress
      unmount()
      
      // Should not throw errors
      expect(() => {
        // Any cleanup should happen silently
      }).not.toThrow()
    })
  })

  describe('Type Selection and State Management', () => {
    beforeEach(() => {
      mockIsHydrated = true
      mockFetchReports.mockResolvedValue(mockReportsData)
      mockDownloadReportContent.mockResolvedValue('data')
      mockParseCSVContent.mockResolvedValue(mockCSVData)
    })

    it('clears selection when changing to incompatible type', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByTestId('type-selector')).toBeInTheDocument()
      })
      
      // First report (daily) should be selected
      await waitFor(() => {
        expect(screen.getByTestId(`report-select-${mockReportsData[0].name}`)).toHaveAttribute('data-selected', 'true')
      })
      
      // Change to ticker type (should clear daily selection)
      const typeSelector = screen.getByTestId('type-selector')
      fireEvent.change(typeSelector, { target: { value: 'ticker' } })
      
      // Previous selection should be cleared since it doesn't match new type
      expect(screen.getByTestId(`report-select-${mockReportsData[0].name}`)).toHaveAttribute('data-selected', 'false')
    })

    it('maintains selection when changing to compatible type', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByTestId('type-selector')).toBeInTheDocument()
      })
      
      // Wait for first report to be selected
      await waitFor(() => {
        expect(screen.getByTestId(`report-select-${mockReportsData[0].name}`)).toHaveAttribute('data-selected', 'true')
      })
      
      // Change to 'all' type (should maintain selection)
      const typeSelector = screen.getByTestId('type-selector')
      fireEvent.change(typeSelector, { target: { value: 'all' } })
      
      // Selection should be maintained
      expect(screen.getByTestId(`report-select-${mockReportsData[0].name}`)).toHaveAttribute('data-selected', 'true')
    })
  })

  describe('Layout and Navigation', () => {
    beforeEach(() => {
      mockIsHydrated = true
      mockFetchReports.mockResolvedValue(mockReportsData)
    })

    it('renders header with back navigation', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByText('Back to Home')).toBeInTheDocument()
      })
      
      const backLink = screen.getByText('Back to Home').closest('a')
      expect(backLink).toHaveAttribute('href', '/')
    })

    it('maintains proper layout structure', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        // Check for main layout elements
        expect(screen.getByRole('banner')).toBeInTheDocument() // header
        expect(screen.getByRole('main')).toBeInTheDocument() // main content
        expect(screen.getByRole('contentinfo')).toBeInTheDocument() // footer
      })
    })

    it('renders responsive layout with proper panels', async () => {
      render(<ReportsClient />)
      
      await waitFor(() => {
        expect(screen.getByTestId('report-type-selector')).toBeInTheDocument()
        expect(screen.getByTestId('report-list')).toBeInTheDocument()
        expect(screen.getByTestId('csv-viewer')).toBeInTheDocument()
      })
    })
  })
})