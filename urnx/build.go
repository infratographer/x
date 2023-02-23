package urnx

import (
	"fmt"

	"github.com/google/uuid"
)

// Build create a new URN with the specified fields
func Build(namespace string, resourceType string, resourceID uuid.UUID) *URN {
	u := &URN{
		Prefix:       prefix,
		Namespace:    namespace,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}

	return u
}

// String returns the string representation of the URN
func (u *URN) String() string {
	return fmt.Sprintf("%s:%s:%s:%s", u.Prefix, u.Namespace, u.ResourceType, u.ResourceID)
}
