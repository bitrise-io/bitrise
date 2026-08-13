package build

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalyml "github.com/bitrise-io/bitrise/v2/internal/yml"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewYMLCommand returns the `build yml` subcommand.
func NewYMLCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yml BUILD_ID",
		Short: "Print the bitrise.yml a specific build ran with",
		Long: `Print the bitrise.yml configuration that a specific build ran with.

This is a shortcut for "bitrise yml get --app ID --build BUILD_ID".`,
		Example: `  bitrise build yml abc123 --app my-app-id
  bitrise build yml abc123 --app my-app-id --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			result, err := internalyml.NewService(client).Get(cmd.Context(), appSlug, args[0])
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), output.Format, result, func(w io.Writer, result internalyml.GetResult) error {
				content := result.Content
				if content != "" && !strings.HasSuffix(content, "\n") {
					content += "\n"
				}
				_, err := fmt.Fprint(w, content)
				return err
			})
		},
	}

	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}
