package telegram

import (
	"errors"
	"fmt"
)

// Common telegram package errors
var (
	ErrUserNotFound = errors.New("user not found")
	ErrOTPInvalid   = errors.New("invalid OTP code")
)

// ErrInvalidConfig creates a configuration validation error
func ErrInvalidConfig(message string) error {
	return fmt.Errorf("invalid telegram config: %s", message)
}

// ErrRateLimit creates a rate limit error with details
func ErrRateLimit(resource string, limit int, window string) error {
	return fmt.Errorf("rate limit exceeded for %s: %d requests per %s", resource, limit, window)
}

// ErrTelegramAPI creates a Telegram API error
func ErrTelegramAPI(operation string, err error) error {
	return fmt.Errorf("telegram API error during %s: %w", operation, err)
}