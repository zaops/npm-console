package core

import (
	"errors"
	"testing"
)

func TestManagerError(t *testing.T) {
	tests := []struct {
		name      string
		manager   string
		operation string
		cause     error
		expected  string
	}{
		{
			name:      "basic manager error",
			manager:   "npm",
			operation: "get cache",
			cause:     errors.New("command not found"),
			expected:  "npm get cache: command not found",
		},
		{
			name:      "empty operation",
			manager:   "pnpm",
			operation: "",
			cause:     errors.New("network error"),
			expected:  "pnpm : network error",
		},
		{
			name:      "nil cause",
			manager:   "yarn",
			operation: "clear cache",
			cause:     nil,
			expected:  "yarn clear cache: <nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewManagerError(tt.manager, tt.operation, tt.cause)
			if err.Error() != tt.expected {
				t.Errorf("ManagerError.Error() = %v, want %v", err.Error(), tt.expected)
			}

			// Test fields directly since we know the type
			if err.Manager != tt.manager {
				t.Errorf("ManagerError.Manager = %v, want %v", err.Manager, tt.manager)
			}
			if err.Op != tt.operation {
				t.Errorf("ManagerError.Op = %v, want %v", err.Op, tt.operation)
			}
			if err.Err != tt.cause {
				t.Errorf("ManagerError.Err = %v, want %v", err.Err, tt.cause)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		message  string
		expected string
	}{
		{
			name:     "basic validation error",
			field:    "registry",
			value:    "invalid-url",
			message:  "invalid URL format",
			expected: "validation error for field 'registry' with value 'invalid-url': invalid URL format",
		},
		{
			name:     "empty value",
			field:    "manager",
			value:    "",
			message:  "cannot be empty",
			expected: "validation error for field 'manager' with value '': cannot be empty",
		},
		{
			name:     "empty message",
			field:    "proxy",
			value:    "http://proxy",
			message:  "",
			expected: "validation error for field 'proxy' with value 'http://proxy': ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.value, tt.message)
			if err.Error() != tt.expected {
				t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), tt.expected)
			}

			// Test fields directly since we know the type
			if err.Field != tt.field {
				t.Errorf("ValidationError.Field = %v, want %v", err.Field, tt.field)
			}
			if err.Value != tt.value {
				t.Errorf("ValidationError.Value = %v, want %v", err.Value, tt.value)
			}
			if err.Message != tt.message {
				t.Errorf("ValidationError.Message = %v, want %v", err.Message, tt.message)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrManagerNotAvailable",
			err:      ErrManagerNotAvailable,
			expected: "package manager not available",
		},
		{
			name:     "ErrProjectNotFound",
			err:      ErrProjectNotFound,
			expected: "project not found",
		},
		{
			name:     "ErrPackageNotFound",
			err:      ErrPackageNotFound,
			expected: "package not found",
		},

		{
			name:     "ErrNetworkError",
			err:      ErrNetworkError,
			expected: "network error",
		},
		{
			name:     "ErrPermissionDenied",
			err:      ErrPermissionDenied,
			expected: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Error message = %v, want %v", tt.err.Error(), tt.expected)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	originalErr := errors.New("original error")
	
	// Test ManagerError wrapping
	managerErr := NewManagerError("npm", "test", originalErr)
	if !errors.Is(managerErr, originalErr) {
		t.Error("ManagerError should wrap original error")
	}

	// Test unwrapping
	unwrapped := errors.Unwrap(managerErr)
	if unwrapped != originalErr {
		t.Errorf("Unwrapped error = %v, want %v", unwrapped, originalErr)
	}
}

func TestErrorChaining(t *testing.T) {
	// Create a chain of errors
	rootErr := errors.New("root cause")
	managerErr := NewManagerError("npm", "operation", rootErr)
	validationErr := NewValidationError("field", "value", managerErr.Error())

	// Test that we can find the root cause
	if !errors.Is(managerErr, rootErr) {
		t.Error("Should be able to find root cause in manager error")
	}

	// Test error messages contain context
	if !contains(managerErr.Error(), "npm") {
		t.Error("Manager error should contain manager name")
	}
	if !contains(managerErr.Error(), "operation") {
		t.Error("Manager error should contain operation")
	}
	if !contains(validationErr.Error(), "field") {
		t.Error("Validation error should contain field name")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
