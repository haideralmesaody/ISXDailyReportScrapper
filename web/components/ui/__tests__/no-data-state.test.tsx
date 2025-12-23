/**
 * @jest-environment jsdom
 */
import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { FileText, Activity, RefreshCw, AlertCircle } from 'lucide-react'
import { NoDataState, type NoDataStateProps, type NoDataAction } from '../no-data-state'

// Mock Next.js Link component
jest.mock('next/link', () => {
  return ({ children, href, ...props }: any) => (
    <a href={href} {...props}>
      {children}
    </a>
  )
})

describe('NoDataState Component', () => {
  const defaultProps: NoDataStateProps = {
    title: 'Test Title',
    description: 'Test description for the no data state',
  }

  describe('Basic Rendering', () => {
    const tests = [
      {
        name: 'renders with required props only',
        props: defaultProps,
        expectedTitle: 'Test Title',
        expectedDescription: 'Test description for the no data state',
      },
      {
        name: 'renders with icon',
        props: {
          ...defaultProps,
          icon: FileText,
        },
        expectedTitle: 'Test Title',
        expectedDescription: 'Test description for the no data state',
      },
      {
        name: 'renders with custom class',
        props: {
          ...defaultProps,
          className: 'custom-class',
        },
        expectedTitle: 'Test Title',
        expectedDescription: 'Test description for the no data state',
      },
    ]

    tests.forEach(({ name, props, expectedTitle, expectedDescription }) => {
      it(name, () => {
        render(<NoDataState {...props} />)
        
        expect(screen.getByText(expectedTitle)).toBeInTheDocument()
        expect(screen.getByText(expectedDescription)).toBeInTheDocument()
        expect(screen.getByRole('alert')).toBeInTheDocument()
        expect(screen.getByRole('alert')).toHaveAttribute('aria-live', 'polite')
      })
    })
  })

  describe('Icon Rendering', () => {
    const iconTests = [
      {
        name: 'renders without icon',
        props: defaultProps,
        shouldHaveIcon: false,
      },
      {
        name: 'renders with FileText icon',
        props: {
          ...defaultProps,
          icon: FileText,
        },
        shouldHaveIcon: true,
      },
      {
        name: 'renders with Activity icon',
        props: {
          ...defaultProps,
          icon: Activity,
        },
        shouldHaveIcon: true,
      },
    ]

    iconTests.forEach(({ name, props, shouldHaveIcon }) => {
      it(name, () => {
        render(<NoDataState {...props} />)
        
        const iconContainer = document.querySelector('.p-4.rounded-full')
        if (shouldHaveIcon) {
          expect(iconContainer).toBeInTheDocument()
          expect(iconContainer?.querySelector('svg')).toBeInTheDocument()
        } else {
          expect(iconContainer).not.toBeInTheDocument()
        }
      })
    })
  })

  describe('Icon Color Variants', () => {
    const colorTests = [
      {
        name: 'blue icon color (default)',
        iconColor: undefined,
        expectedClasses: 'bg-blue-100 text-blue-600',
      },
      {
        name: 'green icon color',
        iconColor: 'green' as const,
        expectedClasses: 'bg-green-100 text-green-600',
      },
      {
        name: 'purple icon color',
        iconColor: 'purple' as const,
        expectedClasses: 'bg-purple-100 text-purple-600',
      },
      {
        name: 'orange icon color',
        iconColor: 'orange' as const,
        expectedClasses: 'bg-orange-100 text-orange-600',
      },
      {
        name: 'red icon color',
        iconColor: 'red' as const,
        expectedClasses: 'bg-red-100 text-red-600',
      },
      {
        name: 'gray icon color',
        iconColor: 'gray' as const,
        expectedClasses: 'bg-gray-100 text-gray-600',
      },
    ]

    colorTests.forEach(({ name, iconColor, expectedClasses }) => {
      it(name, () => {
        render(
          <NoDataState 
            {...defaultProps}
            icon={FileText}
            iconColor={iconColor}
          />
        )
        
        const iconContainer = document.querySelector('.p-4.rounded-full')
        expect(iconContainer).toBeInTheDocument()
        
        const classArray = expectedClasses.split(' ')
        classArray.forEach(className => {
          expect(iconContainer).toHaveClass(className)
        })
      })
    })
  })

  describe('Instructions Rendering', () => {
    const instructionTests = [
      {
        name: 'renders without instructions',
        instructions: undefined,
        shouldHaveInstructions: false,
      },
      {
        name: 'renders with empty instructions array',
        instructions: [],
        shouldHaveInstructions: false,
      },
      {
        name: 'renders with single instruction',
        instructions: ['Step 1: Do something'],
        shouldHaveInstructions: true,
        expectedCount: 1,
      },
      {
        name: 'renders with multiple instructions',
        instructions: [
          'Step 1: Go to the Operations page',
          'Step 2: Run the pipeline',
          'Step 3: Wait for completion',
        ],
        shouldHaveInstructions: true,
        expectedCount: 3,
      },
    ]

    instructionTests.forEach(({ name, instructions, shouldHaveInstructions, expectedCount }) => {
      it(name, () => {
        render(
          <NoDataState 
            {...defaultProps}
            instructions={instructions}
          />
        )
        
        if (shouldHaveInstructions) {
          expect(screen.getByText('How to get started:')).toBeInTheDocument()
          const listItems = screen.getAllByRole('listitem')
          expect(listItems).toHaveLength(expectedCount!)
          
          instructions!.forEach(instruction => {
            expect(screen.getByText(instruction)).toBeInTheDocument()
          })
        } else {
          expect(screen.queryByText('How to get started:')).not.toBeInTheDocument()
          expect(screen.queryByRole('list')).not.toBeInTheDocument()
        }
      })
    })
  })

  describe('Actions Rendering', () => {
    const mockOnClick = jest.fn()

    beforeEach(() => {
      mockOnClick.mockClear()
    })

    const actionTests = [
      {
        name: 'renders without actions',
        actions: undefined,
        shouldHaveActions: false,
      },
      {
        name: 'renders with empty actions array',
        actions: [],
        shouldHaveActions: false,
      },
      {
        name: 'renders with single link action',
        actions: [
          {
            label: 'Go to Operations',
            href: '/operations',
          },
        ] as NoDataAction[],
        shouldHaveActions: true,
        expectedCount: 1,
      },
      {
        name: 'renders with single click action',
        actions: [
          {
            label: 'Retry',
            onClick: mockOnClick,
          },
        ] as NoDataAction[],
        shouldHaveActions: true,
        expectedCount: 1,
      },
      {
        name: 'renders with multiple mixed actions',
        actions: [
          {
            label: 'Go to Operations',
            variant: 'default',
            href: '/operations',
            icon: Activity,
          },
          {
            label: 'Check Again',
            variant: 'outline',
            onClick: mockOnClick,
            icon: RefreshCw,
          },
        ] as NoDataAction[],
        shouldHaveActions: true,
        expectedCount: 2,
      },
    ]

    actionTests.forEach(({ name, actions, shouldHaveActions, expectedCount }) => {
      it(name, () => {
        render(
          <NoDataState 
            {...defaultProps}
            actions={actions}
          />
        )
        
        if (shouldHaveActions) {
          const buttons = screen.queryAllByRole('button')
          const links = screen.queryAllByRole('link')
          const totalActions = buttons.length + links.length
          expect(totalActions).toBe(expectedCount!)
          
          actions!.forEach(action => {
            expect(screen.getByText(action.label)).toBeInTheDocument()
          })
        } else {
          expect(screen.queryByRole('button')).not.toBeInTheDocument()
          expect(screen.queryByRole('link')).not.toBeInTheDocument()
        }
      })
    })
  })

  describe('Link Actions', () => {
    it('renders link action correctly', () => {
      const actions: NoDataAction[] = [
        {
          label: 'Go to Operations',
          href: '/operations',
          variant: 'default',
          icon: Activity,
        },
      ]

      render(<NoDataState {...defaultProps} actions={actions} />)
      
      const link = screen.getByRole('link')
      expect(link).toBeInTheDocument()
      expect(link).toHaveAttribute('href', '/operations')
      expect(screen.getByText('Go to Operations')).toBeInTheDocument()
      
      // Check for icon
      expect(link.querySelector('svg')).toBeInTheDocument()
    })

    it('renders multiple link actions', () => {
      const actions: NoDataAction[] = [
        {
          label: 'Go to Operations',
          href: '/operations',
        },
        {
          label: 'Go to Reports',
          href: '/reports',
        },
      ]

      render(<NoDataState {...defaultProps} actions={actions} />)
      
      const links = screen.getAllByRole('link')
      expect(links).toHaveLength(2)
      
      expect(links[0]).toHaveAttribute('href', '/operations')
      expect(links[1]).toHaveAttribute('href', '/reports')
      
      expect(screen.getByText('Go to Operations')).toBeInTheDocument()
      expect(screen.getByText('Go to Reports')).toBeInTheDocument()
    })
  })

  describe('Click Actions', () => {
    it('renders click action correctly', () => {
      const mockOnClick = jest.fn()
      const actions: NoDataAction[] = [
        {
          label: 'Check Again',
          onClick: mockOnClick,
          variant: 'outline',
          icon: RefreshCw,
        },
      ]

      render(<NoDataState {...defaultProps} actions={actions} />)
      
      const button = screen.getByRole('button')
      expect(button).toBeInTheDocument()
      expect(screen.getByText('Check Again')).toBeInTheDocument()
      
      // Check for icon
      expect(button.querySelector('svg')).toBeInTheDocument()
      
      // Test click functionality
      fireEvent.click(button)
      expect(mockOnClick).toHaveBeenCalledTimes(1)
    })

    it('renders multiple click actions with different variants', () => {
      const mockOnClick1 = jest.fn()
      const mockOnClick2 = jest.fn()
      const actions: NoDataAction[] = [
        {
          label: 'Primary Action',
          onClick: mockOnClick1,
          variant: 'default',
        },
        {
          label: 'Secondary Action',
          onClick: mockOnClick2,
          variant: 'outline',
        },
      ]

      render(<NoDataState {...defaultProps} actions={actions} />)
      
      const buttons = screen.getAllByRole('button')
      expect(buttons).toHaveLength(2)
      
      const primaryButton = screen.getByText('Primary Action')
      const secondaryButton = screen.getByText('Secondary Action')
      
      expect(primaryButton).toBeInTheDocument()
      expect(secondaryButton).toBeInTheDocument()
      
      // Test both click functionalities
      fireEvent.click(primaryButton)
      fireEvent.click(secondaryButton)
      
      expect(mockOnClick1).toHaveBeenCalledTimes(1)
      expect(mockOnClick2).toHaveBeenCalledTimes(1)
    })
  })

  describe('Action Icons', () => {
    it('renders actions without icons', () => {
      const actions: NoDataAction[] = [
        {
          label: 'No Icon Action',
          onClick: jest.fn(),
        },
      ]

      render(<NoDataState {...defaultProps} actions={actions} />)
      
      const button = screen.getByRole('button')
      expect(button).toBeInTheDocument()
      expect(button.querySelector('svg')).not.toBeInTheDocument()
      expect(screen.getByText('No Icon Action')).toBeInTheDocument()
    })

    it('renders actions with different icons', () => {
      const actions: NoDataAction[] = [
        {
          label: 'Activity Action',
          onClick: jest.fn(),
          icon: Activity,
        },
        {
          label: 'Refresh Action',
          onClick: jest.fn(),
          icon: RefreshCw,
        },
      ]

      render(<NoDataState {...defaultProps} actions={actions} />)
      
      const buttons = screen.getAllByRole('button')
      expect(buttons).toHaveLength(2)
      
      buttons.forEach(button => {
        expect(button.querySelector('svg')).toBeInTheDocument()
      })
    })
  })

  describe('Accessibility', () => {
    it('has proper ARIA attributes', () => {
      render(<NoDataState {...defaultProps} />)
      
      const alertElement = screen.getByRole('alert')
      expect(alertElement).toHaveAttribute('aria-live', 'polite')
    })

    it('maintains semantic structure with instructions', () => {
      const instructions = ['Step 1', 'Step 2', 'Step 3']
      
      render(
        <NoDataState 
          {...defaultProps}
          instructions={instructions}
        />
      )
      
      expect(screen.getAllByRole('alert')).toHaveLength(2) // Main alert + instructions alert
      expect(screen.getByRole('list')).toBeInTheDocument()
      expect(screen.getAllByRole('listitem')).toHaveLength(3)
    })

    it('maintains proper heading hierarchy', () => {
      render(<NoDataState {...defaultProps} />)
      
      const heading = screen.getByRole('heading', { level: 2 })
      expect(heading).toBeInTheDocument()
      expect(heading).toHaveTextContent('Test Title')
    })

    it('provides accessible action buttons', () => {
      const actions: NoDataAction[] = [
        {
          label: 'Accessible Action',
          onClick: jest.fn(),
        },
      ]

      render(<NoDataState {...defaultProps} actions={actions} />)
      
      const button = screen.getByRole('button', { name: 'Accessible Action' })
      expect(button).toBeInTheDocument()
    })

    it('provides accessible navigation links', () => {
      const actions: NoDataAction[] = [
        {
          label: 'Accessible Link',
          href: '/test',
        },
      ]

      render(<NoDataState {...defaultProps} actions={actions} />)
      
      const link = screen.getByRole('link', { name: 'Accessible Link' })
      expect(link).toBeInTheDocument()
      expect(link).toHaveAttribute('href', '/test')
    })
  })

  describe('Layout and Styling', () => {
    it('applies custom className', () => {
      const customClass = 'custom-no-data-class'
      render(
        <NoDataState 
          {...defaultProps}
          className={customClass}
        />
      )
      
      const container = document.querySelector(`.${customClass}`)
      expect(container).toBeInTheDocument()
      expect(container).toHaveClass('min-h-screen', 'p-8', customClass)
    })

    it('has proper layout structure', () => {
      render(<NoDataState {...defaultProps} />)
      
      const container = document.querySelector('.min-h-screen')
      expect(container).toBeInTheDocument()
      expect(container).toHaveClass('min-h-screen', 'p-8')
      
      const innerContainer = document.querySelector('.max-w-7xl.mx-auto')
      expect(innerContainer).toBeInTheDocument()
      
      const card = document.querySelector('.border.rounded-lg')
      expect(card).toBeInTheDocument()
      expect(card).toHaveClass('p-8')
    })

    it('centers content properly', () => {
      render(<NoDataState {...defaultProps} />)
      
      const textCenter = document.querySelector('.text-center.space-y-6')
      expect(textCenter).toBeInTheDocument()
      
      const titleContainer = document.querySelector('.space-y-2')
      expect(titleContainer).toBeInTheDocument()
    })

    it('applies responsive design classes', () => {
      render(
        <NoDataState 
          {...defaultProps}
          instructions={['Test instruction']}
        />
      )
      
      const instructionsCard = document.querySelector('.max-w-md.mx-auto')
      expect(instructionsCard).toBeInTheDocument()
      
      const actionsContainer = document.querySelector('.flex.gap-4.justify-center.flex-wrap')
      // Only exists if actions are provided
      expect(actionsContainer).not.toBeInTheDocument()
    })
  })

  describe('Complex Scenarios', () => {
    it('renders complete state with all props', () => {
      const mockOnClick = jest.fn()
      const fullProps: NoDataStateProps = {
        icon: AlertCircle,
        iconColor: 'red',
        title: 'Complex No Data State',
        description: 'This is a complex example with all possible props',
        className: 'complex-state',
        instructions: [
          'First step of the process',
          'Second step with more detail',
          'Final step to complete',
        ],
        actions: [
          {
            label: 'Primary Action',
            variant: 'default',
            href: '/primary',
            icon: Activity,
          },
          {
            label: 'Secondary Action',
            variant: 'outline',
            onClick: mockOnClick,
            icon: RefreshCw,
          },
        ],
      }

      render(<NoDataState {...fullProps} />)
      
      // Check all elements are present
      expect(screen.getByText('Complex No Data State')).toBeInTheDocument()
      expect(screen.getByText('This is a complex example with all possible props')).toBeInTheDocument()
      expect(screen.getByText('How to get started:')).toBeInTheDocument()
      expect(screen.getAllByRole('listitem')).toHaveLength(3)
      expect(screen.getByRole('link')).toHaveAttribute('href', '/primary')
      expect(screen.getByRole('button')).toBeInTheDocument()
      
      // Test interaction
      fireEvent.click(screen.getByText('Secondary Action'))
      expect(mockOnClick).toHaveBeenCalledTimes(1)
      
      // Check icon color
      const iconContainer = document.querySelector('.bg-red-100.text-red-600')
      expect(iconContainer).toBeInTheDocument()
      
      // Check custom class
      expect(document.querySelector('.complex-state')).toBeInTheDocument()
    })

    it('handles mixed action types correctly', () => {
      const mockOnClick = jest.fn()
      const actions: NoDataAction[] = [
        {
          label: 'Link Action',
          href: '/link-target',
          variant: 'default',
        },
        {
          label: 'Click Action',
          onClick: mockOnClick,
          variant: 'secondary',
        },
        {
          label: 'Link with Icon',
          href: '/icon-link',
          icon: FileText,
        },
        {
          label: 'Click with Icon',
          onClick: mockOnClick,
          icon: RefreshCw,
        },
      ]

      render(<NoDataState {...defaultProps} actions={actions} />)
      
      // Should have 2 links and 2 buttons
      const links = screen.queryAllByRole('link')
      const buttons = screen.queryAllByRole('button')
      
      expect(links).toHaveLength(2)
      expect(buttons).toHaveLength(2)
      
      // Test links
      expect(links[0]).toHaveAttribute('href', '/link-target')
      expect(links[1]).toHaveAttribute('href', '/icon-link')
      
      // Test buttons
      fireEvent.click(buttons[0])
      fireEvent.click(buttons[1])
      expect(mockOnClick).toHaveBeenCalledTimes(2)
      
      // Check icons are present
      expect(links[1].querySelector('svg')).toBeInTheDocument()
      expect(buttons[1].querySelector('svg')).toBeInTheDocument()
    })
  })

  describe('Edge Cases', () => {
    it('handles undefined onClick gracefully', () => {
      const actions: NoDataAction[] = [
        {
          label: 'Action without onClick',
          onClick: undefined,
        },
      ]

      expect(() => {
        render(<NoDataState {...defaultProps} actions={actions} />)
      }).not.toThrow()
      
      const button = screen.getByRole('button')
      expect(button).toBeInTheDocument()
      
      // Should not throw when clicked
      expect(() => {
        fireEvent.click(button)
      }).not.toThrow()
    })

    it('handles empty strings gracefully', () => {
      const props = {
        title: '',
        description: '',
      }

      render(<NoDataState {...props} />)
      
      // Should still render structure even with empty strings
      expect(screen.getByRole('alert')).toBeInTheDocument()
      expect(document.querySelector('.text-center')).toBeInTheDocument()
    })

    it('handles very long text content', () => {
      const longProps = {
        title: 'A'.repeat(100),
        description: 'B'.repeat(500),
      }

      render(<NoDataState {...longProps} />)
      
      expect(screen.getByText(longProps.title)).toBeInTheDocument()
      expect(screen.getByText(longProps.description)).toBeInTheDocument()
    })

    it('handles many instructions and actions', () => {
      const manyInstructions = Array.from({ length: 10 }, (_, i) => `Instruction ${i + 1}`)
      const manyActions = Array.from({ length: 5 }, (_, i) => ({
        label: `Action ${i + 1}`,
        onClick: jest.fn(),
      }))

      render(
        <NoDataState 
          {...defaultProps}
          instructions={manyInstructions}
          actions={manyActions}
        />
      )
      
      expect(screen.getAllByRole('listitem')).toHaveLength(10)
      expect(screen.getAllByRole('button')).toHaveLength(5)
    })
  })
})