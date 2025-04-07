package otelx_test

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"

	"go.infratographer.com/x/otelx"
)

func newSpanID() trace.SpanID {
	var id trace.SpanID

	_, _ = rand.Read(id[:])

	return id
}

func testTraceIDs(count int) []trace.TraceID {
	prefix := []byte{0, 0, 0, 0, 0, 0, 0, 0}

	chunkSize := uint64((1 << 63) / uint64(count))

	ids := make([]trace.TraceID, count)

	for i := range ids {
		var chunk [8]byte

		binary.BigEndian.PutUint64(chunk[:], (chunkSize*uint64(i+1)-1)<<1)

		var traceID trace.TraceID

		copy(traceID[:8], prefix)
		copy(traceID[8:16], chunk[:])

		ids[i] = traceID
	}

	return ids
}

func TestSmoke(t *testing.T) {
	logger := &zap.SugaredLogger{}
	cfg := otelx.Config{
		Enabled:  true,
		Provider: otelx.ExporterStdout,
	}

	err := otelx.InitTracer(cfg, "test", logger)
	if err != nil {
		t.Fatalf("could not initialize otelx: %s", err)
	}
}

func testTraceContext(ctx context.Context, traceID trace.TraceID, parent, parentSampling bool) context.Context {
	spanConfig := trace.SpanContextConfig{
		TraceID: traceID,
	}

	// If parent then set a parent span id and mark whether the parent was sampled.
	// If not parent then simply set the SpanContext with only the TraceID configured.
	if parent {
		spanConfig.SpanID = newSpanID()
		spanConfig.TraceFlags = trace.FlagsSampled.WithSampled(parentSampling)

		ctx = trace.ContextWithRemoteSpanContext(ctx, trace.NewSpanContext(spanConfig))
	} else {
		ctx = trace.ContextWithSpanContext(ctx, trace.NewSpanContext(spanConfig))
	}

	return ctx
}

func TestTraceSampling(t *testing.T) {
	testCases := []struct {
		name           string
		ratio          float64
		withParent     bool
		parentSampled  bool
		traces         int
		expectedTraces int
	}{
		{
			"always sampled",
			1.0,
			false,
			false,
			4,
			4,
		},
		{
			"half sampled",
			0.5,
			false,
			false,
			4,
			2,
		},
		{
			"parent trace sampling",
			0.5,
			true,
			true,
			4,
			4,
		},
		{
			"parent trace not sampling",
			0.5,
			true,
			false,
			4,
			0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rootCtx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg := otelx.Config{
				SampleRatio: tc.ratio,
			}

			exporter := tracetest.NewInMemoryExporter()

			defer exporter.Shutdown(rootCtx) //nolint:errcheck

			tp, err := otelx.NewTracerProviderWithExporter(exporter, t.Name(), cfg)

			require.NoError(t, err, "no error expected while initializing a new tracer provider")

			defer tp.Shutdown(rootCtx) //nolint:errcheck

			tracer := tp.Tracer(t.Name())

			for i, traceID := range testTraceIDs(tc.traces) {
				ctx := testTraceContext(rootCtx, traceID, tc.withParent, tc.parentSampled)

				_, span := tracer.Start(ctx, fmt.Sprintf("trace-%d", i))

				span.End()
			}

			require.NoError(t, tp.ForceFlush(rootCtx), "no error expected flushing traces")

			assert.Equal(t, tc.expectedTraces, len(exporter.GetSpans()), "unexpected number of spans traced")
		})
	}
}
