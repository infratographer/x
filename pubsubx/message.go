// Package pubsubx provides common utilities and formats for working with pubsub systems
package pubsubx

import (
	"time"
)

// Message contains the data structure expected to be received when picking
// an event from a message queue
type Message struct {
	SubjectURN            string                 `json:"subject_urn"`
	EventType             string                 `json:"event_type"`
	AdditionalSubjectURNs []string               `json:"additional_subjects"`
	ActorURN              string                 `json:"actor_urn"`
	Source                string                 `json:"source"`
	Timestamp             time.Time              `json:"timestamp"`
	SubjectFields         map[string]string      `json:"fields"`
	AdditionalData        map[string]interface{} `json:"additional_data"`
}
