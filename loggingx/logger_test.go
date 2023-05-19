package loggingx_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.infratographer.com/x/loggingx"
	"go.uber.org/zap"
)

func TestZero(t *testing.T) {
	testCase := []struct {
		name      string
		appName   string
		cfg       loggingx.Config
		out       *bytes.Buffer
		wantMsg   string
		wantLevel string
	}{
		{
			name:      "happy path",
			appName:   "test",
			cfg:       loggingx.Config{},
			out:       &bytes.Buffer{},
			wantMsg:   `"app":"test"`,
			wantLevel: `"level":"info"`,
		},
		{
			name:      "happy path with debug",
			appName:   "test",
			cfg:       loggingx.Config{Debug: true},
			out:       &bytes.Buffer{},
			wantMsg:   `"app":"test"`,
			wantLevel: `"level":"debug"`,
		},
		{
			name:      "happy path with pretty",
			appName:   "test",
			cfg:       loggingx.Config{Pretty: true},
			out:       &bytes.Buffer{},
			wantMsg:   `app=test`,
			wantLevel: `INF`,
		},
		{
			name:      "happy path with pretty and debug",
			appName:   "test",
			cfg:       loggingx.Config{Pretty: true, Debug: true},
			out:       &bytes.Buffer{},
			wantMsg:   `app=test`,
			wantLevel: `DBG`,
		},
		{
			name:      "Time is rf3339",
			appName:   "test",
			cfg:       loggingx.Config{Debug: true},
			out:       &bytes.Buffer{},
			wantMsg:   fmt.Sprintf(`"time":"%d-%02d-%02d`, time.Now().Year(), time.Now().Month(), time.Now().Day()),
			wantLevel: `"level":"debug"`,
		},
		{
			name:    "Time is unix",
			appName: "test",
			cfg:     loggingx.Config{},
			out:     &bytes.Buffer{},
			wantMsg: `"time":1`,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			l := loggingx.InitLoggerWithWriter(tc.appName, tc.cfg, tc.out)

			if tc.cfg.Debug {
				l.Debug().Msg("bar")
			} else {
				l.Info().Msg("foo")
			}

			assert.Contains(t, tc.out.String(), tc.wantMsg)

			assert.Contains(t, tc.out.String(), tc.wantLevel)
		})
	}

	tests := []struct {
		name string
		want interface{}
		got  interface{}
	}{
		{
			name: "ZeroLogger",
			want: &zerolog.Logger{},
			got: loggingx.InitZeroLogger("test", loggingx.Config{
				Debug: true,
				// Pretty: true,
			}),
		},
		{
			name: "ZapLogger",
			want: &zap.SugaredLogger{},
			got: loggingx.InitLogger("test", loggingx.Config{
				Debug: true,
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.IsType(t, tt.want, tt.got)
			switch tt.got.(type) {
			case *zerolog.Logger:
				tt.got.(*zerolog.Logger).Info().Msg("foo")
			}
		})
	}
}
