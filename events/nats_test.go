package events_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	nc "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.infratographer.com/x/events"
	"go.infratographer.com/x/gidx"
	"go.infratographer.com/x/testing/eventtools"
)

var errTimeout = errors.New("timeout waiting for event")

func TestNATSPublishAndSubscribe(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name       string
		queueGroup string
	}{
		{
			name:       "with ephemeral consumer",
			queueGroup: "",
		},
		{
			name:       "with durable consumer",
			queueGroup: "testing-durable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nats, err := eventtools.NewNatsServer()
			require.NoError(t, err)

			defer nats.Close()

			natsCfg := nats.Config.NATS
			natsCfg.QueueGroup = tc.queueGroup
			conn, err := events.NewNATSConnection(natsCfg)
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
		})
	}
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

	resp, err := conn.PublishAuthRelationshipRequest(ctx, "test", authRequest)
	require.Error(t, err)
	require.ErrorIs(t, err, events.ErrRequestNoResponders)
	require.ErrorIs(t, err, nc.ErrNoResponders)
	require.Nil(t, resp)

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

			respMsg, err := reqMsg.Reply(ctx, authResponse)
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
func TestNATSRequestReplyMarshalling(t *testing.T) {
	testCases := []struct {
		name           string
		input          events.AuthRelationshipResponse
		expectDecoded  map[string]any
		expectResponse events.AuthRelationshipResponse
	}{
		{
			"no error",
			events.AuthRelationshipResponse{
				TraceID:      "some-id",
				TraceContext: map[string]string{},
			},
			map[string]any{
				"errors":       nil,
				"spanID":       "",
				"traceContext": map[string]any{},
				"traceID":      "some-id",
			},
			events.AuthRelationshipResponse{
				TraceID:      "some-id",
				TraceContext: map[string]string{},
			},
		},
		{
			"with errors",
			events.AuthRelationshipResponse{
				TraceID:      "some-id",
				TraceContext: map[string]string{},
				Errors: []error{
					os.ErrInvalid,
				},
			},
			map[string]any{
				"errors": []any{
					os.ErrInvalid.Error(),
				},
				"spanID":       "",
				"traceContext": map[string]any{},
				"traceID":      "some-id",
			},
			events.AuthRelationshipResponse{
				TraceID:      "some-id",
				TraceContext: map[string]string{},
				Errors: []error{
					errors.New(os.ErrInvalid.Error()), //nolint:goerr113 // ensure equals same error with text
				},
			},
		},
		{
			"nil errors skipped",
			events.AuthRelationshipResponse{
				TraceID:      "some-id",
				TraceContext: map[string]string{},
				Errors: []error{
					os.ErrInvalid,
					nil,
					os.ErrExist,
					nil,
				},
			},
			map[string]any{
				"errors": []any{
					os.ErrInvalid.Error(),
					os.ErrExist.Error(),
				},
				"spanID":       "",
				"traceContext": map[string]any{},
				"traceID":      "some-id",
			},
			events.AuthRelationshipResponse{
				TraceID:      "some-id",
				TraceContext: map[string]string{},
				Errors: []error{
					errors.New(os.ErrInvalid.Error()), //nolint:goerr113 // ensure equals same error with text
					errors.New(os.ErrExist.Error()),   //nolint:goerr113 // ensure equals same error with text
				},
			},
		},
		{
			"all nil errors skipped",
			events.AuthRelationshipResponse{
				TraceID:      "some-id",
				TraceContext: map[string]string{},
				Errors: []error{
					nil,
					nil,
				},
			},
			map[string]any{
				"errors":       nil,
				"spanID":       "",
				"traceContext": map[string]any{},
				"traceID":      "some-id",
			},
			events.AuthRelationshipResponse{
				TraceID:      "some-id",
				TraceContext: map[string]string{},
				Errors:       nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(tc.input)
			require.NoError(t, err, "unexpected error marshalling input")

			decoded := map[string]any{}

			err = json.Unmarshal(encoded, &decoded)
			require.NoError(t, err, "unexpected error unmarshalling encoded input into map")

			assert.Equal(t, tc.expectDecoded, decoded, "unexpected encoded response")

			var response events.AuthRelationshipResponse

			err = json.Unmarshal(encoded, &response)
			require.NoError(t, err, "unexpected error unmarshalling encoded input into response")

			assert.Equal(t, tc.expectResponse, response, "unexpected response")

			assert.Equal(t, len(tc.expectResponse.Errors), len(response.Errors), "unexpected response error count")
		})
	}
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
