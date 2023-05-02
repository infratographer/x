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

package otelx

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
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

	// ExporterJaeger is for a jaeger exporter capable of connecting to a trace api
	// endpoint as well as using the jaeger agent. You MUST specific either an
	// endpoint or an agent host and port. When connecting with an endpoint you
	// may optionally provide a user and password.
	//
	//  Settings for agent connections:
	//      tracing.jaeger.agent_host    TRACING_JAEGER_AGENT_HOST        host for jaeger agent connection
	//      tracing.jaeger.agent_port    TRACING_JAEGER_AGENT_PORT        port for jaeger agent connection
	//  Settings for endpoint connections:
	//      tracing.jaeger.endpoint      TRACING_JAEGER_AGENT_ENDPOINT    url for jaeger endpoint
	//      tracing.jaeger.user          TRACING_JAEGER_AGENT_USER        user for jaeger endpoint
	//      tracing.jaeger.password      TRACING_JAEGER_AGENT_PASSWORD    password for jaeger endpoint
	ExporterJaeger TraceExporter = "jaeger"

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

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		// Record information about this application in a resource.
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(appName),
			attribute.String("environment", tc.Environment),
		)),
	}

	// exporter could be nil if we are in "passthrough" mode, but we still want
	// to set everything up to traces go work going through the application
	if exp != nil {
		opts = append(opts, sdktrace.WithBatcher(exp))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return nil
}

func newExporter(tc Config) (sdktrace.SpanExporter, error) {
	switch tc.Provider {
	case ExporterStdout:
		return newStdoutExporter(tc)
	case ExporterJaeger:
		return newJaegerExporter(tc)
	case ExporterOTLPHTTP:
		return newOTLPGRPCExporter(tc)
	case ExporterOTLPGRPC:
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

func newJaegerExporter(tc Config) (sdktrace.SpanExporter, error) {
	switch {
	case tc.Jaeger.AgentHost != "" && tc.Jaeger.AgentPort != "":
		return jaeger.New(jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(tc.Jaeger.AgentHost),
			jaeger.WithAgentPort(tc.Jaeger.AgentPort),
		))
	case tc.Jaeger.Endpoint != "":
		opts := []jaeger.CollectorEndpointOption{
			jaeger.WithEndpoint(tc.Jaeger.Endpoint),
		}

		if tc.Jaeger.User != "" {
			opts = append(opts, jaeger.WithUsername(tc.Jaeger.User))
		}

		if tc.Jaeger.Password != "" {
			opts = append(opts, jaeger.WithPassword(tc.Jaeger.Password))
		}

		return jaeger.New(jaeger.WithCollectorEndpoint(opts...))
	default:
		return nil, &ConfigError{Message: "missing jaeger config options; you must pass an endpoint or agent host and port"}
	}
}

func newOTLPGRPCExporter(tc Config) (sdktrace.SpanExporter, error) {
	_, err := url.Parse(tc.OTLP.Endpoint)
	if err != nil {
		return nil, &ConfigError{Message: "missing OTLP config options; you must pass a valid endpoint", Err: err}
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(tc.OTLP.Endpoint),
		otlptracehttp.WithTimeout(tc.OTLP.Timeout),
	}

	if tc.OTLP.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	return otlptrace.New(context.Background(), otlptracehttp.NewClient(opts...))
}

func newOTLPHTTPExporter(tc Config) (sdktrace.SpanExporter, error) {
	_, err := url.Parse(tc.OTLP.Endpoint)
	if err != nil {
		return nil, &ConfigError{Message: "missing OTLP config options; you must pass a valid endpoint", Err: err}
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(tc.OTLP.Endpoint),
		otlptracegrpc.WithTimeout(tc.OTLP.Timeout),
	}

	if tc.OTLP.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	return otlptrace.New(context.Background(), otlptracegrpc.NewClient(opts...))
}
