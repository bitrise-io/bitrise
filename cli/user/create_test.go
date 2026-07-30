package user

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
)

var validCreateArgs = []string{"--email", "a@b.io", "--username", "alice", "--first-name", "Alice", "--last-name", "L"}

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

	cmd, out := newTestCreateCmd(t, srv.URL, "hunter2\n", validCreateArgs...)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "/users", gotPath)
	assert.Equal(t, `✓ Account created
Email:    a@b.io
Username: alice
ID:       u-aaa

Check your email and click the verification link, then run:
  bitrise auth login --email a@b.io
`, out.String())
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

	cmd, _ := newTestCreateCmd(t, srv.URL, "hunter2\n", validCreateArgs...)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email: taken")
}

// TestCreateCmd_MissingRequiredFlags drives Execute rather than RunE, because
// cobra enforces MarkFlagRequired in Execute's ValidateRequiredFlags.
func TestCreateCmd_MissingRequiredFlags(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"missing email", []string{"--username", "a", "--first-name", "A", "--last-name", "B"}, `required flag(s) "email" not set`},
		{"missing username", []string{"--email", "a@b.io", "--first-name", "A", "--last-name", "B"}, `required flag(s) "username" not set`},
		{"missing first name", []string{"--email", "a@b.io", "--username", "a", "--last-name", "B"}, `required flag(s) "first-name" not set`},
		{"missing last name", []string{"--email", "a@b.io", "--username", "a", "--first-name", "A"}, `required flag(s) "last-name" not set`},
		{"all missing", nil, `required flag(s) "email", "first-name", "last-name", "username" not set`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, _ := newTestCreateCmd(t, "", "", tc.args...)
			err := cmd.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestCreateCmd_ExplicitEmptyFlagErrors covers what MarkFlagRequired cannot:
// it only checks that a flag was set, so `--email ""` passes validation.
func TestCreateCmd_ExplicitEmptyFlagErrors(t *testing.T) {
	cmd, _ := newTestCreateCmd(t, "", "", "--email", "", "--username", "alice", "--first-name", "Alice", "--last-name", "L")

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--email requires a non-empty value")
}

func TestCreateCmd_EmptyPasswordErrors(t *testing.T) {
	cmd, _ := newTestCreateCmd(t, "", "\n", validCreateArgs...)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password is empty")
}

func newTestCreateCmd(t *testing.T, webBaseURL, stdin string, args ...string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	// Always point the web base URL somewhere unroutable, so a test that is
	// expected to fail before the signup request can never reach production.
	if webBaseURL == "" {
		webBaseURL = "http://unused.test"
	}
	t.Setenv(cmdutil.EnvWebBaseURL, webBaseURL)

	cmd := NewCreateCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetArgs(args)
	cmd.SetContext(t.Context())
	return cmd, &out
}
