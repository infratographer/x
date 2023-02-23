package urnx

import (
	"fmt"

	"github.com/google/uuid"
)

func Build(namespace string, resourceType string, resourceID uuid.UUID) *URN {
	u := &URN{
		Prefix:       PREFIX,
		Namespace:    namespace,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}

	return u
}

func (u *URN) String() string {
	return fmt.Sprintf("%s:%s:%s:%s", u.Prefix, u.Namespace, u.ResourceType, u.ResourceID)
}

