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

package events

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.infratographer.com/x/viperx"
)

var defaultTimeout = time.Second * 10

// PublisherConfig handles reading in all the config values available for setting up a pubsub publisher
type PublisherConfig struct {
	URL        string        `mapstructure:"url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	Prefix     string        `mapstructure:"prefix"`
	Source     string        `mapstructure:"source"`
	NATSConfig NATSConfig    `mapstructure:"nats"`
}

// SubscriberConfig handles reading in all the config values available for setting up a pubsub publisher
type SubscriberConfig struct {
	URL        string        `mapstructure:"url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	Prefix     string        `mapstructure:"prefix"`
	QueueGroup string        `mapstructure:"queueGroup"`
	NATSConfig NATSConfig    `mapstructure:"nats"`
}

// NATSConfig handles reading in all pubsub values specific to NATS
type NATSConfig struct {
	Token     string `mapstructure:"token"`
	CredsFile string `mapstructure:"credsFile"`
}

// MustViperFlagsForPublisher returns the cobra flags and viper config for an event publisher
func MustViperFlagsForPublisher(v *viper.Viper, flags *pflag.FlagSet, appName string) {
	flags.String("events-publisher-url", "nats://nats:4222", "nats server connection url")
	viperx.MustBindFlag(v, "events.publisher.url", flags.Lookup("events-publisher-url"))

	v.MustBindEnv("events.publisher.timeout")
	v.MustBindEnv("events.publisher.prefix")
	v.MustBindEnv("events.publisher.source")
	v.MustBindEnv("events.publisher.nats.token")
	v.MustBindEnv("events.publisher.nats.credsFile")

	v.SetDefault("events.publisher.timeout", defaultTimeout)
	v.SetDefault("events.publisher.source", appName)
}

// MustViperFlagsForSubscriber returns the cobra flags and viper config for an event subscriber
func MustViperFlagsForSubscriber(v *viper.Viper, flags *pflag.FlagSet) {
	flags.String("events-subscriber-url", "nats://nats:4222", "nats server connection url")
	viperx.MustBindFlag(v, "events.subscriber.url", flags.Lookup("events-subscriber-url"))
	flags.String("events-subscriber-queuegroup", "", "subscriber queue group")
	viperx.MustBindFlag(v, "events.subscriber.queueGroup", flags.Lookup("events-subscriber-queuegroup"))

	v.MustBindEnv("events.subscriber.timeout")
	v.MustBindEnv("events.subscriber.prefix")
	v.MustBindEnv("events.subscriber.nats.token")
	v.MustBindEnv("events.subscriber.nats.credsFile")

	v.SetDefault("events.subscriber.timeout", defaultTimeout)
}
