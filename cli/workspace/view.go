package workspace

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v3/internal/bitriseapi"
	internalworkspace "github.com/bitrise-io/bitrise/v3/internal/workspace"
	"github.com/bitrise-io/bitrise/v3/output"
)

func NewViewCommand() *cobra.Command {
	var web bool

	cmd := &cobra.Command{
		Use:   "view [WORKSPACE_ID]",
		Short: "Show details of a single workspace",
		Long: `Show details for a single workspace identified by its ID or name.

WORKSPACE_ID falls back to --workspace, then $BITRISE_WORKSPACE_ID, then the
default_workspace_id config key. With none of them set, your only workspace is
used, or you pick one interactively when you have several.`,
		Example: `  bitrise workspace view my-workspace-id
  bitrise workspace view "My Workspace"
  bitrise workspace view my-workspace-id --format json
  bitrise workspace view my-workspace-id --web`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(cmd, args, web, cmdutil.OpenBrowser)
		},
	}

	cmd.Flags().String(cmdutil.FlagWorkspace, "", "workspace ID or name to view (or set BITRISE_WORKSPACE_ID / default_workspace_id); overridden by the positional argument")
	cmd.Flags().BoolVar(&web, "web", false, "open the workspace page in the browser instead of printing")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}

// runView is NewViewCommand's RunE, with browser-opening injected so tests
// can exercise --web without launching a real browser.
func runView(cmd *cobra.Command, args []string, web bool, openBrowser func(string) error) error {
	cmdutil.LogCommandParameters(cmd)

	format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
	if err := output.ConfigureOutputFormat(format); err != nil {
		return fmt.Errorf("failed to configure output format: %w", err)
	}

	org, complete, err := resolveWorkspace(cmd, args)
	if err != nil {
		return err
	}

	if web {
		url := fmt.Sprintf("%s/workspaces/%s", cmdutil.ResolveWebBaseURL(cmd), org.Slug)
		if err := openBrowser(url); err != nil {
			return err
		}
		_, err := fmt.Fprintf(cmd.ErrOrStderr(), "Opened %s\n", url)
		return err
	}

	ws := internalworkspace.FromOrganization(org)
	if !complete {
		client, err := cmdutil.NewAPIClient(cmd)
		if err != nil {
			return err
		}
		ws, err = internalworkspace.NewService(client).View(cmd.Context(), org.Slug)
		if err != nil {
			return fmt.Errorf("viewing workspace failed: %w", err)
		}
	}

	return output.Render(cmd.OutOrStdout(), output.Format, ws, printWorkspaceText)
}

// resolveWorkspace resolves only a user-provided value (positional arg or
// --workspace) by name: an ambient value (env/config) is already a canonical
// slug, and ResolveWorkspaceID handles the sole-workspace and picker cases.
// complete reports that a name match already returned the whole workspace.
func resolveWorkspace(cmd *cobra.Command, args []string) (org bitriseapi.Organization, complete bool, err error) {
	value, _ := cmd.Flags().GetString(cmdutil.FlagWorkspace)
	if len(args) > 0 {
		value = args[0]
	}
	if value == "" {
		slug, err := cmdutil.ResolveWorkspaceID(cmd)
		return bitriseapi.Organization{Slug: slug}, false, err
	}
	client, err := cmdutil.NewAPIClient(cmd)
	if err != nil {
		return bitriseapi.Organization{}, false, err
	}
	return cmdutil.NewResolver(client).ResolveWorkspace(cmd.Context(), value)
}

func printWorkspaceText(w io.Writer, ws internalworkspace.Workspace) error {
	var b strings.Builder
	fmt.Fprintf(&b, "Name: %s\n", ws.Name)
	fmt.Fprintf(&b, "ID:   %s\n", ws.ID)
	_, err := io.WriteString(w, b.String())
	return err
}
