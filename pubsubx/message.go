// Package pubsubx provides common utilities and formats for working with pubsub systems
package pubsubx

import (
	"time"
)

// Message contains the data structure expected to be received when picking
// an event from a message queue
type Message struct {
        // SubjectURN is a string representing the identity of the topic of this message
	SubjectURN            string                 `json:"subject_urn"`
        // EventType describes the type of event that has triggered this message
	EventType             string                 `json:"event_type"` 
        // AdditionalSubjectURNs is a group of strings representing additional identities associated with this message
	AdditionalSubjectURNs []string               `json:"additional_subjects"`
        // ActorURN is a string representing the identity of the actor that created this message
	ActorURN              string                 `json:"actor_urn"`
        // Source is a string representing the identity of the source system that created the message
	Source                string                 `json:"source"`
        // Timestamp is the time representing when the message was created
	Timestamp             time.Time              `json:"timestamp"`
	SubjectFields         map[string]string      `json:"fields"`
        // AdditionalData is a field to store any addition information that may be important to include with your message
	AdditionalData        map[string]interface{} `json:"additional_data"`
}
