package urnx

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type pcase struct {
	name           string
	urn            string
	expectError    bool
	expectedErrors []error
}

func TestParse(t *testing.T) {
	pc := []pcase{
		{
			name:           "valid-parse",
			urn:            "urn:namespace:resource-type:" + uuid.New().String(),
			expectError:    false,
			expectedErrors: []error{},
		},
		{
			name:        "invalid-uuid",
			urn:         "urn:namespace:resource-type:invalid-uuid",
			expectError: true,
			expectedErrors: []error{
				ErrInvalidURNResourceID,
			},
		},
		{
			name:        "invalid-prefix",
			urn:         "invalid-prefix:namespace:resource-type:" + uuid.New().String(),
			expectError: true,
			expectedErrors: []error{
				ErrInvalidURNPrefix,
			},
		},
		{
			name:        "too-many-fields",
			urn:         "urn:namespace:resource-type:" + uuid.New().String() + ":extra-field",
			expectError: true,
			expectedErrors: []error{
				ErrInvalidURN,
			},
		},
		{
			name:        "too-few-fields",
			urn:         "urn:namespace:resource-type",
			expectError: true,
			expectedErrors: []error{
				ErrInvalidURN,
			},
		},
		{
			name:        "invalid-separator",
			urn:         "urn-namespace-resource-type:" + uuid.New().String() + "-extra-field",
			expectError: true,
			expectedErrors: []error{
				ErrInvalidURN,
			},
		},
		{
			name:        "invalid-namespace",
			urn:         "urn:invalid-namespace!:resource-type:" + uuid.New().String(),
			expectError: true,
			expectedErrors: []error{
				ErrInvalidURNNamespace,
			},
		},
		{
			name:        "invalid-resource-type",
			urn:         "urn:namespace:invalid-resource-type!:" + uuid.New().String(),
			expectError: true,
			expectedErrors: []error{
				ErrInvalidURNResourceType,
			},
		},
	}

	for _, c := range pc {
		t.Run(c.name, func(t *testing.T) {
			_, err := Parse(c.urn)

			if c.expectError {
				assert.Error(t, err)
				assert.EqualError(t, err, c.expectedErrors[0].Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
