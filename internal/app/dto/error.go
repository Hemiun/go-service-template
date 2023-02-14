package dto

import (
	"errors"
)

type (
	// ErrorDTO represents a common  error dto
	// from this object we can extract certain information about the error
	ErrorDTO struct { //nolint:errname
		Cause    error  `json:"-"`
		Message  string `json:"message,omitempty"`
		TechInfo string `json:"techInfo,omitempty"`
	}
)

// Error implements error interface
func (e *ErrorDTO) Error() string {
	return e.Message
}

// Unwrap returns cause for this error
// it implements interface for errors.Unwrap() function
func (e *ErrorDTO) Unwrap() error {
	return e.Cause
}

var (
	// ErrValidation используется для ошибок найденных в процессе валидации
	ErrValidation = errors.New("validation error")

	// ErrEntityNotFound - сущность не найдена
	ErrEntityNotFound = errors.New("entity not found")
)
