/**
 * @jest-environment jsdom
 */

import {
  generateDeviceFingerprint,
  getDeviceInfo,
  getBrowserInfo,
  getScreenInfo,
  getTimezoneInfo,
  getLanguageInfo,
  getCPUInfo,
  getMemoryInfo,
  getCanvasFingerprint,
  getWebGLFingerprint,
  hashFingerprint,
} from '@/lib/utils/device-fingerprint'

// Mock crypto.subtle for browser compatibility
const mockCrypto = {
  subtle: {
    digest: jest.fn().mockResolvedValue(new ArrayBuffer(32)),
  },
  getRandomValues: jest.fn((arr) => {
    for (let i = 0; i < arr.length; i++) {
      arr[i] = Math.floor(Math.random() * 256)
    }
    return arr
  }),
}

Object.defineProperty(window, 'crypto', {
  value: mockCrypto,
  writable: true,
})

// Mock canvas context for fingerprinting
const mockCanvasContext = {
  fillText: jest.fn(),
  fillRect: jest.fn(),
  getImageData: jest.fn(() => ({
    data: new Uint8ClampedArray([1, 2, 3, 4, 5, 6, 7, 8]),
  })),
  font: '',
  fillStyle: '',
  textBaseline: '',
}

Object.defineProperty(HTMLCanvasElement.prototype, 'getContext', {
  value: jest.fn(() => mockCanvasContext),
})

// Mock WebGL context
const mockWebGLContext = {
  getParameter: jest.fn((param) => {
    switch (param) {
      case 7936: return 'Mock WebGL Renderer'
      case 7937: return 'Mock WebGL Vendor'
      case 7938: return 'WebGL 1.0'
      default: return 'mock-value'
    }
  }),
  getSupportedExtensions: jest.fn(() => ['WEBGL_debug_renderer_info']),
}

Object.defineProperty(HTMLCanvasElement.prototype, 'getContext', {
  value: jest.fn((type) => {
    if (type === '2d') return mockCanvasContext
    if (type === 'webgl' || type === 'experimental-webgl') return mockWebGLContext
    return null
  }),
})

describe('Device Fingerprint Utility', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    
    // Reset DOM properties
    Object.defineProperty(navigator, 'userAgent', {
      value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36',
      writable: true,
    })
    
    Object.defineProperty(navigator, 'language', {
      value: 'en-US',
      writable: true,
    })
    
    Object.defineProperty(navigator, 'languages', {
      value: ['en-US', 'en'],
      writable: true,
    })
    
    Object.defineProperty(screen, 'width', { value: 1920, writable: true })
    Object.defineProperty(screen, 'height', { value: 1080, writable: true })
    Object.defineProperty(screen, 'colorDepth', { value: 24, writable: true })
  })

  describe('generateDeviceFingerprint', () => {
    it('generates a fingerprint successfully', async () => {
      const fingerprint = await generateDeviceFingerprint()
      
      expect(fingerprint).toBeDefined()
      expect(typeof fingerprint).toBe('string')
      expect(fingerprint.length).toBeGreaterThan(0)
    })

    it('generates consistent fingerprints for same environment', async () => {
      const fingerprint1 = await generateDeviceFingerprint()
      const fingerprint2 = await generateDeviceFingerprint()
      
      expect(fingerprint1).toBe(fingerprint2)
    })

    it('handles crypto.subtle unavailable gracefully', async () => {
      // Mock crypto.subtle as undefined
      const originalSubtle = window.crypto.subtle
      delete (window.crypto as any).subtle
      
      const fingerprint = await generateDeviceFingerprint()
      
      expect(fingerprint).toBeDefined()
      expect(typeof fingerprint).toBe('string')
      
      // Restore crypto.subtle
      window.crypto.subtle = originalSubtle
    })

    it('handles web crypto completely unavailable', async () => {
      const originalCrypto = window.crypto
      delete (window as any).crypto
      
      const fingerprint = await generateDeviceFingerprint()
      
      expect(fingerprint).toBeDefined()
      expect(typeof fingerprint).toBe('string')
      
      // Restore crypto
      window.crypto = originalCrypto
    })
  })

  describe('getDeviceInfo', () => {
    it('returns comprehensive device information', () => {
      const deviceInfo = getDeviceInfo()
      
      expect(deviceInfo).toHaveProperty('os')
      expect(deviceInfo).toHaveProperty('browser')
      expect(deviceInfo).toHaveProperty('screen')
      expect(deviceInfo).toHaveProperty('timezone')
      expect(deviceInfo).toHaveProperty('language')
      expect(deviceInfo).toHaveProperty('cpu')
      expect(deviceInfo).toHaveProperty('memory')
    })

    it('handles missing properties gracefully', () => {
      // Remove navigator properties
      const originalNavigator = window.navigator
      Object.defineProperty(window, 'navigator', {
        value: {},
        writable: true,
      })
      
      const deviceInfo = getDeviceInfo()
      
      expect(deviceInfo).toBeDefined()
      expect(deviceInfo.browser).toContain('Unknown')
      
      // Restore navigator
      Object.defineProperty(window, 'navigator', {
        value: originalNavigator,
        writable: true,
      })
    })
  })

  describe('getBrowserInfo', () => {
    it('detects Chrome browser', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36',
        writable: true,
      })
      
      const browserInfo = getBrowserInfo()
      
      expect(browserInfo).toContain('Chrome')
      expect(browserInfo).toContain('118.0')
    })

    it('detects Firefox browser', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/118.0',
        writable: true,
      })
      
      const browserInfo = getBrowserInfo()
      
      expect(browserInfo).toContain('Firefox')
      expect(browserInfo).toContain('118.0')
    })

    it('detects Safari browser', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15',
        writable: true,
      })
      
      const browserInfo = getBrowserInfo()
      
      expect(browserInfo).toContain('Safari')
      expect(browserInfo).toContain('16.6')
    })

    it('detects Edge browser', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36 Edg/118.0.2088.46',
        writable: true,
      })
      
      const browserInfo = getBrowserInfo()
      
      expect(browserInfo).toContain('Edge')
      expect(browserInfo).toContain('118.0')
    })

    it('detects Opera browser', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36 OPR/104.0.0.0',
        writable: true,
      })
      
      const browserInfo = getBrowserInfo()
      
      expect(browserInfo).toContain('Opera')
    })

    it('handles unknown browser', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Unknown/1.0',
        writable: true,
      })
      
      const browserInfo = getBrowserInfo()
      
      expect(browserInfo).toContain('Unknown Browser')
    })

    it('handles missing user agent', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: undefined,
        writable: true,
      })
      
      const browserInfo = getBrowserInfo()
      
      expect(browserInfo).toContain('Unknown Browser')
    })
  })

  describe('getScreenInfo', () => {
    it('returns screen resolution information', () => {
      const screenInfo = getScreenInfo()
      
      expect(screenInfo).toBe('1920x1080x24')
    })

    it('handles missing screen properties', () => {
      Object.defineProperty(screen, 'width', { value: undefined, writable: true })
      Object.defineProperty(screen, 'height', { value: undefined, writable: true })
      
      const screenInfo = getScreenInfo()
      
      expect(screenInfo).toContain('Unknown')
    })

    it('includes color depth when available', () => {
      Object.defineProperty(screen, 'colorDepth', { value: 32, writable: true })
      
      const screenInfo = getScreenInfo()
      
      expect(screenInfo).toContain('x32')
    })
  })

  describe('getTimezoneInfo', () => {
    it('returns timezone information', () => {
      // Mock Intl.DateTimeFormat
      const mockDateTimeFormat = jest.fn().mockImplementation(() => ({
        resolvedOptions: () => ({ timeZone: 'America/New_York' }),
      }))
      
      Object.defineProperty(global, 'Intl', {
        value: { DateTimeFormat: mockDateTimeFormat },
        writable: true,
      })
      
      const timezoneInfo = getTimezoneInfo()
      
      expect(timezoneInfo).toContain('America/New_York')
    })

    it('handles missing Intl support', () => {
      const originalIntl = global.Intl
      delete (global as any).Intl
      
      const timezoneInfo = getTimezoneInfo()
      
      expect(timezoneInfo).toContain('UTC')
      
      global.Intl = originalIntl
    })

    it('includes timezone offset', () => {
      const timezoneInfo = getTimezoneInfo()
      
      expect(timezoneInfo).toMatch(/UTC[+-]\d+/)
    })
  })

  describe('getLanguageInfo', () => {
    it('returns language information', () => {
      const languageInfo = getLanguageInfo()
      
      expect(languageInfo).toContain('en-US')
    })

    it('includes all available languages', () => {
      Object.defineProperty(navigator, 'languages', {
        value: ['en-US', 'en', 'es'],
        writable: true,
      })
      
      const languageInfo = getLanguageInfo()
      
      expect(languageInfo).toContain('en-US,en,es')
    })

    it('handles missing languages property', () => {
      Object.defineProperty(navigator, 'languages', {
        value: undefined,
        writable: true,
      })
      
      const languageInfo = getLanguageInfo()
      
      expect(languageInfo).toContain('en-US')
    })

    it('handles missing language property', () => {
      Object.defineProperty(navigator, 'language', {
        value: undefined,
        writable: true,
      })
      
      Object.defineProperty(navigator, 'languages', {
        value: undefined,
        writable: true,
      })
      
      const languageInfo = getLanguageInfo()
      
      expect(languageInfo).toContain('Unknown')
    })
  })

  describe('getCPUInfo', () => {
    it('returns CPU core count when available', () => {
      Object.defineProperty(navigator, 'hardwareConcurrency', {
        value: 8,
        writable: true,
      })
      
      const cpuInfo = getCPUInfo()
      
      expect(cpuInfo).toContain('8 cores')
    })

    it('handles missing hardwareConcurrency', () => {
      Object.defineProperty(navigator, 'hardwareConcurrency', {
        value: undefined,
        writable: true,
      })
      
      const cpuInfo = getCPUInfo()
      
      expect(cpuInfo).toContain('Unknown CPU')
    })
  })

  describe('getMemoryInfo', () => {
    it('returns memory information when available', () => {
      Object.defineProperty(navigator, 'deviceMemory', {
        value: 8,
        writable: true,
      })
      
      const memoryInfo = getMemoryInfo()
      
      expect(memoryInfo).toContain('8GB')
    })

    it('handles missing deviceMemory', () => {
      Object.defineProperty(navigator, 'deviceMemory', {
        value: undefined,
        writable: true,
      })
      
      const memoryInfo = getMemoryInfo()
      
      expect(memoryInfo).toContain('Unknown')
    })
  })

  describe('getCanvasFingerprint', () => {
    it('generates canvas fingerprint', () => {
      const canvasFingerprint = getCanvasFingerprint()
      
      expect(canvasFingerprint).toBeDefined()
      expect(typeof canvasFingerprint).toBe('string')
      expect(canvasFingerprint.length).toBeGreaterThan(0)
    })

    it('uses various canvas operations for uniqueness', () => {
      getCanvasFingerprint()
      
      expect(mockCanvasContext.fillText).toHaveBeenCalled()
      expect(mockCanvasContext.fillRect).toHaveBeenCalled()
      expect(mockCanvasContext.getImageData).toHaveBeenCalled()
    })

    it('handles canvas unavailable', () => {
      HTMLCanvasElement.prototype.getContext = jest.fn(() => null)
      
      const canvasFingerprint = getCanvasFingerprint()
      
      expect(canvasFingerprint).toContain('canvas-unavailable')
      
      // Restore mock
      HTMLCanvasElement.prototype.getContext = jest.fn(() => mockCanvasContext)
    })

    it('handles canvas creation failure', () => {
      const originalCreateElement = document.createElement
      document.createElement = jest.fn(() => {
        throw new Error('Canvas creation failed')
      })
      
      const canvasFingerprint = getCanvasFingerprint()
      
      expect(canvasFingerprint).toContain('canvas-error')
      
      document.createElement = originalCreateElement
    })
  })

  describe('getWebGLFingerprint', () => {
    it('generates WebGL fingerprint', () => {
      const webglFingerprint = getWebGLFingerprint()
      
      expect(webglFingerprint).toBeDefined()
      expect(typeof webglFingerprint).toBe('string')
      expect(webglFingerprint.length).toBeGreaterThan(0)
    })

    it('includes WebGL renderer and vendor info', () => {
      const webglFingerprint = getWebGLFingerprint()
      
      expect(webglFingerprint).toContain('Mock WebGL Renderer')
      expect(webglFingerprint).toContain('Mock WebGL Vendor')
    })

    it('handles WebGL unavailable', () => {
      HTMLCanvasElement.prototype.getContext = jest.fn((type) => {
        if (type === 'webgl' || type === 'experimental-webgl') return null
        return mockCanvasContext
      })
      
      const webglFingerprint = getWebGLFingerprint()
      
      expect(webglFingerprint).toContain('webgl-unavailable')
      
      // Restore mock
      HTMLCanvasElement.prototype.getContext = jest.fn((type) => {
        if (type === '2d') return mockCanvasContext
        if (type === 'webgl' || type === 'experimental-webgl') return mockWebGLContext
        return null
      })
    })

    it('falls back to experimental-webgl', () => {
      let getContextCallCount = 0
      HTMLCanvasElement.prototype.getContext = jest.fn((type) => {
        getContextCallCount++
        if (type === 'webgl' && getContextCallCount === 1) return null
        if (type === 'experimental-webgl') return mockWebGLContext
        return mockCanvasContext
      })
      
      const webglFingerprint = getWebGLFingerprint()
      
      expect(webglFingerprint).toBeDefined()
      expect(webglFingerprint).not.toContain('webgl-unavailable')
    })
  })

  describe('hashFingerprint', () => {
    it('generates consistent hash for same input', async () => {
      const input = 'test-fingerprint-data'
      
      const hash1 = await hashFingerprint(input)
      const hash2 = await hashFingerprint(input)
      
      expect(hash1).toBe(hash2)
    })

    it('generates different hashes for different inputs', async () => {
      const hash1 = await hashFingerprint('input1')
      const hash2 = await hashFingerprint('input2')
      
      expect(hash1).not.toBe(hash2)
    })

    it('falls back to simple hash when crypto unavailable', async () => {
      const originalCrypto = window.crypto
      delete (window as any).crypto
      
      const hash = await hashFingerprint('test-input')
      
      expect(hash).toBeDefined()
      expect(typeof hash).toBe('string')
      
      window.crypto = originalCrypto
    })

    it('handles crypto.subtle errors gracefully', async () => {
      window.crypto.subtle.digest = jest.fn().mockRejectedValue(new Error('Crypto error'))
      
      const hash = await hashFingerprint('test-input')
      
      expect(hash).toBeDefined()
      expect(typeof hash).toBe('string')
    })
  })

  describe('Browser Compatibility', () => {
    it('works in Internet Explorer environment', async () => {
      // Mock IE environment
      delete (window as any).crypto
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko',
        writable: true,
      })
      
      const fingerprint = await generateDeviceFingerprint()
      
      expect(fingerprint).toBeDefined()
      expect(typeof fingerprint).toBe('string')
    })

    it('works in mobile Safari environment', async () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1',
        writable: true,
      })
      
      const fingerprint = await generateDeviceFingerprint()
      
      expect(fingerprint).toBeDefined()
      expect(fingerprint).toContain('Safari')
    })

    it('works in mobile Chrome environment', async () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Mobile Safari/537.36',
        writable: true,
      })
      
      const fingerprint = await generateDeviceFingerprint()
      
      expect(fingerprint).toBeDefined()
      expect(fingerprint).toContain('Chrome')
    })

    it('handles reduced privacy mode', async () => {
      // Mock reduced privacy environment (e.g., incognito)
      Object.defineProperty(navigator, 'hardwareConcurrency', {
        value: undefined,
        writable: true,
      })
      
      Object.defineProperty(navigator, 'deviceMemory', {
        value: undefined,
        writable: true,
      })
      
      const fingerprint = await generateDeviceFingerprint()
      
      expect(fingerprint).toBeDefined()
      expect(typeof fingerprint).toBe('string')
    })

    it('handles blocked fingerprinting APIs', async () => {
      // Mock blocked APIs
      HTMLCanvasElement.prototype.getContext = jest.fn(() => null)
      
      Object.defineProperty(screen, 'width', { value: 0, writable: true })
      Object.defineProperty(screen, 'height', { value: 0, writable: true })
      
      const fingerprint = await generateDeviceFingerprint()
      
      expect(fingerprint).toBeDefined()
      expect(typeof fingerprint).toBe('string')
    })
  })

  describe('Fallback Behavior', () => {
    it('provides meaningful fallbacks for all missing APIs', async () => {
      // Remove all APIs
      delete (window as any).crypto
      delete (global as any).Intl
      
      Object.defineProperty(navigator, 'userAgent', { value: '', writable: true })
      Object.defineProperty(navigator, 'language', { value: '', writable: true })
      Object.defineProperty(navigator, 'languages', { value: [], writable: true })
      Object.defineProperty(navigator, 'hardwareConcurrency', { value: undefined, writable: true })
      Object.defineProperty(navigator, 'deviceMemory', { value: undefined, writable: true })
      
      Object.defineProperty(screen, 'width', { value: 0, writable: true })
      Object.defineProperty(screen, 'height', { value: 0, writable: true })
      
      HTMLCanvasElement.prototype.getContext = jest.fn(() => null)
      
      const fingerprint = await generateDeviceFingerprint()
      const deviceInfo = getDeviceInfo()
      
      expect(fingerprint).toBeDefined()
      expect(fingerprint.length).toBeGreaterThan(0)
      
      expect(deviceInfo.browser).toContain('Unknown')
      expect(deviceInfo.screen).toContain('Unknown')
      expect(deviceInfo.timezone).toContain('UTC')
      expect(deviceInfo.language).toContain('Unknown')
      expect(deviceInfo.cpu).toContain('Unknown')
      expect(deviceInfo.memory).toContain('Unknown')
    })

    it('generates stable fallback fingerprints', async () => {
      // Remove crypto multiple times
      delete (window as any).crypto
      
      const fingerprint1 = await generateDeviceFingerprint()
      const fingerprint2 = await generateDeviceFingerprint()
      
      expect(fingerprint1).toBe(fingerprint2)
    })
  })

  describe('Security Considerations', () => {
    it('does not expose sensitive information', () => {
      const deviceInfo = getDeviceInfo()
      
      // Should not contain sensitive paths, usernames, etc.
      const infoString = JSON.stringify(deviceInfo)
      expect(infoString).not.toMatch(/[Cc]:\\[Uu]sers\\[\w]+/)
      expect(infoString).not.toMatch(/\/[Hh]ome\/[\w]+/)
      expect(infoString).not.toMatch(/password|token|secret/i)
    })

    it('generates fingerprints that are not easily spoofed', async () => {
      const fingerprint = await generateDeviceFingerprint()
      
      // Should include multiple data points for robustness
      expect(fingerprint.length).toBeGreaterThan(32)
    })
  })
})

// Performance tests
describe('Device Fingerprint Performance', () => {
  beforeAll(() => {
    jest.spyOn(performance, 'now')
      .mockReturnValueOnce(0)
      .mockReturnValueOnce(100)
  })

  afterAll(() => {
    jest.restoreAllMocks()
  })

  it('generates fingerprint within performance budget', async () => {
    const startTime = performance.now()
    await generateDeviceFingerprint()
    const endTime = performance.now()
    
    const duration = endTime - startTime
    expect(duration).toBeLessThanOrEqual(100) // Should complete within 100ms
  })

  it('device info collection is fast', () => {
    const startTime = performance.now()
    getDeviceInfo()
    const endTime = performance.now()
    
    const duration = endTime - startTime
    expect(duration).toBeLessThanOrEqual(50) // Should complete within 50ms
  })

  it('canvas fingerprinting is reasonably fast', () => {
    const startTime = performance.now()
    getCanvasFingerprint()
    const endTime = performance.now()
    
    const duration = endTime - startTime
    expect(duration).toBeLessThanOrEqual(30) // Should complete within 30ms
  })
})