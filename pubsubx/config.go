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

package pubsubx

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var defaultTimeout = time.Second * 10

// PublisherConfig handles reading in all the config values available for setting up a pubsub publisher
type PublisherConfig struct {
	URL        string        `mapstructure:"url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	Prefix     string        `mapstructure:"prefix"`
	NATSConfig NATSConfig    `mapstructure:"nats"`
	Source     string        `mapstructure:"source"`
}

// NATSConfig handles reading in all pubsub values specific to NATS
type NATSConfig struct {
	Token     string `mapstructure:"token"`
	CredsFile string `mapstructure:"creds_file"`
}

// SubscriberConfig handles reading in all the config values available for setting up a pubsub publisher
type SubscriberConfig struct {
	URL        string        `mapstructure:"url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	Prefix     string        `mapstructure:"prefix"`
	QueueGroup string        `mapstructure:"prefix"`
	NATSConfig NATSConfig    `mapstructure:"nats"`
}

// MustViperFlags returns the cobra flags and viper config to prevent code duplication
// and help provide consistent flags across the applications
func MustViperFlags(v *viper.Viper, _ *pflag.FlagSet) {
	v.MustBindEnv("pubsub.timeout")

	v.SetDefault("pubsub.timeout", defaultTimeout)
}
