package build

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v3/internal/build"
	"github.com/bitrise-io/bitrise/v3/internal/style"
	"github.com/bitrise-io/bitrise/v3/output"
)

const watchDivider = "─────────────────────────────────────────────────────────"

// printBuildText writes a full key/value dump of a build. Shared by `build
// view` and the --format json/yml final-record rendering in
// `finishFollowedBuild` (the plain-text footer is a separate short line).
func printBuildText(w io.Writer, b internalbuild.Build) error {
	var buf strings.Builder
	fmt.Fprintf(&buf, "Build #%d\n", b.BuildNumber)
	fmt.Fprintf(&buf, "App:          %s\n", b.AppSlug)
	fmt.Fprintf(&buf, "Status:       %s\n", b.Status)
	if b.StatusText != "" {
		fmt.Fprintf(&buf, "Status Text:  %s\n", b.StatusText)
	}
	if b.AbortReason != "" {
		fmt.Fprintf(&buf, "Abort Reason: %s\n", b.AbortReason)
	}
	if b.IsOnHold {
		fmt.Fprintln(&buf, "On Hold:      yes")
	}
	if b.Rebuildable {
		fmt.Fprintln(&buf, "Rebuildable:  yes")
	}
	if b.Workflow != "" {
		fmt.Fprintf(&buf, "Workflow:     %s\n", b.Workflow)
	}
	if b.PipelineWorkflowID != "" {
		fmt.Fprintf(&buf, "Pipeline WF:  %s\n", b.PipelineWorkflowID)
	}
	if b.Branch != "" {
		fmt.Fprintf(&buf, "Branch:       %s\n", b.Branch)
	}
	if b.Tag != "" {
		fmt.Fprintf(&buf, "Tag:          %s\n", b.Tag)
	}
	if b.PullRequestID != 0 {
		fmt.Fprintf(&buf, "Pull Request: #%d", b.PullRequestID)
		if b.PullRequestTargetBranch != "" {
			fmt.Fprintf(&buf, " -> %s", b.PullRequestTargetBranch)
		}
		fmt.Fprintln(&buf)
		if b.PullRequestViewURL != "" {
			fmt.Fprintf(&buf, "PR URL:       %s\n", b.PullRequestViewURL)
		}
	}
	if b.CommitHash != "" {
		fmt.Fprintf(&buf, "Commit:       %s\n", b.CommitHash)
	}
	if b.CommitMessage != "" {
		fmt.Fprintf(&buf, "Message:      %s\n", b.CommitMessage)
	}
	if !b.TriggeredAt.IsZero() {
		fmt.Fprintf(&buf, "Triggered:    %s\n", b.TriggeredAt.Local().Format(time.RFC3339))
	}
	if b.TriggeredBy != "" {
		fmt.Fprintf(&buf, "Triggered By: %s\n", b.TriggeredBy)
	}
	if b.FinishedAt != nil {
		fmt.Fprintf(&buf, "Finished:     %s\n", b.FinishedAt.Local().Format(time.RFC3339))
	}
	if b.StackIdentifier != "" {
		fmt.Fprintf(&buf, "Stack:        %s\n", b.StackIdentifier)
	}
	if b.MachineTypeID != "" {
		fmt.Fprintf(&buf, "Machine Type: %s\n", b.MachineTypeID)
	}
	if b.CreditCost != 0 {
		fmt.Fprintf(&buf, "Credit Cost:  %d\n", b.CreditCost)
	}
	if b.BuildURL != "" {
		fmt.Fprintf(&buf, "URL:          %s\n", b.BuildURL)
	}
	_, err := io.WriteString(w, buf.String())
	return err
}

// printStartedText is the human-format response for `build trigger` and
// `build rebuild` — a short success line plus the URL. The full build detail
// view would be misleading here because most fields aren't populated yet (the
// build hasn't run).
func printStartedText(w io.Writer, b internalbuild.Build, title string) error {
	s := style.New(w)
	var headline strings.Builder
	headline.WriteString(s.Success.Render("✓"))
	headline.WriteString(" ")
	headline.WriteString(s.Bold.Render(title))
	if b.BuildNumber > 0 {
		headline.WriteString(s.Dim.Render(fmt.Sprintf("  #%d", b.BuildNumber)))
	}
	if b.Workflow != "" {
		headline.WriteString("  ")
		headline.WriteString(b.Workflow)
	}
	switch {
	case b.Branch != "":
		headline.WriteString(s.Dim.Render(" on "))
		headline.WriteString(b.Branch)
	case b.Tag != "":
		headline.WriteString(s.Dim.Render(" tag "))
		headline.WriteString(b.Tag)
	}
	if _, err := fmt.Fprintf(w, "%s\n", headline.String()); err != nil {
		return err
	}
	if b.BuildURL != "" {
		_, err := fmt.Fprintf(w, "  %s\n", s.URL.Render(b.BuildURL))
		return err
	}
	return nil
}

// printAbortText writes the short confirmation line for `build abort`.
func printAbortText(w io.Writer, r internalbuild.AbortResult) error {
	_, err := fmt.Fprintf(w, "Build aborted: %s\n", r.BuildSlug)
	return err
}

// buildWebURL constructs the web URL for a build, for commands where the
// service's Build.BuildURL isn't populated by the API (only the trigger and
// rebuild responses provide one; view/list/watch build the URL themselves
// instead).
func buildWebURL(webBaseURL, appSlug, buildSlug string) string {
	return fmt.Sprintf("%s/app/%s/build/%s", webBaseURL, url.PathEscape(appSlug), url.PathEscape(buildSlug))
}

// runAfterTrigger finishes `build trigger` and `build rebuild` once the new
// build exists: it prints the build, or with --watch/--wait follows it to the
// end and turns its outcome into the exit code.
func runAfterTrigger(cmd *cobra.Command, svc *internalbuild.Service, b internalbuild.Build, title string, wait, watch bool, interval time.Duration) error {
	if !wait && !watch {
		return output.Render(cmd.OutOrStdout(), output.Format, b, func(w io.Writer, b internalbuild.Build) error {
			return printStartedText(w, b, title)
		})
	}

	if watch {
		return runWatch(cmd, svc, b, interval)
	}

	// --wait: silent block until the build finishes; no log output.
	finalBuild, err := waitForBuild(cmd, svc, b, interval, "build watch "+b.Slug)
	if err != nil {
		return err
	}

	// The URL was already printed in the header, so the footer omits it.
	return finishFollowedBuild(cmd, finalBuild, "", "")
}

// addFollowFlags registers the --wait, --watch and --interval flags that
// runAfterTrigger serves.
func addFollowFlags(cmd *cobra.Command, wait, watch *bool, interval *time.Duration) {
	cmd.Flags().BoolVar(wait, "wait", false, "block until the build finishes without streaming logs (exit code reflects build outcome)")
	cmd.Flags().BoolVar(watch, "watch", false, "wait for the build to finish, showing progress (exit code reflects build outcome)")
	cmd.Flags().DurationVar(interval, "interval", 3*time.Second, "polling interval when --wait or --watch is active")
	cmd.MarkFlagsMutuallyExclusive("wait", "watch")
}

// runWatch is the shared implementation for `build watch` and the --watch
// flag of `build trigger` and `build rebuild`. It prints a header/footer to
// stderr and streams the log to stdout. In --format json/yml the log goes to
// stderr and the final build record to stdout instead of the text footer, so
// stdout stays a clean single record.
func runWatch(cmd *cobra.Command, svc *internalbuild.Service, b internalbuild.Build, interval time.Duration) error {
	logWriter := cmd.OutOrStdout()
	switch {
	case output.Format == output.FormatJSON || output.Format == output.FormatYML:
		logWriter = cmd.ErrOrStderr()
	case cmdutil.IsTerminalWriter(cmd.OutOrStdout()):
		return runWatchTUI(cmd, svc, b, interval)
	}

	stderr := cmd.ErrOrStderr()
	if _, err := fmt.Fprintf(stderr, "%s\n%s\n", buildWatchHeader(b), watchDivider); err != nil {
		return err
	}

	// Go's default SIGINT handling would kill the process before the detach
	// path runs. The TUI needs no handler: bubbletea handles Ctrl-C itself.
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()
	finalBuild, err := svc.Watch(ctx, b.AppSlug, b.Slug, logWriter, interval)
	if errors.Is(err, context.Canceled) {
		return detach(stderr, "build watch "+b.Slug, b.AppSlug)
	}
	if err != nil {
		return err
	}

	return finishFollowedBuild(cmd, finalBuild, "\n"+watchDivider+"\n", buildDetailURL(cmd, b))
}

// waitForBuild prints the --wait header in raw format and blocks until the build finishes.
// The SIGINT handler lives only as long as the wait, so a step after it (the
// log download of `build log --wait`) can still be interrupted.
func waitForBuild(cmd *cobra.Command, svc *internalbuild.Service, b internalbuild.Build, interval time.Duration, resumeCmd string) (internalbuild.Build, error) {
	if output.Format == output.FormatRaw {
		header := fmt.Sprintf("Waiting for build #%d to finish\n", b.BuildNumber)
		if url := buildDetailURL(cmd, b); url != "" {
			header += fmt.Sprintf("%s\n", url)
		}
		if _, err := fmt.Fprint(cmd.ErrOrStderr(), header); err != nil {
			return internalbuild.Build{}, err
		}
	}

	// Go's default SIGINT handling would kill the process before the detach
	// path runs.
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()
	finalBuild, err := svc.WaitForCompletion(ctx, b.AppSlug, b.Slug, interval)
	if errors.Is(err, context.Canceled) {
		return internalbuild.Build{}, detach(cmd.ErrOrStderr(), resumeCmd, b.AppSlug)
	}
	return finalBuild, err
}

// finishFollowedBuild reports the final state of a build followed by
// --wait/--watch: the build record on stdout in --format json/yml, otherwise a
// footer on stderr, made of prefix, the finished line and url when non-empty.
func finishFollowedBuild(cmd *cobra.Command, finalBuild internalbuild.Build, prefix, url string) error {
	if output.Format == output.FormatJSON || output.Format == output.FormatYML {
		if err := output.Render(cmd.OutOrStdout(), output.Format, finalBuild, printBuildText); err != nil {
			return err
		}
	} else {
		footer := fmt.Sprintf("%sBuild #%d finished: %s%s\n", prefix, finalBuild.BuildNumber, finalBuild.Status, buildElapsed(finalBuild))
		if url != "" {
			footer += url + "\n"
		}
		if _, err := fmt.Fprint(cmd.ErrOrStderr(), footer); err != nil {
			return err
		}
	}

	// The exit code reflects the build outcome in every mode, including
	// --format json/yml: stdout already carries the build record above.
	if finalBuild.Status != "success" && finalBuild.Status != "aborted-with-success" {
		return fmt.Errorf("build %s", finalBuild.Status)
	}
	return nil
}

// detach ends a wait/watch interrupted by Ctrl-C. It returns an error so the
// exit code is non-zero: the outcome was not reported, and a chained command
// must not run as if the build had succeeded. resumeCmd is the command
// (without the "bitrise " prefix) the user can run to resume. The app is
// always spelled out because the user may have passed it with --app, which
// the resumed command would not inherit.
func detach(stderr io.Writer, resumeCmd, appSlug string) error {
	// Move off the line the terminal echoed "^C" on.
	if _, err := fmt.Fprintln(stderr); err != nil {
		return err
	}
	// No claim about the build's state: Watch may have been interrupted while
	// streaming the log of a build that had already finished.
	return fmt.Errorf("detached: use 'bitrise %s --app %s' to resume", resumeCmd, appSlug)
}

func buildWatchHeader(b internalbuild.Build) string {
	s := fmt.Sprintf("Watching build #%d", b.BuildNumber)
	if b.Workflow != "" {
		s += fmt.Sprintf(" — workflow '%s'", b.Workflow)
	}
	if b.Branch != "" {
		s += fmt.Sprintf(" on branch '%s'", b.Branch)
	} else if b.Tag != "" {
		s += fmt.Sprintf(" on tag '%s'", b.Tag)
	}
	if b.BuildURL != "" {
		s += fmt.Sprintf("\n%s", b.BuildURL)
	}
	return s
}

func buildElapsed(b internalbuild.Build) string {
	if b.FinishedAt == nil || b.TriggeredAt.IsZero() {
		return ""
	}
	d := b.FinishedAt.Sub(b.TriggeredAt).Round(time.Second)
	return fmt.Sprintf(" (%s)", d)
}

// buildDetailURL returns the web URL of the build's detail page. It prefers
// the URL the API supplied (set on triggered builds) and falls back to
// constructing one from the resolved web base URL when the record doesn't
// carry it — e.g. the View path used by `build watch`.
func buildDetailURL(cmd *cobra.Command, b internalbuild.Build) string {
	if b.BuildURL != "" {
		return b.BuildURL
	}
	if b.AppSlug == "" || b.Slug == "" {
		return ""
	}
	return buildWebURL(cmdutil.ResolveWebBaseURL(cmd), b.AppSlug, b.Slug)
}
