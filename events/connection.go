package events

import (
	"context"

	"go.uber.org/multierr"
)

// Connection defines a connection handler.
type Connection interface {
	// Gracefully close the connection.
	Shutdown(ctx context.Context) error

	// Source gives you the raw underlying connection object.
	Source() any

	Subscriber
	Publisher

	AuthRelationshipSubscriber
	AuthRelationshipPublisher
}

// Subscriber specifies subscriber methods.
type Subscriber interface {
	// SubscribeChanges subscribes to the provided topic responding with an ChangeMessage message.
	SubscribeChanges(ctx context.Context, topic string) (<-chan Message[ChangeMessage], error)
	// SubscribeEvents subscribes to the provided topic responding with an EventMessage message.
	SubscribeEvents(ctx context.Context, topic string) (<-chan Message[EventMessage], error)
}

// Publisher specifies publisher methods.
type Publisher interface {
	// PublishChange publishes to the specified topic with the message given.
	PublishChange(ctx context.Context, topic string, message ChangeMessage) (Message[ChangeMessage], error)
	// PublishEvent publishes to the specified topic with the message given.
	PublishEvent(ctx context.Context, topic string, message EventMessage) (Message[EventMessage], error)
}

// AuthRelationshipSubscriber specifies the auth relationship subscriber methods.
type AuthRelationshipSubscriber interface {
	// SubscribeAuthRelationshipRequests subscribes to the provided topic responding with an AuthRelationshipRequest message.
	SubscribeAuthRelationshipRequests(ctx context.Context, topic string) (<-chan Message[AuthRelationshipRequest], error)
}

// AuthRelationshipPublisher specifies the auth relationship publisher methods.
type AuthRelationshipPublisher interface {
	// PublishAuthRelationshipRequest publishes to the specified topic with the message given.
	PublishAuthRelationshipRequest(ctx context.Context, topic string, message AuthRelationshipRequest) (Message[AuthRelationshipResponse], error)
}

// NewConnection creates a new Connection from the provided config.
func NewConnection(config Config, options ...Option) (Connection, error) {
	var err error

	for _, opt := range options {
		err = multierr.Append(err, opt(&config))
	}

	if err != nil {
		return nil, err
	}

	if config.NATS.Configured() {
		return NewNATSConnection(config.NATS)
	}

	return nil, ErrProviderNotConfigured
}
