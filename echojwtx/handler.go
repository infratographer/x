package echojwtx

import (
	"context"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

var (
	errInvalidAudience = errors.New("invalid audience")
	errInvalidIssuer   = errors.New("invalid issuer")
)

// jwtHandler validates the token claims and sets the ActorKey to the token subject.
func (a *Auth) jwtHandler(c echo.Context) error {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		a.logger.Warn("jwt user is not jwt.Token")

		return nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		a.logger.Warn("jwt user claims are not jwt.MapClaims type")

		return nil
	}

	if err := a.validateClaims(claims); err != nil {
		a.logger.Error("jwt user claims are not valid", zap.Error(err))

		return err
	}

	if subject, ok := claims["sub"]; ok {
		// store the actor in the request context as well so it's available outside of echo contexts
		req := c.Request()
		req = req.WithContext(context.WithValue(req.Context(), ActorCtxKey, subject))
		c.SetRequest(req)
		c.Set(ActorKey, subject)
	}

	return nil
}

// Actor retrieves the ActorKey from echo Context.
func Actor(c echo.Context) string {
	if actor, ok := c.Get(ActorKey).(string); ok {
		return actor
	}

	return ""
}

func (a *Auth) validateClaims(claims jwt.MapClaims) error {
	if a.audience != "" {
		if audiences, err := claims.GetAudience(); err != nil {
			a.logger.Error("jwt user failed to get audience", zap.Error(err), zap.Any("audience", claims["aud"]))
		} else if !slices.Contains(audiences, a.audience) {
			a.logger.Error("jwt user claim invalid audience", zap.Any("audience", claims["aud"]))

			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired jwt").SetInternal(errInvalidAudience)
		}
	}

	if a.issuer != "" {
		if issuer, err := claims.GetIssuer(); err != nil {
			a.logger.Error("jwt user failed to get issuer", zap.Error(err), zap.Any("issuer", claims["iss"]))
		} else if issuer != a.issuer {
			a.logger.Error("jwt user claim invalid issuer", zap.Any("issuer", claims["iss"]))

			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired jwt").SetInternal(errInvalidIssuer)
		}
	}

	return nil
}
