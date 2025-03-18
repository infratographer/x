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
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jaevor/go-nanoid"
)

const (
	// IDPartLength is the number of characters passed to nanoid when generating an ID value
	IDPartLength = 21
	// PrefixPartMinLength is the minimum required number of characters required for a prefix
	PrefixPartMinLength = 2
	// Parts represents how many parts of an ID there are
	Parts = 2
	// NullPrefixedID represents a null value PrefixedID
	NullPrefixedID = PrefixedID("")
)

// PrefixRegexp is the regular expression used to validate a prefix
var PrefixRegexp = regexp.MustCompile(`^[a-z0-9]{2,}$`)

// PrefixedID represents an ID that is formatted as prefix-id. PrefixedIDs are used
// to implement the relay spec for graphql, which required all IDs to be globally
// unique between objects and that you have the ability to resolve an object
// with only the id. Prefixed IDs make it possible by using a prefix representing
// the object. This makes it possible to receive an ID and programatically tell
// you the object type the ID represents.
//
// Prefixes can be of any length greater than 2 characters, Infratographer
// projects use a prefix that follows the pattern of 4 characters representing
// the application and 3 characters representing the object type. For example,
// instance-api uses the 4 character prefix of inst and has an object type of
// instance. The 3 character code for instance is anc, so combined the prefix is
// instanc, resulting an in instance having an id that looks like instanc-myrandomidvalue.
type PrefixedID string

// Prefix will return the Prefix value of an ID
func (p PrefixedID) Prefix() string {
	prefix, _ := parts(string(p))
	return prefix
}

// MustNewID wraps NewID and panics in the event of an error
func MustNewID(prefix string) PrefixedID {
	id, err := NewID(prefix)
	if err != nil {
		panic(err)
	}

	return id
}

func validPrefix(s string) error {
	if len(s) <= PrefixPartMinLength {
		return newErrInvalidID(fmt.Sprintf("expected prefix length is at least %d, '%s' is %d", PrefixPartMinLength, s, len(s)))
	}

	if !PrefixRegexp.MatchString(s) {
		return newErrInvalidID(fmt.Sprintf("expected prefix must match %s, '%s' does not", PrefixRegexp.String(), s))
	}

	return nil
}

// NewID will return a new PrefixedID with the given prefix and a generated ID value.
// The ID value will be a 21 character nanoID value.
func NewID(prefix string) (PrefixedID, error) {
	prefix = strings.ToLower(prefix)
	if err := validPrefix(prefix); err != nil {
		return "", err
	}

	id, err := newIDValue()
	if err != nil {
		return "", err
	}

	return PrefixedID(fmt.Sprintf("%s-%s", prefix, id)), nil
}

func newIDValue() (string, error) {
	// This would only return an error if the const value for IDPartLength is
	// changed to an invalid value, must be within 2-255 per nanoid docs.
	id, err := nanoid.Standard(IDPartLength)
	if err != nil {
		return "", err
	}

	return id(), nil
}

func parts(str string) (string, string) {
	if cnt := strings.Count(str, "-"); cnt == 0 {
		return "", ""
	}

	p := strings.SplitN(string(str), "-", Parts)

	return p[0], p[1]
}

// Parse reads in a string and returns a PrefixedID if the string is a properly
// formatted PrefixedID value
func Parse(str string) (PrefixedID, error) {
	if str == "" {
		return PrefixedID(""), nil
	}

	prefix, id := parts(str)

	if prefix == "" || id == "" {
		return "", newErrInvalidID(fmt.Sprintf("expected id format is prefix-id, but received %s", str))
	}

	if err := validPrefix(prefix); err != nil {
		return "", err
	}

	// ensure the string isn't a UUID
	if _, err := uuid.Parse(str); err == nil {
		return "", newErrInvalidID("uuids are not valid prefix-ids")
	}

	return PrefixedID(str), nil
}

// String returns PrefixedID as a string.
func (p PrefixedID) String() string {
	return string(p)
}

// Value implements sql.Valuer so that PrefixedIDs can be written to databases
// transparently. PrefixedIDs map to strings.
func (p PrefixedID) Value() (driver.Value, error) {
	if _, err := Parse(string(p)); err != nil {
		return "", err
	}

	return string(p), nil
}

// Scan implements sql.Scanner so PrefixedIDs can be read from databases
// transparently. The value returned is not checked to ensure it's a
// properly formatted PrefixedID.
func (p *PrefixedID) Scan(v any) error {
	if v == nil {
		*p = PrefixedID("")
		return nil
	}

	switch src := v.(type) {
	case string:
		*p = PrefixedID(src)
	case []byte:
		*p = PrefixedID(string(src))
	case PrefixedID:
		*p = src
	default:
		return ErrUnsupportedType
	}

	return nil
}

// MarshalGQL provides GraphQL marshaling so that PrefixedIDs can be returned
// in GraphQL results transparently. Only types that map to a string are supported.
func (p PrefixedID) MarshalGQL(w io.Writer) {
	// graphql ID is a scalar which must be quoted
	io.WriteString(w, strconv.Quote(string(p))) //nolint:errcheck
}

// UnmarshalGQL provides GraphQL unmarshaling so that PrefixedIDs can be parsed
// in GraphQL requests transparently. Only input types that map to a string are supported.
func (p *PrefixedID) UnmarshalGQL(v interface{}) error {
	return p.Scan(v)
}

// Verify interfaces are satisfied
var (
	_ driver.Valuer = PrefixedID("")
	_ sql.Scanner   = (*PrefixedID)(nil)
)
