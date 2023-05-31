// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

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
	Random      string        `mapstructure:"random"`
	Stdout      struct {
		PrettyPrint       bool `mapstructure:"pretty_print"`
		DisableTimestamps bool `mapstructure:"disable_timestamps"`
	}
	Jaeger struct {
		AgentHost string `mapstructure:"agent_host"`
		AgentPort string `mapstructure:"agent_port"`
		Endpoint  string `mapstructure:"endpoint"`
		User      string `mapstructure:"user"`
		Password  string `mapstructure:"password"`
	}
	OTLP struct {
		Endpoint    string        `mapstructure:"endpoint"`
		Insecure    bool          `mapstructure:"insecure"`
		Certificate string        `mapstructure:"certificate"`
		Headers     []string      `mapstructure:"headers"`
		Compression string        `mapstructure:"compression"`
		Timeout     time.Duration `mapstructure:"timeout"`
	}
}

// MustViperFlags returns the cobra flags and viper config to prevent code duplication
// and help provide consistent flags across the applications
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.Bool("tracing", false, "enable tracing support")
	viperx.MustBindFlag(v, "tracing.enabled", flags.Lookup("tracing"))
	flags.String("tracing-provider", "", `tracing provider to use options: "stdout", "jaeger", "otlphttp", "otlpgrpc"`)
	viperx.MustBindFlag(v, "tracing.provider", flags.Lookup("tracing-provider"))
	flags.String("tracing-environment", "production", "environment value in traces")
	viperx.MustBindFlag(v, "tracing.environment", flags.Lookup("tracing-environment"))

	v.MustBindEnv("tracing.stdout.pretty_print")
	v.MustBindEnv("tracing.stdout.disable_timestamps")
	v.MustBindEnv("tracing.jaeger.agent_host")
	v.MustBindEnv("tracing.jaeger.agent_port")
	v.MustBindEnv("tracing.jaeger.endpoint")
	v.MustBindEnv("tracing.jaeger.user")
	v.MustBindEnv("tracing.jaeger.password")
	v.MustBindEnv("tracing.otlp.endpoint")
	v.MustBindEnv("tracing.otlp.insecure")
	v.MustBindEnv("tracing.otlp.certificate")
	v.MustBindEnv("tracing.otlp.headers")
	v.MustBindEnv("tracing.otlp.compression")
	v.MustBindEnv("tracing.otlp.timeout")

	v.SetDefault("tracing.otlp.timeout", defaultTimeout)
}
