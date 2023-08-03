package eventtools

import (
	"context"

	"github.com/stretchr/testify/mock"

	"go.infratographer.com/x/events"
)

var _ events.Connection = (*MockConnection)(nil)

// MockConnection implements events.Connection
type MockConnection struct {
	mock.Mock
}

// Shutdown implements events.Connection
func (c *MockConnection) Shutdown(_ context.Context) error {
	args := c.Called()

	return args.Error(0)
}

// PublishAuthRelationshipRequest implements events.Connection
func (c *MockConnection) PublishAuthRelationshipRequest(_ context.Context, topic string, message events.AuthRelationshipRequest) (events.Message[events.AuthRelationshipResponse], error) {
	args := c.Called(topic, message)

	return args.Get(0).(events.Message[events.AuthRelationshipResponse]), args.Error(1)
}

// PublishChange implements events.Connection
func (c *MockConnection) PublishChange(_ context.Context, topic string, message events.ChangeMessage) (events.Message[events.ChangeMessage], error) {
	args := c.Called(topic, message)

	return args.Get(0).(events.Message[events.ChangeMessage]), args.Error(1)
}

// PublishEvent implements events.Connection
func (c *MockConnection) PublishEvent(_ context.Context, topic string, message events.EventMessage) (events.Message[events.EventMessage], error) {
	args := c.Called(topic, message)

	return args.Get(0).(events.Message[events.EventMessage]), args.Error(1)
}

// Source implements events.Connection
func (c *MockConnection) Source() any {
	args := c.Called()

	return args.Error(0)
}

// SubscribeAuthRelationshipRequests implements events.Connection
func (c *MockConnection) SubscribeAuthRelationshipRequests(_ context.Context, topic string) (<-chan events.Request[events.AuthRelationshipRequest, events.AuthRelationshipResponse], error) {
	args := c.Called(topic)

	return args.Get(0).(<-chan events.Request[events.AuthRelationshipRequest, events.AuthRelationshipResponse]), args.Error(1)
}

// SubscribeChanges implements events.Connection
func (c *MockConnection) SubscribeChanges(_ context.Context, topic string) (<-chan events.Message[events.ChangeMessage], error) {
	args := c.Called(topic)

	return args.Get(0).(<-chan events.Message[events.ChangeMessage]), args.Error(1)
}

// SubscribeEvents implements events.Connection
func (c *MockConnection) SubscribeEvents(_ context.Context, topic string) (<-chan events.Message[events.EventMessage], error) {
	args := c.Called(topic)

	return args.Get(0).(<-chan events.Message[events.EventMessage]), args.Error(1)
}
