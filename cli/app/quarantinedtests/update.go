package quarantinedtests

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalapp "github.com/bitrise-io/bitrise/v3/internal/app"
	"github.com/bitrise-io/bitrise/v3/output"
)

// NewUpdateCommand returns the `app quarantined-tests update` subcommand.
func NewUpdateCommand() *cobra.Command {
	var (
		comment                 string
		branches, targets       []string
		allBranches, allTargets bool
	)

	cmd := &cobra.Command{
		Use:   "update ID",
		Short: "Change the comment or scope of a quarantined test",
		Long: `Change the comment or scope of a quarantined test. Only what you pass changes:
--branch replaces the branches but keeps the targets and the comment.

--all-branches and --all-targets remove the branch or target limit, so the
quarantine applies to every branch or every Workflow and Pipeline.`,
		Example: `  bitrise app quarantined-tests update 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id --comment "Tracked in JIRA-123"
  bitrise app quarantined-tests update 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id --branch main --branch "release/*"
  bitrise app quarantined-tests update 0b8a9c34-5d6e-4f70-8192-a3b4c5d6e7f8 --app my-app-id --all-branches --all-targets`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			var opts internalapp.UpdateQuarantinedTestOptions
			flags := cmd.Flags()
			if flags.Changed("comment") {
				opts.Comment = &comment
			}
			if flags.Changed("branch") {
				opts.Branches = &branches
			} else if allBranches {
				opts.Branches = new([]string)
			}
			if flags.Changed("target") {
				opts.Targets = &targets
			} else if allTargets {
				opts.Targets = new([]string)
			}
			if opts.Comment == nil && opts.Branches == nil && opts.Targets == nil {
				return errors.New("nothing to update: pass --comment, --branch, --target, --all-branches or --all-targets")
			}

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

			t, err := internalapp.NewService(client).UpdateQuarantinedTest(cmd.Context(), appSlug, args[0], opts)
			if err != nil {
				return fmt.Errorf("updating quarantined test failed: %w", err)
			}

			return output.Render(cmd.OutOrStdout(), output.Format, t, printQuarantinedTestText)
		},
	}

	cmd.Flags().StringVar(&comment, "comment", "", "why the test is quarantined; an empty value removes the comment")
	cmd.Flags().StringSliceVar(&branches, "branch", nil, "quarantine the test only on this branch, replacing the current branches (repeated or comma-separated)")
	cmd.Flags().StringSliceVar(&targets, "target", nil, "quarantine the test only in this Workflow or Pipeline, replacing the current targets (repeated or comma-separated)")
	cmd.Flags().BoolVar(&allBranches, "all-branches", false, "quarantine the test on every branch")
	cmd.Flags().BoolVar(&allTargets, "all-targets", false, "quarantine the test in every Workflow and Pipeline")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	cmd.MarkFlagsMutuallyExclusive("branch", "all-branches")
	cmd.MarkFlagsMutuallyExclusive("target", "all-targets")

	return cmd
}
