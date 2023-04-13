package echojwtx

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

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
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

// testHelperJWKSProvider returns a url for a webserver that will return JSONWebKeySets
func testHelperJWKSProvider(keyIDs ...string) (string, func()) {
	e := echo.New()

	keySet := testHelperJoseJWKSProvider(keyIDs...)

	e.GET("/.well-known/jwks.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, keySet)
	})

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	s := &http.Server{
		Handler: e,
	}

	go func() {
		if err := s.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	jwksURI := fmt.Sprintf("http://localhost:%d/.well-known/jwks.json", listener.Addr().(*net.TCPAddr).Port)

	close := func() {
		s.Close() //nolint:errcheck // error check not needed
	}

	return jwksURI, close
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

// TestOAuthClient creates a new http client handling OAuth automatically.
// Returned is the new HTTP Client, JWKS URI and a close function.
func TestOAuthClient(subject string, audience string, issuer string) (*http.Client, string, func()) {
	jwksuri, close := testHelperJWKSProvider(TestPrivRSAKey1ID, TestPrivRSAKey2ID)

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
	})), jwksuri, close
}
