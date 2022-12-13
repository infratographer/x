package pubsubx

import (
	"time"
)

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
