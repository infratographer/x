package urnx

import (
	"strings"

	"github.com/google/uuid"
)

// Parse parses a string into a URN object
func Parse(urn string) (*URN, error) {
	conv := strings.Split(urn, ":")

	if conv[0] != prefix {
		return nil, ErrInvalidURNPrefix
	}

	ns, err := validateNamespace(conv[1])
	if err != nil || !ns {
		return nil, ErrInvalidURNNamespace
	}

	rt, err := validateResourceType(conv[2])
	if err != nil || !rt {
		return nil, ErrInvalidURNResourceType
	}

	if len(conv) != urnLength {
		return nil, ErrInvalidURN
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
