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

// Package events provides common utilities and formats for working with infratographer events
package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/multierr"

	"go.infratographer.com/x/gidx"
)

// Message contains a message which has been published or received from a subscription.
type Message[T any] interface {
	// Connection returns the underlying connection the message was received on.
	Connection() Connection

	// ID returns the unique message id.
	ID() string
	// Topic returns the topic the message was sent to.
	Topic() string
	// Message returns the decoded message object.
	Message() T
	// Ack acks the message.
	Ack() error
	// Nak nacks the message.
	Nak(delay time.Duration) error
	// Term terminates the message.
	Term() error
	// Timestamp returns the time the message was submitted.
	Timestamp() time.Time
	// Deliveries returns the number of times the message was delivered.
	Deliveries() uint64

	// Error returns any error encountered while decoding the message
	Error() error

	// Source returns the underlying message object.
	Source() any
}

// Request extends Message by allowing replies to be sent for the received message.
type Request[TRequest, TResponse any] interface {
	Message[TRequest]

	// Reply publishes a response to the received message.
	Reply(ctx context.Context, message TResponse) (Message[TResponse], error)
}

// ChangeType represents the possible event types for a ChangeMessage
type ChangeType string

// AuthRelationshipAction represents the possible auth relationship request actions for an AuthRelationshipRequest
type AuthRelationshipAction string

var (
	// CreateChangeType provides the event type for create events
	CreateChangeType ChangeType = "create"
	// UpdateChangeType provides the event type for update events
	UpdateChangeType ChangeType = "update"
	// DeleteChangeType provides the event type for delete events
	DeleteChangeType ChangeType = "delete"
	// WriteAuthRelationshipAction provides the auth relationship action for write requests
	WriteAuthRelationshipAction AuthRelationshipAction = "write"
	// DeleteAuthRelationshipAction provides the auth relationship action for delete requests
	DeleteAuthRelationshipAction AuthRelationshipAction = "delete"
)

// FieldChange represents a single field that was changed in a changeset and is used to map fields to the old and new values
type FieldChange struct {
	// Field is the name of the field that changed
	Field string `json:"field"`
	// PreviousValue is the value the field had before the change
	PreviousValue string `json:"previousValue"`
	// CurrentValue is the new value of the field after the change
	CurrentValue string `json:"currentValue"`
}

// ChangeMessage contains the data structure expected to be received when picking
// an event from a changes message queue
type ChangeMessage struct {
	// SubjectID is the PrefixedID representing the node of the topic of this message
	SubjectID gidx.PrefixedID `json:"subjectID"`
	// EventType describes the type of event that has triggered this message
	EventType string `json:"eventType"`
	// AdditionalSubjectIDs is a group of PrefixedIDs representing additional nodes associated with this message
	AdditionalSubjectIDs []gidx.PrefixedID `json:"additionalSubjects"`
	// ActorID is the PrefixedID representing the identity of the actor that caused this message to be triggered
	ActorID gidx.PrefixedID `json:"actorID"`
	// Source is a string representing the identity of the source system that created the message
	Source string `json:"source"`
	// Timestamp is the time representing when the message was created
	Timestamp time.Time `json:"timestamp"`
	// TraceContext is a map of values used for OpenTelemetry context propagation.
	TraceContext map[string]string `json:"traceContext"`
	// TraceID is the ID of the trace for this event
	// Deprecated: Use TraceContext with OpenTelemetry context propagation instead.
	TraceID string `json:"traceID"`
	// SpanID is the ID of the span that additional traces should based off of
	// Deprecated: Use TraceContext with OpenTelemetry context propagation instead.
	SpanID string `json:"spanID"`
	// SubjectFields is a map of the fields on the subject
	SubjectFields map[string]string `json:"subjectFields"`
	// Changeset is an optional map of the fields that changed triggering this message, this should be provided if the source can provide a changeset
	FieldChanges []FieldChange `json:"fieldChanges"`
	// AdditionalData is a field to store any addition information that may be important to include with your message
	AdditionalData map[string]interface{} `json:"additionalData"`
}

// GetTraceContext creates a new OpenTelementry context for the message.
func (m ChangeMessage) GetTraceContext(ctx context.Context) context.Context {
	tp := otel.GetTextMapPropagator()

	return tp.Extract(ctx, propagation.MapCarrier(m.TraceContext))
}

// GetSubject returns the subject of the message
func (m ChangeMessage) GetSubject() gidx.PrefixedID {
	return m.SubjectID
}

// GetSubject returns the subject of the message
func (m EventMessage) GetSubject() gidx.PrefixedID {
	return m.SubjectID
}

// GetAddSubjects returns the additional subjects of the message
func (m ChangeMessage) GetAddSubjects() []gidx.PrefixedID {
	return m.AdditionalSubjectIDs
}

// GetAddSubjects returns the additional subjects of the message
func (m EventMessage) GetAddSubjects() []gidx.PrefixedID {
	return m.AdditionalSubjectIDs
}

// GetEventType returns the event type of the message
func (m ChangeMessage) GetEventType() string {
	return m.EventType
}

// GetEventType returns the event type of the message
func (m EventMessage) GetEventType() string {
	return m.EventType
}

// Validate ensures the message has all the required fields.
func (m ChangeMessage) Validate() error {
	var err error

	if m.SubjectID == "" {
		err = multierr.Append(err, ErrMissingChangeMessageSubjectID)
	}

	if m.EventType == "" {
		err = multierr.Append(err, ErrMissingChangeMessageEventType)
	}

	return err
}

// EventMessage contains the data structure expected to be received when picking
// an event from an events message queue
type EventMessage struct {
	// SubjectID is the PrefixedID representing the node of the topic of this message
	SubjectID gidx.PrefixedID `json:"subjectID"`
	// EventType describes the type of event that has triggered this message
	EventType string `json:"eventType"`
	// AdditionalSubjectIDs is a group of PrefixedIDs representing additional nodes associated with this message
	AdditionalSubjectIDs []gidx.PrefixedID `json:"additionalSubjects"`
	// Source is a string representing the identity of the source system that created the message
	Source string `json:"source"`
	// Timestamp is the time representing when the message was created
	Timestamp time.Time `json:"timestamp"`
	// TraceContext is a map of values used for OpenTelemetry context propagation.
	TraceContext map[string]string `json:"traceContext"`
	// TraceID is the ID of the trace for this event
	// Deprecated: Use TraceContext with OpenTelemetry context propagation instead.
	TraceID string `json:"traceID"`
	// SpanID is the ID of the span that additional traces should based off of
	// Deprecated: Use TraceContext with OpenTelemetry context propagation instead.
	SpanID string `json:"spanID"`
	// Data is a field to store any information that may be important to include about the event
	Data map[string]interface{} `json:"data"`
}

// GetTraceContext creates a new OpenTelementry context for the message.
func (m EventMessage) GetTraceContext(ctx context.Context) context.Context {
	tp := otel.GetTextMapPropagator()

	return tp.Extract(ctx, propagation.MapCarrier(m.TraceContext))
}

// Validate ensures the message has all the required fields.
func (m EventMessage) Validate() error {
	var err error

	if m.SubjectID == "" {
		err = multierr.Append(err, ErrMissingEventMessageSubjectID)
	}

	if m.EventType == "" {
		err = multierr.Append(err, ErrMissingEventMessageEventType)
	}

	return err
}

// AuthRelationshipRequest contains the data structure expected to be used to write or delete
// an auth relationship from PermissionsAPI
type AuthRelationshipRequest struct {
	// Action describes the type of action being performed. Valid options are "write" and "delete".
	Action AuthRelationshipAction `json:"action"`
	// ObjectID is the PrefixedID of the object the permissions will be granted on
	ObjectID gidx.PrefixedID `json:"objectID"`
	// Relations defines all relations which should be written or deleted for this object.
	Relations []AuthRelationshipRelation `json:"relations"`
	// ConditionName represents the name of a conditional check that will be applied to this relationship. (Optional)
	// In SpiceDB this would be a caveat name
	ConditionName string `json:"conditionName"`
	// ConditionValues are the condition values to be used on the condition check. (Optional)
	ConditionValues map[string]interface{} `json:"conditionValue"`
	// TraceContext is a map of values used for OpenTelemetry context propagation.
	TraceContext map[string]string `json:"traceContext"`
	// TraceID is the ID of the trace for this event
	// Deprecated: Use TraceContext with OpenTelemetry context propagation instead.
	TraceID string `json:"traceID"`
	// SpanID is the ID of the span that additional traces should based off of
	// Deprecated: Use TraceContext with OpenTelemetry context propagation instead.
	SpanID string `json:"spanID"`
}

// GetTraceContext creates a new OpenTelementry context for the message.
func (m AuthRelationshipRequest) GetTraceContext(ctx context.Context) context.Context {
	tp := otel.GetTextMapPropagator()

	return tp.Extract(ctx, propagation.MapCarrier(m.TraceContext))
}

// Validate ensures the message has all the required fields.
func (m AuthRelationshipRequest) Validate() error {
	var err error

	if m.Action == "" || m.Action != WriteAuthRelationshipAction && m.Action != DeleteAuthRelationshipAction {
		err = multierr.Append(err, ErrInvalidAuthRelationshipRequestAction)
	}

	if m.ObjectID == "" {
		err = multierr.Append(err, ErrMissingAuthRelationshipRequestObjectID)
	}

	if len(m.Relations) == 0 {
		err = multierr.Append(err, ErrMissingAuthRelationshipRequestRelation)
	}

	for i, rel := range m.Relations {
		if rErr := rel.Validate(); rErr != nil {
			err = multierr.Append(err, fmt.Errorf("%w: relation %d", rErr, i))
		}
	}

	return err
}

// AuthRelationshipRelation defines the relation an object from an AuthRelationshipRequest has to a subject.
type AuthRelationshipRelation struct {
	// Relation is the name of the relation the object from AuthRelationshipRequest has to the subject.
	Relation string `json:"relation"`
	// The subject the relation is to.
	SubjectID gidx.PrefixedID `json:"subjectID"`
}

// Validate ensures the message has all the required fields.
func (r AuthRelationshipRelation) Validate() error {
	var err error

	if r.Relation == "" {
		err = multierr.Append(err, ErrMissingAuthRelationshipRequestRelationRelation)
	}

	if r.SubjectID == "" {
		err = multierr.Append(err, ErrMissingAuthRelationshipRequestRelationSubjectID)
	}

	return err
}

// Errors contains one or more errors and handles marshalling the errors.
// See [Errors.MarshalJSON] and [Errors.UnmarshalJSON] for details on how marshalling is done.
type Errors []error

// MarshalJSON encodes a string of arrays with each errors Error string.
// Entries which are nil are skipped.
// If no non nil errors are provided, null is returned.
func (e Errors) MarshalJSON() ([]byte, error) {
	errs := make([]string, 0, len(e))

	for _, err := range e {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) == 0 {
		return []byte("null"), nil
	}

	return json.Marshal(errs)
}

// UnmarshalJSON converts a list of string errors into new errors.
// All errors unmarshalled are new errors and cannot be compared directly to another error.
// Errors should be checked using string comparison.
func (e *Errors) UnmarshalJSON(b []byte) error {
	var errs []string

	if err := json.Unmarshal(b, &errs); err != nil {
		return err
	}

	if len(errs) == 0 {
		*e = nil

		return nil
	}

	*e = make(Errors, len(errs))

	for i, err := range errs {
		(*e)[i] = errors.New(err) //nolint:err113 // errors are dynamically returned
	}

	return nil
}

// Error returns each error on a new line.
// Nil error are not included.
func (e Errors) Error() string {
	errs := make([]string, 0, len(e))

	for _, err := range e {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	return strings.Join(errs, "\n")
}

// AuthRelationshipResponse contains the data structure expected to be received from an AuthRelationshipRequest
// message to write or delete an auth relationship from PermissionsAPI
type AuthRelationshipResponse struct {
	// Errors contains any errors, if empty the request was successful
	Errors Errors `json:"errors"`
	// TraceContext is a map of values used for OpenTelemetry context propagation.
	TraceContext map[string]string `json:"traceContext"`
	// TraceID is the ID of the trace for this event
	// Deprecated: Use TraceContext with OpenTelemetry context propagation instead.
	TraceID string `json:"traceID"`
	// SpanID is the ID of the span that additional traces should based off of
	// Deprecated: Use TraceContext with OpenTelemetry context propagation instead.
	SpanID string `json:"spanID"`
}

// GetTraceContext creates a new OpenTelementry context for the message.
func (m AuthRelationshipResponse) GetTraceContext(ctx context.Context) context.Context {
	tp := otel.GetTextMapPropagator()

	return tp.Extract(ctx, propagation.MapCarrier(m.TraceContext))
}

// Validate ensures the message has all the required fields.
func (m AuthRelationshipResponse) Validate() error {
	return nil
}

// UnmarshalChangeMessage returns a ChangeMessage from a json []byte.
func UnmarshalChangeMessage(b []byte) (ChangeMessage, error) {
	var c ChangeMessage

	err := json.Unmarshal(b, &c)

	return c, err
}

// UnmarshalEventMessage returns a EventMessage from a json []byte.
func UnmarshalEventMessage(b []byte) (EventMessage, error) {
	var m EventMessage

	err := json.Unmarshal(b, &m)

	return m, err
}

// UnmarshalAuthRelationshipRequest returns an AuthRelationshipRequest from a json []byte.
func UnmarshalAuthRelationshipRequest(b []byte) (AuthRelationshipRequest, error) {
	var m AuthRelationshipRequest

	err := json.Unmarshal(b, &m)

	return m, err
}

// UnmarshalAuthRelationshipResponse returns an AuthRelationshipRsponse from a json []byte.
func UnmarshalAuthRelationshipResponse(b []byte) (AuthRelationshipResponse, error) {
	var m AuthRelationshipResponse

	err := json.Unmarshal(b, &m)

	return m, err
}
