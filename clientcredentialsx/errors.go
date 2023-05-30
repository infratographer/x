package clientcredentialsx

import (
	"errors"
)

var (
	// ErrClientIDRequired is returned when the OIDC client id is missing
	ErrClientIDRequired = errors.New("oauth2 client id is required and cannot be empty")

	// ErrClientSecretRequired is returned when the OIDC client secret is missing
	ErrClientSecretRequired = errors.New("oauth2 client secret is required and cannot be empty")

	// ErrTokenURLRequired is returned when the OIDC token url is missing
	ErrTokenURLRequired = errors.New("oauth2 token url is required and cannot be empty")
)
