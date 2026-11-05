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

// NewTestResultsCommand returns the `build test-results` subcommand.
func NewTestResultsCommand() *cobra.Command {
	var filter internalbuild.TestResultsFilter
	var flaky bool

	cmd := &cobra.Command{
		Use:   "test-results BUILD_ID",
		Short: "List a build's test suites and test cases with their results",
		Long: `List the test suites a build reported and the test cases in them, with each
case's status and duration, and the message of failed ones.

Each test case also lists the test suites it belongs to, which can differ from
where it is nested, and the identifier its test report gives it, if any.

A quarantine matches a test case by those exact suites, its name and class
name, or by its identifier. 'bitrise app quarantined-tests add --build
BUILD_ID --name NAME' looks them up in the build, with NAME as listed here.

Test results are available once the build has finished, and only if the plan
of the app's workspace includes test reports. Filters narrow the test cases
listed, and suites left without any are left out; a suite's counts always
cover all of its test cases. --sort orders the test cases within each suite,
and --limit keeps the first ones across suites.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").`,
		Example: `  bitrise build test-results abc123 --app my-app-id
  bitrise build test-results abc123 --app my-app-id --status failed,error
  bitrise build test-results abc123 --app my-app-id --status failed,error --sort duration --order desc --limit 20
  bitrise build test-results abc123 --app my-app-id --flaky
  bitrise build test-results abc123 --app my-app-id --suite LoginTests --name signin
  bitrise build test-results abc123 --app my-app-id --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)
			if cmd.Flags().Changed("flaky") {
				filter.Flaky = &flaky
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

			res, err := internalbuild.NewService(client).TestResults(cmd.Context(), appSlug, args[0], filter)
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), output.Format, res, func(w io.Writer, res internalbuild.TestResultsResult) error {
				return printTestResultsText(w, res, filter.IsSet(), appSlug, args[0])
			})
		},
	}

	cmd.Flags().StringSliceVar(&filter.Statuses, "status", nil, fmt.Sprintf("only test cases with one of these statuses (comma-separated or repeated): %s", strings.Join(internalbuild.TestStatuses, ", ")))
	cmd.Flags().BoolVar(&flaky, "flaky", false, "only flaky test cases, or with --flaky=false only the ones that are not flaky")
	cmd.Flags().StringVar(&filter.Suite, "suite", "", "only the test suite with this exact name, at any depth, with the suites nested in it")
	cmd.Flags().StringVar(&filter.ClassName, "class-name", "", "only test cases with this exact class name")
	cmd.Flags().StringVar(&filter.Name, "name", "", "only test cases whose name, as listed (Class.name), contains this text (case-insensitive)")
	cmd.Flags().StringVar(&filter.SortBy, "sort", "", fmt.Sprintf("sort the test cases within each suite by: %s. status puts error first, then failed, skipped and passed", strings.Join(internalbuild.TestSortFields, ", ")))
	cmd.Flags().StringVar(&filter.SortOrder, "order", "", fmt.Sprintf("direction of --sort: %s (default asc)", strings.Join(internalbuild.TestSortOrders, ", ")))
	cmd.Flags().IntVar(&filter.Limit, "limit", 0, fmt.Sprintf("list at most this many test cases, up to %d (default: all)", internalbuild.MaxTestResultsLimit))
	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	_ = cmd.RegisterFlagCompletionFunc("status", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return internalbuild.TestStatuses, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("sort", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return internalbuild.TestSortFields, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("order", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return internalbuild.TestSortOrders, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

// printTestResultsText renders the suites as an indented tree: each suite
// with its totals, then its test cases, then its child suites. Messages can
// span lines, so they go under their test case rather than in a column. A
// footer shows how to quarantine one of the test cases listed.
func printTestResultsText(w io.Writer, res internalbuild.TestResultsResult, filtered bool, appSlug, buildSlug string) error {
	if len(res.TestSuites) == 0 {
		msg := "No test results found."
		if filtered {
			msg = "No test cases match the filters."
		}
		_, err := fmt.Fprintln(w, msg)
		return err
	}

	s := style.New(w)
	var b strings.Builder
	shown := 0
	var printSuite func(su internalbuild.TestSuite, indent string)
	printSuite = func(su internalbuild.TestSuite, indent string) {
		fmt.Fprintf(&b, "%s%s  %s\n", indent, s.Bold.Render(su.Name), s.Dim.Render(testCounts(su.TestCount, su.FailureCount, su.ErrorCount, su.SkippedCount, su.DurationMS)))
		for _, tc := range su.TestCases {
			shown++
			// Pad outside the styled text: ANSI codes would count toward the width.
			status := s.TestStatus(tc.Status).Render(tc.Status) + strings.Repeat(" ", max(0, len("skipped")-len(tc.Status)))
			fmt.Fprintf(&b, "%s  %s  %s  %s", indent, status, internalbuild.TestCaseDisplayName(tc.ClassName, tc.Name), s.Dim.Render(testDuration(tc.DurationMS)))
			if tc.Flaky {
				fmt.Fprintf(&b, "  %s", s.Bold.Render("flaky"))
			}
			if len(tc.Runs) > 1 {
				fmt.Fprintf(&b, "  %s", s.Dim.Render(fmt.Sprintf("%d runs", len(tc.Runs))))
			}
			if len(tc.TestSuites) > 0 {
				fmt.Fprintf(&b, "  %s %s", s.Dim.Render("suites:"), strings.Join(tc.TestSuites, " › "))
			}
			if tc.Identifier != "" {
				fmt.Fprintf(&b, "  %s", s.Dim.Render("identifier: "+tc.Identifier))
			}
			b.WriteString("\n")
			if msg := strings.TrimSpace(tc.Message); msg != "" && (tc.Status == "failed" || tc.Status == "error") {
				for _, line := range strings.Split(msg, "\n") {
					fmt.Fprintf(&b, "%s    %s\n", indent, line)
				}
			}
		}
		for _, child := range su.TestSuites {
			printSuite(child, indent+"  ")
		}
	}
	var total internalbuild.TestSuite
	for _, su := range res.TestSuites {
		printSuite(su, "")
		total.TestCount += su.TestCount
		total.FailureCount += su.FailureCount
		total.ErrorCount += su.ErrorCount
		total.SkippedCount += su.SkippedCount
		total.DurationMS += su.DurationMS
	}

	// A filter drops whole suites, so their totals would no longer add up.
	if filtered {
		noun := "test cases"
		if shown == 1 {
			noun = "test case"
		}
		fmt.Fprintf(&b, "\n%d matching %s.\n", shown, noun)
	} else {
		fmt.Fprintf(&b, "\nTotal: %s\n", testCounts(total.TestCount, total.FailureCount, total.ErrorCount, total.SkippedCount, total.DurationMS))
	}
	if shown > 0 {
		fmt.Fprintf(&b, "\n%s\n%s\n",
			s.Dim.Render("Quarantining a test case needs its exact suites, name and class name, or its identifier. To look them up in this build:"),
			s.Dim.Render(fmt.Sprintf("  bitrise app quarantined-tests add --app %s --build %s --name NAME", appSlug, buildSlug)))
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func testCounts(tests, failures, errs, skipped int, durationMS int64) string {
	return fmt.Sprintf("%d tests, %d failed, %d errors, %d skipped, %s", tests, failures, errs, skipped, testDuration(durationMS))
}

// testDuration keeps millisecond precision below a second, where most test
// cases finish, and a tenth of a second above it.
func testDuration(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	if d < time.Second {
		return d.String()
	}
	return d.Round(100 * time.Millisecond).String()
}

