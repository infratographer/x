package eventtools_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/brianvoe/gofakeit/v6"
	nc "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.infratographer.com/x/events"
	"go.infratographer.com/x/gidx"
	"go.infratographer.com/x/testing/eventtools"
)

var errTimeout = errors.New("timeout waiting for event")

func TestNats(t *testing.T) {
	ctx := context.Background()

	nats, err := eventtools.NewNatsServer()
	require.NoError(t, err)

	defer nats.Close()

	publisher, err := events.NewPublisher(nats.PublisherConfig)
	require.NoError(t, err)

	change1 := testCreateChange()

	err = publisher.PublishChange(ctx, "test", change1)
	require.NoError(t, err)

	change2 := testCreateChange()

	err = publisher.PublishChange(ctx, "test", change2)
	require.NoError(t, err)

	sub, err := events.NewSubscriber(nats.SubscriberConfig,
		nc.ManualAck(),
		nc.AckExplicit(),
		nc.Durable("test-consumer"),
	)
	require.NoError(t, err)

	messages, err := sub.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	err = nats.SetConsumerSampleFrequency("test-consumer", "100")
	require.NoError(t, err)

	receivedMsg, err := getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err := events.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change1, chgMsg)
	assert.True(t, receivedMsg.Ack())

	err = nats.WaitForAck("test-consumer", time.Second)
	require.NoError(t, err)

	receivedMsg, err = getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err = events.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change2, chgMsg)
	assert.True(t, receivedMsg.Nack())

	err = nats.WaitForAck("test-consumer", time.Second)
	require.ErrorIs(t, err, eventtools.ErrNack)

	err = nats.WaitForAck("test-consumer", time.Second)
	require.ErrorIs(t, err, eventtools.ErrNoAck)
}

func getSingleMessage(messages <-chan *message.Message, timeout time.Duration) (*message.Message, error) {
	select {
	case message := <-messages:
		return message, nil
	case <-time.After(timeout):
		return nil, errTimeout
	}
}

func testChange(eventType string) events.ChangeMessage {
	js, err := gofakeit.JSON(nil)
	if err != nil {
		js = []byte("json-failed-so-have-a-string")
	}

	return events.ChangeMessage{
		EventType:            eventType,
		SubjectID:            gidx.MustNewID("testing"),
		ActorID:              gidx.MustNewID("testusr"),
		AdditionalSubjectIDs: []gidx.PrefixedID{gidx.MustNewID("testusr"), gidx.MustNewID("testtnt")},
		FieldChanges: []events.FieldChange{
			{
				Field:         "name",
				PreviousValue: gofakeit.Name(),
				CurrentValue:  gofakeit.Name(),
			},
			{
				Field:         "random_data",
				PreviousValue: gofakeit.Adjective(),
				CurrentValue:  string(js),
			},
		},
		TraceContext: map[string]string{},
	}
}

func testCreateChange() events.ChangeMessage {
	return testChange("create")
}
