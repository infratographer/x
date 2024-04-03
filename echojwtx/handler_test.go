package echojwtx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.infratographer.com/x/echojwtx"
)

const (
	chanTimeout = 3 * time.Second
)

func TestNoAuth(t *testing.T) {
	_, issuer, closer := OAuthTestClient("testing-user", "")
	defer closer()

	auth, err := echojwtx.NewAuth(context.Background(), echojwtx.AuthConfig{
		Issuer:         issuer,
		RefreshTimeout: 5 * time.Second,
	})

	require.NoError(t, err, "no error expected for NewAuth")

	gotUserTokenCh := make(chan *jwt.Token, 1)
	gotActorCh := make(chan string, 1)

	e := echo.New()

	e.Use(auth.Middleware())

	e.GET("/test", func(c echo.Context) error {
		token, _ := c.Get("user").(*jwt.Token)
		actor, _ := c.Get(echojwtx.ActorKey).(string)

		gotUserTokenCh <- token
		gotActorCh <- actor

		return nil
	})

	srv := httptest.NewServer(e)

	defer srv.Close()

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, srv.URL+"/test", nil)

	require.NoError(t, err, "expected new request without error")

	resp, err := http.DefaultClient.Do(req)
	_ = resp.Body.Close()

	require.NoError(t, err, "expected response without error")

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 response from test server")
}

func TestAudienceValidation(t *testing.T) {
	testCases := []struct {
		name             string
		clientAudience   string
		serverAudience   string
		expectActor      bool
		expectStatusCode int
	}{
		{
			"no audience or issuer",
			"",
			"",
			true,
			http.StatusOK,
		},
		{
			"skip audience",
			"skipaud",
			"",
			true,
			http.StatusOK,
		},
		{
			"missing audience",
			"",
			"missaud",
			false,
			http.StatusUnauthorized,
		},
		{
			"audience mismatch",
			"testaud",
			"audmismatch",
			false,
			http.StatusUnauthorized,
		},
		{
			"audience match",
			"testaud",
			"testaud",
			true,
			http.StatusOK,
		},
	}

	loggerConfig := zap.NewProductionConfig()

	loggerConfig.Level.SetLevel(zap.DebugLevel)

	logger, err := loggerConfig.Build()

	require.NoError(t, err, "no error expected for logger build")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oauthClient, issuer, closer := OAuthTestClient("urn:test:user", tc.clientAudience)
			defer closer()

			auth, err := echojwtx.NewAuth(context.Background(),
				echojwtx.AuthConfig{
					Audience: tc.serverAudience,
					Issuer:   issuer,
				},
				echojwtx.WithLogger(logger), echojwtx.WithHTTPClientStorageOptions(jwkset.HTTPClientStorageOptions{
					HTTPTimeout: 5 * time.Second,
				}),
			)

			require.NoError(t, err, "no error expected for NewAuth")

			gotUserTokenCh := make(chan *jwt.Token, 1)
			gotActorCh := make(chan string, 1)
			gotActorCtxCh := make(chan string, 1)

			e := echo.New()

			e.Use(auth.Middleware())

			e.GET("/test", func(c echo.Context) error {
				token, _ := c.Get("user").(*jwt.Token)
				actor, _ := c.Get(echojwtx.ActorKey).(string)
				actorCtx, _ := c.Request().Context().Value(echojwtx.ActorCtxKey).(string)

				gotUserTokenCh <- token
				gotActorCh <- actor
				gotActorCtxCh <- actorCtx

				return nil
			})

			srv := httptest.NewServer(e)

			defer srv.Close()

			req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, srv.URL+"/test", nil)

			require.NoError(t, err, "expected new request without error")

			resp, err := oauthClient.Do(req)
			_ = resp.Body.Close()

			require.NoError(t, err, "expected response without error")

			assert.Equalf(t, tc.expectStatusCode, resp.StatusCode, "expected %d response from test server", tc.expectStatusCode)

			if tc.expectStatusCode != http.StatusUnauthorized {
				select {
				case token := <-gotUserTokenCh:
					assert.NotNil(t, token, "expected user token")
				case <-time.After(chanTimeout):
					t.Error("failed to receive user token")
				}

				select {
				case actor := <-gotActorCh:
					if tc.expectActor {
						assert.NotEmpty(t, actor, "expected actor not to be empty")
					} else {
						assert.Empty(t, actor, "expected actor to be empty")
					}
				case <-time.After(chanTimeout):
					t.Error("failed to receive actor result")
				}

				select {
				case actor := <-gotActorCtxCh:
					if tc.expectActor {
						assert.NotEmpty(t, actor, "expected actor not to be empty")
					} else {
						assert.Empty(t, actor, "expected actor to be empty")
					}
				case <-time.After(chanTimeout):
					t.Error("failed to receive actor result")
				}
			}
		})
	}
}
