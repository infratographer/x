// Copyright 2023 The Infratographer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package echojwtx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

var (
	jwksClient = &http.Client{
		Timeout:   5 * time.Second, // nolint:mnd // clear and unexported
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
)

type actorContext struct{}

const (
	// ActorKey defines the context key an actor is stored in for an echo context
	ActorKey = "actor"

	// DefaultHTTPClientStorageOptionRefreshInterval defines the frequency at which the jwks file is refreshed.
	DefaultHTTPClientStorageOptionRefreshInterval = time.Hour

	// DefaultHTTPClientStorageOptionHTTPTimeout limits the runtime of a reload of jwks.
	DefaultHTTPClientStorageOptionHTTPTimeout = 10 * time.Second

	// DefaultRateLimitWaitMax is the default timeout for waiting for rate limiting to end.
	DefaultRateLimitWaitMax = time.Minute
)

var (
	// ActorCtxKey defines the context key an actor is stored in for a plain context
	ActorCtxKey = actorContext{}

	// ErrJWKSURIMissing is returned when the jwks_uri field is not found in the issuer's oidc well-known configuration.
	ErrJWKSURIMissing = errors.New("jwks_uri missing from oidc provider")
)

// Opts defines options for the Auth middleware.
type Opts func(*Auth)

// AuthConfig provides configuration for JWT validation using JWKS.
type AuthConfig struct {
	// Issuer is the Auth Issuer
	Issuer string `mapstructure:"issuer"`

	// Audience is the Auth Audience
	Audience string `mapstructure:"audience"`

	// RefreshTimeout is the timeout for fetching the JWKS from the issuer.
	RefreshTimeout time.Duration `mapstructure:"refresh_timeout"`

	// RateLimitWaitMax is the timeout for waiting for rate limiting to end.
	RateLimitWaitMax time.Duration `mapstructure:"rate_limit_wait_max"`
}

// Auth handles JWT Authentication as echo middleware.
type Auth struct {
	logger *zap.Logger

	middleware echo.MiddlewareFunc

	// JWTConfig configuration for handling JWT validation.
	JWTConfig echojwt.Config

	// HTTPClientStorageOptions configuration for fetching JWKS.
	HTTPClientStorageOptions jwkset.HTTPClientStorageOptions

	issuer   string
	audience string
}

// WithLogger sets the logger for the auth middleware.
func WithLogger(logger *zap.Logger) Opts {
	return func(a *Auth) {
		a.logger = logger
	}
}

// WithJWTConfig sets the JWTConfig for the auth middleware.
func WithJWTConfig(jwtConfig echojwt.Config) Opts {
	return func(a *Auth) {
		a.JWTConfig = jwtConfig
	}
}

// WithHTTPClientStorageOptions sets the HTTPClientStorageOptions for the auth middleware.
func WithHTTPClientStorageOptions(options jwkset.HTTPClientStorageOptions) Opts {
	return func(a *Auth) {
		a.HTTPClientStorageOptions = options
	}
}

func (a *Auth) setup(ctx context.Context, config AuthConfig, options ...Opts) error {
	for _, opt := range options {
		opt(a)
	}

	// Ensure the logger is not nil
	if a.logger == nil {
		a.logger = zap.NewNop()
	}

	if config.RefreshTimeout > 0 {
		a.HTTPClientStorageOptions.HTTPTimeout = config.RefreshTimeout
	}

	if config.RateLimitWaitMax == 0 {
		config.RateLimitWaitMax = DefaultRateLimitWaitMax
	}

	a.issuer = config.Issuer
	a.audience = config.Audience

	if a.JWTConfig.KeyFunc == nil {
		jwksURI, err := jwksURI(ctx, a.issuer)
		if err != nil {
			return err
		}

		if a.HTTPClientStorageOptions.Ctx == nil {
			a.HTTPClientStorageOptions.Ctx = ctx
		}

		if a.HTTPClientStorageOptions.RefreshErrorHandler == nil {
			a.HTTPClientStorageOptions.RefreshErrorHandler = func(_ context.Context, err error) {
				a.logger.Error("error refreshing jwks", zap.Error(err))
			}
		}

		if a.HTTPClientStorageOptions.RefreshInterval == 0 {
			a.HTTPClientStorageOptions.RefreshInterval = DefaultHTTPClientStorageOptionRefreshInterval
		}

		if a.HTTPClientStorageOptions.HTTPTimeout == 0 {
			a.HTTPClientStorageOptions.HTTPTimeout = DefaultHTTPClientStorageOptionHTTPTimeout
		}

		storage, err := jwkset.NewStorageFromHTTP(jwksURI, a.HTTPClientStorageOptions)
		if err != nil {
			return err
		}

		clientOptions := jwkset.HTTPClientOptions{
			Given:            storage,
			RateLimitWaitMax: config.RateLimitWaitMax,
		}

		clientStorage, err := jwkset.NewHTTPClient(clientOptions)
		if err != nil {
			return err
		}

		keyfuncOptions := keyfunc.Options{
			Ctx:     ctx,
			Storage: clientStorage,
		}

		jwks, err := keyfunc.New(keyfuncOptions)
		if err != nil {
			return err
		}

		a.JWTConfig.KeyFunc = jwks.Keyfunc
	}

	mdw, err := a.JWTConfig.ToMiddleware()
	if err != nil {
		return err
	}

	// intercepts the next function to run final validation.
	a.middleware = func(next echo.HandlerFunc) echo.HandlerFunc {
		skipper := a.JWTConfig.Skipper
		if skipper == nil {
			skipper = middleware.DefaultSkipper
		}

		postActions := func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			if err := a.jwtHandler(c); err != nil {
				return err
			}

			return next(c)
		}

		return mdw(postActions)
	}

	return nil
}

// Middleware returns echo middleware for validation jwt tokens.
func (a *Auth) Middleware() echo.MiddlewareFunc {
	if a == nil || a.middleware == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	return a.middleware
}

// NewAuth creates a new auth middleware handler for JWTs using JWKS.
func NewAuth(ctx context.Context, config AuthConfig, options ...Opts) (*Auth, error) {
	auth := new(Auth)

	if err := auth.setup(ctx, config, options...); err != nil {
		return nil, err
	}

	return auth, nil
}

func jwksURI(ctx context.Context, issuer string) (*url.URL, error) {
	uri, err := url.JoinPath(issuer, ".well-known", "openid-configuration")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	res, err := jwksClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck // no need to check

	var m map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return nil, err
	}

	jwksURL, ok := m["jwks_uri"]
	if !ok {
		return nil, ErrJWKSURIMissing
	}

	return url.Parse(jwksURL.(string))
}
