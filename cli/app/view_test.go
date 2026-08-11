package app

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestViewCmd_PositionalArg(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"my-app","title":"My App","provider":"github","repo_url":"https://github.com/x/y","owner":{"slug":"acme"}}}`))
	})

	cmd, out := newTestViewCmd(t, srv.URL)
	require.NoError(t, runView(cmd, []string{"my-app"}, false, unusedBrowser(t)))

	assert.Equal(t, "/apps/my-app", gotPath)
	// Matched padding-agnostically: the label/value pairing is the contract,
	// the column width is cosmetic.
	assert.Regexp(t, `Title:\s+My App`, out.String())
	assert.Regexp(t, `ID:\s+my-app`, out.String())
	assert.Regexp(t, `Workspace:\s+acme`, out.String())
}

func TestViewCmd_FlagFallback(t *testing.T) {
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"my-app","title":"My App","provider":"github","owner":{}}}`))
	})

	cmd, _ := newTestViewCmd(t, srv.URL)
	require.NoError(t, cmd.Flags().Set("app", "my-app"))
	require.NoError(t, runView(cmd, nil, false, unusedBrowser(t)))

	assert.Equal(t, "/apps/my-app", gotPath)
}

func TestViewCmd_RequiresAppSlug(t *testing.T) {
	cmd, _ := newTestViewCmd(t, "https://unused.test")
	err := runView(cmd, nil, false, unusedBrowser(t))
	require.EqualError(t, err, "--app is required")
}

func TestViewCmd_AppNotFound(t *testing.T) {
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})

	cmd, _ := newTestViewCmd(t, srv.URL)
	err := runView(cmd, []string{"missing-app"}, false, unusedBrowser(t))
	require.EqualError(t, err, `app "missing-app" not found`)
}

func TestViewCmd_Web_OpensBrowserAndSkipsAPICall(t *testing.T) {
	cmd, _ := newTestViewCmd(t, "https://unused.test")
	t.Setenv(cmdutil.EnvWebBaseURL, "https://app.bitrise.io")

	var gotURL string
	err := runView(cmd, []string{"my-app"}, true, func(url string) error {
		gotURL = url
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "https://app.bitrise.io/app/my-app", gotURL)
}

func TestViewCmd_RejectsMultipleArgs(t *testing.T) {
	cmd := NewViewCommand()
	cmd.SetArgs([]string{"app-1", "app-2"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	assert.Error(t, cmd.Execute(), "more than one positional arg should be rejected before the command runs")
}

// unusedBrowser fails the test if the browser opener is ever invoked, for
// cases where --web is off and no browser call is expected.
func unusedBrowser(t *testing.T) func(string) error {
	t.Helper()
	return func(url string) error {
		t.Fatalf("unexpected browser open: %s", url)
		return nil
	}
}

// newTestViewCmd builds a NewViewCommand() wired to apiBaseURL, with its
// output captured in the returned buffer.
func newTestViewCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewViewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out
}
