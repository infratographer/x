package events

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	base10 = 10

	natsTracerName = tracerName + ":nats"
)

var _ Connection = (*NATSConnection)(nil)

// NATSConnection implements Connection.
type NATSConnection struct {
	logger    *zap.SugaredLogger
	tracer    trace.Tracer
	conn      *nats.Conn
	jetstream nats.JetStreamContext
	cfg       NATSConfig
}

// Shutdown gracefully drains the connection.
func (c *NATSConnection) Shutdown(ctx context.Context) error {
	ctx, cancelTimeout := context.WithTimeout(ctx, c.cfg.ShutdownTimeout)
	ctx, cancel := context.WithCancelCause(ctx)

	defer cancelTimeout()

	closedCB := c.conn.Opts.ClosedCB

	c.conn.Opts.ClosedCB = func(c *nats.Conn) {
		defer cancel(nil)

		if closedCB != nil {
			closedCB(c)
		}
	}

	if err := c.conn.Drain(); err != nil {
		cancel(err)

		return err
	}

	<-ctx.Done()

	return ctx.Err()
}

// Source returns the underlying NATS Connection.
func (c *NATSConnection) Source() any {
	return c.conn
}

func (c *NATSConnection) durableName(topic string) string {
	return NATSConsumerDurableName(c.cfg.QueueGroup, topic)
}

func (c *NATSConnection) buildSubscribeSubject(parts ...string) string {
	var subjectParts []string

	if c.cfg.SubscribePrefix != "" {
		subjectParts = append(subjectParts, c.cfg.SubscribePrefix)
	}

	subjectParts = append(subjectParts, parts...)

	return strings.Join(subjectParts, ".")
}

func (c *NATSConnection) buildPublishSubject(parts ...string) string {
	var subjectParts []string

	if c.cfg.PublishPrefix != "" {
		subjectParts = append(subjectParts, c.cfg.PublishPrefix)
	}

	subjectParts = append(subjectParts, parts...)

	return strings.Join(subjectParts, ".")
}

func newNATSMessage[T any](conn *NATSConnection, subject string, message T) (*NATSMessage[T], error) {
	data, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	return &NATSMessage[T]{
		conn: conn,
		source: &nats.Msg{
			Subject: subject,
			Data:    data,
		},
		message: message,
	}, nil
}

// NewNATSConnection creates a new nats jetstream connection.
func NewNATSConnection(config NATSConfig, options ...NATSOption) (*NATSConnection, error) {
	nc := config.WithDefaults()

	if err := nc.Validate(); err != nil {
		return nil, err
	}

	for _, opt := range options {
		if err := opt(&nc); err != nil {
			return nil, err
		}
	}

	if config.QueueGroup == "" {
		nc.logger.Warn("NATS QueueGroup is not set. Subscriptions will not be durable.")
	}

	conn, err := nats.Connect(config.URL, nc.connectOptions...)
	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream(nc.jetStreamOptions...)
	if err != nil {
		conn.Close()

		return nil, err
	}

	return &NATSConnection{
		logger:    nc.logger,
		tracer:    otel.GetTracerProvider().Tracer(natsTracerName),
		conn:      conn,
		jetstream: js,
		cfg:       nc,
	}, nil
}

// NATSConsumerDurableName is the generator function to create a new durable consumer name.
// If queueGroup is empty, an empty durable name is returned to support ephemeral consumers.
func NATSConsumerDurableName(queueGroup, subject string) string {
	if queueGroup == "" {
		return ""
	}

	hash := md5.Sum([]byte(subject))

	return queueGroup + hex.EncodeToString(hash[:])
}
