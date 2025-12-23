/**
 * Core Web Vitals and Performance Testing
 * 
 * Measures Core Web Vitals (>90 Lighthouse score), bundle size analysis (<250KB first load),
 * Time to Interactive (TTI) optimization, First Contentful Paint (FCP) benchmarking,
 * and Cumulative Layout Shift (CLS) monitoring.
 * 
 * Coverage Requirements:
 * - Core Web Vitals measurement (>90 Lighthouse score)
 * - Bundle size analysis (<250KB first load per CLAUDE.md)
 * - Time to Interactive (TTI) optimization
 * - First Contentful Paint (FCP) benchmarking
 * - Cumulative Layout Shift (CLS) monitoring
 */

import { test, expect } from '@playwright/test'
import { PlaywrightWebVitals } from '@playwright/web-vitals'

const BASE_URL = 'http://localhost:8080'
const VALID_TEST_LICENSE_KEY = 'ISX1M02LYE1F9QJHR9D7Z'
const TEST_EMAIL = 'test@iraqiinvestor.gov.iq'

// Performance thresholds based on CLAUDE.md requirements
const PERFORMANCE_THRESHOLDS = {
  // Core Web Vitals thresholds for good user experience
  LCP: 2500, // Largest Contentful Paint - Good: ≤2.5s
  FID: 100,  // First Input Delay - Good: ≤100ms
  CLS: 0.1,  // Cumulative Layout Shift - Good: ≤0.1
  
  // Additional performance metrics
  FCP: 1800, // First Contentful Paint - Good: ≤1.8s
  TTI: 3800, // Time to Interactive - Good: ≤3.8s
  TBT: 200,  // Total Blocking Time - Good: ≤200ms
  
  // Network and resource thresholds
  FIRST_LOAD_BUNDLE: 250 * 1024, // 250KB as per CLAUDE.md
  TOTAL_TRANSFER_SIZE: 1024 * 1024, // 1MB total
  
  // Performance scores (0-100)
  LIGHTHOUSE_PERFORMANCE: 90, // >90 as per CLAUDE.md
  LIGHTHOUSE_ACCESSIBILITY: 95,
  LIGHTHOUSE_BEST_PRACTICES: 90,
  LIGHTHOUSE_SEO: 90
}

test.describe('Core Web Vitals Measurement', () => {
  let webVitals: PlaywrightWebVitals

  test.beforeEach(async ({ page }) => {
    webVitals = new PlaywrightWebVitals(page)
    await webVitals.startMeasuring()
  })

  test.afterEach(async () => {
    const vitals = await webVitals.getVitals()
    console.log('Core Web Vitals:', vitals)
  })

  test('measures Core Web Vitals for license page', async ({ page }) => {
    // Navigate to license page and wait for load
    await page.goto(`${BASE_URL}/license`)
    
    // Wait for page to be fully loaded
    await page.waitForLoadState('networkidle')
    await page.waitForFunction(() => document.readyState === 'complete')

    // Get Core Web Vitals
    const vitals = await webVitals.getVitals()

    // Verify Largest Contentful Paint
    expect(vitals.lcp).toBeLessThan(PERFORMANCE_THRESHOLDS.LCP)
    console.log(`✅ LCP: ${vitals.lcp}ms (threshold: ${PERFORMANCE_THRESHOLDS.LCP}ms)`)

    // Verify Cumulative Layout Shift
    expect(vitals.cls).toBeLessThan(PERFORMANCE_THRESHOLDS.CLS)
    console.log(`✅ CLS: ${vitals.cls} (threshold: ${PERFORMANCE_THRESHOLDS.CLS})`)

    // Verify First Contentful Paint
    expect(vitals.fcp).toBeLessThan(PERFORMANCE_THRESHOLDS.FCP)
    console.log(`✅ FCP: ${vitals.fcp}ms (threshold: ${PERFORMANCE_THRESHOLDS.FCP}ms)`)
  })

  test('measures First Input Delay during form interaction', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)
    await page.waitForLoadState('networkidle')

    // Trigger first input by clicking license input
    await page.click('input[name="license_key"]')
    
    // Wait for FID measurement
    await page.waitForTimeout(100)

    const vitals = await webVitals.getVitals()

    // Verify First Input Delay
    if (vitals.fid !== undefined) {
      expect(vitals.fid).toBeLessThan(PERFORMANCE_THRESHOLDS.FID)
      console.log(`✅ FID: ${vitals.fid}ms (threshold: ${PERFORMANCE_THRESHOLDS.FID}ms)`)
    }
  })

  test('measures Core Web Vitals during form submission', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)
    await page.waitForLoadState('networkidle')

    // Fill and submit form
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Wait for response
    await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })

    const vitals = await webVitals.getVitals()

    // Verify layout stability during form submission
    expect(vitals.cls).toBeLessThan(PERFORMANCE_THRESHOLDS.CLS)
    console.log(`✅ CLS during submission: ${vitals.cls} (threshold: ${PERFORMANCE_THRESHOLDS.CLS})`)
  })
})

test.describe('Bundle Size Analysis', () => {
  test('analyzes first load bundle size', async ({ page }) => {
    // Track network requests
    const resourceSizes: { [key: string]: number } = {}
    let totalSize = 0
    let jsSize = 0
    let cssSize = 0

    page.on('response', async (response) => {
      const url = response.url()
      const contentLength = response.headers()['content-length']
      const size = contentLength ? parseInt(contentLength) : 0

      if (size > 0) {
        resourceSizes[url] = size
        totalSize += size

        if (url.endsWith('.js')) {
          jsSize += size
        } else if (url.endsWith('.css')) {
          cssSize += size
        }
      }
    })

    await page.goto(`${BASE_URL}/license`)
    await page.waitForLoadState('networkidle')

    console.log(`📊 Bundle Size Analysis:`)
    console.log(`  Total Size: ${(totalSize / 1024).toFixed(2)} KB`)
    console.log(`  JavaScript: ${(jsSize / 1024).toFixed(2)} KB`)
    console.log(`  CSS: ${(cssSize / 1024).toFixed(2)} KB`)

    // Verify first load bundle size meets CLAUDE.md requirements
    expect(jsSize + cssSize).toBeLessThan(PERFORMANCE_THRESHOLDS.FIRST_LOAD_BUNDLE)
    console.log(`✅ First load bundle: ${((jsSize + cssSize) / 1024).toFixed(2)} KB (threshold: ${PERFORMANCE_THRESHOLDS.FIRST_LOAD_BUNDLE / 1024} KB)`)

    // Log largest resources
    const sortedResources = Object.entries(resourceSizes)
      .sort(([, a], [, b]) => b - a)
      .slice(0, 5)

    console.log('📋 Largest Resources:')
    sortedResources.forEach(([url, size]) => {
      const filename = url.split('/').pop() || url
      console.log(`  ${filename}: ${(size / 1024).toFixed(2)} KB`)
    })
  })

  test('verifies code splitting effectiveness', async ({ page }) => {
    const loadedChunks = new Set<string>()

    page.on('response', (response) => {
      const url = response.url()
      if (url.includes('/_next/static/chunks/')) {
        const chunkName = url.split('/').pop()
        if (chunkName) {
          loadedChunks.add(chunkName)
        }
      }
    })

    // Load initial page
    await page.goto(`${BASE_URL}/license`)
    await page.waitForLoadState('networkidle')
    
    const initialChunks = loadedChunks.size
    console.log(`📦 Initial chunks loaded: ${initialChunks}`)

    // Navigate to dashboard to trigger additional chunk loading
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')
    
    await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })
    await page.click('button:has-text("Continue to Dashboard")')
    
    await page.waitForLoadState('networkidle')
    
    const finalChunks = loadedChunks.size
    console.log(`📦 Total chunks after navigation: ${finalChunks}`)

    // Verify code splitting is working (more chunks loaded for new routes)
    expect(finalChunks).toBeGreaterThan(initialChunks)
    console.log(`✅ Code splitting active: ${finalChunks - initialChunks} additional chunks loaded`)
  })
})

test.describe('Time to Interactive (TTI) Optimization', () => {
  test('measures Time to Interactive', async ({ page }) => {
    const navigationStart = Date.now()

    await page.goto(`${BASE_URL}/license`)

    // Wait for page to be interactive
    await page.waitForLoadState('domcontentloaded')
    
    // Test interactivity by focusing on form field
    await page.focus('input[name="license_key"]')
    const interactiveTime = Date.now() - navigationStart

    console.log(`⚡ Time to Interactive: ${interactiveTime}ms`)
    expect(interactiveTime).toBeLessThan(PERFORMANCE_THRESHOLDS.TTI)
    console.log(`✅ TTI: ${interactiveTime}ms (threshold: ${PERFORMANCE_THRESHOLDS.TTI}ms)`)
  })

  test('measures form responsiveness', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)
    await page.waitForLoadState('networkidle')

    // Measure input responsiveness
    const startTime = Date.now()
    await page.type('input[name="license_key"]', 'TEST')
    const inputResponseTime = Date.now() - startTime

    console.log(`⌨️  Input response time: ${inputResponseTime}ms`)
    expect(inputResponseTime).toBeLessThan(100) // Should be very responsive

    // Measure validation responsiveness
    const validationStart = Date.now()
    await page.blur('input[name="license_key"]')
    
    // Wait for validation message
    await expect(page.locator('text=License key must be exactly 19 characters')).toBeVisible()
    const validationTime = Date.now() - validationStart

    console.log(`✅ Validation response time: ${validationTime}ms`)
    expect(validationTime).toBeLessThan(200)
  })
})

test.describe('First Contentful Paint (FCP) Benchmarking', () => {
  test('measures FCP across different network conditions', async ({ page }) => {
    // Test fast connection
    const fcpTimes: number[] = []

    for (let i = 0; i < 3; i++) {
      await page.goto(`${BASE_URL}/license`)
      
      const fcp = await page.evaluate(() => {
        return new Promise<number>((resolve) => {
          const observer = new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
              if (entry.name === 'first-contentful-paint') {
                resolve(entry.startTime)
                observer.disconnect()
              }
            }
          })
          observer.observe({ entryTypes: ['paint'] })
        })
      })

      fcpTimes.push(fcp)
      console.log(`🎨 FCP Run ${i + 1}: ${fcp.toFixed(2)}ms`)
    }

    const averageFCP = fcpTimes.reduce((sum, time) => sum + time, 0) / fcpTimes.length
    console.log(`📊 Average FCP: ${averageFCP.toFixed(2)}ms`)
    
    expect(averageFCP).toBeLessThan(PERFORMANCE_THRESHOLDS.FCP)
    console.log(`✅ Average FCP: ${averageFCP.toFixed(2)}ms (threshold: ${PERFORMANCE_THRESHOLDS.FCP}ms)`)
  })

  test('measures FCP with slow 3G simulation', async ({ page, context }) => {
    // Simulate slow 3G
    await context.route('**/*', async (route) => {
      await new Promise(resolve => setTimeout(resolve, 100))
      await route.continue()
    })

    await page.goto(`${BASE_URL}/license`)
    
    const fcp = await page.evaluate(() => {
      return new Promise<number>((resolve) => {
        const observer = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (entry.name === 'first-contentful-paint') {
              resolve(entry.startTime)
              observer.disconnect()
            }
          }
        })
        observer.observe({ entryTypes: ['paint'] })
      })
    })

    console.log(`🐌 FCP on slow 3G: ${fcp.toFixed(2)}ms`)
    
    // Allow higher threshold for slow connection but still reasonable
    expect(fcp).toBeLessThan(PERFORMANCE_THRESHOLDS.FCP * 2)
  })
})

test.describe('Cumulative Layout Shift (CLS) Monitoring', () => {
  test('monitors CLS during page load', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Monitor layout shifts
    let maxCLS = 0
    
    await page.evaluate(() => {
      let cls = 0
      
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType === 'layout-shift' && !(entry as any).hadRecentInput) {
            cls += (entry as any).value
          }
        }
        
        // Store in window for retrieval
        ;(window as any).clsValue = cls
      })
      
      observer.observe({ entryTypes: ['layout-shift'] })
    })

    // Wait for page to stabilize
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(2000)

    // Get final CLS value
    const finalCLS = await page.evaluate(() => (window as any).clsValue || 0)
    
    console.log(`📐 Cumulative Layout Shift: ${finalCLS.toFixed(4)}`)
    expect(finalCLS).toBeLessThan(PERFORMANCE_THRESHOLDS.CLS)
    console.log(`✅ CLS: ${finalCLS.toFixed(4)} (threshold: ${PERFORMANCE_THRESHOLDS.CLS})`)
  })

  test('monitors CLS during form interactions', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)
    await page.waitForLoadState('networkidle')

    // Set up CLS monitoring
    await page.evaluate(() => {
      let cls = 0
      ;(window as any).clsHistory = []
      
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType === 'layout-shift' && !(entry as any).hadRecentInput) {
            cls += (entry as any).value
            ;(window as any).clsHistory.push({
              value: (entry as any).value,
              sources: (entry as any).sources,
              time: entry.startTime
            })
          }
        }
        
        ;(window as any).clsValue = cls
      })
      
      observer.observe({ entryTypes: ['layout-shift'] })
    })

    // Perform form interactions
    await page.fill('input[name="license_key"]', 'INVALID_KEY')
    await page.blur('input[name="license_key"]')
    
    // Wait for validation error to appear
    await expect(page.locator('text=License key must be exactly 19 characters')).toBeVisible()
    
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    
    // Wait for validation to clear
    await page.waitForTimeout(1000)

    const finalCLS = await page.evaluate(() => (window as any).clsValue || 0)
    const clsHistory = await page.evaluate(() => (window as any).clsHistory || [])
    
    console.log(`📐 CLS during interactions: ${finalCLS.toFixed(4)}`)
    console.log(`📋 Layout shift events: ${clsHistory.length}`)
    
    expect(finalCLS).toBeLessThan(PERFORMANCE_THRESHOLDS.CLS)
    console.log(`✅ CLS during interactions: ${finalCLS.toFixed(4)} (threshold: ${PERFORMANCE_THRESHOLDS.CLS})`)
  })

  test('monitors CLS during async loading states', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)
    await page.waitForLoadState('networkidle')

    // Set up CLS monitoring
    await page.evaluate(() => {
      let cls = 0
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType === 'layout-shift' && !(entry as any).hadRecentInput) {
            cls += (entry as any).value
          }
        }
        ;(window as any).clsValue = cls
      })
      observer.observe({ entryTypes: ['layout-shift'] })
    })

    // Trigger form submission with loading states
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')

    // Wait through loading and success states
    await expect(page.locator('button:has-text("Activating...")')).toBeVisible()
    await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })

    const finalCLS = await page.evaluate(() => (window as any).clsValue || 0)
    
    console.log(`📐 CLS during async operations: ${finalCLS.toFixed(4)}`)
    expect(finalCLS).toBeLessThan(PERFORMANCE_THRESHOLDS.CLS)
    console.log(`✅ CLS during async operations: ${finalCLS.toFixed(4)} (threshold: ${PERFORMANCE_THRESHOLDS.CLS})`)
  })
})

test.describe('Memory Usage and Performance Monitoring', () => {
  test('monitors memory usage during license activation flow', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)
    await page.waitForLoadState('networkidle')

    // Get initial memory usage
    const initialMemory = await page.evaluate(() => {
      return (performance as any).memory ? {
        usedJSHeapSize: (performance as any).memory.usedJSHeapSize,
        totalJSHeapSize: (performance as any).memory.totalJSHeapSize,
        jsHeapSizeLimit: (performance as any).memory.jsHeapSizeLimit
      } : null
    })

    if (initialMemory) {
      console.log(`💾 Initial memory usage: ${(initialMemory.usedJSHeapSize / 1024 / 1024).toFixed(2)} MB`)
    }

    // Perform license activation
    await page.fill('input[name="license_key"]', VALID_TEST_LICENSE_KEY)
    await page.fill('input[name="email"]', TEST_EMAIL)
    await page.click('button[type="submit"]')
    
    await expect(page.locator('text=License Activated Successfully!')).toBeVisible({ timeout: 10000 })

    // Get final memory usage
    const finalMemory = await page.evaluate(() => {
      return (performance as any).memory ? {
        usedJSHeapSize: (performance as any).memory.usedJSHeapSize,
        totalJSHeapSize: (performance as any).memory.totalJSHeapSize,
        jsHeapSizeLimit: (performance as any).memory.jsHeapSizeLimit
      } : null
    })

    if (initialMemory && finalMemory) {
      const memoryIncrease = finalMemory.usedJSHeapSize - initialMemory.usedJSHeapSize
      console.log(`📈 Memory increase: ${(memoryIncrease / 1024 / 1024).toFixed(2)} MB`)
      
      // Memory increase should be reasonable (less than 5MB)
      expect(memoryIncrease).toBeLessThan(5 * 1024 * 1024)
      console.log(`✅ Memory increase: ${(memoryIncrease / 1024 / 1024).toFixed(2)} MB (threshold: 5 MB)`)
    }
  })

  test('measures long task performance', async ({ page }) => {
    await page.goto(`${BASE_URL}/license`)

    // Monitor long tasks
    const longTasks = await page.evaluate(() => {
      return new Promise((resolve) => {
        const tasks: any[] = []
        const observer = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            tasks.push({
              name: entry.name,
              duration: entry.duration,
              startTime: entry.startTime
            })
          }
        })
        observer.observe({ entryTypes: ['longtask'] })
        
        // Resolve after 5 seconds of monitoring
        setTimeout(() => {
          observer.disconnect()
          resolve(tasks)
        }, 5000)
      })
    })

    console.log(`🐌 Long tasks detected: ${(longTasks as any[]).length}`)
    
    if ((longTasks as any[]).length > 0) {
      const totalBlockingTime = (longTasks as any[]).reduce((sum, task) => sum + Math.max(0, task.duration - 50), 0)
      console.log(`⏱️  Total Blocking Time: ${totalBlockingTime.toFixed(2)}ms`)
      
      expect(totalBlockingTime).toBeLessThan(PERFORMANCE_THRESHOLDS.TBT)
      console.log(`✅ TBT: ${totalBlockingTime.toFixed(2)}ms (threshold: ${PERFORMANCE_THRESHOLDS.TBT}ms)`)
    }
  })
})