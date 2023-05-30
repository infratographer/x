package clientcredentialsx

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/clientcredentials"

	"go.infratographer.com/x/viperx"
)

// MustViperFlags adds oidc oauth2 client config to the provided flagset and binds to viper
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.Bool("oidc-client", true, "use oidc client auth")
	viperx.MustBindFlag(v, "oidc-client.enabled", flags.Lookup("oidc-client"))

	flags.String("oidc-client-id", "", "expected oidc client identifier")
	viperx.MustBindFlag(v, "oidc-client.id", flags.Lookup("oidc-client-id"))

	flags.String("oidc-client-secret", "", "expected oidc client secret")
	viperx.MustBindFlag(v, "oidc-client.secret", flags.Lookup("oidc-client-secret"))

	flags.String("oidc-client-token-url", "", "expected oidc token url")
	viperx.MustBindFlag(v, "oidc-client.token-url", flags.Lookup("oidc-client-token-url"))
}

// ConfigFromViper returns an oauth2 client credentials config from the provided viper config
func ConfigFromViper(v *viper.Viper) (*clientcredentials.Config, error) {
	if !v.GetBool("oidc-client.enabled") {
		return nil, nil
	}

	if v.GetString("oidc-client.id") == "" {
		return nil, ErrClientIDRequired
	}

	if v.GetString("oidc-client.secret") == "" {
		return nil, ErrClientSecretRequired
	}

	if v.GetString("oidc-client.token-url") == "" {
		return nil, ErrTokenURLRequired
	}

	return &clientcredentials.Config{
		ClientID:     v.GetString("oidc-client.id"),
		ClientSecret: v.GetString("oidc-client.secret"),
		TokenURL:     v.GetString("oidc-client.token-url"),
		Scopes:       []string{},
	}, nil
}
