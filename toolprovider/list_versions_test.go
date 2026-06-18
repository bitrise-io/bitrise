package toolprovider

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeVersionProvider struct {
	versions map[provider.ToolID][]string
	err      error
}

func (f fakeVersionProvider) ID() string       { return "fake" }
func (f fakeVersionProvider) Bootstrap() error { return nil }
func (f fakeVersionProvider) InstallTool(provider.ToolRequest) (provider.ToolInstallResult, error) {
	return provider.ToolInstallResult{}, nil
}
func (f fakeVersionProvider) ActivateEnv(provider.ToolInstallResult) (provider.EnvironmentActivation, error) {
	return provider.EnvironmentActivation{}, nil
}
func (f fakeVersionProvider) ListReleasedVersions(toolName provider.ToolID) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.versions[toolName], nil
}

func TestListToolVersions(t *testing.T) {
	t.Run("returns sorted versions from provider", func(t *testing.T) {
		fp := fakeVersionProvider{
			versions: map[provider.ToolID][]string{
				"nodejs": {"1.0.0", "3.0.0", "2.0.0"},
			},
		}

		versions, err := ListToolVersions("nodejs", "", fp)
		require.NoError(t, err)
		assert.Equal(t, []string{"3.0.0", "2.0.0", "1.0.0"}, versions)
	})

	t.Run("returns error from provider", func(t *testing.T) {
		fp := fakeVersionProvider{
			err: fmt.Errorf("connection failed"),
		}

		_, err := ListToolVersions("nodejs", "", fp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection failed")
	})

	t.Run("returns empty list for tool with no versions", func(t *testing.T) {
		fp := fakeVersionProvider{
			versions: map[provider.ToolID][]string{
				"nodejs": {},
			},
		}

		versions, err := ListToolVersions("nodejs", "", fp)
		require.NoError(t, err)
		assert.Empty(t, versions)
	})

	t.Run("resolves alias to canonical name", func(t *testing.T) {
		fp := fakeVersionProvider{
			versions: map[provider.ToolID][]string{
				"golang": {"1.21.0", "1.22.0"},
			},
		}

		versions, err := ListToolVersions("go", "", fp)
		require.NoError(t, err)
		assert.Equal(t, []string{"1.22.0", "1.21.0"}, versions)
	})

	t.Run("resolves node alias to nodejs", func(t *testing.T) {
		fp := fakeVersionProvider{
			versions: map[provider.ToolID][]string{
				"nodejs": {"18.0.0", "20.0.0"},
			},
		}

		versions, err := ListToolVersions("node", "", fp)
		require.NoError(t, err)
		assert.Equal(t, []string{"20.0.0", "18.0.0"}, versions)
	})

	t.Run("rejects unsupported tool", func(t *testing.T) {
		fp := fakeVersionProvider{}

		_, err := ListToolVersions("nonexistent", "", fp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a supported tool")
	})

	t.Run("sorts pre-release versions after their release", func(t *testing.T) {
		fp := fakeVersionProvider{
			versions: map[provider.ToolID][]string{
				"nodejs": {"1.0.0", "2.0.0-rc.1", "2.0.0", "1.0.0-beta.1"},
			},
		}

		versions, err := ListToolVersions("nodejs", "", fp)
		require.NoError(t, err)
		assert.Equal(t, []string{"2.0.0", "2.0.0-rc.1", "1.0.0", "1.0.0-beta.1"}, versions)
	})

	t.Run("places non-semver versions after semver", func(t *testing.T) {
		fp := fakeVersionProvider{
			versions: map[provider.ToolID][]string{
				"nodejs": {"nightly", "2.0.0", "1.0.0", "3.15.0a8", "latest"},
			},
		}

		versions, err := ListToolVersions("nodejs", "", fp)
		require.NoError(t, err)
		assert.Equal(t, []string{"3.15.0a8", "2.0.0", "1.0.0", "nightly", "latest"}, versions)
	})

	t.Run("filters by version prefix", func(t *testing.T) {
		fp := fakeVersionProvider{
			versions: map[provider.ToolID][]string{
				"nodejs": {"18.0.0", "18.1.0", "20.0.0", "20.1.0", "22.0.0"},
			},
		}

		versions, err := ListToolVersions("nodejs", "20", fp)
		require.NoError(t, err)
		assert.Equal(t, []string{"20.1.0", "20.0.0"}, versions)
	})

	t.Run("version prefix matches at boundary only", func(t *testing.T) {
		fp := fakeVersionProvider{
			versions: map[provider.ToolID][]string{
				"nodejs": {"18.0.0", "18.1.0", "18.1.2", "18.10.0", "20.0.0"},
			},
		}

		versions, err := ListToolVersions("nodejs", "18.1", fp)
		require.NoError(t, err)
		assert.Equal(t, []string{"18.1.2", "18.1.0"}, versions)
	})

	t.Run("version prefix with trailing dot", func(t *testing.T) {
		fp := fakeVersionProvider{
			versions: map[provider.ToolID][]string{
				"nodejs": {"18.0.0", "18.1.0", "20.0.0", "20.1.0", "22.0.0"},
			},
		}

		versions, err := ListToolVersions("nodejs", "20.", fp)
		require.NoError(t, err)
		assert.Equal(t, []string{"20.1.0", "20.0.0"}, versions)
	})

	t.Run("version prefix matches nothing", func(t *testing.T) {
		fp := fakeVersionProvider{
			versions: map[provider.ToolID][]string{
				"nodejs": {"18.0.0", "20.0.0"},
			},
		}

		versions, err := ListToolVersions("nodejs", "22", fp)
		require.NoError(t, err)
		assert.Empty(t, versions)
	})
}
