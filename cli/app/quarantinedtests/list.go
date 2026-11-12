package quarantinedtests

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v3/cli/cmdutil"
	internalapp "github.com/bitrise-io/bitrise/v3/internal/app"
	internalbuild "github.com/bitrise-io/bitrise/v3/internal/build"
	"github.com/bitrise-io/bitrise/v3/internal/style"
	"github.com/bitrise-io/bitrise/v3/output"
)

// NewListCommand returns the `app quarantined-tests list` subcommand.
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the tests quarantined on an app",
		Example: `  bitrise app quarantined-tests list --app my-app-id
  bitrise app quarantined-tests list --app my-app-id --format json`,
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

			res, err := internalapp.NewService(client).ListQuarantinedTests(cmd.Context(), appSlug)
			if err != nil {
				return fmt.Errorf("listing quarantined tests failed: %w", err)
			}

			return output.Render(cmd.OutOrStdout(), output.Format, res, func(w io.Writer, res internalapp.QuarantinedTestsResult) error {
				return printQuarantinedTestsTable(w, res.Items)
			})
		},
	}

	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}

func printQuarantinedTestsTable(w io.Writer, tests []internalapp.QuarantinedTest) error {
	if len(tests) == 0 {
		_, err := fmt.Fprintln(w, "No quarantined tests found.")
		return err
	}

	s := style.New(w)
	headers := []string{"ID", "TEST", "BRANCHES", "TARGETS", "QUARANTINED"}
	rows := make([][]string, 0, len(tests))
	for _, t := range tests {
		quarantined := t.QuarantinedAt.Local().Format("2006-01-02")
		if t.QuarantinedBy != "" {
			quarantined += " by " + t.QuarantinedBy
		}
		rows = append(rows, []string{t.ID, testPath(t), filterText(t.Filters.Branches), filterText(t.Filters.Targets), quarantined})
	}
	styler := func(_, col int, content string) string {
		if col == 0 {
			return s.Slug.Render(content)
		}
		return content
	}
	return style.Table(w, headers, rows, s.Header, styler)
}

// testPath identifies a test the way a test report nests it: suites
// outermost first, then the test case as `build test-results` names it.
func testPath(t internalapp.QuarantinedTest) string {
	return strings.Join(append(append([]string{}, t.TestSuiteName...), internalbuild.TestCaseDisplayName(t.ClassName, t.TestCaseName)), " › ")
}

func filterText(values []string) string {
	if len(values) == 0 {
		return "all"
	}
	return strings.Join(values, ", ")
}
