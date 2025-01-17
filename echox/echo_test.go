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
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"go.infratographer.com/x/versionx"
)

const (
	maxRetries      uint = 10
	backoffDuration      = 5 * time.Millisecond
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

	srv, err := NewServer(zap.NewNop(), config, versionx.BuildDetails())

	require.NoError(t, err, "no error expected for new server")

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

	_, err := backoff.Retry(context.Background(),
		func() (struct{}, error) {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testURL, nil)
			if err != nil {
				return struct{}{}, err
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return struct{}{}, err
			}

			defer resp.Body.Close() //nolint:errcheck // no need to check error in test

			if resp.StatusCode != http.StatusOK {
				return struct{}{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode) //nolint:err113 // this is fine for a test
			}

			return struct{}{}, nil
		},
		backoff.WithBackOff(backoff.NewConstantBackOff(backoffDuration)),
		backoff.WithMaxTries(maxRetries),
	)

	require.NoError(t, err, "error waiting for server to be ready")
}

func TestNewServer(t *testing.T) {
	parseNet := func(s string) *net.IPNet {
		_, ipnet, _ := net.ParseCIDR(s)

		return ipnet
	}

	testCases := []struct {
		name         string
		config       Config
		expectServer *Server
		expectError  string
	}{
		{
			"empty config",
			Config{},
			&Server{
				logger:          zap.NewNop().Named("echox"),
				listen:          ":8080",
				readinessChecks: map[string]CheckFunc{},
				shutdownTimeout: DefaultServerShutdownTimeout,
			},
			"",
		},
		{
			"with valid trusted proxies",
			Config{
				TrustedProxies: []string{
					"1.2.3.4",
					"2.3.4.5/32",
					"3.4.5.6/24",
					"2001:db8:abcd:0012::0",
					"2001:db8:abcd:0012::1/128",
					"2001:db8:abcd:0013::0/112",
				},
			},
			&Server{
				logger:          zap.NewNop().Named("echox"),
				listen:          ":8080",
				readinessChecks: map[string]CheckFunc{},
				shutdownTimeout: DefaultServerShutdownTimeout,
				trustedProxies: []*net.IPNet{
					parseNet("1.2.3.4/32"),
					parseNet("2.3.4.5/32"),
					parseNet("3.4.5.6/24"),
					parseNet("2001:db8:abcd:0012::0/128"),
					parseNet("2001:db8:abcd:0012::1/128"),
					parseNet("2001:db8:abcd:0013::0/112"),
				},
			},
			"",
		},
		{
			"with invalid ipv4 trusted proxies",
			Config{
				TrustedProxies: []string{
					"1.2.bad.4",
				},
			},
			nil,
			ErrInvalidTrustedProxyIP.Error(),
		},
		{
			"with invalid ipv4 net trusted proxies",
			Config{
				TrustedProxies: []string{
					"1.2.3.4/bad",
				},
			},
			nil,
			"invalid CIDR address: 1.2.3.4/bad",
		},
		{
			"with invalid ipv6 trusted proxies",
			Config{
				TrustedProxies: []string{
					"2001:db8:-:0012::0",
				},
			},
			nil,
			ErrInvalidTrustedProxyIP.Error(),
		},
		{
			"with invalid ipv6 net trusted proxies",
			Config{
				TrustedProxies: []string{
					"2001:db8:abcd:0012::0/bad",
				},
			},
			nil,
			"invalid CIDR address: 2001:db8:abcd:0012::0/bad",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := NewServer(zap.NewNop(), tc.config, nil)

			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError, "unexpected error returned")

				return
			}

			require.NoError(t, err, "unexpected error returned from NewServer")

			assert.Equal(t, tc.expectServer, srv, "server result doesn't match expectation")
		})
	}
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
			obsZapCore, obsLogs := observer.New(zap.InfoLevel)

			srv := &Server{
				logger: zap.New(obsZapCore),
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

			assert.Contains(t, routes, "/baseline", "expected baseline to exist in route list")
			assert.Contains(t, routes, "/"+strings.TrimLeft(tc.path, "/"), "expected /%s to exist in route list", tc.path)

			w := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/test", nil)

			require.NoError(t, err, "no error expected creating test request")

			engine.ServeHTTP(w, req)

			assert.Equal(t, tc.expectStatus, w.Code, "unexpected response code for %s", tc.path)

			if assert.Len(t, obsLogs.All(), 1, "expected a single request log to be logged") {
				logs := obsLogs.TakeAll()

				assert.Equal(t, "/test", logs[len(logs)-1].ContextMap()["path"], "expected path to be /test")
			}

			w = httptest.NewRecorder()

			req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/baseline", nil)

			require.NoError(t, err, "no error expected creating baseline request")

			engine.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "unexpected response code for baseline")

			if assert.Len(t, obsLogs.All(), 1, "expected a single request log to be logged") {
				logs := obsLogs.TakeAll()

				assert.Equal(t, "/baseline", logs[len(logs)-1].ContextMap()["path"], "expected path to be /baseline")
			}
		})
	}
}

func TestHandlerTrustedProxies(t *testing.T) {
	testCases := []struct {
		name           string
		clientIP       string
		proxyIP        string
		trustedProxies []string
		expectIP       string
	}{
		{
			"none ipv4",
			"1.2.3.10",
			"1.2.3.20",
			nil,
			"1.2.3.20",
		},
		{
			"none ipv6",
			"2001:db8:abcd:12::a7",
			"2001:db8:abcd:12::b3",
			nil,
			"2001:db8:abcd:12::b3",
		},
		{
			"no header ipv4",
			"1.2.3.10",
			"",
			[]string{
				"1.2.3.20/32",
			},
			"1.2.3.10",
		},
		{
			"no header ipv6",
			"2001:db8:abcd:12::a7",
			"2001:db8:abcd:12::b3",
			[]string{
				"2001:db8:abcd:12::b3",
			},
			"2001:db8:abcd:12::a7",
		},
		{
			"trusted ipv4/32",
			"1.2.3.10",
			"1.2.3.20",
			[]string{
				"1.2.3.20/32",
			},
			"1.2.3.10",
		},
		{
			"trusted ipv6/128",
			"2001:db8:abcd:12::a7",
			"2001:db8:abcd:12::b3",
			[]string{
				"2001:db8:abcd:12::b3",
			},
			"2001:db8:abcd:12::a7",
		},
		{
			"trusted ipv4 subnet",
			"1.2.3.10",
			"1.2.3.20",
			[]string{
				"1.2.3.16/28",
			},
			"1.2.3.10",
		},
		{
			"trusted ipv6 subnet",
			"2001:db8:abcd:12::a7",
			"2001:db8:abcd:12::b3",
			[]string{
				"2001:db8:abcd:12::b0/125",
			},
			"2001:db8:abcd:12::a7",
		},
		{
			"untrusted ipv4 subnet",
			"1.2.3.10",
			"1.2.3.40",
			[]string{
				"1.2.3.16/28",
			},
			"1.2.3.40",
		},
		{
			"untrusted ipv6 subnet",
			"2001:db8:abcd:12::a7",
			"2001:db8:abcd:12::c3",
			[]string{
				"2001:db8:abcd:12::b0/125",
			},
			"2001:db8:abcd:12::c3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{
				logger: zap.NewNop(),
				trustedProxies: func() []*net.IPNet {
					nets, _ := parseIPNets(tc.trustedProxies)
					return nets
				}(),
			}

			engine := srv.Handler().(*echo.Echo)

			engine.GET("/ip", func(c echo.Context) error {
				return c.String(http.StatusOK, c.RealIP())
			})

			w := httptest.NewRecorder()
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/ip", nil)

			require.NoError(t, err, "no error expected creating new request")

			if tc.proxyIP != "" {
				ip := tc.proxyIP

				if strings.Contains(tc.name, "ipv6") {
					ip = "[" + ip + "]"
				}

				req.RemoteAddr = ip + ":2345"
			} else {
				ip := tc.clientIP

				if strings.Contains(tc.name, "ipv6") {
					ip = "[" + ip + "]"
				}

				req.RemoteAddr = ip + ":1234"
			}

			req.Header.Add("X-Forwarded-For", tc.clientIP)

			engine.ServeHTTP(w, req)

			assert.Equal(t, tc.expectIP, w.Body.String(), "unexpected real ip")
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
			20 * time.Millisecond,
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
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			defer func() {
				// If a delay is being tested, wait until the test sleep completes
				// to ensure all routines have exited.
				time.Sleep(tc.reqSleep)
			}()

			srv, err := NewServer(
				zap.NewNop(),
				Config{
					ShutdownGracePeriod: tc.gracefulTimeout,
				},
				nil,
			)

			require.NoError(t, err, "unexpected error creating new server")

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
				"test": func(_ context.Context) error {
					return nil
				},
			},
			`{"test":"OK"}`,
			http.StatusOK,
		},
		{
			"single check, fail",
			map[string]CheckFunc{
				"test": func(_ context.Context) error {
					return errored
				},
			},
			`{"test":"errored"}`,
			http.StatusServiceUnavailable,
		},
		{
			"multiple checks, ok",
			map[string]CheckFunc{
				"test1": func(_ context.Context) error {
					return nil
				},
				"test2": func(_ context.Context) error {
					return nil
				},
			},
			`{"test1":"OK","test2":"OK"}`,
			http.StatusOK,
		},
		{
			"multiple checks, first fail",
			map[string]CheckFunc{
				"test1": func(_ context.Context) error {
					return errored
				},
				"test2": func(_ context.Context) error {
					return nil
				},
			},
			`{"test1":"errored","test2":"OK"}`,
			http.StatusServiceUnavailable,
		},
		{
			"multiple checks, last fail",
			map[string]CheckFunc{
				"test1": func(_ context.Context) error {
					return nil
				},
				"test2": func(_ context.Context) error {
					return errored
				},
			},
			`{"test1":"OK","test2":"errored"}`,
			http.StatusServiceUnavailable,
		},
		{
			"multiple checks, all fail",
			map[string]CheckFunc{
				"test1": func(_ context.Context) error {
					return errored
				},
				"test2": func(_ context.Context) error {
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
