package echojwtx

import (
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.infratographer.com/x/viperx"
)

// MustViperFlags adds jwks-uri to the provided flagset and binds to viper jwks.uri.
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.Bool("oidc", true, "use oidc auth")
	viperx.MustBindFlag(v, "oidc.enabled", flags.Lookup("oidc"))

	flags.String("oidc-aud", "", "expected audience on OIDC JWT")
	viperx.MustBindFlag(v, "oidc.audience", flags.Lookup("oidc-aud"))

	flags.String("oidc-issuer", "", "expected issuer of OIDC JWT")
	viperx.MustBindFlag(v, "oidc.issuer", flags.Lookup("oidc-issuer"))

	flags.String("oidc-jwks-uri", "", "URI for JWKS listing for JWTs")
	viperx.MustBindFlag(v, "oidc.jwks.uri", flags.Lookup("oidc-jwks-uri"))

	flags.Duration("oidc-jwks-remote-timeout", 1*time.Minute, "timeout for remote JWKS fetching")
	viperx.MustBindFlag(v, "oidc.jwks.remote-timeout", flags.Lookup("oidc-jwks-remote-timeout"))
}

// AuthConfigFromViper builds a new AuthConfig from viper.
func AuthConfigFromViper(v *viper.Viper) *AuthConfig {
	if !v.GetBool("oidc.enabled") {
		return nil
	}

	return &AuthConfig{
		Audience: viper.GetString("oidc.audience"),
		Issuer:   viper.GetString("oidc.issuer"),
		JWKSURI:  viper.GetString("oidc.jwks.uri"),
		KeyFuncOptions: keyfunc.Options{
			RefreshTimeout: viper.GetDuration("oidc.jwks.remote-timeout"),
		},
	}
}
