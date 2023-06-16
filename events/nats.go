// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package events

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/garsue/watermillzap"
	nc "github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

var natsMarshaler = &nats.JSONMarshaler{}

func newNATSPublisher(cfg PublisherConfig, logger *zap.SugaredLogger) (message.Publisher, error) {
	logAdapter := watermillzap.NewLogger(logger.Desugar())

	options := []nc.Option{
		nc.Timeout(cfg.Timeout),
	}

	switch {
	case cfg.NATSConfig.CredsFile != "":
		options = append(options, nc.UserCredentials(cfg.NATSConfig.CredsFile))
	case cfg.NATSConfig.Token != "":
		options = append(options, nc.Token(cfg.NATSConfig.Token))
	}

	jsConfig := nats.JetStreamConfig{
		Disabled:       false,
		AutoProvision:  false,
		ConnectOptions: nil,
		PublishOptions: nil,
		TrackMsgId:     false,
		AckAsync:       false,
		DurablePrefix:  "",
	}

	return nats.NewPublisher(
		nats.PublisherConfig{
			URL:         cfg.URL,
			NatsOptions: options,
			Marshaler:   natsMarshaler,
			JetStream:   jsConfig,
		},
		logAdapter,
	)
}

func newNATSSubscriber(cfg SubscriberConfig, logger *zap.SugaredLogger, subOpts ...nc.SubOpt) (message.Subscriber, error) {
	logAdapter := watermillzap.NewLogger(logger.Desugar())

	options := []nc.Option{
		nc.Timeout(cfg.Timeout),
	}

	switch {
	case cfg.NATSConfig.CredsFile != "":
		options = append(options, nc.UserCredentials(cfg.NATSConfig.CredsFile))
	case cfg.NATSConfig.Token != "":
		options = append(options, nc.Token(cfg.NATSConfig.Token))
	}

	jsConfig := nats.JetStreamConfig{
		Disabled:         false,
		AutoProvision:    false,
		ConnectOptions:   nil,
		PublishOptions:   nil,
		SubscribeOptions: subOpts,
		TrackMsgId:       false,
		AckAsync:         false,
		DurablePrefix:    "",
	}

	if cfg.QueueGroup != "" {
		jsConfig.DurableCalculator = func(_ string, topic string) string {
			hash := md5.Sum([]byte(topic))
			return cfg.QueueGroup + hex.EncodeToString(hash[:])
		}
	}

	sub, err := nats.NewSubscriber(
		nats.SubscriberConfig{
			URL:              cfg.URL,
			NatsOptions:      options,
			Unmarshaler:      natsMarshaler,
			JetStream:        jsConfig,
			QueueGroupPrefix: cfg.QueueGroup,
		},
		logAdapter,
	)

	return sub, err
}
