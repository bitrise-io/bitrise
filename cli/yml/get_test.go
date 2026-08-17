package yml

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestGetCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _ := newTestGetCmd(t, "http://unused.test")
	err := cmd.RunE(cmd, nil)
	require.EqualError(t, err, "--app is required")
}

func TestGetCmd_App(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte("format_version: \"13\"\n"))
	}))
	t.Cleanup(srv.Close)

	cmd, out := newTestGetCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "app-slug"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "/apps/app-slug/bitrise.yml", gotPath)
	assert.Equal(t, "format_version: \"13\"\n", out.String())
}

func TestGetCmd_Build(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte("format_version: \"13\"\n"))
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newTestGetCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "app-slug"))
	require.NoError(t, cmd.Flags().Set("build", "build-slug"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "/apps/app-slug/builds/build-slug/bitrise.yml", gotPath)
}

func TestGetCmd_AppendsMissingTrailingNewline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("format_version: \"13\""))
	}))
	t.Cleanup(srv.Close)

	cmd, out := newTestGetCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "app-slug"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Equal(t, "format_version: \"13\"\n", out.String())
}

func TestGetCmd_JSONOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("format_version: \"13\"\n"))
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newTestGetCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "app-slug"))
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, cmd.RunE(cmd, nil))
}

func newTestGetCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("BITRISE_TOKEN", "") // an exported token would outrank the fixture in cmdutil.ResolveToken
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewGetCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out
}
