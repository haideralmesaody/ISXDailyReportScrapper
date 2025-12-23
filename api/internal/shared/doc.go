// Package shared provides common utilities and test helpers used across the ISX codebase.
// It serves as a central location for shared functionality that doesn't belong to any
// specific domain or architectural layer.
//
// # Structure
//
// The package is organized into the following components:
//
// - testutil: Testing utilities including fixtures, mocks, and assertions
// - helpers: Common helper functions used across packages
//
// # Usage Guidelines
//
// This package should only contain:
//
// 1. Test utilities used by multiple packages
// 2. Generic helper functions with no domain-specific logic
// 3. Common constants or types used across packages
//
// It should NOT contain:
//
// 1. Business logic or domain-specific code
// 2. External dependencies beyond standard library
// 3. Circular dependencies with other internal packages
//
// # Test Utilities
//
// The testutil subpackage provides:
//
//	- Fixture generators for common test data
//	- Mock implementations of interfaces
//	- Custom assertions for domain objects
//	- Test database setup and teardown helpers
//
// Example usage:
//
//	func TestSomething(t *testing.T) {
//	    fixture := shared.NewTestFixture()
//	    defer fixture.Cleanup()
//	    
//	    // Use fixture in tests
//	}
//
// # Performance
//
// All utilities in this package are designed to have minimal overhead
// and should not impact production performance when used correctly.
package shared