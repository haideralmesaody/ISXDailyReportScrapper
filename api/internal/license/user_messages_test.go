package license

import (
	"errors"
	"testing"
	"time"
)

func TestGetUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		expectedTitle    string
		expectedSeverity string
		expectSuggestion bool
	}{
		{
			name:             "HMAC signature verification failed",
			err:              errors.New("HMAC signature verification failed"),
			expectedTitle:    "License Key Not Recognized",
			expectedSeverity: "error",
			expectSuggestion: true,
		},
		{
			name:             "Already activated on different device",
			err:              errors.New("already activated on different device"),
			expectedTitle:    "License Registered to Another Computer",
			expectedSeverity: "error",
			expectSuggestion: true,
		},
		{
			name:             "Reactivation limit exceeded",
			err:              errors.New("reactivation limit exceeded"),
			expectedTitle:    "Reactivation Limit Reached",
			expectedSeverity: "error",
			expectSuggestion: true,
		},
		{
			name:             "Network timeout",
			err:              errors.New("network timeout occurred"),
			expectedTitle:    "Connection Timeout",
			expectedSeverity: "error",
			expectSuggestion: true,
		},
		{
			name:             "Invalid license format",
			err:              errors.New("invalid license format"),
			expectedTitle:    "Invalid License Key Format",
			expectedSeverity: "error",
			expectSuggestion: true,
		},
		{
			name:             "License expired",
			err:              errors.New("license expired"),
			expectedTitle:    "License Has Expired",
			expectedSeverity: "error",
			expectSuggestion: true,
		},
		{
			name:             "Rate limited",
			err:              errors.New("rate limited"),
			expectedTitle:    "Too Many Attempts",
			expectedSeverity: "warning",
			expectSuggestion: true,
		},
		{
			name:             "Unknown error",
			err:              errors.New("some unknown error"),
			expectedTitle:    "Activation Error",
			expectedSeverity: "error",
			expectSuggestion: true,
		},
		{
			name:             "Nil error",
			err:              nil,
			expectedTitle:    "Success",
			expectedSeverity: "success",
			expectSuggestion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := GetUserFriendlyMessage(tt.err)

			if msg.Title != tt.expectedTitle {
				t.Errorf("Expected title %q, got %q", tt.expectedTitle, msg.Title)
			}

			if msg.SeverityLevel != tt.expectedSeverity {
				t.Errorf("Expected severity %q, got %q", tt.expectedSeverity, msg.SeverityLevel)
			}

			if tt.expectSuggestion && msg.Suggestion == "" {
				t.Error("Expected suggestion but got empty string")
			}

			if !tt.expectSuggestion && msg.Suggestion != "" {
				t.Errorf("Expected no suggestion but got %q", msg.Suggestion)
			}

			// Basic validation that message is not empty
			if msg.Message == "" {
				t.Error("Expected non-empty message")
			}
		})
	}
}

func TestGetUserFriendlyMessageWithContext(t *testing.T) {
	reactivationInfo := &ReactivationInfo{
		Count:         3,
		Limit:         5,
		Remaining:     2,
		NextResetDate: time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name              string
		err               error
		reactivationInfo  *ReactivationInfo
		expectEnhancement bool
	}{
		{
			name:              "Reactivation limit with context",
			err:               errors.New("reactivation limit exceeded"),
			reactivationInfo:  reactivationInfo,
			expectEnhancement: true,
		},
		{
			name:              "Already activated with remaining reactivations",
			err:               errors.New("already activated on different device"),
			reactivationInfo:  reactivationInfo,
			expectEnhancement: true,
		},
		{
			name:              "Regular error without context",
			err:               errors.New("network timeout"),
			reactivationInfo:  nil,
			expectEnhancement: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgWithContext := GetUserFriendlyMessageWithContext(tt.err, tt.reactivationInfo)
			msgWithoutContext := GetUserFriendlyMessage(tt.err)

			if tt.expectEnhancement {
				// The message or suggestion should be different/enhanced
				if msgWithContext.Message == msgWithoutContext.Message &&
					msgWithContext.Suggestion == msgWithoutContext.Suggestion {
					t.Error("Expected message enhancement but messages are identical")
				}
			}

			// Basic validation
			if msgWithContext.Title == "" {
				t.Error("Expected non-empty title")
			}
		})
	}
}

func TestFormatSuccessMessage(t *testing.T) {
	tests := []struct {
		name             string
		info             *LicenseSuccessInfo
		expectedSeverity string
		expectWarning    bool
	}{
		{
			name: "New activation with plenty of time",
			info: &LicenseSuccessInfo{
				ExpiryDate:     time.Now().AddDate(0, 6, 0), // 6 months from now
				DaysRemaining:  180,
				IsReactivation: false,
				ReactivationInfo: &ReactivationInfo{
					Count:     0,
					Limit:     5,
					Remaining: 5,
				},
			},
			expectedSeverity: "success",
			expectWarning:    false,
		},
		{
			name: "Reactivation with warning threshold",
			info: &LicenseSuccessInfo{
				ExpiryDate:     time.Now().AddDate(0, 3, 0), // 3 months from now
				DaysRemaining:  90,
				IsReactivation: true,
				ReactivationInfo: &ReactivationInfo{
					Count:     3,
					Limit:     5,
					Remaining: 2,
				},
			},
			expectedSeverity: "warning",
			expectWarning:    true,
		},
		{
			name: "All reactivations used",
			info: &LicenseSuccessInfo{
				ExpiryDate:     time.Now().AddDate(0, 1, 0), // 1 month from now
				DaysRemaining:  30,
				IsReactivation: true,
				ReactivationInfo: &ReactivationInfo{
					Count:         5,
					Limit:         5,
					Remaining:     0,
					NextResetDate: time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedSeverity: "warning",
			expectWarning:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := FormatSuccessMessage(tt.info)

			if msg.SeverityLevel != tt.expectedSeverity {
				t.Errorf("Expected severity %q, got %q", tt.expectedSeverity, msg.SeverityLevel)
			}

			if tt.expectWarning && msg.Suggestion == "" {
				t.Error("Expected warning/suggestion but got empty string")
			}

			// Validate that reactivation info is mentioned if provided
			if tt.info.IsReactivation && tt.info.ReactivationInfo != nil {
				if !containsSubstring(msg.Message, "Reactivations used") {
					t.Error("Expected reactivation information in message")
				}
			}

			// Validate expiry information is included
			if !tt.info.ExpiryDate.IsZero() {
				if !containsSubstring(msg.Message, "Expiry:") {
					t.Error("Expected expiry information in message")
				}
			}
		})
	}
}

func TestCalculateReactivationInfo(t *testing.T) {
	tests := []struct {
		name              string
		currentCount      int
		maxCount          int
		expectedRemaining int
	}{
		{
			name:              "No reactivations used",
			currentCount:      0,
			maxCount:          5,
			expectedRemaining: 5,
		},
		{
			name:              "Some reactivations used",
			currentCount:      2,
			maxCount:          5,
			expectedRemaining: 3,
		},
		{
			name:              "All reactivations used",
			currentCount:      5,
			maxCount:          5,
			expectedRemaining: 0,
		},
		{
			name:              "Exceeded (should cap at 0)",
			currentCount:      7,
			maxCount:          5,
			expectedRemaining: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := CalculateReactivationInfo(tt.currentCount, tt.maxCount)

			if info.Count != tt.currentCount {
				t.Errorf("Expected count %d, got %d", tt.currentCount, info.Count)
			}

			if info.Limit != tt.maxCount {
				t.Errorf("Expected limit %d, got %d", tt.maxCount, info.Limit)
			}

			if info.Remaining != tt.expectedRemaining {
				t.Errorf("Expected remaining %d, got %d", tt.expectedRemaining, info.Remaining)
			}

			// Verify next reset date is set to next month
			now := time.Now()
			expectedNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
			if !info.NextResetDate.Equal(expectedNextMonth) {
				t.Errorf("Expected next reset date %v, got %v", expectedNextMonth, info.NextResetDate)
			}
		})
	}
}

func TestGetLicenseStatusMessage(t *testing.T) {
	tests := []struct {
		name             string
		license          *LicenseInfo
		expectedSeverity string
	}{
		{
			name:             "Nil license",
			license:          nil,
			expectedSeverity: "warning",
		},
		{
			name: "Active license with plenty of time",
			license: &LicenseInfo{
				Status:     "active",
				ExpiryDate: time.Now().AddDate(0, 6, 0), // 6 months from now
			},
			expectedSeverity: "success",
		},
		{
			name: "Active license expiring soon",
			license: &LicenseInfo{
				Status:     "active",
				ExpiryDate: time.Now().AddDate(0, 0, 5), // 5 days from now
			},
			expectedSeverity: "warning",
		},
		{
			name: "Expired license",
			license: &LicenseInfo{
				Status:     "expired",
				ExpiryDate: time.Now().AddDate(0, 0, -10), // 10 days ago
			},
			expectedSeverity: "error",
		},
		{
			name: "Revoked license",
			license: &LicenseInfo{
				Status:     "revoked",
				ExpiryDate: time.Now().AddDate(0, 3, 0),
			},
			expectedSeverity: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := GetLicenseStatusMessage(tt.license)

			if msg.SeverityLevel != tt.expectedSeverity {
				t.Errorf("Expected severity %q, got %q", tt.expectedSeverity, msg.SeverityLevel)
			}

			if msg.Title == "" {
				t.Error("Expected non-empty title")
			}

			if msg.Message == "" {
				t.Error("Expected non-empty message")
			}
		})
	}
}

// Helper function to check if a string contains a substring (simple implementation)
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}
