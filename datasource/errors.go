package datasource

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure cases.
var (
	// ErrNotFound indicates a resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrConnectionFailed indicates a connection could not be established.
	ErrConnectionFailed = errors.New("connection failed")

	// ErrQueryFailed indicates a query execution failed.
	ErrQueryFailed = errors.New("query failed")

	// ErrConfigInvalid indicates the configuration is invalid.
	ErrConfigInvalid = errors.New("invalid configuration")

	// ErrProviderNotFound indicates the requested provider is not registered.
	ErrProviderNotFound = errors.New("provider not found")

	// ErrReadOnlyViolation indicates a write operation was attempted on a read-only connection.
	ErrReadOnlyViolation = errors.New("write operation not allowed on read-only connection")

	// ErrQueryTimeout indicates a query exceeded its timeout.
	ErrQueryTimeout = errors.New("query timeout")

	// ErrConnectionClosed indicates the connection is closed.
	ErrConnectionClosed = errors.New("connection closed")
)

// ConnectionError wraps connection-related errors with additional context.
type ConnectionError struct {
	Provider string
	Host     string
	Port     int
	Database string
	Err      error
}

func (e *ConnectionError) Error() string {
	if e.Host != "" {
		return fmt.Sprintf("connection to %s://%s:%d/%s failed: %v",
			e.Provider, e.Host, e.Port, e.Database, e.Err)
	}
	return fmt.Sprintf("connection to %s failed: %v", e.Provider, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// QueryError wraps query-related errors with additional context.
type QueryError struct {
	Query string
	Err   error
}

func (e *QueryError) Error() string {
	// Truncate query for logging if too long
	query := e.Query
	if len(query) > 100 {
		query = query[:100] + "..."
	}
	return fmt.Sprintf("query failed: %v (query: %s)", e.Err, query)
}

func (e *QueryError) Unwrap() error {
	return e.Err
}

// ConfigError wraps configuration validation errors.
type ConfigError struct {
	Field   string
	Message string
	Err     error
}

func (e *ConfigError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("config error: %s: %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("config error: %s: %s", e.Field, e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConnectionError creates a new ConnectionError.
func NewConnectionError(provider, host string, port int, database string, err error) error {
	return &ConnectionError{
		Provider: provider,
		Host:     host,
		Port:     port,
		Database: database,
		Err:      err,
	}
}

// NewQueryError creates a new QueryError.
func NewQueryError(query string, err error) error {
	return &QueryError{
		Query: query,
		Err:   err,
	}
}

// NewConfigError creates a new ConfigError.
func NewConfigError(field, message string, err error) error {
	return &ConfigError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// IsConnectionError returns true if the error is a ConnectionError.
func IsConnectionError(err error) bool {
	var ce *ConnectionError
	return errors.As(err, &ce)
}

// IsQueryError returns true if the error is a QueryError.
func IsQueryError(err error) bool {
	var qe *QueryError
	return errors.As(err, &qe)
}

// IsConfigError returns true if the error is a ConfigError.
func IsConfigError(err error) bool {
	var ce *ConfigError
	return errors.As(err, &ce)
}

// IsTimeout returns true if the error indicates a timeout.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrQueryTimeout)
}

// IsReadOnlyViolation returns true if the error indicates a read-only violation.
func IsReadOnlyViolation(err error) bool {
	return errors.Is(err, ErrReadOnlyViolation)
}
