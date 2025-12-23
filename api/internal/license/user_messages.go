package license

import (
	"fmt"
	"strings"
	"time"
)

// UserMessage represents a user-friendly error message with context
type UserMessage struct {
	Title         string `json:"title"`
	Message       string `json:"message"`
	Suggestion    string `json:"suggestion,omitempty"`
	CanRetry      bool   `json:"can_retry"`
	SeverityLevel string `json:"severity"` // "error", "warning", "info", "success"
}

// ReactivationInfo provides context about reactivation status
type ReactivationInfo struct {
	Count         int       `json:"count"`
	Limit         int       `json:"limit"`
	Remaining     int       `json:"remaining"`
	NextResetDate time.Time `json:"next_reset_date,omitempty"`
}

// LicenseSuccessInfo provides context for successful operations
type LicenseSuccessInfo struct {
	ExpiryDate       time.Time         `json:"expiry_date"`
	DaysRemaining    int               `json:"days_remaining"`
	ReactivationInfo *ReactivationInfo `json:"reactivation_info,omitempty"`
	IsReactivation   bool              `json:"is_reactivation"`
	ShowWarning      bool              `json:"show_warning"`
	WarningMessage   string            `json:"warning_message,omitempty"`
}

// userMessageMap maps technical error patterns to user-friendly messages
var userMessageMap = map[string]UserMessage{
	"signature verification failed": {
		Title:         "License Key Not Recognized",
		Message:       "The license key you entered is not recognized by our system.",
		Suggestion:    "Please verify that you've entered your license key correctly. License keys look like: ISX-XXXXX-XXXXX-XXXXX-XXXXX",
		CanRetry:      true,
		SeverityLevel: "error",
	},
	"already activated on different device": {
		Title:         "License Registered to Another Computer",
		Message:       "This license is currently registered to a different computer.",
		Suggestion:    "If you want to transfer this license to this computer, please contact support. Each license can be reactivated up to 5 times per month.",
		CanRetry:      false,
		SeverityLevel: "error",
	},
	"already activated": {
		Title:         "License Already Activated",
		Message:       "This license has already been activated.",
		Suggestion:    "If this is the same device, the license should still work. If you're on a different device, contact support for assistance.",
		CanRetry:      false,
		SeverityLevel: "warning",
	},
	"reactivation limit exceeded": {
		Title:         "Reactivation Limit Reached",
		Message:       "You've used all 5 reactivations this month.",
		Suggestion:    "The reactivation limit resets on the first day of each month. For immediate assistance, contact support.",
		CanRetry:      false,
		SeverityLevel: "error",
	},
	"network timeout": {
		Title:         "Connection Timeout",
		Message:       "The connection to our activation server timed out after multiple attempts.",
		Suggestion:    "Please check your internet connection and try again. If the problem persists, check your firewall settings or try again later.",
		CanRetry:      true,
		SeverityLevel: "error",
	},
	"network error": {
		Title:         "Connection Problem",
		Message:       "Cannot connect to the activation server after multiple retry attempts.",
		Suggestion:    "Please check your internet connection and firewall settings, then try again. The server may be temporarily unavailable.",
		CanRetry:      true,
		SeverityLevel: "error",
	},
	"connection failed": {
		Title:         "Connection Failed After Retries",
		Message:       "Failed to connect to the activation server after 3 attempts with exponential backoff.",
		Suggestion:    "This usually indicates a network connectivity issue. Please check your internet connection, firewall settings, and try again in a few minutes.",
		CanRetry:      true,
		SeverityLevel: "error",
	},
	"deadline exceeded": {
		Title:         "Connection Timeout",
		Message:       "The activation request took too long to complete.",
		Suggestion:    "The activation server may be busy or your connection is slow. Please try again in a few minutes.",
		CanRetry:      true,
		SeverityLevel: "error",
	},
	"invalid license format": {
		Title:         "Invalid License Key Format",
		Message:       "The license key format is not valid.",
		Suggestion:    "License keys should look like: ISX-XXXXX-XXXXX-XXXXX-XXXXX (with dashes and exactly 5 groups of 5 characters each)",
		CanRetry:      true,
		SeverityLevel: "error",
	},
	"license not found": {
		Title:         "License Key Not Found",
		Message:       "This license key was not found in our database.",
		Suggestion:    "Please verify that you've entered the license key correctly. If you continue to have issues, contact support.",
		CanRetry:      true,
		SeverityLevel: "error",
	},
	"license expired": {
		Title:         "License Has Expired",
		Message:       "This license has expired and can no longer be used.",
		Suggestion:    "Please obtain a new license or contact support to renew your existing license.",
		CanRetry:      false,
		SeverityLevel: "error",
	},
	"invalid request": {
		Title:         "Invalid Request",
		Message:       "The activation request contains invalid data.",
		Suggestion:    "Please ensure you're using the latest version of the activation tool and try again.",
		CanRetry:      true,
		SeverityLevel: "error",
	},
	"rate limited": {
		Title:         "Too Many Attempts",
		Message:       "You've made too many activation attempts in a short period.",
		Suggestion:    "Please wait 15 minutes before trying again.",
		CanRetry:      true,
		SeverityLevel: "warning",
	},
}

// GetUserFriendlyMessage converts a technical error to a user-friendly message
func GetUserFriendlyMessage(err error) UserMessage {
	if err == nil {
		return UserMessage{
			Title:         "Success",
			Message:       "Operation completed successfully.",
			SeverityLevel: "success",
		}
	}

	errorText := strings.ToLower(err.Error())

	// Prioritize specific patterns that could be shadowed by broader matches
	if strings.Contains(errorText, "different device") {
		return userMessageMap["already activated on different device"]
	}

	// Check for specific error patterns
	for pattern, message := range userMessageMap {
		if strings.Contains(errorText, pattern) {
			return message
		}
	}

	// Check for HMAC-related errors
	if strings.Contains(errorText, "hmac") || strings.Contains(errorText, "signature") {
		return userMessageMap["signature verification failed"]
	}

	// Check for timeout errors
	if strings.Contains(errorText, "timeout") || strings.Contains(errorText, "deadline exceeded") {
		if strings.Contains(errorText, "deadline exceeded") {
			return userMessageMap["deadline exceeded"]
		}
		return userMessageMap["network timeout"]
	}

	// Check for retry-related errors
	if strings.Contains(errorText, "connection failed after") && strings.Contains(errorText, "attempts") {
		return userMessageMap["connection failed"]
	}

	// Check for validation errors
	if strings.Contains(errorText, "validation") || strings.Contains(errorText, "invalid") {
		if strings.Contains(errorText, "format") || strings.Contains(errorText, "start with") || strings.Contains(errorText, "must be") {
			return userMessageMap["invalid license format"]
		}
		return userMessageMap["license not found"]
	}

	// Default error message for unrecognized errors
	return UserMessage{
		Title:         "Activation Error",
		Message:       "An unexpected error occurred during activation.",
		Suggestion:    "Please try again. If the problem persists, contact support with the error details.",
		CanRetry:      true,
		SeverityLevel: "error",
	}
}

// GetUserFriendlyMessageWithContext provides enhanced error messages with additional context
func GetUserFriendlyMessageWithContext(err error, reactivationInfo *ReactivationInfo) UserMessage {
	baseMessage := GetUserFriendlyMessage(err)

	// Enhance reactivation limit messages with specific dates
	if strings.Contains(strings.ToLower(err.Error()), "reactivation limit") && reactivationInfo != nil {
		if !reactivationInfo.NextResetDate.IsZero() {
			resetDate := reactivationInfo.NextResetDate.Format("January 2, 2006")
			baseMessage.Message = fmt.Sprintf("You've used all %d reactivations this month (%d/%d used).",
				reactivationInfo.Limit, reactivationInfo.Count, reactivationInfo.Limit)
			baseMessage.Suggestion = fmt.Sprintf("The reactivation limit resets on %s. For immediate assistance, contact support.", resetDate)
		}
	}

	// Enhance already activated messages with reactivation info
	if strings.Contains(strings.ToLower(err.Error()), "already activated") && reactivationInfo != nil {
		if reactivationInfo.Remaining > 0 {
			baseMessage.Message += fmt.Sprintf(" You have %d reactivations remaining this month.", reactivationInfo.Remaining)
			baseMessage.Suggestion = "You may be able to reactivate this license on this device. Contact support for assistance."
			baseMessage.CanRetry = true
		} else {
			baseMessage.Message += " You have used all reactivations this month."
			if !reactivationInfo.NextResetDate.IsZero() {
				resetDate := reactivationInfo.NextResetDate.Format("January 2, 2006")
				baseMessage.Suggestion = fmt.Sprintf("Reactivations reset on %s, or contact support for immediate assistance.", resetDate)
			}
		}
	}

	return baseMessage
}

// FormatSuccessMessage creates a detailed success message for license activation
func FormatSuccessMessage(info *LicenseSuccessInfo) UserMessage {
	var title, message, suggestion string

	if info.IsReactivation {
		title = "License Reactivated Successfully"
		message = "Your license has been successfully reactivated on this device."
	} else {
		title = "License Activated Successfully"
		message = "Your license has been successfully activated."
	}

	// Add expiry information
	if !info.ExpiryDate.IsZero() {
		expiryStr := info.ExpiryDate.Format("January 2, 2006")
		if info.DaysRemaining > 0 {
			message += fmt.Sprintf("\n\nExpiry: %s (%d days remaining)", expiryStr, info.DaysRemaining)
		} else {
			message += fmt.Sprintf("\n\nExpiry: %s", expiryStr)
		}
	}

	// Add reactivation information if available
	if info.ReactivationInfo != nil && info.ReactivationInfo.Count > 0 {
		message += fmt.Sprintf("\n\nReactivations used: %d/%d",
			info.ReactivationInfo.Count, info.ReactivationInfo.Limit)
	}

	// Determine severity and warnings
	severity := "success"
	if info.ShowWarning {
		severity = "warning"
		if info.WarningMessage != "" {
			suggestion = info.WarningMessage
		}
	}

	// Add warnings based on reactivation count
	if info.ReactivationInfo != nil {
		if info.ReactivationInfo.Remaining <= 2 && info.ReactivationInfo.Remaining > 0 {
			warningMsg := fmt.Sprintf("⚠️ Warning: You have only %d reactivation(s) remaining this month.", info.ReactivationInfo.Remaining)
			if suggestion != "" {
				suggestion = warningMsg + " " + suggestion
			} else {
				suggestion = warningMsg
			}
			severity = "warning"
		} else if info.ReactivationInfo.Remaining == 0 {
			warningMsg := "⚠️ Warning: You have used all reactivations for this month."
			if !info.ReactivationInfo.NextResetDate.IsZero() {
				resetDate := info.ReactivationInfo.NextResetDate.Format("January 2, 2006")
				warningMsg += fmt.Sprintf(" They will reset on %s.", resetDate)
			}
			if suggestion != "" {
				suggestion = warningMsg + " " + suggestion
			} else {
				suggestion = warningMsg
			}
			severity = "warning"
		}
	}

	return UserMessage{
		Title:         title,
		Message:       message,
		Suggestion:    suggestion,
		CanRetry:      false,
		SeverityLevel: severity,
	}
}

// CalculateReactivationInfo creates reactivation info with next reset date
func CalculateReactivationInfo(currentCount, maxCount int) *ReactivationInfo {
	now := time.Now()
	// Reset date is the first day of the next month
	nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())

	remaining := maxCount - currentCount
	if remaining < 0 {
		remaining = 0
	}

	return &ReactivationInfo{
		Count:         currentCount,
		Limit:         maxCount,
		Remaining:     remaining,
		NextResetDate: nextMonth,
	}
}

// GetLicenseStatusMessage creates user-friendly status messages for license information
func GetLicenseStatusMessage(licenseInfo *LicenseInfo) UserMessage {
	if licenseInfo == nil {
		return UserMessage{
			Title:         "No License Found",
			Message:       "No license has been activated on this system.",
			Suggestion:    "Please activate a license to use this software.",
			CanRetry:      true,
			SeverityLevel: "warning",
		}
	}

	daysUntilExpiry := int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)

	var title, message, suggestion, severity string

	switch licenseInfo.Status {
	case "active":
		if daysUntilExpiry <= 0 {
			title = "License Expired"
			message = "Your license has expired."
			suggestion = "Please renew your license to continue using this software."
			severity = "error"
		} else if daysUntilExpiry <= 7 {
			title = "License Expiring Soon"
			message = fmt.Sprintf("Your license expires in %d days (%s).",
				daysUntilExpiry, licenseInfo.ExpiryDate.Format("January 2, 2006"))
			suggestion = "Please renew your license soon to avoid interruption."
			severity = "warning"
		} else if daysUntilExpiry <= 30 {
			title = "License Active"
			message = fmt.Sprintf("Your license is active and expires in %d days (%s).",
				daysUntilExpiry, licenseInfo.ExpiryDate.Format("January 2, 2006"))
			suggestion = "Consider renewing your license soon."
			severity = "info"
		} else {
			title = "License Active"
			message = fmt.Sprintf("Your license is active and expires on %s (%d days remaining).",
				licenseInfo.ExpiryDate.Format("January 2, 2006"), daysUntilExpiry)
			severity = "success"
		}
	case "expired":
		title = "License Expired"
		expiredDays := int(time.Since(licenseInfo.ExpiryDate).Hours() / 24)
		message = fmt.Sprintf("Your license expired %d days ago on %s.",
			expiredDays, licenseInfo.ExpiryDate.Format("January 2, 2006"))
		suggestion = "Please renew your license to continue using this software."
		severity = "error"
	case "revoked":
		title = "License Revoked"
		message = "Your license has been revoked."
		suggestion = "Please contact support for assistance."
		severity = "error"
	default:
		title = "License Status Unknown"
		message = fmt.Sprintf("License status: %s", licenseInfo.Status)
		suggestion = "Please contact support if you need assistance."
		severity = "warning"
	}

	return UserMessage{
		Title:         title,
		Message:       message,
		Suggestion:    suggestion,
		CanRetry:      false,
		SeverityLevel: severity,
	}
}
