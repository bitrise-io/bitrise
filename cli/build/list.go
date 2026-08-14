package build

import (
	"fmt"
	"io"
	"strings"
	"time"

	"al.essio.dev/pkg/shellescape"
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v2/internal/build"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewListCommand returns the `build list` subcommand.
func NewListCommand() *cobra.Command {
	var (
		branch           string
		workflow         string
		status           string
		sortBy           string
		commitMessage    string
		triggerEventType string
		pullRequestID    int
		buildNumber      int
		after            string
		before           string
		pipelineBuild    bool
		limit            int
		cursor           string
		fetchAll         bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List builds for an app",
		Long: `List builds for an app.

Pagination:
  --limit N     max items per page (server default if 0)
  --cursor TOKEN opaque token from a previous page's next_cursor
  --all         fetch all pages automatically

In JSON mode (--format json), next_cursor holds the cursor value for scripting:
  bitrise build list --app my-app-id --format json | jq -r '.next_cursor'`,
		Example: `  bitrise build list --app my-app-id
  bitrise build list --app my-app-id --branch main --status failed
  bitrise build list --app my-app-id --all`,
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

			appSlug, err := cmdutil.ResolveAppSlug(cmd)
			if err != nil {
				return err
			}

			var afterTime, beforeTime *time.Time
			if after != "" {
				t, err := time.Parse(time.RFC3339, after)
				if err != nil {
					return fmt.Errorf("invalid --after value: %w", err)
				}
				afterTime = &t
			}
			if before != "" {
				t, err := time.Parse(time.RFC3339, before)
				if err != nil {
					return fmt.Errorf("invalid --before value: %w", err)
				}
				beforeTime = &t
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}
			svc := internalbuild.NewService(client)

			var isPipelineBuild *bool
			if cmd.Flags().Changed("pipeline-build") {
				isPipelineBuild = &pipelineBuild
			}

			makeOpts := func(cur string) internalbuild.ListOptions {
				return internalbuild.ListOptions{
					AppSlug:          appSlug,
					Branch:           branch,
					Workflow:         workflow,
					Status:           status,
					SortBy:           sortBy,
					CommitMessage:    commitMessage,
					TriggerEventType: triggerEventType,
					PullRequestID:    pullRequestID,
					BuildNumber:      buildNumber,
					After:            afterTime,
					Before:           beforeTime,
					IsPipelineBuild:  isPipelineBuild,
					Limit:            limit,
					Cursor:           cur,
				}
			}

			var res internalbuild.ListResult
			if fetchAll {
				allItems := []internalbuild.Build{}
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
				res = internalbuild.ListResult{Items: allItems}
			} else {
				res, err = svc.List(cmd.Context(), makeOpts(cursor))
				if err != nil {
					return err
				}
			}

			return output.Render(cmd.OutOrStdout(), output.Format, res, func(w io.Writer, res internalbuild.ListResult) error {
				return printBuildsTable(w, res, nextPageCmd(cmd))
			})
		},
	}

	cmd.Flags().StringVar(&branch, "branch", "", "filter by branch")
	cmd.Flags().StringVar(&workflow, "workflow", "", "filter by workflow")
	cmd.Flags().StringVar(&status, "status", "", "filter by status: in-progress, success, failed, aborted, aborted-with-success")
	cmd.Flags().StringVar(&sortBy, "sort-by", "", "ordering: created_at (default) or running_first")
	cmd.Flags().StringVar(&commitMessage, "commit-message", "", "filter by commit message")
	cmd.Flags().StringVar(&triggerEventType, "trigger-event-type", "", "filter by trigger event type: push, pull-request, tag")
	cmd.Flags().IntVar(&pullRequestID, "pull-request-id", 0, "filter by pull request ID")
	cmd.Flags().IntVar(&buildNumber, "build-number", 0, "filter by build number")
	cmd.Flags().StringVar(&after, "after", "", "only builds triggered after this time (RFC3339)")
	cmd.Flags().StringVar(&before, "before", "", "only builds triggered before this time (RFC3339)")
	cmd.Flags().BoolVar(&pipelineBuild, "pipeline-build", false, "filter by whether the build is part of a pipeline")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page (server default if 0)")
	cmd.Flags().StringVar(&cursor, "cursor", "", "pagination cursor from a previous response")
	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages automatically")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	_ = cmd.RegisterFlagCompletionFunc("status", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"in-progress", "success", "failed", "aborted", "aborted-with-success"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("sort-by", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"created_at", "running_first"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("trigger-event-type", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"push", "pull-request", "tag"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

// nextPageCmd builds the "fetch the next page" hint shown under the table,
// reproducing every filter flag the user actually set so the suggested
// invocation stays scoped the same way.
func nextPageCmd(cmd *cobra.Command) func(nextCursor string) string {
	return func(nextCursor string) string {
		parts := []string{"bitrise build list"}
		for _, name := range []string{
			"app", "branch", "workflow", "status", "sort-by", "commit-message",
			"trigger-event-type", "pull-request-id", "build-number", "after",
			"before", "pipeline-build", "limit",
		} {
			if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
				parts = append(parts, "--"+name, shellescape.Quote(f.Value.String()))
			}
		}
		parts = append(parts, "--cursor", shellescape.Quote(nextCursor))
		return strings.Join(parts, " ")
	}
}

func printBuildsTable(w io.Writer, res internalbuild.ListResult, nextPageCmd func(string) string) error {
	if len(res.Items) == 0 {
		_, err := fmt.Fprintln(w, "No builds found.")
		return err
	}

	s := style.New(w)
	headers := []string{"NUMBER", "STATUS", "BRANCH", "WORKFLOW", "TRIGGERED", "ID"}
	rows := make([][]string, 0, len(res.Items))
	statuses := make([]string, 0, len(res.Items))
	onHold := make([]bool, 0, len(res.Items))
	for _, b := range res.Items {
		statuses = append(statuses, b.Status)
		onHold = append(onHold, b.IsOnHold)
		triggered := ""
		if !b.TriggeredAt.IsZero() {
			triggered = b.TriggeredAt.Local().Format("2006-01-02 15:04")
		}
		rows = append(rows, []string{fmt.Sprintf("#%d", b.BuildNumber), b.Status, b.Branch, b.Workflow, triggered, b.Slug})
	}
	const colStatus = 1
	const colSlug = 5
	styler := func(row, col int, content string) string {
		if col == colStatus {
			rendered := s.BuildStatus(statuses[row]).Render(content)
			if onHold[row] {
				rendered += " " + s.Dim.Render("(held)")
			}
			return rendered
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
