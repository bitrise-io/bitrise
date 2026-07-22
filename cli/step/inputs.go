package step

import (
	"strings"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/output"
	"github.com/spf13/cobra"
)

func NewInputsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inputs STEP_REF",
		Short: "List inputs of a step version",
		Long: `List the inputs (and their defaults) for a given step version.

STEP_REF must include an exact version: step_id@version
For custom step sources: step_lib_source::step_id@version`,
		Example: `  bitrise step inputs git-clone@8.3.1
  bitrise step inputs git-clone@8.3.1 --format json`,
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

			inputs, err := client.StepInputs(cmd.Context(), args[0])
			if err != nil {
				cmdutil.Failf("Fetching step inputs failed, error: %s", err)
			}

			if output.Format == output.FormatRaw {
				printInputsTable(inputs)
			} else {
				output.Print(inputs, output.Format)
			}
			return nil
		},
	}

	cmd.Flags().StringP(cmdutil.OuputFormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}

func printInputsTable(inputs []bitriseapi.StepInputOutputResponse) {
	if len(inputs) == 0 {
		log.Print("No inputs found.")
		return
	}
	rows := make([][]string, 0, len(inputs))
	for _, in := range inputs {
		required, sensitive := "", ""
		if in.IsRequired {
			required = "yes"
		}
		if in.IsSensitive {
			sensitive = "yes"
		}
		rows = append(rows, []string{in.Name, in.Title, in.DefaultValue, required, sensitive, strings.Join(in.ValueOptions, ", ")})
	}
	log.Print(renderTable([]string{"NAME", "TITLE", "DEFAULT", "REQUIRED", "SENSITIVE", "OPTIONS"}, rows))
}
