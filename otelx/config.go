// Copyright 2023 The Infratographer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelx

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.infratographer.com/x/viperx"
)

// Config provides a struct for reading in all the config values available
// from viper. If you are not using viper it is still able to be configured
// manually and passed in.
type Config struct {
	Enabled     bool          `mapstructure:"enabled"`
	Provider    TraceExporter `mapstructure:"provider"`
	Environment string        `mapstructure:"environment"`
	SampleRatio float64       `mapstructure:"sample_ratio"`
	Stdout      StdoutConfig
	OTLP        OTLPConfig
}

// StdoutConfig defines additional options when using the stdout provider.
type StdoutConfig struct {
	PrettyPrint       bool `mapstructure:"pretty_print"`
	DisableTimestamps bool `mapstructure:"disable_timestamps"`
}

// OTLPConfig defines additional options when using an OTLP provider.
type OTLPConfig struct {
	Endpoint    string        `mapstructure:"endpoint"`
	Insecure    bool          `mapstructure:"insecure"`
	Certificate string        `mapstructure:"certificate"`
	Headers     []string      `mapstructure:"headers"`
	Compression string        `mapstructure:"compression"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

// MustViperFlags returns the cobra flags and viper config to prevent code duplication
// and help provide consistent flags across the applications
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.Bool("tracing", false, "enable tracing support")
	viperx.MustBindFlag(v, "tracing.enabled", flags.Lookup("tracing"))
	flags.String("tracing-provider", "", `tracing provider to use options: "stdout", "otlphttp", "otlpgrpc", "passthrough"`)
	viperx.MustBindFlag(v, "tracing.provider", flags.Lookup("tracing-provider"))
	flags.String("tracing-environment", "production", "environment value in traces")
	viperx.MustBindFlag(v, "tracing.environment", flags.Lookup("tracing-environment"))
	flags.Float64("tracing-sample-ratio", 1.0, "ratio of traces sampled (0.0 - 1.0)")
	viperx.MustBindFlag(v, "tracing.sample_ratio", flags.Lookup("tracing-sample-ratio"))
	flags.String("tracing-otlp-endpoint", "", "OTLP exporter endpoint")
	viperx.MustBindFlag(v, "tracing.otlp.endpoint", flags.Lookup("tracing-otlp-endpoint"))

	v.MustBindEnv("tracing.stdout.pretty_print")
	v.MustBindEnv("tracing.stdout.disable_timestamps")
	v.MustBindEnv("tracing.otlp.endpoint")
	v.MustBindEnv("tracing.otlp.insecure")
	v.MustBindEnv("tracing.otlp.certificate")
	v.MustBindEnv("tracing.otlp.headers")
	v.MustBindEnv("tracing.otlp.compression")
	v.MustBindEnv("tracing.otlp.timeout")

	v.SetDefault("tracing.otlp.timeout", defaultTimeout)
}
