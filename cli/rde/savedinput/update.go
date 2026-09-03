package savedinput

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newUpdateCmd() *cobra.Command {
	var (
		value      string
		valueStdin bool
		isSecret   bool
		format     string
	)
	c := &cobra.Command{
		Use:   "update SAVED_INPUT_ID",
		Short: "Update a saved input's value and/or secret flag",
		Long: `Update a saved input's value and/or secret flag.

Pass --value VALUE to set a new value, or --value-stdin to read it from stdin
without prompting (keeps secrets out of shell history). Omit both to change only
the --secret flag.`,
		Args: cmdutil.RequireArgs("SAVED_INPUT_ID"),
		Example: `  bitrise rde saved-input update ID --value new-value
  echo -n "ghp_xxx" | bitrise rde saved-input update ID --value-stdin --secret`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			req := internalrde.UpdateSavedInputRequest{}
			v, provided, err := resolveValue(cmd, value, cmd.Flags().Changed("value"), valueStdin, false)
			if err != nil {
				return err
			}
			if provided {
				req.Value = &v
			}
			if cmd.Flags().Changed("secret") {
				b := isSecret
				req.IsSecret = &b
			}
			if req.Value == nil && req.IsSecret == nil {
				return fmt.Errorf("at least one of --value, --value-stdin, or --secret is required")
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			in, err := internalrde.NewService(client).UpdateSavedInput(cmd.Context(), args[0], req)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, in, renderDetail)
		},
	}
	c.Flags().StringVar(&value, "value", "", "new value (literal)")
	c.Flags().BoolVar(&valueStdin, "value-stdin", false, "read the new value from stdin without prompting")
	c.Flags().BoolVar(&isSecret, "secret", false, "set/unset the secret flag")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	c.MarkFlagsMutuallyExclusive("value", "value-stdin")
	return c
}
