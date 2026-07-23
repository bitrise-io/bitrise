package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/user"
	"github.com/bitrise-io/bitrise/v2/internal/webclient"
	"github.com/bitrise-io/bitrise/v2/log"
)

func NewLoginCommand() *cobra.Command {
	var (
		emailLogin    string
		passwordStdin bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save a Bitrise access token",
		Long: `Save a Bitrise access token for future commands to use.

  Token (default). Reads a Personal Access Token from stdin — pasted
  interactively (masked, not echoed) or piped in non-interactively (CI,
  scripts):

      bitrise auth login
      echo "$BITRISE_PAT" | bitrise auth login

  Email and password (--email).
     Signs in to app.bitrise.io with your account credentials, then asks the
     server to mint a fresh Personal Access Token and stores it. The cookie
     session used to mint the token is dropped immediately. Your account must
     have its email verified:

         bitrise auth login --email alice@example.com
         printf '%s' "$PW" | bitrise auth login --email alice@example.com --password-stdin

The resulting token is written to $XDG_CONFIG_HOME/bitrise/cli/auth.yaml with
0600 permissions and is never echoed (use 'auth status' to verify, 'auth
logout' to clear).`,
		Example: `  bitrise auth login                                     # paste a token
  echo "$BITRISE_PAT" | bitrise auth login               # pipe a token
  bitrise auth login --email alice@example.com           # email/password`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			switch {
			case emailLogin != "":
				return runEmailLogin(cmd, emailLogin, passwordStdin)
			case passwordStdin:
				return fmt.Errorf("--password-stdin requires --email (token login reads the token, not a password)")
			default:
				return runTokenLogin(cmd)
			}
		},
	}
	cmd.Flags().StringVar(&emailLogin, "email", "", "sign in by email/password and mint a Personal Access Token")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "with --email, read the password from stdin without prompting")
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

func runEmailLogin(cmd *cobra.Command, email string, passwordStdin bool) error {
	pw, err := cmdutil.ReadSecretInput(cmd.InOrStdin(), cmd.ErrOrStderr(), "Password: ", passwordStdin)
	if err != nil {
		return err
	}
	if pw == "" {
		return fmt.Errorf("password is empty")
	}
	wc, err := webclient.New(cmdutil.ResolveWebBaseURL())
	if err != nil {
		return err
	}
	host, _ := os.Hostname()
	if host == "" {
		host = "unknown-host"
	}
	svc := user.NewService(wc)
	tok, err := svc.Login(cmd.Context(), user.LoginInput{Login: email, Password: pw}, fmt.Sprintf("bitrise (%s)", host))
	if err != nil {
		if user.IsUnconfirmedEmailErr(err) {
			return fmt.Errorf("this account hasn't verified its email yet — click the link in the confirmation email, then re-run")
		}
		return err
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
