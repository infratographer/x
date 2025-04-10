package otelx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildOTLPEndpoint(t *testing.T) {
	testCases := []struct {
		name           string
		config         OTLPConfig
		expectEndpoint string
		expectError    string
	}{
		{
			"no endpoint",
			OTLPConfig{},
			"",
			"",
		},
		{
			"with scheme insecure",
			OTLPConfig{
				Endpoint: "http://localhost:4317",
			},
			"http://localhost:4317",
			"",
		},
		{
			"with scheme secure",
			OTLPConfig{
				Endpoint: "https://localhost:4317",
			},
			"https://localhost:4317",
			"",
		},
		{
			"without scheme insecure",
			OTLPConfig{
				Endpoint: "localhost:4317",
				Insecure: true,
			},
			"http://localhost:4317",
			"",
		},
		{
			"without scheme secure",
			OTLPConfig{
				Endpoint: "localhost:4317",
				Insecure: false,
			},
			"https://localhost:4317",
			"",
		},
		{
			"invalid with scheme",
			OTLPConfig{
				Endpoint: "https://localhost:error",
			},
			"",
			"invalid port",
		},
		{
			"invalid without scheme",
			OTLPConfig{
				Endpoint: "localhost:error",
				Insecure: true,
			},
			"",
			"invalid port",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			endpoint, err := buildOTLPEndpoint(tc.config)

			if tc.expectError != "" {
				require.Error(t, err, "expected error to be returned")

				assert.ErrorContains(t, err, tc.expectError, "unexpected error returned")

				return
			}

			require.NoError(t, err, "no error expected")

			assert.Equal(t, tc.expectEndpoint, endpoint, "unexpected endpoint returned")
		})
	}
}
