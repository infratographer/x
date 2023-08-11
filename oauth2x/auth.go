package oauth2x

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.infratographer.com/x/viperx"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	tokenEndpointClient = &http.Client{
		Timeout:   5 * time.Second, // nolint:gomnd // clear and unexported
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	// ErrTokenEndpointMissing is returned when the issuers .well-known/openid-configuration is missing the token_endpoint key.
	ErrTokenEndpointMissing = errors.New("token endpoint missing from issuer well-known openid-configuration")
)

// NewClientCredentialsTokenSrc returns an oauth2 client credentials token source
func NewClientCredentialsTokenSrc(ctx context.Context, cfg Config) (oauth2.TokenSource, error) {
	tokenEndpoint, err := fetchIssuerTokenEndpoint(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}

	ccCfg := clientcredentials.Config{
		ClientID:     cfg.ID,
		ClientSecret: cfg.Secret,
		TokenURL:     tokenEndpoint,
	}

	return ccCfg.TokenSource(ctx), nil
}

// NewClient returns a http client using requested token source
func NewClient(_ context.Context, tokenSrc oauth2.TokenSource) *http.Client {
	return &http.Client{
		Transport: &oauth2.Transport{
			Base:   otelhttp.NewTransport(http.DefaultTransport),
			Source: oauth2.ReuseTokenSource(nil, tokenSrc),
		},
	}
}

// Config handles reading in all the config values available
// for setting up an oauth2 configuration
type Config struct {
	ID     string `mapstructure:"id"`
	Secret string `mapstructure:"secret"`
	Issuer string `mapstructure:"issuer"`
}

// MustViperFlags adds oidc oauth2 client credentials config to the provided flagset and binds to viper
func MustViperFlags(v *viper.Viper, flags *pflag.FlagSet) {
	flags.String("oidc-client-id", "", "oidc client identifier")
	viperx.MustBindFlag(v, "oidc.client.id", flags.Lookup("oidc-client-id"))

	flags.String("oidc-client-secret", "", "oidc client secret")
	viperx.MustBindFlag(v, "oidc.client.secret", flags.Lookup("oidc-client-secret"))

	flags.String("oidc-client-issuer", "", "oidc issuer")
	viperx.MustBindFlag(v, "oidc.client.issuer", flags.Lookup("oidc-client-issuer"))
}

func fetchIssuerTokenEndpoint(ctx context.Context, issuer string) (string, error) {
	uri, err := url.JoinPath(issuer, ".well-known", "openid-configuration")
	if err != nil {
		return "", fmt.Errorf("invalid issuer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return "", err
	}

	res, err := tokenEndpointClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close() //nolint:errcheck // no need to check

	var m map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", err
	}

	tokenEndpoint, ok := m["token_endpoint"]
	if !ok {
		return "", ErrTokenEndpointMissing
	}

	return tokenEndpoint.(string), nil
}
