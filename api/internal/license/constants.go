package license

// Action constants for Apps Script API
// These must match the action names expected by the Google Apps Script
const (
	// ActionActivate is used when activating a new license
	ActionActivate = "activate"

	// ActionValidate is used to validate an existing license
	ActionValidate = "validate"

	// ActionRevoke is used to revoke/deactivate a license
	ActionRevoke = "revoke"

	// ActionCheckStatus is used to check the status of a license
	ActionCheckStatus = "checkStatus"

	// ActionPing is used for connectivity testing
	ActionPing = "ping"
)

// Response status constants
const (
	StatusActivated                 = "activated"
	StatusReactivated               = "reactivated"
	StatusValid                     = "valid"
	StatusExpired                   = "expired"
	StatusAlreadyActivated          = "already_activated"
	StatusAlreadyActivatedDifferent = "already_activated_different_device"
	StatusReactivationBlocked       = "reactivation_blocked"
	StatusInvalid                   = "invalid"
)
