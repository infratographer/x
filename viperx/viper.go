// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package viperx

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// MustBindFlag provides a wrapper around the viper bindings that panics if an error occurs
func MustBindFlag(v *viper.Viper, name string, flag *pflag.Flag) {
	err := v.BindPFlag(name, flag)
	if err != nil {
		panic(err)
	}
}
