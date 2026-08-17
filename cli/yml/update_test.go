package yml

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
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestUpdateCmd_RequiresApp(t *testing.T) {
	t.Setenv(cmdutil.EnvAppID, "")
	t.Setenv(cmdutil.EnvAppIDLegacy, "")

	cmd, _ := newTestUpdateCmd(t, "http://unused.test", "")
	err := cmd.RunE(cmd, nil)
	require.EqualError(t, err, "--app is required")
}

func TestUpdateCmd_FromStdin(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	cmd, stderr := newTestUpdateCmd(t, srv.URL, "format_version: \"13\"\n")
	require.NoError(t, cmd.Flags().Set("app", "app-slug"))
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.JSONEq(t, `{"app_config_datastore_yaml":{"format_version":"13"}}`, gotBody)
	assert.Equal(t, "bitrise.yml updated for app app-slug\n", stderr.String())
}

func TestUpdateCmd_EmptyContent(t *testing.T) {
	cmd, _ := newTestUpdateCmd(t, "http://unused.test", "")
	require.NoError(t, cmd.Flags().Set("app", "app-slug"))

	err := cmd.RunE(cmd, nil)
	require.Error(t, err)
}

func newTestUpdateCmd(t *testing.T, apiBaseURL, stdin string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("BITRISE_TOKEN", "") // an exported token would outrank the fixture in cmdutil.ResolveToken
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewUpdateCommand()
	cmd.SetIn(strings.NewReader(stdin))
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &stderr
}
