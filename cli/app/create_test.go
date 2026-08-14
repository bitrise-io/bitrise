package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/internal/config"
)

func TestCreateCmd_HappyPath_PersistsAppID(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"new-app"}`)
	})
	api.handle("/apps/new-app/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"btt","branch_name":"main"}`)
	})

	cmd, out, errOut := newTestCreateCmd(t, api.baseURL)
	err := runCreate(cmd, noCallDetector{t: t}, createFlags{
		repoURL:   "https://github.com/acme/widget.git",
		branch:    "main",
		title:     "Widget",
		workspace: "acme",
	})
	require.NoError(t, err)

	assert.Contains(t, out.String(), "Created app new-app")
	assert.Regexp(t, `Title:\s+Widget`, out.String())
	assert.Contains(t, errOut.String(), "Set app_id=new-app in")
	assert.NotContains(t, errOut.String(), "Detected repo URL from git")

	// The trigger token is a credential: raw output masks it (like
	// 'auth status'), --format json still carries the real value.
	assert.Regexp(t, `Build trigger token:\s+\*+ \(set`, out.String())
	assert.NotContains(t, out.String(), "btt", "the raw output must not leak the trigger token")

	globalCfg, loadErr := config.Load()
	require.NoError(t, loadErr)
	assert.Equal(t, "new-app", globalCfg.AppID)
}

func TestCreateCmd_PrintsResultEvenWhenPersistAppIDFails(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"new-app"}`)
	})
	api.handle("/apps/new-app/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"btt","branch_name":"main"}`)
	})

	cmd, out, errOut := newTestCreateCmd(t, api.baseURL)
	// internalconfig.Save renames a temp file onto this path; making it an
	// existing directory forces that rename to fail regardless of the
	// process's privileges (unlike a read-only permission bit, which root
	// ignores), without touching auth.yaml alongside it.
	cfgPath, err := config.Path()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(cfgPath, 0o700))

	runErr := runCreate(cmd, noCallDetector{t: t}, createFlags{
		repoURL:   "https://github.com/acme/widget.git",
		branch:    "main",
		workspace: "acme",
	})
	require.NoError(t, runErr, "the app was already created remotely; a local save failure must not turn into a command error")

	assert.Contains(t, out.String(), "Created app new-app", "the result must still print despite the save failure")
	assert.Contains(t, errOut.String(), "Warning: failed to save app_id")
}

func TestCreateCmd_DetectsRepoURLAndBranchFromGit(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"x"}`)
	})
	api.handle("/apps/x/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"t","branch_name":"detected-branch"}`)
	})

	cmd, _, errOut := newTestCreateCmd(t, api.baseURL)
	detector := stubGitDetector{remoteURL: "https://github.com/a/b.git", branch: "detected-branch"}
	err := runCreate(cmd, detector, createFlags{workspace: "acme"})
	require.NoError(t, err)

	assert.Contains(t, errOut.String(), "Detected repo URL from git: https://github.com/a/b.git")
	assert.Contains(t, errOut.String(), "Detected branch from git: detected-branch")

	var reg bitriseapi.RegisterAppRequest
	require.NoError(t, json.Unmarshal(api.bodies["/apps/register"], &reg))
	assert.Equal(t, "https://github.com/a/b.git", reg.RepoURL)
	assert.Equal(t, "detected-branch", reg.DefaultBranchName)
}

func TestCreateCmd_NoRepoURLDetected_Fails(t *testing.T) {
	cmd, _, _ := newTestCreateCmd(t, "https://unused.test")
	err := runCreate(cmd, stubGitDetector{}, createFlags{workspace: "acme"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--repo-url is required")
}

func TestCreateCmd_UploadsExplicitBitriseYML(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"x"}`)
	})
	api.handle("/apps/x/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"t","branch_name":"main"}`)
	})
	api.handle("/apps/x/bitrise.yml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	})

	ymlPath := t.TempDir() + "/bitrise.yml"
	require.NoError(t, os.WriteFile(ymlPath, []byte("format_version: \"11\"\n"), 0o600))

	cmd, out, errOut := newTestCreateCmd(t, api.baseURL)
	err := runCreate(cmd, stubGitDetector{}, createFlags{
		repoURL: "https://github.com/a/b.git", workspace: "acme", bitriseYML: ymlPath,
	})
	require.NoError(t, err)

	assert.Contains(t, errOut.String(), "Uploading bitrise.yml from "+ymlPath)
	assert.NotContains(t, out.String(), "Config:") // "server preset" line only appears when not uploaded
	assert.JSONEq(t, `{"app_config_datastore_yaml":{"format_version":"11"}}`, string(api.bodies["/apps/x/bitrise.yml"]))
}

func TestCreateCmd_NoBitriseYML_UsesServerPreset(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"x"}`)
	})
	api.handle("/apps/x/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"t","branch_name":"main"}`)
	})

	t.Chdir(t.TempDir())
	cmd, out, errOut := newTestCreateCmd(t, api.baseURL)
	err := runCreate(cmd, stubGitDetector{}, createFlags{repoURL: "https://github.com/a/b.git", workspace: "acme"})
	require.NoError(t, err)

	assert.Contains(t, errOut.String(), "No bitrise.yml provided — using server preset for project-type=other")
	assert.Regexp(t, `Config:\s+server preset`, out.String())
}

func TestCreateCmd_WarnsWhenPerDirConfigOverridesAppID(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"new-app"}`)
	})
	api.handle("/apps/new-app/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"t","branch_name":"main"}`)
	})

	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.WriteFile(dir+"/.bitrise-cli.yml", []byte("app_id: other-app\n"), 0o600))

	cmd, _, errOut := newTestCreateCmd(t, api.baseURL)
	err := runCreate(cmd, stubGitDetector{}, createFlags{repoURL: "https://github.com/a/b.git", workspace: "acme"})
	require.NoError(t, err)

	assert.Contains(t, errOut.String(), "pins app_id=other-app, which still wins at runtime")
}

func TestCreateCmd_PropagatesAutoDetectWorkspaceError(t *testing.T) {
	t.Setenv(cmdutil.EnvWorkspaceID, "")
	api := newStubAPI(t)
	api.handle("/organizations", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[]}`)
	})

	cmd, _, _ := newTestCreateCmd(t, api.baseURL)
	err := runCreate(cmd, stubGitDetector{}, createFlags{repoURL: "https://github.com/a/b.git"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no workspaces")
}

func TestCreateCmd_UsesDefaultWorkspaceFromConfig(t *testing.T) {
	t.Setenv(cmdutil.EnvWorkspaceID, "")
	api := newStubAPI(t)
	api.handle("/organizations", func(w http.ResponseWriter, _ *http.Request) {
		t.Error("auto-detect must not run when default_workspace_id is set")
		_, _ = io.WriteString(w, `{"data":[]}`)
	})
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"new-app"}`)
	})
	api.handle("/apps/new-app/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"t","branch_name":"main"}`)
	})

	cmd, _, errOut := newTestCreateCmdWithConfig(t, api.baseURL, config.Config{DefaultWorkspaceID: "cfg-ws"})
	require.NoError(t, runCreate(cmd, noCallDetector{t: t}, createFlags{
		repoURL: "https://github.com/a/b.git",
		branch:  "main",
	}))

	var reg map[string]any
	require.NoError(t, json.Unmarshal(api.bodies["/apps/register"], &reg))
	assert.Equal(t, "cfg-ws", reg["organization_slug"])
	assert.Contains(t, errOut.String(), "Using default workspace: cfg-ws")
}

func TestCreateCmd_WorkspaceFlagWinsOverDefault(t *testing.T) {
	t.Setenv(cmdutil.EnvWorkspaceID, "")
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"new-app"}`)
	})
	api.handle("/apps/new-app/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"t","branch_name":"main"}`)
	})

	cmd, _, errOut := newTestCreateCmdWithConfig(t, api.baseURL, config.Config{DefaultWorkspaceID: "cfg-ws"})
	require.NoError(t, runCreate(cmd, noCallDetector{t: t}, createFlags{
		repoURL:   "https://github.com/a/b.git",
		branch:    "main",
		workspace: "flag-ws",
	}))

	var reg map[string]any
	require.NoError(t, json.Unmarshal(api.bodies["/apps/register"], &reg))
	assert.Equal(t, "flag-ws", reg["organization_slug"])
	assert.NotContains(t, errOut.String(), "Using default workspace", "an explicit --workspace is not a default")
}

func TestCreateCmd_SuppressesInferenceBreadcrumbsForJSON(t *testing.T) {
	t.Setenv(cmdutil.EnvWorkspaceID, "")
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"new-app"}`)
	})
	api.handle("/apps/new-app/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"t","branch_name":"main"}`)
	})

	cmd, _, errOut := newTestCreateCmdWithConfig(t, api.baseURL, config.Config{DefaultWorkspaceID: "cfg-ws"})
	require.NoError(t, cmd.Flags().Set("format", "json"))
	require.NoError(t, runCreate(cmd, stubGitDetector{remoteURL: "https://github.com/a/b.git", branch: "dev"}, createFlags{}))

	errText := errOut.String()
	assert.NotContains(t, errText, "Detected repo URL from git")
	assert.NotContains(t, errText, "Detected branch from git")
	assert.NotContains(t, errText, "Using default workspace")
	assert.NotContains(t, errText, "No bitrise.yml provided")
	// The app_id notice is not an inference hint — it reports a file this
	// command wrote, so it stays even for machine-readable output.
	assert.Contains(t, errText, "Set app_id=new-app in")
}

func TestCreateCmd_RejectsArgs(t *testing.T) {
	cmd := NewCreateCommand()
	cmd.SetArgs([]string{"unexpected"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	assert.Error(t, cmd.Execute(), "positional args should be rejected before the command runs")
}

// stubGitDetector is an internalapp.GitDetector stand-in for tests that
// exercise git auto-detection.
type stubGitDetector struct {
	remoteURL string
	branch    string
}

func (d stubGitDetector) RemoteURL(context.Context) (string, error)     { return d.remoteURL, nil }
func (d stubGitDetector) CurrentBranch(context.Context) (string, error) { return d.branch, nil }

// noCallDetector fails the test if git detection is invoked, for cases where
// --repo-url/--branch are set explicitly and detection should be skipped.
type noCallDetector struct{ t *testing.T }

func (d noCallDetector) RemoteURL(context.Context) (string, error) {
	d.t.Fatal("RemoteURL should not be called when --repo-url is set")
	return "", nil
}

func (d noCallDetector) CurrentBranch(context.Context) (string, error) {
	d.t.Fatal("CurrentBranch should not be called when --branch is set")
	return "", nil
}

// newTestCreateCmd builds a NewCreateCommand() wired to apiBaseURL, with
// stdout/stderr captured in the returned buffers. Tests call runCreate
// directly with their own createFlags rather than setting flags on cmd, so
// this cmd only needs to supply the --format flag and a resolved context.
func newTestCreateCmd(t *testing.T, apiBaseURL string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, auth.Save(auth.Auth{Token: "test-token"}))

	cmd := NewCreateCommand()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	resolved := config.Resolve(config.Config{}, config.Config{}, config.Config{APIBaseURL: apiBaseURL})
	cmd.SetContext(config.WithResolved(t.Context(), resolved))
	return cmd, &out, &errOut
}

// newTestCreateCmdWithConfig is newTestCreateCmd for tests that need extra
// resolved config keys (e.g. default_workspace_id) alongside the API base URL.
func newTestCreateCmdWithConfig(t *testing.T, apiBaseURL string, global config.Config) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	cmd, out, errOut := newTestCreateCmd(t, apiBaseURL)
	global.APIBaseURL = apiBaseURL
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolve(config.Config{}, config.Config{}, global)))
	return cmd, out, errOut
}

// stubAPI is the multiplexer used by create's tests. It maps URL paths to
// canned responses and records the request bodies that arrived.
type stubAPI struct {
	baseURL  string
	registry map[string]http.HandlerFunc
	bodies   map[string][]byte
}

func newStubAPI(t *testing.T) *stubAPI {
	t.Helper()
	s := &stubAPI{registry: map[string]http.HandlerFunc{}, bodies: map[string][]byte{}}
	srv := newFakeServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.bodies[r.URL.Path] = body
		h, ok := s.registry[r.URL.Path]
		if !ok {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		h(w, r)
	})
	s.baseURL = srv.URL
	return s
}

func (s *stubAPI) handle(path string, h http.HandlerFunc) { s.registry[path] = h }
