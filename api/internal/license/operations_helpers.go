package license

import (
	"fmt"
	"time"
)

// TransferLicense transfers a license - simplified since no machine binding.
func (m *Manager) TransferLicense(licenseKey string, forceTransfer bool) error {
	return m.ActivateLicense(licenseKey)
}

// UpdateLastConnected updates the last connected time in both local storage and Google Sheets.
func (m *Manager) UpdateLastConnected() error {
	license, err := m.loadLicenseLocal()
	if err != nil {
		return fmt.Errorf("no local license found: %v", err)
	}

	license.LastChecked = time.Now()

	if err := m.saveLicenseLocal(license); err != nil {
		return fmt.Errorf("failed to save license locally: %v", err)
	}

	if err := m.updateLicenseInSheets(license); err != nil {
		return fmt.Errorf("failed to update last connected time in sheets: %v", err)
	}

	return nil
}
