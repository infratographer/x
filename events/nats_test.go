package events_test

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

func TestNATSPublishAndSubscribe(t *testing.T) {
	ctx := context.Background()
	nats, err := eventtools.NewNatsServer()
	require.NoError(t, err)

	defer nats.Close()

	conn, err := events.NewNATSConnection(nats.Config.NATS)
	require.NoError(t, err)

	defer conn.Shutdown(ctx) //nolint:errcheck // within test

	change := testCreateChange()

	msg, err := conn.PublishChange(ctx, "test", change)
	require.NoError(t, err)
	require.Equal(t, change, msg.Message())

	change2 := testCreateChange()

	msg, err = conn.PublishChange(ctx, "test", change2)
	require.NoError(t, err)
	require.Equal(t, change2, msg.Message())

	change3 := testCreateChange()
	change3.ActorID = ""

	msg, err = conn.PublishChange(ctx, "test", change3)
	require.NoError(t, err)
	require.NotEqual(t, change3, msg.Message())

	messages, err := conn.SubscribeChanges(ctx, ">")
	require.NoError(t, err)

	receivedMsg, err := getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)
	require.NoError(t, receivedMsg.Error())
	assert.EqualValues(t, change, receivedMsg.Message())

	receivedMsg, err = getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)
	require.NoError(t, receivedMsg.Error())
	assert.EqualValues(t, change2, receivedMsg.Message())
	assert.NoError(t, receivedMsg.Ack())

	receivedMsg, err = getSingleMessage(messages, time.Second*1)
	require.NoError(t, err)
	require.NoError(t, receivedMsg.Error())
	assert.NotEqualValues(t, change3, receivedMsg.Message())
	assert.Equal(t, "unknown-actor", receivedMsg.Message().ActorID.String())
	assert.NoError(t, receivedMsg.Ack())
}

func TestNATSRequestReply(t *testing.T) {
	ctx := context.Background()
	nats, err := eventtools.NewNatsServer()
	require.NoError(t, err)

	defer nats.Close()

	conn, err := events.NewNATSConnection(nats.Config.NATS)
	require.NoError(t, err)

	defer conn.Shutdown(ctx) //nolint:errcheck // within test

	authRequest := events.AuthRelationshipRequest{
		Action:   events.WriteAuthRelationshipAction,
		ObjectID: gidx.PrefixedID("prntobj-abc123"),
		Relations: []events.AuthRelationshipRelation{
			{
				Relation:  "owner",
				SubjectID: gidx.PrefixedID("chldobj-abc123"),
			},
		},
		TraceContext: map[string]string{},
	}

	authResponse := events.AuthRelationshipResponse{
		TraceID:      "some-id",
		TraceContext: map[string]string{},
	}

	reqGot := make(chan events.Message[events.AuthRelationshipRequest], 1)
	respGot := make(chan events.Message[events.AuthRelationshipResponse], 1)

	authSubscribed := make(chan bool, 1)

	go func() {
		ctx, cancel := context.WithCancel(ctx)

		defer cancel()

		msgs, err := conn.SubscribeAuthRelationshipRequests(ctx, "*.test")

		close(authSubscribed)

		require.NoError(t, err)

		select {
		case reqMsg, ok := <-msgs:
			if !ok {
				return
			}

			reqGot <- reqMsg

			respMsg, err := reqMsg.ReplyAuthRelationshipRequest(ctx, authResponse)
			assert.NoError(t, err)
			assert.NotNil(t, respMsg)
		case <-time.After(time.Second * 2):
		}
	}()

	<-authSubscribed

	go func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*2)

		defer cancel()

		resp, err := conn.PublishAuthRelationshipRequest(ctx, "test", authRequest)
		assert.NoError(t, err)

		respGot <- resp
	}()

	select {
	case authRequestGot := <-reqGot:
		require.NotNil(t, authRequestGot)
		require.NoError(t, authRequestGot.Error())
		require.EqualValues(t, authRequest, authRequestGot.Message())
	case <-time.After(time.Second * 2):
		t.Error("timed out waiting for auth relationship request")
	}

	close(reqGot)

	select {
	case authResponseGot := <-respGot:
		require.NotNil(t, authResponseGot)
		require.NoError(t, authResponseGot.Error())
		require.EqualValues(t, authResponse, authResponseGot.Message())
	case <-time.After(time.Second * 2):
		t.Error("timed out waiting for auth relationship response")
	}

	close(respGot)
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
