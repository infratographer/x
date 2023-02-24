package urnx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"math/rand"
	"time"
)

type vcase struct {
	name        string
	fieldName   string
	expectError bool
}

func TestValidateNamespace(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	tc := []vcase{
		{
			name:        "valid-namespace",
			fieldName:   "valid-namespace",
			expectError: false,
		},
		{
			name:        "invalid-namespace",
			fieldName:   "invalid-namespace!",
			expectError: true,
		},
		{
			name:        "invalid-namespace-too-long",
			fieldName:   randomString(31),
			expectError: true,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			check := validateNamespace(c.fieldName)

			if c.expectError {
				assert.False(t, check)
			} else {
				assert.True(t, check)
			}
		})
	}
}

func TestValidateResourceType(t *testing.T) {
	tc := []vcase{
		{
			name:        "valid-resource-type",
			fieldName:   "valid-resource-type",
			expectError: false,
		},
		{
			name:        "invalid-resource-type",
			fieldName:   "invalid-resource-type!",
			expectError: true,
		},
		{
			name:        "invalid-resource-type-too-long",
			fieldName:   randomString(256),
			expectError: true,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			check := validateResourceType(c.fieldName)
			fmt.Println(check)
			if c.expectError {
				assert.False(t, check)
			} else {
				assert.True(t, check)
			}
		})
	}
}

func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
