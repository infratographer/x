// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package ginx

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.infratographer.com/x/viperx"
)

// Config is used to configure a new ginx server
type Config struct {
	Listen string `mapstructure:"listen"`
}

// MustViperFlags returns the cobra flags and wires them up with viper to prevent code duplication
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet, defaultListen string) {
	flags.String("listen", defaultListen, "address to listen on")
	viperx.MustBindFlag(v, "server.listen", flags.Lookup("listen"))
}
