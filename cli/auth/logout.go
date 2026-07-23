package auth

import (
	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/log"
)

func NewLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the saved access token",
		Long: `Remove the auth.yaml file. Does not affect a token set via the
BITRISE_TOKEN environment variable.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)
			return runLogout()
		},
	}
}

func runLogout() error {
	if err := auth.Clear(); err != nil {
		return err
	}
	log.Print("Cleared saved access token")
	return nil
}
