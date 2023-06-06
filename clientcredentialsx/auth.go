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
	// ClientID is the applications's ID
	ClientID string

	// ClientSecret is the application's secret
	ClientSecret string

	// TokenURL is the resource server's token endpoint URL. This is a constant specific to each server.
	TokenURL string

	// Scopes specifies optional requested permissions
	Scopes []string
}

// Auth is an oauth2 http client
type Auth struct {
	cfg clientcredentials.Config
}

// NewAuth returns a new oauth2 http client
func NewAuth(cfg AuthConfig) *Auth {
	ccCfg := clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenURL,
		Scopes:       cfg.Scopes,
	}

	cli := &Auth{
		cfg: ccCfg,
	}

	return cli
}

// HTTPClient returns an http client using the configured token.
// The token will auto-refresh as necessary.
func (a Auth) HTTPClient(ctx context.Context) *http.Client {
	return a.cfg.Client(ctx)
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
