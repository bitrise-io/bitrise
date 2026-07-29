package stack

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalstack "github.com/bitrise-io/bitrise/v2/internal/stack"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

func NewListCommand() *cobra.Command {
	var workspaceSlug string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available stacks and their machine configurations",
		Long: `List all available stacks with their OS, status, and version information.

When --workspace is provided, returns stacks available for that workspace,
including any custom stacks configured for it.
Without --workspace, returns globally available stacks.`,
		Example: `  bitrise stack list
  bitrise stack list --workspace my-workspace-id
  bitrise stack list --format json`,
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

			result, err := internalstack.NewService(client).List(cmd.Context(), workspaceSlug)
			if err != nil {
				return fmt.Errorf("listing stacks failed: %w", err)
			}

			if output.Format == output.FormatRaw {
				return printStacksTable(cmd.OutOrStdout(), result.Items)
			}
			output.Print(result, output.Format)
			return nil
		},
	}

	cmd.Flags().StringVar(&workspaceSlug, "workspace", "", "workspace ID for workspace-specific stacks (including custom stacks)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}

// printStacksTable writes directly to w (rather than composing a string for
// log.Print, like other commands' table helpers) so style.New(w) sees the
// real output stream: TTY/NO_COLOR detection needs the actual writer, not a
// throwaway buffer, which would always look non-interactive and disable
// color. w is cmd.OutOrStdout() — os.Stdout in real runs, swappable in tests.
func printStacksTable(w io.Writer, stacks []internalstack.Stack) error {
	if len(stacks) == 0 {
		_, err := fmt.Fprintln(w, "No stacks found.")
		return err
	}

	s := style.New(w)
	headers := []string{"ID", "TITLE", "OS", "STATUS", "REMOVAL_DATE"}
	rows := make([][]string, 0, len(stacks))
	statuses := make([]string, 0, len(stacks))
	for _, st := range stacks {
		statuses = append(statuses, st.Status)
		rows = append(rows, []string{st.ID, st.Title, st.OS, st.Status, st.RemovalDate})
	}
	const colStatus = 3
	styler := func(row, col int, content string) string {
		if col == colStatus {
			switch statuses[row] {
			case "stable":
				return s.Success.Render(content)
			case "edge":
				return s.Warn.Render(content)
			case "frozen":
				return s.Dim.Render(content)
			}
		}
		if col == 0 {
			return s.Slug.Render(content)
		}
		return content
	}
	return style.Table(w, headers, rows, s.Header, styler)
}
