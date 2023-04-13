// Copyright 2022 The Infratographer Authors
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

package echox

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.infratographer.com/x/viperx"
)

// Config is used to configure a new ginx server
type Config struct {
	Listen string
}

// MustViperFlags returns the cobra flags and wires them up with viper to prevent code duplication
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet, defaultListen string) {
	flags.String("listen", defaultListen, "address to listen on")
	viperx.MustBindFlag(v, "server.listen", flags.Lookup("listen"))
}
