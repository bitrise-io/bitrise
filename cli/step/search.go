package step

import (
	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/output"
	"github.com/spf13/cobra"
)

func NewSearchCommand() *cobra.Command {
	var categories, maintainers []string

	cmd := &cobra.Command{
		Use:   "search QUERY",
		Short: "Find steps by name, description, or tags",
		Long: `Find steps for use in workflows or step bundles.

Returns only the latest, non-deprecated version of each matching step.

Valid categories:
  build, code-sign, test, deploy, notification, access-control,
  artifact-info, installer, dependency, utility

Valid maintainers:
  bitrise   official Bitrise steps
  verified  verified community steps
  community all community steps`,
		Example: `  bitrise step search clone
  bitrise step search deploy --category deploy --maintainer bitrise
  bitrise step search npm --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.OuputFormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				cmdutil.Failf("Failed to configure output format, error: %s", err)
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				cmdutil.Failf("%s", err)
			}

			steps, err := client.SearchSteps(cmd.Context(), bitriseapi.StepSearchOptions{
				Query:       args[0],
				Categories:  categories,
				Maintainers: maintainers,
			})
			if err != nil {
				cmdutil.Failf("Step search failed, error: %s", err)
			}

			if output.Format == output.FormatRaw {
				printStepsTable(steps)
			} else {
				output.Print(steps, output.Format)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&categories, "category", nil, "filter by category (may be repeated)")
	cmd.Flags().StringArrayVar(&maintainers, "maintainer", nil, "filter by maintainer: bitrise, verified, community (may be repeated)")
	cmd.Flags().StringP(cmdutil.OuputFormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}

func printStepsTable(steps []bitriseapi.StepResponse) {
	if len(steps) == 0 {
		log.Print("No steps found.")
		return
	}
	rows := make([][]string, 0, len(steps))
	for _, s := range steps {
		rows = append(rows, []string{s.StepRef, s.Title, s.Maintainer, s.Summary})
	}
	log.Print(renderTable([]string{"STEP_REF", "TITLE", "MAINTAINER", "SUMMARY"}, rows))
}
