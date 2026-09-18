package template

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

type listResult struct {
	Items []internalrde.Template `json:"items" yaml:"items"`
}

func newListCmd() *cobra.Command {
	var format string
	c := &cobra.Command{
		Use:   "list",
		Short: "List RDE templates in the workspace",
		Example: `  bitrise rde template list
  bitrise rde template list --format json | jq '.items[].id'`,
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
			items, err := internalrde.NewService(client).ListTemplates(cmd.Context(), workspaceID)
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
		_, err := fmt.Fprintln(w, "No templates found.")
		return err
	}
	s := style.New(w)
	headers := []string{"NAME", "STACK", "MACHINE", "OWNER", "ID"}
	rows := make([][]string, 0, len(res.Items))
	for _, t := range res.Items {
		rows = append(rows, []string{t.Name, t.StackID, t.MachineType, t.CreatedByEmail, t.ID})
	}
	const colID = 4
	styler := func(_, col int, content string) string {
		if col == colID {
			return s.Slug.Render(content)
		}
		return content
	}
	return style.Table(w, headers, rows, s.Header, styler)
}
