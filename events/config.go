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

package events

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	defaultTimeout = time.Second * 10
	tracerName     = "go.infratographer.com/x/events"
)

// Config contains event provider configs.
type Config struct {
	NATS NATSConfig `mapstructure:"nats"`
}

// MustViperFlags returns the cobra flags and viper config for events.
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet, appName string) {
	MustViperFlagsForNATS(v, flags, appName)
}

// Option configures a connection option.
type Option func(config *Config) error

// WithLogger sets the logger for the connection.
func WithLogger(logger *zap.SugaredLogger) Option {
	return func(config *Config) error {
		config.NATS.logger = logger

		return nil
	}
}

// WithNATSOptions configures nats options.
func WithNATSOptions(options ...NATSOption) Option {
	return func(config *Config) error {
		var err error

		for _, opt := range options {
			err = multierr.Append(err, opt(&config.NATS))
		}

		return err
	}
}
