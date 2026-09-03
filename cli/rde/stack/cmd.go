package stack

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

type listResult struct {
	Items []internalrde.Stack `json:"items" yaml:"items"`
}

// NewCmd returns the `rde stack` parent command.
func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "stack",
		Short: "List machine stacks available to the workspace",
		Args:  cobra.NoArgs,
		RunE:  cmdutil.DelegateToList,
	}
	c.AddCommand(newListCmd())
	return c
}

func newListCmd() *cobra.Command {
	var format string
	c := &cobra.Command{
		Use:   "list",
		Short: "List machine stacks",
		Example: `  bitrise rde stack list
  bitrise rde stack list --format json | jq '.items[].id'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			items, err := internalrde.NewService(client).ListStacks(cmd.Context(), workspaceID)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, listResult{Items: items}, renderList)
		},
	}
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}

func renderList(w io.Writer, res listResult) error {
	if len(res.Items) == 0 {
		_, err := fmt.Fprintln(w, "No stacks found.")
		return err
	}
	s := style.New(w)
	headers := []string{"ID", "TITLE", "OS", "STATUS"}
	rows := make([][]string, 0, len(res.Items))
	for _, st := range res.Items {
		rows = append(rows, []string{st.ID, st.Title, osLabel(st), st.Status})
	}
	styler := func(_, col int, content string) string {
		if col == 0 {
			return s.Slug.Render(content)
		}
		return content
	}
	return style.Table(w, headers, rows, s.Header, styler)
}

// osLabel renders the OS column with a human-friendly OS name, appending the
// numeric OS version when present (e.g. "macOS 26"), so the table conveys both
// at a glance.
func osLabel(st internalrde.Stack) string {
	os := cmdutil.OSDisplayName(st.OS)
	if os == "" {
		return ""
	}
	if st.OSVersion > 0 {
		return os + " " + strconv.Itoa(int(st.OSVersion))
	}
	return os
}
