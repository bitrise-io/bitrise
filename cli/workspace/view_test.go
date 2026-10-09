package workspace

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v3/internal/config"
)

func TestViewCmd_PositionalArg(t *testing.T) {
	t.Setenv(cmdutil.EnvWorkspaceID, "env-ws")
	var gotPath string
	srv := newViewFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"a-ws","name":"Alpha"}}`))
	})

	cmd, out := newTestCmd(t, NewViewCommand(), srv.URL)
	require.NoError(t, cmd.Flags().Set(cmdutil.FlagWorkspace, "flag-ws"))
	require.NoError(t, runView(cmd, []string{"a-ws"}, false, unusedBrowser(t)))

	assert.Equal(t, "/organizations/a-ws", gotPath)
	assert.Regexp(t, `Name:\s+Alpha`, out.String())
	assert.Regexp(t, `ID:\s+a-ws`, out.String())
}

func TestViewCmd_WorkspaceFlagByName_SkipsSecondFetch(t *testing.T) {
	var paths []string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		_, _ = w.Write([]byte(`{"data":[{"slug":"acme","name":"Acme Corp"}]}`))
	})

	cmd, out := newTestCmd(t, NewViewCommand(), srv.URL)
	require.NoError(t, cmd.Flags().Set(cmdutil.FlagWorkspace, "Acme Corp"))
	require.NoError(t, runView(cmd, nil, false, unusedBrowser(t)))

	assert.Equal(t, []string{"/organizations"}, paths, "a name match already holds the workspace")
	assert.Regexp(t, `Name:\s+Acme Corp`, out.String())
	assert.Regexp(t, `ID:\s+acme`, out.String())
}

func TestViewCmd_PositionalSlug_SkipsSecondFetch(t *testing.T) {
	var paths []string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		_, _ = w.Write([]byte(`{"data":[{"slug":"acme","name":"Acme Corp"}]}`))
	})

	cmd, out := newTestCmd(t, NewViewCommand(), srv.URL)
	require.NoError(t, runView(cmd, []string{"acme"}, false, unusedBrowser(t)))

	assert.Equal(t, []string{"/organizations"}, paths, "the workspace list already holds the workspace")
	assert.Equal(t, "Name: Acme Corp\nID:   acme\n", out.String())
}

func TestViewCmd_EnvSkipsNameResolution(t *testing.T) {
	t.Setenv(cmdutil.EnvWorkspaceID, "env-ws")
	var paths []string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		_, _ = w.Write([]byte(`{"data":{"slug":"env-ws","name":"Env"}}`))
	})

	cmd, _ := newTestCmd(t, NewViewCommand(), srv.URL)
	require.NoError(t, runView(cmd, nil, false, unusedBrowser(t)))

	assert.Equal(t, []string{"/organizations/env-ws"}, paths)
}

func TestViewCmd_DefaultFromConfig(t *testing.T) {
	t.Setenv(cmdutil.EnvWorkspaceID, "")
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"cfg-ws","name":"Cfg"}}`))
	})

	cmd, _ := newTestCmd(t, NewViewCommand(), srv.URL, config.Config{DefaultWorkspaceID: "cfg-ws"})
	require.NoError(t, runView(cmd, nil, false, unusedBrowser(t)))

	assert.Equal(t, "/organizations/cfg-ws", gotPath)
}

func TestViewCmd_SoleWorkspace(t *testing.T) {
	t.Setenv(cmdutil.EnvWorkspaceID, "")
	var gotPath string
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/organizations" {
			_, _ = w.Write([]byte(`{"data":[{"slug":"only-ws","name":"Only"}]}`))
			return
		}
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"slug":"only-ws","name":"Only"}}`))
	})

	cmd, _ := newTestCmd(t, NewViewCommand(), srv.URL)
	require.NoError(t, runView(cmd, nil, false, unusedBrowser(t)))

	assert.Equal(t, "/organizations/only-ws", gotPath)
}

func TestViewCmd_NotFound(t *testing.T) {
	srv := newViewFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	})

	cmd, _ := newTestCmd(t, NewViewCommand(), srv.URL)
	err := runView(cmd, []string{"missing-ws"}, false, unusedBrowser(t))
	require.EqualError(t, err, `viewing workspace failed: workspace "missing-ws" not found`)
}

func TestViewCmd_WrapsAPIError(t *testing.T) {
	srv := newViewFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	cmd, _ := newTestCmd(t, NewViewCommand(), srv.URL)
	err := runView(cmd, []string{"a-ws"}, false, unusedBrowser(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "viewing workspace failed")
	assert.Contains(t, err.Error(), "boom")
}

func TestViewCmd_JSON(t *testing.T) {
	srv := newViewFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"slug":"a-ws","name":"Alpha"}}`))
	})

	cmd, out := newTestCmd(t, NewViewCommand(), srv.URL)
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, runView(cmd, []string{"a-ws"}, false, unusedBrowser(t)))

	assert.JSONEq(t, `{"id":"a-ws","name":"Alpha"}`, out.String())
}

func TestViewCmd_Web_EnvSkipsAPICall(t *testing.T) {
	t.Setenv(cmdutil.EnvWebBaseURL, "https://app.bitrise.io")
	t.Setenv(cmdutil.EnvWorkspaceID, "env-ws")
	cmd, _ := newTestCmd(t, NewViewCommand(), "https://unused.test")

	var gotURL string
	require.NoError(t, runView(cmd, nil, true, func(url string) error {
		gotURL = url
		return nil
	}))
	assert.Equal(t, "https://app.bitrise.io/workspaces/env-ws", gotURL)
}

func TestViewCmd_Web_ResolvesUserProvidedName(t *testing.T) {
	t.Setenv(cmdutil.EnvWebBaseURL, "https://app.bitrise.io")
	srv := newFakeServer(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"slug":"acme","name":"Acme Corp"}]}`))
	})
	cmd, _ := newTestCmd(t, NewViewCommand(), srv.URL)

	var gotURL string
	require.NoError(t, runView(cmd, []string{"Acme Corp"}, true, func(url string) error {
		gotURL = url
		return nil
	}))
	assert.Equal(t, "https://app.bitrise.io/workspaces/acme", gotURL)
}

func TestViewCmd_Web_EscapesID(t *testing.T) {
	t.Setenv(cmdutil.EnvWebBaseURL, "https://app.bitrise.io")
	srv := newViewFakeServer(t, func(_ http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s", r.URL.Path)
	})
	cmd, _ := newTestCmd(t, NewViewCommand(), srv.URL)

	var gotURL string
	require.NoError(t, runView(cmd, []string{"acme/x#y"}, true, func(url string) error {
		gotURL = url
		return nil
	}))
	assert.Equal(t, "https://app.bitrise.io/workspaces/acme%2Fx%23y", gotURL)
}

func TestViewCmd_RejectsMultipleArgs(t *testing.T) {
	cmd := NewViewCommand()
	cmd.SetArgs([]string{"ws-1", "ws-2"})
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

// newViewFakeServer answers the name→slug resolver's GET /organizations
// lookup with no matches, so the value passes through as an ID, before
// deferring to handler.
func newViewFakeServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/organizations" {
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		handler(w, r)
	})
}
