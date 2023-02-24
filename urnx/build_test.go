package urnx

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type bcase struct {
	name           string
	namespace      string
	resourceType   string
	resourceID     uuid.UUID
	expectError    bool
	expectedErrors []error
}

func TestBuild(t *testing.T) {
	t.Parallel()

	bc := []bcase{
		{
			name:           "valid-build",
			namespace:      "namespace",
			resourceType:   "resource-type",
			resourceID:     uuid.New(),
			expectError:    false,
			expectedErrors: []error{},
		},
		{
			name:         "invalid-namespace",
			namespace:    "invalid-namespace!",
			resourceType: "resource-type",
			resourceID:   uuid.New(),
			expectError:  true,
			expectedErrors: []error{
				ErrInvalidURNNamespace,
			},
		},
		{
			name:         "invalid-resource-type",
			namespace:    "namespace",
			resourceType: "invalid-resource-type!",
			resourceID:   uuid.New(),
			expectError:  true,
			expectedErrors: []error{
				ErrInvalidURNResourceType,
			},
		},
	}

	for _, c := range bc {
		t.Run(c.name, func(t *testing.T) {
			urn, err := Build(c.namespace, c.resourceType, c.resourceID)

			if c.expectError {
				assert.Error(t, err)
				assert.EqualError(t, err, c.expectedErrors[0].Error())
			} else {
				assert.NoError(t, err)
				expectedURN := &URN{Namespace: c.namespace, ResourceType: c.resourceType, ResourceID: c.resourceID}
				assert.Equal(t, expectedURN, urn)
			}
		})
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	urn := &URN{Namespace: "namespace", ResourceType: "resource-type", ResourceID: uuid.New()}
	assert.Equal(t, fmt.Sprintf("%s:%s:%s:%s", prefix, urn.Namespace, urn.ResourceType, urn.ResourceID), urn.String())
}
