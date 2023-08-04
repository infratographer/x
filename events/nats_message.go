package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

func natsSubscriptionMessageChan[T any](ctx context.Context, conn *NATSConnection, batchSize int, natsCh <-chan *nats.Msg) chan Message[T] {
	msgCh := make(chan Message[T], batchSize)

	go func() {
		defer close(msgCh)

		for nMsg := range natsCh {
			msg := natsDecodeMessage[T](conn, nMsg)

			select {
			case msgCh <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()

	return msgCh
}

func natsSubscriptionAuthRelationshipRequestChan(ctx context.Context, conn *NATSConnection, batchSize int, natsCh <-chan *nats.Msg) chan Request[AuthRelationshipRequest, AuthRelationshipResponse] {
	msgCh := make(chan Request[AuthRelationshipRequest, AuthRelationshipResponse], batchSize)

	go func() {
		defer close(msgCh)

		for nMsg := range natsCh {
			msg := natsDecodeMessage[AuthRelationshipRequest](conn, nMsg)

			req := &NATSAuthRelationshipRequest{
				NATSMessage: msg.(*NATSMessage[AuthRelationshipRequest]),
			}

			select {
			case msgCh <- req:
			case <-ctx.Done():
				return
			}
		}
	}()

	return msgCh
}

func natsDecodeMessage[T any](conn *NATSConnection, nMsg *nats.Msg) Message[T] {
	msg := &NATSMessage[T]{
		conn:   conn,
		source: nMsg,
	}

	if err := json.Unmarshal(nMsg.Data, &msg.message); err != nil {
		msg.err = err
	}

	return msg
}

var _ Message[any] = (*NATSMessage[any])(nil)

// NATSMessage implements Message
type NATSMessage[T any] struct {
	conn           *NATSConnection
	source         *nats.Msg
	sourceMetadata *nats.MsgMetadata
	message        T
	err            error
}

// Connection returns the underlying Connection.
func (m *NATSMessage[T]) Connection() Connection {
	return m.conn
}

func (m *NATSMessage[T]) metadata() nats.MsgMetadata {
	if m.sourceMetadata != nil {
		return *m.sourceMetadata
	}

	metadata, err := m.source.Metadata()
	if err != nil {
		m.conn.logger.Errorw("failed to load metadata for nats message", "nats.subject", m.source.Subject)

		return nats.MsgMetadata{}
	}

	m.sourceMetadata = metadata

	return *m.sourceMetadata
}

// ID returns the nats message sequence number for the consumer.
func (m *NATSMessage[T]) ID() string {
	return strconv.FormatUint(m.metadata().Sequence.Consumer, base10)
}

// Topic returns the nats subject.
func (m *NATSMessage[T]) Topic() string {
	return m.source.Subject
}

// Message returns the decoded message object.
func (m *NATSMessage[T]) Message() T {
	return m.message
}

// Ack acks the message.
func (m *NATSMessage[T]) Ack() error {
	return m.source.Ack()
}

// Nak calls a Nak with the provided delay.
func (m *NATSMessage[T]) Nak(delay time.Duration) error {
	return m.source.NakWithDelay(delay)
}

// Term terminates the message from being processed again.
func (m *NATSMessage[T]) Term() error {
	return m.source.Term()
}

// Timestamp returns the timestamp of the message.
func (m *NATSMessage[T]) Timestamp() time.Time {
	return m.metadata().Timestamp
}

// Deliveries returns the number of times the message was delivered.
func (m *NATSMessage[T]) Deliveries() uint64 {
	return m.metadata().NumDelivered
}

// Error returns any error with the message.
func (m *NATSMessage[T]) Error() error {
	if m.err != nil {
		return m.err
	}

	return nil
}

// Source returns the underlying nats message.
func (m *NATSMessage[T]) Source() any {
	return m.source
}

func (m *NATSMessage[T]) publish() error {
	return m.conn.conn.PublishMsg(m.source)
}

func (m *NATSMessage[T]) request(ctx context.Context) (Message[AuthRelationshipResponse], error) {
	if m.source.Reply == "" {
		m.source.Reply = m.conn.conn.NewRespInbox()
	}

	nMsg, err := m.conn.conn.RequestMsgWithContext(ctx, m.source)
	if err != nil {
		// ensure we wrap no responder errors with ErrRequestNoResponders.
		if errors.Is(err, nats.ErrNoResponders) {
			return nil, fmt.Errorf("%w: %w", ErrRequestNoResponders, err)
		}

		return nil, err
	}

	respMsg := natsDecodeMessage[AuthRelationshipResponse](m.conn, nMsg)

	return respMsg, nil
}

var _ Request[AuthRelationshipRequest, AuthRelationshipResponse] = (*NATSAuthRelationshipRequest)(nil)

// NATSAuthRelationshipRequest implements Request for AuthRelationshipRequest / AuthRelationshipResponse
type NATSAuthRelationshipRequest struct {
	*NATSMessage[AuthRelationshipRequest]
}

// Reply responds to an AuthRelationshipRequest with an AuthRelationshipResponse.
func (r *NATSAuthRelationshipRequest) Reply(ctx context.Context, message AuthRelationshipResponse) (Message[AuthRelationshipResponse], error) {
	ctx, span := r.conn.tracer.Start(ctx, "events.Reply")

	defer span.End()

	if r.source.Reply == "" {
		span.RecordError(ErrNATSMessageNoReplySubject)
		span.SetStatus(codes.Error, ErrNATSMessageNoReplySubject.Error())

		return nil, ErrNATSMessageNoReplySubject
	}

	if err := message.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	// Propagate trace context into the message for the subscriber
	var mapCarrier propagation.MapCarrier = make(map[string]string)

	otel.GetTextMapPropagator().Inject(ctx, mapCarrier)

	message.TraceContext = mapCarrier

	respMsg, err := newNATSMessage(r.conn, r.source.Reply, message)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	if err := r.source.RespondMsg(respMsg.source); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return respMsg, err
	}

	return respMsg, nil
}
