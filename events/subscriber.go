package events

import (
	"context"
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
)

// Subscriber provides a pubsub subscriber that uses the watermill pubsub package
type Subscriber struct {
	prefix     string
	subscriber message.Subscriber
	logger     *zap.SugaredLogger
}

// NewSubscriberWithLogger returns a subscriber for the given config provided
func NewSubscriberWithLogger(cfg SubscriberConfig, logger *zap.SugaredLogger) (*Subscriber, error) {
	s := &Subscriber{
		prefix: cfg.Prefix,
		logger: logger,
	}

	switch {
	case strings.HasPrefix(cfg.URL, "nats://"):
		ns, err := newNATSSubscriber(cfg, s.logger)
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
func NewSubscriber(cfg SubscriberConfig) (*Subscriber, error) {
	return NewSubscriberWithLogger(cfg, zap.NewNop().Sugar())
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
