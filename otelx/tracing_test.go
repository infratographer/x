package otelx_test

import (
	"testing"

	"go.infratographer.com/x/otelx"

	"go.uber.org/zap"
)

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
