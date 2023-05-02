package gidx

import (
	"errors"
	"fmt"
)

var (
	// ErrNilValue is returned when a nil value is provided
	ErrNilValue = errors.New("no value provided")

	// ErrUnsupportedType is returned when a value is provided of an unsupported type
	ErrUnsupportedType = errors.New("unsupported type")
)

// ErrInvalidID is returned when a provided ID value is invalid
type ErrInvalidID struct {
	msg string
}

func (e *ErrInvalidID) Error() string {
	return fmt.Sprintf("invalid id: %s", e.msg)
}

func newErrInvalidID(s string) error {
	return &ErrInvalidID{msg: s}
}
