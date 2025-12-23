/**
 * Pipeline Tickets Visual Regression Tests
 *
 * Visual regression tests to ensure pipeline ticket rendering consistency
 * and proper display of stage information, progress, and connections.
 */

import { test, expect } from '@playwright/test'

test.describe('Pipeline Tickets Visual Regression', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API responses for operations data
    await page.route('/api/operations/types', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { id: 'scraping', name: 'Data Collection', description: 'Download ISX daily reports' },
          { id: 'processing', name: 'Excel to CSV Processing', description: 'Convert Excel files to CSV format' },
          { id: 'indices', name: 'Index Extraction', description: 'Extract ISX60 and ISX15 indices' },
          { id: 'liquidity', name: 'Liquidity Analysis', description: 'Calculate liquidity metrics' },
          { id: 'full_pipeline', name: 'Full Pipeline', description: 'Run all stages in sequence' }
        ])
      })
    })

    // Mock WebSocket connection
    await page.route('/ws', async (route) => {
      // Keep the connection open for WebSocket
      await route.fulfill({
        status: 101,
        headers: {
          'Upgrade': 'websocket',
          'Connection': 'Upgrade',
          'Sec-WebSocket-Accept': 'mock-websocket-accept'
        }
      })
    })
  })

  test('pipeline stages display in correct order with connectors', async ({ page }) => {
    // Mock active pipeline operation with multiple stages
    await page.addInitScript(() => {
      window.mockOperations = [
        {
          operation_id: 'pipeline-visual-test',
          status: 'running',
          progress: 35,
          steps: [
            {
              id: 'scraping-step',
              name: 'Data Collection',
              status: 'completed',
              progress: 100,
              metadata: {
                stage_id: 'scraping',
                files_downloaded: 15,
                duration: '3m 15s'
              }
            },
            {
              id: 'processing-step',
              name: 'Excel to CSV Processing',
              status: 'active',
              progress: 45,
              metadata: {
                stage_id: 'processing',
                files_processed: 7,
                total_files: 15,
                current_file: 'market_data.xlsx'
              }
            },
            {
              id: 'indices-step',
              name: 'Index Extraction',
              status: 'pending',
              progress: 0,
              metadata: {
                stage_id: 'indices'
              }
            },
            {
              id: 'liquidity-step',
              name: 'Liquidity Analysis',
              status: 'pending',
              progress: 0,
              metadata: {
                stage_id: 'liquidity'
              }
            }
          ],
          metadata: {
            pipeline_summary: {
              stage_timeline: {
                scraping: { started_at: '2025-10-17T17:00:00Z', completed_at: '2025-10-17T17:03:15Z' },
                processing: { started_at: '2025-10-17T17:03:20Z' }
              }
            }
          }
        }
      ]
    })

    await page.goto('/operations')
    await page.waitForLoadState('networkidle')

    // Wait for operations to be displayed
    await page.waitForSelector('[data-testid="operation-ticket"]')

    // Take screenshot for visual comparison
    await expect(page.locator('body')).toHaveScreenshot('pipeline-stages-with-connectors.png', {
      fullPage: true,
      animations: 'disabled'
    })

    // Verify specific elements are visible
    await expect(page.locator('text=Data Collection')).toBeVisible()
    await expect(page.locator('text=Stage 1 of 4')).toBeVisible()
    await expect(page.locator('text=Excel to CSV Processing')).toBeVisible()
    await expect(page.locator('text=Stage 2 of 4')).toBeVisible()

    // Verify connectors are visible between stages
    await expect(page.locator('[role="separator"]')).toHaveCount(3) // 3 connectors for 4 stages

    // Verify status indicators
    await expect(page.locator('text=Completed')).toBeVisible()
    await expect(page.locator('text=Active')).toBeVisible()
    await expect(page.locator('text=Pending')).toHaveCount(2)
  })

  test('completed pipeline shows all stages with proper completion states', async ({ page }) => {
    // Mock completed pipeline operation
    await page.addInitScript(() => {
      window.mockOperations = [
        {
          operation_id: 'completed-pipeline-test',
          status: 'completed',
          progress: 100,
          steps: [
            {
              id: 'scraping-step',
              name: 'Data Collection',
              status: 'completed',
              progress: 100,
              metadata: {
                stage_id: 'scraping',
                files_downloaded: 15,
                duration: '3m 15s'
              }
            },
            {
              id: 'processing-step',
              name: 'Excel to CSV Processing',
              status: 'completed',
              progress: 100,
              metadata: {
                stage_id: 'processing',
                files_processed: 15,
                duration: '2m 30s'
              }
            },
            {
              id: 'indices-step',
              name: 'Index Extraction',
              status: 'completed',
              progress: 100,
              metadata: {
                stage_id: 'indices',
                indices_extracted: 2,
                duration: '1m 45s'
              }
            },
            {
              id: 'liquidity-step',
              name: 'Liquidity Analysis',
              status: 'completed',
              progress: 100,
              metadata: {
                stage_id: 'liquidity',
                analysis_files: 1,
                duration: '2m 10s'
              }
            }
          ],
          metadata: {
            pipeline_summary: {
              stage_timeline: {
                scraping: { started_at: '2025-10-17T17:00:00Z', completed_at: '2025-10-17T17:03:15Z' },
                processing: { started_at: '2025-10-17T17:03:20Z', completed_at: '2025-10-17T17:05:50Z' },
                indices: { started_at: '2025-10-17T17:05:55Z', completed_at: '2025-10-17T17:07:40Z' },
                liquidity: { started_at: '2025-10-17T17:07:45Z', completed_at: '2025-10-17T17:09:55Z' }
              }
            }
          }
        }
      ]
    })

    await page.goto('/operations')
    await page.waitForLoadState('networkidle')

    // Take screenshot of completed pipeline
    await expect(page.locator('body')).toHaveScreenshot('completed-pipeline-stages.png', {
      fullPage: true,
      animations: 'disabled'
    })

    // Verify all stages show as completed
    await expect(page.locator('text=Completed')).toHaveCount(4)

    // Verify final stage shows next steps (since it's the last stage)
    await expect(page.locator('button:has-text("Next Operation")')).toBeVisible()
  })

  test('single stage operations display without pipeline connectors', async ({ page }) => {
    // Mock single stage operations
    await page.addInitScript(() => {
      window.mockOperations = [
        {
          operation_id: 'single-scraping',
          status: 'completed',
          progress: 100,
          steps: [{
            id: 'scraping-step',
            name: 'Data Collection',
            status: 'completed',
            progress: 100,
            metadata: {
              stage_id: 'scraping',
              files_downloaded: 15
            }
          }]
        },
        {
          operation_id: 'single-processing',
          status: 'running',
          progress: 60,
          steps: [{
            id: 'processing-step',
            name: 'Excel to CSV Processing',
            status: 'active',
            progress: 60,
            metadata: {
              stage_id: 'processing',
              files_processed: 9,
              total_files: 15
            }
          }]
        }
      ]
    })

    await page.goto('/operations')
    await page.waitForLoadState('networkidle')

    // Take screenshot of single stage operations
    await expect(page.locator('body')).toHaveScreenshot('single-stage-operations.png', {
      fullPage: true,
      animations: 'disabled'
    })

    // Verify no stage numbers or connectors for single stages
    await expect(page.locator('text=Stage 1 of')).toHaveCount(0)
    await expect(page.locator('[role="separator"]')).toHaveCount(0)

    // Verify both operations are visible
    await expect(page.locator('text=Data Collection')).toBeVisible()
    await expect(page.locator('text=Excel to CSV Processing')).toBeVisible()
  })

  test('KPI information displays correctly for different stages', async ({ page }) => {
    // Mock operations with rich KPI data
    await page.addInitScript(() => {
      window.mockOperations = [
        {
          operation_id: 'kpi-display-test',
          status: 'running',
          progress: 70,
          steps: [
            {
              id: 'indicators-step',
              name: 'Indicator Calculation',
              status: 'active',
              progress: 70,
              metadata: {
                stage_id: 'indicators',
                indicators_calculated: 14,
                total_indicators: 20,
                tickers_processed: 67,
                total_tickers: 95,
                current_indicator: 'RSI',
                estimated_completion: '1m 30s'
              }
            }
          ]
        }
      ]
    })

    await page.goto('/operations')
    await page.waitForLoadState('networkidle')

    // Take screenshot focusing on KPI display
    await expect(page.locator('[data-testid="operation-ticket"]').first()).toHaveScreenshot('kpi-display-indicators.png', {
      animations: 'disabled'
    })

    // Verify KPI information is displayed
    await expect(page.locator('text=/14.*20.*indicators/i')).toBeVisible()
    await expect(page.locator('text=/67.*95.*tickers/i')).toBeVisible()
    await expect(page.locator('text=/calculating RSI/i')).toBeVisible()
    await expect(page.locator('text=/1m 30s/i')).toBeVisible()
  })

  test('error states display properly in pipeline stages', async ({ page }) => {
    // Mock pipeline with failed stage
    await page.addInitScript(() => {
      window.mockOperations = [
        {
          operation_id: 'error-pipeline-test',
          status: 'failed',
          progress: 40,
          steps: [
            {
              id: 'scraping-step',
              name: 'Data Collection',
              status: 'completed',
              progress: 100,
              metadata: {
                stage_id: 'scraping',
                files_downloaded: 15
              }
            },
            {
              id: 'processing-step',
              name: 'Excel to CSV Processing',
              status: 'failed',
              progress: 40,
              error: 'Failed to process market_data.xlsx: Invalid file format',
              metadata: {
                stage_id: 'processing',
                files_processed: 6,
                error_details: 'File corruption detected in Excel format'
              }
            }
          ]
        }
      ]
    })

    await page.goto('/operations')
    await page.waitForLoadState('networkidle')

    // Take screenshot of error state
    await expect(page.locator('[data-testid="operation-ticket"]').first()).toHaveScreenshot('pipeline-error-state.png', {
      animations: 'disabled'
    })

    // Verify error information is displayed
    await expect(page.locator('text=Failed')).toBeVisible()
    await expect(page.locator('text=/Invalid file format/i')).toBeVisible()

    // Verify completed stage still shows properly
    await expect(page.locator('text=Completed')).toBeVisible()
  })

  test('responsive design works correctly on different screen sizes', async ({ page }) => {
    // Mock pipeline operation
    await page.addInitScript(() => {
      window.mockOperations = [
        {
          operation_id: 'responsive-test',
          status: 'running',
          progress: 50,
          steps: [
            {
              id: 'scraping-step',
              name: 'Data Collection',
              status: 'completed',
              progress: 100,
              metadata: { stage_id: 'scraping', files_downloaded: 15 }
            },
            {
              id: 'processing-step',
              name: 'Excel to CSV Processing',
              status: 'active',
              progress: 50,
              metadata: { stage_id: 'processing', files_processed: 8 }
            },
            {
              id: 'indices-step',
              name: 'Index Extraction',
              status: 'pending',
              progress: 0,
              metadata: { stage_id: 'indices' }
            }
          ]
        }
      ]
    })

    await page.goto('/operations')
    await page.waitForLoadState('networkidle')

    // Test different screen sizes
    const sizes = [
      { width: 1920, height: 1080, name: 'desktop' },
      { width: 1024, height: 768, name: 'tablet' },
      { width: 375, height: 667, name: 'mobile' }
    ]

    for (const size of sizes) {
      await page.setViewportSize({ width: size.width, height: size.height })
      await page.waitForTimeout(500) // Allow layout to adjust

      await expect(page.locator('body')).toHaveScreenshot(`pipeline-responsive-${size.name}.png`, {
        fullPage: true,
        animations: 'disabled'
      })

      // Verify content is still accessible
      await expect(page.locator('text=Data Collection')).toBeVisible()
      await expect(page.locator('text=Excel to CSV Processing')).toBeVisible()
      await expect(page.locator('text=Index Extraction')).toBeVisible()
    }
  })

  test('accessibility features work correctly', async ({ page }) => {
    // Mock pipeline operation
    await page.addInitScript(() => {
      window.mockOperations = [
        {
          operation_id: 'a11y-test',
          status: 'running',
          progress: 50,
          steps: [
            {
              id: 'scraping-step',
              name: 'Data Collection',
              status: 'completed',
              progress: 100,
              metadata: { stage_id: 'scraping', files_downloaded: 15 }
            },
            {
              id: 'processing-step',
              name: 'Excel to CSV Processing',
              status: 'active',
              progress: 50,
              metadata: { stage_id: 'processing', files_processed: 8 }
            }
          ]
        }
      ]
    })

    await page.goto('/operations')
    await page.waitForLoadState('networkidle')

    // Test keyboard navigation
    await page.keyboard.press('Tab')
    await expect(page.locator(':focus')).toBeVisible()

    // Test ARIA attributes
    const progressBars = page.locator('[role="progressbar"]')
    await expect(progressBars).toHaveCount(2)

    for (let i = 0; i < await progressBars.count(); i++) {
      const progressBar = progressBars.nth(i)
      await expect(progressBar).toHaveAttribute('aria-valuenow')
      await expect(progressBar).toHaveAttribute('aria-valuemin')
      await expect(progressBar).toHaveAttribute('aria-valuemax')
    }

    // Test screen reader announcements
    const statusElements = page.locator('[aria-live]')
    await expect(statusElements).toHaveCount.greaterThan(0)

    // Test reduced motion preference
    await page.emulateMedia({ reducedMotion: 'reduce' })
    await page.waitForTimeout(500)

    // Verify animations are disabled
    const connectorElements = page.locator('[role="separator"]')
    for (let i = 0; i < await connectorElements.count(); i++) {
      const connector = connectorElements.nth(i)
      const styles = await connector.evaluate(el => getComputedStyle(el))
      expect(styles.transition).toContain('0s') // Transitions should be disabled
    }
  })
})