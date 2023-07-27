package events

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"go.infratographer.com/x/echojwtx"
	"go.infratographer.com/x/gidx"
)

// PublishAuthRelationshipRequest publishes an AuthRelationshipRequest message and blocks until an AuthRelationshipResponse is provided.
func (c *NATSConnection) PublishAuthRelationshipRequest(ctx context.Context, topic string, message AuthRelationshipRequest) (Message[AuthRelationshipResponse], error) {
	ctx, span := c.tracer.Start(ctx, "events.nats.PublishAuthRelationshipRequest", trace.WithAttributes(
		attribute.String("events.subject_type", topic),
		attribute.String("events.subject_id", message.ObjectID.String()),
		attribute.String("events.event_type", string(message.Action)),
	))

	defer span.End()

	if err := message.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	// Propagate trace context into the message for the subscriber
	var mapCarrier propagation.MapCarrier = make(map[string]string)

	otel.GetTextMapPropagator().Inject(ctx, mapCarrier)

	message.TraceContext = mapCarrier

	topic = c.buildPublishSubject("auth", "relationships", string(message.Action), topic)

	reqMsg, err := newNATSMessage(c, topic, message)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	c.logger.Debugf("publishing auth relation request message to topic %s", topic)

	respMsg, err := reqMsg.request(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	return respMsg, nil
}

// PublishChange publishes a ChangeMessage.
func (c *NATSConnection) PublishChange(ctx context.Context, topic string, message ChangeMessage) (Message[ChangeMessage], error) {
	ctx, span := c.tracer.Start(ctx, "events.nats.PublishChange", trace.WithAttributes(
		attribute.String("events.subject_type", topic),
		attribute.String("events.subject_id", message.SubjectID.String()),
		attribute.String("events.event_type", message.EventType),
		attribute.String("events.source", message.Source),
	))

	defer span.End()

	if err := message.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	// Propagate trace context into the message for the subscriber
	var mapCarrier propagation.MapCarrier = make(map[string]string)

	otel.GetTextMapPropagator().Inject(ctx, mapCarrier)

	message.TraceContext = mapCarrier

	topic = c.buildPublishSubject("changes", message.EventType, topic)

	message.Source = c.cfg.Source

	if message.ActorID == gidx.NullPrefixedID {
		id, ok := ctx.Value(echojwtx.ActorCtxKey).(string)
		if ok {
			message.ActorID = gidx.PrefixedID(id)
		} else {
			message.ActorID = "unknown-actor"
		}
	}

	span.SetAttributes(
		attribute.String(
			"events.actor_id",
			message.ActorID.String(),
		),
	)

	msg, err := newNATSMessage(c, topic, message)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	c.logger.Debugf("publishing change message to topic %s", topic)

	if err = msg.publish(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return msg, err
	}

	return msg, nil
}

// PublishEvent publishes an EventMessage.
func (c *NATSConnection) PublishEvent(ctx context.Context, topic string, message EventMessage) (Message[EventMessage], error) {
	ctx, span := c.tracer.Start(ctx, "events.nats.PublishEvent", trace.WithAttributes(
		attribute.String("events.subject_type", topic),
		attribute.String("events.subject_id", message.SubjectID.String()),
		attribute.String("events.event_type", message.EventType),
		attribute.String("events.source", message.Source),
	))

	defer span.End()

	if err := message.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	// Propagate trace context into the message for the subscriber
	var mapCarrier propagation.MapCarrier = make(map[string]string)

	otel.GetTextMapPropagator().Inject(ctx, mapCarrier)

	message.TraceContext = mapCarrier

	topic = c.buildPublishSubject("events", message.EventType, topic)

	msg, err := newNATSMessage(c, topic, message)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	c.logger.Debugf("publishing event message to topic %s", topic)

	if err = msg.publish(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return msg, err
	}

	return msg, nil
}
