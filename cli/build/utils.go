package build

import (
	"fmt"
	"io"
	"strings"
	"time"

	internalbuild "github.com/bitrise-io/bitrise/v2/internal/build"
)

// printBuildText writes a full key/value dump of a build. Used by `build
// view`'s --format raw output.
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
