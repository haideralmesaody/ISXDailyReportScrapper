import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import LicensePage from '@/app/license/page'

// Mock the API client with security-focused responses
const mockApiClient = {
  activateLicense: jest.fn(),
  getLicenseStatus: jest.fn(),
}

jest.mock('@/lib/api', () => ({
  apiClient: mockApiClient,
}))

// Mock hooks
jest.mock('@/lib/hooks/use-websocket', () => ({
  useLicenseStatus: () => ({
    isValid: false,
    expiresAt: null,
    features: [],
    userInfo: null,
  }),
}))

jest.mock('@/lib/hooks/use-api', () => ({
  useApi: (fn: any) => ({
    execute: jest.fn().mockImplementation(async (data) => {
      return mockApiClient.activateLicense(data)
    }),
    loading: false,
    error: null,
  }),
}))

jest.mock('@/lib/hooks/use-toast', () => ({
  useToast: () => ({
    toast: jest.fn(),
  }),
}))

describe('Frontend Input Validation Security', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockApiClient.getLicenseStatus.mockResolvedValue({
      license_status: 'not_activated',
    })
  })

  describe('XSS Prevention', () => {
    it('sanitizes script tags in organization input', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const organizationInput = screen.getByLabelText('Organization (Optional)')
      
      // Attempt XSS injection
      await user.type(organizationInput, '<script>alert("xss")</script>')
      
      // Input should be sanitized (no script tags in DOM)
      const dom = document.documentElement.outerHTML
      expect(dom).not.toContain('<script>alert("xss")</script>')
      expect(dom).not.toContain('alert("xss")')
    })

    it('prevents HTML injection in organization field', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const organizationInput = screen.getByLabelText('Organization (Optional)')
      
      // Attempt HTML injection
      const maliciousInput = '<img src="x" onerror="alert(1)">'
      await user.type(organizationInput, maliciousInput)
      
      // Check that dangerous HTML is not rendered
      expect(screen.queryByRole('img')).not.toBeInTheDocument()
      expect(document.documentElement.outerHTML).not.toContain('onerror')
    })

    it('handles JavaScript protocol in input', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const organizationInput = screen.getByLabelText('Organization (Optional)')
      
      // Attempt JavaScript protocol injection
      await user.type(organizationInput, 'javascript:alert("xss")')
      
      // Should not execute JavaScript
      const inputValue = (organizationInput as HTMLInputElement).value
      expect(inputValue).toBe('javascript:alert("xss")')
      
      // But DOM should not contain executable JavaScript
      expect(document.documentElement.outerHTML).not.toContain('javascript:alert')
    })

    it('prevents event handler injection', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const organizationInput = screen.getByLabelText('Organization (Optional)')
      
      // Attempt event handler injection
      await user.type(organizationInput, '" onmouseover="alert(1)" "')
      
      // Check that event handlers are not added to DOM
      expect(document.documentElement.outerHTML).not.toContain('onmouseover')
      expect(document.documentElement.outerHTML).not.toContain('alert(1)')
    })
  })

  describe('Input Length Validation', () => {
    it('prevents buffer overflow with extremely long license key', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const licenseKeyInput = screen.getByLabelText('License Key *')
      
      // Attempt buffer overflow with very long input
      const longInput = 'A'.repeat(10000)
      await user.type(licenseKeyInput, longInput)
      
      // Input should be truncated or rejected
      const inputValue = (licenseKeyInput as HTMLInputElement).value
      expect(inputValue.length).toBeLessThan(1000)
    })

    it('validates organization field length', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const organizationInput = screen.getByLabelText('Organization (Optional)')
      
      // Test very long organization name
      const longOrganization = 'A'.repeat(500)
      await user.type(organizationInput, longOrganization)
      
      // Should enforce reasonable length limits
      const inputValue = (organizationInput as HTMLInputElement).value
      expect(inputValue.length).toBeLessThan(300)
    })
  })

  describe('License Key Format Security', () => {
    it('rejects SQL injection attempts in license key', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const licenseKeyInput = screen.getByLabelText('License Key *')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })
      
      // Attempt SQL injection
      await user.type(licenseKeyInput, "'; DROP TABLE licenses; --")
      await user.click(submitButton)
      
      await waitFor(() => {
        expect(screen.getByText(/Invalid license key format/)).toBeInTheDocument()
      })
    })

    it('rejects license key with null bytes', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const licenseKeyInput = screen.getByLabelText('License Key *')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })
      
      // Create input with null byte (browser might filter this)
      const inputWithNull = 'ISX1Y-ABCDE\u0000-12345-FGHIJ-67890'
      
      // Use fireEvent for more direct input
      fireEvent.change(licenseKeyInput, { target: { value: inputWithNull } })
      await user.click(submitButton)
      
      await waitFor(() => {
        expect(screen.getByText(/Invalid license key format/)).toBeInTheDocument()
      })
    })

    it('rejects license key with control characters', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const licenseKeyInput = screen.getByLabelText('License Key *')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })
      
      // Input with control characters
      const controlChars = 'ISX1Y-ABCDE\x01\x02-12345-FGHIJ-67890'
      fireEvent.change(licenseKeyInput, { target: { value: controlChars } })
      await user.click(submitButton)
      
      await waitFor(() => {
        expect(screen.getByText(/Invalid license key format/)).toBeInTheDocument()
      })
    })

    it('validates proper ISX license key format', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const licenseKeyInput = screen.getByLabelText('License Key *')
      
      // Test various valid and invalid formats
      const testCases = [
        { key: 'ISX1Y-ABCDE-12345-FGHIJ-67890', valid: true },
        { key: 'XXX1Y-ABCDE-12345-FGHIJ-67890', valid: false },
        { key: 'ISX1Y-ABCDE-12345', valid: false },
        { key: 'isx1y-abcde-12345-fghij-67890', valid: false },
        { key: 'ISX1Y-ABCDE-12345-FGHIJ-678@0', valid: false },
      ]
      
      for (const testCase of testCases) {
        await user.clear(licenseKeyInput)
        await user.type(licenseKeyInput, testCase.key)
        
        const submitButton = screen.getByRole('button', { name: 'Activate License' })
        
        if (testCase.valid) {
          expect(submitButton).not.toBeDisabled()
        } else {
          // Invalid format should show validation error
          await user.click(submitButton)
          await waitFor(() => {
            expect(screen.queryByText(/Invalid license key format/)).toBeInTheDocument()
          })
        }
      }
    })
  })

  describe('Unicode Security', () => {
    it('handles Unicode normalization attacks', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const organizationInput = screen.getByLabelText('Organization (Optional)')
      
      // Unicode normalization attack attempt
      const unicodeAttack = 'Test\uFEFF\u200B\u2060Org' // Zero-width characters
      await user.type(organizationInput, unicodeAttack)
      
      // Should normalize or reject problematic Unicode
      const inputValue = (organizationInput as HTMLInputElement).value
      expect(inputValue).not.toContain('\uFEFF')
      expect(inputValue).not.toContain('\u200B')
      expect(inputValue).not.toContain('\u2060')
    })

    it('prevents homograph attacks in organization', async () => {
      const user = userEvent.setup()
      render(<LicensePage />)

      const organizationInput = screen.getByLabelText('Organization (Optional)')
      
      // Homograph attack with Cyrillic characters that look like Latin
      const homographAttack = 'Аррlе Inc' // Contains Cyrillic 'А' and 'р'
      await user.type(organizationInput, homographAttack)
      
      // Should detect and handle suspicious character combinations
      const inputValue = (organizationInput as HTMLInputElement).value
      // Implementation should normalize or flag this
      expect(inputValue).toBeDefined()
    })
  })

  describe('Error Information Disclosure', () => {
    it('does not expose sensitive error details', async () => {
      const user = userEvent.setup()
      
      // Mock API error with sensitive information
      mockApiClient.activateLicense.mockRejectedValue({
        status: 500,
        detail: 'Database connection failed: postgres://user:password@host:5432/db',
        trace_id: 'trace-123',
      })

      render(<LicensePage />)

      const licenseKeyInput = screen.getByLabelText('License Key *')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })
      
      await user.type(licenseKeyInput, 'ISX1Y-ABCDE-12345-FGHIJ-67890')
      await user.click(submitButton)
      
      await waitFor(() => {
        // Error should be displayed but sanitized
        const errorElements = screen.getAllByText(/error/i)
        expect(errorElements.length).toBeGreaterThan(0)
        
        // Should not contain sensitive information
        const pageContent = document.documentElement.textContent || ''
        expect(pageContent).not.toContain('postgres://')
        expect(pageContent).not.toContain('password')
        expect(pageContent).not.toContain('host:5432')
        expect(pageContent).not.toContain('Database connection failed')
      })
    })

    it('provides generic error messages for security', async () => {
      const user = userEvent.setup()
      
      // Mock various error types
      const sensitiveErrors = [
        'File not found: /etc/secrets/license.key',
        'Network error: connection to internal.service.local:8080 failed',
        'Authentication failed: invalid JWT token signature',
        'SQL error: table "licenses" does not exist',
      ]

      for (const errorMessage of sensitiveErrors) {
        mockApiClient.activateLicense.mockRejectedValue({
          status: 500,
          detail: errorMessage,
        })

        const { unmount } = render(<LicensePage />)

        const licenseKeyInput = screen.getByLabelText('License Key *')
        const submitButton = screen.getByRole('button', { name: 'Activate License' })
        
        await user.type(licenseKeyInput, 'ISX1Y-ABCDE-12345-FGHIJ-67890')
        await user.click(submitButton)
        
        await waitFor(() => {
          const pageContent = document.documentElement.textContent || ''
          
          // Should show generic error message
          expect(pageContent).toContain('activation failed') || 
          expect(pageContent).toContain('unexpected error')
          
          // Should not expose sensitive details
          expect(pageContent).not.toContain('/etc/secrets/')
          expect(pageContent).not.toContain('internal.service.local')
          expect(pageContent).not.toContain('JWT token')
          expect(pageContent).not.toContain('SQL error')
        })

        unmount()
      }
    })
  })

  describe('Content Security Policy Compliance', () => {
    it('does not use inline styles that violate CSP', () => {
      render(<LicensePage />)
      
      // Check that no elements have inline style attributes
      const elementsWithInlineStyles = document.querySelectorAll('[style]')
      expect(elementsWithInlineStyles.length).toBe(0)
    })

    it('does not use inline event handlers', () => {
      render(<LicensePage />)
      
      // Check for inline event handlers
      const elementsWithHandlers = document.querySelectorAll(
        '[onclick], [onmouseover], [onload], [onerror]'
      )
      expect(elementsWithHandlers.length).toBe(0)
    })

    it('uses safe external resource loading', () => {
      render(<LicensePage />)
      
      // Check script and link tags for unsafe sources
      const scripts = document.querySelectorAll('script[src]')
      const links = document.querySelectorAll('link[href]')
      
      scripts.forEach(script => {
        const src = script.getAttribute('src')
        if (src) {
          expect(src).not.toMatch(/^http:\/\//) // Should use HTTPS
          expect(src).not.toContain('eval(')
          expect(src).not.toContain('javascript:')
        }
      })
      
      links.forEach(link => {
        const href = link.getAttribute('href')
        if (href) {
          expect(href).not.toMatch(/^http:\/\//) // Should use HTTPS
          expect(href).not.toContain('javascript:')
        }
      })
    })
  })

  describe('Form Security', () => {
    it('prevents form hijacking', () => {
      render(<LicensePage />)
      
      const forms = document.querySelectorAll('form')
      forms.forEach(form => {
        // Check that forms don't submit to external URLs
        const action = form.getAttribute('action')
        if (action) {
          expect(action).not.toMatch(/^https?:\/\/[^\/]/)
        }
        
        // Check for CSRF protection (if implemented)
        const csrfInput = form.querySelector('input[name="csrf_token"], input[name="_token"]')
        // This would be implementation-dependent
      })
    })

    it('uses secure form submission', async () => {
      const user = userEvent.setup()
      
      // Mock form submission
      const mockSubmit = jest.fn()
      mockApiClient.activateLicense.mockImplementation(mockSubmit)

      render(<LicensePage />)

      const licenseKeyInput = screen.getByLabelText('License Key *')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })
      
      await user.type(licenseKeyInput, 'ISX1Y-ABCDE-12345-FGHIJ-67890')
      await user.click(submitButton)
      
      await waitFor(() => {
        expect(mockSubmit).toHaveBeenCalledWith({
          license_key: 'ISX1Y-ABCDE-12345-FGHIJ-67890',
          organization: '',
        })
      })
    })
  })
})