package echojwtx

import (
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.infratographer.com/x/viperx"
)

const (
	// DefaultOIDCJWKSRemoteTimeout defines the default timeout for fetching the OIDC JWKS file.
	DefaultOIDCJWKSRemoteTimeout = 5 * time.Second
)

// MustViperFlags adds jwks-uri to the provided flagset and binds to viper jwks.uri.
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.Bool("oidc", true, "use oidc auth")
	viperx.MustBindFlag(v, "oidc.enabled", flags.Lookup("oidc"))

	flags.String("oidc-aud", "", "expected audience on OIDC JWT")
	viperx.MustBindFlag(v, "oidc.audience", flags.Lookup("oidc-aud"))

	flags.String("oidc-issuer", "", "expected issuer of OIDC JWT")
	viperx.MustBindFlag(v, "oidc.issuer", flags.Lookup("oidc-issuer"))

	flags.Duration("oidc-jwks-remote-timeout", DefaultOIDCJWKSRemoteTimeout, "timeout for remote JWKS fetching")
	viperx.MustBindFlag(v, "oidc.jwks.remote-timeout", flags.Lookup("oidc-jwks-remote-timeout"))
}

// AuthConfigFromViper builds a new AuthConfig from viper.
func AuthConfigFromViper(v *viper.Viper) (*AuthConfig, error) {
	if !v.GetBool("oidc.enabled") || v.GetBool("dev") {
		return nil, nil
	}

	return &AuthConfig{
		Audience: v.GetString("oidc.audience"),
		Issuer:   v.GetString("oidc.issuer"),
		KeyFuncOptions: keyfunc.Options{
			RefreshTimeout: v.GetDuration("oidc.jwks.remote-timeout"),
		},
	}, nil
}
