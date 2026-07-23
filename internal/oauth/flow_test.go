package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
)

// oauthMock is a test double for the WorkOS token endpoint (/oauth2/token) and
// the monolith OIDC exchange (/oidc/token), with call counters.
type oauthMock struct {
	server *httptest.Server

	mu            sync.Mutex
	tokenCalls    int // /oauth2/token (authorization_code + refresh)
	exchangeCalls int // /oidc/token (JWT → PAT)

	jwt          string
	refreshToken string
	jwtExpiresIn int64
	pat          string
	patExpiresIn int64
	failRefresh  bool
}

func newOAuthMock() *oauthMock {
	m := &oauthMock{
		jwt:          makeJWT(time.Now().Add(time.Hour).Unix()),
		refreshToken: "refresh-1",
		jwtExpiresIn: 3600,
		pat:          "bitpat_minted",
		patExpiresIn: 3600,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		m.tokenCalls++
		fail := m.failRefresh
		m.mu.Unlock()
		_ = r.ParseForm()
		if r.FormValue("grant_type") == "refresh_token" && fail {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  m.jwt,
			"refresh_token": m.refreshToken,
			"expires_in":    m.jwtExpiresIn,
			"token_type":    "Bearer",
		})
	})
	mux.HandleFunc("/oidc/token", func(w http.ResponseWriter, _ *http.Request) {
		m.mu.Lock()
		m.exchangeCalls++
		m.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": m.pat,
			"token_type":   "bearer",
			"expires_in":   m.patExpiresIn,
		})
	})
	m.server = httptest.NewServer(mux)
	return m
}

func (m *oauthMock) config() Config {
	return Config{
		Issuer:            m.server.URL,
		OIDCTokenEndpoint: m.server.URL + "/oidc/token",
		ClientID:          "https://cli.example/cimd.json",
		Resource:          "https://cli.example",
	}
}

func (m *oauthMock) close() { m.server.Close() }

func (m *oauthMock) counts() (tokenCalls, exchangeCalls int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.tokenCalls, m.exchangeCalls
}

func TestEnsureFreshPAT_ManualTokenPassthrough(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := auth.Save(auth.Auth{Token: "manual-pat"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m := newOAuthMock()
	defer m.close()

	got, err := m.config().EnsureFreshPAT(context.Background(), "manual-pat")
	if err != nil {
		t.Fatalf("EnsureFreshPAT: %v", err)
	}
	if got != "manual-pat" {
		t.Fatalf("got %q, want manual-pat", got)
	}
	if tc, ec := m.counts(); tc != 0 || ec != 0 {
		t.Fatalf("manual token should make no HTTP calls; got token=%d exchange=%d", tc, ec)
	}
}

func TestEnsureFreshPAT_ValidPAT(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := auth.Save(auth.Auth{
		Token: "still-good", TokenExpiry: time.Now().Add(time.Hour),
		JWT: "j", JWTExpiry: time.Now().Add(time.Hour),
		RefreshToken: "r",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m := newOAuthMock()
	defer m.close()

	got, err := m.config().EnsureFreshPAT(context.Background(), "still-good")
	if err != nil {
		t.Fatalf("EnsureFreshPAT: %v", err)
	}
	if got != "still-good" {
		t.Fatalf("got %q, want still-good", got)
	}
	if tc, ec := m.counts(); tc != 0 || ec != 0 {
		t.Fatalf("valid PAT should make no HTTP calls; got token=%d exchange=%d", tc, ec)
	}
}

func TestEnsureFreshPAT_ExpiredPAT_ValidJWT(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := auth.Save(auth.Auth{
		Token: "old-pat", TokenExpiry: time.Now().Add(-time.Minute),
		JWT: "good-jwt", JWTExpiry: time.Now().Add(time.Hour),
		RefreshToken: "r",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m := newOAuthMock()
	defer m.close()

	got, err := m.config().EnsureFreshPAT(context.Background(), "old-pat")
	if err != nil {
		t.Fatalf("EnsureFreshPAT: %v", err)
	}
	if got != "bitpat_minted" {
		t.Fatalf("got %q, want bitpat_minted", got)
	}
	if tc, ec := m.counts(); tc != 0 || ec != 1 {
		t.Fatalf("expected 0 token + 1 exchange; got token=%d exchange=%d", tc, ec)
	}
	if saved, _ := auth.Load(); saved.Token != "bitpat_minted" {
		t.Fatalf("new PAT not persisted: %q", saved.Token)
	}
}

func TestEnsureFreshPAT_ExpiredPATAndJWT_RefreshesAndRotates(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := auth.Save(auth.Auth{
		Token: "old", TokenExpiry: time.Now().Add(-time.Hour),
		JWT: "old-jwt", JWTExpiry: time.Now().Add(-time.Minute),
		RefreshToken: "refresh-old",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m := newOAuthMock()
	defer m.close()
	m.refreshToken = "refresh-rotated" // WorkOS rotates the refresh token

	got, err := m.config().EnsureFreshPAT(context.Background(), "old")
	if err != nil {
		t.Fatalf("EnsureFreshPAT: %v", err)
	}
	if got != "bitpat_minted" {
		t.Fatalf("got %q, want bitpat_minted", got)
	}
	if tc, ec := m.counts(); tc != 1 || ec != 1 {
		t.Fatalf("expected 1 refresh + 1 exchange; got token=%d exchange=%d", tc, ec)
	}
	saved, _ := auth.Load()
	if saved.Token != "bitpat_minted" {
		t.Fatalf("PAT not persisted: %q", saved.Token)
	}
	if saved.RefreshToken != "refresh-rotated" {
		t.Fatalf("rotated refresh token not persisted: %q", saved.RefreshToken)
	}
}

func TestEnsureFreshPAT_RefreshRejected(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := auth.Save(auth.Auth{
		Token: "old", TokenExpiry: time.Now().Add(-time.Hour),
		JWT: "old-jwt", JWTExpiry: time.Now().Add(-time.Hour),
		RefreshToken: "expired-refresh",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m := newOAuthMock()
	defer m.close()
	m.failRefresh = true

	_, err := m.config().EnsureFreshPAT(context.Background(), "old")
	if !errors.Is(err, ErrLoginRequired) {
		t.Fatalf("expected ErrLoginRequired, got %v", err)
	}
}

// callbackOpener returns a fake browser that completes the loopback callback
// with the given code and (optionally overridden) state.
func callbackOpener(t *testing.T, code, overrideState string) func(string) error {
	t.Helper()
	return func(rawURL string) error {
		u, err := url.Parse(rawURL)
		if err != nil {
			return err
		}
		q := u.Query()
		state := q.Get("state")
		if overrideState != "" {
			state = overrideState
		}
		cb := q.Get("redirect_uri") + "?code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(state)
		resp, err := http.Get(cb)
		if err != nil {
			return err
		}
		return resp.Body.Close()
	}
}

func TestLogin_HappyPath(t *testing.T) {
	m := newOAuthMock()
	defer m.close()

	a, err := m.config().Login(context.Background(), callbackOpener(t, "auth-code", ""), io.Discard)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if a.Token != "bitpat_minted" {
		t.Fatalf("token = %q, want bitpat_minted", a.Token)
	}
	if a.RefreshToken != "refresh-1" {
		t.Fatalf("refresh token = %q, want refresh-1", a.RefreshToken)
	}
	if !a.IsOAuthManaged() {
		t.Fatal("login result should be OAuth-managed")
	}
	if a.TokenExpiry.IsZero() || a.JWTExpiry.IsZero() {
		t.Fatal("expiries should be set after login")
	}
	if tc, ec := m.counts(); tc != 1 || ec != 1 {
		t.Fatalf("expected 1 code-exchange + 1 PAT-exchange; got token=%d exchange=%d", tc, ec)
	}
}

func TestLogin_StateMismatch(t *testing.T) {
	m := newOAuthMock()
	defer m.close()

	_, err := m.config().Login(context.Background(), callbackOpener(t, "auth-code", "WRONG-STATE"), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("expected state-mismatch error, got %v", err)
	}
}

func TestLogin_GuardsMissingConfig(t *testing.T) {
	if _, err := (Config{ClientID: "x"}).Login(context.Background(), nil, io.Discard); err == nil || !strings.Contains(err.Error(), "issuer") {
		t.Fatalf("expected missing-issuer error, got %v", err)
	}
	if _, err := (Config{Issuer: "https://x"}).Login(context.Background(), nil, io.Discard); err == nil || !strings.Contains(err.Error(), "client_id") {
		t.Fatalf("expected missing-client_id error, got %v", err)
	}
}
