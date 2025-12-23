package license

import "context"

// ActivationService defines the contract for activating licenses.
// It delegates to the existing ManagerInterface to avoid duplicating logic.
type ActivationService interface {
	Activate(ctx context.Context, licenseKey string) error
}

// DefaultActivationService is the production implementation that wraps a ManagerInterface.
type DefaultActivationService struct {
	manager ManagerInterface
}

// NewActivationService builds an ActivationService backed by a ManagerInterface.
func NewActivationService(manager ManagerInterface) ActivationService {
	return &DefaultActivationService{manager: manager}
}

// Activate performs license activation, preferring context-aware activation if available.
func (s *DefaultActivationService) Activate(ctx context.Context, licenseKey string) error {
	type ctxActivator interface {
		ActivateLicenseWithContext(context.Context, string) error
	}

	if m, ok := s.manager.(ctxActivator); ok {
		return m.ActivateLicenseWithContext(ctx, licenseKey)
	}

	return s.manager.ActivateLicense(licenseKey)
}
