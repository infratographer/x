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

package otelx

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.infratographer.com/x/versionx"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.uber.org/zap"
)

// TraceExporter provides a string representation of the tracing exporters that
// are supported. For example to print to stdout you would use ExporterStdout
// which is the same as using "stdout" as a CLI flag.
type TraceExporter string

const (
	// ExporterStdout is for a stdout exporter.
	//  Settings for stdout:
	//      tracing.stdout.pretty_print          TRACING_STDOUT_PRETTY_PRINT            enable pretty printing the trace output
	//      tracing.stdout.disable_timestamps    TRACING_STDOUT_DISABLE_TIMESTAMPS      disable timestamps in the trace output
	ExporterStdout TraceExporter = "stdout"

	// ExporterOTLPHTTP is for a OTLP exporter capable of connecting over secure or
	// insecure HTTP.
	//
	//      tracing.otlp.endpoint        TRACING_OTLP_ENDPOINT        url for otlp http endpoint
	//      tracing.otlp.insecure        TRACING_OTLP_INSECURE        use an insecure connection to the endpoint
	//      tracing.otlp.timeout         TRACING_OTLP_TIMEOUT         timeout for sending to the endpoint (defaults to 10s)
	ExporterOTLPHTTP TraceExporter = "otlphttp"

	// ExporterOTLPGRPC is for a OTLP exporter capable of connecting over secure or
	// insecure GRPC.
	//
	//      tracing.otlp.endpoint        TRACING_OTLP_ENDPOINT        url for otlp http endpoint
	//      tracing.otlp.insecure        TRACING_OTLP_INSECURE        use an insecure connection to the endpoint
	//      tracing.otlp.timeout         TRACING_OTLP_TIMEOUT         timeout for sending to the endpoint (defaults to 10s)
	ExporterOTLPGRPC TraceExporter = "otlpgrpc"

	// ExporterPassthrough is to configure tracing as a passthrough service. This
	// will setup a tracer and read incoming parent trace info from request and
	// pass parent trace info to downstream services, but will not export any spans
	// or other trace data from the application directly.
	ExporterPassthrough TraceExporter = "passthrough"
)

var (
	// ErrUnknownExporter is returned when the exporter passed in via the config
	// is not a known TraceExporter type.
	ErrUnknownExporter = errors.New("unknown tracing exporter")

	defaultTimeout = time.Second * 10
)

// ConfigError is returned when there is a problem with the provided TracingConfig
// during configuration.
type ConfigError struct {
	Message string
	Err     error
}

func (c *ConfigError) Error() string {
	if c.Err != nil {
		return fmt.Errorf("%s, error: %w", c.Message, c.Err).Error()
	}

	return c.Message
}

// InitTracer will configure the exporter setup in the provided config and create
// an otel TracerProvider. The new TracerProvider will be set as the global trace
// provider.
func InitTracer(tc Config, appName string, _ *zap.SugaredLogger) error {
	if !tc.Enabled {
		return nil
	}

	exp, err := newExporter(tc)
	if err != nil {
		return err
	}

	tp, err := NewTracerProviderWithExporter(exp, appName, tc)
	if err != nil {
		return err
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}

// NewTracerProviderWithExporter creates a new tracer provider using the provided exporter instead of initializing one based on the provided config.
func NewTracerProviderWithExporter(exporter sdktrace.SpanExporter, appName string, tc Config) (*sdktrace.TracerProvider, error) {
	appDetails := versionx.BuildDetails()

	r, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceName(appName),
			semconv.ServiceVersion(appDetails.Version),
			semconv.DeploymentEnvironmentName(tc.Environment),
		),
	)
	if err != nil {
		return nil, &ConfigError{
			Message: "could not construct otel resource",
			Err:     err,
		}
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(
			sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(tc.SampleRatio),
			),
		),
		// Record information about this application in a resource.
		sdktrace.WithResource(r),
	}

	// exporter could be nil if we are in "passthrough" mode, but we still want
	// to set everything up to traces go work going through the application
	if exporter != nil {
		opts = append(opts, sdktrace.WithBatcher(exporter))
	}

	return sdktrace.NewTracerProvider(opts...), nil
}

func newExporter(tc Config) (sdktrace.SpanExporter, error) {
	switch tc.Provider {
	case ExporterStdout:
		return newStdoutExporter(tc)
	case ExporterOTLPGRPC:
		return newOTLPGRPCExporter(tc)
	case ExporterOTLPHTTP:
		return newOTLPHTTPExporter(tc)
	case ExporterPassthrough:
		// in the case of passthrough we don't want to configure an exporter but we still want all the rest of the setup
		return nil, nil
	default:
		return nil, ErrUnknownExporter
	}
}

func newStdoutExporter(tc Config) (sdktrace.SpanExporter, error) {
	opts := []stdouttrace.Option{}

	if tc.Stdout.PrettyPrint {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	}

	if tc.Stdout.DisableTimestamps {
		opts = append(opts, stdouttrace.WithoutTimestamps())
	}

	return stdouttrace.New(opts...)
}

func newOTLPHTTPExporter(tc Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithTimeout(tc.OTLP.Timeout),
	}

	// If an endpoint is provided and it is valid, use the endpoint.
	// If the endpoint is invalid return an error.
	// If no endpoint is provided, fallback to otel sdk env vars.
	if endpoint, err := buildOTLPEndpoint(tc.OTLP); endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpointURL(endpoint))
	} else if err != nil {
		return nil, err
	}

	if tc.OTLP.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	return otlptrace.New(context.Background(), otlptracehttp.NewClient(opts...))
}

func newOTLPGRPCExporter(tc Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithTimeout(tc.OTLP.Timeout),
	}

	// If an endpoint is provided and it is valid, use the endpoint.
	// If the endpoint is invalid return an error.
	// If no endpoint is provided, fallback to otel sdk env vars.
	if endpoint, err := buildOTLPEndpoint(tc.OTLP); endpoint != "" {
		opts = append(opts, otlptracegrpc.WithEndpointURL(endpoint))
	} else if err != nil {
		return nil, err
	}

	if tc.OTLP.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	return otlptrace.New(context.Background(), otlptracegrpc.NewClient(opts...))
}

// buildOTLPEndpoint formats and validates the otlp endpoint if one is provided.
// If no endpoint is provided no error is returned.
func buildOTLPEndpoint(config OTLPConfig) (string, error) {
	if config.Endpoint == "" {
		return "", nil
	}

	// If scheme is not found, add a scheme.
	if !strings.Contains(config.Endpoint, "://") {
		scheme := "http"

		if !config.Insecure {
			scheme = "https"
		}

		config.Endpoint = scheme + "://" + config.Endpoint
	}

	if _, err := url.Parse(config.Endpoint); err != nil {
		return "", fmt.Errorf("invalid OTLP endpoint '%s': %w", config.Endpoint, err)
	}

	return config.Endpoint, nil
}
