package urnx

import "errors"

// ErrInvalidURNPrefix is returned when the URN prefix is invalid
var ErrInvalidURNPrefix = errors.New("invalid urn prefix: expected '" + prefix + "'")

// ErrInvalidURN is returned when the URN is invalid
var ErrInvalidURN = errors.New("invalid urn: expected 'urn:<namespace>:<resource type>:<resource id>")

// ErrInvalidURNNamespace is returned when the URN namespace is invalid and does not match
// the regex [A-za-z0-9-]{1,30}
var ErrInvalidURNNamespace = errors.New("invalid urn namespace: expected string consisting of [A-za-z0-9-]{1,30}")

// ErrInvalidURNResourceType is returned when the URN resource type is invalid and does not match
// the regex [A-za-z0-9-]{1,}
var ErrInvalidURNResourceType = errors.New("invalid urn resource type: expected string consisting of [A-za-z0-9-]{1,}")

// ErrInvalidURNResourceID is returned when the URN resource ID is invalid and not a valid UUID
var ErrInvalidURNResourceID = errors.New("invalid urn resource id: expected valid uuid")
