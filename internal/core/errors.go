package core

import (
	"errors"
	"fmt"
)

// Common error variables
var (
	ErrManagerNotFound     = errors.New("package manager not found")
	ErrManagerNotAvailable = errors.New("package manager not available")
	ErrInvalidPath         = errors.New("invalid path")
	ErrInvalidConfig       = errors.New("invalid configuration")
	ErrCacheNotFound       = errors.New("cache not found")
	ErrProjectNotFound     = errors.New("project not found")
	ErrPackageNotFound     = errors.New("package not found")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrNetworkError        = errors.New("network error")
	ErrInvalidRegistry     = errors.New("invalid registry URL")
	ErrInvalidProxy        = errors.New("invalid proxy configuration")
)

// ManagerError represents an error specific to a package manager
type ManagerError struct {
	Manager string
	Op      string
	Err     error
}

func (e *ManagerError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Manager, e.Op, e.Err)
}

func (e *ManagerError) Unwrap() error {
	return e.Err
}

// NewManagerError creates a new ManagerError
func NewManagerError(manager, op string, err error) *ManagerError {
	return &ManagerError{
		Manager: manager,
		Op:      op,
		Err:     err,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' with value '%s': %s", e.Field, e.Value, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// ConfigError represents a configuration error
type ConfigError struct {
	Manager string
	Key     string
	Err     error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error for %s.%s: %v", e.Manager, e.Key, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new ConfigError
func NewConfigError(manager, key string, err error) *ConfigError {
	return &ConfigError{
		Manager: manager,
		Key:     key,
		Err:     err,
	}
}

// IsManagerError checks if an error is a ManagerError
func IsManagerError(err error) bool {
	var managerErr *ManagerError
	return errors.As(err, &managerErr)
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// IsConfigError checks if an error is a ConfigError
func IsConfigError(err error) bool {
	var configErr *ConfigError
	return errors.As(err, &configErr)
}
