package license

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// ShowRenewalNotification displays renewal notification if needed.
// Note: kept as a console helper to avoid mixing UI output into the core manager.
func (m *Manager) ShowRenewalNotification() error {
	renewalInfo, err := m.CheckRenewalStatus()
	if err != nil {
		// Missing/corrupted license isn't fatal for showing a notification; just log and continue
		if strings.Contains(err.Error(), "no local license found") {
			m.logDebug(context.Background(), "license_renewal_notice", "No license found; skipping renewal notice",
				slog.String("error", err.Error()),
			)
			return nil
		}
		return err
	}

	if !renewalInfo.NeedsRenewal {
		return nil
	}

	fmt.Printf("\n")
	fmt.Printf("==============================================================\n")
	fmt.Printf("                      LICENSE RENEWAL NOTICE                  \n")
	fmt.Printf("==============================================================\n")

	switch renewalInfo.Status {
	case "Expired":
		fmt.Printf("| STATUS: EXPIRED                                            |\n")
	case "Critical":
		fmt.Printf("| STATUS: CRITICAL - EXPIRES SOON                           |\n")
	case "Warning":
		fmt.Printf("| STATUS: WARNING - RENEWAL RECOMMENDED                     |\n")
	}

	fmt.Printf("| %-58s |\n", renewalInfo.Message)

	if renewalInfo.IsExpired {
		fmt.Printf("| Application functionality is limited until renewal.       |\n")
	} else {
		fmt.Printf("| Contact support for license renewal options.              |\n")
	}

	fmt.Printf("| Support: contact your license provider                    |\n")
	fmt.Printf("==============================================================\n\n")

	return nil
}

// TestNetworkConnectivity tests connectivity to Google/HTTPS endpoints for diagnosis.
func (m *Manager) TestNetworkConnectivity() error {
	ctx := context.Background()
	m.logInfo(ctx, "network_connectivity_test", "Starting network connectivity tests")
	fmt.Printf("Testing network connectivity...\n")

	if m.sheetsService == nil {
		m.logError(ctx, "connectivity_test", "Google Sheets service not initialized")
		return fmt.Errorf("google sheets service not initialized")
	}

	// Test basic internet connectivity
	fmt.Printf("   - Testing basic internet connectivity...")
	m.logDebug(ctx, "connectivity_test", "Testing basic internet connectivity")
	resp, err := http.Get("https://www.google.com")
	if err != nil {
		m.logError(ctx, "connectivity_test", "Basic internet connectivity failed", slog.String("error", err.Error()))
		fmt.Printf(" FAILED\n")
		return fmt.Errorf("no internet connection: %v", err)
	}
	resp.Body.Close()
	fmt.Printf(" OK\n")

	// Test time sync (rough check via time API)
	fmt.Printf("   - Testing time synchronization...")
	client := http.Client{Timeout: 3 * time.Second}
	resp, err = client.Get("https://worldtimeapi.org/api/ip")
	if err != nil {
		m.logWarn(ctx, "time_sync_test", "Time synchronization check failed", slog.String("error", err.Error()))
		fmt.Printf(" WARNING\n")
	} else {
		resp.Body.Close()
		fmt.Printf(" OK\n")
	}

	fmt.Printf("Network connectivity tests completed.\n")
	m.logInfo(ctx, "network_connectivity_test_complete", "Network connectivity tests completed")

	return nil
}
