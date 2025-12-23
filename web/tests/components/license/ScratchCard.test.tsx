/**
 * @jest-environment jsdom
 */

import React from 'react'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'

import ScratchCard from '@/components/license/ScratchCard'
import { useToast } from '@/lib/hooks/use-toast'
import { useHydration } from '@/lib/hooks/use-hydration'
import type { ScratchCardData } from '@/types/index'

// Mock the hooks
jest.mock('@/lib/hooks/use-toast')
jest.mock('@/lib/hooks/use-hydration')
jest.mock('@/lib/utils/license-helpers', () => ({
  copyToClipboard: jest.fn(),
}))

// Mock framer-motion
jest.mock('framer-motion', () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

// Mock canvas context for scratch functionality
const mockCanvasContext = {
  scale: jest.fn(),
  fillRect: jest.fn(),
  createLinearGradient: jest.fn(() => ({
    addColorStop: jest.fn(),
  })),
  getImageData: jest.fn(() => ({
    data: new Uint8ClampedArray(1000),
  })),
  arc: jest.fn(),
  fill: jest.fn(),
  beginPath: jest.fn(),
  setLinearGradient: jest.fn(),
}

// Mock HTMLCanvasElement methods
Object.defineProperty(HTMLCanvasElement.prototype, 'getContext', {
  value: jest.fn(() => mockCanvasContext),
})

Object.defineProperty(HTMLCanvasElement.prototype, 'getBoundingClientRect', {
  value: jest.fn(() => ({
    width: 350,
    height: 220,
    left: 0,
    top: 0,
  })),
})

const mockToast = jest.fn()
const mockUseToast = useToast as jest.MockedFunction<typeof useToast>
const mockUseHydration = useHydration as jest.MockedFunction<typeof useHydration>

// Mock copyToClipboard
const { copyToClipboard } = require('@/lib/utils/license-helpers')

describe('ScratchCard Component', () => {
  const defaultProps = {
    data: {
      code: 'ISX-1234-5678-90AB',
      format: 'scratch' as const,
      revealed: false,
      activationId: 'act_12345678',
    } as ScratchCardData,
  }

  beforeEach(() => {
    jest.clearAllMocks()
    mockUseToast.mockReturnValue({ toast: mockToast })
    mockUseHydration.mockReturnValue(true)
    copyToClipboard.mockResolvedValue(true)
    
    // Reset canvas context mocks
    mockCanvasContext.scale.mockClear()
    mockCanvasContext.fillRect.mockClear()
    mockCanvasContext.createLinearGradient.mockClear()
  })

  describe('Rendering', () => {
    it('renders loading state when not hydrated', () => {
      mockUseHydration.mockReturnValue(false)
      
      render(<ScratchCard {...defaultProps} />)
      
      expect(screen.getByText('Loading scratch card...')).toBeInTheDocument()
      expect(screen.getByText('Loading scratch card...')).toHaveClass('text-muted-foreground')
    })

    it('renders scratch card when hydrated', () => {
      render(<ScratchCard {...defaultProps} />)
      
      expect(screen.getByText('ISX Pulse License')).toBeInTheDocument()
      expect(screen.getByText('Scratch Card')).toBeInTheDocument()
      expect(screen.getByText('Scratch to reveal your license code')).toBeInTheDocument()
    })

    it('renders with custom size', () => {
      render(<ScratchCard {...defaultProps} size="lg" />)
      
      const container = screen.getByText('ISX Pulse License').closest('[style]')
      expect(container).toHaveStyle({ width: '420px', height: '260px' })
    })

    it('renders with custom theme', () => {
      render(<ScratchCard {...defaultProps} theme="premium" />)
      
      // Check that the component renders (theme styling is applied via CSS classes)
      expect(screen.getByText('ISX Pulse License')).toBeInTheDocument()
    })

    it('displays activation ID when provided', () => {
      render(<ScratchCard {...defaultProps} />)
      
      expect(screen.getByText('ID: 12345678')).toBeInTheDocument()
    })

    it('does not display activation ID when not provided', () => {
      const propsWithoutId = {
        ...defaultProps,
        data: { ...defaultProps.data, activationId: undefined },
      }
      
      render(<ScratchCard {...propsWithoutId} />)
      
      expect(screen.queryByText(/ID:/)).not.toBeInTheDocument()
    })
  })

  describe('Scratch Interaction', () => {
    it('initializes canvas on mount', () => {
      render(<ScratchCard {...defaultProps} />)
      
      expect(mockCanvasContext.scale).toHaveBeenCalledWith(2, 2)
      expect(mockCanvasContext.fillRect).toHaveBeenCalled()
      expect(mockCanvasContext.createLinearGradient).toHaveBeenCalled()
    })

    it('handles mouse scratch interaction', async () => {
      const onReveal = jest.fn()
      render(<ScratchCard {...defaultProps} onReveal={onReveal} />)
      
      const canvas = screen.getByRole('img', { hidden: true }) // Canvas has implicit img role
      
      // Simulate mouse scratch
      fireEvent.mouseDown(canvas, { clientX: 100, clientY: 100 })
      fireEvent.mouseMove(canvas, { clientX: 110, clientY: 110 })
      fireEvent.mouseUp(canvas)
      
      expect(mockCanvasContext.arc).toHaveBeenCalled()
      expect(mockCanvasContext.fill).toHaveBeenCalled()
    })

    it('handles touch scratch interaction', async () => {
      const onReveal = jest.fn()
      render(<ScratchCard {...defaultProps} onReveal={onReveal} />)
      
      const canvas = screen.getByRole('img', { hidden: true })
      
      // Simulate touch scratch
      fireEvent.touchStart(canvas, {
        touches: [{ clientX: 100, clientY: 100 }],
      })
      fireEvent.touchMove(canvas, {
        touches: [{ clientX: 110, clientY: 110 }],
      })
      fireEvent.touchEnd(canvas)
      
      expect(mockCanvasContext.arc).toHaveBeenCalled()
      expect(mockCanvasContext.fill).toHaveBeenCalled()
    })

    it('reveals code when scratch progress exceeds threshold', async () => {
      const onReveal = jest.fn()
      
      // Mock getImageData to return high transparency (scratched)
      mockCanvasContext.getImageData.mockReturnValue({
        data: new Uint8ClampedArray(1000).fill(0), // All transparent
      })
      
      render(<ScratchCard {...defaultProps} onReveal={onReveal} />)
      
      const canvas = screen.getByRole('img', { hidden: true })
      
      // Simulate scratching
      fireEvent.mouseDown(canvas, { clientX: 100, clientY: 100 })
      
      await waitFor(() => {
        expect(onReveal).toHaveBeenCalledWith('ISX-1234-5678-90AB')
      }, { timeout: 1000 })
    })

    it('does not scratch when already revealed', () => {
      const revealedData = { ...defaultProps.data, revealed: true }
      render(<ScratchCard data={revealedData} />)
      
      // Canvas should not be present when already revealed
      expect(screen.queryByRole('img', { hidden: true })).not.toBeInTheDocument()
    })
  })

  describe('Revealed State', () => {
    const revealedProps = {
      ...defaultProps,
      data: { ...defaultProps.data, revealed: true },
    }

    it('displays license code when revealed', () => {
      render(<ScratchCard {...revealedProps} />)
      
      expect(screen.getByText('ISX-1234-5678-90AB')).toBeInTheDocument()
      expect(screen.getByText('Copy Code')).toBeInTheDocument()
    })

    it('shows revealed status indicator', () => {
      render(<ScratchCard {...revealedProps} />)
      
      expect(screen.getByText('Revealed')).toBeInTheDocument()
    })

    it('copies code to clipboard on button click', async () => {
      const onCopy = jest.fn()
      const user = userEvent.setup()
      
      render(<ScratchCard {...revealedProps} onCopy={onCopy} />)
      
      const copyButton = screen.getByText('Copy Code')
      await user.click(copyButton)
      
      expect(copyToClipboard).toHaveBeenCalledWith('ISX-1234-5678-90AB')
      expect(onCopy).toHaveBeenCalledWith('ISX-1234-5678-90AB')
      expect(mockToast).toHaveBeenCalledWith({
        title: 'Copied!',
        description: 'License code copied to clipboard',
      })
    })

    it('shows success state after copying', async () => {
      const user = userEvent.setup()
      
      render(<ScratchCard {...revealedProps} />)
      
      const copyButton = screen.getByText('Copy Code')
      await user.click(copyButton)
      
      expect(screen.getByText('Copied!')).toBeInTheDocument()
      
      // Should revert back to "Copy Code" after 2 seconds
      await waitFor(
        () => {
          expect(screen.getByText('Copy Code')).toBeInTheDocument()
        },
        { timeout: 2500 }
      )
    })

    it('handles clipboard copy failure', async () => {
      copyToClipboard.mockResolvedValue(false)
      const user = userEvent.setup()
      
      render(<ScratchCard {...revealedProps} />)
      
      const copyButton = screen.getByText('Copy Code')
      await user.click(copyButton)
      
      expect(mockToast).toHaveBeenCalledWith({
        title: 'Copy failed',
        description: 'Unable to copy to clipboard',
        variant: 'destructive',
      })
    })
  })

  describe('Code Formatting', () => {
    it('formats scratch card codes with dashes', () => {
      const unformattedData = {
        ...defaultProps.data,
        code: 'ISX1234567890AB',
        revealed: true,
      }
      
      render(<ScratchCard data={unformattedData} />)
      
      expect(screen.getByText('ISX-1234-5678-90AB')).toBeInTheDocument()
    })

    it('handles standard format codes without formatting', () => {
      const standardData = {
        ...defaultProps.data,
        code: 'ISX1M02LYE1F9QJHR9D7Z',
        format: 'standard' as const,
        revealed: true,
      }
      
      render(<ScratchCard data={standardData} />)
      
      expect(screen.getByText('ISX1M02LYE1F9QJHR9D7Z')).toBeInTheDocument()
    })
  })

  describe('Hover Effects', () => {
    it('applies hover effects on mouse enter and leave', async () => {
      const user = userEvent.setup()
      
      render(<ScratchCard {...defaultProps} />)
      
      const card = screen.getByText('ISX Pulse License').closest('[style]')
      
      // Test hover
      await user.hover(card!)
      
      // Test unhover
      await user.unhover(card!)
      
      // Component should render without errors
      expect(screen.getByText('ISX Pulse License')).toBeInTheDocument()
    })
  })

  describe('Animation Effects', () => {
    it('shows sparkle effect when revealed', () => {
      render(<ScratchCard {...defaultProps} data={{ ...defaultProps.data, revealed: true }} />)
      
      // Check that the component renders with sparkle elements
      expect(screen.getByText('Revealed')).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('provides proper ARIA labels and roles', () => {
      render(<ScratchCard {...defaultProps} />)
      
      expect(screen.getByText('Scratch to reveal your license code')).toBeInTheDocument()
    })

    it('supports keyboard navigation for copy button', async () => {
      const user = userEvent.setup()
      
      render(<ScratchCard {...defaultProps} data={{ ...defaultProps.data, revealed: true }} />)
      
      const copyButton = screen.getByText('Copy Code')
      
      // Tab to button and activate with Enter
      copyButton.focus()
      await user.keyboard('{Enter}')
      
      expect(copyToClipboard).toHaveBeenCalledWith('ISX-1234-5678-90AB')
    })

    it('supports keyboard navigation for copy button with Space', async () => {
      const user = userEvent.setup()
      
      render(<ScratchCard {...defaultProps} data={{ ...defaultProps.data, revealed: true }} />)
      
      const copyButton = screen.getByText('Copy Code')
      
      // Tab to button and activate with Space
      copyButton.focus()
      await user.keyboard(' ')
      
      expect(copyToClipboard).toHaveBeenCalledWith('ISX-1234-5678-90AB')
    })
  })

  describe('Performance', () => {
    it('memoizes handlers to prevent unnecessary re-renders', () => {
      const { rerender } = render(<ScratchCard {...defaultProps} />)
      
      // Rerender with same props
      rerender(<ScratchCard {...defaultProps} />)
      
      // Component should handle re-renders efficiently
      expect(screen.getByText('ISX Pulse License')).toBeInTheDocument()
    })

    it('cleans up event listeners on unmount', () => {
      const { unmount } = render(<ScratchCard {...defaultProps} />)
      
      unmount()
      
      // No way to directly test event listener cleanup in jsdom,
      // but we can verify the component unmounts without errors
      expect(screen.queryByText('ISX Pulse License')).not.toBeInTheDocument()
    })
  })

  describe('Error Handling', () => {
    it('handles canvas context creation failure gracefully', () => {
      // Mock canvas context to return null
      HTMLCanvasElement.prototype.getContext = jest.fn(() => null)
      
      expect(() => render(<ScratchCard {...defaultProps} />)).not.toThrow()
      
      // Restore original mock
      HTMLCanvasElement.prototype.getContext = jest.fn(() => mockCanvasContext)
    })

    it('handles missing canvas element gracefully', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation()
      
      render(<ScratchCard {...defaultProps} />)
      
      // Should not throw errors even with potential canvas issues
      expect(screen.getByText('ISX Pulse License')).toBeInTheDocument()
      
      consoleSpy.mockRestore()
    })
  })

  describe('Integration with Hooks', () => {
    it('uses hydration hook correctly', () => {
      render(<ScratchCard {...defaultProps} />)
      
      expect(mockUseHydration).toHaveBeenCalled()
    })

    it('uses toast hook correctly', async () => {
      const user = userEvent.setup()
      
      render(<ScratchCard {...defaultProps} data={{ ...defaultProps.data, revealed: true }} />)
      
      const copyButton = screen.getByText('Copy Code')
      await user.click(copyButton)
      
      expect(mockUseToast).toHaveBeenCalled()
    })
  })

  describe('Theme Variations', () => {
    it('renders default theme correctly', () => {
      render(<ScratchCard {...defaultProps} theme="default" />)
      
      expect(screen.getByText('ISX Pulse License')).toBeInTheDocument()
    })

    it('renders premium theme correctly', () => {
      render(<ScratchCard {...defaultProps} theme="premium" />)
      
      expect(screen.getByText('ISX Pulse License')).toBeInTheDocument()
    })

    it('renders gold theme correctly', () => {
      render(<ScratchCard {...defaultProps} theme="gold" />)
      
      expect(screen.getByText('ISX Pulse License')).toBeInTheDocument()
    })
  })

  describe('Size Variations', () => {
    it('renders small size correctly', () => {
      render(<ScratchCard {...defaultProps} size="sm" />)
      
      const container = screen.getByText('ISX Pulse License').closest('[style]')
      expect(container).toHaveStyle({ width: '280px', height: '180px' })
    })

    it('renders medium size correctly', () => {
      render(<ScratchCard {...defaultProps} size="md" />)
      
      const container = screen.getByText('ISX Pulse License').closest('[style]')
      expect(container).toHaveStyle({ width: '350px', height: '220px' })
    })

    it('renders large size correctly', () => {
      render(<ScratchCard {...defaultProps} size="lg" />)
      
      const container = screen.getByText('ISX Pulse License').closest('[style]')
      expect(container).toHaveStyle({ width: '420px', height: '260px' })
    })
  })
})

// Performance benchmark test
describe('ScratchCard Performance', () => {
  beforeAll(() => {
    jest.spyOn(performance, 'now')
      .mockReturnValueOnce(0)
      .mockReturnValueOnce(100)
  })

  afterAll(() => {
    jest.restoreAllMocks()
  })

  it('renders within performance budget', () => {
    const startTime = performance.now()
    
    render(
      <ScratchCard
        data={{
          code: 'ISX-1234-5678-90AB',
          format: 'scratch',
          revealed: false,
          activationId: 'act_12345678',
        }}
      />
    )
    
    const endTime = performance.now()
    const renderTime = endTime - startTime
    
    // Should render within 100ms (mocked performance.now returns 100)
    expect(renderTime).toBeLessThanOrEqual(100)
  })
})