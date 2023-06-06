package clientcredentialsx

import (
	"context"
	"net/http"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"go.infratographer.com/x/viperx"

	"golang.org/x/oauth2/clientcredentials"
)

// AuthConfig is an oauth2 client config
type AuthConfig struct {
	config *clientcredentials.Config
	ctx    context.Context
}

// MustViperFlags adds oidc oauth2 client config to the provided flagset and binds to viper
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.Bool("oidc-client", true, "use oidc client auth")
	viperx.MustBindFlag(v, "oidc.client.enabled", flags.Lookup("oidc-client"))

	flags.String("oidc-client-id", "", "expected oidc client identifier")
	viperx.MustBindFlag(v, "oidc.client.id", flags.Lookup("oidc-client-id"))

	flags.String("oidc-client-secret", "", "expected oidc client secret")
	viperx.MustBindFlag(v, "oidc.client.secret", flags.Lookup("oidc-client-secret"))

	flags.String("oidc-client-token-url", "", "expected oidc token url")
	viperx.MustBindFlag(v, "oidc.client.token-url", flags.Lookup("oidc-client-token-url"))
}

// NewAuth returns a new jwt auth
func NewAuth(ctx context.Context, cfg *clientcredentials.Config) *AuthConfig {
	return &AuthConfig{
		config: cfg,
		ctx:    ctx,
	}
}

// HTTPClient returns an oauth2 http client with valid cfg, otherwise returns the default http client
func (a AuthConfig) HTTPClient() *http.Client {
	if a.config != nil {
		return a.config.Client(a.ctx)
	}

	return http.DefaultClient
}
