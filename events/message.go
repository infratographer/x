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
	"encoding/json"
	"time"

	"go.infratographer.com/x/gidx"
)

// ChangeType represents the possible event types for a ChangeMessage
type ChangeType string

var (
	// CreateChangeType provides the event type for create events
	CreateChangeType ChangeType = "create"
	// UpdateChangeType provides the event type for update events
	UpdateChangeType ChangeType = "update"
	// DeleteChangeType provides the event type for delete events
	DeleteChangeType ChangeType = "delete"
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
	// TraceID is the ID of the trace for this event
	TraceID string `json:"traceID"`
	// SpanID is the ID of the span that additional traces should based off of
	SpanID string `json:"spanID"`
	// SubjectFields is a map of the fields on the subject
	SubjectFields map[string]string `json:"subjectFields"`
	// Changeset is an optional map of the fields that changed triggering this message, this should be provided if the source can provide a changeset
	FieldChanges []FieldChange `json:"fieldChanges"`
	// AdditionalData is a field to store any addition information that may be important to include with your message
	AdditionalData map[string]interface{} `json:"additionalData"`
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
	// TraceID is the ID of the trace for this event
	TraceID string `json:"traceID"`
	// SpanID is the ID of the span that additional traces should based off of
	SpanID string `json:"spanID"`
	// Data is a field to store any information that may be important to include about the event
	Data map[string]interface{} `json:"data"`
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
