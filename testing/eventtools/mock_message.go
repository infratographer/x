package eventtools

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"go.infratographer.com/x/events"
)

var _ events.Message[any] = (*MockMessage[interface{}])(nil)

// MockMessage implements events.Message.
type MockMessage[T any] struct {
	mock.Mock
}

// Connection implements events.Message.
func (m *MockMessage[T]) Connection() events.Connection {
	args := m.Called()

	return args.Get(0).(events.Connection)
}

// ID implements events.Message.
func (m *MockMessage[T]) ID() string {
	args := m.Called()

	return args.String(0)
}

// Topic implements events.Message.
func (m *MockMessage[T]) Topic() string {
	args := m.Called()

	return args.String(0)
}

// Message implements events.Message.
func (m *MockMessage[T]) Message() T {
	args := m.Called()

	return args.Get(0).(T)
}

// Ack implements events.Message.
func (m *MockMessage[T]) Ack() error {
	args := m.Called()

	return args.Error(0)
}

// Nak implements events.Message.
func (m *MockMessage[T]) Nak(delay time.Duration) error {
	args := m.Called(delay)

	return args.Error(0)
}

// Term implements events.Message.
func (m *MockMessage[T]) Term() error {
	args := m.Called()

	return args.Error(0)
}

// Timestamp implements events.Message.
func (m *MockMessage[T]) Timestamp() time.Time {
	args := m.Called()

	return args.Get(0).(time.Time)
}

// Deliveries implements events.Message.
func (m *MockMessage[T]) Deliveries() uint64 {
	args := m.Called()

	return args.Get(0).(uint64)
}

// Error implements events.Message.
func (m *MockMessage[T]) Error() error {
	args := m.Called()

	return args.Error(0)
}

// ReplyAuthRelationshipRequest implements events.Message.
func (m *MockMessage[T]) ReplyAuthRelationshipRequest(_ context.Context, message events.AuthRelationshipResponse) (events.Message[events.AuthRelationshipResponse], error) {
	args := m.Called(message)

	return args.Get(0).(events.Message[events.AuthRelationshipResponse]), args.Error(1)
}

// Source implements events.Message.
func (m *MockMessage[T]) Source() any {
	args := m.Called()

	return args.Error(0)
}
