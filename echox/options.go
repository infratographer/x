package echox

import (
	"github.com/labstack/echo/v4/middleware"

	"go.infratographer.com/x/echox/echozap"
)

// Option sets the server function definition
type Option func(c *Server)

// WithLoggingSkipper sets the echozap middleware Skipper config option
func WithLoggingSkipper(skipper middleware.Skipper) Option {
	return func(s *Server) {
		s.echozapOpts = append(s.echozapOpts, echozap.WithSkipper(skipper))
	}
}
