package otelx_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.infratographer.com/x/otelx"
)

var envReplacer = strings.NewReplacer(".", "_", "-", "_")

type appConfig struct {
	Tracing otelx.Config
}

func initConfig(t *testing.T, env map[string]string, flags []string) (otelx.Config, *viper.Viper, *pflag.FlagSet, error) {
	t.Helper()

	for k, v := range env {
		envName := strings.ToUpper(envReplacer.Replace(t.Name() + "_" + k))

		require.NoErrorf(t, os.Setenv(envName, v), "No error expected while setting env var: %s = %s", envName, v)

		defer func() {
			require.NoErrorf(t, os.Unsetenv(envName), "No error expected while unsetting env var: %s", envName)
		}()
	}

	v := viper.New()

	v.SetEnvPrefix(t.Name())
	v.SetEnvKeyReplacer(envReplacer)
	v.AutomaticEnv()

	flagset := pflag.NewFlagSet("flags", pflag.ContinueOnError)

	otelx.MustViperFlags(v, flagset)

	err := flagset.Parse(flags)
	if err != nil {
		return otelx.Config{}, v, flagset, fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg := new(appConfig)

	err = v.Unmarshal(cfg)
	if err != nil {
		return cfg.Tracing, v, flagset, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg.Tracing, v, flagset, nil
}

func TestViperFlags(t *testing.T) {
	defaultOTLPTimeout := 10 * time.Second

	testCases := []struct {
		name         string
		env          map[string]string
		flags        []string
		expectConfig map[string]any
		expectError  string
	}{
		{
			"defaults",
			nil,
			nil,
			map[string]any{
				"environment":  "production",
				"sample_ratio": 1.0,
				"otlp": map[string]any{
					"timeout": defaultOTLPTimeout,
				},
			},
			"",
		},
		{
			"env values",
			map[string]string{
				"TRACING_ENABLED":             "true",
				"TRACING_PROVIDER":            "otlpgrpc",
				"TRACING_ENVIRONMENT":         "test",
				"TRACING_SAMPLE_RATIO":        "0.5",
				"TRACING_STDOUT_PRETTY_PRINT": "true",
				"TRACING_OTLP_ENDPOINT":       "grpc.example.com:1234",
				"TRACING_OTLP_TIMEOUT":        "5s",
			},
			nil,
			map[string]any{
				"enabled":      true,
				"provider":     "otlpgrpc",
				"environment":  "test",
				"sample_ratio": 0.5,
				"stdout": map[string]any{
					"pretty_print": true,
				},
				"otlp": map[string]any{
					"endpoint": "grpc.example.com:1234",
					"timeout":  5 * time.Second,
				},
			},
			"",
		},
		{
			"flag values",
			nil,
			[]string{
				"--tracing",
				"--tracing-provider", "otlphttp",
				"--tracing-environment", "test1",
				"--tracing-sample-ratio", "0.7",
			},
			map[string]any{
				"enabled":      true,
				"provider":     "otlphttp",
				"environment":  "test1",
				"sample_ratio": 0.7,
				"otlp": map[string]any{
					"timeout": defaultOTLPTimeout,
				},
			},
			"",
		},
		{
			"flag preferred values",
			map[string]string{
				"TRACING_ENABLED":      "true",
				"TRACING_PROVIDER":     "otlpgrpc",
				"TRACING_SAMPLE_RATIO": "0.5",
			},
			[]string{
				"--tracing-provider", "otlphttp",
				"--tracing-environment", "test1",
				"--tracing-sample-ratio", "0.7",
			},
			map[string]any{
				"enabled":      true,
				"provider":     "otlphttp",
				"environment":  "test1",
				"sample_ratio": 0.7,
				"otlp": map[string]any{
					"timeout": defaultOTLPTimeout,
				},
			},
			"",
		},
		{
			"invalid env value",
			map[string]string{
				"TRACING_ENABLED":      "true",
				"TRACING_PROVIDER":     "otlpgrpc",
				"TRACING_SAMPLE_RATIO": "bad",
			},
			nil,
			nil,
			"cannot parse",
		},
		{
			"invalid flag value",
			map[string]string{
				"TRACING_ENABLED":      "true",
				"TRACING_PROVIDER":     "otlpgrpc",
				"TRACING_SAMPLE_RATIO": "0.5",
			},
			[]string{
				"--tracing-sample-ratio", "bad",
			},
			nil,
			"invalid argument",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var expectCfg otelx.Config

			if tc.expectConfig != nil {
				require.NoError(t, mapstructure.Decode(tc.expectConfig, &expectCfg), "unexpected error decoding expected config")
			}

			cfg, _, _, err := initConfig(t, tc.env, tc.flags)

			if tc.expectError != "" {
				require.Error(t, err, "expected error to be returned")

				require.ErrorContains(t, err, tc.expectError, "unexpected error returned")

				return
			}

			require.NoError(t, err, "no error expected to be returned")

			assert.Equal(t, expectCfg, cfg, "unexpected configuration")
		})
	}
}
