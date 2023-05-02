package gidx_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.infratographer.com/x/gidx"
)

func TestNewID(t *testing.T) {
	// ensure the prefix length hasn't changed. if it does change these tests may need updated
	require.Equal(t, 7, gidx.PrefixPartLength)

	cases := []struct {
		name     string
		prefix   string
		errorMsg string
	}{
		{name: "corrent prefix length", prefix: "testpre"},
		{name: "prefix length too short", prefix: "short", errorMsg: "invalid id: expected prefix length is 7"},
		{name: "prefix length too long", prefix: "notthatshort", errorMsg: "invalid id: expected prefix length is 7"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			id, err := gidx.NewID(tt.prefix)
			if tt.errorMsg == "" {
				assert.NoError(t, err)
				assert.NotNil(t, id)
			} else {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.errorMsg)
			}
		})
	}
}

func TestPrefix(t *testing.T) {
	id, err := gidx.NewID("testpre")
	assert.NoError(t, err)

	assert.Equal(t, "testpre", id.Prefix())
}

func TestParsers(t *testing.T) {
	cases := []struct {
		name     string
		id       string
		errorMsg string
	}{
		{name: "valid id", id: string(gidx.MustNewID("testing"))},
		{name: "valid prefix with any length id", id: "testing-any_random#string@i*want(to-put-in-here-of-any-length"},
		{name: "valid prefix with a uuid ", id: "testing-" + uuid.New().String()},
		{name: "invalid id; no separator", id: "somestringthatisalltogether", errorMsg: "invalid id: expected id format is prefix-id"},
		{name: "invalid id; prefix length too short", id: "short-fm21VlAHHrGf6utn1JsKc", errorMsg: "invalid id: expected prefix length is 7"},
		{name: "invalid id; prefix length too long", id: "notthatshort-fm21VlAHHrGf6utn1JsKc", errorMsg: "invalid id: expected prefix length is 7"},
	}

	t.Run("Test globalid.Parse", func(t *testing.T) {
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				id, err := gidx.Parse(tt.id)
				if tt.errorMsg == "" {
					assert.NoError(t, err)
					assert.NotNil(t, id)
				} else {
					assert.Error(t, err)
					assert.ErrorContains(t, err, tt.errorMsg)
				}
			})
		}
	})

	t.Run("Test Value", func(t *testing.T) {
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				v, err := gidx.PrefixedID(tt.id).Value()
				if tt.errorMsg == "" {
					assert.NoError(t, err)
					assert.NotNil(t, v)
				} else {
					assert.Error(t, err)
					assert.ErrorContains(t, err, tt.errorMsg)
				}
			})
		}
	})

	t.Run("Test Scan", func(t *testing.T) {
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				id := gidx.PrefixedID("")
				err := id.Scan(tt.id)
				// scan should never return an error, if it's in the database treat it like it's valid
				assert.NoError(t, err)
				assert.NotEqual(t, "", string(id))
			})
		}
	})
}

func TestMarshalGQL(t *testing.T) {
	id := gidx.MustNewID("testing")

	var b bytes.Buffer

	id.MarshalGQL(&b)
	assert.Equal(t, fmt.Sprintf(`"%s"`, string(id)), b.String())
}
