package errors

import "errors"

var (
	ErrInternal     = errors.New("internal error")
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("resource already exists")
)
