package echozap

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap/zapcore"
)

// MiddlewareOption sets the middleware function definition.
type MiddlewareOption func(c *MiddlewareConfig)

// WithSkipper sets the middleware Skipper config option.
func WithSkipper(skipper middleware.Skipper) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Skipper = skipper
	}
}

// WithCustomTimeFormat sets the middleware CustomTimeFormat config option.
func WithCustomTimeFormat(format string) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.CustomTimeFormat = format
	}
}

// WithExtraFields sets the middleware ExtraFields config option.
func WithExtraFields(fields []zapcore.Field) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.ExtraFields = append(c.ExtraFields, fields...)
	}
}

// WithExtraFieldsHook sets the middleware ExtraFieldsHook config option.
func WithExtraFieldsHook(hook func(echo.Context) []zapcore.Field) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.ExtraFieldsHook = hook
	}
}
