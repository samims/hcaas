package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

func NewInternal(format string, a ...interface{}) error {
	return fmt.Errorf("INTERNAL: "+format, a...)
}

func NewNotFound(format string, a ...interface{}) error {
	return fmt.Errorf("NOT FOUND: "+format, a...)
}

func NewConflict(format string, a ...interface{}) error {
	return fmt.Errorf("CONFLICT: "+format, a...)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

func IsInternal(err error) bool {
	return err != nil && !IsNotFound(err) && !IsConflict(err)
}
