package savedinput

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete SAVED_INPUT_ID",
		Short: "Delete a saved input",
		Args:  cmdutil.RequireArgs("SAVED_INPUT_ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			if err := internalrde.NewService(client).DeleteSavedInput(cmd.Context(), args[0]); err != nil {
				return err
			}
			if !cmdutil.IsQuiet(cmd) {
				_, err := fmt.Fprintf(cmd.ErrOrStderr(), "Deleted saved input %s\n", args[0])
				return err
			}
			return nil
		},
	}
}
