package middleware

import "context"

// LicenseManagerInterface defines the interface for license validation
// This allows for easier testing and decoupling from the concrete implementation
type LicenseManagerInterface interface {
	ValidateLicense() (bool, error)
}

type ValidationServiceInterface interface {
	Validate(ctx context.Context) (bool, error)
}
