package quarantinedtests

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalapp "github.com/bitrise-io/bitrise/v3/internal/app"
	"github.com/bitrise-io/bitrise/v3/output"
)

// NewAddCommand returns the `app quarantined-tests add` subcommand.
func NewAddCommand() *cobra.Command {
	var opts internalapp.AddQuarantinedTestOptions

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a test to quarantine",
		Long: `Quarantine a test, so it is excluded from the test runs of the app's builds.

A quarantine matches a test case by its exact suites, name and class name, or
by its identifier when the test report gives one. With --build, the test case
is looked up in that build's test results and quarantined with those exact
values, the same way the build page does it. Pass --name as
'bitrise build test-results BUILD_ID' lists it; if more than one test case
matches, --class-name, --suite and --identifier narrow it down.

Without --build, the values are sent as given: --suite takes the name of each
suite the test case belongs to, outermost first, as the test results list
them, which can differ from where the test case is nested.

The quarantine applies to every branch and every Workflow and Pipeline unless
--branch or --target narrows it. Both accept * wildcards.`,
		Example: `  bitrise app quarantined-tests add --app my-app-id --build abc123 --name LoginTests.testWrongPassword
  bitrise app quarantined-tests add --app my-app-id --build abc123 --name testWrongPassword --suite LoginTests --comment "Flaky on CI"
  bitrise app quarantined-tests add --app my-app-id --build abc123 --name testWrongPassword --branch main --branch "release/*" --target primary
  bitrise app quarantined-tests add --app my-app-id --suite AppTests --suite LoginTests --class-name LoginTests --name testWrongPassword
  bitrise app quarantined-tests add --app my-app-id --suite BullsEyeTests --name testScore --identifier "BullsEyeTests/testScore()"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			t, err := internalapp.NewService(client).AddQuarantinedTest(cmd.Context(), appSlug, opts)
			if err != nil {
				return fmt.Errorf("quarantining test failed: %w", err)
			}

			return output.Render(cmd.OutOrStdout(), output.Format, t, printQuarantinedTestText)
		},
	}

	cmd.Flags().StringVar(&opts.TestCaseName, "name", "", "name of the test case; with --build, also as Class.name")
	cmd.Flags().StringArrayVar(&opts.TestSuiteName, "suite", nil, "name of a suite the test case belongs to; repeat it for each, outermost first. With --build, it only narrows the lookup")
	cmd.Flags().StringVar(&opts.ClassName, "class-name", "", "class name of the test case. With --build, it only narrows the lookup")
	cmd.Flags().StringVar(&opts.TestCaseIdentifier, "identifier", "", "identifier of the test case, if the test report gives one. With --build, it only narrows the lookup")
	cmd.Flags().StringVar(&opts.BuildID, "build", "", "ID of a build the test ran in: the test case is looked up in its test results, and the build is recorded as where it was quarantined from")
	cmd.Flags().StringVar(&opts.Comment, "comment", "", "why the test is quarantined")
	cmd.Flags().StringSliceVar(&opts.Branches, "branch", nil, "quarantine the test only on this branch (repeated or comma-separated)")
	cmd.Flags().StringSliceVar(&opts.Targets, "target", nil, "quarantine the test only in this Workflow or Pipeline (repeated or comma-separated)")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
