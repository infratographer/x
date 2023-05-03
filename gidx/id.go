package gidx

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jaevor/go-nanoid"
)

const (
	// PrefixPartLength is the number of characters expected in a prefix
	PrefixPartLength = 7
	// IDPartLength is the number of characters passed to nanoid when generating an ID value
	IDPartLength = 21
	// Parts represents how many parts of an ID there are
	Parts = 2
	// TotalLength is the length of a idx generated PrefixID
	TotalLength = PrefixPartLength + IDPartLength + Parts - 1
)

// PrefixedID represents an ID that is formatted as prefix-id. PrefixedIDs are used
// to implement the relay spec for graphql, which required all IDs to be globally
// unique between objects and that you have the ability to resolve an object
// with only the id. Prefixed IDs make it possible by using a 7 characters long
// prefix with the first 4 characters representing the application the ID belongs
// to and the next 3 characters representing the object. This makes it possible
// to receive an ID and programatically tell you the object type the ID represents.
//
// Examples scenario: instance-api uses the 4 character prefix of inst and has an
// object type of instance. The 3 character code for instance is anc, so combined
// the prefix is instanc, resulting an in instance having an id that looks like
// instanc-myrandomidvalue.
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

// NewID will return a new PrefixedID with the given prefix and a generated ID value.
// The ID value will be a 21 character nanoID value.
func NewID(prefix string) (PrefixedID, error) {
	prefix = strings.ToLower(prefix)
	if len(prefix) != PrefixPartLength {
		return "", newErrInvalidID(fmt.Sprintf("expected prefix length is %d, '%s' is %d", PrefixPartLength, prefix, len(prefix)))
	}

	id, err := newIDValue()
	if err != nil {
		return "", err
	}

	return PrefixedID(fmt.Sprintf("%s-%s", prefix, id)), nil
}

func newIDValue() (string, error) {
	id, err := nanoid.Standard(IDPartLength)
	if err != nil {
		return "", err
	}

	return id(), nil
}

func parts(str string) (string, string) {
	p := strings.SplitN(string(str), "-", Parts)

	if len(p) != Parts && len(p[0]) != PrefixPartLength {
		return "", ""
	}

	return p[0], p[1]
}

// Parse reads in a string and returns a PrefixedID if the string is a properly
// formatted PrefixedID value
func Parse(str string) (PrefixedID, error) {
	prefix, id := parts(str)

	if prefix == "" || id == "" {
		return "", newErrInvalidID(fmt.Sprintf("expected id format is prefix-id, but received %s", str))
	}

	if len(prefix) != PrefixPartLength {
		return "", newErrInvalidID(fmt.Sprintf("expected prefix length is %d, '%s' is %d", PrefixPartLength, prefix, len(prefix)))
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
		return ErrNilValue
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
var _ driver.Valuer = PrefixedID("")
var _ sql.Scanner = (*PrefixedID)(nil)
