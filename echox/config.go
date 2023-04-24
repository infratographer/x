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
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.infratographer.com/x/viperx"
)

var (
	// DefaultServerShutdownTimeout sets the default for how long we give the sever
	// to shutdown before forcefully stopping the server.
	DefaultServerShutdownTimeout = 5 * time.Second
)

// Config is used to configure a new ginx server
type Config struct {
	Listen              string
	ShutdownGracePeriod time.Duration
}

// MustViperFlags returns the cobra flags and wires them up with viper to prevent code duplication
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet, defaultListen string) {
	flags.String("listen", defaultListen, "address to listen on")
	viperx.MustBindFlag(v, "server.listen", flags.Lookup("listen"))

	flags.Duration("shutdown-grace-period", DefaultServerShutdownTimeout, "server shutdown grace period")
	viperx.MustBindFlag(v, "server.shutdown-grace-period", flags.Lookup("shutdown-grace-period"))
}

// ConfigFromViper builds a new Config from viper.
func ConfigFromViper(v *viper.Viper) Config {
	return Config{
		Listen:              v.GetString("server.listen"),
		ShutdownGracePeriod: v.GetDuration("server.shutdown-grace-period"),
	}
}
