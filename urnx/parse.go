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

	if len(conv) != urnLength {
		return nil, ErrInvalidURN
	}

	u := &URN{
		Namespace:    conv[1],
		ResourceType: conv[2],
		ResourceID:   uuid.MustParse(conv[3]),
	}

	return u, nil
}
