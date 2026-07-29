package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExchangeJWTForPAT(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		assert.NoError(t, r.ParseForm())
		assert.Equal(t, "urn:ietf:params:oauth:grant-type:token-exchange", r.FormValue("grant_type"))
		assert.Equal(t, "the-jwt", r.FormValue("subject_token"))
		assert.Equal(t, "urn:ietf:params:oauth:token-type:access_token", r.FormValue("subject_token_type"))
		_, _ = w.Write([]byte(`{"access_token":"bitpat_x","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	c := Config{OIDCTokenEndpoint: srv.URL}
	pat, expiry, err := c.exchangeJWTForPAT(context.Background(), "the-jwt")
	require.NoError(t, err)
	assert.Equal(t, "bitpat_x", pat)
	assert.WithinDuration(t, time.Now().Add(time.Hour), expiry, time.Minute, "expiry should come from expires_in")
}

func TestExchangeJWTForPAT_Errors(t *testing.T) {
	t.Run("non-200", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
		}))
		defer srv.Close()

		_, _, err := (Config{OIDCTokenEndpoint: srv.URL}).exchangeJWTForPAT(context.Background(), "j")
		assert.Error(t, err)
	})

	t.Run("missing access_token", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"token_type":"bearer"}`))
		}))
		defer srv.Close()

		_, _, err := (Config{OIDCTokenEndpoint: srv.URL}).exchangeJWTForPAT(context.Background(), "j")
		assert.ErrorContains(t, err, "missing access_token")
	})
}

func TestParseJWTExp(t *testing.T) {
	exp := time.Now().Add(42 * time.Minute).Unix()
	got, ok := parseJWTExp(makeJWT(exp))
	require.True(t, ok, "expected to parse the exp claim")
	assert.Equal(t, exp, got.Unix())

	_, ok = parseJWTExp("only.two")
	assert.False(t, ok, "a two-part token should not parse")

	_, ok = parseJWTExp("aaa.@@@notbase64.ccc")
	assert.False(t, ok, "an undecodable payload should not parse")
}

func TestJWTExpiry_Precedence(t *testing.T) {
	now := time.Now()

	// expires_in takes precedence over the embedded exp claim.
	got := jwtExpiry(tokenResponse{ExpiresIn: 600, AccessToken: makeJWT(now.Add(time.Hour).Unix())}, now)
	assert.WithinDuration(t, now.Add(10*time.Minute), got, time.Minute)

	// no expires_in → fall back to the JWT's exp claim.
	expUnix := now.Add(20 * time.Minute).Unix()
	got = jwtExpiry(tokenResponse{AccessToken: makeJWT(expUnix)}, now)
	assert.Equal(t, expUnix, got.Unix())

	// neither → short conservative fallback.
	got = jwtExpiry(tokenResponse{AccessToken: "garbage"}, now)
	assert.WithinDuration(t, now.Add(5*time.Minute), got, time.Minute)
}

// makeJWT builds a minimal unsigned JWT carrying the given exp claim. Shared
// across the oauth package tests.
func makeJWT(expUnix int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload, _ := json.Marshal(map[string]any{"sub": "user", "exp": expUnix})
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
}
