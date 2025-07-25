package middleware

// LicenseManagerInterface defines the interface for license validation
// This allows for easier testing and decoupling from the concrete implementation
type LicenseManagerInterface interface {
	ValidateLicense() (bool, error)
}