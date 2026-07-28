package auth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/log"
)

func TestRunTokenLogin_SavesToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	require.NoError(t, runTokenLogin(newTestCmd(t, "bitpat_faketoken\n")))

	saved, err := auth.Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_faketoken", saved.Token)
	assert.False(t, saved.IsOAuthManaged())
}

func TestRunTokenLogin_EmptyTokenErrors(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := runTokenLogin(newTestCmd(t, "\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token is empty")
}

func TestRunEmailLogin_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/sign_in":
			if r.Method == http.MethodGet {
				_, _ = w.Write([]byte(`<meta name="csrf-token" content="t" />`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		case "/me/profile/security":
			_, _ = w.Write([]byte(`<meta name="csrf-token" content="t2" />`))
		case "/me/profile/security/user_auth_tokens":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"token":"bitpat_minted"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(cmdutil.EnvWebBaseURL, srv.URL)

	require.NoError(t, runEmailLogin(newTestCmd(t, "hunter2\n"), "alice@example.com", true))

	saved, err := auth.Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_minted", saved.Token)
}

func TestRunEmailLogin_UnconfirmedEmail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/sign_in" && r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`<meta name="csrf-token" content="t" />`))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"You have to confirm your email address before continuing."}`))
	}))
	defer srv.Close()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(cmdutil.EnvWebBaseURL, srv.URL)

	err := runEmailLogin(newTestCmd(t, "pw\n"), "alice@example.com", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hasn't verified its email yet")
}

func TestRunEmailLogin_EmptyPasswordErrors(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := runEmailLogin(newTestCmd(t, "\n"), "alice@example.com", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password is empty")
}

func TestAuthLogin_ExplicitEmptyEmailErrors(t *testing.T) {
	cmd := NewLoginCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--email", ""})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--email requires a non-empty value")
}

func TestAuthLogin_PasswordStdinWithoutEmailErrors(t *testing.T) {
	cmd := NewLoginCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--password-stdin"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--password-stdin requires --email")
}

func TestCheckPasswordStdinPiped_ErrorsWhenTerminal(t *testing.T) {
	err := checkPasswordStdinPiped(true, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--password-stdin requires piped stdin")
}

func TestCheckPasswordStdinPiped_OKCases(t *testing.T) {
	assert.NoError(t, checkPasswordStdinPiped(true, false))  // --password-stdin + piped input: fine
	assert.NoError(t, checkPasswordStdinPiped(false, true))  // no --password-stdin, terminal: interactive masked prompt, fine
	assert.NoError(t, checkPasswordStdinPiped(false, false)) // no --password-stdin, non-terminal: fine
}

func TestRunOAuthLogin_SavesOAuthManagedToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_, _ = w.Write([]byte(`{"access_token":"jwt-1","refresh_token":"refresh-1","expires_in":3600}`))
	})
	mux.HandleFunc("/oidc/token", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"bitpat_oauth","expires_in":3600}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(cmdutil.EnvOAuthIssuer, srv.URL)
	t.Setenv(cmdutil.EnvOIDCTokenEndpoint, srv.URL+"/oidc/token")
	t.Setenv(cmdutil.EnvOAuthClientID, "https://cli.example/cimd.json")

	// Fake "browser": hit the loopback callback directly instead of opening a real one.
	fakeBrowser := func(rawURL string) error {
		u, err := url.Parse(rawURL)
		if err != nil {
			return err
		}
		q := u.Query()
		cb := q.Get("redirect_uri") + "?code=auth-code&state=" + url.QueryEscape(q.Get("state"))
		resp, err := http.Get(cb)
		if err != nil {
			return err
		}
		return resp.Body.Close()
	}

	cmd := newTestCmd(t, "")
	require.NoError(t, doOAuthLogin(cmd, fakeBrowser))

	saved, err := auth.Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_oauth", saved.Token)
	assert.True(t, saved.IsOAuthManaged())
}

func TestDoOAuthLogin_FailsFastOverSSH(t *testing.T) {
	t.Setenv("SSH_CONNECTION", "10.0.0.1 1234 10.0.0.2 22")

	browserCalled := false
	err := doOAuthLogin(newTestCmd(t, ""), func(string) error {
		browserCalled = true
		return nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--with-token")
	assert.False(t, browserCalled, "should fail before ever trying to open a browser")
}

// The tests below exercise NewLoginCommand()'s actual cobra dispatch (flag
// parsing, mutual exclusivity, and the interactive-vs-piped default) end to
// end, rather than calling the run*Login functions directly.

func TestAuthLogin_EmailAndWithTokenMutuallyExclusive(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd := NewLoginCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetIn(strings.NewReader("anything\n"))
	cmd.SetArgs([]string{"--email", "alice@example.com", "--with-token"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestAuthLogin_OAuthRejectsWithToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd := NewLoginCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--oauth", "--with-token"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestAuthLogin_WarnsWhenEnvTokenShadowsSavedToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(auth.EnvToken, "ci-env-token")

	var logBuf strings.Builder
	log.InitGlobalLogger(log.LoggerOpts{LoggerType: log.ConsoleLogger, Producer: log.BitriseCLI, Writer: &logBuf})

	cmd := NewLoginCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetIn(strings.NewReader("bitpat_saved\n"))
	cmd.SetArgs([]string{"--with-token"})

	require.NoError(t, cmd.Execute())

	saved, err := auth.Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_saved", saved.Token)

	out := logBuf.String()
	assert.Contains(t, out, auth.EnvToken)
	assert.Contains(t, out, "takes precedence")
}

func TestAuthLogin_NoShadowWarningWhenEnvUnset(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv(auth.EnvToken, "")

	var logBuf strings.Builder
	log.InitGlobalLogger(log.LoggerOpts{LoggerType: log.ConsoleLogger, Producer: log.BitriseCLI, Writer: &logBuf})

	cmd := NewLoginCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetIn(strings.NewReader("bitpat_saved\n"))
	cmd.SetArgs([]string{"--with-token"})

	require.NoError(t, cmd.Execute())

	out := logBuf.String()
	assert.Contains(t, out, "Saved access token")
	assert.NotContains(t, out, "takes precedence")
}

func TestAuthLogin_DefaultNonInteractive_ReadsTokenFromStdin(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd := NewLoginCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// A strings.Reader isn't a terminal, so the default routes to
	// token-from-stdin even with no mode flag at all, not the browser flow.
	cmd.SetIn(strings.NewReader("bitpat_piped\n"))
	cmd.SetArgs(nil)

	require.NoError(t, cmd.Execute())

	saved, err := auth.Load()
	require.NoError(t, err)
	assert.Equal(t, "bitpat_piped", saved.Token)
}

func newTestCmd(t *testing.T, stdin string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	return cmd
}
