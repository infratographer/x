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
	"context"
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Subscriber provides a pubsub subscriber that uses the watermill pubsub package
type Subscriber struct {
	prefix     string
	subscriber message.Subscriber
	logger     *zap.SugaredLogger
}

// NewSubscriberWithLogger returns a subscriber for the given config provided
func NewSubscriberWithLogger(cfg SubscriberConfig, logger *zap.SugaredLogger, options ...nats.SubOpt) (*Subscriber, error) {
	s := &Subscriber{
		prefix: cfg.Prefix,
		logger: logger,
	}

	switch {
	case strings.HasPrefix(cfg.URL, "nats://"):
		ns, err := newNATSSubscriber(cfg, s.logger, options...)
		if err != nil {
			return nil, err
		}

		s.subscriber = ns
	default:
		return nil, ErrUnsupportedPubsub
	}

	return s, nil
}

// NewSubscriber returns a subscriber for the given config provided
func NewSubscriber(cfg SubscriberConfig, options ...nats.SubOpt) (*Subscriber, error) {
	return NewSubscriberWithLogger(cfg, zap.NewNop().Sugar(), options...)
}

// SubscribeChanges will subscribe you to the changes for a given topic. To receive all changes of any kind you can
// pass in ">".
func (s *Subscriber) SubscribeChanges(ctx context.Context, topic string) (<-chan *message.Message, error) {
	topic = strings.Join([]string{s.prefix, "changes", topic}, ".")

	return s.subscriber.Subscribe(ctx, topic)
}

// SubscribeEvents will subscribe you to the events for a given topic. To receive all changes of any kind you can
// pass in ">".
func (s *Subscriber) SubscribeEvents(ctx context.Context, topic string) (<-chan *message.Message, error) {
	topic = strings.Join([]string{s.prefix, "events", topic}, ".")

	return s.subscriber.Subscribe(ctx, topic)
}

// Close will close the subscriber
func (s *Subscriber) Close() error {
	return s.subscriber.Close()
}
