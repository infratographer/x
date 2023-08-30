package events

import (
	"context"
	"errors"
	"time"

	"github.com/nats-io/nats.go"
)

func (c *NATSConnection) coreSubscribe(ctx context.Context, subject string) (<-chan *nats.Msg, error) {
	logger := c.logger.With(
		"nats.provider", "core",
		"nats.subject", subject,
	)

	sub, err := c.conn.QueueSubscribeSync(subject, c.cfg.QueueGroup)
	if err != nil {
		return nil, err
	}

	msgCh := make(chan *nats.Msg, c.cfg.SubscriberFetchBatchSize)

	go func() {
		for {
			if err := c.nextMessage(ctx, sub, msgCh); err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					continue
				}

				logger.Errorw("error fetching messages", "error", err)

				select {
				case <-ctx.Done():
				case <-time.After(c.cfg.SubscriberFetchBackoff):
				}
			}

			select {
			case <-ctx.Done():
				close(msgCh)

				if err := sub.Unsubscribe(); err != nil {
					logger.Warnw("error unsubscribing", "error", err)
				}

				return
			default:
			}
		}
	}()

	return msgCh, nil
}

func (c *NATSConnection) jsSubscribe(ctx context.Context, subject string) (<-chan *nats.Msg, error) {
	durableName := c.durableName(subject)

	logger := c.logger.With(
		"nats.provider", "jetstream",
		"nats.subject", subject,
		"nats.durable_name", durableName,
	)

	sub, err := c.jetstream.PullSubscribe(subject, durableName, c.cfg.subscribeOptions...)
	if err != nil {
		return nil, err
	}

	msgCh := make(chan *nats.Msg, c.cfg.SubscriberFetchBatchSize)

	go func() {
		for {
			if err := c.fetchMessages(ctx, sub, msgCh); err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					continue
				}

				logger.Errorw("error fetching messages", "error", err)

				select {
				case <-ctx.Done():
				case <-time.After(c.cfg.SubscriberFetchBackoff):
				}
			}

			select {
			case <-ctx.Done():
				close(msgCh)

				if err := sub.Unsubscribe(); err != nil {
					logger.Warnw("error unsubscribing", "error", err)
				}

				return
			default:
			}
		}
	}()

	return msgCh, nil
}

func (c *NATSConnection) fetchMessages(ctx context.Context, sub *nats.Subscription, msgCh chan<- *nats.Msg) error {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.SubscriberFetchTimeout)

	defer cancel()

	batch, err := sub.FetchBatch(c.cfg.SubscriberFetchBatchSize, nats.Context(ctx))
	if err != nil {
		return err
	}

	for msg := range batch.Messages() {
		select {
		case msgCh <- msg:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return batch.Error()
}

func (c *NATSConnection) nextMessage(ctx context.Context, sub *nats.Subscription, msgCh chan<- *nats.Msg) error {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.SubscriberFetchTimeout)

	defer cancel()

	msg, err := sub.NextMsgWithContext(ctx)
	if err != nil {
		return err
	}

	select {
	case msgCh <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// SubscribeAuthRelationshipRequests creates a new pull subscription parsing incoming messages as AuthRelationshipRequest messages and returning a new Message channel.
func (c *NATSConnection) SubscribeAuthRelationshipRequests(ctx context.Context, topic string) (<-chan Request[AuthRelationshipRequest, AuthRelationshipResponse], error) {
	topic = c.buildSubscribeSubject("auth", "relationships", topic)

	natsCh, err := c.coreSubscribe(ctx, topic)
	if err != nil {
		return nil, err
	}

	c.logger.Debugf("subscribing to auth relation request message on topic %s", topic)

	return natsSubscriptionAuthRelationshipRequestChan(ctx, c, c.cfg.SubscriberFetchBatchSize, natsCh), nil
}

// SubscribeChanges creates a new pull subscription parsing incoming messages as ChangeMessage messages and returning a new Message channel.
func (c *NATSConnection) SubscribeChanges(ctx context.Context, topic string) (<-chan Message[ChangeMessage], error) {
	topic = c.buildSubscribeSubject("changes", topic)

	natsCh, err := c.jsSubscribe(ctx, topic)
	if err != nil {
		return nil, err
	}

	c.logger.Debugf("subscribing to changes message on topic %s", topic)

	return natsSubscriptionMessageChan[ChangeMessage](ctx, c, c.cfg.SubscriberFetchBatchSize, natsCh), nil
}

// SubscribeEvents creates a new pull subscription parsing incoming messages as EventMessage messages and returning a new Message channel.
func (c *NATSConnection) SubscribeEvents(ctx context.Context, topic string) (<-chan Message[EventMessage], error) {
	topic = c.buildSubscribeSubject("events", topic)

	natsCh, err := c.jsSubscribe(ctx, topic)
	if err != nil {
		return nil, err
	}

	c.logger.Debugf("subscribing to events message on topic %s", topic)

	return natsSubscriptionMessageChan[EventMessage](ctx, c, c.cfg.SubscriberFetchBatchSize, natsCh), nil
}
