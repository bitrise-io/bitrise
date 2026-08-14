package build

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v2/internal/build"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

const watchDivider = "─────────────────────────────────────────────────────────"

// printBuildText writes a full key/value dump of a build. Shared by `build
// view` and the --format json/yml final-record rendering in `runWatch`/
// `runWatchTUI` (the plain-text watch footer is a separate short line).
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

// printTriggerText is the human-format response for `build trigger` — a
// short success line plus the URL. The full build detail view would be
// misleading here because most fields aren't populated yet (the build
// hasn't run).
func printTriggerText(w io.Writer, b internalbuild.Build) error {
	s := style.New(w)
	var headline strings.Builder
	headline.WriteString(s.Success.Render("✓"))
	headline.WriteString(" ")
	headline.WriteString(s.Bold.Render("Build triggered"))
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
// service's Build.BuildURL isn't populated by the API (only the trigger
// response provides one; view/list/watch build the URL themselves instead).
func buildWebURL(webBaseURL, appSlug, buildSlug string) string {
	return fmt.Sprintf("%s/app/%s/build/%s", webBaseURL, appSlug, buildSlug)
}

// writeDetachNotice writes the standard Ctrl-C detach message to w. resumeCmd
// is the command (without the "bitrise " prefix) the user can run to resume.
// Shared by every wait/watch path so the wording stays consistent.
func writeDetachNotice(w io.Writer, resumeCmd string) error {
	_, err := fmt.Fprintf(w, "\nDetached — build is still running.\nUse 'bitrise %s' to resume.\n", resumeCmd)
	return err
}

// runWatch is the shared implementation for `build watch` and
// `build trigger --watch`. It prints a header/footer to stderr and streams
// log content to logWriter. In --format json/yml it renders the final build
// record to cmd.OutOrStdout() instead of the text footer, so stdout stays a
// clean single record.
func runWatch(cmd *cobra.Command, svc *internalbuild.Service, b internalbuild.Build, interval time.Duration, logWriter io.Writer, format string) error {
	if format == output.FormatRaw && cmdutil.IsTerminalWriter(cmd.OutOrStdout()) {
		return runWatchTUI(cmd, svc, b, interval)
	}

	stderr := cmd.ErrOrStderr()
	if _, err := fmt.Fprintf(stderr, "%s\n%s\n", buildWatchHeader(b), watchDivider); err != nil {
		return err
	}

	finalBuild, err := svc.Watch(cmd.Context(), b.AppSlug, b.Slug, logWriter, interval)
	if errors.Is(err, context.Canceled) {
		return writeDetachNotice(stderr, "build watch "+b.Slug)
	}
	if err != nil {
		return err
	}

	if format == output.FormatJSON || format == output.FormatYML {
		if err := output.Render(cmd.OutOrStdout(), format, finalBuild, printBuildText); err != nil {
			return err
		}
	} else {
		footer := fmt.Sprintf("\n%s\nBuild #%d finished: %s%s\n", watchDivider, finalBuild.BuildNumber, finalBuild.Status, buildElapsed(finalBuild))
		if url := buildDetailURL(cmd, b); url != "" {
			footer += fmt.Sprintf("%s\n", url)
		}
		if _, err := fmt.Fprint(stderr, footer); err != nil {
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
