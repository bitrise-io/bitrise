package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
)

func newTestCmd(t *testing.T, stdin string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&strings.Builder{})
	return cmd
}

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
