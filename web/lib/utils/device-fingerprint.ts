/**
 * Device Fingerprint Utility for License Activation
 * Generates unique device identification for license tracking
 */

import type { DeviceFingerprint } from '@/types/index'

/**
 * Generate a browser-based device fingerprint
 * Used for license activation and device tracking
 */
export async function generateDeviceFingerprint(): Promise<DeviceFingerprint> {
  const userAgent = navigator.userAgent
  const platform = navigator.platform
  const language = navigator.language
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
  const screenResolution = `${screen.width}x${screen.height}x${screen.colorDepth}`
  
  // Get browser information
  const browserInfo = getBrowserInfo(userAgent)
  const osInfo = getOSInfo(userAgent, platform)
  
  // Collect additional entropy
  const additionalData = await collectAdditionalEntropy()
  
  // Create fingerprint data
  const fingerprintData = {
    browser: browserInfo.name,
    browserVersion: browserInfo.version,
    os: osInfo.name,
    osVersion: osInfo.version,
    platform: platform,
    screenResolution: screenResolution,
    timezone: timezone,
    language: language,
    userAgent: userAgent,
    timestamp: new Date().toISOString(),
    ...additionalData
  }

  // Generate hash
  const hash = await generateFingerprightHash(fingerprintData)
  
  return {
    ...fingerprintData,
    hash
  }
}

/**
 * Extract browser information from user agent
 */
function getBrowserInfo(userAgent: string): { name: string; version: string } {
  // Chrome
  if (userAgent.includes('Chrome/') && !userAgent.includes('Edg/')) {
    const match = userAgent.match(/Chrome\/([0-9\.]+)/)
    return {
      name: 'Chrome',
      version: match?.[1] || 'unknown'
    }
  }
  
  // Edge
  if (userAgent.includes('Edg/')) {
    const match = userAgent.match(/Edg\/([0-9\.]+)/)
    return {
      name: 'Edge',
      version: match?.[1] || 'unknown'
    }
  }
  
  // Firefox
  if (userAgent.includes('Firefox/')) {
    const match = userAgent.match(/Firefox\/([0-9\.]+)/)
    return {
      name: 'Firefox',
      version: match?.[1] || 'unknown'
    }
  }
  
  // Safari
  if (userAgent.includes('Safari/') && !userAgent.includes('Chrome/')) {
    const match = userAgent.match(/Version\/([0-9\.]+)/)
    return {
      name: 'Safari',
      version: match?.[1] || 'unknown'
    }
  }
  
  return {
    name: 'Unknown',
    version: 'unknown'
  }
}

/**
 * Extract OS information from user agent and platform
 */
function getOSInfo(userAgent: string, platform: string): { name: string; version: string } {
  // Windows
  if (userAgent.includes('Windows NT')) {
    const match = userAgent.match(/Windows NT ([0-9\.]+)/)
    const version = match?.[1] || 'unknown'
    
    // Map Windows NT versions to friendly names
    const windowsVersions: Record<string, string> = {
      '10.0': 'Windows 10/11',
      '6.3': 'Windows 8.1',
      '6.2': 'Windows 8',
      '6.1': 'Windows 7',
      '6.0': 'Windows Vista'
    }
    
    return {
      name: 'Windows',
      version: windowsVersions[version] || `Windows NT ${version}`
    }
  }
  
  // macOS
  if (userAgent.includes('Mac OS X')) {
    const match = userAgent.match(/Mac OS X ([0-9_]+)/)
    const version = match?.[1]?.replace(/_/g, '.') || 'unknown'
    return {
      name: 'macOS',
      version: version
    }
  }
  
  // Linux
  if (userAgent.includes('Linux') || platform.includes('Linux')) {
    return {
      name: 'Linux',
      version: 'unknown'
    }
  }
  
  // Mobile
  if (userAgent.includes('iPhone OS')) {
    const match = userAgent.match(/iPhone OS ([0-9_]+)/)
    const version = match?.[1]?.replace(/_/g, '.') || 'unknown'
    return {
      name: 'iOS',
      version: version
    }
  }
  
  if (userAgent.includes('Android')) {
    const match = userAgent.match(/Android ([0-9\.]+)/)
    return {
      name: 'Android',
      version: match?.[1] || 'unknown'
    }
  }
  
  return {
    name: 'Unknown',
    version: 'unknown'
  }
}

/**
 * Collect additional entropy for fingerprinting
 */
async function collectAdditionalEntropy(): Promise<Record<string, any>> {
  const entropy: Record<string, any> = {}
  
  try {
    // Canvas fingerprinting (basic)
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    if (ctx) {
      canvas.width = 200
      canvas.height = 50
      ctx.textBaseline = 'top'
      ctx.font = '14px Arial'
      ctx.fillText('ISX Pulse Fingerprint Test', 2, 2)
      entropy.canvasFingerprint = canvas.toDataURL().slice(-50) // Last 50 chars
    }
  } catch {
    entropy.canvasFingerprint = 'unavailable'
  }
  
  try {
    // WebGL fingerprinting (basic)
    const canvas = document.createElement('canvas')
    const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl')
    if (gl) {
      const renderer = gl.getParameter(gl.RENDERER)
      const vendor = gl.getParameter(gl.VENDOR)
      entropy.webglRenderer = renderer || 'unknown'
      entropy.webglVendor = vendor || 'unknown'
    }
  } catch {
    entropy.webglRenderer = 'unavailable'
    entropy.webglVendor = 'unavailable'
  }
  
  try {
    // Hardware concurrency
    entropy.hardwareConcurrency = navigator.hardwareConcurrency || 'unknown'
  } catch {
    entropy.hardwareConcurrency = 'unavailable'
  }
  
  try {
    // Memory information (if available)
    const memory = (navigator as any).deviceMemory
    if (memory) {
      entropy.deviceMemory = memory
    }
  } catch {
    entropy.deviceMemory = 'unavailable'
  }
  
  try {
    // Connection information (if available)
    const connection = (navigator as any).connection
    if (connection) {
      entropy.connectionType = connection.effectiveType || 'unknown'
      entropy.downlink = connection.downlink || 'unknown'
    }
  } catch {
    entropy.connectionType = 'unavailable'
  }
  
  return entropy
}

/**
 * Generate SHA-256 hash of fingerprint data
 */
async function generateFingerprightHash(data: any): Promise<string> {
  try {
    // Convert data to string
    const dataString = JSON.stringify(data, Object.keys(data).sort())
    
    // Generate hash using Web Crypto API
    const encoder = new TextEncoder()
    const dataBuffer = encoder.encode(dataString)
    const hashBuffer = await crypto.subtle.digest('SHA-256', dataBuffer)
    
    // Convert to hex string
    const hashArray = Array.from(new Uint8Array(hashBuffer))
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
    
    return hashHex
  } catch (error) {
    // Fallback to simple hash if crypto API is not available
    console.warn('Web Crypto API not available, using fallback hash')
    return generateFallbackHash(JSON.stringify(data))
  }
}

/**
 * Fallback hash function for environments without Web Crypto API
 */
function generateFallbackHash(str: string): string {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = ((hash << 5) - hash) + char
    hash = hash & hash // Convert to 32-bit integer
  }
  
  // Convert to positive hex
  return Math.abs(hash).toString(16).padStart(8, '0')
}

/**
 * Get a simplified device identifier for display purposes
 */
export function getDeviceDisplayName(fingerprint: DeviceFingerprint): string {
  const { browser, browserVersion, os, osVersion } = fingerprint
  
  // Format version numbers to be more readable
  const formatVersion = (version: string) => {
    return version.split('.')[0] // Just major version
  }
  
  const browserDisplay = browserVersion !== 'unknown' 
    ? `${browser} ${formatVersion(browserVersion)}`
    : browser
    
  const osDisplay = osVersion !== 'unknown' && osVersion !== 'unknown' 
    ? `${os} ${osVersion}`
    : os
  
  return `${browserDisplay} on ${osDisplay}`
}

/**
 * Compare two device fingerprints for similarity
 * Returns a score from 0 (completely different) to 1 (identical)
 */
export function compareFingerprints(fp1: DeviceFingerprint, fp2: DeviceFingerprint): number {
  if (fp1.hash === fp2.hash) return 1
  
  let score = 0
  let totalChecks = 0
  
  // Check core properties
  const coreProperties = ['browser', 'os', 'platform', 'screenResolution']
  coreProperties.forEach(prop => {
    totalChecks++
    if (fp1[prop as keyof DeviceFingerprint] === fp2[prop as keyof DeviceFingerprint]) {
      score += 0.25 // Core properties are weighted more
    }
  })
  
  // Check secondary properties
  const secondaryProperties = ['timezone', 'language']
  secondaryProperties.forEach(prop => {
    totalChecks++
    if (fp1[prop as keyof DeviceFingerprint] === fp2[prop as keyof DeviceFingerprint]) {
      score += 0.125 // Secondary properties are weighted less
    }
  })
  
  return Math.min(score, 1)
}

/**
 * Validate that a device fingerprint is complete and valid
 */
export function validateFingerprint(fingerprint: DeviceFingerprint): boolean {
  const requiredFields = [
    'browser', 'browserVersion', 'os', 'osVersion', 
    'platform', 'screenResolution', 'timezone', 
    'language', 'userAgent', 'hash', 'timestamp'
  ]
  
  return requiredFields.every(field => 
    fingerprint[field as keyof DeviceFingerprint] !== undefined &&
    fingerprint[field as keyof DeviceFingerprint] !== ''
  )
}

/**
 * Generate a human-readable device summary for security purposes
 */
export function generateDeviceSummary(fingerprint: DeviceFingerprint): string {
  const browserInfo = `${fingerprint.browser} ${fingerprint.browserVersion}`
  const osInfo = `${fingerprint.os} ${fingerprint.osVersion}`
  const location = fingerprint.timezone
  
  return `${browserInfo} on ${osInfo} (${location})`
}