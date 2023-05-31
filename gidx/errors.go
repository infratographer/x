// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package gidx

import (
	"errors"
)

// ErrUnsupportedType is returned when a value is provided of an unsupported type
var ErrUnsupportedType = errors.New("unsupported type")

// ErrInvalidID is returned when a provided ID value is invalid
type ErrInvalidID struct {
	msg string
}

func (e *ErrInvalidID) Error() string {
	return "invalid id: " + e.msg
}

func newErrInvalidID(s string) error {
	return &ErrInvalidID{msg: s}
}
