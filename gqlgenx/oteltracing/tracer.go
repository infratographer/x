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

package oteltracing

import (
	"context"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type (
	// Tracer provides an otel tracer for gqlgen servers
	Tracer struct {
		// FieldSpans allow you to enable a span for each field in the response
		FieldSpans bool
	}
)

var _ interface {
	graphql.HandlerExtension
	graphql.FieldInterceptor
} = Tracer{}

var tracer = otel.Tracer("go.infratographer.com/x/gqlgenx/oteltracing")

// ExtensionName returns the name of this extension
func (Tracer) ExtensionName() string {
	return "OpenTelemetryTracing"
}

// Validate is required to meet HandlerExtension interface
func (Tracer) Validate(graphql.ExecutableSchema) error {
	return nil
}

// InterceptField adds the middleware that lets us add traces to each field of a request
func (t Tracer) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	fc := graphql.GetFieldContext(ctx)

	// check if this is a method or a resolver, if it's not and we aren't emitting
	// field spans skip tracing
	if !fc.IsMethod && !fc.IsResolver && !t.FieldSpans {
		return next(ctx)
	}

	attrs := []attribute.KeyValue{
		attribute.String("fieldName", fc.Field.Name),
		attribute.String("returnType", fc.Field.Definition.Type.String()),
	}

	if fc.IsMethod {
		for k, v := range fc.Args {
			if strings.HasSuffix(k, "ID") || k == "id" {
				attrs = append(attrs, attribute.String("args."+k, fmt.Sprintf("%s", v)))
			}
		}
	}

	ctx, span := tracer.Start(ctx, fc.Path().String(), trace.WithAttributes(attrs...))
	defer span.End()

	return next(ctx)
}
