// Copyright 2023 The Infratographer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"go.infratographer.com/x/echojwtx"
	"go.infratographer.com/x/gidx"
)

const instrumentationName = "go.infratographer.com/x/events"

// ErrUnsupportedPubsub is returned when the pubsub URL is not a supported provider
var ErrUnsupportedPubsub = errors.New("unsupported pubsub provider")

// ErrMissingEventType is returned when attempting to publish an event without an event type specified
var ErrMissingEventType = errors.New("event type missing")

// InvalidAuthRelationshipAction is returned when attempting to publish an AuthRelationshipAction that isn't write or delete
var ErrInvalidAuthRelationshipAction = errors.New("invalid auth relationship action")

// Publisher provides a pubsub publisher that uses the watermill pubsub package
type Publisher struct {
	prefix    string
	source    string
	publisher message.Publisher
	logger    *zap.SugaredLogger
	tracer    trace.Tracer
}

// NewPublisherWithLogger returns a publisher for the given config provided
func NewPublisherWithLogger(cfg PublisherConfig, logger *zap.SugaredLogger) (*Publisher, error) {
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)

	p := &Publisher{
		prefix: cfg.Prefix,
		source: cfg.Source,
		logger: logger,
		tracer: tracer,
	}

	switch {
	case strings.HasPrefix(cfg.URL, "nats://"):
		np, err := newNATSPublisher(cfg, p.logger)
		if err != nil {
			return nil, err
		}

		p.publisher = np
	default:
		return nil, ErrUnsupportedPubsub
	}

	return p, nil
}

// NewPublisher returns a publisher for the given config provided
func NewPublisher(cfg PublisherConfig) (*Publisher, error) {
	return NewPublisherWithLogger(cfg, zap.NewNop().Sugar())
}

func (p *Publisher) PublishAuthRelationshipRequest(ctx context.Context, msg AuthRelationshipRequest) error {
	ctx, span := p.tracer.Start(
		ctx,
		"events.publishAuthRelationshipRequest",
		trace.WithAttributes(
			attribute.String(
				"events.action",
				string(msg.Action),
			),
			attribute.String(
				"events.subject_id",
				msg.SubjectID.String(),
			),
			attribute.String(
				"events.object_id",
				msg.ObjectID.String(),
			),
			attribute.String(
				"events.relationship_name",
				msg.RelationshipName,
			),
		),
	)
	defer span.End()

	var mapCarrier propagation.MapCarrier = make(map[string]string)

	otel.GetTextMapPropagator().Inject(ctx, mapCarrier)

	msg.TraceContext = mapCarrier

	if strings.ToLower(string(msg.Action)) != string(WriteAuthRelationshipAction) || strings.ToLower(string(msg.Action)) != string(DeleteAuthRelationshipAction) {
		span.RecordError(ErrInvalidAuthRelationshipAction)
		span.SetStatus(codes.Error, ErrInvalidAuthRelationshipAction.Error())

		return ErrInvalidAuthRelationshipAction
	}

	topic := strings.Join([]string{p.prefix, "permissions.relationship", string(msg.Action)}, ".")

	span.SetAttributes(
		attribute.String(
			"events.topic",
			topic,
		),
	)

	v, err := json.Marshal(msg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	m := message.NewMessage(watermill.NewUUID(), v)

	span.SetAttributes(
		attribute.String(
			"events.message_id",
			m.UUID,
		),
	)

	if err := p.publisher.Publish(topic, m); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

func (p *Publisher) PublishAuthRelationshipResponse(ctx context.Context, msg AuthRelationshipResponse) AuthRelationshipResponse {
	ctx, span := p.tracer.Start(
		ctx,
		"events.publishAuthRelationshipResponse",
	)
	defer span.End()

	// Propagate trace context into the message for the subscriber
	var mapCarrier propagation.MapCarrier = make(map[string]string)

	otel.GetTextMapPropagator().Inject(ctx, mapCarrier)

	msg.TraceContext = mapCarrier

	return msg
}

// PublishChange will publish a ChangeMessage to the topic for the change
func (p *Publisher) PublishChange(ctx context.Context, subjectType string, change ChangeMessage) error {
	ctx, span := p.tracer.Start(
		ctx,
		"events.publishChange",
		trace.WithAttributes(
			attribute.String(
				"events.subject_type",
				subjectType,
			),
			attribute.String(
				"events.subject_id",
				change.SubjectID.String(),
			),
			attribute.String(
				"events.event_type",
				change.EventType,
			),
			attribute.String(
				"events.source",
				change.Source,
			),
		),
	)

	defer span.End()

	// Propagate trace context into the message for the subscriber
	var mapCarrier propagation.MapCarrier = make(map[string]string)

	otel.GetTextMapPropagator().Inject(ctx, mapCarrier)

	change.TraceContext = mapCarrier

	if change.EventType == "" {
		span.RecordError(ErrMissingEventType)
		span.SetStatus(codes.Error, ErrMissingEventType.Error())

		return ErrMissingEventType
	}

	topic := strings.Join([]string{p.prefix, "changes", change.EventType, subjectType}, ".")

	span.SetAttributes(
		attribute.String(
			"events.topic",
			topic,
		),
	)

	change.Source = p.source
	if change.ActorID == gidx.NullPrefixedID {
		id, ok := ctx.Value(echojwtx.ActorCtxKey).(string)
		if ok {
			change.ActorID = gidx.PrefixedID(id)
		} else {
			change.ActorID = "unknown-actor"
		}
	}

	span.SetAttributes(
		attribute.String(
			"events.actor_id",
			change.ActorID.String(),
		),
	)

	v, err := json.Marshal(change)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), v)

	span.SetAttributes(
		attribute.String(
			"events.message_id",
			msg.UUID,
		),
	)

	if err := p.publisher.Publish(topic, msg); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

// PublishEvent will publish an EventMessage to the proper topic for that event
func (p *Publisher) PublishEvent(ctx context.Context, subjectType string, event EventMessage) error {
	ctx, span := p.tracer.Start(
		ctx,
		"events.publishEvent",
		trace.WithAttributes(
			attribute.String(
				"events.subject_type",
				subjectType,
			),
			attribute.String(
				"events.subject_id",
				event.SubjectID.String(),
			),
			attribute.String(
				"events.event_type",
				event.EventType,
			),
			attribute.String(
				"events.source",
				event.Source,
			),
		),
	)

	defer span.End()

	// Propagate trace context into the message for the subscriber
	var mapCarrier propagation.MapCarrier = make(map[string]string)

	otel.GetTextMapPropagator().Inject(ctx, mapCarrier)

	event.TraceContext = mapCarrier

	if event.EventType == "" {
		span.RecordError(ErrMissingEventType)
		span.SetStatus(codes.Error, ErrMissingEventType.Error())

		return ErrMissingEventType
	}

	topic := strings.Join([]string{p.prefix, "events", subjectType, event.EventType}, ".")

	span.SetAttributes(
		attribute.String(
			"events.topic",
			topic,
		),
	)

	event.Source = p.source

	v, err := json.Marshal(event)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), v)

	span.SetAttributes(
		attribute.String(
			"events.message_id",
			msg.UUID,
		),
	)

	if err := p.publisher.Publish(topic, msg); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

// Close will close the publisher
func (p *Publisher) Close() error {
	return p.publisher.Close()
}
