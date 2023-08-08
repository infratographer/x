package events

import "errors"

var (
	// ErrNATSInvalidAuthConfiguration is returned when the config has both Tokena nd CredsFile specified.
	ErrNATSInvalidAuthConfiguration = errors.New("invalid nats confinguration, both token and creds file are specified")

	// ErrNATSInvalidDeliveryPolicy is returned when an incorrect delivery policy is provided.
	ErrNATSInvalidDeliveryPolicy = errors.New("invalid delivery policy, expected all|last|last-per-subject|new|start-sequence|start-time")

	// ErrNATSMessageNoReplySubject is returned when calling ReplyAuthRelationshipRequest when the request has no reply subject defined.
	ErrNATSMessageNoReplySubject = errors.New("unable to reply to auth relationship request, no reply subject specified")
)
