package echojwtx_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

var (
	testKeySize = 2048

	// TestPrivRSAKey1 provides an RSA key used to sign tokens
	TestPrivRSAKey1, _ = rsa.GenerateKey(rand.Reader, testKeySize)
	// TestPrivRSAKey1ID is the ID of this signing key in tokens
	TestPrivRSAKey1ID = "testKey1"
	// TestPrivRSAKey2 provides an RSA key used to sign tokens
	TestPrivRSAKey2, _ = rsa.GenerateKey(rand.Reader, testKeySize)
	// TestPrivRSAKey2ID is the ID of this signing key in tokens
	TestPrivRSAKey2ID = "testKey2"

	keyMap sync.Map
)

func init() {
	keyMap.Store(TestPrivRSAKey1ID, TestPrivRSAKey1)
	keyMap.Store(TestPrivRSAKey2ID, TestPrivRSAKey2)
}

// testHelperMustMakeSigner will return a JWT signer from the given key
func testHelperMustMakeSigner(alg jose.SignatureAlgorithm, kid string, k interface{}) jose.Signer {
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: alg, Key: k}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", kid))
	if err != nil {
		panic("failed to create signer:" + err.Error())
	}

	return sig
}

// testHelperJoseJWKSProvider returns a JWKS
func testHelperJoseJWKSProvider(keyIDs ...string) jose.JSONWebKeySet {
	jwks := make([]jose.JSONWebKey, len(keyIDs))

	for idx, keyID := range keyIDs {
		rawKey, found := keyMap.Load(keyID)
		if !found {
			panic("Failed finding private key to create test JWKS provider. Fix the test.")
		}

		privKey := rawKey.(*rsa.PrivateKey)

		jwks[idx] = jose.JSONWebKey{
			KeyID: keyID,
			Key:   &privKey.PublicKey,
		}
	}

	return jose.JSONWebKeySet{
		Keys: jwks,
	}
}

// testHelperOIDCProvider returns an issuer and a close function.
func testHelperOIDCProvider(keyIDs ...string) (string, func()) {
	e := echo.New()

	var lc net.ListenConfig

	listener, err := lc.Listen(context.Background(), "tcp", ":0")
	if err != nil {
		panic(err)
	}

	s := &http.Server{
		Handler: e,
	}

	issuer := fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)

	keySet := testHelperJoseJWKSProvider(keyIDs...)

	e.GET("/.well-known/openid-configuration", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{
			"jwks_uri": issuer + "/.well-known/jwks.json",
		})
	})

	e.GET("/.well-known/jwks.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, keySet)
	})

	go func() {
		if err := s.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	closer := func() {
		s.Close() //nolint:errcheck // error check not needed
	}

	return issuer, closer
}

// testHelperGetToken will return a signed token
func testHelperGetToken(signer jose.Signer, cl jwt.Claims, key string, value interface{}) string {
	sc := map[string]interface{}{}

	sc[key] = value

	raw, err := jwt.Signed(signer).Claims(cl).Claims(sc).CompactSerialize()
	if err != nil {
		panic(err)
	}

	return raw
}

// OAuthTestClient creates a new http client handling OAuth automatically.
// Returned is the new HTTP Client, OIDC URI and a close function.
func OAuthTestClient(subject string, audience string) (*http.Client, string, func()) {
	issuer, closer := testHelperOIDCProvider(TestPrivRSAKey1ID, TestPrivRSAKey2ID)

	ctx := context.Background()

	var audiences jwt.Audience

	if audience != "" {
		audiences = append(audiences, audience)
	}

	authClaim := jwt.Claims{
		Issuer:    issuer,
		Subject:   subject,
		Audience:  audiences,
		NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
	}

	signer := testHelperMustMakeSigner(jose.RS256, TestPrivRSAKey1ID, TestPrivRSAKey1)
	rawToken := testHelperGetToken(signer, authClaim, "scope", "test")

	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: rawToken,
	})), issuer, closer
}
