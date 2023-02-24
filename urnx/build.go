package urnx

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Build create a new URN with the specified fields
func Build(namespace string, resourceType string, resourceID uuid.UUID) (*URN, error) {
	ns := validateNamespace(namespace)
	if !ns {
		return nil, ErrInvalidURNNamespace
	}

	rt := validateResourceType(resourceType)
	if !rt {
		return nil, ErrInvalidURNResourceType
	}

	u := &URN{
		Namespace:    strings.ToLower(namespace),
		ResourceType: strings.ToLower(resourceType),
		ResourceID:   resourceID,
	}

	return u, nil
}

// String returns the string representation of the URN
func (u *URN) String() string {
	return fmt.Sprintf("%s:%s:%s:%s", prefix, u.Namespace, u.ResourceType, u.ResourceID)
}
