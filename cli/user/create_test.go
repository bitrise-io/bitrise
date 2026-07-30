package user

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

func TestCreateCmd_HappyPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/sign_up":
			_, _ = w.Write([]byte(`<meta name="csrf-token" content="t" />`))
		case "/users":
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"slug":"u-aaa","email":"a@b.io","username":"alice","first_name":"Alice","last_name":"L","confirmed_at":null}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	cmd := newTestCreateCmd(t, srv.URL, "hunter2\n")
	require.NoError(t, cmd.Flags().Set("email", "a@b.io"))
	require.NoError(t, cmd.Flags().Set("username", "alice"))
	require.NoError(t, cmd.Flags().Set("first-name", "Alice"))
	require.NoError(t, cmd.Flags().Set("last-name", "L"))

	require.NoError(t, cmd.RunE(cmd, nil))
	assert.Equal(t, "/users", gotPath)
}

func TestCreateCmd_FieldErrors422(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/sign_up" {
			_, _ = w.Write([]byte(`<meta name="csrf-token" content="t" />`))
			return
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"errors":{"email":[{"error":"taken"}]}}`))
	}))
	t.Cleanup(srv.Close)

	cmd := newTestCreateCmd(t, srv.URL, "hunter2\n")
	require.NoError(t, cmd.Flags().Set("email", "a@b.io"))
	require.NoError(t, cmd.Flags().Set("username", "alice"))
	require.NoError(t, cmd.Flags().Set("first-name", "Alice"))
	require.NoError(t, cmd.Flags().Set("last-name", "L"))

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email: taken")
}

func TestCreateCmd_MissingRequiredFlags(t *testing.T) {
	cases := []struct {
		name    string
		set     map[string]string
		wantErr string
	}{
		{"missing email", map[string]string{"username": "a", "first-name": "A", "last-name": "B"}, "--email is required"},
		{"missing username", map[string]string{"email": "a@b.io", "first-name": "A", "last-name": "B"}, "--username is required"},
		{"missing first name", map[string]string{"email": "a@b.io", "username": "a", "last-name": "B"}, "--first-name is required"},
		{"missing last name", map[string]string{"email": "a@b.io", "username": "a", "first-name": "A"}, "--last-name is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newTestCreateCmd(t, "", "")
			for k, v := range tc.set {
				require.NoError(t, cmd.Flags().Set(k, v))
			}
			err := cmd.RunE(cmd, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestCreateCmd_EmptyPasswordErrors(t *testing.T) {
	cmd := newTestCreateCmd(t, "", "\n")
	require.NoError(t, cmd.Flags().Set("email", "a@b.io"))
	require.NoError(t, cmd.Flags().Set("username", "alice"))
	require.NoError(t, cmd.Flags().Set("first-name", "Alice"))
	require.NoError(t, cmd.Flags().Set("last-name", "L"))

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password is empty")
}

func newTestCreateCmd(t *testing.T, webBaseURL, stdin string) *cobra.Command {
	t.Helper()
	if webBaseURL != "" {
		t.Setenv(cmdutil.EnvWebBaseURL, webBaseURL)
	}

	cmd := NewCreateCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetContext(t.Context())
	return cmd
}
