package workspace

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/internal/auth"
	"github.com/bitrise-io/bitrise/v3/internal/config"
)

func TestListCmd_HumanTable(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":[{"slug":"b-ws","name":"Bravo"},{"slug":"a-ws","name":"Alpha"}]}`))
	})

	cmd, out := newTestCmd(t, NewListCommand(), srv.URL)
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "/organizations", gotPath)
	assert.Contains(t, out.String(), "ID")
	assert.Contains(t, out.String(), "NAME")
	assert.Less(t, bytes.Index(out.Bytes(), []byte("Alpha")), bytes.Index(out.Bytes(), []byte("Bravo")))
	assert.Contains(t, out.String(), "a-ws")
}

func TestListCmd_JSON(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"slug":"a-ws","name":"Alpha"}]}`))
	})

	cmd, out := newTestCmd(t, NewListCommand(), srv.URL)
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.JSONEq(t, `{"items":[{"id":"a-ws","name":"Alpha"}]}`, out.String())
}

func TestListCmd_EmptyHuman(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	})

	cmd, out := newTestCmd(t, NewListCommand(), srv.URL)
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "No workspaces found.\n", out.String())
}

func TestListCmd_PropagatesAPIError(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"forbidden"}`))
	})

	cmd, _ := newTestCmd(t, NewListCommand(), srv.URL)
	err := cmd.RunE(cmd, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing workspaces failed")
	assert.Contains(t, err.Error(), "forbidden")
}

func TestListCmd_RejectsPositionalArgs(t *testing.T) {
	cmd := NewListCommand()
	cmd.SetArgs([]string{"unexpected"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	assert.Error(t, cmd.Execute(), "positional args should be rejected before the command runs")
}

// newTestCmd wires cmd to apiBaseURL with global config extended by extra,
// and captures its stdout in the returned buffer.
func newTestCmd(t *testing.T, cmd *cobra.Command, apiBaseURL string, extra ...config.Config) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("BITRISE_TOKEN", "") // an exported token would outrank the fixture in cmdutil.ResolveToken
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	global := config.Config{}
	if len(extra) > 0 {
		global = extra[0]
	}
	global.APIBaseURL = apiBaseURL

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolve(config.Config{}, config.Config{}, global)))
	return cmd, &out
}

func newFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}
