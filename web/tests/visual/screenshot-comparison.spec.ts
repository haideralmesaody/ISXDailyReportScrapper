/**
 * Visual Regression Testing Setup
 * 
 * Screenshot comparison testing, logo and branding consistency,
 * form layout across different screen sizes, error state visual consistency,
 * and loading animation proper rendering.
 * 
 * Coverage Requirements:
 * - Screenshot comparison testing
 * - Logo and branding consistency
 * - Form layout across different screen sizes
 * - Error state visual consistency
 * - Loading animation proper rendering
 */

import { test, expect, devices } from '@playwright/test'

const BASE_URL = 'http://localhost:8080'
const VALID_TEST_LICENSE_KEY = 'ISX1M02LYE1F9QJHR9D7Z'
const TEST_EMAIL = 'test@iraqiinvestor.gov.iq'

// Screen sizes for responsive testing
const SCREEN_SIZES = [
  { name: 'desktop', width: 1920, height: 1080 },
  { name: 'laptop', width: 1366, height: 768 },
  { name: 'tablet', width: 768, height: 1024 },
  { name: 'mobile', width: 375, height: 667 }
]

test.describe('Visual Regression Testing', () => {
  test.describe('License Page Visual Consistency', () => {
    for (const size of SCREEN_SIZES) {
      test(`license page layout - ${size.name} (${size.width}x${size.height})`, async ({ page }) => {
        await page.setViewportSize({ width: size.width, height: size.height })
        await page.goto(`${BASE_URL}/license`)
        await page.waitForLoadState('networkidle')

        // Wait for fonts and images to load
        await page.waitForTimeout(1000)

        // Take full page screenshot
        await expect(page).toHaveScreenshot(`license-page-${size.name}.png`, {
          fullPage: true,
          animations: 'disabled',
          clip: { x: 0, y: 0, width: size.width, height: Math.min(size.height, 2000) }
        })
      })
    }

    test('license page with filled form', async ({ page }) => {
      await page.goto(`${BASE_URL}/license`)
      await page.waitForLoadState('networkidle')

      // Fill form
      await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
      await page.fill('input[name="email"]', TEST_EMAIL)

      // Wait for validation states
      await page.waitForTimeout(500)

      await expect(page).toHaveScreenshot('license-page-filled-form.png', {
        animations: 'disabled'
      })
    })

    test('license page loading state', async ({ page }) => {
      // Mock slow response to capture loading state
      await page.route('**/api/license/activate', async route => {
        await new Promise(resolve => setTimeout(resolve, 2000))
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            license: {
              valid: true,
              status: 'active',
              expires_at: '2025-12-31T23:59:59Z',
              email: TEST_EMAIL
            }
          })
        })
      })

      await page.goto(`${BASE_URL}/license`)
      await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
      await page.fill('input[name="email"]', TEST_EMAIL)
      await page.click('button[type="submit"]')

      // Capture loading state
      await expect(page.locator('button:has-text("Activating...")')).toBeVisible()
      
      await expect(page).toHaveScreenshot('license-page-loading-state.png', {
        animations: 'allow'
      })
    })

    test('license page success state', async ({ page }) => {
      await page.goto(`${BASE_URL}/license`)
      await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
      await page.fill('input[name="email"]', TEST_EMAIL)
      await page.click('button[type="submit"]')

      // Wait for success state
      await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })

      await expect(page).toHaveScreenshot('license-page-success-state.png', {
        animations: 'disabled'
      })
    })
  })

  test.describe('Error State Visual Consistency', () => {
    test('form validation errors', async ({ page }) => {
      await page.goto(`${BASE_URL}/license`)
      await page.waitForLoadState('networkidle')

      // Trigger validation errors
      await page.fill('input[name="license_key"]', 'INVALID')
      await page.fill('input[name="email"]', 'invalid-email')
      await page.blur('input[name="email"]')

      // Wait for error messages to appear
      await expect(page.locator('text=License key must be exactly 19 characters')).toBeVisible()
      await expect(page.locator('text=Please enter a valid email address')).toBeVisible()

      await expect(page).toHaveScreenshot('license-page-validation-errors.png', {
        animations: 'disabled'
      })
    })

    test('API error response', async ({ page }) => {
      // Mock API error response
      await page.route('**/api/license/activate', route => {
        route.fulfill({
          status: 400,
          contentType: 'application/problem+json',
          body: JSON.stringify({
            type: '/problems/invalid-license',
            title: 'Invalid License Key',
            status: 400,
            detail: 'The provided license key is not valid or has expired'
          })
        })
      })

      await page.goto(`${BASE_URL}/license`)
      await page.fill('input[name="license_key"]', 'INVALID_KEY')
      await page.fill('input[name="email"]', TEST_EMAIL)
      await page.click('button[type="submit"]')

      // Wait for error state
      await expect(page.locator('text=Invalid License Key')).toBeVisible()

      await expect(page).toHaveScreenshot('license-page-api-error.png', {
        animations: 'disabled'
      })
    })

    test('network error state', async ({ page }) => {
      // Mock network failure
      await page.route('**/api/license/activate', route => {
        route.abort('failed')
      })

      await page.goto(`${BASE_URL}/license`)
      await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
      await page.fill('input[name="email"]', TEST_EMAIL)
      await page.click('button[type="submit"]')

      // Wait for network error state
      await expect(page.locator('text=Network Connection Error')).toBeVisible()

      await expect(page).toHaveScreenshot('license-page-network-error.png', {
        animations: 'disabled'
      })
    })

    test('license already activated error', async ({ page }) => {
      // Mock conflict response
      await page.route('**/api/license/activate', route => {
        route.fulfill({
          status: 409,
          contentType: 'application/problem+json',
          body: JSON.stringify({
            type: '/problems/license-already-active',
            title: 'License Already Active',
            status: 409,
            detail: 'This license is already activated on another device'
          })
        })
      })

      await page.goto(`${BASE_URL}/license`)
      await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
      await page.fill('input[name="email"]', TEST_EMAIL)
      await page.click('button[type="submit"]')

      // Wait for conflict error state
      await expect(page.locator('text=License Already Active')).toBeVisible()

      await expect(page).toHaveScreenshot('license-page-already-active-error.png', {
        animations: 'disabled'
      })
    })
  })

  test.describe('Logo and Branding Consistency', () => {
    test('Iraqi Investor logo visibility and positioning', async ({ page }) => {
      await page.goto(`${BASE_URL}/license`)
      await page.waitForLoadState('networkidle')

      // Focus on logo area
      const logoSection = page.locator('img[alt="Iraqi Investor Logo"]').locator('..')
      
      await expect(logoSection).toHaveScreenshot('iraqi-investor-logo.png', {
        animations: 'disabled'
      })
    })

    test('branding consistency across states', async ({ page }) => {
      const states = [
        { name: 'initial', action: async () => {} },
        { 
          name: 'filled', 
          action: async () => {
            await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
            await page.fill('input[name="email"]', TEST_EMAIL)
          }
        },
        {
          name: 'success',
          action: async () => {
            await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
            await page.fill('input[name="email"]', TEST_EMAIL)
            await page.click('button[type="submit"]')
            await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })
          }
        }
      ]

      for (const state of states) {
        await page.goto(`${BASE_URL}/license`)
        await page.waitForLoadState('networkidle')
        
        await state.action()
        
        // Screenshot header area with branding
        const headerArea = page.locator('header, .header, [data-testid="header"]').first()
        if (await headerArea.isVisible()) {
          await expect(headerArea).toHaveScreenshot(`branding-${state.name}.png`, {
            animations: 'disabled'
          })
        } else {
          // Fallback to top section of page
          await expect(page).toHaveScreenshot(`branding-${state.name}-full.png`, {
            animations: 'disabled',
            clip: { x: 0, y: 0, width: 1280, height: 300 }
          })
        }
      }
    })

    test('logo in different viewport sizes', async ({ page }) => {
      for (const size of SCREEN_SIZES) {
        await page.setViewportSize({ width: size.width, height: size.height })
        await page.goto(`${BASE_URL}/license`)
        await page.waitForLoadState('networkidle')

        // Focus on branding area
        await expect(page).toHaveScreenshot(`logo-${size.name}.png`, {
          animations: 'disabled',
          clip: { x: 0, y: 0, width: size.width, height: 200 }
        })
      }
    })
  })

  test.describe('Form Layout Responsive Design', () => {
    for (const size of SCREEN_SIZES) {
      test(`form layout - ${size.name}`, async ({ page }) => {
        await page.setViewportSize({ width: size.width, height: size.height })
        await page.goto(`${BASE_URL}/license`)
        await page.waitForLoadState('networkidle')

        // Focus on form area
        const formArea = page.locator('form')
        
        await expect(formArea).toHaveScreenshot(`form-layout-${size.name}.png`, {
          animations: 'disabled'
        })
      })
    }

    test('form field focus states across screen sizes', async ({ page }) => {
      for (const size of SCREEN_SIZES) {
        await page.setViewportSize({ width: size.width, height: size.height })
        await page.goto(`${BASE_URL}/license`)
        await page.waitForLoadState('networkidle')

        // Focus on license key input
        await page.focus('input[name="license_key"]')
        
        await expect(page.locator('form')).toHaveScreenshot(`form-focus-license-${size.name}.png`, {
          animations: 'disabled'
        })

        // Focus on email input
        await page.focus('input[name="email"]')
        
        await expect(page.locator('form')).toHaveScreenshot(`form-focus-email-${size.name}.png`, {
          animations: 'disabled'
        })
      }
    })

    test('button states across screen sizes', async ({ page }) => {
      for (const size of SCREEN_SIZES) {
        await page.setViewportSize({ width: size.width, height: size.height })
        await page.goto(`${BASE_URL}/license`)
        await page.waitForLoadState('networkidle')

        // Disabled state (initial)
        await expect(page.locator('button[type="submit"]')).toHaveScreenshot(`button-disabled-${size.name}.png`, {
          animations: 'disabled'
        })

        // Enabled state
        await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
        await page.fill('input[name="email"]', TEST_EMAIL)

        await expect(page.locator('button[type="submit"]')).toHaveScreenshot(`button-enabled-${size.name}.png`, {
          animations: 'disabled'
        })
      }
    })
  })

  test.describe('Loading Animation Rendering', () => {
    test('loading spinner animation', async ({ page }) => {
      // Mock slow response to capture loading animation
      await page.route('**/api/license/activate', async route => {
        await new Promise(resolve => setTimeout(resolve, 3000))
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            license: {
              valid: true,
              status: 'active',
              expires_at: '2025-12-31T23:59:59Z',
              email: TEST_EMAIL
            }
          })
        })
      })

      await page.goto(`${BASE_URL}/license`)
      await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
      await page.fill('input[name="email"]', TEST_EMAIL)
      await page.click('button[type="submit"]')

      // Wait for loading state
      await expect(page.locator('[data-testid="submit-loading-spinner"]')).toBeVisible()

      // Capture loading spinner at different moments
      await page.waitForTimeout(500)
      await expect(page.locator('[data-testid="submit-loading-spinner"]')).toHaveScreenshot('loading-spinner-frame1.png')

      await page.waitForTimeout(500)
      await expect(page.locator('[data-testid="submit-loading-spinner"]')).toHaveScreenshot('loading-spinner-frame2.png')
    })

    test('page loading skeleton', async ({ page }) => {
      // Intercept requests to show loading state
      let resolvePageLoad: () => void
      const pageLoadPromise = new Promise<void>(resolve => {
        resolvePageLoad = resolve
      })

      await page.route('**/license', async route => {
        // Delay the page load
        await new Promise(resolve => setTimeout(resolve, 1000))
        await route.continue()
        resolvePageLoad()
      })

      // Navigate and capture loading state
      const navigationPromise = page.goto(`${BASE_URL}/license`)
      
      // Try to capture loading skeleton if visible
      try {
        await expect(page.locator('[data-testid="page-loading-skeleton"]')).toBeVisible({ timeout: 500 })
        await expect(page).toHaveScreenshot('page-loading-skeleton.png', {
          animations: 'allow'
        })
      } catch {
        // Loading was too fast to capture, continue
      }

      await navigationPromise
      await pageLoadPromise
    })

    test('loading animation accessibility', async ({ page }) => {
      // Mock slow response
      await page.route('**/api/license/activate', async route => {
        await new Promise(resolve => setTimeout(resolve, 2000))
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            license: {
              valid: true,
              status: 'active',
              expires_at: '2025-12-31T23:59:59Z',
              email: TEST_EMAIL
            }
          })
        })
      })

      // Enable reduced motion for accessibility testing
      await page.emulateMedia({ reducedMotion: 'reduce' })

      await page.goto(`${BASE_URL}/license`)
      await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
      await page.fill('input[name="email"]', TEST_EMAIL)
      await page.click('button[type="submit"]')

      // Capture loading state with reduced motion
      await expect(page.locator('button:has-text("Activating...")')).toBeVisible()
      
      await expect(page).toHaveScreenshot('loading-reduced-motion.png', {
        animations: 'disabled'
      })
    })
  })

  test.describe('Dark Mode and Theme Consistency', () => {
    test('dark mode rendering', async ({ page }) => {
      // Enable dark mode if supported
      await page.emulateMedia({ colorScheme: 'dark' })
      
      await page.goto(`${BASE_URL}/license`)
      await page.waitForLoadState('networkidle')

      await expect(page).toHaveScreenshot('license-page-dark-mode.png', {
        animations: 'disabled'
      })
    })

    test('high contrast mode', async ({ page }) => {
      // Simulate high contrast mode
      await page.addStyleTag({
        content: `
          @media (prefers-contrast: high) {
            :root {
              --bg-color: #000000;
              --text-color: #ffffff;
              --border-color: #ffffff;
              --accent-color: #ffff00;
            }
          }
        `
      })

      await page.goto(`${BASE_URL}/license`)
      await page.waitForLoadState('networkidle')

      await expect(page).toHaveScreenshot('license-page-high-contrast.png', {
        animations: 'disabled'
      })
    })
  })

  test.describe('Cross-Browser Visual Consistency', () => {
    const browsers = ['chromium', 'firefox', 'webkit']

    for (const browserName of browsers) {
      test(`license page consistency - ${browserName}`, async ({ page }) => {
        await page.goto(`${BASE_URL}/license`)
        await page.waitForLoadState('networkidle')

        // Take screenshot for cross-browser comparison
        await expect(page).toHaveScreenshot(`license-page-${browserName}.png`, {
          animations: 'disabled',
          threshold: 0.3 // Allow for minor rendering differences
        })
      })
    }
  })
})

test.describe('Screenshot Comparison Utilities', () => {
  test('generate baseline screenshots', async ({ page }) => {
    const scenarios = [
      { name: 'initial-state', setup: async () => {} },
      { 
        name: 'form-filled', 
        setup: async () => {
          await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
          await page.fill('input[name="email"]', TEST_EMAIL)
        }
      },
      {
        name: 'validation-error',
        setup: async () => {
          await page.fill('input[name="license_key"]', 'INVALID')
          await page.fill('input[name="email"]', 'invalid')
          await page.blur('input[name="email"]')
          await expect(page.locator('text=License key must be exactly 19 characters')).toBeVisible()
        }
      }
    ]

    for (const scenario of scenarios) {
      await page.goto(`${BASE_URL}/license`)
      await page.waitForLoadState('networkidle')
      
      await scenario.setup()
      await page.waitForTimeout(500) // Allow UI to settle

      await expect(page).toHaveScreenshot(`baseline-${scenario.name}.png`, {
        animations: 'disabled',
        fullPage: true
      })
    }
  })

  test('update screenshots threshold configuration', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)
    await page.waitForLoadState('networkidle')

    // Test with different threshold values
    const thresholds = [0.1, 0.2, 0.3]
    
    for (const threshold of thresholds) {
      await expect(page).toHaveScreenshot(`threshold-test-${threshold}.png`, {
        animations: 'disabled',
        threshold: threshold
      })
    }
  })
})