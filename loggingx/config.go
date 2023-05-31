// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package loggingx

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.infratographer.com/x/viperx"
)

// Config handles reading in all the config values available for setting up a logger
type Config struct {
	Debug  bool `mapstructure:"debug"`
	Pretty bool `mapstructure:"pretty"`
}

// MustViperFlags returns the cobra flags and viper config to prevent code duplication
// and help provide consistent flags across the applications
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.Bool("debug", false, "enable debug logging")
	viperx.MustBindFlag(v, "logging.debug", flags.Lookup("debug"))
	flags.Bool("pretty", false, "enable pretty (human readable) logging output")
	viperx.MustBindFlag(v, "logging.pretty", flags.Lookup("pretty"))
}
