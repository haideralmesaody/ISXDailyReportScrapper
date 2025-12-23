import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright configuration for debugging analysis page
 */
export default defineConfig({
  testDir: './__tests__/e2e',
  timeout: 30 * 1000,
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: 1,
  reporter: 'list',
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  webServer: {
    command: 'echo "Using existing server on port 8080"',
    port: 8080,
    reuseExistingServer: true,
  },
});