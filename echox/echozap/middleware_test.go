package echozap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestToMiddleware(t *testing.T) {
	returnStatus := func(status int) echo.HandlerFunc {
		return func(c echo.Context) error {
			return c.String(status, strconv.Itoa(status))
		}
	}
	returnError := func(iErr any) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err, ok := iErr.(error); ok {
				return err
			}

			return fmt.Errorf("%v", iErr) //nolint:goerr113 // fine for tests
		}
	}

	testCases := []struct {
		name              string
		config            MiddlewareConfig
		requestMethod     string
		requestPath       string
		setRequestID      bool
		returnFn          echo.HandlerFunc
		expectFields      map[string]interface{}
		expectConfigError error
	}{
		{
			"empty config",
			MiddlewareConfig{},
			"",
			"",
			false,
			nil,
			nil,
			ErrLoggerRequired,
		},
		{
			"success log",
			MiddlewareConfig{
				Logger: zap.NewNop(),
			},
			http.MethodGet,
			"/test",
			false,
			returnStatus(http.StatusOK),
			map[string]interface{}{
				"status":         int64(http.StatusOK),
				"method":         http.MethodGet,
				"path":           "/test",
				"query":          "",
				"response_bytes": int64(3),
			},
			nil,
		},
		{
			"with request id (from request)",
			MiddlewareConfig{
				Logger: zap.NewNop(),
			},
			http.MethodGet,
			"/test",
			true,
			returnStatus(http.StatusOK),
			map[string]interface{}{
				"status":         int64(http.StatusOK),
				"method":         http.MethodGet,
				"path":           "/test",
				"query":          "",
				"response_bytes": int64(3),
				"request_id":     "test-request-id",
			},
			nil,
		},
		{
			"simple error log",
			MiddlewareConfig{
				Logger: zap.NewNop(),
			},
			http.MethodGet,
			"/test",
			false,
			returnError(errors.New("test error")), //nolint:goerr113 // fine for tests
			map[string]interface{}{
				"status":         int64(http.StatusInternalServerError),
				"method":         http.MethodGet,
				"path":           "/test",
				"query":          "",
				"response_bytes": int64(36),
				"error":          "test error",
			},
			nil,
		},
		{
			"log with formatted time",
			MiddlewareConfig{
				Logger:           zap.NewNop(),
				CustomTimeFormat: time.RFC3339,
			},
			http.MethodGet,
			"/test",
			false,
			returnError(errors.New("test error")), //nolint:goerr113 // fine for tests
			map[string]interface{}{
				"status":         int64(http.StatusInternalServerError),
				"method":         http.MethodGet,
				"path":           "/test",
				"query":          "",
				"response_bytes": int64(36),
				"error":          "test error",
			},
			nil,
		},
		{
			"simple http error",
			MiddlewareConfig{
				Logger: zap.NewNop(),
			},
			http.MethodGet,
			"/test",
			false,
			returnError(echo.NewHTTPError(http.StatusForbidden, "http error")),
			map[string]interface{}{
				"status":         int64(http.StatusForbidden),
				"method":         http.MethodGet,
				"path":           "/test",
				"query":          "",
				"response_bytes": int64(25),
				"error":          "code=403, message=http error",
				"error_code":     int64(http.StatusForbidden),
				"error_message":  "http error",
			},
			nil,
		},
		{
			"extended http error",
			MiddlewareConfig{
				Logger: zap.NewNop(),
			},
			http.MethodGet,
			"/test",
			false,
			returnError(
				echo.NewHTTPError(
					http.StatusForbidden,
					"http error",
				).WithInternal(errors.New("internal error")), //nolint:goerr113 // fine for tests
			),
			map[string]interface{}{
				"status":         int64(http.StatusForbidden),
				"method":         http.MethodGet,
				"path":           "/test",
				"query":          "",
				"response_bytes": int64(25),
				"error":          "code=403, message=http error, internal=internal error",
				"error_code":     int64(http.StatusForbidden),
				"error_message":  "http error",
				"error_internal": "internal error",
			},
			nil,
		},
		{
			"extra static fields",
			MiddlewareConfig{
				Logger: zap.NewNop(),
				ExtraFields: []zapcore.Field{
					zap.String("static_field_1", "static_value_1"),
					zap.String("static_field_2", "static_value_2"),
				},
			},
			http.MethodGet,
			"/test",
			false,
			returnStatus(http.StatusOK),
			map[string]interface{}{
				"status":         int64(http.StatusOK),
				"method":         http.MethodGet,
				"path":           "/test",
				"query":          "",
				"response_bytes": int64(3),
				"static_field_1": "static_value_1",
				"static_field_2": "static_value_2",
			},
			nil,
		},
		{
			"extra dynamic fields",
			MiddlewareConfig{
				Logger: zap.NewNop(),
				ExtraFieldsHook: func(c echo.Context) []zapcore.Field {
					return []zapcore.Field{
						zap.String("status_text", http.StatusText(c.Response().Status)),
					}
				},
			},
			http.MethodGet,
			"/test",
			false,
			returnStatus(http.StatusOK),
			map[string]interface{}{
				"status":         int64(http.StatusOK),
				"method":         http.MethodGet,
				"path":           "/test",
				"query":          "",
				"response_bytes": int64(3),
				"status_text":    "OK",
			},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			obsZapCore, obsLogs := observer.New(zap.InfoLevel)
			if tc.config.Logger != nil {
				tc.config.Logger = zap.New(obsZapCore)
			}
			mdw, err := tc.config.ToMiddleware()

			if tc.expectConfigError != nil {
				require.Error(t, err, "expected error to be returned")

				assert.ErrorIs(t, err, tc.expectConfigError, "error mismatched")

				return
			}

			require.NoError(t, err, "unexpected error returned")

			e := echo.New()
			e.Use(middleware.RequestID())
			e.Use(mdw)
			e.GET(tc.requestPath, func(c echo.Context) error {
				time.Sleep(20 * time.Millisecond)
				return tc.returnFn(c)
			})

			req, err := http.NewRequestWithContext(
				context.Background(),
				tc.requestMethod,
				tc.requestPath,
				strings.NewReader("body content"),
			)

			require.NoError(t, err, "expected no error creating request")

			req.RemoteAddr = "127.0.0.5:1234"
			req.Host = "test-host.local"

			req.Header.Add("User-Agent", "test/agent")

			if tc.setRequestID {
				req.Header.Add(echo.HeaderXRequestID, "test-request-id")
			}

			resp := httptest.NewRecorder()

			e.ServeHTTP(resp, req)

			assert.Equal(t, 1, obsLogs.Len(), "expected 1 log to have been logged")

			if obsLogs.Len() == 0 {
				return
			}

			var expectedKeys []string

			lastLog := obsLogs.All()[obsLogs.Len()-1]
			lastLogFields := lastLog.ContextMap()

			// add static default expectations
			tc.expectFields["host"] = "test-host.local"
			tc.expectFields["remote_addr"] = "127.0.0.5"
			tc.expectFields["path"] = tc.requestPath
			tc.expectFields["user_agent"] = "test/agent"
			tc.expectFields["content_length"] = int64(12)

			for ek, ev := range tc.expectFields {
				gv, ok := lastLogFields[ek]

				assert.Truef(t, ok, "expected field in log: %s", ek)

				assert.Equalf(t, ev, gv, "unexpected field value for %s", ek)
			}

			if _, ok := lastLogFields["time"].(int64); ok {
				assert.NotZero(t, lastLogFields["time"], "unexpected time")
			} else {
				assert.NotEmpty(t, lastLogFields["time"])
			}

			if !tc.setRequestID {
				assert.NotEmpty(t, lastLogFields["request_id"], "expected request_id to be defined")

				expectedKeys = append(expectedKeys, "request_id")
			}

			assert.GreaterOrEqual(t, lastLogFields["duration"], float64(20), "expected duration to be greater than 20 milliseconds")

			latency, err := time.ParseDuration(lastLogFields["latency"].(string))

			require.NoError(t, err, "unexpected error parsing duration: %s", lastLogFields["latency"])

			assert.GreaterOrEqual(t, latency, 20*time.Millisecond, "expected latency to be greater than 20 milliseconds")

			// add manually checked fields
			expectedKeys = append(expectedKeys, "duration", "latency", "time")

			var gotKeys []string

			for k := range tc.expectFields {
				expectedKeys = append(expectedKeys, k)
			}

			for k := range lastLogFields {
				gotKeys = append(gotKeys, k)
			}

			assert.ElementsMatch(t, expectedKeys, gotKeys, "expected keys mismatch")
		})
	}
}
