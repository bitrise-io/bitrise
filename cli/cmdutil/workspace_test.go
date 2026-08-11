package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/bitrise-io/bitrise/v2/internal/config"
)

// A bare command has a nil Context(), which must read as "no default set"
// rather than panicking.
func TestDefaultWorkspaceSlug_EmptyWhenNothingSet(t *testing.T) {
	t.Setenv(EnvWorkspaceID, "")

	assert.Empty(t, DefaultWorkspaceSlug(&cobra.Command{}))
}

func TestDefaultWorkspaceSlug_FromConfig(t *testing.T) {
	t.Setenv(EnvWorkspaceID, "")

	cmd := &cobra.Command{}
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolve(
		config.Config{}, config.Config{}, config.Config{DefaultWorkspaceID: "cfg-ws"},
	)))

	assert.Equal(t, "cfg-ws", DefaultWorkspaceSlug(cmd))
}

func TestDefaultWorkspaceSlug_EnvWinsOverConfig(t *testing.T) {
	t.Setenv(EnvWorkspaceID, "env-ws")

	cmd := &cobra.Command{}
	cmd.SetContext(config.WithResolved(t.Context(), config.Resolve(
		config.Config{}, config.Config{}, config.Config{DefaultWorkspaceID: "cfg-ws"},
	)))

	assert.Equal(t, "env-ws", DefaultWorkspaceSlug(cmd))
}
