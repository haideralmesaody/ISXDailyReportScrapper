/**
 * Unit tests for ChartTooltip component
 */

import React from 'react'
import { render, screen } from '@testing-library/react'
import { ChartTooltip } from '@/components/analysis/ChartTooltip'
import type { EnhancedTooltipData } from '@/lib/utils/tooltip-helpers'

describe('ChartTooltip Component', () => {
  const mockData: EnhancedTooltipData = {
    date: '2024-01-15',
    open: 100,
    high: 110,
    low: 95,
    close: 105,
    volume: 1000000,
    value: 105000000,
    trades: 500,
    change: 5,
    changePercent: 5,
    priceMA20: 102,
    volumeMA20: 950000,
    vsYesterday: {
      price: 100,
      change: 5,
      changePercent: 5,
      volume: 900000,
      date: '2024-01-14',
    },
    vsWeekAgo: {
      price: 95,
      change: 10,
      changePercent: 10.53,
      volume: 850000,
      date: '2024-01-08',
    },
    vsMonthAgo: {
      price: 90,
      change: 15,
      changePercent: 16.67,
      volume: 800000,
      date: '2023-12-15',
    },
  }

  const defaultProps = {
    data: mockData,
    position: { x: 100, y: 100 },
    visible: true,
    containerWidth: 800,
    containerHeight: 600,
  }

  it('should render when visible with data', () => {
    render(<ChartTooltip {...defaultProps} />)
    
    // Check date is displayed
    expect(screen.getByText(/Jan 15, 2024/)).toBeInTheDocument()
    
    // Check OHLCV labels are present
    expect(screen.getByText('Open:')).toBeInTheDocument()
    expect(screen.getByText('High:')).toBeInTheDocument()
    expect(screen.getByText('Low:')).toBeInTheDocument()
    expect(screen.getByText('Close:')).toBeInTheDocument()
    expect(screen.getByText('Volume:')).toBeInTheDocument()
  })

  it('should not render when not visible', () => {
    const { container } = render(<ChartTooltip {...defaultProps} visible={false} />)
    expect(container.firstChild).toBeNull()
  })

  it('should not render when data is null', () => {
    const { container } = render(<ChartTooltip {...defaultProps} data={null} />)
    expect(container.firstChild).toBeNull()
  })

  it('should display percentage change with correct color', () => {
    render(<ChartTooltip {...defaultProps} />)
    
    // Check for positive change indicator
    const changeElement = screen.getByText('+5.00%')
    expect(changeElement).toBeInTheDocument()
    expect(changeElement.parentElement).toHaveClass('text-green-600')
  })

  it('should display negative change with correct color', () => {
    const negativeData = {
      ...mockData,
      change: -5,
      changePercent: -5,
    }
    
    render(<ChartTooltip {...defaultProps} data={negativeData} />)
    
    const changeElement = screen.getByText('-5.00%')
    expect(changeElement).toBeInTheDocument()
    expect(changeElement.parentElement).toHaveClass('text-red-600')
  })

  it('should display moving averages when available', () => {
    render(<ChartTooltip {...defaultProps} />)
    
    expect(screen.getByText('Moving Averages (20-day)')).toBeInTheDocument()
    expect(screen.getByText(/Price MA:/)).toBeInTheDocument()
    expect(screen.getByText(/Volume MA:/)).toBeInTheDocument()
  })

  it('should not display moving averages section when not available', () => {
    const dataWithoutMA = {
      ...mockData,
      priceMA20: undefined,
      volumeMA20: undefined,
    }
    
    render(<ChartTooltip {...defaultProps} data={dataWithoutMA} />)
    
    expect(screen.queryByText('Moving Averages (20-day)')).not.toBeInTheDocument()
  })

  it('should display historical comparisons', () => {
    render(<ChartTooltip {...defaultProps} />)
    
    expect(screen.getByText('Historical Comparison')).toBeInTheDocument()
    expect(screen.getByText('vs Yesterday:')).toBeInTheDocument()
    expect(screen.getByText('vs Week Ago:')).toBeInTheDocument()
    expect(screen.getByText('vs Month Ago:')).toBeInTheDocument()
  })

  it('should handle missing historical comparison data', () => {
    const dataWithoutComparisons = {
      ...mockData,
      vsYesterday: null,
      vsWeekAgo: null,
      vsMonthAgo: null,
    }
    
    render(<ChartTooltip {...defaultProps} data={dataWithoutComparisons} />)
    
    expect(screen.getByText('Historical Comparison')).toBeInTheDocument()
    expect(screen.queryByText('vs Yesterday:')).not.toBeInTheDocument()
    expect(screen.queryByText('vs Week Ago:')).not.toBeInTheDocument()
    expect(screen.queryByText('vs Month Ago:')).not.toBeInTheDocument()
  })

  it('should position tooltip within container bounds', () => {
    const { container } = render(<ChartTooltip {...defaultProps} />)
    const tooltip = container.querySelector('.absolute')
    
    expect(tooltip).toBeInTheDocument()
    expect(tooltip).toHaveStyle({ left: '110px', top: '110px' })
  })

  it('should adjust position when near right edge', () => {
    const { container } = render(
      <ChartTooltip
        {...defaultProps}
        position={{ x: 700, y: 100 }}
        containerWidth={800}
      />
    )
    
    const tooltip = container.querySelector('.absolute')
    expect(tooltip).toBeInTheDocument()
    // Tooltip should be positioned to the left of cursor
    const leftValue = parseInt(tooltip?.style.left || '0')
    expect(leftValue).toBeLessThan(700)
  })

  it('should adjust position when near bottom edge', () => {
    const { container } = render(
      <ChartTooltip
        {...defaultProps}
        position={{ x: 100, y: 500 }}
        containerHeight={600}
      />
    )
    
    const tooltip = container.querySelector('.absolute')
    expect(tooltip).toBeInTheDocument()
    // Tooltip should be adjusted upward
    const topValue = parseInt(tooltip?.style.top || '0')
    expect(topValue).toBeLessThan(500)
  })

  it('should display trades count when available', () => {
    render(<ChartTooltip {...defaultProps} />)
    expect(screen.getByText(/Trades: 500/)).toBeInTheDocument()
  })

  it('should not display trades when not available', () => {
    const dataWithoutTrades = {
      ...mockData,
      trades: undefined,
    }
    
    render(<ChartTooltip {...defaultProps} data={dataWithoutTrades} />)
    expect(screen.queryByText(/Trades:/)).not.toBeInTheDocument()
  })

  it('should display value when available', () => {
    render(<ChartTooltip {...defaultProps} />)
    expect(screen.getByText('Value:')).toBeInTheDocument()
  })

  it('should not display value when not available', () => {
    const dataWithoutValue = {
      ...mockData,
      value: undefined,
    }
    
    render(<ChartTooltip {...defaultProps} data={dataWithoutValue} />)
    expect(screen.queryByText('Value:')).not.toBeInTheDocument()
  })
})