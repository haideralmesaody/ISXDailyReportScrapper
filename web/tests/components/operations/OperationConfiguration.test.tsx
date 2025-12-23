import React from 'react'
import { render, screen, waitFor, fireEvent, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OperationConfiguration } from '@/components/operations/OperationConfiguration'
import { apiClient } from '@/lib/api'
import '@testing-library/jest-dom'

// Mock dependencies
jest.mock('@/lib/api')

describe('OperationConfiguration', () => {
  const defaultProps = {
    operationId: 'op-123',
    operationType: 'report_generation',
    currentConfig: {
      schedule: 'daily',
      timezone: 'Asia/Baghdad',
      retryAttempts: 3,
      timeout: 300,
    },
    onConfigSave: jest.fn(),
    onCancel: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('Rendering', () => {
    const testCases = [
      {
        name: 'renders configuration form with current values',
        props: defaultProps,
        assertions: () => {
          expect(screen.getByLabelText(/Schedule/i)).toHaveValue('daily')
          expect(screen.getByLabelText(/Timezone/i)).toHaveValue('Asia/Baghdad')
          expect(screen.getByLabelText(/Retry Attempts/i)).toHaveValue(3)
          expect(screen.getByLabelText(/Timeout/i)).toHaveValue(300)
        },
      },
      {
        name: 'renders operation-specific fields for report generation',
        props: {
          ...defaultProps,
          operationType: 'report_generation',
          currentConfig: {
            ...defaultProps.currentConfig,
            reportFormat: 'pdf',
            includeSummary: true,
            emailRecipients: ['admin@example.com'],
          },
        },
        assertions: () => {
          expect(screen.getByLabelText(/Report Format/i)).toBeInTheDocument()
          expect(screen.getByLabelText(/Include Summary/i)).toBeChecked()
          expect(screen.getByLabelText(/Email Recipients/i)).toHaveValue('admin@example.com')
        },
      },
      {
        name: 'renders operation-specific fields for data scraping',
        props: {
          ...defaultProps,
          operationType: 'data_scraping',
          currentConfig: {
            ...defaultProps.currentConfig,
            dataSource: 'isx_website',
            scrapeInterval: 60,
            maxRetries: 5,
          },
        },
        assertions: () => {
          expect(screen.getByLabelText(/Data Source/i)).toHaveValue('isx_website')
          expect(screen.getByLabelText(/Scrape Interval/i)).toHaveValue(60)
          expect(screen.getByLabelText(/Max Retries/i)).toHaveValue(5)
        },
      },
      {
        name: 'renders operation-specific fields for data processing',
        props: {
          ...defaultProps,
          operationType: 'data_processing',
          currentConfig: {
            ...defaultProps.currentConfig,
            batchSize: 100,
            parallelWorkers: 4,
            processingMode: 'batch',
          },
        },
        assertions: () => {
          expect(screen.getByLabelText(/Batch Size/i)).toHaveValue(100)
          expect(screen.getByLabelText(/Parallel Workers/i)).toHaveValue(4)
          expect(screen.getByLabelText(/Processing Mode/i)).toHaveValue('batch')
        },
      },
      {
        name: 'displays loading state when fetching configuration schema',
        props: {
          ...defaultProps,
          isLoading: true,
        },
        assertions: () => {
          expect(screen.getByTestId('config-loading')).toBeInTheDocument()
          expect(screen.getByText(/Loading configuration/i)).toBeInTheDocument()
          expect(screen.queryByRole('form')).not.toBeInTheDocument()
        },
      },
    ]

    testCases.forEach(({ name, props, assertions }) => {
      it(name, () => {
        render(<OperationConfiguration {...props} />)
        assertions()
      })
    })
  })

  describe('Form Validation', () => {
    const validationTestCases = [
      {
        name: 'validates required fields',
        actions: async () => {
          const scheduleInput = screen.getByLabelText(/Schedule/i)
          await userEvent.clear(scheduleInput)
          await userEvent.tab() // Trigger blur
        },
        assertions: () => {
          expect(screen.getByText(/Schedule is required/i)).toBeInTheDocument()
          expect(screen.getByRole('button', { name: /Save/i })).toBeDisabled()
        },
      },
      {
        name: 'validates numeric ranges',
        actions: async () => {
          const retryInput = screen.getByLabelText(/Retry Attempts/i)
          await userEvent.clear(retryInput)
          await userEvent.type(retryInput, '-1')
        },
        assertions: () => {
          expect(screen.getByText(/Must be between 0 and 10/i)).toBeInTheDocument()
        },
      },
      {
        name: 'validates email format',
        props: {
          ...defaultProps,
          operationType: 'report_generation',
          currentConfig: {
            ...defaultProps.currentConfig,
            emailRecipients: [],
          },
        },
        actions: async () => {
          const emailInput = screen.getByLabelText(/Email Recipients/i)
          await userEvent.type(emailInput, 'invalid-email')
          await userEvent.keyboard('{Enter}')
        },
        assertions: () => {
          expect(screen.getByText(/Invalid email format/i)).toBeInTheDocument()
        },
      },
      {
        name: 'validates cron expression for custom schedules',
        actions: async () => {
          const scheduleSelect = screen.getByLabelText(/Schedule/i)
          await userEvent.selectOptions(scheduleSelect, 'custom')
          
          const cronInput = screen.getByLabelText(/Cron Expression/i)
          await userEvent.type(cronInput, 'invalid cron')
        },
        assertions: () => {
          expect(screen.getByText(/Invalid cron expression/i)).toBeInTheDocument()
        },
      },
      {
        name: 'validates URL format for webhook configurations',
        props: {
          ...defaultProps,
          currentConfig: {
            ...defaultProps.currentConfig,
            webhookUrl: '',
          },
        },
        actions: async () => {
          const webhookInput = screen.getByLabelText(/Webhook URL/i)
          await userEvent.type(webhookInput, 'not-a-url')
        },
        assertions: () => {
          expect(screen.getByText(/Must be a valid URL/i)).toBeInTheDocument()
        },
      },
    ]

    validationTestCases.forEach(({ name, props = defaultProps, actions, assertions }) => {
      it(name, async () => {
        render(<OperationConfiguration {...props} />)
        await actions()
        assertions()
      })
    })
  })

  describe('User Interactions', () => {
    it('updates configuration values on input', async () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      const timeoutInput = screen.getByLabelText(/Timeout/i)
      await userEvent.clear(timeoutInput)
      await userEvent.type(timeoutInput, '600')
      
      expect(timeoutInput).toHaveValue(600)
    })

    it('handles schedule type changes', async () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      const scheduleSelect = screen.getByLabelText(/Schedule/i)
      await userEvent.selectOptions(scheduleSelect, 'hourly')
      
      // Should show hour selection
      expect(screen.getByLabelText(/Hour of day/i)).toBeInTheDocument()
    })

    it('adds and removes email recipients', async () => {
      const props = {
        ...defaultProps,
        operationType: 'report_generation',
        currentConfig: {
          ...defaultProps.currentConfig,
          emailRecipients: ['existing@example.com'],
        },
      }
      
      render(<OperationConfiguration {...props} />)
      
      // Add new email
      const emailInput = screen.getByLabelText(/Add Email/i)
      await userEvent.type(emailInput, 'new@example.com')
      await userEvent.keyboard('{Enter}')
      
      expect(screen.getByText('new@example.com')).toBeInTheDocument()
      
      // Remove email
      const removeButton = screen.getByTestId('remove-email-new@example.com')
      await userEvent.click(removeButton)
      
      expect(screen.queryByText('new@example.com')).not.toBeInTheDocument()
    })

    it('toggles advanced settings', async () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      expect(screen.queryByText(/Advanced Settings/i)).not.toBeInTheDocument()
      
      const toggleButton = screen.getByRole('button', { name: /Show Advanced/i })
      await userEvent.click(toggleButton)
      
      expect(screen.getByText(/Advanced Settings/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Enable Debug Mode/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Custom Headers/i)).toBeInTheDocument()
    })

    it('saves configuration with validation', async () => {
      ;(apiClient.updateOperationConfig as jest.Mock).mockResolvedValue({
        success: true,
        config: { ...defaultProps.currentConfig, timeout: 600 },
      })
      
      render(<OperationConfiguration {...defaultProps} />)
      
      const timeoutInput = screen.getByLabelText(/Timeout/i)
      await userEvent.clear(timeoutInput)
      await userEvent.type(timeoutInput, '600')
      
      const saveButton = screen.getByRole('button', { name: /Save/i })
      await userEvent.click(saveButton)
      
      await waitFor(() => {
        expect(apiClient.updateOperationConfig).toHaveBeenCalledWith('op-123', {
          schedule: 'daily',
          timezone: 'Asia/Baghdad',
          retryAttempts: 3,
          timeout: 600,
        })
      })
      
      expect(defaultProps.onConfigSave).toHaveBeenCalledWith({
        schedule: 'daily',
        timezone: 'Asia/Baghdad',
        retryAttempts: 3,
        timeout: 600,
      })
    })

    it('shows confirmation dialog for critical changes', async () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      // Change schedule from daily to manual
      const scheduleSelect = screen.getByLabelText(/Schedule/i)
      await userEvent.selectOptions(scheduleSelect, 'manual')
      
      const saveButton = screen.getByRole('button', { name: /Save/i })
      await userEvent.click(saveButton)
      
      // Confirmation dialog should appear
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByText(/This will disable automatic execution/i)).toBeInTheDocument()
      
      const confirmButton = within(screen.getByRole('dialog')).getByRole('button', { name: /Confirm/i })
      await userEvent.click(confirmButton)
      
      expect(apiClient.updateOperationConfig).toHaveBeenCalled()
    })

    it('cancels without saving changes', async () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      const timeoutInput = screen.getByLabelText(/Timeout/i)
      await userEvent.clear(timeoutInput)
      await userEvent.type(timeoutInput, '600')
      
      const cancelButton = screen.getByRole('button', { name: /Cancel/i })
      await userEvent.click(cancelButton)
      
      expect(apiClient.updateOperationConfig).not.toHaveBeenCalled()
      expect(defaultProps.onCancel).toHaveBeenCalled()
    })

    it('shows unsaved changes warning', async () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      const timeoutInput = screen.getByLabelText(/Timeout/i)
      await userEvent.clear(timeoutInput)
      await userEvent.type(timeoutInput, '600')
      
      // Try to navigate away
      const cancelButton = screen.getByRole('button', { name: /Cancel/i })
      await userEvent.click(cancelButton)
      
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByText(/You have unsaved changes/i)).toBeInTheDocument()
    })
  })

  describe('Dynamic Field Rendering', () => {
    it('shows conditional fields based on other values', async () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      // Enable notifications
      const notificationToggle = screen.getByLabelText(/Enable Notifications/i)
      await userEvent.click(notificationToggle)
      
      // Should show notification-related fields
      expect(screen.getByLabelText(/Notification Channel/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Notification Threshold/i)).toBeInTheDocument()
    })

    it('validates dependent fields', async () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      // Enable webhook
      const webhookToggle = screen.getByLabelText(/Enable Webhook/i)
      await userEvent.click(webhookToggle)
      
      // Try to save without webhook URL
      const saveButton = screen.getByRole('button', { name: /Save/i })
      await userEvent.click(saveButton)
      
      expect(screen.getByText(/Webhook URL is required when webhook is enabled/i)).toBeInTheDocument()
    })
  })

  describe('Error Handling', () => {
    it('displays error when configuration save fails', async () => {
      ;(apiClient.updateOperationConfig as jest.Mock).mockRejectedValue(
        new Error('Configuration update failed')
      )
      
      render(<OperationConfiguration {...defaultProps} />)
      
      const saveButton = screen.getByRole('button', { name: /Save/i })
      await userEvent.click(saveButton)
      
      await waitFor(() => {
        expect(screen.getByTestId('error-alert')).toBeInTheDocument()
        expect(screen.getByText(/Configuration update failed/i)).toBeInTheDocument()
      })
    })

    it('handles network errors gracefully', async () => {
      ;(apiClient.updateOperationConfig as jest.Mock).mockRejectedValue(
        new Error('Network error')
      )
      
      render(<OperationConfiguration {...defaultProps} />)
      
      const saveButton = screen.getByRole('button', { name: /Save/i })
      await userEvent.click(saveButton)
      
      await waitFor(() => {
        expect(screen.getByText(/Network error/i)).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /Retry/i })).toBeInTheDocument()
      })
    })
  })

  describe('Accessibility', () => {
    it('has proper form labels and ARIA attributes', () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      const form = screen.getByRole('form')
      expect(form).toHaveAccessibleName(/Operation Configuration/i)
      
      // All inputs should have labels
      const inputs = screen.getAllByRole('textbox')
      inputs.forEach(input => {
        expect(input).toHaveAccessibleName()
      })
    })

    it('shows validation errors with proper ARIA', async () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      const scheduleInput = screen.getByLabelText(/Schedule/i)
      await userEvent.clear(scheduleInput)
      await userEvent.tab()
      
      const errorMessage = screen.getByText(/Schedule is required/i)
      expect(errorMessage).toHaveAttribute('role', 'alert')
      expect(scheduleInput).toHaveAttribute('aria-invalid', 'true')
      expect(scheduleInput).toHaveAttribute('aria-describedby', expect.stringContaining('error'))
    })

    it('supports keyboard navigation', async () => {
      render(<OperationConfiguration {...defaultProps} />)
      
      // Tab through form fields
      await userEvent.tab() // Schedule
      expect(screen.getByLabelText(/Schedule/i)).toHaveFocus()
      
      await userEvent.tab() // Timezone
      expect(screen.getByLabelText(/Timezone/i)).toHaveFocus()
      
      await userEvent.tab() // Retry Attempts
      expect(screen.getByLabelText(/Retry Attempts/i)).toHaveFocus()
    })
  })

  describe('Performance', () => {
    it('debounces validation for text inputs', async () => {
      const validateSpy = jest.fn()
      render(
        <OperationConfiguration 
          {...defaultProps} 
          onValidate={validateSpy}
        />
      )
      
      const timeoutInput = screen.getByLabelText(/Timeout/i)
      
      // Type quickly
      await userEvent.type(timeoutInput, '123456')
      
      // Validation should be debounced
      expect(validateSpy).toHaveBeenCalledTimes(1) // Only once after debounce
    })

    it('memoizes complex calculations', () => {
      const { rerender } = render(<OperationConfiguration {...defaultProps} />)
      
      // Re-render with same props
      rerender(<OperationConfiguration {...defaultProps} />)
      
      // Should not recalculate derived state (implementation-specific test)
      expect(screen.getByTestId('render-count')).toHaveTextContent('1')
    })
  })
})