package cmdutil

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/internal/config"
)

// FlagWorkspace is the workspace a command acts on.
const FlagWorkspace = "workspace"

// EnvWorkspaceID supplies the workspace when --workspace isn't passed. Safe to
// honor unconditionally, unlike BITRISE_APP_SLUG (see EnvAppID): Bitrise does
// not inject a workspace ID into builds, so this can't make a bare command
// silently act on whatever workspace a build happens to run in.
const EnvWorkspaceID = "BITRISE_WORKSPACE_ID"

// DefaultWorkspaceSlug returns the workspace configured as the default —
// BITRISE_WORKSPACE_ID, then the default_workspace_id config key — or "" when
// neither is set. It deliberately does not read --workspace: callers own that
// flag's precedence, and they need to know whether a value came from the flag
// (explicit) or from here (implicit) to report it to the user.
func DefaultWorkspaceSlug(cmd *cobra.Command) string {
	if v := os.Getenv(EnvWorkspaceID); v != "" {
		return v
	}
	return config.FromContext(cmd.Context()).DefaultWorkspaceID
}
