package auth

import (
	"os"

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
	confirmLogoutCleared()
	return nil
}

// confirmLogoutCleared warns when BITRISE_TOKEN is set, since it shadows the
// removal just performed (see liveToken) — otherwise commands stay
// authenticated via the env var and the user believes they're signed out.
func confirmLogoutCleared() {
	log.Print("Cleared saved access token")
	if os.Getenv(auth.EnvToken) != "" {
		log.Warnf("%s is still set and will be used by commands — run 'unset %s' to fully sign out.", auth.EnvToken, auth.EnvToken)
	}
}
