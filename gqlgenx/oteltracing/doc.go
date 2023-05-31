// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

// Package oteltracing provides a gqlgen middleware that adds otel tracing.
//
// You can use it like:
//
//	srv := handler.New(es)
//	srv.Use(oteltracing.Tracer{})
//
// If you would like spans for every field in the response you can enable FieldSpans:
//
//	srv := handler.New(es)
//	srv.Use(oteltracing.Tracer{FieldSpans: true})
package oteltracing
