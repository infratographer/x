package pubsubx_test

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

	"go.infratographer.com/x/pubsubx"
	"go.infratographer.com/x/testing/pubsubxtools"
)

func TestNatsPublishAndSubscribe(t *testing.T) {
	ctx := context.Background()
	pubCfg, subCfg, err := pubsubxtools.NewNatsServer()
	require.NoError(t, err)

	publisher, err := pubsubx.NewPublisher(pubCfg)
	require.NoError(t, err)

	change := testCreateChange()

	err = publisher.PublishChange(ctx, "test", change)
	require.NoError(t, err)

	change2 := testCreateChange()

	err = publisher.PublishChange(ctx, "test", change2)
	require.NoError(t, err)

	sub, err := pubsubx.NewSubscriber(subCfg)
	require.NoError(t, err)

	messages, err := sub.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	receivedMsg, err := getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err := pubsubx.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change, chgMsg)
	assert.True(t, receivedMsg.Ack())

	receivedMsg, err = getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err = pubsubx.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change2, chgMsg)
	assert.True(t, receivedMsg.Ack())
}

func TestNatsMultipleSubscribers(t *testing.T) {
	ctx := context.Background()
	pubCfg, subCfg, err := pubsubxtools.NewNatsServer()
	require.NoError(t, err)

	publisher, err := pubsubx.NewPublisher(pubCfg)
	require.NoError(t, err)

	change := testCreateChange()

	err = publisher.PublishChange(ctx, "test", change)
	require.NoError(t, err)

	sub, err := pubsubx.NewSubscriber(subCfg)
	require.NoError(t, err)

	messages, err := sub.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	receivedMsg, err := getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err := pubsubx.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change, chgMsg)
	assert.True(t, receivedMsg.Ack())

	sub2, err := pubsubx.NewSubscriber(subCfg)
	require.NoError(t, err)

	messages, err = sub2.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	receivedMsg, err = getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err = pubsubx.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change, chgMsg)
	assert.True(t, receivedMsg.Ack())
}

func TestNatsGroupedSubscribers(t *testing.T) {
	ctx := context.Background()
	pubCfg, subCfg, err := pubsubxtools.NewNatsServer()
	require.NoError(t, err)

	publisher, err := pubsubx.NewPublisher(pubCfg)
	require.NoError(t, err)

	change := testCreateChange()

	err = publisher.PublishChange(ctx, "test", change)
	require.NoError(t, err)

	// put both subscribers in the same queue group so that combined the message is only delivered once
	subCfg.QueueGroup = "queue-test"

	sub, err := pubsubx.NewSubscriber(subCfg)
	require.NoError(t, err)

	messages, err := sub.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	receivedMsg, err := getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)

	chgMsg, err := pubsubx.UnmarshalChangeMessage(receivedMsg.Payload)
	require.NoError(t, err)
	assert.EqualValues(t, change, chgMsg)
	assert.True(t, receivedMsg.Ack())

	sub2, err := pubsubx.NewSubscriber(subCfg)
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
		return nil, errors.New("timeout")
	}
}

func testChange(eventType string) pubsubx.ChangeMessage {
	js, err := gofakeit.JSON(nil)
	if err != nil {
		js = []byte("json-failed-so-have-a-string")
	}

	return pubsubx.ChangeMessage{
		EventType:            eventType,
		SubjectID:            gidx.MustNewID("testing"),
		ActorID:              gidx.MustNewID("testusr"),
		AdditionalSubjectIDs: []gidx.PrefixedID{gidx.MustNewID("testusr"), gidx.MustNewID("testtnt")},
		FieldChanges: []pubsubx.FieldChange{
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

func testCreateChange() pubsubx.ChangeMessage {
	return testChange("create")
}
