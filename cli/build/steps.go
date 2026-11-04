package build

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v3/internal/build"
	"github.com/bitrise-io/bitrise/v3/internal/style"
	"github.com/bitrise-io/bitrise/v3/output"
)

// NewStepsCommand returns the `build steps` subcommand.
func NewStepsCommand() *cobra.Command {
	var status, stepID string

	cmd := &cobra.Command{
		Use:   "steps BUILD_ID",
		Short: "List a build's steps with their status and failure or skip reasons",
		Long: `List the steps a build ran, across all its workflows, in execution order,
with each step's status and timing. Failed steps include the errors they
reported, and skipped steps the reason they were skipped.

While the build is still running, or its log is still being archived, only
the steps that have started so far are listed.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").`,
		Example: `  bitrise build steps abc123 --app my-app-id
  bitrise build steps abc123 --app my-app-id --status failed
  bitrise build steps abc123 --app my-app-id --step-id xcode-test
  bitrise build steps abc123 --app my-app-id --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}
			appSlug, err := cmdutil.ResolveAndLookupAppSlug(cmd, client)
			if err != nil {
				return err
			}

			res, err := internalbuild.NewService(client).Steps(cmd.Context(), appSlug, args[0], status, stepID)
			if err != nil {
				return err
			}

			if err := output.Render(cmd.OutOrStdout(), output.Format, res, func(w io.Writer, res internalbuild.StepsResult) error {
				return printStepsTable(w, res, status != "" || stepID != "")
			}); err != nil {
				return err
			}
			if output.Format == output.FormatRaw && !res.Complete && !cmdutil.IsQuiet(cmd) {
				_, err := fmt.Fprintln(cmd.ErrOrStderr(), "The step summary is not final yet: the build has not finished, or has just finished and its log is being archived. Only the steps that have started so far are listed.")
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", fmt.Sprintf("only steps with this status: %s. failed matches every status that fails the build: %s",
		strings.Join(internalbuild.StepStatuses, ", "), strings.Join(internalbuild.FailedStepStatuses, ", ")))
	cmd.Flags().StringVar(&stepID, "step-id", "", "only the runs of this step, for example xcode-test")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	_ = cmd.RegisterFlagCompletionFunc("status", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return internalbuild.StepStatuses, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

// printStepsTable renders the steps as a table, then the failure and skip
// reasons of every step that has any. Reasons can span lines, which a table
// cell would mangle.
func printStepsTable(w io.Writer, res internalbuild.StepsResult, filtered bool) error {
	// Before any step has started, the notice about the summary not being
	// final is all there is to say.
	if len(res.Items) == 0 && !res.Complete {
		return nil
	}
	if len(res.Items) == 0 {
		msg := "No steps found."
		if filtered {
			msg = "No steps match the filters."
		}
		_, err := fmt.Fprintln(w, msg)
		return err
	}

	s := style.New(w)
	headers := []string{"WORKFLOW", "STEP", "STATUS", "DURATION"}
	rows := make([][]string, 0, len(res.Items))
	for _, st := range res.Items {
		rows = append(rows, []string{st.Workflow, stepName(st), st.Status, stepDuration(st)})
	}
	const colStatus = 2
	styler := func(row, col int, content string) string {
		if col == colStatus {
			return stepStatusStyle(s, res.Items[row]).Render(content)
		}
		return content
	}
	if err := style.Table(w, headers, rows, s.Header, styler); err != nil {
		return err
	}

	var b strings.Builder
	// A filter can leave out the failed step a skipped step refers to.
	failureListed := false
	for _, st := range res.Items {
		var reasons []string
		if st.StatusReason != "" && (!st.SkippedAfterFailure() || !failureListed) {
			reasons = append(reasons, strings.TrimSpace(st.StatusReason))
		}
		for _, e := range st.Errors {
			msg := strings.TrimSpace(e.Message)
			if msg == "" {
				msg = fmt.Sprintf("(no message, code %d)", e.Code)
			}
			reasons = append(reasons, msg)
		}
		failureListed = failureListed || st.FailedBuild()
		if len(reasons) == 0 {
			continue
		}
		if b.Len() == 0 {
			fmt.Fprintf(&b, "\n%s\n", s.Bold.Render("Reasons:"))
		}
		fmt.Fprintf(&b, "  %s › %s\n", st.Workflow, stepStatusStyle(s, st).Render(stepName(st)))
		for _, r := range reasons {
			for _, line := range strings.Split(r, "\n") {
				fmt.Fprintf(&b, "    %s\n", line)
			}
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// stepStatusStyle colors a step like the build outcome it stands for: a step
// that fails the build as a failed build, failed_skippable (the build went on)
// and aborted (the build itself was aborted) as an aborted build, and started
// as a running one. The skipped ones and not_started are Dim.
func stepStatusStyle(s style.Styles, st internalbuild.Step) lipgloss.Style {
	switch {
	case st.Status == "started":
		return s.BuildStatus("in-progress")
	case st.Status == "success":
		return s.BuildStatus("success")
	case st.FailedBuild():
		return s.BuildStatus("failed")
	case st.Status == "failed_skippable", st.Status == "aborted":
		return s.BuildStatus("aborted")
	default:
		return s.Dim
	}
}

func stepName(st internalbuild.Step) string {
	if st.Title != "" {
		return st.Title
	}
	return st.StepID
}

func stepDuration(st internalbuild.Step) string {
	d, ok := st.RunTime()
	if !ok {
		return ""
	}
	if d >= time.Second {
		return d.Round(time.Second).String()
	}
	return "<1s"
}
