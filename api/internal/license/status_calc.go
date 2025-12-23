package license

import "time"

// calculateLicenseStatus determines license status and days left based on expiry and current time.
func calculateLicenseStatus(license LicenseInfo, now time.Time) (status string, daysLeft int) {
	daysLeft = daysLeftRounded(license.ExpiryDate)

	if now.After(license.ExpiryDate) {
		return "Expired", daysLeft
	}
	if daysLeft <= 7 {
		return "Critical", daysLeft
	}
	if daysLeft <= 30 {
		return "Warning", daysLeft
	}
	return "Active", daysLeft
}
