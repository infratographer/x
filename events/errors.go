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
	// ErrMissingAuthRelationshipRequestRelation is returned when the event message has no relations defined.
	ErrMissingAuthRelationshipRequestRelation = errors.New("auth relationship request message Relations field required")
	// ErrMissingAuthRelationshipRequestRelationRelation is returned when the event message Relations has the incorrect field for Relation value.
	ErrMissingAuthRelationshipRequestRelationRelation = errors.New("auth relationship request message Relations Relation field required")
	// ErrMissingAuthRelationshipRequestRelationSubjectID is returned when the event message Relations has the incorrect field SubjectID value.
	ErrMissingAuthRelationshipRequestRelationSubjectID = errors.New("auth relationship request message Relations SubjectID field required")
)
