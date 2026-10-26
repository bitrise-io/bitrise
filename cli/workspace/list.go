package workspace

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v3/internal/style"
	internalworkspace "github.com/bitrise-io/bitrise/v3/internal/workspace"
	"github.com/bitrise-io/bitrise/v3/output"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspaces the authenticated user belongs to",
		Example: `  bitrise workspace list
  bitrise workspace list --format json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}

			result, err := internalworkspace.NewService(client).List(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing workspaces failed: %w", err)
			}

			return output.Render(cmd.OutOrStdout(), output.Format, result, func(w io.Writer, result internalworkspace.WorkspacesResult) error {
				return printWorkspacesTable(w, result.Items)
			})
		},
	}

	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return cmd
}

func printWorkspacesTable(w io.Writer, workspaces []internalworkspace.Workspace) error {
	if len(workspaces) == 0 {
		_, err := fmt.Fprintln(w, "No workspaces found.")
		return err
	}
	s := style.New(w)
	headers := []string{"ID", "NAME"}
	rows := make([][]string, 0, len(workspaces))
	for _, ws := range workspaces {
		rows = append(rows, []string{ws.ID, ws.Name})
	}
	styler := func(_, col int, content string) string {
		if col == 0 {
			return s.Slug.Render(content)
		}
		return content
	}
	return style.Table(w, headers, rows, s.Header, styler)
}
