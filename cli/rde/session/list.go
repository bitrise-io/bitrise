package session

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

// listResult wraps the slice so --format json emits {"items": [...]}
// instead of a bare array — matches the other CLI commands' shape.
type listResult struct {
	Items []internalrde.Session `json:"items" yaml:"items"`
}

func newListCmd() *cobra.Command {
	var (
		selectors []string
		scope     string
		format    string
	)
	c := &cobra.Command{
		Use:   "list",
		Short: "List RDE sessions in the workspace",
		Long: `List RDE sessions in the workspace.

Filter by labels with --label-selector key=value (repeatable; selectors are
exact matches and are ANDed, at most 8 per request).

By default every session the authenticated user has in the workspace is
listed. Pass --scope workspace for sessions owned by the workspace itself
rather than by a user (for example sessions spawned by workspace device
preview links) — every workspace member sees the same list. The owning user
or workspace is reported via the owner_type and owner_id fields in
--format json/yml.

The session list comes from the backend in arbitrary order; the CLI does
not paginate (the API doesn't paginate this endpoint either).`,
		Example: `  bitrise rde session list
  bitrise rde session list --workspace my-workspace
  bitrise rde session list --scope workspace
  bitrise rde session list -l team=mobile -l branch=main
  bitrise rde session list --format json | jq '.items[].id'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			if err := validateLabelSelectors(selectors); err != nil {
				return err
			}
			switch scope {
			case internalrde.SessionScopeMine, internalrde.SessionScopeWorkspace:
			default:
				return fmt.Errorf("--scope must be %s or %s", internalrde.SessionScopeMine, internalrde.SessionScopeWorkspace)
			}
			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			sessions, err := internalrde.NewService(client).ListSessions(cmd.Context(), workspaceID, selectors, scope)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, listResult{Items: sessions}, renderSessionList)
		},
	}
	c.Flags().StringArrayVarP(&selectors, "label-selector", "l", nil, "only sessions whose labels match key=value exactly (repeatable; multiple selectors must all match)")
	c.Flags().StringVar(&scope, "scope", internalrde.SessionScopeMine, "which sessions to list: mine (sessions you created) or workspace (sessions owned by the workspace itself, visible to every member)")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	_ = c.RegisterFlagCompletionFunc("scope", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"mine\tsessions you created (default)", "workspace\tsessions owned by the workspace itself"}, cobra.ShellCompDirectiveNoFileComp
	})
	return c
}

func renderSessionList(w io.Writer, res listResult) error {
	if len(res.Items) == 0 {
		_, err := fmt.Fprintln(w, "No sessions found.")
		return err
	}
	s := style.New(w)
	headers := []string{"NAME", "STATUS", "TEMPLATE", "CREATED", "ID"}
	rows := make([][]string, 0, len(res.Items))
	statuses := make([]string, 0, len(res.Items))
	for _, sess := range res.Items {
		statuses = append(statuses, sess.Status)
		rows = append(rows, []string{
			sess.Name,
			sess.Status,
			sess.TemplateName,
			formatTime(sess.CreatedAt),
			sess.ID,
		})
	}
	const (
		colStatus = 1
		colID     = 4
	)
	styler := func(row, col int, content string) string {
		switch col {
		case colStatus:
			return statusStyle(s, statuses[row]).Render(content)
		case colID:
			return s.Slug.Render(content)
		}
		return content
	}
	return style.Table(w, headers, rows, s.Header, styler)
}
