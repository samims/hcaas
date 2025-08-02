package errors

import "errors"

var (
	ErrInternal        = errors.New("internal error")
	ErrNotFound        = errors.New("resource not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrConflict        = errors.New("resource already exists")
	ErrTokenGeneration = errors.New("token generation failed ")
	ErrInvalidEmail    = errors.New("invalid Email")
	ErrInvalidInput    = errors.New("invalid input")
)
