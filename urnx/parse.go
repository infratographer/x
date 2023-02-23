package urnx

import (
	"strings"

	"github.com/google/uuid"
)

func Parse(urn string) (*URN, error) {
	conv := strings.Split(urn, ":")

	if conv[0] != PREFIX {
		return nil, ErrInvalidURNPrefix
	}

	if len(conv) != URNLENGTH {
		return nil, ErrInvalidURN
	}

	u := &URN{
		Prefix:       conv[0],
		Namespace:    conv[1],
		ResourceType: conv[2],
		ResourceID:   uuid.MustParse(conv[3]),
	}

	return u, nil
}

