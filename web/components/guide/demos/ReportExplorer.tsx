'use client'

import { useState, useMemo } from 'react'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import {
  Search,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Download,
  TrendingUp,
  TrendingDown
} from 'lucide-react'

type SortDirection = 'asc' | 'desc' | null
type SortField = 'date' | 'code' | 'close' | 'volume' | 'change' | null

interface ReportRow {
  date: string
  code: string
  company: string
  open: number
  high: number
  low: number
  close: number
  volume: number
  value: number
  trades: number
  change: number
  changePct: number
}

/**
 * Report Explorer Demo
 * Interactive CSV report viewer with sorting, filtering, and search
 */
export function ReportExplorer() {
  const [searchTerm, setSearchTerm] = useState('')
  const [sortField, setSortField] = useState<SortField>(null)
  const [sortDirection, setSortDirection] = useState<SortDirection>(null)
  const [currentPage, setCurrentPage] = useState(1)
  const rowsPerPage = 10

  // Sample data (mock ISX report data)
  const sampleData: ReportRow[] = [
    { date: '2025-10-11', code: 'BAQL', company: 'Bank of Baghdad', open: 1.450, high: 1.480, low: 1.440, close: 1.470, volume: 2500000, value: 3687500, trades: 245, change: 0.020, changePct: 1.38 },
    { date: '2025-10-11', code: 'BBOB', company: 'Babylon Bank', open: 0.950, high: 0.965, low: 0.945, close: 0.960, volume: 1800000, value: 1728000, trades: 180, change: 0.010, changePct: 1.05 },
    { date: '2025-10-11', code: 'TASC', company: 'Al-Taif Islamic Bank', open: 1.200, high: 1.215, low: 1.195, close: 1.210, volume: 3200000, value: 3872000, trades: 320, change: 0.010, changePct: 0.83 },
    { date: '2025-10-11', code: 'BGUC', company: 'United Bank', open: 0.780, high: 0.795, low: 0.775, close: 0.780, volume: 950000, value: 741000, trades: 95, change: 0.000, changePct: 0.00 },
    { date: '2025-10-11', code: 'BAMS', company: 'Al-Mansour Bank', open: 1.100, high: 1.125, low: 1.095, close: 1.115, volume: 1600000, value: 1784000, trades: 160, change: 0.015, changePct: 1.36 },
    { date: '2025-10-10', code: 'BAQL', company: 'Bank of Baghdad', open: 1.430, high: 1.455, low: 1.425, close: 1.450, volume: 2300000, value: 3335000, trades: 230, change: 0.020, changePct: 1.40 },
    { date: '2025-10-10', code: 'BBOB', company: 'Babylon Bank', open: 0.940, high: 0.955, low: 0.935, close: 0.950, volume: 1700000, value: 1615000, trades: 170, change: 0.010, changePct: 1.06 },
    { date: '2025-10-10', code: 'TASC', company: 'Al-Taif Islamic Bank', open: 1.190, high: 1.205, low: 1.185, close: 1.200, volume: 3000000, value: 3600000, trades: 300, change: 0.010, changePct: 0.84 },
    { date: '2025-10-10', code: 'BGUC', company: 'United Bank', open: 0.770, high: 0.785, low: 0.765, close: 0.780, volume: 900000, value: 702000, trades: 90, change: 0.010, changePct: 1.30 },
    { date: '2025-10-10', code: 'BAMS', company: 'Al-Mansour Bank', open: 1.090, high: 1.105, low: 1.085, close: 1.100, volume: 1500000, value: 1650000, trades: 150, change: 0.010, changePct: 0.92 },
    { date: '2025-10-09', code: 'BAQL', company: 'Bank of Baghdad', open: 1.420, high: 1.435, low: 1.415, close: 1.430, volume: 2100000, value: 3003000, trades: 210, change: 0.010, changePct: 0.70 },
    { date: '2025-10-09', code: 'BBOB', company: 'Babylon Bank', open: 0.930, high: 0.945, low: 0.925, close: 0.940, volume: 1600000, value: 1504000, trades: 160, change: 0.010, changePct: 1.08 },
    { date: '2025-10-09', code: 'TASC', company: 'Al-Taif Islamic Bank', open: 1.180, high: 1.195, low: 1.175, close: 1.190, volume: 2800000, value: 3332000, trades: 280, change: 0.010, changePct: 0.85 },
    { date: '2025-10-09', code: 'BGUC', company: 'United Bank', open: 0.760, high: 0.775, low: 0.755, close: 0.770, volume: 850000, value: 654500, trades: 85, change: 0.010, changePct: 1.32 },
    { date: '2025-10-09', code: 'BAMS', company: 'Al-Mansour Bank', open: 1.080, high: 1.095, low: 1.075, close: 1.090, volume: 1400000, value: 1526000, trades: 140, change: 0.010, changePct: 0.93 }
  ]

  // Filter data based on search term
  const filteredData = useMemo(() => {
    if (!searchTerm) return sampleData

    const term = searchTerm.toLowerCase()
    return sampleData.filter(row =>
      row.code.toLowerCase().includes(term) ||
      row.company.toLowerCase().includes(term) ||
      row.date.includes(term)
    )
  }, [searchTerm])

  // Sort data
  const sortedData = useMemo(() => {
    if (!sortField || !sortDirection) return filteredData

    return [...filteredData].sort((a, b) => {
      let aVal: any = a[sortField === 'change' ? 'changePct' : sortField as keyof ReportRow]
      let bVal: any = b[sortField === 'change' ? 'changePct' : sortField as keyof ReportRow]

      if (sortField === 'date') {
        aVal = new Date(aVal).getTime()
        bVal = new Date(bVal).getTime()
      }

      if (sortDirection === 'asc') {
        return aVal > bVal ? 1 : -1
      } else {
        return aVal < bVal ? 1 : -1
      }
    })
  }, [filteredData, sortField, sortDirection])

  // Paginate data
  const paginatedData = useMemo(() => {
    const startIndex = (currentPage - 1) * rowsPerPage
    const endIndex = startIndex + rowsPerPage
    return sortedData.slice(startIndex, endIndex)
  }, [sortedData, currentPage])

  const totalPages = Math.ceil(sortedData.length / rowsPerPage)

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      // Cycle through: asc -> desc -> null
      if (sortDirection === 'asc') {
        setSortDirection('desc')
      } else if (sortDirection === 'desc') {
        setSortDirection(null)
        setSortField(null)
      }
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
  }

  const getSortIcon = (field: SortField) => {
    if (sortField !== field) {
      return <ArrowUpDown className="h-3 w-3 text-muted-foreground" />
    }
    if (sortDirection === 'asc') {
      return <ArrowUp className="h-3 w-3 text-primary" />
    }
    if (sortDirection === 'desc') {
      return <ArrowDown className="h-3 w-3 text-primary" />
    }
    return <ArrowUpDown className="h-3 w-3 text-muted-foreground" />
  }

  const handleDownload = () => {
    const csv = [
      ['Date', 'Code', 'Company', 'Open', 'High', 'Low', 'Close', 'Volume', 'Value', 'Trades', 'Change%'],
      ...sortedData.map(row => [
        row.date,
        row.code,
        row.company,
        row.open.toFixed(3),
        row.high.toFixed(3),
        row.low.toFixed(3),
        row.close.toFixed(3),
        row.volume.toString(),
        row.value.toString(),
        row.trades.toString(),
        row.changePct.toFixed(2)
      ])
    ].map(row => row.join(',')).join('\n')

    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'isx_report_sample.csv'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-4">
      {/* Controls */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by ticker, company, or date..."
            value={searchTerm}
            onChange={(e) => {
              setSearchTerm(e.target.value)
              setCurrentPage(1)
            }}
            className="pl-10"
          />
        </div>
        <Button onClick={handleDownload} variant="outline" className="gap-2 shrink-0">
          <Download className="h-4 w-4" />
          Download CSV
        </Button>
      </div>

      {/* Stats */}
      <div className="flex items-center gap-4 text-sm text-muted-foreground">
        <span>Showing {paginatedData.length} of {sortedData.length} rows</span>
        {sortField && (
          <Badge variant="secondary">
            Sorted by {sortField === 'change' ? 'change %' : sortField} ({sortDirection})
          </Badge>
        )}
      </div>

      {/* Table */}
      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>
                  <button
                    onClick={() => handleSort('date')}
                    className="flex items-center gap-1 hover:text-primary transition-colors"
                  >
                    Date {getSortIcon('date')}
                  </button>
                </TableHead>
                <TableHead>
                  <button
                    onClick={() => handleSort('code')}
                    className="flex items-center gap-1 hover:text-primary transition-colors"
                  >
                    Code {getSortIcon('code')}
                  </button>
                </TableHead>
                <TableHead>Company</TableHead>
                <TableHead className="text-right">Open</TableHead>
                <TableHead className="text-right">High</TableHead>
                <TableHead className="text-right">Low</TableHead>
                <TableHead className="text-right">
                  <button
                    onClick={() => handleSort('close')}
                    className="flex items-center gap-1 hover:text-primary transition-colors ml-auto"
                  >
                    Close {getSortIcon('close')}
                  </button>
                </TableHead>
                <TableHead className="text-right">
                  <button
                    onClick={() => handleSort('volume')}
                    className="flex items-center gap-1 hover:text-primary transition-colors ml-auto"
                  >
                    Volume {getSortIcon('volume')}
                  </button>
                </TableHead>
                <TableHead className="text-right">
                  <button
                    onClick={() => handleSort('change')}
                    className="flex items-center gap-1 hover:text-primary transition-colors ml-auto"
                  >
                    Change% {getSortIcon('change')}
                  </button>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {paginatedData.map((row, i) => (
                <TableRow key={i}>
                  <TableCell className="font-mono text-xs">{row.date}</TableCell>
                  <TableCell>
                    <Badge variant="outline">{row.code}</Badge>
                  </TableCell>
                  <TableCell className="max-w-[200px] truncate">{row.company}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{row.open.toFixed(3)}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{row.high.toFixed(3)}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{row.low.toFixed(3)}</TableCell>
                  <TableCell className="text-right font-mono text-sm font-semibold">{row.close.toFixed(3)}</TableCell>
                  <TableCell className="text-right font-mono text-sm">{row.volume.toLocaleString()}</TableCell>
                  <TableCell className="text-right">
                    <div className={`inline-flex items-center gap-1 ${row.changePct > 0 ? 'text-green-600 dark:text-green-400' : row.changePct < 0 ? 'text-red-600 dark:text-red-400' : 'text-muted-foreground'}`}>
                      {row.changePct > 0 && <TrendingUp className="h-3 w-3" />}
                      {row.changePct < 0 && <TrendingDown className="h-3 w-3" />}
                      <span className="font-mono text-sm">{row.changePct > 0 ? '+' : ''}{row.changePct.toFixed(2)}%</span>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
            disabled={currentPage === 1}
          >
            Previous
          </Button>
          <div className="text-sm text-muted-foreground">
            Page {currentPage} of {totalPages}
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
            disabled={currentPage === totalPages}
          >
            Next
          </Button>
        </div>
      )}

      {/* Instructions */}
      <Card className="bg-muted/50">
        <div className="p-4">
          <h4 className="font-semibold mb-2 text-sm">Features:</h4>
          <ul className="space-y-1 text-xs text-muted-foreground">
            <li>• <strong>Search:</strong> Filter by ticker code, company name, or date</li>
            <li>• <strong>Sort:</strong> Click column headers to sort (asc → desc → none)</li>
            <li>• <strong>Pagination:</strong> Navigate through multiple pages</li>
            <li>• <strong>Download:</strong> Export filtered results as CSV</li>
            <li>• <strong>Real Data:</strong> Sample ISX data for demonstration</li>
          </ul>
        </div>
      </Card>
    </div>
  )
}
