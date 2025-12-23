package license

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/isxcli/isxcli/internal/config"
)

// saveLicenseToSheets saves license to Google Sheets.
func (m *Manager) saveLicenseToSheets(license LicenseInfo) error {
	values := []interface{}{
		license.LicenseKey,
		license.UserEmail,
		license.ExpiryDate.Format("2006-01-02 15:04:05"),
		license.Duration,
		"", // Machine ID removed
		license.IssuedDate.Format("2006-01-02 15:04:05"),
		license.Status,
		license.LastChecked.Format("2006-01-02 15:04:05"),
	}

	url := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s:append?valueInputOption=RAW&key=%s",
		m.config.SheetID, m.config.SheetName, m.config.APIKey)

	payload := map[string]interface{}{
		"values": [][]interface{}{values},
	}

	return m.makeSheetRequest("POST", url, payload)
}

// validateLicenseFromAppsScript validates license via Google Apps Script endpoint.
func (m *Manager) validateLicenseFromAppsScript(licenseKey string) (LicenseInfo, error) {
	var license LicenseInfo

	ctx := context.Background()
	creds := config.GetCredentials()
	m.logInfo(ctx, "apps_script_validation", "Validating license via Apps Script",
		slog.String("license_key_prefix", licenseKey[:min(8, len(licenseKey))]),
		slog.String("endpoint", creds.AppsScriptURL),
	)

	requestData := map[string]interface{}{
		"action": "validate",
		"code":   licenseKey,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		m.logError(ctx, "apps_script_validation", "Failed to marshal request data",
			slog.String("error", err.Error()),
		)
		return license, fmt.Errorf("failed to prepare request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", creds.AppsScriptURL, bytes.NewBuffer(jsonData))
	if err != nil {
		m.logError(ctx, "apps_script_validation", "Failed to create HTTP request",
			slog.String("error", err.Error()),
		)
		return license, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ISX-Pulse-License-Client/1.0")

	client := &http.Client{Timeout: 30 * time.Second}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		m.logError(ctx, "apps_script_validation", "HTTP request failed",
			slog.String("error", err.Error()),
			slog.Duration("duration", time.Since(start)),
		)
		return license, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		m.logError(ctx, "apps_script_validation", "Failed to read response",
			slog.String("error", err.Error()),
		)
		return license, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		m.logError(ctx, "apps_script_validation", "Apps Script returned error status",
			slog.Int("status_code", resp.StatusCode),
			slog.String("response_body", string(body)),
		)
		return license, fmt.Errorf("Apps Script returned status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		m.logError(ctx, "apps_script_validation", "Failed to parse response JSON",
			slog.String("error", err.Error()),
			slog.String("response_body", string(body)),
		)
		return license, fmt.Errorf("failed to parse response: %w", err)
	}

	success, ok := response["success"].(bool)
	if !ok || !success {
		errorMsg := "unknown error"
		if msg, exists := response["error"]; exists {
			errorMsg = fmt.Sprintf("%v", msg)
		}

		m.logWarn(ctx, "apps_script_validation", "Apps Script validation failed",
			slog.String("error", errorMsg),
			slog.String("license_key_prefix", licenseKey[:min(8, len(licenseKey))]),
		)
		return license, fmt.Errorf("validation failed: %s", errorMsg)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		m.logError(ctx, "apps_script_validation", "Invalid response format - missing data field",
			slog.String("response_body", string(body)),
		)
		return license, fmt.Errorf("invalid response format")
	}

	license.LicenseKey = licenseKey

	if duration, exists := data["duration"]; exists {
		license.Duration = fmt.Sprintf("%v", duration)
	}

	if status, exists := data["status"]; exists {
		license.Status = fmt.Sprintf("%v", status)
	}

	if expiryStr, exists := data["expiry_date"]; exists && fmt.Sprintf("%v", expiryStr) != "" {
		if expiryDate, err := time.Parse("2006-01-02", fmt.Sprintf("%v", expiryStr)); err == nil {
			license.ExpiryDate = expiryDate
			m.logInfo(ctx, "apps_script_activation", "Parsed expiry date from 'expiry_date' field",
				slog.String("expiry_date", expiryDate.Format("2006-01-02")),
			)
		}
	} else if expiryStr, exists := data["expires_at"]; exists && fmt.Sprintf("%v", expiryStr) != "" {
		if expiryDate, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", expiryStr)); err == nil {
			license.ExpiryDate = expiryDate
			m.logInfo(ctx, "apps_script_activation", "Parsed expiry date from 'expires_at' field",
				slog.String("expiry_date", expiryDate.Format("2006-01-02")),
			)
		} else if expiryDate, err := time.Parse("2006-01-02", fmt.Sprintf("%v", expiryStr)); err == nil {
			license.ExpiryDate = expiryDate
			m.logInfo(ctx, "apps_script_activation", "Parsed expiry date from 'expires_at' field (date format)",
				slog.String("expiry_date", expiryDate.Format("2006-01-02")),
			)
		}
	}

	if issuedStr, exists := data["issued_date"]; exists && fmt.Sprintf("%v", issuedStr) != "" {
		if issuedDate, err := time.Parse("2006-01-02", fmt.Sprintf("%v", issuedStr)); err == nil {
			license.IssuedDate = issuedDate
		}
	}

	if activationID, exists := data["activation_id"]; exists {
		license.ActivationID = fmt.Sprintf("%v", activationID)
	}

	if deviceFingerprint, exists := data["device_fingerprint"]; exists {
		license.DeviceFingerprint = fmt.Sprintf("%v", deviceFingerprint)
	}

	license.UserEmail = ""
	license.LastChecked = time.Now()

	duration := time.Since(start)
	m.logInfo(ctx, "apps_script_validation", "License validated successfully via Apps Script",
		slog.String("license_key_prefix", licenseKey[:min(8, len(licenseKey))]),
		slog.String("status", license.Status),
		slog.String("duration", license.Duration),
		slog.String("activation_id", license.ActivationID),
		slog.Duration("request_duration", duration),
	)

	return license, nil
}

// calculateExpiryDateFromDuration calculates expiry date from duration string.
func (m *Manager) calculateExpiryDateFromDuration(duration string) time.Time {
	baseTime := m.now()
	var standardExpiry time.Time
	switch duration {
	case "1m":
		standardExpiry = baseTime.AddDate(0, 1, 0)
	case "3m":
		standardExpiry = baseTime.AddDate(0, 3, 0)
	case "6m":
		standardExpiry = baseTime.AddDate(0, 6, 0)
	case "1y":
		standardExpiry = baseTime.AddDate(1, 0, 0)
	default:
		standardExpiry = baseTime.AddDate(0, 1, 0)
	}

	return time.Date(standardExpiry.Year(), standardExpiry.Month(), standardExpiry.Day()+1, 0, 0, 0, 0, standardExpiry.Location())
}

// validateWithAppsScript performs periodic remote validation and updates local license.
func (m *Manager) validateWithAppsScript(license *LicenseInfo) error {
	ctx := context.Background()

	if license.ActivationID == "" {
		if id, err := generateShortID(10); err == nil {
			license.ActivationID = id
		}
	}

	requestData := map[string]interface{}{
		"action":      ActionValidate,
		"license_key": license.LicenseKey,
		"activation_id": func() string {
			if license.ActivationID != "" {
				return license.ActivationID
			}
			return ""
		}(),
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.GetCredentials().AppsScriptURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ISX-Pulse-License-Client/1.0")

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Apps Script returned status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	success, ok := response["success"].(bool)
	if !ok || !success {
		errorMsg := "unknown error"
		if msg, exists := response["error"]; exists {
			errorMsg = fmt.Sprintf("%v", msg)
		}
		return fmt.Errorf("remote validation failed: %s", errorMsg)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid response format")
	}

	if status, exists := data["status"]; exists {
		license.Status = fmt.Sprintf("%v", status)
	}

	if expiryStr, exists := data["expiry_date"]; exists && fmt.Sprintf("%v", expiryStr) != "" {
		if expiryDate, err := time.Parse("2006-01-02", fmt.Sprintf("%v", expiryStr)); err == nil {
			license.ExpiryDate = expiryDate
		}
	} else if expiryStr, exists := data["expires_at"]; exists && fmt.Sprintf("%v", expiryStr) != "" {
		if expiryDate, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", expiryStr)); err == nil {
			license.ExpiryDate = expiryDate
		} else if expiryDate, err := time.Parse("2006-01-02", fmt.Sprintf("%v", expiryStr)); err == nil {
			license.ExpiryDate = expiryDate
		}
	}

	license.LastChecked = time.Now()

	if err := m.saveLicenseLocal(*license); err != nil {
		m.logWarn(ctx, "license_validation", "Failed to save updated license locally",
			slog.String("error", err.Error()),
		)
	}

	return nil
}

// updateLicenseInSheets updates license information in Google Sheets.
func (m *Manager) updateLicenseInSheets(license LicenseInfo) error {
	if m.sheetsService == nil {
		return fmt.Errorf("Google Sheets service not initialized")
	}

	expireStatus := m.calculateExpireStatus(license.ExpiryDate)

	rangeName := fmt.Sprintf("%s!A:H", m.config.SheetName)
	resp, err := m.sheetsService.Spreadsheets.Values.Get(m.config.SheetID, rangeName).Do()
	if err != nil {
		return fmt.Errorf("failed to fetch sheet: %w", err)
	}

	var rowIndex int64 = -1
	for i, row := range resp.Values {
		if len(row) > 0 && row[0] == license.LicenseKey {
			rowIndex = int64(i)
			break
		}
	}

	if rowIndex == -1 {
		return fmt.Errorf("license not found in sheet")
	}

	values := [][]interface{}{{
		license.LicenseKey,
		"", // Machine ID removed
		license.IssuedDate.Format("2006-01-02 15:04:05"),
		license.ExpiryDate.Format("2006-01-02 15:04:05"),
		license.Duration,
		expireStatus,
		license.LastChecked.Format("2006-01-02 15:04:05"),
		"", // Client ID not available in embedded credentials
	}}

	rangeToUpdate := fmt.Sprintf("%s!A%d:H%d", m.config.SheetName, rowIndex+1, rowIndex+1)

	payload := map[string]interface{}{
		"values": values,
	}

	updateURL := fmt.Sprintf("https://sheets.googleapis.com/v4/spreadsheets/%s/values/%s?valueInputOption=RAW&key=%s",
		m.config.SheetID, rangeToUpdate, m.config.APIKey)

	return m.makeSheetRequest("PUT", updateURL, payload)
}
