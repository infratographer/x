// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package ginx

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.infratographer.com/x/versionx"
)

// LogFunc is a function that can be used to add additional fields to the log
// output.
type LogFunc func(c *gin.Context) []zapcore.Field

// CheckFunc is a function that can be used to check the status of a service.
type CheckFunc func(ctx context.Context) error

var (
	emptyLogFn = func(c *gin.Context) []zapcore.Field { return []zapcore.Field{} }

	// DefaultServerShutdownTimeout sets the default for how long we give the sever
	// to shutdown before forcefully stopping the server.
	DefaultServerShutdownTimeout = 5 * time.Second
)

// Server implements the HTTP Server
type Server struct {
	listen          string
	DB              *sqlx.DB
	Debug           bool
	handlers        []handler
	logger          *zap.Logger
	version         *versionx.Details
	readinessChecks map[string]CheckFunc
}

// NewServer will return an opinionated gin server for processing API requests.
func NewServer(lgr *zap.Logger, cfg Config, version *versionx.Details) Server {
	return Server{
		listen:          cfg.Listen,
		logger:          lgr.Named("ginx"),
		version:         version,
		readinessChecks: map[string]CheckFunc{},
	}
}

// DefaultEngine returns a base gin engine for processing requests.
// This setups logging, requestid, and otel middleware.
func DefaultEngine(lgr *zap.Logger, f LogFunc) *gin.Engine {
	tp := otel.GetTracerProvider()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	engine := gin.New()

	logF := func(c *gin.Context) []zapcore.Field {
		fields := []zapcore.Field{
			zap.String("request_id", requestid.Get(c)),
		}
		fields = append(fields, f(c)...)

		return fields
	}

	engine.Use(
		requestid.New(),
		otelgin.Middleware(hostname, otelgin.WithTracerProvider(tp)),
		ginzap.GinzapWithConfig(lgr, &ginzap.Config{
			TimeFormat: time.RFC3339,
			UTC:        true,
			TraceID:    true,
			Context:    logF,
		}),
		ginzap.RecoveryWithZap(lgr, true),
	)

	return engine
}

type handler interface {
	Routes(*gin.RouterGroup)
}

// AddHandler provides the ability to add additional HTTP handlers that process
// requests. The handler that is provided should have a Routes(*gin.RouterGroup)
// function, which allows the routes to be added to the server.
func (s Server) AddHandler(h handler) Server {
	s.handlers = append(s.handlers, h)
	return s
}

// AddReadinessCheck will accept a function to be ran during calls to /readyx.
// These functions should accept a context and only return an error. When adding
// a readiness check a name is also provided, this name will be used when returning
// the state of all the checks
func (s Server) AddReadinessCheck(name string, f CheckFunc) Server {
	s.readinessChecks[name] = f

	return s
}

func (s *Server) engine() *gin.Engine {
	// Setup default gin router
	r := DefaultEngine(s.logger, emptyLogFn)

	p := ginprometheus.NewPrometheus("gin")

	// Remove any params from the URL string to keep the number of labels down
	p.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		return c.FullPath()
	}

	p.Use(r)

	if s.version != nil {
		// Version endpoint returns build information
		r.GET("/version", s.versionHandler)
	}

	// Health endpoints
	r.GET("/livez", s.livenessCheckHandler)
	r.GET("/readyz", s.readinessCheckHandler)

	r.Use(func(c *gin.Context) {
		u := c.GetHeader("User")
		if u != "" {
			c.Set("current_actor", u)
			c.Set("actor_type", "user")
		}
	})

	for _, handler := range s.handlers {
		handler.Routes(r.Group("/"))
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "invalid request - route not found"})
	})

	return r
}

// Run will start the server listening on the specified address and listens for
// os signals to shutdown the server
func (s Server) Run() {
	if !s.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	s.logger.Sugar().Infow("starting server",
		"address", s.listen,
	)

	srv := &http.Server{
		Addr:    s.listen,
		Handler: s.engine(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 2) //nolint:gomnd
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.logger.Info("shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), DefaultServerShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.logger.Fatal("server forced to shutdown", zap.Error(err))
	}
}
