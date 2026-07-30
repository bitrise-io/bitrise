package yml

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalyml "github.com/bitrise-io/bitrise/v2/internal/yml"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewGetCommand returns the `yml get` subcommand.
func NewGetCommand() *cobra.Command {
	var buildSlug string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Print the bitrise.yml stored on Bitrise",
		Long: `Print the bitrise.yml configuration stored on Bitrise for an app.

When --build is provided, prints the bitrise.yml that a specific build ran with
instead of the app's current stored configuration.`,
		Example: `  bitrise yml get --app my-app-id
  bitrise yml get --app my-app-id --build abc123
  bitrise yml get --app my-app-id --format json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return err
			}

			appSlug, err := cmdutil.ResolveAppSlug(cmd)
			if err != nil {
				return err
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}

			result, err := internalyml.NewService(client).Get(cmd.Context(), appSlug, buildSlug)
			if err != nil {
				return err
			}

			if output.Format == output.FormatRaw {
				content := result.Content
				if content != "" && !strings.HasSuffix(content, "\n") {
					content += "\n"
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), content)
				return err
			}
			return output.Print(result, output.Format)
		},
	}

	cmd.Flags().StringVar(&buildSlug, "build", "", "build ID to retrieve the yml for")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID to retrieve the bitrise.yml for (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}
