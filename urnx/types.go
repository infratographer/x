package urnx

import "github.com/google/uuid"

const PREFIX = "urn"
const URNLENGTH = 4

// URN is an infratographer based URN consisting of a namespace, resource type and resource ID
type URN struct {
	Prefix       string
	Namespace    string
	ResourceType string
	ResourceID   uuid.UUID
}

