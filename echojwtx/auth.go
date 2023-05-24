package echojwtx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type actorContext struct{}

const (
	// ActorKey defines the context key an actor is stored in for an echo context
	ActorKey = "actor"
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
}

// Auth handles JWT Authentication as echo middleware.
type Auth struct {
	logger *zap.Logger

	middleware echo.MiddlewareFunc

	// JWTConfig configuration for handling JWT validation.
	JWTConfig echojwt.Config

	// KeyFuncOptions configuration for fetching JWKS.
	KeyFuncOptions keyfunc.Options

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

// WithKeyFuncOptions sets the KeyFuncOptions for the auth middleware.
func WithKeyFuncOptions(keyFuncOptions keyfunc.Options) Opts {
	return func(a *Auth) {
		a.KeyFuncOptions = keyFuncOptions
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
		a.KeyFuncOptions.RefreshTimeout = config.RefreshTimeout
	}

	a.issuer = config.Issuer
	a.audience = config.Audience

	if a.JWTConfig.KeyFunc == nil {
		jwksURI, err := jwksURI(ctx, a.issuer)
		if err != nil {
			return err
		}

		jwks, err := keyfunc.Get(jwksURI, a.KeyFuncOptions)
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

func jwksURI(ctx context.Context, issuer string) (string, error) {
	uri, err := url.JoinPath(issuer, ".well-known", "openid-configuration")
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close() //nolint:errcheck // no need to check

	var m map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", err
	}

	jwksURL, ok := m["jwks_uri"]
	if !ok {
		return "", ErrJWKSURIMissing
	}

	return jwksURL.(string), nil
}
