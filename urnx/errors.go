package urnx

import "errors"

// ErrInvalidURNPrefix is returned when the URN prefix is invalid
var ErrInvalidURNPrefix = errors.New("invalid urn prefix, expected '" + prefix + "'")

// ErrInvalidURN is returned when the URN is invalid
var ErrInvalidURN = errors.New("invalid urn, expected 'urn:<namespace>:<resource type>:<resource id>")
