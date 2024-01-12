package eventtools_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.infratographer.com/x/events"
	"go.infratographer.com/x/gidx"
	"go.infratographer.com/x/testing/eventtools"
)

var errTimeout = errors.New("timeout waiting for event")

func TestNats(t *testing.T) {
	ctx := context.Background()

	consumerName := events.NATSConsumerDurableName("nats-testing", eventtools.Prefix+".changes.>")

	nats, err := eventtools.NewNatsServer()
	require.NoError(t, err)

	defer nats.Close()

	natsCfg := nats.Config.NATS
	natsCfg.QueueGroup = "nats-testing"

	conn, err := events.NewNATSConnection(natsCfg)
	require.NoError(t, err)

	change1 := testCreateChange()

	chgMsg, err := conn.PublishChange(ctx, "test", change1)
	require.NoError(t, err)
	require.NoError(t, chgMsg.Error())
	require.Equal(t, change1, chgMsg.Message())

	change2 := testCreateChange()

	chgMsg, err = conn.PublishChange(ctx, "test", change2)
	require.NoError(t, err)
	require.NoError(t, chgMsg.Error())
	require.Equal(t, change2, chgMsg.Message())

	messages, err := conn.SubscribeChanges(context.Background(), ">")
	require.NoError(t, err)

	err = nats.SetConsumerSampleFrequency(consumerName, "100")
	require.NoError(t, err)

	receivedMsg, err := getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)
	require.NoError(t, receivedMsg.Error())
	assert.EqualValues(t, change1, receivedMsg.Message())
	assert.NoError(t, receivedMsg.Ack())

	err = nats.WaitForAck(consumerName, time.Second*2)
	require.NoError(t, err)

	receivedMsg, err = getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)
	require.NoError(t, receivedMsg.Error())
	assert.EqualValues(t, change2, receivedMsg.Message())
	assert.NoError(t, receivedMsg.Nak(0))

	err = nats.WaitForAck(consumerName, time.Second*2)
	require.ErrorIs(t, err, eventtools.ErrNack)

	err = nats.WaitForAck(consumerName, time.Second*2)
	require.ErrorIs(t, err, eventtools.ErrNoAck)
}

func getSingleMessage[T any](messages <-chan events.Message[T], timeout time.Duration) (events.Message[T], error) {
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
