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
