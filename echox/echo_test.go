package echox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap"

	"go.infratographer.com/x/versionx"
)

const (
	maxRetries      = 10
	backoffDuration = 5 * time.Millisecond
)

type testRoute struct {
	Method  string
	Path    string
	Handler echo.HandlerFunc
}

type testHandler struct {
	routes []testRoute
}

func (h testHandler) Routes(group *echo.Group) {
	for _, route := range h.routes {
		group.Add(route.Method, route.Path, route.Handler)
	}
}

func testServer(t *testing.T, config Config, preRun func(srv *Server)) (*Server, string, func()) {
	t.Helper()

	listen := config.Listen
	if listen == "" {
		listen = "127.0.0.1:0"
	}

	listener, err := net.Listen("tcp", listen)

	require.NoError(t, err, "no error expected listening")

	srv := NewServer(zap.NewNop(), config, versionx.BuildDetails())

	if preRun != nil {
		preRun(srv)
	}

	go srv.Serve(listener) //nolint:errcheck // no need to check error in test

	url := "http://" + listener.Addr().String()

	waitForServer(t, url+"/livez")

	return srv, url, func() {
		_ = listener.Close()
	}
}

func waitForServer(t *testing.T, testURL string) {
	t.Helper()

	err := backoff.Retry(
		func() error {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testURL, nil)
			if err != nil {
				return err
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}

			defer resp.Body.Close() //nolint:errcheck // no need to check error in test

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code: %d", resp.StatusCode) //nolint:goerr113 // this is fine for a test
			}

			return nil
		},
		backoff.WithMaxRetries(backoff.NewConstantBackOff(backoffDuration), maxRetries),
	)

	require.NoError(t, err, "error waiting for server to be ready")
}

func TestHandler(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectStatus int
	}{
		{
			"without /",
			"test",
			http.StatusOK,
		},
		{
			"with /",
			"/test",
			http.StatusOK,
		},
		{
			"not found",
			"other",
			http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{
				logger: zap.NewNop(),
			}

			srv.AddHandler(testHandler{
				routes: []testRoute{
					{
						Method: http.MethodGet,
						Path:   "baseline",
						Handler: func(c echo.Context) error {
							return c.String(http.StatusOK, "ok")
						},
					},
				},
			})

			srv.AddHandler(testHandler{
				routes: []testRoute{
					{
						Method: http.MethodGet,
						Path:   tc.path,
						Handler: func(c echo.Context) error {
							return c.String(http.StatusOK, "ok")
						},
					},
				},
			})

			engine := srv.Handler().(*echo.Echo)

			var routes []string

			for _, route := range engine.Routes() {
				routes = append(routes, route.Path)
			}

			assert.Contains(t, routes, "baseline", "expected baseline to exist in route list")
			assert.Contains(t, routes, tc.path, "expected %s to exist in route list", tc.path)

			w := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/test", nil)

			require.NoError(t, err, "no error expected creating test request")

			engine.ServeHTTP(w, req)

			assert.Equal(t, tc.expectStatus, w.Code, "unexpected response code for %s", tc.path)

			w = httptest.NewRecorder()

			req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/baseline", nil)

			require.NoError(t, err, "no error expected creating baseline request")

			engine.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "unexpected response code for baseline")
		})
	}
}

func TestServe(t *testing.T) {
	testCases := []struct {
		name            string
		reqSleep        time.Duration
		gracefulTimeout time.Duration
		ctxCancelDelay  time.Duration
		expectReqErr    string
		expectSrvErr    string
	}{
		{
			"requests complete",
			0,
			10 * time.Millisecond,
			0,
			"",
			"",
		},
		{
			"signal, request finishes",
			5 * time.Millisecond,
			10 * time.Millisecond,
			0,
			"",
			"",
		},
		{
			"signal, request fails",
			2 * time.Second,
			10 * time.Millisecond,
			0,
			"EOF",
			"deadline exceeded",
		},
		{
			"context, request finishes",
			10 * time.Millisecond,
			20 * time.Millisecond,
			1 * time.Millisecond,
			"",
			"",
		},
		{
			"context, request fails",
			2 * time.Second,
			10 * time.Millisecond,
			10 * time.Millisecond,
			"EOF",
			"deadline exceeded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			defer func() {
				// If a delay is being tested, wait until the test sleep completes
				// to ensure all routines have exited.
				time.Sleep(tc.reqSleep)
			}()

			srv := NewServer(
				zap.NewNop(),
				Config{
					ShutdownGracePeriod: tc.gracefulTimeout,
				},
				nil,
			)

			handler := testHandler{
				routes: []testRoute{
					{
						Method: http.MethodGet,
						Path:   "test",
						Handler: func(c echo.Context) error {
							time.Sleep(tc.reqSleep)

							return c.String(http.StatusOK, fmt.Sprintf("slept for %s", tc.reqSleep.String()))
						},
					},
				},
			}

			srv.AddHandler(handler)

			listener, err := net.Listen("tcp", "127.0.0.1:0")

			require.NoError(t, err, "net.Listen should not return an error")

			defer listener.Close() //nolint:errcheck // not needed in test

			url := fmt.Sprintf("http://%s/test", listener.Addr().String())

			var (
				wg sync.WaitGroup

				srvErr  error
				reqResp *http.Response
				reqErr  error

				ctx    = context.Background()
				cancel func()
			)

			if tc.ctxCancelDelay != 0 {
				ctx, cancel = context.WithCancel(ctx)

				defer cancel()
			}

			// Wait group for server and client
			wg.Add(2)

			// Start server
			go func() {
				defer wg.Done()

				srvErr = srv.ServeWithContext(ctx, listener)
			}()

			waitForServer(t, url)

			establishedCh := make(chan bool, 1)

			// Create a client which times out quick
			// Once the connection is established, we write to a channel so we can kill the server
			client := &http.Client{
				Transport: &http.Transport{
					DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
						conn, err := (&net.Dialer{
							Timeout: 5 * time.Second,
						}).DialContext(ctx, network, addr)
						if err == nil && conn != nil {
							establishedCh <- true
						}

						return conn, err
					},
				},
			}

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)

			require.NoError(t, err)

			// Start client request
			go func() {
				defer wg.Done()

				resp, err := client.Do(req)
				if err != nil {
					reqErr = err

					return
				}
				reqResp = resp

				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close() //nolint:errcheck // not needed in test
			}()

			// Wait for client to establish a connection before triggering the kill
			select {
			case established := <-establishedCh:
				require.True(t, established, "expected connection to be established")
			case <-time.After(time.Second):
				t.Error("failed to establish connection to server")
				t.FailNow()
			}

			// if ctxCancelDelay is defined, we're using the context to kill the service
			if cancel == nil {
				// Ask for SIGTERM (ensures we don't exit when we trigger killing ourself)
				c := make(chan os.Signal, 1)
				signal.Notify(c, syscall.SIGTERM)
				defer signal.Stop(c)

				// Send SIGTERM to trigger shutdown process
				err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

				require.NoError(t, err, "syscall.Kill should not return an error")
			} else {
				time.Sleep(tc.ctxCancelDelay)
				cancel()
			}

			// Wait for server and request to complete
			wg.Wait()

			if tc.expectSrvErr != "" {
				require.Error(t, srvErr, "expected server to have errored")

				assert.Contains(t, srvErr.Error(), tc.expectSrvErr, "unexpected server error")
			} else {
				require.NoError(t, srvErr, "expected server to not have errored")
			}

			if tc.expectReqErr != "" {
				require.Error(t, reqErr, "expected request to have errored")

				assert.Contains(t, reqErr.Error(), tc.expectReqErr, "unexpected request error")
			} else {
				require.NoError(t, reqErr, "expected request to not have errored")
			}

			if tc.expectSrvErr == "" && tc.expectReqErr == "" {
				assert.Equal(t, http.StatusOK, reqResp.StatusCode, "unexpected status code")
			} else {
				assert.Nil(t, reqResp, "expected response to be nil")
			}
		})
	}
}

var (
	errored = errors.New("errored")
)

func TestAddReadinessCheck(t *testing.T) {
	testCases := []struct {
		name         string
		checks       map[string]CheckFunc
		expectBody   string
		expectStatus int
	}{
		{
			"no checks",
			nil,
			`{}`,
			http.StatusOK,
		},
		{
			"single check, ok",
			map[string]CheckFunc{
				"test": func(ctx context.Context) error {
					return nil
				},
			},
			`{"test":"OK"}`,
			http.StatusOK,
		},
		{
			"single check, fail",
			map[string]CheckFunc{
				"test": func(ctx context.Context) error {
					return errored
				},
			},
			`{"test":"errored"}`,
			http.StatusServiceUnavailable,
		},
		{
			"multiple checks, ok",
			map[string]CheckFunc{
				"test1": func(ctx context.Context) error {
					return nil
				},
				"test2": func(ctx context.Context) error {
					return nil
				},
			},
			`{"test1":"OK","test2":"OK"}`,
			http.StatusOK,
		},
		{
			"multiple checks, first fail",
			map[string]CheckFunc{
				"test1": func(ctx context.Context) error {
					return errored
				},
				"test2": func(ctx context.Context) error {
					return nil
				},
			},
			`{"test1":"errored","test2":"OK"}`,
			http.StatusServiceUnavailable,
		},
		{
			"multiple checks, last fail",
			map[string]CheckFunc{
				"test1": func(ctx context.Context) error {
					return nil
				},
				"test2": func(ctx context.Context) error {
					return errored
				},
			},
			`{"test1":"OK","test2":"errored"}`,
			http.StatusServiceUnavailable,
		},
		{
			"multiple checks, all fail",
			map[string]CheckFunc{
				"test1": func(ctx context.Context) error {
					return errored
				},
				"test2": func(ctx context.Context) error {
					return errored
				},
			},
			`{"test1":"errored","test2":"errored"}`,
			http.StatusServiceUnavailable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, url, closeFn := testServer(t, Config{}, func(srv *Server) {
				for check, checkFn := range tc.checks {
					srv.AddReadinessCheck(check, checkFn)
				}
			})

			defer closeFn()

			readinessURL := url + "/readyz"

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, readinessURL, nil)

			require.NoError(t, err, "no error expected for new request")

			resp, err := http.DefaultClient.Do(req)

			require.NoError(t, err, "no error expected for client request")
			require.NotNil(t, resp, "response expected")

			defer resp.Body.Close() //nolint:errcheck // no need to check error in test

			body, err := io.ReadAll(resp.Body)

			require.NoError(t, err, "no error expected reading response body")

			assert.Equal(t, tc.expectStatus, resp.StatusCode, "unexpected status code")
			assert.Equal(t, tc.expectBody+"\n", string(body), "unexpected body response")
		})
	}
}
