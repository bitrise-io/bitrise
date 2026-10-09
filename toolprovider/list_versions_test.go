package toolprovider

import (
	"fmt"
	"testing"

	"github.com/bitrise-io/bitrise/v3/toolprovider/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeVersionProvider struct {
	versions        []string
	err             error
	requestedTool   *provider.ToolID
	requestedPrefix *string
}

func (f fakeVersionProvider) ID() string       { return "fake" }
func (f fakeVersionProvider) Bootstrap() error { return nil }
func (f fakeVersionProvider) InstallTool(provider.ToolRequest) (provider.ToolInstallResult, error) {
	return provider.ToolInstallResult{}, nil
}
func (f fakeVersionProvider) ActivateEnv(provider.ToolInstallResult) (provider.EnvironmentActivation, error) {
	return provider.EnvironmentActivation{}, nil
}
func (f fakeVersionProvider) ListReleasedVersions(toolName provider.ToolID, prefix string) ([]string, error) {
	*f.requestedTool = toolName
	*f.requestedPrefix = prefix
	return f.versions, f.err
}

func TestListToolVersions(t *testing.T) {
	t.Run("returns the provider's listing for the canonical tool name", func(t *testing.T) {
		fp := newFakeVersionProvider([]string{"1.22.0", "1.21.0"}, nil)

		versions, err := ListToolVersions("go", "1.", fp)
		require.NoError(t, err)
		assert.Equal(t, []string{"1.22.0", "1.21.0"}, versions)
		assert.Equal(t, provider.ToolID("golang"), *fp.requestedTool)
		assert.Equal(t, "1.", *fp.requestedPrefix, "each provider reads the prefix its own way")
	})

	t.Run("returns error from provider", func(t *testing.T) {
		fp := newFakeVersionProvider(nil, fmt.Errorf("connection failed"))

		_, err := ListToolVersions("nodejs", "", fp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection failed")
	})

	t.Run("rejects unsupported tool", func(t *testing.T) {
		fp := newFakeVersionProvider(nil, nil)

		_, err := ListToolVersions("nonexistent", "", fp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a supported tool")
	})
}

func newFakeVersionProvider(versions []string, err error) fakeVersionProvider {
	return fakeVersionProvider{
		versions:        versions,
		err:             err,
		requestedTool:   new(provider.ToolID),
		requestedPrefix: new(string),
	}
}
