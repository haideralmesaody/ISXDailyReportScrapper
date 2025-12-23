package license

import "context"

// ValidationService defines the contract for validating licenses and retrieving status.
type ValidationService interface {
	Validate(ctx context.Context) (bool, error)
	Status(ctx context.Context) (*LicenseInfo, string, error)
	Renewal(ctx context.Context) (*RenewalInfo, error)
}

// DefaultValidationService is the production implementation backed by a ManagerInterface.
type DefaultValidationService struct {
	Manager ManagerInterface
}

// NewValidationService builds a ValidationService backed by a ManagerInterface.
func NewValidationService(manager ManagerInterface) ValidationService {
	return &DefaultValidationService{Manager: manager}
}

// Validate performs license validation, preferring context-aware methods.
func (s *DefaultValidationService) Validate(ctx context.Context) (bool, error) {
	type ctxValidator interface {
		ValidateLicenseWithContext(context.Context) (bool, error)
	}

	if m, ok := s.Manager.(ctxValidator); ok {
		return m.ValidateLicenseWithContext(ctx)
	}

	return s.Manager.ValidateLicense()
}

// Status returns the current license status with comprehensive info.
func (s *DefaultValidationService) Status(ctx context.Context) (*LicenseInfo, string, error) {
	return s.Manager.GetLicenseStatus()
}

// Renewal returns renewal information.
func (s *DefaultValidationService) Renewal(ctx context.Context) (*RenewalInfo, error) {
	return s.Manager.CheckRenewalStatus()
}
