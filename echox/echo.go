// Copyright 2022 The Infratographer Authors
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

package echox

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brpaz/echozap"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.infratographer.com/x/versionx"
)

// LogFunc is a function that can be used to add additional fields to the log
// output.
type LogFunc func(c echo.Context) []zapcore.Field

// CheckFunc is a function that can be used to check the status of a service.
type CheckFunc func(ctx context.Context) error

var (
	emptyLogFn = func(c echo.Context) []zapcore.Field { return []zapcore.Field{} }
)

// Server implements the HTTP Server
type Server struct {
	listen          string
	handlers        []handler
	logger          *zap.Logger
	version         *versionx.Details
	readinessChecks map[string]CheckFunc
	shutdownTimeout time.Duration
}

// NewServer will return an opinionated echo server for processing API requests.
func NewServer(logger *zap.Logger, cfg Config, version *versionx.Details) *Server {
	shutdownTimeout := cfg.ShutdownGracePeriod
	if shutdownTimeout == 0 {
		shutdownTimeout = DefaultServerShutdownTimeout
	}

	return &Server{
		listen:          cfg.Listen,
		logger:          logger.Named("echox"),
		version:         version,
		readinessChecks: map[string]CheckFunc{},
		shutdownTimeout: shutdownTimeout,
	}
}

// DefaultEngine returns a base echo instance for processing requests.
// This setups logging, requestid, and otel middleware.
func DefaultEngine(logger *zap.Logger, f LogFunc) *echo.Echo {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	engine := echo.New()

	defaultSkipper := func(c echo.Context) bool {
		switch c.Request().URL.Path {
		case "/version", "/livez", "/readyz", "/metrics":
			return true
		default:
			return false
		}
	}

	engine.Use(middleware.RequestID())
	engine.Use(echozap.ZapLogger(logger))
	engine.Use(middleware.Recover())
	engine.Use(otelecho.Middleware(hostname, otelecho.WithSkipper(defaultSkipper)))

	engine.HideBanner = true
	engine.HidePort = true

	return engine
}

type handler interface {
	Routes(*echo.Group)
}

// AddHandler provides the ability to add additional HTTP handlers that process
// requests. The handler that is provided should have a Routes(*echo.Group)
// function, which allows the routes to be added to the server.
func (s *Server) AddHandler(h handler) *Server {
	s.handlers = append(s.handlers, h)
	return s
}

// AddReadinessCheck will accept a function to be ran during calls to /readyx.
// These functions should accept a context and only return an error. When adding
// a readiness check a name is also provided, this name will be used when returning
// the state of all the checks
func (s *Server) AddReadinessCheck(name string, f CheckFunc) *Server {
	s.readinessChecks[name] = f

	return s
}

func (s *Server) engine() *echo.Echo {
	// Setup default echo router
	r := DefaultEngine(s.logger, emptyLogFn)

	p := prometheus.NewPrometheus("echo", nil)

	p.Use(r)

	if s.version != nil {
		// Version endpoint returns build information
		r.GET("/version", s.versionHandler)
	}

	// Health endpoints
	r.GET("/livez", s.livenessCheckHandler)
	r.GET("/readyz", s.readinessCheckHandler)

	for _, handler := range s.handlers {
		handler.Routes(r.Group("/"))
	}

	return r
}

// Serve serves an http server on the provided listener.
// Serve blocks until SIGINT or SIGTERM are signalled,
// or if the http serve fails.
// A graceful shutdown will be attempted
func (s *Server) Serve(listener net.Listener) error {
	logger := s.logger.With(zap.String("address", listener.Addr().String()))

	logger.Info("starting server")

	srv := &http.Server{
		Handler: s.engine(),
	}

	var (
		exit = make(chan error, 1)
		quit = make(chan os.Signal, 2) //nolint:gomnd
	)

	// Serve in a go routine.
	// If serve returns an error, capture the error to return later.
	go func() {
		if err := srv.Serve(listener); err != nil {
			exit <- err

			return
		}

		exit <- nil
	}()

	// close server to kill active connections.
	defer srv.Close() //nolint:errcheck // server is being closed, we'll ignore this.

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var err error

	select {
	case err = <-exit:
		return err
	case <-quit:
		logger.Warn("server shutting down")
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown timed out", zap.Error(err))

		return err
	}

	return nil
}

// Run listens and serves the echo server on the specified address.
// See Serve for more details.
func (s *Server) Run() error {
	listener, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}

	defer listener.Close() //nolint:errcheck // No need to check error.

	return s.Serve(listener)
}
