package echozap

import (
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// ErrLoggerRequired is returned when no logger was provided in the config.
	ErrLoggerRequired = errors.New("logger required")
)

// MiddlewareConfig defines the config for Zap middleware.
type MiddlewareConfig struct {
	// Logger defines the logger to write to.
	Logger *zap.Logger

	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper

	// CustomTimeFormat defines the time format the log should use.
	// Default time format is unix time.
	CustomTimeFormat string

	// ExtraFields defines additional fields to be added to the log.
	ExtraFields []zapcore.Field

	// ExtraFieldsHook allows for additional fields to be added dynamically.
	ExtraFieldsHook func(echo.Context) []zapcore.Field
}

// ToMiddleware converts Config to middleware or returns an error for invalid configuration.
func (config MiddlewareConfig) ToMiddleware() (echo.MiddlewareFunc, error) {
	if config.Logger == nil {
		return nil, ErrLoggerRequired
	}

	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			end := time.Now()

			var fields []zapcore.Field

			if config.ExtraFields != nil {
				fields = append(fields, config.ExtraFields...)
			}

			if config.ExtraFieldsHook != nil {
				fields = append(fields, config.ExtraFieldsHook(c)...)
			}

			logger := config.Logger

			request := c.Request()
			response := c.Response()

			fields = append(fields,
				zap.Int("status", response.Status),
				zap.String("method", request.Method),
				zap.String("path", request.URL.Path),
				zap.String("query", request.URL.RawQuery),
				zap.String("remote_addr", c.RealIP()),
				zap.String("user_agent", request.UserAgent()),
				zap.Int64("content_length", request.ContentLength),
				zap.Float64("duration", float64(end.Sub(start)/time.Millisecond)),
				zap.String("latency", end.Sub(start).String()),
				zap.String("host", request.Host),
				zap.Int64("response_bytes", response.Size),
			)

			reqID := request.Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = response.Header().Get(echo.HeaderXRequestID)
			}

			fields = append(fields, zap.String("request_id", reqID))

			if err != nil {
				fields = append(fields, zap.Error(err))

				if httpErr, ok := err.(*echo.HTTPError); ok {
					fields = append(fields,
						zap.Int("error_code", httpErr.Code),
						zap.Any("error_message", httpErr.Message),
						zap.NamedError("error_internal", httpErr.Internal),
					)
				}
			}

			if config.CustomTimeFormat != "" {
				fields = append(fields, zap.String("time", end.UTC().Format(config.CustomTimeFormat)))
			} else {
				fields = append(fields, zap.Int64("time", end.Unix()))
			}

			if err != nil {
				logger.Error(request.URL.Path, fields...)
			} else {
				logger.Info(request.URL.Path, fields...)
			}

			return err
		}
	}, nil
}

// MiddlewareWithConfig calls config.ToMiddleware returning
// the middleware if no error, if error is returned
// a panic is raised.
func MiddlewareWithConfig(config MiddlewareConfig) echo.MiddlewareFunc {
	mdwfn, err := config.ToMiddleware()
	if err != nil {
		panic(err)
	}

	return mdwfn
}

// Middleware returns a new middleware using the provided logger.
func Middleware(logger *zap.Logger, options ...MiddlewareOption) echo.MiddlewareFunc {
	cfg := &MiddlewareConfig{
		Logger: logger,
	}

	for _, opt := range options {
		opt(cfg)
	}

	return MiddlewareWithConfig(*cfg)
}
