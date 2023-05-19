package events

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"go.infratographer.com/x/echojwtx"
	"go.infratographer.com/x/gidx"
)

// ErrUnsupportedPubsub is returned when the pubsub URL is not a supported provider
var ErrUnsupportedPubsub = errors.New("unsupported pubsub provider")

// ErrMissingEventType is returned when attempting to publish an event without an event type specified
var ErrMissingEventType = errors.New("event type missing")

// Publisher provides a pubsub publisher that uses the watermill pubsub package
type Publisher struct {
	prefix    string
	source    string
	publisher message.Publisher
}

// NewPublisher returns a publisher for the given config provided
func NewPublisher(cfg PublisherConfig) (*Publisher, error) {
	p := &Publisher{
		prefix: cfg.Prefix,
		source: cfg.Source,
	}

	switch {
	case strings.HasPrefix(cfg.URL, "nats://"):
		np, err := newNATSPublisher(cfg)
		if err != nil {
			return nil, err
		}

		p.publisher = np
	default:
		return nil, ErrUnsupportedPubsub
	}

	return p, nil
}

// PublishChange will publish a ChangeMessage to the topic for the change
func (p *Publisher) PublishChange(ctx context.Context, subjectType string, change ChangeMessage) error {
	if change.EventType == "" {
		return ErrMissingEventType
	}

	topic := strings.Join([]string{p.prefix, "changes", change.EventType, subjectType}, ".")

	change.Source = p.source
	if change.ActorID == gidx.NullPrefixedID {
		id, ok := ctx.Value(echojwtx.ActorCtxKey).(string)
		if ok {
			change.ActorID = gidx.PrefixedID(id)
		} else {
			change.ActorID = "unknown-actor"
		}
	}

	v, err := json.Marshal(change)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), v)

	return p.publisher.Publish(topic, msg)
}

// PublishEvent will publish an EventMessage to the proper topic for that event
func (p *Publisher) PublishEvent(_ context.Context, event EventMessage) error {
	if event.EventType == "" {
		return ErrMissingEventType
	}

	topic := strings.Join([]string{p.prefix, "events", event.EventType}, ".")

	event.Source = p.source

	v, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), v)

	return p.publisher.Publish(topic, msg)
}
