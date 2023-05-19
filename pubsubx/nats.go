package pubsubx

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	nc "github.com/nats-io/nats.go"
)

var natsMarshaler = &nats.JSONMarshaler{}

func newNATSPublisher(cfg PublisherConfig) (message.Publisher, error) {
	logger := watermill.NewStdLogger(false, false)

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
		logger,
	)
}

func newNATSSubscriber(cfg SubscriberConfig) (message.Subscriber, error) {
	logger := watermill.NewStdLogger(false, false)

	options := []nc.Option{
		nc.Timeout(cfg.Timeout),
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

	sub, err := nats.NewSubscriber(
		nats.SubscriberConfig{
			URL:              cfg.URL,
			NatsOptions:      options,
			Unmarshaler:      natsMarshaler,
			JetStream:        jsConfig,
			QueueGroupPrefix: cfg.QueueGroup,
		},
		logger,
	)

	return sub, err
}
