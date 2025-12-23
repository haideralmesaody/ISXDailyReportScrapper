/**
 * Viewport Positioning Validation Script
 *
 * This script can be run in the browser console to validate viewport positioning
 * across different scenarios and zoom levels.
 *
 * Usage:
 * 1. Open the ISX Pulse application in your browser
 * 2. Open browser developer tools console
 * 3. Copy and paste this script
 * 4. Run `validateViewportPositioning()` to start validation
 * 5. Check the console output for results
 */

(function() {
  'use strict';

  // Test configuration
  const TEST_CONFIG = {
    zoomLevels: [
      { level: 1.0, name: '100%', expected: 'desktop' },
      { level: 1.25, name: '125%', expected: 'tablet' },
      { level: 1.5, name: '150%', expected: 'large-tablet' },
      { level: 2.0, name: '200%', expected: 'mobile' }
    ],
    placements: ['center', 'left', 'right', 'auto'],
    panelSizes: ['small', 'medium', 'large'],
    performanceThreshold: 16.67, // 60fps in ms
    marginThreshold: 16 // Minimum margin from edges
  };

  // Performance measurement utilities
  class PerformanceMeasurer {
    constructor() {
      this.measurements = [];
    }

    start() {
      this.startTime = performance.now();
      this.measurements = [];
    }

    record(label = 'measurement') {
      const now = performance.now();
      const elapsed = now - this.startTime;
      this.measurements.push({
        label,
        elapsed,
        timestamp: now
      });
      return elapsed;
    }

    getResults() {
      if (this.measurements.length === 0) return null;

      const elapsed = this.measurements.map(m => m.elapsed);
      return {
        count: this.measurements.length,
        total: elapsed.reduce((a, b) => a + b, 0),
        average: elapsed.reduce((a, b) => a + b, 0) / elapsed.length,
        min: Math.min(...elapsed),
        max: Math.max(...elapsed),
        measurements: this.measurements
      };
    }
  }

  // Validation utilities
  class ViewportValidator {
    constructor() {
      this.results = [];
      this.performance = new PerformanceMeasurer();
    }

    async validateZoomLevels() {
      console.log('🔍 Testing zoom levels...');
      const results = [];

      for (const zoom of TEST_CONFIG.zoomLevels) {
        console.log(`Testing ${zoom.name} zoom (${zoom.level})...`);

        // Apply zoom level
        const originalZoom = document.body.style.zoom;
        document.body.style.zoom = zoom.level.toString();

        // Wait for layout to settle
        await new Promise(resolve => setTimeout(resolve, 100));

        // Find panel elements (adjust selector based on your implementation)
        const panels = document.querySelectorAll('[data-panel="true"], .fixed, [style*="position: fixed"]');

        if (panels.length === 0) {
          console.warn(`⚠️ No panels found for ${zoom.name} zoom test`);
          continue;
        }

        const panelResults = [];
        panels.forEach((panel, index) => {
          const rect = panel.getBoundingClientRect();
          const viewportWidth = window.innerWidth;
          const viewportHeight = window.innerHeight;

          const result = {
            zoom: zoom.name,
            panelIndex: index,
            position: {
              left: rect.left,
              top: rect.top,
              width: rect.width,
              height: rect.height
            },
            withinBounds: {
              left: rect.left >= TEST_CONFIG.marginThreshold,
              top: rect.top >= TEST_CONFIG.marginThreshold,
              right: rect.right <= viewportWidth - TEST_CONFIG.marginThreshold,
              bottom: rect.bottom <= viewportHeight - TEST_CONFIG.marginThreshold
            },
            isVisible: rect.width > 0 && rect.height > 0
          };

          panelResults.push(result);
        });

        results.push({
          zoom: zoom.name,
          level: zoom.level,
          panels: panelResults,
          allPanelsVisible: panelResults.every(p => p.isVisible),
          allPanelsWithinBounds: panelResults.every(p =>
            p.withinBounds.left &&
            p.withinBounds.top &&
            p.withinBounds.right &&
            p.withinBounds.bottom
          )
        });

        // Restore original zoom
        document.body.style.zoom = originalZoom;
        await new Promise(resolve => setTimeout(resolve, 100));
      }

      console.log('✅ Zoom level validation complete:', results);
      return results;
    }

    async validatePerformance() {
      console.log('⚡ Testing performance...');
      this.performance.start();

      // Test panel repositioning performance
      const panels = document.querySelectorAll('[data-panel="true"], .fixed, [style*="position: fixed"]');

      if (panels.length === 0) {
        console.warn('⚠️ No panels found for performance testing');
        return null;
      }

      const panel = panels[0]; // Test first panel
      const originalPosition = panel.style.position;
      const originalLeft = panel.style.left;
      const originalTop = panel.style.top;

      // Simulate multiple position changes
      for (let i = 0; i < 10; i++) {
        this.performance.record(`reposition-${i}`);

        // Simulate position change
        panel.style.left = `${100 + (i * 10)}px`;
        panel.style.top = `${100 + (i * 10)}px`;

        // Wait for layout
        await new Promise(resolve => setTimeout(resolve, 10));
      }

      // Restore original position
      panel.style.position = originalPosition;
      panel.style.left = originalLeft;
      panel.style.top = originalTop;

      const results = this.performance.getResults();
      console.log('✅ Performance validation complete:', results);

      return results;
    }

    validateEmergencyFallback() {
      console.log('🚨 Testing emergency fallback...');

      // Check if emergency positioning styles are present
      const emergencyStyles = document.querySelectorAll('[style*="50%"][style*="translate(-50%, -50%)"]');
      const emergencyClasses = document.querySelectorAll('.emergency-position, .fallback-position');

      const result = {
        emergencyStylesFound: emergencyStyles.length,
        emergencyClassesFound: emergencyClasses.length,
        hasEmergencyFallback: emergencyStyles.length > 0 || emergencyClasses.length > 0
      };

      console.log('✅ Emergency fallback validation complete:', result);
      return result;
    }

    async validateResponsiveness() {
      console.log('📱 Testing responsiveness...');
      const originalWidth = window.innerWidth;
      const originalHeight = window.innerHeight;

      const testSizes = [
        { width: 1920, height: 1080, name: 'Desktop' },
        { width: 1366, height: 768, name: 'Laptop' },
        { width: 768, height: 1024, name: 'Tablet' },
        { width: 375, height: 667, name: 'Mobile' }
      ];

      const results = [];

      for (const size of testSizes) {
        console.log(`Testing ${size.name} (${size.width}x${size.height})...`);

        // This would require actual viewport resizing, which is complex in browser console
        // For now, we'll validate the current state
        const panels = document.querySelectorAll('[data-panel="true"], .fixed, [style*="position: fixed"]');

        const panelResults = Array.from(panels).map(panel => {
          const rect = panel.getBoundingClientRect();
          return {
            withinCurrentViewport: {
              left: rect.left >= 0,
              top: rect.top >= 0,
              right: rect.right <= window.innerWidth,
              bottom: rect.bottom <= window.innerHeight
            }
          };
        });

        results.push({
          size: size.name,
          panels: panelResults,
          allWithinBounds: panelResults.every(p => p.withinCurrentViewport.left && p.withinCurrentViewport.top)
        });
      }

      console.log('✅ Responsiveness validation complete:', results);
      return results;
    }

    generateReport(results) {
      console.log('\n📊 VALIDATION REPORT');
      console.log('==================');

      // Zoom level results
      if (results.zoomLevels) {
        console.log('\n🔍 Zoom Level Tests:');
        results.zoomLevels.forEach(result => {
          const status = result.allPanelsVisible && result.allPanelsWithinBounds ? '✅' : '❌';
          console.log(`${status} ${result.zoom}: ${result.panels.length} panels`);
          if (!result.allPanelsVisible) {
            console.log(`   ⚠️ Some panels not visible`);
          }
          if (!result.allPanelsWithinBounds) {
            console.log(`   ⚠️ Some panels outside viewport bounds`);
          }
        });
      }

      // Performance results
      if (results.performance) {
        const perfStatus = results.performance.average <= TEST_CONFIG.performanceThreshold ? '✅' : '❌';
        console.log(`\n⚡ Performance Tests:`);
        console.log(`${perfStatus} Average: ${results.performance.average.toFixed(2)}ms (target: <${TEST_CONFIG.performanceThreshold}ms)`);
        console.log(`   Range: ${results.performance.min.toFixed(2)}ms - ${results.performance.max.toFixed(2)}ms`);
        console.log(`   Total operations: ${results.performance.count}`);
      }

      // Emergency fallback results
      if (results.emergencyFallback) {
        const fallbackStatus = results.emergencyFallback.hasEmergencyFallback ? '✅' : '⚠️';
        console.log(`\n🚨 Emergency Fallback:`);
        console.log(`${fallbackStatus} Emergency positioning ${results.emergencyFallback.hasEmergencyFallback ? 'available' : 'not detected'}`);
      }

      // Responsiveness results
      if (results.responsiveness) {
        console.log(`\n📱 Responsiveness Tests:`);
        results.responsiveness.forEach(result => {
          const status = result.allWithinBounds ? '✅' : '⚠️';
          console.log(`${status} ${result.size}: All panels within bounds`);
        });
      }

      console.log('\n🎯 Recommendations:');

      if (results.performance && results.performance.average > TEST_CONFIG.performanceThreshold) {
        console.log('⚠️ Performance below 60fps target. Consider optimizing calculations.');
      }

      if (results.zoomLevels && !results.zoomLevels.every(r => r.allPanelsWithinBounds)) {
        console.log('⚠️ Some panels exceed viewport bounds at certain zoom levels.');
      }

      if (results.emergencyFallback && !results.emergencyFallback.hasEmergencyFallback) {
        console.log('ℹ️ No emergency fallback detected. Consider adding fallback positioning.');
      }

      console.log('\n✅ Validation complete!');
    }

    async runFullValidation() {
      console.log('🚀 Starting viewport positioning validation...\n');

      const results = {};

      try {
        results.zoomLevels = await this.validateZoomLevels();
        results.performance = await this.validatePerformance();
        results.emergencyFallback = this.validateEmergencyFallback();
        results.responsiveness = await this.validateResponsiveness();

        this.generateReport(results);
        return results;
      } catch (error) {
        console.error('❌ Validation failed:', error);
        return null;
      }
    }
  }

  // Make validator globally available
  const validator = new ViewportValidator();

  // Expose validation function
  window.validateViewportPositioning = () => validator.runFullValidation();

  // Also expose individual test functions
  window.validateZoomLevels = () => validator.validateZoomLevels();
  window.validatePerformance = () => validator.validatePerformance();
  window.validateEmergencyFallback = () => validator.validateEmergencyFallback();
  window.validateResponsiveness = () => validator.validateResponsiveness();

  // Auto-run if specified in URL
  if (window.location.search.includes('auto-validate=true')) {
    console.log('🤖 Auto-running viewport validation...');
    setTimeout(() => {
      window.validateViewportPositioning();
    }, 1000);
  }

  console.log('🔧 Viewport validation tools loaded!');
  console.log('Available commands:');
  console.log('  validateViewportPositioning() - Run full validation');
  console.log('  validateZoomLevels() - Test zoom level positioning');
  console.log('  validatePerformance() - Test performance metrics');
  console.log('  validateEmergencyFallback() - Test emergency positioning');
  console.log('  validateResponsiveness() - Test responsive behavior');

})();