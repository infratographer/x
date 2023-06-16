// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

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
