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
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"

	"go.infratographer.com/x/echox/echozap"
	"go.infratographer.com/x/versionx"
)

const (
	ipv4SingleHostCIDR = 32
	ipv4BitLength      = 8 * net.IPv4len

	ipv6SingleHostCIDR = 128
	ipv6BitLength      = 8 * net.IPv6len
)

var (
	// ErrInvalidTrustedProxyIP is returned when an invalid ip is provided as a trusted proxy.
	ErrInvalidTrustedProxyIP = errors.New("invalid trusted proxy ip")
)

// CheckFunc is a function that can be used to check the status of a service.
type CheckFunc func(ctx context.Context) error

// Server implements the HTTP Server
type Server struct {
	debug           bool
	listen          string
	logger          *zap.Logger
	handlers        []handler
	middleware      []echo.MiddlewareFunc
	readinessChecks map[string]CheckFunc
	shutdownTimeout time.Duration
	trustedProxies  []*net.IPNet
	version         *versionx.Details
}

// NewServer will return an opinionated echo server for processing API requests.
func NewServer(logger *zap.Logger, cfg Config, version *versionx.Details) (*Server, error) {
	cfg = cfg.withDefaults()

	trustedProxies, err := parseIPNets(cfg.TrustedProxies)
	if err != nil {
		return nil, err
	}

	return &Server{
		debug:           cfg.Debug,
		listen:          cfg.Listen,
		logger:          logger.Named("echox"),
		middleware:      cfg.Middleware,
		readinessChecks: map[string]CheckFunc{},
		shutdownTimeout: cfg.ShutdownGracePeriod,
		trustedProxies:  trustedProxies,
		version:         version,
	}, nil
}

func parseIPNets(sNets []string) ([]*net.IPNet, error) {
	var nets []*net.IPNet

	for _, entry := range sNets {
		var (
			ipnet *net.IPNet
			err   error
		)

		if strings.Contains(entry, "/") {
			_, ipnet, err = net.ParseCIDR(entry)
			if err != nil {
				return nil, err
			}
		} else {
			ip := net.ParseIP(entry)
			if ip == nil {
				return nil, ErrInvalidTrustedProxyIP
			}

			if ipv4 := ip.To4(); ipv4 != nil {
				ipnet = &net.IPNet{
					IP:   ipv4,
					Mask: net.CIDRMask(ipv4SingleHostCIDR, ipv4BitLength),
				}
			} else {
				ipnet = &net.IPNet{
					IP:   ip,
					Mask: net.CIDRMask(ipv6SingleHostCIDR, ipv6BitLength),
				}
			}
		}

		nets = append(nets, ipnet)
	}

	return nets, nil
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

// Handler returns a new http.Handler for serving requests.
func (s *Server) Handler() http.Handler {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	engine := echo.New()

	engine.Use(middleware.RequestID())
	engine.Use(echozap.Middleware(s.logger))
	engine.Use(middleware.Recover())
	engine.Use(otelecho.Middleware(hostname, otelecho.WithSkipper(SkipDefaultEndpoints)))

	engine.Use(s.middleware...)

	engine.HideBanner = true
	engine.HidePort = true

	engine.Debug = s.debug

	if s.trustedProxies != nil {
		ranges := make([]echo.TrustOption, len(s.trustedProxies))
		for i, trust := range s.trustedProxies {
			ranges[i] = echo.TrustIPRange(trust)
		}

		engine.IPExtractor = echo.ExtractIPFromXFFHeader(ranges...)
	} else {
		engine.IPExtractor = echo.ExtractIPDirect()
	}

	p := prometheus.NewPrometheus("echo", nil)

	p.Use(engine)

	if s.version != nil {
		// Version endpoint returns build information
		engine.GET("/version", s.versionHandler)
	}

	// Health endpoints
	engine.GET("/livez", s.livenessCheckHandler)
	engine.GET("/readyz", s.readinessCheckHandler)

	for _, handler := range s.handlers {
		handler.Routes(engine.Group(""))
	}

	return engine
}

// Serve serves an http server on the provided listener.
// See ServeWithContext for more details.
func (s *Server) Serve(listener net.Listener) error {
	return s.ServeWithContext(context.Background(), listener)
}

// ServeWithContext serves an http server on the provided listener.
// Serve blocks until SIGINT or SIGTERM are signalled,
// or if the http serve fails.
// A graceful shutdown will be attempted
func (s *Server) ServeWithContext(ctx context.Context, listener net.Listener) error {
	logger := s.logger.With(zap.String("address", listener.Addr().String()))

	logger.Info("starting server")

	srv := &http.Server{
		Handler: s.Handler(),
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
	case sig := <-quit:
		logger.Warn(fmt.Sprintf("%s received, server shutting down", sig.String()))
	case <-ctx.Done():
		logger.Warn("context done, server shutting down")

		// Since the context has already been canceled, the server would immediately shutdown.
		// We'll reset the context to allow for the proper grace period to be given.
		ctx = context.Background()
	}

	ctx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown timed out", zap.Error(err))

		return err
	}

	return nil
}

// Run listens and serves the echo server on the configured address.
func (s *Server) Run() error {
	return s.RunWithContext(context.Background())
}

// RunWithContext listens and serves the echo server on the configured address.
// See ServeWithContext for more details.
func (s *Server) RunWithContext(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}

	defer listener.Close() //nolint:errcheck // No need to check error.

	return s.ServeWithContext(ctx, listener)
}

// SkipDefaultEndpoints returns true when the provided context request is for /version /livez /readyz or /metrics
func SkipDefaultEndpoints(c echo.Context) bool {
	switch c.Request().URL.Path {
	case "/version", "/livez", "/readyz", "/metrics":
		return true
	default:
		return false
	}
}
