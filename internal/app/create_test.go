package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

func TestCreate_RegisterFinishUpload_WithExplicitOrg(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"new-app"}`)
	})
	api.handle("/apps/new-app/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"btt","branch_name":"main"}`)
	})
	api.handle("/apps/new-app/bitrise.yml", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	})

	svc := NewService(api.client())
	got, err := svc.Create(context.Background(), CreateOptions{
		RepoURL:    "https://github.com/acme/widget.git",
		Branch:     "main",
		Title:      "Widget",
		Provider:   "auto",
		OrgSlug:    "acme",
		BitriseYML: "format_version: \"11\"\n",
	})
	require.NoError(t, err)

	want := CreateResult{
		Slug: "new-app", Title: "Widget", RepoURL: "https://github.com/acme/widget.git",
		DefaultBranch: "main", BuildTriggerToken: "btt",
		OrgSlug: "acme", StackID: DefaultStackID, ProjectType: DefaultProjectType,
		BitriseYMLUploaded: true,
	}
	assert.Equal(t, want, got)

	// Provider defaults to "custom" — matches the website's add-new-app flow.
	var reg bitriseapi.RegisterAppRequest
	require.NoError(t, json.Unmarshal(api.bodies["/apps/register"], &reg))
	assert.Equal(t, DefaultProvider, reg.Provider)
	assert.Equal(t, "acme", reg.OrganizationSlug)
	assert.Equal(t, "main", reg.DefaultBranchName)
	assert.Equal(t, FlowTypeCLI, reg.FlowType)

	// Finish defaults applied; config maps to project_type.
	var fin bitriseapi.FinishAppRequest
	require.NoError(t, json.Unmarshal(api.bodies["/apps/new-app/finish"], &fin))
	assert.Equal(t, DefaultStackID, fin.StackID)
	assert.Equal(t, "manual", fin.Mode)
	assert.Equal(t, DefaultProjectType, fin.ProjectType)
	assert.Equal(t, "other-config", fin.Config)
	assert.Equal(t, FlowTypeCLI, fin.FlowType)

	// Upload happened with the parsed YAML payload.
	assert.JSONEq(t, `{"app_config_datastore_yaml":{"format_version":"11"}}`, string(api.bodies["/apps/new-app/bitrise.yml"]))
}

func TestCreate_SkipsUpload_WhenNoYML(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"x"}`)
	})
	api.handle("/apps/x/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"t","branch_name":"main"}`)
	})

	svc := NewService(api.client())
	got, err := svc.Create(context.Background(), CreateOptions{RepoURL: "https://github.com/a/b.git", OrgSlug: "acme"})
	require.NoError(t, err)
	assert.False(t, got.BitriseYMLUploaded)
}

func TestCreate_InvalidYAML(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"x"}`)
	})
	api.handle("/apps/x/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"t","branch_name":"main"}`)
	})

	svc := NewService(api.client())
	_, err := svc.Create(context.Background(), CreateOptions{RepoURL: "https://github.com/a/b.git", OrgSlug: "acme", BitriseYML: "not: valid: yaml:"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse bitrise.yml")
}

func TestCreate_AutoDetectOrg_SingleWorkspaceWins(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/organizations", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"slug":"only-org","name":"Only"}]}`)
	})
	api.handle("/apps/register", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"slug":"x"}`)
	})
	api.handle("/apps/x/finish", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"build_trigger_token":"t","branch_name":"main"}`)
	})

	svc := NewService(api.client())
	got, err := svc.Create(context.Background(), CreateOptions{RepoURL: "https://github.com/a/b.git"})
	require.NoError(t, err)
	assert.Equal(t, "only-org", got.OrgSlug)

	var reg bitriseapi.RegisterAppRequest
	require.NoError(t, json.Unmarshal(api.bodies["/apps/register"], &reg))
	assert.Equal(t, "only-org", reg.OrganizationSlug)
}

func TestCreate_AutoDetectOrg_NoneFails(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/organizations", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[]}`)
	})

	svc := NewService(api.client())
	_, err := svc.Create(context.Background(), CreateOptions{RepoURL: "https://github.com/a/b.git"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no workspaces")
}

func TestCreate_AutoDetectOrg_MultipleFails(t *testing.T) {
	api := newStubAPI(t)
	api.handle("/organizations", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"slug":"a","name":"A"},{"slug":"b","name":"B"}]}`)
	})

	svc := NewService(api.client())
	_, err := svc.Create(context.Background(), CreateOptions{RepoURL: "https://github.com/a/b.git"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple workspaces")
}

func TestCreate_RequiresRepoURL(t *testing.T) {
	svc := NewService(newAPIClient(t, "https://unused.test"))
	_, err := svc.Create(context.Background(), CreateOptions{OrgSlug: "acme"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repo URL is required")
}

func TestResolveProvider(t *testing.T) {
	for input, want := range map[string]string{"": DefaultProvider, "auto": DefaultProvider, "github": "github"} {
		got, err := resolveProvider(input)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}
	_, err := resolveProvider("hg")
	require.Error(t, err)
}

func TestConfigIDForProjectType(t *testing.T) {
	// Pins the full project-type → config_id mapping. The expected values are
	// the server's own preset IDs (see the doc comment on
	// configIDForProjectType); any change here should be a reviewed,
	// intentional diff checked against that source.
	cases := map[string]string{
		"android":              "default-android-config",
		"cordova":              "default-cordova-config",
		"fastlane":             "default-fastlane-ios-config",
		"flutter":              "flutter-config-test-ios-android-web-0",
		"ionic":                "default-ionic-config",
		"ios":                  "default-ios-config",
		"java":                 "default-java-gradle-config",
		"kotlin-multiplatform": "default-kotlin-multiplatform-config",
		"macos":                "default-macos-config",
		"node-js":              "default-node-js-npm-config",
		"python":               "default-python-pip-config",
		"react-native":         "default-react-native-config",
		"ruby":                 "default-ruby-config",
		"":                     "other-config",
		"other":                "other-config",
		"xamarin":              "other-config",
		"unknown":              "other-config",
	}
	for input, want := range cases {
		assert.Equal(t, want, configIDForProjectType(input), "input %q", input)
	}
}

func TestDeriveTitle(t *testing.T) {
	cases := map[string]string{
		"https://github.com/acme/widget.git": "widget",
		"https://github.com/acme/widget":     "widget",
		"git@github.com:acme/widget.git":     "widget",
		"":                                   "",
	}
	for input, want := range cases {
		assert.Equal(t, want, deriveTitle(input), "input %q", input)
	}
}

// stubAPI is the multiplexer used by Create's tests. It maps URL paths to
// canned responses and records the request bodies that arrived.
type stubAPI struct {
	t        *testing.T
	baseURL  string
	registry map[string]http.HandlerFunc
	bodies   map[string][]byte
}

func newStubAPI(t *testing.T) *stubAPI {
	t.Helper()
	s := &stubAPI{t: t, registry: map[string]http.HandlerFunc{}, bodies: map[string][]byte{}}
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

func (s *stubAPI) client() *bitriseapi.Client { return newAPIClient(s.t, s.baseURL) }
