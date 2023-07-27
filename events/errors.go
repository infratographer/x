package events

import "errors"

var (
	// ErrProviderNotConfigured is an error packages should return if no events provider is configured.
	ErrProviderNotConfigured = errors.New("events provider not configured")

	// ErrMissingChangeMessageEventType is returned when the event message has the incorrect field EventType value.
	ErrMissingChangeMessageEventType = errors.New("change message EventType field required")
	// ErrMissingChangeMessageSubjectID is returned when the event message has the incorrect field SubjectID value.
	ErrMissingChangeMessageSubjectID = errors.New("change message SubjectID field required")

	// ErrMissingEventMessageEventType is returned when the event message has the incorrect field EventType value.
	ErrMissingEventMessageEventType = errors.New("event message EventType field required")
	// ErrMissingEventMessageSubjectID is returned when the event message has the incorrect field SubjectID value.
	ErrMissingEventMessageSubjectID = errors.New("event message SubjectID field required")

	// ErrInvalidAuthRelationshipRequestAction is returned when the event message has the incorrect field Action value.
	ErrInvalidAuthRelationshipRequestAction = errors.New("auth relationship request message Action field must be write or delete")
	// ErrMissingAuthRelationshipRequestObjectID is returned when the event message has the incorrect field ObjectID value.
	ErrMissingAuthRelationshipRequestObjectID = errors.New("auth relationship request message ObjectID field required")
	// ErrMissingAuthRelationshipRequestRelationshipName is returned when the event message has the incorrect field RelationshipName value.
	ErrMissingAuthRelationshipRequestRelationshipName = errors.New("auth relationship request message RelationshipName field required")
	// ErrMissingAuthRelationshipRequestSubjectID is returned when the event message has the incorrect field SubjectID value.
	ErrMissingAuthRelationshipRequestSubjectID = errors.New("auth relationship request message SubjectID field required")
)
