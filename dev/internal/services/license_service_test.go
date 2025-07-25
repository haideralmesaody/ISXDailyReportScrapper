package services

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"isxcli/internal/license"
)

// MockLicenseManager is a mock implementation of license.Manager
type MockLicenseManager struct {
	mock.Mock
}

func (m *MockLicenseManager) GetLicenseStatus() (*license.LicenseInfo, string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*license.LicenseInfo), args.String(1), args.Error(2)
}

func (m *MockLicenseManager) ActivateLicense(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockLicenseManager) ValidateLicense() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *MockLicenseManager) ReadLicenseFile() (*license.LicenseInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*license.LicenseInfo), args.Error(1)
}

func (m *MockLicenseManager) WriteLicenseFile(info *license.LicenseInfo) error {
	args := m.Called(info)
	return args.Error(0)
}

func (m *MockLicenseManager) GetMachineID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockLicenseManager) IsLicenseValid(info *license.LicenseInfo) bool {
	args := m.Called(info)
	return args.Bool(0)
}

func TestLicenseService_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockLicenseManager)
		expectedStatus string
		expectedError  bool
		checkResponse  func(*testing.T, *LicenseStatusResponse)
	}{
		{
			name: "not activated license",
			mockSetup: func(m *MockLicenseManager) {
				m.On("GetLicenseStatus").Return(nil, "Not Activated", nil)
			},
			expectedStatus: "not_activated",
			expectedError:  false,
			checkResponse: func(t *testing.T, resp *LicenseStatusResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, "not_activated", resp.LicenseStatus)
				assert.Equal(t, "No license activated. Please activate a license to use this application.", resp.Message)
				assert.Equal(t, "/license/not-activated", resp.Type)
				assert.Nil(t, resp.LicenseInfo)
			},
		},
		{
			name: "active license with 30 days left",
			mockSetup: func(m *MockLicenseManager) {
				info := &license.LicenseInfo{
					LicenseKey:  "TEST-KEY-123",
					ExpiryDate:  time.Now().Add(30 * 24 * time.Hour),
					MachineID:   "MACHINE-123",
					CompanyName: "Test Company",
				}
				m.On("GetLicenseStatus").Return(info, "Active", nil)
			},
			expectedStatus: "warning",
			expectedError:  false,
			checkResponse: func(t *testing.T, resp *LicenseStatusResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, "warning", resp.LicenseStatus)
				assert.Equal(t, 30, resp.DaysLeft)
				assert.Contains(t, resp.Message, "expires in 30 days")
				assert.NotNil(t, resp.LicenseInfo)
			},
		},
		{
			name: "active license with 7 days left (critical)",
			mockSetup: func(m *MockLicenseManager) {
				info := &license.LicenseInfo{
					LicenseKey:  "TEST-KEY-123",
					ExpiryDate:  time.Now().Add(7 * 24 * time.Hour),
					MachineID:   "MACHINE-123",
					CompanyName: "Test Company",
				}
				m.On("GetLicenseStatus").Return(info, "Critical", nil)
			},
			expectedStatus: "critical",
			expectedError:  false,
			checkResponse: func(t *testing.T, resp *LicenseStatusResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, "critical", resp.LicenseStatus)
				assert.Equal(t, 7, resp.DaysLeft)
				assert.Contains(t, resp.Message, "expires in 7 days")
				assert.Contains(t, resp.Message, "renew soon")
			},
		},
		{
			name: "expired license",
			mockSetup: func(m *MockLicenseManager) {
				info := &license.LicenseInfo{
					LicenseKey:  "TEST-KEY-123",
					ExpiryDate:  time.Now().Add(-1 * 24 * time.Hour),
					MachineID:   "MACHINE-123",
					CompanyName: "Test Company",
				}
				m.On("GetLicenseStatus").Return(info, "Expired", nil)
			},
			expectedStatus: "expired",
			expectedError:  false,
			checkResponse: func(t *testing.T, resp *LicenseStatusResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, "expired", resp.LicenseStatus)
				assert.Contains(t, resp.Message, "expired")
				assert.Contains(t, resp.Message, "renew")
			},
		},
		{
			name: "error getting license status",
			mockSetup: func(m *MockLicenseManager) {
				m.On("GetLicenseStatus").Return(nil, "", errors.New("database error"))
			},
			expectedStatus: "error",
			expectedError:  false,
			checkResponse: func(t *testing.T, resp *LicenseStatusResponse) {
				assert.Equal(t, 500, resp.Status)
				assert.Equal(t, "error", resp.LicenseStatus)
				assert.Equal(t, "/errors/license-check-failed", resp.Type)
				assert.Contains(t, resp.Message, "Unable to retrieve license information")
			},
		},
		{
			name: "valid license with many days left",
			mockSetup: func(m *MockLicenseManager) {
				info := &license.LicenseInfo{
					LicenseKey:  "TEST-KEY-123",
					ExpiryDate:  time.Now().Add(365 * 24 * time.Hour),
					MachineID:   "MACHINE-123",
					CompanyName: "Test Company",
				}
				m.On("GetLicenseStatus").Return(info, "Active", nil)
			},
			expectedStatus: "active",
			expectedError:  false,
			checkResponse: func(t *testing.T, resp *LicenseStatusResponse) {
				assert.Equal(t, 200, resp.Status)
				assert.Equal(t, "active", resp.LicenseStatus)
				assert.Equal(t, 365, resp.DaysLeft)
				assert.Contains(t, resp.Message, "License is active")
				assert.Contains(t, resp.Message, "365 days remaining")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockManager := new(MockLicenseManager)
			tt.mockSetup(mockManager)
			
			logger := slog.New(slog.NewTextHandler(nil, nil))
			service := NewLicenseService(mockManager, logger)
			
			ctx := context.Background()
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-trace-id")
			
			// Execute
			resp, err := service.GetStatus(ctx)
			
			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				tt.checkResponse(t, resp)
				assert.Equal(t, "test-trace-id", resp.TraceID)
			}
			
			mockManager.AssertExpectations(t)
		})
	}
}

func TestLicenseService_Activate(t *testing.T) {
	tests := []struct {
		name          string
		licenseKey    string
		mockSetup     func(*MockLicenseManager)
		expectedError bool
		errorContains string
	}{
		{
			name:       "successful activation",
			licenseKey: "VALID-LICENSE-KEY",
			mockSetup: func(m *MockLicenseManager) {
				m.On("ActivateLicense", "VALID-LICENSE-KEY").Return(nil)
			},
			expectedError: false,
		},
		{
			name:       "invalid license key",
			licenseKey: "INVALID-KEY",
			mockSetup: func(m *MockLicenseManager) {
				m.On("ActivateLicense", "INVALID-KEY").Return(errors.New("invalid license key"))
			},
			expectedError: true,
			errorContains: "activation failed",
		},
		{
			name:       "network error during activation",
			licenseKey: "VALID-KEY",
			mockSetup: func(m *MockLicenseManager) {
				m.On("ActivateLicense", "VALID-KEY").Return(errors.New("network error"))
			},
			expectedError: true,
			errorContains: "activation failed",
		},
		{
			name:       "expired license key",
			licenseKey: "EXPIRED-KEY",
			mockSetup: func(m *MockLicenseManager) {
				m.On("ActivateLicense", "EXPIRED-KEY").Return(errors.New("license expired"))
			},
			expectedError: true,
			errorContains: "activation failed",
		},
		{
			name:       "machine mismatch",
			licenseKey: "VALID-KEY",
			mockSetup: func(m *MockLicenseManager) {
				m.On("ActivateLicense", "VALID-KEY").Return(errors.New("machine mismatch"))
			},
			expectedError: true,
			errorContains: "activation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockManager := new(MockLicenseManager)
			tt.mockSetup(mockManager)
			
			logger := slog.New(slog.NewTextHandler(nil, nil))
			service := NewLicenseService(mockManager, logger)
			
			ctx := context.Background()
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-trace-id")
			
			// Execute
			err := service.Activate(ctx, tt.licenseKey)
			
			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
			
			mockManager.AssertExpectations(t)
		})
	}
}

func TestLicenseService_ValidateWithContext(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockLicenseManager)
		expectedValid  bool
		expectedError  bool
		contextTimeout bool
	}{
		{
			name: "valid license",
			mockSetup: func(m *MockLicenseManager) {
				m.On("ValidateLicense").Return(true, nil).Maybe()
			},
			expectedValid: true,
			expectedError: false,
		},
		{
			name: "invalid license",
			mockSetup: func(m *MockLicenseManager) {
				m.On("ValidateLicense").Return(false, nil).Maybe()
			},
			expectedValid: false,
			expectedError: false,
		},
		{
			name: "validation error",
			mockSetup: func(m *MockLicenseManager) {
				m.On("ValidateLicense").Return(false, errors.New("validation error")).Maybe()
			},
			expectedValid: false,
			expectedError: true,
		},
		{
			name: "context timeout",
			mockSetup: func(m *MockLicenseManager) {
				// Simulate slow validation
				m.On("ValidateLicense").Return(true, nil).After(100 * time.Millisecond).Maybe()
			},
			expectedValid:  false,
			expectedError:  true,
			contextTimeout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockManager := new(MockLicenseManager)
			tt.mockSetup(mockManager)
			
			logger := slog.New(slog.NewTextHandler(nil, nil))
			service := NewLicenseService(mockManager, logger)
			
			ctx := context.Background()
			if tt.contextTimeout {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 10*time.Millisecond)
				defer cancel()
			}
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-trace-id")
			
			// Execute
			valid, err := service.ValidateWithContext(ctx)
			
			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.contextTimeout {
					assert.ErrorIs(t, err, context.DeadlineExceeded)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValid, valid)
			}
			
			mockManager.AssertExpectations(t)
		})
	}
}

func TestLicenseService_StatusMessageGeneration(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		daysLeft       int
		expectedStatus string
		messageContains []string
	}{
		{
			name:            "expired license",
			status:          "Expired",
			daysLeft:        -5,
			expectedStatus:  "expired",
			messageContains: []string{"expired", "renew"},
		},
		{
			name:            "critical - 3 days left",
			status:          "Critical",
			daysLeft:        3,
			expectedStatus:  "critical",
			messageContains: []string{"expires in 3 days", "renew soon"},
		},
		{
			name:            "warning - 20 days left",
			status:          "Warning",
			daysLeft:        20,
			expectedStatus:  "warning",
			messageContains: []string{"expires in 20 days", "Consider renewing"},
		},
		{
			name:            "active - 100 days left",
			status:          "Active",
			daysLeft:        100,
			expectedStatus:  "active",
			messageContains: []string{"License is active", "100 days remaining"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockManager := new(MockLicenseManager)
			expiryDate := time.Now().Add(time.Duration(tt.daysLeft) * 24 * time.Hour)
			info := &license.LicenseInfo{
				LicenseKey:  "TEST-KEY",
				ExpiryDate:  expiryDate,
				MachineID:   "MACHINE-123",
				CompanyName: "Test Company",
			}
			mockManager.On("GetLicenseStatus").Return(info, tt.status, nil)
			
			logger := slog.New(slog.NewTextHandler(nil, nil))
			service := NewLicenseService(mockManager, logger)
			
			ctx := context.Background()
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-trace-id")
			
			// Execute
			resp, err := service.GetStatus(ctx)
			
			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectedStatus, resp.LicenseStatus)
			
			for _, contains := range tt.messageContains {
				assert.Contains(t, resp.Message, contains)
			}
			
			mockManager.AssertExpectations(t)
		})
	}
}

func TestMaskLicenseKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short key",
			input:    "ABC",
			expected: "***",
		},
		{
			name:     "exact 8 chars",
			input:    "12345678",
			expected: "***",
		},
		{
			name:     "long key",
			input:    "ABC123DEF456GHI789",
			expected: "ABC123DE...",
		},
		{
			name:     "empty key",
			input:    "",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskLicenseKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}