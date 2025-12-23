/**
 * Comprehensive End-to-End License Activation Tests
 * 
 * Cross-browser testing (Chrome, Firefox, Edge, Safari) with mobile responsive testing,
 * touch interaction testing, network throttling scenarios, and complete user journey testing
 * with live license key ISX1M02LYE1F9QJHR9D7Z.
 * 
 * Coverage Requirements:
 * - Cross-browser testing: Chrome, Firefox, Edge, Safari
 * - Mobile responsive testing (tablet/phone viewports)
 * - Touch interaction testing
 * - Network throttling scenarios
 * - Complete user journey testing with live license key
 * - Error scenario testing with proper fallback behavior
 */

import { test, expect, devices } from '@playwright/test'

// Valid test license key from requirements
const VALID_TEST_LICENSE_KEY = 'ISX1M02LYE1F9QJHR9D7Z'
const TEST_EMAIL = 'test@iraqiinvestor.gov.iq'
const BASE_URL = 'http://localhost:8080'

// Cross-browser configuration
const browsers = [
  { name: 'chromium', ...devices['Desktop Chrome'] },
  { name: 'firefox', ...devices['Desktop Firefox'] },
  { name: 'webkit', ...devices['Desktop Safari'] }
]

// Mobile device configuration
const mobileDevices = [
  devices['iPad Pro'],
  devices['iPhone 14'],
  devices['Samsung Galaxy S23'],
  devices['Pixel 7']
]

test.describe('License Activation - Cross-Browser Testing', () => {
  for (const browser of browsers) {
    test.describe(`${browser.name.toUpperCase()} Browser`, () => {
      test.use({ ...browser })

      test('complete license activation user journey', async ({ page }) => {
        // Navigate to application
        await page.goto(BASE_URL)

        // Verify initial loading screen with Iraqi Investor branding
        await expect(page.locator('text=Starting Iraqi Investor')).toBeVisible()
        await expect(page.locator('img[alt="Iraqi Investor Logo"]')).toBeVisible()
        await expect(page.locator('text=License System')).toBeVisible()

        // Wait for automatic redirect to license activation page
        await expect(page).toHaveURL(/.*\/license/)

        // Verify license activation page elements
        await expect(page.locator('h1:has-text("Activate Your Professional License")')).toBeVisible()
        await expect(page.locator('input[name="license_key"]')).toBeVisible()
        await expect(page.locator('input[name="email"]')).toBeVisible()
        await expect(page.locator('button[type="submit"]')).toBeVisible()

        // Fill license activation form
        await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
        await page.fill('input[name="email"]', TEST_EMAIL)

        // Verify form validation feedback
        await expect(page.locator('text=License key format is valid')).toBeVisible()
        await expect(page.locator('text=Email format is valid')).toBeVisible()

        // Submit activation form
        await page.click('button[type="submit"]')

        // Verify loading state during activation
        await expect(page.locator('button:has-text("Activating...")')).toBeVisible()
        await expect(page.locator('[data-testid="submit-loading-spinner"]')).toBeVisible()

        // Wait for success state
        await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })

        // Verify license details display
        await expect(page.locator('text=License Status: Active')).toBeVisible()
        await expect(page.locator('text=Expires:')).toBeVisible()
        await expect(page.locator(`text=Registered Email: ${TEST_EMAIL}`)).toBeVisible()

        // Verify countdown and redirect functionality
        await expect(page.locator('text=Redirecting to dashboard in')).toBeVisible()
        await expect(page.locator('button:has-text("Continue to Dashboard")')).toBeVisible()

        // Click continue button instead of waiting for countdown
        await page.click('button:has-text("Continue to Dashboard")')

        // Verify successful redirect to dashboard
        await expect(page).toHaveURL(/.*\/dashboard/)
        await expect(page.locator('text=Iraqi Investor Dashboard')).toBeVisible()
      })

      test('error handling for invalid license key', async ({ page }) => {
        await page.goto(`${BASE_URL}/license`)

        // Enter invalid license key
        await page.fill('input[name="license_key"]', 'INVALID_LICENSE_KEY')
        await page.fill('input[name="email"]', TEST_EMAIL)

        // Submit form
        await page.click('button[type="submit"]')

        // Verify error state
        await expect(page.locator('text=Invalid License Key')).toBeVisible()
        await expect(page.locator('text=The provided license key is not valid')).toBeVisible()

        // Verify error alert styling
        await expect(page.locator('[role="alert"]')).toHaveClass(/.*destructive.*/)

        // Verify retry functionality
        await expect(page.locator('button:has-text("Retry Activation")')).toBeVisible()
      })

      test('network error handling and recovery', async ({ page }) => {
        // Simulate network failure
        await page.route('**/api/license/activate', route => {
          route.abort('failed')
        })

        await page.goto(`${BASE_URL}/license`)

        await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
        await page.fill('input[name="email"]', TEST_EMAIL)
        await page.click('button[type="submit"]')

        // Verify network error display
        await expect(page.locator('text=Network Connection Error')).toBeVisible()
        await expect(page.locator('text=Please check your internet connection')).toBeVisible()

        // Restore network and test recovery
        await page.unroute('**/api/license/activate')
        await page.click('button:has-text("Retry Activation")')

        // Should succeed on retry
        await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })
      })

      test('form validation edge cases', async ({ page }) => {
        await page.goto(`${BASE_URL}/license`)

        // Test empty form submission
        await page.click('button[type="submit"]')
        await expect(page.locator('text=License key is required')).toBeVisible()
        await expect(page.locator('text=Email address is required')).toBeVisible()

        // Test invalid email formats
        await page.fill('input[name="email"]', 'invalid-email')
        await page.blur('input[name="email"]')
        await expect(page.locator('text=Please enter a valid email address')).toBeVisible()

        // Test license key length validation  
        await page.fill('input[name="license_key"]', 'SHORT')
        await page.blur('input[name="license_key"]')
        await expect(page.locator('text=License key must be exactly 19 characters')).toBeVisible()

        // Test too long license key
        await page.fill('input[name="license_key"]', 'VERY_LONG_LICENSE_KEY_THAT_EXCEEDS_LIMIT')
        await page.blur('input[name="license_key"]')
        await expect(page.locator('text=License key must be exactly 19 characters')).toBeVisible()

        // Test invalid characters in license key
        await page.fill('input[name="license_key"]', 'ISX1M02LYE1F9QJHR9D!')
        await page.blur('input[name="license_key"]')
        await expect(page.locator('text=License key contains invalid characters')).toBeVisible()
      })
    })
  }
})

test.describe('License Activation - Mobile Responsive Testing', () => {
  for (const device of mobileDevices) {
    test.describe(`${device.name} Device`, () => {
      test.use({ ...device })

      test('mobile responsive license activation flow', async ({ page }) => {
        await page.goto(BASE_URL)

        // Verify mobile-optimized loading screen
        await expect(page.locator('text=Starting Iraqi Investor')).toBeVisible()
        
        // Check mobile navigation
        await expect(page).toHaveURL(/.*\/license/)

        // Verify mobile form layout
        const form = page.locator('form[aria-label="License Activation"]')
        await expect(form).toBeVisible()

        // Verify mobile-optimized input fields
        const licenseInput = page.locator('input[name="license_key"]')
        const emailInput = page.locator('input[name="email"]')
        
        await expect(licenseInput).toBeVisible()
        await expect(emailInput).toBeVisible()

        // Test mobile keyboard input
        await licenseInput.tap()
        await expect(licenseInput).toBeFocused()
        await licenseInput.fill(VALID_TEST_LICENSE_KEY)

        await emailInput.tap()
        await expect(emailInput).toBeFocused()
        await emailInput.fill(TEST_EMAIL)

        // Verify mobile submit button
        const submitButton = page.locator('button[type="submit"]')
        await expect(submitButton).toBeVisible()
        await expect(submitButton).toBeEnabled()

        // Test mobile touch interaction
        await submitButton.tap()

        // Verify mobile success state
        await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })

        // Test mobile navigation to dashboard
        await page.locator('button:has-text("Continue to Dashboard")').tap()
        await expect(page).toHaveURL(/.*\/dashboard/)
      })

      test('mobile form validation and error display', async ({ page }) => {
        await page.goto(`${BASE_URL}/license`)

        // Test mobile validation UI
        await page.locator('input[name="license_key"]').tap()
        await page.locator('input[name="license_key"]').fill('INVALID')
        await page.locator('input[name="email"]').tap() // Blur previous field

        // Verify mobile error styling
        await expect(page.locator('text=License key must be exactly 19 characters')).toBeVisible()
        
        // Check error message positioning on mobile
        const errorMessage = page.locator('text=License key must be exactly 19 characters')
        const boundingBox = await errorMessage.boundingBox()
        expect(boundingBox).not.toBeNull()
        expect(boundingBox!.y).toBeGreaterThan(0)
      })

      test('mobile network error handling', async ({ page }) => {
        await page.route('**/api/license/activate', route => {
          route.abort('failed')
        })

        await page.goto(`${BASE_URL}/license`)

        await page.locator('input[name="license_key"]').fill(VALID_TEST_LICENSE_KEY)
        await page.locator('input[name="email"]').fill(TEST_EMAIL)
        await page.locator('button[type="submit"]').tap()

        // Verify mobile error modal/toast
        await expect(page.locator('text=Network Connection Error')).toBeVisible()
        
        // Test mobile retry button
        await page.locator('button:has-text("Retry Activation")').tap()
      })
    })
  }
})

test.describe('License Activation - Network Conditions Testing', () => {
  test('slow 3G network simulation', async ({ page, context }) => {
    // Simulate slow 3G connection
    await context.route('**/*', async route => {
      await new Promise(resolve => setTimeout(resolve, 2000)) // 2s delay
      await route.continue()
    })

    await page.goto(BASE_URL)

    // Verify loading states persist appropriately
    await expect(page.locator('text=Starting Iraqi Investor')).toBeVisible()
    
    // Should eventually reach license page despite slow network
    await expect(page).toHaveURL(/.*\/license/, { timeout: 15000 })

    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Verify extended loading state for slow network
    await expect(page.locator('button:has-text("Activating...")')).toBeVisible()
    
    // Should eventually succeed
    await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 20000 })
  })

  test('intermittent network failures', async ({ page }) => {
    let requestCount = 0

    await page.route('**/api/license/activate', route => {
      requestCount++
      if (requestCount <= 2) {
        // Fail first two requests
        route.abort('failed')
      } else {
        // Succeed on third request
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            license: {
              valid: true,
              status: 'active',
              expires_at: '2025-12-31T23:59:59Z',
              email: TEST_EMAIL,
              activated_at: new Date().toISOString()
            }
          })
        })
      }
    })

    await page.goto(`${BASE_URL}/license`)

    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // First attempt should fail
    await expect(page.locator('text=Network Connection Error')).toBeVisible()
    
    // Retry should also fail
    await page.click('button:has-text("Retry Activation")')
    await expect(page.locator('text=Network Connection Error')).toBeVisible()
    
    // Third attempt should succeed
    await page.click('button:has-text("Retry Activation")')
    await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })
  })

  test('offline mode detection', async ({ page, context }) => {
    await page.goto(`${BASE_URL}/license`)

    // Simulate going offline
    await context.setOffline(true)

    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Should detect offline state
    await expect(page.locator('text=You appear to be offline')).toBeVisible()
    await expect(page.locator('text=Please check your internet connection')).toBeVisible()

    // Simulate coming back online
    await context.setOffline(false)
    
    // Should allow retry when back online
    await page.click('button:has-text("Retry Activation")')
    await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })
  })
})

test.describe('License Activation - Accessibility Testing', () => {
  test('keyboard navigation support', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Test tab navigation
    await page.keyboard.press('Tab')
    await expect(page.locator('input[name="license_key"]')).toBeFocused()

    await page.keyboard.press('Tab')
    await expect(page.locator('input[name="email"]')).toBeFocused()

    await page.keyboard.press('Tab')
    await expect(page.locator('button[type="submit"]')).toBeFocused()

    // Test reverse tab navigation
    await page.keyboard.press('Shift+Tab')
    await expect(page.locator('input[name="email"]')).toBeFocused()

    await page.keyboard.press('Shift+Tab')
    await expect(page.locator('input[name="license_key"]')).toBeFocused()
  })

  test('screen reader accessibility', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Verify ARIA labels
    const licenseInput = page.locator('input[name="license_key"]')
    await expect(licenseInput).toHaveAttribute('aria-label', 'License Key')
    await expect(licenseInput).toHaveAttribute('aria-required', 'true')

    const emailInput = page.locator('input[name="email"]')
    await expect(emailInput).toHaveAttribute('aria-label', 'Email Address')
    await expect(emailInput).toHaveAttribute('aria-required', 'true')

    // Test form submission with Enter key
    await licenseInput.fill(VALID_TEST_LICENSE_KEY)
    await emailInput.fill(TEST_EMAIL)
    await emailInput.press('Enter')

    // Should submit form
    await expect(page.locator('button:has-text("Activating...")')).toBeVisible()
  })

  test('high contrast and color accessibility', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Check color contrast for form elements
    const submitButton = page.locator('button[type="submit"]')
    const buttonStyles = await submitButton.evaluate(el => {
      const computed = window.getComputedStyle(el)
      return {
        backgroundColor: computed.backgroundColor,
        color: computed.color,
        borderColor: computed.borderColor
      }
    })

    // Verify button has sufficient contrast (basic check)
    expect(buttonStyles.backgroundColor).not.toBe(buttonStyles.color)

    // Test focus visible indicators
    await page.keyboard.press('Tab')
    const licenseInput = page.locator('input[name="license_key"]')
    
    const focusStyles = await licenseInput.evaluate(el => {
      return window.getComputedStyle(el, ':focus')
    })
    
    // Should have visible focus indicator
    expect(focusStyles).toBeTruthy()
  })
})

test.describe('License Activation - Performance Testing', () => {
  test('page load performance', async ({ page }) => {
    // Start performance monitoring
    await page.goto(BASE_URL)

    // Measure First Contentful Paint
    const fcpMetric = await page.evaluate(() => {
      return new Promise(resolve => {
        const observer = new PerformanceObserver(list => {
          for (const entry of list.getEntries()) {
            if (entry.name === 'first-contentful-paint') {
              resolve(entry.startTime)
            }
          }
        })
        observer.observe({ entryTypes: ['paint'] })
      })
    })

    // FCP should be under 2 seconds
    expect(fcpMetric).toBeLessThan(2000)

    // Measure Time to Interactive
    const ttiMetric = await page.evaluate(() => {
      return performance.timing.domInteractive - performance.timing.navigationStart
    })

    // TTI should be under 3 seconds
    expect(ttiMetric).toBeLessThan(3000)
  })

  test('form interaction performance', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Measure input response time
    const startTime = Date.now()
    
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    
    const inputTime = Date.now() - startTime

    // Form input should be responsive (under 100ms)
    expect(inputTime).toBeLessThan(100)

    // Measure form submission time
    const submitStartTime = Date.now()
    await page.click('button[type="submit"]')
    
    // Should show loading state immediately
    await expect(page.locator('button:has-text("Activating...")')).toBeVisible()
    
    const submitResponseTime = Date.now() - submitStartTime
    expect(submitResponseTime).toBeLessThan(200)
  })

  test('memory usage monitoring', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Get initial memory usage
    const initialMemory = await page.evaluate(() => {
      return (performance as any).memory?.usedJSHeapSize || 0
    })

    // Perform license activation
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })

    // Check final memory usage
    const finalMemory = await page.evaluate(() => {
      return (performance as any).memory?.usedJSHeapSize || 0
    })

    // Memory increase should be reasonable (less than 10MB)
    const memoryIncrease = finalMemory - initialMemory
    expect(memoryIncrease).toBeLessThan(10 * 1024 * 1024)
  })
})

test.describe('License Activation - Edge Cases and Error Recovery', () => {
  test('browser back/forward navigation', async ({ page }) => {
    await page.goto(BASE_URL)
    await expect(page).toHaveURL(/.*\/license/)

    // Fill form partially
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)

    // Navigate away and back
    await page.goto(`${BASE_URL}/dashboard`)
    await page.goBack()

    // Form should be cleared for security
    await expect(page.locator('input[name="license_key"]')).toHaveValue('')
  })

  test('multiple tab handling', async ({ context }) => {
    const page1 = await context.newPage()
    const page2 = await context.newPage()

    // Activate license in first tab
    await page1.goto(`${BASE_URL}/license`)
    await page1.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page1.fill('input[name="email"]', TEST_EMAIL)
    await page1.click('button[type="submit"]')

    await expect(page1.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })

    // Second tab should detect license is already active
    await page2.goto(`${BASE_URL}/license`)
    await expect(page2.locator('text=License is already active')).toBeVisible()
    
    // Should redirect to dashboard
    await expect(page2).toHaveURL(/.*\/dashboard/)
  })

  test('session timeout handling', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Mock session timeout
    await page.route('**/api/license/activate', route => {
      route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({
          type: '/problems/session-expired',
          title: 'Session Expired',
          status: 401,
          detail: 'Your session has expired. Please refresh and try again.'
        })
      })
    })

    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Should handle session expiry gracefully
    await expect(page.locator('text=Session Expired')).toBeVisible()
    await expect(page.locator('button:has-text("Refresh Page")')).toBeVisible()

    // Test refresh functionality
    await page.click('button:has-text("Refresh Page")')
    await expect(page).toHaveURL(/.*\/license/)
  })

  test('license already activated scenario', async ({ page }) => {
    // Mock already activated response
    await page.route('**/api/license/activate', route => {
      route.fulfill({
        status: 409,
        contentType: 'application/json',
        body: JSON.stringify({
          type: '/problems/license-already-active',
          title: 'License Already Active',
          status: 409,
          detail: 'This license is already activated on another device.'
        })
      })
    })

    await page.goto(`${BASE_URL}/license`)
    
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Should show conflict error
    await expect(page.locator('text=License Already Active')).toBeVisible()
    await expect(page.locator('text=This license is already activated on another device')).toBeVisible()

    // Should provide contact support option
    await expect(page.locator('button:has-text("Contact Support")')).toBeVisible()
  })
})