package build

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v3/internal/build"
	"github.com/bitrise-io/bitrise/v3/internal/style"
	"github.com/bitrise-io/bitrise/v3/output"
)

// NewStepsCommand returns the `build steps` subcommand.
func NewStepsCommand() *cobra.Command {
	var status string

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

			res, err := internalbuild.NewService(client).Steps(cmd.Context(), appSlug, args[0], status)
			if err != nil {
				return err
			}

			if err := output.Render(cmd.OutOrStdout(), output.Format, res, func(w io.Writer, res internalbuild.StepsResult) error {
				return printStepsTable(w, res, status)
			}); err != nil {
				return err
			}
			if output.Format == output.FormatRaw && !res.Complete {
				_, err := fmt.Fprintln(cmd.ErrOrStderr(), "The step summary is not final yet: the build is still running, or has just finished and its log is being archived. Only the steps that have started so far are listed.")
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", fmt.Sprintf("only steps with this status: %s. failed matches every status that fails the build: %s",
		strings.Join(internalbuild.StepStatuses, ", "), strings.Join(internalbuild.FailedStepStatuses, ", ")))
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
func printStepsTable(w io.Writer, res internalbuild.StepsResult, status string) error {
	if len(res.Items) == 0 {
		msg := "No steps found."
		if status != "" {
			msg = fmt.Sprintf("No steps with status %q.", status)
		}
		_, err := fmt.Fprintln(w, msg)
		return err
	}

	s := style.New(w)
	headers := []string{"WORKFLOW", "STEP", "STATUS", "DURATION"}
	rows := make([][]string, 0, len(res.Items))
	for _, st := range res.Items {
		duration := ""
		// Skipped steps carry timestamps too, but a duration would suggest they ran.
		if st.StartedAt != nil && st.FinishedAt != nil && !strings.HasPrefix(st.Status, "skipped") {
			duration = "<1s"
			if d := st.FinishedAt.Sub(*st.StartedAt); d >= time.Second {
				duration = d.Round(time.Second).String()
			}
		}
		rows = append(rows, []string{st.Workflow, stepName(st), st.Status, duration})
	}
	const colStatus = 2
	styler := func(row, col int, content string) string {
		if col == colStatus {
			return s.StepStatus(res.Items[row].Status).Render(content)
		}
		return content
	}
	if err := style.Table(w, headers, rows, s.Header, styler); err != nil {
		return err
	}

	var b strings.Builder
	for _, st := range res.Items {
		var reasons []string
		// A plain skipped step's reason is always "a previous Step failed",
		// which the failed step listed before it already says.
		if st.StatusReason != "" && st.Status != "skipped" {
			reasons = append(reasons, strings.TrimSpace(st.StatusReason))
		}
		for _, e := range st.Errors {
			msg := strings.TrimSpace(e.Message)
			if msg == "" {
				msg = fmt.Sprintf("(no message, code %d)", e.Code)
			}
			reasons = append(reasons, msg)
		}
		if len(reasons) == 0 {
			continue
		}
		if b.Len() == 0 {
			fmt.Fprintf(&b, "\n%s\n", s.Bold.Render("Reasons:"))
		}
		fmt.Fprintf(&b, "  %s › %s\n", st.Workflow, s.StepStatus(st.Status).Render(stepName(st)))
		for _, r := range reasons {
			for _, line := range strings.Split(r, "\n") {
				fmt.Fprintf(&b, "    %s\n", line)
			}
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func stepName(st internalbuild.Step) string {
	if st.Title != "" {
		return st.Title
	}
	return st.StepID
}
