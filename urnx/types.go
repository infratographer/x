package urnx

import "github.com/google/uuid"

const prefix = "urn"
const urnLength = 4

// URN is an infratographer based URN consisting of a namespace, resource type and resource ID
type URN struct {
	Namespace    string
	ResourceType string
	ResourceID   uuid.UUID
}
