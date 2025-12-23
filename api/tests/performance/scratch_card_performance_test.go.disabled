package performance

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"isxcli/internal/license"
	"isxcli/internal/security"
)

// BenchmarkScratchCardActivationSpeed benchmarks the speed of scratch card activation
func BenchmarkScratchCardActivationSpeed(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "perf_activation_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	// Setup mock Apps Script server
	mockServer := setupMockAppsScriptServer()
	defer mockServer.Close()

	licenseFile := filepath.Join(tempDir, "bench_license.dat")
	manager, err := license.NewManager(licenseFile)
	require.NoError(b, err)

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		licenseKey := fmt.Sprintf("ISX-PERF-%04d-%03d", i%10000, i%1000)
		
		_, err := manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
		if err != nil {
			b.Fatalf("Activation failed at iteration %d: %v", i, err)
		}
		
		// Clean up for next iteration
		os.Remove(licenseFile)
		manager, _ = license.NewManager(licenseFile)
	}
}

// BenchmarkDeviceFingerprintGeneration benchmarks device fingerprint generation speed
func BenchmarkDeviceFingerprintGeneration(b *testing.B) {
	fingerprintMgr := security.NewFingerprintManager()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := fingerprintMgr.GenerateFingerprint()
		if err != nil {
			b.Fatalf("Fingerprint generation failed at iteration %d: %v", i, err)
		}
	}
}

// BenchmarkConcurrentActivations benchmarks concurrent scratch card activations
func BenchmarkConcurrentActivations(b *testing.B) {
	benchmarks := []struct {
		name        string
		concurrency int
	}{
		{"Concurrent_1", 1},
		{"Concurrent_2", 2},
		{"Concurrent_4", 4},
		{"Concurrent_8", 8},
		{"Concurrent_16", 16},
		{"Concurrent_32", 32},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			benchmarkConcurrentActivations(b, bm.concurrency)
		})
	}
}

func benchmarkConcurrentActivations(b *testing.B, concurrency int) {
	tempDir, err := os.MkdirTemp("", "perf_concurrent_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	// Setup mock Apps Script server
	mockServer := setupMockAppsScriptServer()
	defer mockServer.Close()

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()
	
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		var counter int64
		for pb.Next() {
			id := atomic.AddInt64(&counter, 1)
			
			licenseFile := filepath.Join(tempDir, fmt.Sprintf("license_%d_%d.dat", id, time.Now().UnixNano()))
			manager, err := license.NewManager(licenseFile)
			if err != nil {
				b.Fatalf("Failed to create manager: %v", err)
			}

			licenseKey := fmt.Sprintf("ISX-CONC-%04d-%03d", id%10000, id%1000)
			_, err = manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
			if err != nil {
				b.Fatalf("Concurrent activation failed: %v", err)
			}
		}
	})
}

// BenchmarkLicenseValidation benchmarks license validation speed
func BenchmarkLicenseValidation(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "perf_validation_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	// Setup and activate a license first
	mockServer := setupMockAppsScriptServer()
	defer mockServer.Close()

	licenseFile := filepath.Join(tempDir, "validation_license.dat")
	manager, err := license.NewManager(licenseFile)
	require.NoError(b, err)

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()
	licenseKey := "ISX-VALID-TEST-001"

	_, err = manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		valid, err := manager.ValidateLicenseWithContext(ctx)
		if err != nil {
			b.Fatalf("Validation failed at iteration %d: %v", i, err)
		}
		if !valid {
			b.Fatalf("License should be valid at iteration %d", i)
		}
	}
}

// BenchmarkMemoryUsage benchmarks memory usage during operations
func BenchmarkMemoryUsage(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "perf_memory_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	mockServer := setupMockAppsScriptServer()
	defer mockServer.Close()

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		licenseFile := filepath.Join(tempDir, fmt.Sprintf("memory_test_%d.dat", i))
		manager, err := license.NewManager(licenseFile)
		if err != nil {
			b.Fatalf("Failed to create manager: %v", err)
		}

		licenseKey := fmt.Sprintf("ISX-MEM-%04d-%03d", i%10000, i%1000)
		_, err = manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
		if err != nil {
			b.Fatalf("Activation failed: %v", err)
		}

		// Validate license
		_, err = manager.ValidateLicenseWithContext(ctx)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "B/op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
}

// BenchmarkHighThroughputActivations benchmarks high throughput activation scenarios
func BenchmarkHighThroughputActivations(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "perf_throughput_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	mockServer := setupMockAppsScriptServer()
	defer mockServer.Close()

	const batchSize = 100
	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		
		for j := 0; j < batchSize; j++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				licenseFile := filepath.Join(tempDir, fmt.Sprintf("throughput_%d_%d.dat", i, id))
				manager, err := license.NewManager(licenseFile)
				if err != nil {
					b.Errorf("Failed to create manager: %v", err)
					return
				}

				licenseKey := fmt.Sprintf("ISX-THRU-%04d-%03d", (i*batchSize+id)%10000, id%1000)
				_, err = manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
				if err != nil {
					b.Errorf("Throughput activation failed: %v", err)
				}
			}(j)
		}
		
		wg.Wait()
	}
}

// BenchmarkAppsScriptIntegration benchmarks Apps Script integration performance
func BenchmarkAppsScriptIntegration(b *testing.B) {
	mockServer := setupMockAppsScriptServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "perf_apps_script_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		licenseFile := filepath.Join(tempDir, fmt.Sprintf("apps_script_%d.dat", i))
		manager, err := license.NewManager(licenseFile)
		if err != nil {
			b.Fatalf("Failed to create manager: %v", err)
		}

		licenseKey := fmt.Sprintf("ISX-APPS-%04d-%03d", i%10000, i%1000)
		
		// This will make HTTP requests to the mock server
		_, err = manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
		if err != nil {
			b.Fatalf("Apps Script integration failed: %v", err)
		}
	}
}

// BenchmarkFingerprintCaching benchmarks fingerprint caching performance
func BenchmarkFingerprintCaching(b *testing.B) {
	fingerprintMgr := security.NewFingerprintManager()
	
	// Generate initial fingerprint to populate cache
	_, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// This should hit cache after first generation
		_, err := fingerprintMgr.GenerateFingerprint()
		if err != nil {
			b.Fatalf("Cached fingerprint generation failed: %v", err)
		}
	}
}

// BenchmarkLicenseDataPersistence benchmarks license data persistence performance
func BenchmarkLicenseDataPersistence(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "perf_persistence_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	mockServer := setupMockAppsScriptServer()
	defer mockServer.Close()

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		licenseFile := filepath.Join(tempDir, fmt.Sprintf("persistence_%d.dat", i))
		
		// Create and activate license
		manager1, err := license.NewManager(licenseFile)
		if err != nil {
			b.Fatalf("Failed to create manager: %v", err)
		}

		licenseKey := fmt.Sprintf("ISX-PERS-%04d-%03d", i%10000, i%1000)
		_, err = manager1.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
		if err != nil {
			b.Fatalf("Activation failed: %v", err)
		}

		// Recreate manager (simulating restart) and validate persistence
		manager2, err := license.NewManager(licenseFile)
		if err != nil {
			b.Fatalf("Failed to recreate manager: %v", err)
		}

		valid, err := manager2.ValidateLicenseWithContext(ctx)
		if err != nil {
			b.Fatalf("Persistence validation failed: %v", err)
		}
		if !valid {
			b.Fatalf("License should be valid after persistence at iteration %d", i)
		}
	}
}

// BenchmarkStressTest benchmarks system under stress conditions
func BenchmarkStressTest(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "perf_stress_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	mockServer := setupMockAppsScriptServer()
	defer mockServer.Close()

	const (
		numWorkers = 50
		workersPerCPU = 8
	)

	maxWorkers := runtime.NumCPU() * workersPerCPU
	workers := numWorkers
	if workers > maxWorkers {
		workers = maxWorkers
	}

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		var successCount, errorCount int64
		
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				licenseFile := filepath.Join(tempDir, fmt.Sprintf("stress_%d_%d.dat", i, workerID))
				manager, err := license.NewManager(licenseFile)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					return
				}

				licenseKey := fmt.Sprintf("ISX-STRS-%04d-%03d", (i*workers+workerID)%10000, workerID%1000)
				_, err = manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}(w)
		}
		
		wg.Wait()

		if errorCount > 0 {
			b.Logf("Stress test iteration %d: %d successes, %d errors", i, successCount, errorCount)
		}
	}
}

// BenchmarkNetworkLatencySimulation benchmarks performance under network latency
func BenchmarkNetworkLatencySimulation(b *testing.B) {
	latencies := []time.Duration{
		0 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		250 * time.Millisecond,
		500 * time.Millisecond,
	}

	for _, latency := range latencies {
		b.Run(fmt.Sprintf("Latency_%dms", latency.Milliseconds()), func(b *testing.B) {
			benchmarkWithNetworkLatency(b, latency)
		})
	}
}

func benchmarkWithNetworkLatency(b *testing.B, latency time.Duration) {
	tempDir, err := os.MkdirTemp("", "perf_latency_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	// Setup mock server with artificial latency
	mockServer := setupMockAppsScriptServerWithLatency(latency)
	defer mockServer.Close()

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		licenseFile := filepath.Join(tempDir, fmt.Sprintf("latency_%d.dat", i))
		manager, err := license.NewManager(licenseFile)
		if err != nil {
			b.Fatalf("Failed to create manager: %v", err)
		}

		licenseKey := fmt.Sprintf("ISX-LAT-%04d-%03d", i%10000, i%1000)
		_, err = manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
		if err != nil {
			b.Fatalf("Latency simulation failed: %v", err)
		}
	}
}

// BenchmarkErrorRecovery benchmarks performance during error conditions
func BenchmarkErrorRecovery(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "perf_error_recovery_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	// Setup server that fails intermittently
	mockServer := setupFlakyAppsScriptServer(0.2) // 20% failure rate
	defer mockServer.Close()

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		licenseFile := filepath.Join(tempDir, fmt.Sprintf("error_recovery_%d.dat", i))
		manager, err := license.NewManager(licenseFile)
		if err != nil {
			b.Fatalf("Failed to create manager: %v", err)
		}

		licenseKey := fmt.Sprintf("ISX-ERR-%04d-%03d", i%10000, i%1000)
		_, err = manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
		
		// Don't fail the benchmark on expected errors from flaky server
		// This tests error recovery performance
	}
}

// Helper functions for setting up test servers

func setupMockAppsScriptServer() *httptest.Server {
	activatedLicenses := make(map[string]map[string]interface{})
	var mutex sync.RWMutex

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add minimal processing delay to simulate real server
		time.Sleep(1 * time.Millisecond)

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Simple successful response for benchmarking
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "activationId": "benchmark_activation", "message": "Success"}`))
	}))
}

func setupMockAppsScriptServerWithLatency(latency time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate network latency
		time.Sleep(latency)

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "activationId": "latency_test_activation", "message": "Success"}`))
	}))
}

func setupFlakyAppsScriptServer(failureRate float64) *httptest.Server {
	var requestCount int64

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt64(&requestCount, 1)
		
		// Simulate intermittent failures
		if float64(count%10)/10.0 < failureRate {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"success": false, "error": "Simulated server error"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "activationId": "flaky_test_activation", "message": "Success"}`))
	}))
}

// Performance reporting helpers

func init() {
	// Set up performance testing environment
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// BenchmarkResourceUtilization provides insights into resource utilization
func BenchmarkResourceUtilization(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "perf_resources_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	mockServer := setupMockAppsScriptServer()
	defer mockServer.Close()

	fingerprintMgr := security.NewFingerprintManager()
	fingerprint, err := fingerprintMgr.GenerateFingerprint()
	require.NoError(b, err)

	ctx := context.Background()

	// Track goroutine count
	initialGoroutines := runtime.NumGoroutine()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		licenseFile := filepath.Join(tempDir, fmt.Sprintf("resource_%d.dat", i))
		manager, err := license.NewManager(licenseFile)
		if err != nil {
			b.Fatalf("Failed to create manager: %v", err)
		}

		licenseKey := fmt.Sprintf("ISX-RES-%04d-%03d", i%10000, i%1000)
		_, err = manager.ActivateScratchCard(ctx, licenseKey, fingerprint.Fingerprint)
		if err != nil {
			b.Fatalf("Resource utilization test failed: %v", err)
		}

		// Check for goroutine leaks periodically
		if i%100 == 0 {
			currentGoroutines := runtime.NumGoroutine()
			if currentGoroutines > initialGoroutines+10 {
				b.Logf("Potential goroutine leak detected: initial=%d, current=%d", initialGoroutines, currentGoroutines)
			}
		}
	}

	// Final check for resource cleanup
	runtime.GC()
	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > initialGoroutines+5 {
		b.Logf("Resource cleanup issue: initial=%d, final=%d goroutines", initialGoroutines, finalGoroutines)
	}
}