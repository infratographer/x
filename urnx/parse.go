package urnx

import (
	"strings"

	"github.com/google/uuid"
)

// Parse parses a string into a URN object
func Parse(urn string) (*URN, error) {
	conv := strings.Split(urn, ":")

	if len(conv) != urnLength {
		return nil, ErrInvalidURN
	}

	if conv[0] != prefix {
		return nil, ErrInvalidURNPrefix
	}

	ns := validateNamespace(conv[1])
	if !ns {
		return nil, ErrInvalidURNNamespace
	}

	rt := validateResourceType(conv[2])
	if !rt {
		return nil, ErrInvalidURNResourceType
	}

	id, err := uuid.Parse(conv[3])
	if err != nil {
		return nil, ErrInvalidURNResourceID
	}

	u := &URN{
		Namespace:    conv[1],
		ResourceType: conv[2],
		ResourceID:   id,
	}

	return u, nil
}
