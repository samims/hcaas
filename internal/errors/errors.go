package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource already exists")
)

type ErrorType string

const (
	NotFound ErrorType = "NOT_FOUND"
	Internal ErrorType = "INTERNAL"
)

type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewNotFound(msg string) *AppError {
	return &AppError{Type: NotFound, Message: msg}
}

func NewInternal(format string, args ...interface{}) *AppError {
	return &AppError{
		Type:    Internal,
		Message: fmt.Sprintf(format, args...),
	}
}

func IsNotFound(err error) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Type == NotFound
	}
	return false
}

func IsInternal(err error) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Type == Internal
	}
	return false
}
