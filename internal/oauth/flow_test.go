package oauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
)

func TestEnsureFreshPAT_ManualTokenPassthrough(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "manual-pat"}))
	m := newOAuthMock(t)

	got, err := m.config().EnsureFreshPAT(context.Background(), "manual-pat")
	require.NoError(t, err)
	assert.Equal(t, "manual-pat", got)
	m.assertCounts(t, 0, 0, "a manual token is never refreshed")
}

func TestEnsureFreshPAT_ValidPAT(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{
		Token: "still-good", TokenExpiry: time.Now().Add(time.Hour),
		JWT: "j", JWTExpiry: time.Now().Add(time.Hour),
		RefreshToken: "r",
	}))
	m := newOAuthMock(t)

	got, err := m.config().EnsureFreshPAT(context.Background(), "still-good")
	require.NoError(t, err)
	assert.Equal(t, "still-good", got)
	m.assertCounts(t, 0, 0, "a still-valid PAT needs no network calls")
}

func TestEnsureFreshPAT_ExpiredPAT_ValidJWT(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{
		Token: "old-pat", TokenExpiry: time.Now().Add(-time.Minute),
		JWT: "good-jwt", JWTExpiry: time.Now().Add(time.Hour),
		RefreshToken: "r",
	}))
	m := newOAuthMock(t)

	got, err := m.config().EnsureFreshPAT(context.Background(), "old-pat")
	require.NoError(t, err)
	assert.Equal(t, "bitpat_minted", got)
	m.assertCounts(t, 0, 1, "a valid JWT refreshes the PAT with one exchange, no refresh-token grant")

	saved, err := auth.Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_minted", saved.Token, "the new PAT should be persisted")
}

func TestEnsureFreshPAT_JWTLooksValidButExchangeRejected_FallsBackToFullRefresh(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{
		Token: "old-pat", TokenExpiry: time.Now().Add(-time.Minute),
		JWT: "revoked-jwt", JWTExpiry: time.Now().Add(time.Hour), // unexpired by the clock, but the server rejects it anyway
		RefreshToken: "r",
	}))
	m := newOAuthMock(t)
	m.failExchangeTimes = 1 // first exchange (using the stale-looking JWT) is rejected

	got, err := m.config().EnsureFreshPAT(context.Background(), "old-pat")
	require.NoError(t, err)
	assert.Equal(t, "bitpat_minted", got)
	m.assertCounts(t, 1, 2, "a rejected exchange should fall back to a full refresh, then exchange again")

	saved, err := auth.Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_minted", saved.Token, "the new PAT should be persisted")
}

func TestEnsureFreshPAT_ExpiredPATAndJWT_RefreshesAndRotates(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{
		Token: "old", TokenExpiry: time.Now().Add(-time.Hour),
		JWT: "old-jwt", JWTExpiry: time.Now().Add(-time.Minute),
		RefreshToken: "refresh-old",
	}))
	m := newOAuthMock(t)
	m.refreshToken = "refresh-rotated" // WorkOS rotates the refresh token

	got, err := m.config().EnsureFreshPAT(context.Background(), "old")
	require.NoError(t, err)
	assert.Equal(t, "bitpat_minted", got)
	m.assertCounts(t, 1, 1, "a stale PAT and JWT need one refresh plus one exchange")

	saved, err := auth.Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_minted", saved.Token, "the new PAT should be persisted")
	assert.Equal(t, "refresh-rotated", saved.RefreshToken, "the rotated refresh token should be persisted")
}

// TestEnsureFreshPAT_ConcurrentCallsRefreshOnlyOnce simulates two overlapping
// CLI invocations (e.g. two commands started at once) both finding a stale
// PAT+JWT. Only the first to grab the lock should actually spend the refresh
// token; the second, after acquiring the lock, should see the already-fresh
// PAT written by the first and skip its own refresh entirely.
func TestEnsureFreshPAT_ConcurrentCallsRefreshOnlyOnce(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{
		Token: "old", TokenExpiry: time.Now().Add(-time.Hour),
		JWT: "old-jwt", JWTExpiry: time.Now().Add(-time.Minute),
		RefreshToken: "refresh-old",
	}))
	m := newOAuthMock(t)

	cfg := m.config()
	var wg sync.WaitGroup
	results := make([]string, 2)
	errs := make([]error, 2)
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = cfg.EnsureFreshPAT(context.Background(), "old")
		}(i)
	}
	wg.Wait()

	for i := range 2 {
		require.NoError(t, errs[i], "call %d", i)
		assert.Equal(t, "bitpat_minted", results[i], "call %d", i)
	}
	m.assertCounts(t, 1, 1, "only one of the two concurrent calls should spend the refresh token")
}

func TestEnsureFreshPAT_RefreshRejected(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{
		Token: "old", TokenExpiry: time.Now().Add(-time.Hour),
		JWT: "old-jwt", JWTExpiry: time.Now().Add(-time.Hour),
		RefreshToken: "expired-refresh",
	}))
	m := newOAuthMock(t)
	m.failRefresh = true

	_, err := m.config().EnsureFreshPAT(context.Background(), "old")
	assert.ErrorIs(t, err, ErrLoginRequired)
}

func TestLogin_HappyPath(t *testing.T) {
	m := newOAuthMock(t)

	a, err := m.config().Login(context.Background(), callbackOpener(t, "auth-code", ""), io.Discard)
	require.NoError(t, err)
	assert.Equal(t, "bitpat_minted", a.Token)
	assert.Equal(t, "refresh-1", a.RefreshToken)
	assert.True(t, a.IsOAuthManaged(), "a login result should be OAuth-managed")
	assert.False(t, a.TokenExpiry.IsZero(), "the PAT expiry should be set after login")
	assert.False(t, a.JWTExpiry.IsZero(), "the JWT expiry should be set after login")
	m.assertCounts(t, 1, 1, "login is one code exchange plus one PAT exchange")
}

func TestLogin_StateMismatch(t *testing.T) {
	m := newOAuthMock(t)

	_, err := m.config().Login(context.Background(), callbackOpener(t, "auth-code", "WRONG-STATE"), io.Discard)
	assert.ErrorContains(t, err, "state mismatch")
}

func TestLogin_GuardsMissingConfig(t *testing.T) {
	_, err := (Config{ClientID: "x"}).Login(context.Background(), nil, io.Discard)
	assert.ErrorContains(t, err, "issuer")

	_, err = (Config{Issuer: "https://x"}).Login(context.Background(), nil, io.Discard)
	assert.ErrorContains(t, err, "client_id")
}

// oauthMock is a test double for the WorkOS token endpoint (/oauth2/token) and
// the monolith OIDC exchange (/oidc/token), with call counters.
type oauthMock struct {
	server *httptest.Server

	mu            sync.Mutex
	tokenCalls    int // /oauth2/token (authorization_code + refresh)
	exchangeCalls int // /oidc/token (JWT → PAT)

	jwt               string
	refreshToken      string
	jwtExpiresIn      int64
	pat               string
	patExpiresIn      int64
	failRefresh       bool
	failExchangeTimes int // number of leading /oidc/token calls to fail before succeeding
}

func newOAuthMock(t *testing.T) *oauthMock {
	t.Helper()
	m := &oauthMock{
		jwt:          makeJWT(time.Now().Add(time.Hour).Unix()),
		refreshToken: "refresh-1",
		jwtExpiresIn: 3600,
		pat:          "bitpat_minted",
		patExpiresIn: 3600,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()

		m.mu.Lock()
		m.tokenCalls++
		fail := m.failRefresh
		body := map[string]any{
			"access_token":  m.jwt,
			"refresh_token": m.refreshToken,
			"expires_in":    m.jwtExpiresIn,
			"token_type":    "Bearer",
		}
		m.mu.Unlock()

		if r.FormValue("grant_type") == "refresh_token" && fail {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
			return
		}
		_ = json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("/oidc/token", func(w http.ResponseWriter, _ *http.Request) {
		m.mu.Lock()
		m.exchangeCalls++
		shouldFail := m.failExchangeTimes > 0
		if shouldFail {
			m.failExchangeTimes--
		}
		body := map[string]any{
			"access_token": m.pat,
			"token_type":   "bearer",
			"expires_in":   m.patExpiresIn,
		}
		m.mu.Unlock()

		if shouldFail {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"error":"invalid_token"}`)
			return
		}
		_ = json.NewEncoder(w).Encode(body)
	})
	m.server = httptest.NewServer(mux)
	t.Cleanup(m.server.Close)
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

// assertCounts checks how many calls reached each endpoint. why states what the
// expected call pattern proves — which is the point of most of these tests, as
// the refresh ladder is defined by which steps it does and doesn't take.
func (m *oauthMock) assertCounts(t *testing.T, wantToken, wantExchange int, why string) {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	assert.Equal(t, wantToken, m.tokenCalls, "/oauth2/token calls: %s", why)
	assert.Equal(t, wantExchange, m.exchangeCalls, "/oidc/token calls: %s", why)
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
