package echojwtx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/MicahParks/keyfunc"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

const (
	// ActorKey defines the context key an actor is stored in.
	ActorKey = "actor"
)

var (
	// ErrJWKSURIMissing is returned when the jwks_uri field is not found in the issuer's oidc well-known configuration.
	ErrJWKSURIMissing = errors.New("jwks_uri missing from oidc provider")
)

func noopMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}

// AuthConfig provides configuration for JWT validation using JWKS.
type AuthConfig struct {
	// Logger defines the auth logger to use.
	Logger *zap.Logger

	// Issuer is the Auth Issuer
	Issuer string

	// Audience is the Auth Audience
	Audience string

	// JWTConfig configuration for handling JWT validation.
	JWTConfig echojwt.Config

	// KeyFuncOptions configuration for fetching JWKS.
	KeyFuncOptions keyfunc.Options
}

// Auth handles JWT Authentication as echo middleware.
type Auth struct {
	logger *zap.Logger

	jwtConfig echojwt.Config

	middleware echo.MiddlewareFunc

	issuer   string
	audience string
}

func (a *Auth) setup(ctx context.Context, config AuthConfig) error {
	if config.Logger != nil {
		a.logger = config.Logger
	} else {
		a.logger = zap.NewNop()
	}

	a.issuer = config.Issuer
	a.audience = config.Audience

	jwksURI, err := jwksURI(ctx, a.issuer)
	if err != nil {
		return err
	}

	jwks, err := keyfunc.Get(jwksURI, config.KeyFuncOptions)
	if err != nil {
		return err
	}

	a.jwtConfig = config.JWTConfig

	jwtConfig := &config.JWTConfig
	jwtConfig.KeyFunc = jwks.Keyfunc

	mdw, err := jwtConfig.ToMiddleware()
	if err != nil {
		return err
	}

	// intercepts the next function to run final validation.
	a.middleware = func(next echo.HandlerFunc) echo.HandlerFunc {
		skipper := jwtConfig.Skipper
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
		return noopMiddleware
	}

	return a.middleware
}

// NewAuth creates a new auth middleware handler for JWTs using JWKS.
func NewAuth(ctx context.Context, config AuthConfig) (*Auth, error) {
	auth := new(Auth)

	if err := auth.setup(ctx, config); err != nil {
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
