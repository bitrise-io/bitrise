package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/log"
)

func NewLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save a Bitrise access token",
		Long: `Save a Bitrise access token for future commands to use.

Reads a Personal Access Token from stdin — pasted interactively (masked, not
echoed) or piped in non-interactively (CI, scripts):

    bitrise auth login
    echo "$BITRISE_PAT" | bitrise auth login

The token is written to $XDG_CONFIG_HOME/bitrise/cli/auth.yaml with 0600
permissions and is never echoed (use 'auth status' to verify, 'auth logout'
to clear).`,
		Example: `  bitrise auth login                                     # paste a token
  echo "$BITRISE_PAT" | bitrise auth login               # pipe a token`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)
			return runTokenLogin(cmd)
		},
	}
	return cmd
}

func runTokenLogin(cmd *cobra.Command) error {
	tok, err := cmdutil.ReadSecretInput(cmd.InOrStdin(), cmd.ErrOrStderr(), "Token: ", false)
	if err != nil {
		return err
	}
	if tok == "" {
		return fmt.Errorf("token is empty")
	}
	if err := auth.Save(auth.Auth{Token: tok}); err != nil {
		return err
	}
	confirmLoginSaved()
	return nil
}

// confirmLoginSaved warns when BITRISE_TOKEN is set, since it shadows the
// token just saved (see liveToken) — otherwise the login silently has no
// effect on later commands.
func confirmLoginSaved() {
	log.Print("Saved access token")
	if os.Getenv(auth.EnvToken) != "" {
		log.Warnf("%s is set and takes precedence over the token just saved.", auth.EnvToken)
		log.Warnf("Commands will use it, not this login — run 'unset %s' to use the saved token.", auth.EnvToken)
	}
}
