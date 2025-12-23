package license

import (
	"fmt"
	"math"
	"time"
)

// CheckRenewalStatus checks if license needs renewal and returns detailed info.
func (m *Manager) CheckRenewalStatus() (*RenewalInfo, error) {
	license, err := m.loadLicenseLocal()
	if err != nil {
		return &RenewalInfo{
			Status:       "No License",
			Message:      "No license found. Please activate a license.",
			NeedsRenewal: true,
			IsExpired:    true,
		}, fmt.Errorf("no local license found: %v", err)
	}

	daysLeft := int(math.Ceil(time.Until(license.ExpiryDate).Hours() / 24))
	renewalInfo := &RenewalInfo{DaysLeft: daysLeft}

	switch {
	case time.Now().After(license.ExpiryDate):
		renewalInfo.Status = "Expired"
		renewalInfo.Message = fmt.Sprintf("License expired %d days ago. Please renew immediately.", -daysLeft)
		renewalInfo.NeedsRenewal = true
		renewalInfo.IsExpired = true
	case daysLeft <= 7:
		renewalInfo.Status = "Critical"
		renewalInfo.Message = fmt.Sprintf("License expires in %d days! Please renew soon to avoid interruption.", daysLeft)
		renewalInfo.NeedsRenewal = true
	case daysLeft <= 30:
		renewalInfo.Status = "Warning"
		renewalInfo.Message = fmt.Sprintf("License expires in %d days. Consider renewing soon.", daysLeft)
		renewalInfo.NeedsRenewal = true
	default:
		renewalInfo.Status = "Active"
		renewalInfo.Message = fmt.Sprintf("License is active with %d days remaining.", daysLeft)
		renewalInfo.NeedsRenewal = false
	}

	return renewalInfo, nil
}

// ValidateWithRenewalCheck performs validation and checks for renewal needs.
func (m *Manager) ValidateWithRenewalCheck() (bool, *RenewalInfo, error) {
	isValid, err := m.ValidateLicense()

	renewalInfo, renewalErr := m.CheckRenewalStatus()
	if renewalErr != nil {
		renewalInfo = &RenewalInfo{
			Status:       "No License",
			Message:      "No license found",
			NeedsRenewal: true,
			IsExpired:    true,
		}
	}

	if renewalInfo.NeedsRenewal {
		m.ShowRenewalNotification()
	}

	return isValid, renewalInfo, err
}
