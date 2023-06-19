// Copyright 2023 The Infratographer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
