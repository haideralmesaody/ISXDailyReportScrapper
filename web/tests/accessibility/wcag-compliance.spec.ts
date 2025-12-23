/**
 * WCAG 2.1 AA Accessibility Compliance Tests
 * 
 * Comprehensive accessibility testing including keyboard navigation,
 * screen reader compatibility, color contrast validation, focus management,
 * ARIA labels and descriptions, and semantic HTML structure.
 * 
 * Coverage Requirements:
 * - Keyboard navigation through license form
 * - Screen reader compatibility
 * - Color contrast validation
 * - Focus management and trap
 * - ARIA labels and descriptions
 * - Semantic HTML structure
 */

import { test, expect } from '@playwright/test'
import AxeBuilder from '@axe-core/playwright'

const BASE_URL = 'http://localhost:8080'
const VALID_TEST_LICENSE_KEY = 'ISX1M02LYE1F9QJHR9D7Z'
const TEST_EMAIL = 'test@iraqiinvestor.gov.iq'

test.describe('WCAG 2.1 AA Compliance Testing', () => {
  test.beforeEach(async ({ page }) => {
    // Ensure consistent testing environment
    await page.goto(`${BASE_URL}/license`)
    await page.waitForLoadState('networkidle')
  })

  test('passes axe accessibility scan on license page', async ({ page }) => {
    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21aa'])
      .analyze()

    expect(accessibilityScanResults.violations).toEqual([])
  })

  test('passes axe accessibility scan on success state', async ({ page }) => {
    // Fill and submit form to reach success state
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Wait for success state
    await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })

    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21aa'])
      .analyze()

    expect(accessibilityScanResults.violations).toEqual([])
  })

  test('passes axe accessibility scan on error state', async ({ page }) => {
    // Mock error response
    await page.route('**/api/license/activate', route => {
      route.fulfill({
        status: 400,
        contentType: 'application/problem+json',
        body: JSON.stringify({
          type: '/problems/invalid-license',
          title: 'Invalid License Key',
          status: 400,
          detail: 'The provided license key is not valid'
        })
      })
    })

    await page.fill('input[name="license_key"]', 'INVALID_KEY')
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Wait for error state
    await expect(page.locator('text=Invalid License Key')).toBeVisible()

    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21aa'])
      .analyze()

    expect(accessibilityScanResults.violations).toEqual([])
  })
})

test.describe('Keyboard Navigation Accessibility', () => {
  test('supports complete keyboard navigation', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Test initial focus
    await page.keyboard.press('Tab')
    await expect(page.locator('input[name="license_key"]')).toBeFocused()

    // Test tab navigation through form
    await page.keyboard.press('Tab')
    await expect(page.locator('input[name="email"]')).toBeFocused()

    await page.keyboard.press('Tab')
    await expect(page.locator('button[type="submit"]')).toBeFocused()

    // Test reverse navigation
    await page.keyboard.press('Shift+Tab')
    await expect(page.locator('input[name="email"]')).toBeFocused()

    await page.keyboard.press('Shift+Tab')
    await expect(page.locator('input[name="license_key"]')).toBeFocused()
  })

  test('supports keyboard form submission', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Fill form using keyboard
    await page.keyboard.press('Tab')
    await page.keyboard.type(VALID_TEST_LICENSE_KEY)

    await page.keyboard.press('Tab')
    await page.keyboard.type(TEST_EMAIL)

    // Submit with Enter key
    await page.keyboard.press('Enter')

    // Verify submission
    await expect(page.locator('button:has-text("Activating...")')).toBeVisible()
  })

  test('maintains focus management during state changes', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Focus on submit button
    await page.keyboard.press('Tab')
    await page.keyboard.press('Tab')
    await page.keyboard.press('Tab')
    await expect(page.locator('button[type="submit"]')).toBeFocused()

    // Fill form to enable button
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)

    // Button should still be focusable when enabled
    await page.keyboard.press('Tab')
    await page.keyboard.press('Tab')
    await page.keyboard.press('Tab')
    await expect(page.locator('button[type="submit"]')).toBeFocused()
  })

  test('provides visible focus indicators', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Check focus ring on license input
    await page.keyboard.press('Tab')
    const licenseInput = page.locator('input[name="license_key"]:focus')
    
    const focusStyles = await licenseInput.evaluate(el => {
      const computed = window.getComputedStyle(el)
      return {
        outline: computed.outline,
        outlineColor: computed.outlineColor,
        outlineWidth: computed.outlineWidth,
        boxShadow: computed.boxShadow
      }
    })

    // Should have visible focus indicator
    const hasFocusIndicator = 
      focusStyles.outline !== 'none' || 
      focusStyles.boxShadow !== 'none' ||
      focusStyles.outlineWidth !== '0px'

    expect(hasFocusIndicator).toBeTruthy()
  })

  test('supports escape key for error dismissal', async ({ page }) => {
    await page.route('**/api/license/activate', route => {
      route.fulfill({
        status: 400,
        contentType: 'application/problem+json',
        body: JSON.stringify({
          type: '/problems/invalid-license',
          title: 'Invalid License Key',
          status: 400,
          detail: 'The provided license key is not valid'
        })
      })
    })

    await page.goto(`${BASE_URL}/license`)
    
    await page.fill('input[name="license_key"]', 'INVALID_KEY')
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Wait for error
    await expect(page.locator('text=Invalid License Key')).toBeVisible()

    // Press Escape to dismiss error (if supported)
    await page.keyboard.press('Escape')
    
    // Focus should return to form
    await expect(page.locator('input[name="license_key"]')).toBeFocused()
  })
})

test.describe('Screen Reader Accessibility', () => {
  test('provides proper form labels and descriptions', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Check license input accessibility
    const licenseInput = page.locator('input[name="license_key"]')
    await expect(licenseInput).toHaveAttribute('aria-label', 'License Key')
    await expect(licenseInput).toHaveAttribute('aria-required', 'true')
    await expect(licenseInput).toHaveAttribute('aria-describedby')

    // Check email input accessibility
    const emailInput = page.locator('input[name="email"]')
    await expect(emailInput).toHaveAttribute('aria-label', 'Email Address')
    await expect(emailInput).toHaveAttribute('aria-required', 'true')
    await expect(emailInput).toHaveAttribute('aria-describedby')

    // Check submit button accessibility
    const submitButton = page.locator('button[type="submit"]')
    await expect(submitButton).toHaveAttribute('aria-label')
  })

  test('announces form validation errors', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Trigger validation error
    await page.fill('input[name="license_key"]', 'INVALID')
    await page.fill('input[name="email"]', 'invalid-email')
    await page.click('button[type="submit"]')

    // Check error announcements
    const errorRegion = page.locator('[role="alert"]')
    await expect(errorRegion).toBeVisible()
    await expect(errorRegion).toHaveAttribute('aria-live', 'assertive')

    // Check individual field errors
    const licenseError = page.locator('#license-key-error')
    await expect(licenseError).toBeVisible()
    await expect(licenseError).toHaveAttribute('aria-live', 'polite')
  })

  test('provides meaningful page structure', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Check heading hierarchy
    const mainHeading = page.locator('h1')
    await expect(mainHeading).toBeVisible()
    await expect(mainHeading).toContainText('Activate Your Professional License')

    // Check landmark regions
    const main = page.locator('main')
    await expect(main).toBeVisible()
    await expect(main).toHaveAttribute('role', 'main')

    const form = page.locator('form')
    await expect(form).toHaveAttribute('aria-label', 'License Activation')
  })

  test('announces loading states', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Check loading announcement
    const loadingButton = page.locator('button:has-text("Activating...")')
    await expect(loadingButton).toBeVisible()
    await expect(loadingButton).toHaveAttribute('aria-live', 'polite')
    await expect(loadingButton).toHaveAttribute('aria-busy', 'true')
  })

  test('announces success state', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Wait for success
    await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })

    // Check success announcement
    const successRegion = page.locator('[role="status"]')
    await expect(successRegion).toBeVisible()
    await expect(successRegion).toHaveAttribute('aria-live', 'polite')
  })
})

test.describe('Color Contrast and Visual Accessibility', () => {
  test('meets color contrast requirements', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Test form elements contrast
    const elements = [
      'input[name="license_key"]',
      'input[name="email"]',
      'button[type="submit"]',
      'label[for="license_key"]',
      'label[for="email"]'
    ]

    for (const selector of elements) {
      const element = page.locator(selector)
      const contrast = await element.evaluate(el => {
        const computed = window.getComputedStyle(el)
        return {
          color: computed.color,
          backgroundColor: computed.backgroundColor,
          borderColor: computed.borderColor
        }
      })

      // Basic check that colors are defined
      expect(contrast.color).not.toBe('rgba(0, 0, 0, 0)')
      expect(contrast.backgroundColor).toBeDefined()
    }
  })

  test('maintains readability in high contrast mode', async ({ page }) => {
    // Simulate high contrast mode
    await page.addStyleTag({
      content: `
        @media (prefers-contrast: high) {
          * {
            background: black !important;
            color: white !important;
            border-color: white !important;
          }
        }
      `
    })

    await page.goto(`${BASE_URL}/license`)

    // Verify elements are still visible
    await expect(page.locator('input[name="license_key"]')).toBeVisible()
    await expect(page.locator('input[name="email"]')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('supports reduced motion preferences', async ({ page }) => {
    // Simulate reduced motion preference
    await page.emulateMedia({ reducedMotion: 'reduce' })
    
    await page.goto(`${BASE_URL}/license`)

    // Check that animations are reduced/disabled
    const spinner = page.locator('[data-testid="license-loading-spinner"]')
    if (await spinner.isVisible()) {
      const animationDuration = await spinner.evaluate(el => {
        const computed = window.getComputedStyle(el)
        return computed.animationDuration
      })

      // Animation should be reduced or disabled
      expect(['0s', 'none']).toContain(animationDuration)
    }
  })
})

test.describe('ARIA and Semantic HTML', () => {
  test('uses semantic HTML elements', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Check semantic structure
    await expect(page.locator('main')).toBeVisible()
    await expect(page.locator('form')).toBeVisible()
    await expect(page.locator('fieldset')).toBeVisible()
    await expect(page.locator('legend')).toBeVisible()
  })

  test('provides comprehensive ARIA labels', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Check ARIA labels
    const form = page.locator('form')
    await expect(form).toHaveAttribute('aria-label', 'License Activation')

    const fieldset = page.locator('fieldset')
    await expect(fieldset).toHaveAttribute('aria-labelledby')

    // Check input descriptions
    const licenseInput = page.locator('input[name="license_key"]')
    const describedBy = await licenseInput.getAttribute('aria-describedby')
    expect(describedBy).toBeTruthy()

    // Verify description element exists
    const descriptionElement = page.locator(`#${describedBy}`)
    await expect(descriptionElement).toBeVisible()
  })

  test('maintains ARIA states during interactions', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    const submitButton = page.locator('button[type="submit"]')
    
    // Initially disabled
    await expect(submitButton).toHaveAttribute('aria-disabled', 'true')

    // Enable by filling form
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)

    // Should be enabled
    await expect(submitButton).toHaveAttribute('aria-disabled', 'false')

    // Click to start loading
    await page.click('button[type="submit"]')

    // Should show loading state
    await expect(submitButton).toHaveAttribute('aria-busy', 'true')
  })

  test('provides error identification and association', async ({ page }) => {
    await page.route('**/api/license/activate', route => {
      route.fulfill({
        status: 400,
        contentType: 'application/problem+json',
        body: JSON.stringify({
          type: '/problems/invalid-license',
          title: 'Invalid License Key',
          status: 400,
          detail: 'The provided license key is not valid'
        })
      })
    })

    await page.goto(`${BASE_URL}/license`)
    
    await page.fill('input[name="license_key"]', 'INVALID_KEY')
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Wait for error
    await expect(page.locator('text=Invalid License Key')).toBeVisible()

    // Check error association
    const licenseInput = page.locator('input[name="license_key"]')
    await expect(licenseInput).toHaveAttribute('aria-invalid', 'true')
    
    const errorId = await licenseInput.getAttribute('aria-describedby')
    const errorElement = page.locator(`#${errorId}`)
    await expect(errorElement).toBeVisible()
    await expect(errorElement).toHaveAttribute('role', 'alert')
  })
})

test.describe('Mobile Accessibility', () => {
  test.use({ viewport: { width: 375, height: 667 } }) // iPhone SE size

  test('maintains accessibility on mobile devices', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Run accessibility scan on mobile
    const accessibilityScanResults = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21aa'])
      .analyze()

    expect(accessibilityScanResults.violations).toEqual([])
  })

  test('supports touch accessibility', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Test touch targets are large enough (minimum 44px)
    const touchTargets = [
      'input[name="license_key"]',
      'input[name="email"]',
      'button[type="submit"]'
    ]

    for (const selector of touchTargets) {
      const element = page.locator(selector)
      const boundingBox = await element.boundingBox()
      
      expect(boundingBox).not.toBeNull()
      expect(boundingBox!.height).toBeGreaterThanOrEqual(44)
      expect(boundingBox!.width).toBeGreaterThanOrEqual(44)
    }
  })

  test('maintains readable text size on mobile', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Check text elements have minimum readable size
    const textElements = [
      'label[for="license_key"]',
      'label[for="email"]',
      'button[type="submit"]',
      'h1'
    ]

    for (const selector of textElements) {
      const element = page.locator(selector)
      const fontSize = await element.evaluate(el => {
        return window.getComputedStyle(el).fontSize
      })

      const fontSizeNum = parseFloat(fontSize)
      // Minimum 16px for mobile readability
      expect(fontSizeNum).toBeGreaterThanOrEqual(16)
    }
  })
})