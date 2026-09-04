package session

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

// deleteTerminatedResult is the --format json/yml shape: {"deleted_count": N}.
type deleteTerminatedResult struct {
	DeletedCount int `json:"deleted_count" yaml:"deleted_count"`
}

func newDeleteTerminatedCmd() *cobra.Command {
	var (
		assumeYes bool
		format    string
	)
	c := &cobra.Command{
		Use:   "delete-terminated",
		Short: "Permanently delete every terminated session in the workspace",
		Long: `Permanently delete every terminated session in the workspace.
This cannot be undone. Pass --yes to skip the confirmation prompt.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			if !assumeYes {
				if _, err := fmt.Fprint(cmd.ErrOrStderr(),
					"This will permanently delete every terminated session in the workspace.\nProceed? [y/N]: "); err != nil {
					return err
				}
				answer, err := cmdutil.ReadSecretInput(cmd.InOrStdin(), cmd.ErrOrStderr(), "", true)
				if err != nil {
					return err
				}
				if answer != "y" && answer != "Y" && answer != "yes" {
					return fmt.Errorf("aborted")
				}
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			count, err := internalrde.NewService(client).DeleteTerminatedSessions(cmd.Context(), workspaceID)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, deleteTerminatedResult{DeletedCount: count}, renderDeleteTerminated)
		},
	}
	c.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}

func renderDeleteTerminated(w io.Writer, r deleteTerminatedResult) error {
	_, err := fmt.Fprintf(w, "Deleted %d terminated session(s)\n", r.DeletedCount)
	return err
}
