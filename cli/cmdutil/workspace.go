package cmdutil

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil/picker"
	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/internal/workspace"
	"github.com/bitrise-io/bitrise/v2/output"
)

// FlagWorkspace is the workspace a command acts on.
const FlagWorkspace = "workspace"

// EnvWorkspaceID supplies the workspace when --workspace isn't passed. Bitrise
// auto-injects it into every build, naming the workspace that owns the app the
// build runs for, so a bare command inside a build acts on that workspace. Same
// intentional ambient targeting as EnvAppIDLegacy. Defined in internal/workspace
// because the shared sole-workspace error names it.
const EnvWorkspaceID = workspace.EnvWorkspaceID

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

// ResolveWorkspaceID returns the workspace ID for this command: --workspace,
// then DefaultWorkspaceSlug (BITRISE_WORKSPACE_ID / default_workspace_id),
// then one GET /organizations call — a sole workspace auto-detects (with a
// stderr breadcrumb), 2+ workspaces show an interactive picker on a TTY or a
// friendly sorted error otherwise. Zero workspaces always errors.
func ResolveWorkspaceID(cmd *cobra.Command) (string, error) {
	if v, _ := cmd.Flags().GetString(FlagWorkspace); v != "" {
		return v, nil
	}
	if v := DefaultWorkspaceSlug(cmd); v != "" {
		return v, nil
	}
	client, err := NewAPIClient(cmd)
	if err != nil {
		return "", err
	}
	orgs, err := client.Organizations(cmd.Context())
	if err != nil {
		return "", fmt.Errorf("list workspaces: %w", err)
	}
	if len(orgs) > 1 && interactiveWorkspacePicker(cmd) {
		return pickWorkspace(cmd, orgs)
	}
	ws, err := workspace.Sole(orgs)
	if err != nil {
		return "", err
	}
	if !IsQuiet(cmd) && output.Format == output.FormatRaw {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Using your only workspace: %s\n", workspaceLabel(ws))
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Set it permanently to skip this lookup: bitrise config set %s %s\n", config.KeyDefaultWorkspaceID, ws.Slug)
	}
	return ws.Slug, nil
}

// interactiveWorkspacePicker reports whether the workspace picker can run: it
// reads keys from stdin and draws to stderr, so both must be a terminal, and it
// only makes sense for raw/human output (json/yml callers want the
// deterministic error, not a prompt).
func interactiveWorkspacePicker(cmd *cobra.Command) bool {
	return output.Format == output.FormatRaw &&
		IsTerminal(os.Stdin) &&
		IsTerminalWriter(cmd.ErrOrStderr())
}

// pickWorkspace shows the interactive workspace picker for orgs (assumed 2+)
// and returns the chosen workspace ID for this command only. It prints the
// command that pins the choice as the default so the next run skips the
// prompt. A backout (Esc/q/Ctrl-C) prints "Cancelled." and returns an error
// that aborts the command without cobra's redundant "Error:" line.
func pickWorkspace(cmd *cobra.Command, orgs []bitriseapi.Organization) (string, error) {
	sorted := workspace.Sort(orgs)
	items := make([]picker.Item, len(sorted))
	for i, o := range sorted {
		items[i] = picker.Item{Title: workspaceName(o), Desc: o.Slug}
	}
	idx, err := picker.Select(cmd.Context(), picker.Config{
		Prompt:     "Select a workspace",
		Items:      items,
		Cursor:     0,
		DefaultIdx: -1,
		In:         os.Stdin,
		Out:        cmd.ErrOrStderr(),
	})
	if errors.Is(err, picker.ErrCancelled) {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
		SilenceRootErrors(cmd)
		return "", errors.New("workspace selection cancelled")
	}
	if err != nil {
		return "", err
	}
	ws := sorted[idx]
	if !IsQuiet(cmd) {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Using workspace: %s\n", workspaceLabel(ws))
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Set it permanently to skip this prompt: bitrise config set %s %s\n", config.KeyDefaultWorkspaceID, ws.Slug)
	}
	return ws.Slug, nil
}

// workspaceLabel renders a workspace for a human breadcrumb as "name (slug)",
// falling back to the bare slug when the API omitted a name.
func workspaceLabel(ws bitriseapi.Organization) string {
	if ws.Name != "" {
		return fmt.Sprintf("%s (%s)", ws.Name, ws.Slug)
	}
	return ws.Slug
}

// workspaceName returns the workspace's display name, falling back to its
// slug when the API omitted a name — used as the picker row's primary text.
func workspaceName(ws bitriseapi.Organization) string {
	if ws.Name != "" {
		return ws.Name
	}
	return ws.Slug
}
