package events_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.infratographer.com/x/gidx"

	"go.infratographer.com/x/events"
	"go.infratographer.com/x/testing/eventtools"
)

var errTimeout = errors.New("timeout waiting for event")

func TestNatsPublishAndSubscribe(t *testing.T) {
	ctx := context.Background()
	pubCfg, subCfg, err := eventtools.NewNatsServer()
	require.NoError(t, err)

	publisher, err := events.NewPublisher(pubCfg)
	require.NoError(t, err)

	change := testCreateChange()

	err = publisher.PublishChange(ctx, "test", change)
	require.NoError(t, err)

	change2 := testCreateChange()

	err = publisher.PublishChange(ctx, "test", change2)
	require.NoError(t, err)

	change3 := testCreateChange()
	change3.ActorID = ""

	err = publisher.PublishChange(ctx, "test", change3)
	require.NoError(t, err)

	sub, err := events.NewSubscriber(subCfg)
	require.NoError(t, err)

	messages, err := sub.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	receivedMsg, err := getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err := events.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change, chgMsg)
	assert.True(t, receivedMsg.Ack())

	receivedMsg, err = getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err = events.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change2, chgMsg)
	assert.True(t, receivedMsg.Ack())

	receivedMsg, err = getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err = events.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.NotEqualValues(t, change3, chgMsg)
	assert.Equal(t, "unknown-actor", chgMsg.ActorID.String())
	assert.True(t, receivedMsg.Ack())
}

func TestNatsMultipleSubscribers(t *testing.T) {
	ctx := context.Background()
	pubCfg, subCfg, err := eventtools.NewNatsServer()
	require.NoError(t, err)

	publisher, err := events.NewPublisher(pubCfg)
	require.NoError(t, err)

	change := testCreateChange()

	err = publisher.PublishChange(ctx, "test", change)
	require.NoError(t, err)

	sub, err := events.NewSubscriber(subCfg)
	require.NoError(t, err)

	messages, err := sub.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	receivedMsg, err := getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err := events.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change, chgMsg)
	assert.True(t, receivedMsg.Ack())

	sub2, err := events.NewSubscriber(subCfg)
	require.NoError(t, err)

	messages, err = sub2.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	receivedMsg, err = getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err = events.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change, chgMsg)
	assert.True(t, receivedMsg.Ack())
}

func TestNatsGroupedSubscribers(t *testing.T) {
	ctx := context.Background()
	pubCfg, subCfg, err := eventtools.NewNatsServer()
	require.NoError(t, err)

	publisher, err := events.NewPublisher(pubCfg)
	require.NoError(t, err)

	change := testCreateChange()

	err = publisher.PublishChange(ctx, "test", change)
	require.NoError(t, err)

	// put both subscribers in the same queue group so that combined the message is only delivered once
	subCfg.QueueGroup = "queue-test"

	sub, err := events.NewSubscriber(subCfg)
	require.NoError(t, err)

	messages, err := sub.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	receivedMsg, err := getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err := events.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change, chgMsg)
	assert.True(t, receivedMsg.Ack())

	sub2, err := events.NewSubscriber(subCfg)
	require.NoError(t, err)

	messages, err = sub2.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	receivedMsg, err = getSingleMessage(messages, time.Second*1)
	assert.Error(t, err, "this should fail since the other subscriber in the group already received the message")
	assert.ErrorContains(t, err, "timeout")
	assert.Nil(t, receivedMsg)
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
	}
}

func testCreateChange() events.ChangeMessage {
	return testChange("create")
}
