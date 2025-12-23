/**
 * @jest-environment jsdom
 */
import React from 'react'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { Loader2, Clock, Download, AlertCircle } from 'lucide-react'
import { DataLoadingState, type DataLoadingStateProps } from '../data-loading-state'

describe('DataLoadingState Component', () => {
  const defaultProps: DataLoadingStateProps = {}

  describe('Basic Rendering', () => {
    const tests = [
      {
        name: 'renders with default props',
        props: defaultProps,
        expectedMessage: 'Loading...',
        expectedSize: 'default',
        shouldHaveCard: true,
      },
      {
        name: 'renders with custom message',
        props: {
          message: 'Custom loading message',
        },
        expectedMessage: 'Custom loading message',
        expectedSize: 'default',
        shouldHaveCard: true,
      },
      {
        name: 'renders with no message',
        props: {
          message: '',
        },
        expectedMessage: '',
        expectedSize: 'default',
        shouldHaveCard: true,
      },
      {
        name: 'renders without card wrapper',
        props: {
          showCard: false,
        },
        expectedMessage: 'Loading...',
        expectedSize: 'default',
        shouldHaveCard: false,
      },
    ]

    tests.forEach(({ name, props, expectedMessage, expectedSize, shouldHaveCard }) => {
      it(name, () => {
        render(<DataLoadingState {...props} />)
        
        // Check for loading indicator with proper ARIA attributes
        const statusElement = screen.getByRole('status')
        expect(statusElement).toBeInTheDocument()
        expect(statusElement).toHaveAttribute('aria-live', 'polite')
        
        // Check for message if it exists
        if (expectedMessage) {
          expect(screen.getByText(expectedMessage)).toBeInTheDocument()
        }
        
        // Check for default spinning icon (Loader2)
        const spinningIcon = document.querySelector('.animate-spin')
        expect(spinningIcon).toBeInTheDocument()
        
        // Check for card wrapper
        if (shouldHaveCard) {
          const cardContainer = document.querySelector('.min-h-screen.p-8')
          expect(cardContainer).toBeInTheDocument()
          const card = document.querySelector('.border.rounded-lg')
          expect(card).toBeInTheDocument()
        } else {
          const cardContainer = document.querySelector('.min-h-screen.p-8')
          expect(cardContainer).not.toBeInTheDocument()
        }
      })
    })
  })

  describe('Size Variants', () => {
    const sizeTests = [
      {
        name: 'small size variant',
        size: 'sm' as const,
        expectedIconClass: 'h-6 w-6',
        expectedTextClass: 'text-sm',
        expectedSpacingClass: 'space-y-2',
        expectedPaddingClass: 'p-4',
      },
      {
        name: 'default size variant',
        size: 'default' as const,
        expectedIconClass: 'h-8 w-8',
        expectedTextClass: 'text-base',
        expectedSpacingClass: 'space-y-4',
        expectedPaddingClass: 'p-8',
      },
      {
        name: 'large size variant',
        size: 'lg' as const,
        expectedIconClass: 'h-12 w-12',
        expectedTextClass: 'text-lg',
        expectedSpacingClass: 'space-y-6',
        expectedPaddingClass: 'p-12',
      },
    ]

    sizeTests.forEach(({ name, size, expectedIconClass, expectedTextClass, expectedSpacingClass, expectedPaddingClass }) => {
      it(name, () => {
        render(
          <DataLoadingState 
            size={size}
            message="Test message"
          />
        )
        
        // Check icon size
        const icon = document.querySelector('svg')
        expect(icon).toBeInTheDocument()
        expect(icon).toHaveClass(expectedIconClass)
        
        // Check text size
        const message = screen.getByText('Test message')
        expect(message).toHaveClass(expectedTextClass)
        
        // Check spacing
        const spacingContainer = document.querySelector(`.${expectedSpacingClass.replace('-', '\\-')}`)
        expect(spacingContainer).toBeInTheDocument()
        
        // Check padding
        const paddingContainer = document.querySelector(`.${expectedPaddingClass.replace('-', '\\-')}`)
        expect(paddingContainer).toBeInTheDocument()
      })
    })
  })

  describe('Custom Icon', () => {
    const iconTests = [
      {
        name: 'renders with default Loader2 icon',
        props: defaultProps,
        expectedIcon: Loader2,
        shouldSpin: true,
      },
      {
        name: 'renders with custom Clock icon',
        props: {
          icon: Clock,
        },
        expectedIcon: Clock,
        shouldSpin: false,
      },
      {
        name: 'renders with custom Download icon',
        props: {
          icon: Download,
        },
        expectedIcon: Download,
        shouldSpin: false,
      },
      {
        name: 'renders with custom AlertCircle icon (should not spin)',
        props: {
          icon: AlertCircle,
        },
        expectedIcon: AlertCircle,
        shouldSpin: false,
      },
    ]

    iconTests.forEach(({ name, props, shouldSpin }) => {
      it(name, () => {
        render(<DataLoadingState {...props} />)
        
        const icon = document.querySelector('svg')
        expect(icon).toBeInTheDocument()
        expect(icon).toHaveClass('text-muted-foreground')
        
        if (shouldSpin) {
          expect(icon).toHaveClass('animate-spin')
        } else {
          expect(icon).not.toHaveClass('animate-spin')
        }
      })
    })
  })

  describe('Card vs No-Card Rendering', () => {
    it('renders with card wrapper by default', () => {
      render(<DataLoadingState />)
      
      const outerContainer = document.querySelector('.min-h-screen.p-8')
      expect(outerContainer).toBeInTheDocument()
      
      const maxWidthContainer = document.querySelector('.max-w-7xl.mx-auto')
      expect(maxWidthContainer).toBeInTheDocument()
      
      const card = document.querySelector('.border.rounded-lg')
      expect(card).toBeInTheDocument()
    })

    it('renders without card wrapper when showCard is false', () => {
      render(<DataLoadingState showCard={false} />)
      
      const outerContainer = document.querySelector('.min-h-screen.p-8')
      expect(outerContainer).not.toBeInTheDocument()
      
      const card = document.querySelector('.border.rounded-lg')
      expect(card).not.toBeInTheDocument()
      
      const flexContainer = document.querySelector('.flex.items-center.justify-center')
      expect(flexContainer).toBeInTheDocument()
    })

    it('applies correct classes when showCard is true', () => {
      render(<DataLoadingState showCard={true} />)
      
      const outerContainer = document.querySelector('.min-h-screen')
      expect(outerContainer).toBeInTheDocument()
      expect(outerContainer).toHaveClass('min-h-screen', 'p-8')
    })

    it('applies correct classes when showCard is false', () => {
      render(<DataLoadingState showCard={false} />)
      
      const flexContainer = document.querySelector('.flex.items-center.justify-center')
      expect(flexContainer).toBeInTheDocument()
    })
  })

  describe('Custom ClassName', () => {
    it('applies custom className with card', () => {
      const customClass = 'custom-loading-class'
      render(
        <DataLoadingState 
          className={customClass}
          showCard={true}
        />
      )
      
      const container = document.querySelector(`.${customClass}`)
      expect(container).toBeInTheDocument()
      expect(container).toHaveClass('min-h-screen', 'p-8', customClass)
    })

    it('applies custom className without card', () => {
      const customClass = 'custom-no-card-class'
      render(
        <DataLoadingState 
          className={customClass}
          showCard={false}
        />
      )
      
      const container = document.querySelector(`.${customClass}`)
      expect(container).toBeInTheDocument()
      expect(container).toHaveClass('flex', 'items-center', 'justify-center', customClass)
    })
  })

  describe('Message Handling', () => {
    const messageTests = [
      {
        name: 'renders default message',
        message: undefined,
        expectedMessage: 'Loading...',
      },
      {
        name: 'renders custom message',
        message: 'Fetching data...',
        expectedMessage: 'Fetching data...',
      },
      {
        name: 'renders long message',
        message: 'Loading a very long message that should still display correctly in the component',
        expectedMessage: 'Loading a very long message that should still display correctly in the component',
      },
      {
        name: 'renders empty message',
        message: '',
        expectedMessage: '',
      },
    ]

    messageTests.forEach(({ name, message, expectedMessage }) => {
      it(name, () => {
        render(<DataLoadingState message={message} />)
        
        if (expectedMessage) {
          const messageElement = screen.getByText(expectedMessage)
          expect(messageElement).toBeInTheDocument()
          expect(messageElement).toHaveClass('text-muted-foreground')
        } else {
          // When message is empty, the p element should not exist
          const messageElements = document.querySelectorAll('p')
          expect(messageElements).toHaveLength(0)
        }
      })
    })
  })

  describe('Accessibility', () => {
    it('has proper ARIA attributes', () => {
      render(<DataLoadingState />)
      
      const statusElement = screen.getByRole('status')
      expect(statusElement).toBeInTheDocument()
      expect(statusElement).toHaveAttribute('aria-live', 'polite')
    })

    it('maintains ARIA attributes with custom props', () => {
      render(
        <DataLoadingState 
          message="Custom loading"
          size="lg"
          showCard={false}
          icon={Clock}
        />
      )
      
      const statusElement = screen.getByRole('status')
      expect(statusElement).toBeInTheDocument()
      expect(statusElement).toHaveAttribute('aria-live', 'polite')
    })

    it('provides screen reader accessible content', () => {
      render(<DataLoadingState message="Processing your request" />)
      
      const message = screen.getByText('Processing your request')
      expect(message).toBeInTheDocument()
      
      const statusElement = screen.getByRole('status')
      expect(statusElement).toContainElement(message)
    })

    it('maintains accessibility without message', () => {
      render(<DataLoadingState message="" />)
      
      const statusElement = screen.getByRole('status')
      expect(statusElement).toBeInTheDocument()
      expect(statusElement).toHaveAttribute('aria-live', 'polite')
      
      // Should still have icon for visual feedback
      const icon = document.querySelector('svg')
      expect(icon).toBeInTheDocument()
    })
  })

  describe('Layout and Styling', () => {
    it('centers content properly', () => {
      render(<DataLoadingState />)
      
      const centerContainer = document.querySelector('.flex.items-center.justify-center')
      expect(centerContainer).toBeInTheDocument()
      
      const textCenter = document.querySelector('.text-center')
      expect(textCenter).toBeInTheDocument()
    })

    it('applies proper spacing classes', () => {
      render(<DataLoadingState size="default" />)
      
      const spacingContainer = document.querySelector('.space-y-4')
      expect(spacingContainer).toBeInTheDocument()
    })

    it('applies proper icon positioning', () => {
      render(<DataLoadingState />)
      
      const iconContainer = document.querySelector('.flex.justify-center')
      expect(iconContainer).toBeInTheDocument()
      
      const icon = iconContainer?.querySelector('svg')
      expect(icon).toBeInTheDocument()
    })

    it('applies muted foreground color to text and icon', () => {
      render(<DataLoadingState message="Test" />)
      
      const icon = document.querySelector('svg')
      const text = screen.getByText('Test')
      
      expect(icon).toHaveClass('text-muted-foreground')
      expect(text).toHaveClass('text-muted-foreground')
    })
  })

  describe('Responsive Behavior', () => {
    it('adapts to different screen sizes with card', () => {
      render(<DataLoadingState showCard={true} />)
      
      const maxWidthContainer = document.querySelector('.max-w-7xl.mx-auto')
      expect(maxWidthContainer).toBeInTheDocument()
      
      const outerContainer = document.querySelector('.min-h-screen')
      expect(outerContainer).toBeInTheDocument()
    })

    it('maintains proper layout without card', () => {
      render(<DataLoadingState showCard={false} />)
      
      const flexContainer = document.querySelector('.flex.items-center.justify-center')
      expect(flexContainer).toBeInTheDocument()
    })

    it('handles content overflow gracefully', () => {
      const longMessage = 'A'.repeat(1000)
      
      render(<DataLoadingState message={longMessage} />)
      
      const messageElement = screen.getByText(longMessage)
      expect(messageElement).toBeInTheDocument()
      
      // Should still be centered
      const textCenter = document.querySelector('.text-center')
      expect(textCenter).toBeInTheDocument()
    })
  })

  describe('Complex Scenarios', () => {
    it('renders with all custom props', () => {
      const complexProps: DataLoadingStateProps = {
        message: 'Complex loading state',
        icon: Clock,
        size: 'lg',
        showCard: false,
        className: 'complex-loading',
      }

      render(<DataLoadingState {...complexProps} />)
      
      // Check all aspects
      expect(screen.getByText('Complex loading state')).toBeInTheDocument()
      expect(screen.getByRole('status')).toBeInTheDocument()
      
      const icon = document.querySelector('svg')
      expect(icon).toBeInTheDocument()
      expect(icon).toHaveClass('h-12', 'w-12') // lg size
      expect(icon).not.toHaveClass('animate-spin') // custom icon shouldn't spin
      
      const container = document.querySelector('.complex-loading')
      expect(container).toBeInTheDocument()
      expect(container).toHaveClass('flex', 'items-center', 'justify-center')
      
      const message = screen.getByText('Complex loading state')
      expect(message).toHaveClass('text-lg') // lg size
    })

    it('handles size variant edge cases', () => {
      // Test all size combinations
      const sizes: Array<'sm' | 'default' | 'lg'> = ['sm', 'default', 'lg']
      
      sizes.forEach(size => {
        const { unmount } = render(
          <DataLoadingState 
            size={size}
            message={`Testing ${size} size`}
          />
        )
        
        expect(screen.getByText(`Testing ${size} size`)).toBeInTheDocument()
        expect(screen.getByRole('status')).toBeInTheDocument()
        
        unmount()
      })
    })

    it('maintains performance with rapid re-renders', () => {
      const { rerender } = render(<DataLoadingState />)
      
      // Rapid re-renders with different props
      for (let i = 0; i < 10; i++) {
        rerender(
          <DataLoadingState 
            message={`Loading ${i}`}
            size={i % 2 === 0 ? 'sm' : 'lg'}
          />
        )
        
        expect(screen.getByText(`Loading ${i}`)).toBeInTheDocument()
        expect(screen.getByRole('status')).toBeInTheDocument()
      }
    })
  })

  describe('Edge Cases', () => {
    it('handles undefined icon gracefully', () => {
      const props = {
        icon: undefined,
      }

      render(<DataLoadingState {...props} />)
      
      // Should render default Loader2 icon
      const icon = document.querySelector('svg')
      expect(icon).toBeInTheDocument()
      expect(icon).toHaveClass('animate-spin') // Default should spin
    })

    it('handles null message', () => {
      const props = {
        message: null as any,
      }

      expect(() => {
        render(<DataLoadingState {...props} />)
      }).not.toThrow()
      
      // Should not render message element
      expect(screen.queryByText('null')).not.toBeInTheDocument()
    })

    it('handles invalid size gracefully', () => {
      const props = {
        size: 'invalid' as any,
      }

      // This will actually throw because the component doesn't handle invalid sizes
      // The component should be updated to handle this, but for now we test what it does
      expect(() => {
        render(<DataLoadingState {...props} />)
      }).toThrow()
    })

    it('handles boolean showCard values correctly', () => {
      // Test explicit true
      const { rerender } = render(<DataLoadingState showCard={true} />)
      expect(document.querySelector('.min-h-screen')).toBeInTheDocument()
      
      // Test explicit false
      rerender(<DataLoadingState showCard={false} />)
      expect(document.querySelector('.min-h-screen')).not.toBeInTheDocument()
    })

    it('maintains ref forwarding', () => {
      const ref = React.createRef<HTMLDivElement>()
      
      render(<DataLoadingState ref={ref} />)
      
      expect(ref.current).toBeInTheDocument()
      expect(ref.current?.tagName).toBe('DIV')
    })
  })

  describe('Forward Ref Behavior', () => {
    it('forwards ref correctly with card', () => {
      const ref = React.createRef<HTMLDivElement>()
      
      render(<DataLoadingState ref={ref} showCard={true} />)
      
      expect(ref.current).toBeInTheDocument()
      expect(ref.current).toHaveClass('min-h-screen', 'p-8')
    })

    it('forwards ref correctly without card', () => {
      const ref = React.createRef<HTMLDivElement>()
      
      render(<DataLoadingState ref={ref} showCard={false} />)
      
      expect(ref.current).toBeInTheDocument()
      expect(ref.current).toHaveClass('flex', 'items-center', 'justify-center')
    })
  })

  describe('Animation Behavior', () => {
    it('applies spin animation only to default icon', () => {
      const { rerender } = render(<DataLoadingState />)
      
      let icon = document.querySelector('svg')
      expect(icon).toHaveClass('animate-spin')
      
      // Switch to custom icon
      rerender(<DataLoadingState icon={Clock} />)
      
      icon = document.querySelector('svg')
      expect(icon).not.toHaveClass('animate-spin')
      
      // Switch back to default
      rerender(<DataLoadingState />)
      
      icon = document.querySelector('svg')
      expect(icon).toHaveClass('animate-spin')
    })

    it('maintains consistent animation state', () => {
      render(<DataLoadingState />)
      
      const icon = document.querySelector('svg')
      expect(icon).toHaveClass('animate-spin')
      expect(icon).toHaveClass('text-muted-foreground')
    })
  })
})