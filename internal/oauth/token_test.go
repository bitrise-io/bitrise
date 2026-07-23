package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// makeJWT builds a minimal unsigned JWT carrying the given exp claim. Shared
// across the oauth package tests.
func makeJWT(expUnix int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload, _ := json.Marshal(map[string]any{"sub": "user", "exp": expUnix})
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
}

func TestExchangeJWTForPAT(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q", ct)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		if g := r.FormValue("grant_type"); g != "urn:ietf:params:oauth:grant-type:token-exchange" {
			t.Errorf("grant_type = %q", g)
		}
		if r.FormValue("subject_token") != "the-jwt" {
			t.Errorf("subject_token = %q", r.FormValue("subject_token"))
		}
		if r.FormValue("subject_token_type") != "urn:ietf:params:oauth:token-type:access_token" {
			t.Errorf("subject_token_type = %q", r.FormValue("subject_token_type"))
		}
		_, _ = w.Write([]byte(`{"access_token":"bitpat_x","token_type":"bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	c := Config{OIDCTokenEndpoint: srv.URL}
	pat, expiry, err := c.exchangeJWTForPAT(context.Background(), "the-jwt")
	if err != nil {
		t.Fatalf("exchangeJWTForPAT: %v", err)
	}
	if pat != "bitpat_x" {
		t.Fatalf("pat = %q, want bitpat_x", pat)
	}
	if d := time.Until(expiry); d < 59*time.Minute || d > 61*time.Minute {
		t.Fatalf("expiry ~1h expected from expires_in, got %v", d)
	}
}

func TestExchangeJWTForPAT_Errors(t *testing.T) {
	t.Run("non-200", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
		}))
		defer srv.Close()
		if _, _, err := (Config{OIDCTokenEndpoint: srv.URL}).exchangeJWTForPAT(context.Background(), "j"); err == nil {
			t.Fatal("expected error on 401")
		}
	})

	t.Run("missing access_token", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"token_type":"bearer"}`))
		}))
		defer srv.Close()
		if _, _, err := (Config{OIDCTokenEndpoint: srv.URL}).exchangeJWTForPAT(context.Background(), "j"); err == nil {
			t.Fatal("expected error on missing access_token")
		}
	})
}

func TestParseJWTExp(t *testing.T) {
	exp := time.Now().Add(42 * time.Minute).Unix()
	got, ok := parseJWTExp(makeJWT(exp))
	if !ok {
		t.Fatal("expected to parse exp claim")
	}
	if got.Unix() != exp {
		t.Fatalf("exp = %d, want %d", got.Unix(), exp)
	}

	if _, ok := parseJWTExp("only.two"); ok {
		t.Fatal("a two-part token should not parse")
	}
	if _, ok := parseJWTExp("aaa.@@@notbase64.ccc"); ok {
		t.Fatal("undecodable payload should not parse")
	}
}

func TestJWTExpiry_Precedence(t *testing.T) {
	now := time.Now()

	// expires_in takes precedence over the embedded exp claim.
	got := jwtExpiry(tokenResponse{ExpiresIn: 600, AccessToken: makeJWT(now.Add(time.Hour).Unix())}, now)
	if d := got.Sub(now); d < 9*time.Minute || d > 11*time.Minute {
		t.Fatalf("expected ~10m from expires_in, got %v", d)
	}

	// no expires_in → fall back to the JWT's exp claim.
	expUnix := now.Add(20 * time.Minute).Unix()
	if got := jwtExpiry(tokenResponse{AccessToken: makeJWT(expUnix)}, now); got.Unix() != expUnix {
		t.Fatalf("expected JWT exp %d, got %d", expUnix, got.Unix())
	}

	// neither → short conservative fallback.
	if got := jwtExpiry(tokenResponse{AccessToken: "garbage"}, now); got.Sub(now) < 4*time.Minute || got.Sub(now) > 6*time.Minute {
		t.Fatalf("expected ~5m fallback, got %v", got.Sub(now))
	}
}
