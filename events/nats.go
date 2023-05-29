package events

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/garsue/watermillzap"
	nc "github.com/nats-io/nats.go"
	"github.com/spf13/viper"
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

func newNATSSubscriber(cfg SubscriberConfig, logger *zap.SugaredLogger) (message.Subscriber, error) {
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
		SubscribeOptions: nil,
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

// WithNATS sets the NATS config for a SubscriberConfig from viper
func (s *SubscriberConfig) WithNATS(v *viper.Viper) {
	s.NATSConfig = NATSConfig{
		Token:     v.GetString("events.subscriber.nats.token"),
		CredsFile: v.GetString("events.subscriber.nats.credsFile"),
	}
}

// WithNATS sets the NATS config for a SubscriberConfig from viper
func (p *PublisherConfig) WithNATS(v *viper.Viper) {
	p.NATSConfig = NATSConfig{
		Token:     v.GetString("events.publisher.nats.token"),
		CredsFile: v.GetString("events.publisher.nats.credsFile"),
	}
}
