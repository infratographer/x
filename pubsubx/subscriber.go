package pubsubx

import (
	"context"
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"
)

// Subscriber provides a pubsub subscriber that uses the watermill pubsub package
type Subscriber struct {
	prefix     string
	subscriber message.Subscriber
}

// NewSubscriber returns a subscriber for the given config provided
func NewSubscriber(cfg SubscriberConfig) (*Subscriber, error) {
	s := &Subscriber{
		prefix: cfg.Prefix,
	}

	switch {
	case strings.HasPrefix(cfg.URL, "nats://"):
		ns, err := newNATSSubscriber(cfg)
		if err != nil {
			return nil, err
		}

		s.subscriber = ns
	default:
		return nil, ErrUnsupportedPubsub
	}

	return s, nil
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
