package app

import (
	"fmt"
	"io"
	"strings"

	"al.essio.dev/pkg/shellescape"
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalapp "github.com/bitrise-io/bitrise/v2/internal/app"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewListCommand returns the `app list` subcommand.
func NewListCommand() *cobra.Command {
	var (
		limit       int
		cursor      string
		sortBy      string
		title       string
		projectType string
		fetchAll    bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List apps the authenticated user can access",
		Long: `List all apps the authenticated user can access.

Filters:
  --title TITLE          filter apps by title
  --project-type TYPE    e.g. ios, android
  --sort-by FIELD        ordering accepted by the API (created_at, last_build_at)

Pagination:
  --limit N     max items per page (server default if 0)
  --cursor TOKEN opaque token from a previous page's next_cursor
  --all         fetch all pages automatically

In JSON mode (--format json), next_cursor holds the cursor value for scripting:
  bitrise app list --format json | jq -r '.next_cursor'`,
		Example: `  bitrise app list
  bitrise app list --all
  bitrise app list --format json | jq -r '.next_cursor'
  bitrise app list --project-type ios --limit 100`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			if fetchAll && cursor != "" {
				return fmt.Errorf("--all and --cursor cannot be used together")
			}

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}
			svc := internalapp.NewService(client)

			makeOpts := func(cur string) internalapp.ListOptions {
				return internalapp.ListOptions{
					Limit:       limit,
					Cursor:      cur,
					SortBy:      sortBy,
					Title:       title,
					ProjectType: projectType,
				}
			}

			var res internalapp.AppsResult
			if fetchAll {
				allItems := []internalapp.App{}
				seenCursors := map[string]bool{}
				cur := ""
				for {
					page, pageErr := svc.List(cmd.Context(), makeOpts(cur))
					if pageErr != nil {
						return pageErr
					}
					allItems = append(allItems, page.Items...)
					if page.NextCursor == "" {
						break
					}
					if seenCursors[page.NextCursor] {
						return fmt.Errorf("pagination stalled: the API returned the cursor %q twice", page.NextCursor)
					}
					seenCursors[page.NextCursor] = true
					cur = page.NextCursor
				}
				res = internalapp.AppsResult{Items: allItems}
			} else {
				res, err = svc.List(cmd.Context(), makeOpts(cursor))
				if err != nil {
					return err
				}
			}

			return output.Render(cmd.OutOrStdout(), output.Format, res, func(w io.Writer, res internalapp.AppsResult) error {
				return printAppsTable(w, res, nextPageCmd(cmd))
			})
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page (server default if 0)")
	cmd.Flags().StringVar(&cursor, "cursor", "", "pagination cursor from a previous response")
	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages automatically")
	cmd.Flags().StringVar(&sortBy, "sort-by", "", "ordering accepted by the API (created_at, last_build_at)")
	cmd.Flags().StringVar(&title, "title", "", "filter apps by title")
	cmd.Flags().StringVar(&projectType, "project-type", "", "filter by project type (ios, android, ...)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	_ = cmd.RegisterFlagCompletionFunc("sort-by", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"created_at", "last_build_at"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("project-type", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"ios", "android", "flutter", "react-native", "xamarin", "cordova", "ionic", "other"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

// nextPageCmd builds the "fetch the next page" hint shown under the table,
// reproducing every filter flag the user actually set so the suggested
// invocation stays scoped the same way.
func nextPageCmd(cmd *cobra.Command) func(nextCursor string) string {
	return func(nextCursor string) string {
		parts := []string{"bitrise app list"}
		for _, name := range []string{"title", "project-type", "sort-by", "limit"} {
			if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
				parts = append(parts, "--"+name, shellescape.Quote(f.Value.String()))
			}
		}
		parts = append(parts, "--cursor", shellescape.Quote(nextCursor))
		return strings.Join(parts, " ")
	}
}

func printAppsTable(w io.Writer, res internalapp.AppsResult, nextPageCmd func(string) string) error {
	if len(res.Items) == 0 {
		_, err := fmt.Fprintln(w, "No apps found.")
		return err
	}

	s := style.New(w)
	headers := []string{"TITLE", "PROVIDER", "PROJECT_TYPE", "WORKSPACE", "DISABLED", "ID"}
	rows := make([][]string, 0, len(res.Items))
	disabled := make([]bool, 0, len(res.Items))
	for _, a := range res.Items {
		dis := ""
		if a.IsDisabled {
			dis = "yes"
		}
		disabled = append(disabled, a.IsDisabled)
		rows = append(rows, []string{a.Title, a.Provider, a.ProjectType, a.OwnerSlug, dis, a.Slug})
	}
	const colSlug = 5
	styler := func(row, col int, content string) string {
		if disabled[row] {
			// Whole row dimmed when the app is disabled.
			return s.Dim.Render(content)
		}
		if col == colSlug {
			return s.Slug.Render(content)
		}
		return content
	}
	if err := style.Table(w, headers, rows, s.Header, styler); err != nil {
		return err
	}
	if res.NextCursor != "" {
		hint := fmt.Sprintf("More results available. To fetch the next page:\n  %s", nextPageCmd(res.NextCursor))
		_, err := fmt.Fprintf(w, "\n%s\n", s.Dim.Render(hint))
		return err
	}
	return nil
}
