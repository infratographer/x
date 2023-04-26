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
	// Debug enables echo's Debug option.
	Debug bool

	// Listen sets the listen address to serve the echo server on.
	Listen string

	// ShutdownGracePeriod sets the grace period for in flight requests before shutting down.
	ShutdownGracePeriod time.Duration

	// TrustedProxies defines the allowed ip / network ranges to trust a proxy from.
	TrustedProxies []string
}

// withDefaults returns a new config with defaults set if not already defined.
func (c Config) withDefaults() Config {
	if c.Listen == "" {
		c.Listen = ":8080"
	}

	if c.ShutdownGracePeriod <= 0 {
		c.ShutdownGracePeriod = DefaultServerShutdownTimeout
	}

	return c
}

// WithDebug enables echo's Debug option.
func (c Config) WithDebug(debug bool) Config {
	c.Debug = debug

	return c
}

// WithListen sets the listen address to serve the echo server on.
func (c Config) WithListen(listen string) Config {
	c.Listen = listen

	return c
}

// WithShutdownGracePeriod sets the grace period for in flight requests before shutting down.
func (c Config) WithShutdownGracePeriod(period time.Duration) Config {
	c.ShutdownGracePeriod = period

	return c
}

// WithTrustedProxies defines the allowed ip / network ranges to trust a proxy from.
func (c Config) WithTrustedProxies(trust ...string) Config {
	c.TrustedProxies = append(c.TrustedProxies, trust...)

	return c
}

// MustViperFlags returns the cobra flags and wires them up with viper to prevent code duplication
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet, defaultListen string) {
	flags.Bool("debug", false, "enable server debug")
	viperx.MustBindFlag(v, "server.debug", flags.Lookup("debug"))

	flags.String("listen", defaultListen, "address to listen on")
	viperx.MustBindFlag(v, "server.listen", flags.Lookup("listen"))

	flags.Duration("shutdown-grace-period", DefaultServerShutdownTimeout, "server shutdown grace period")
	viperx.MustBindFlag(v, "server.shutdown-grace-period", flags.Lookup("shutdown-grace-period"))

	flags.StringSlice("trusted-proxies", nil, "server trusted proxies")
	viperx.MustBindFlag(v, "server.trusted-proxies", flags.Lookup("trusted-proxies"))
}

// ConfigFromViper builds a new Config from viper.
func ConfigFromViper(v *viper.Viper) Config {
	return Config{
		Debug:               v.GetBool("server.debug"),
		Listen:              v.GetString("server.listen"),
		ShutdownGracePeriod: v.GetDuration("server.shutdown-grace-period"),
		TrustedProxies:      v.GetStringSlice("server.trusted-proxies"),
	}
}
