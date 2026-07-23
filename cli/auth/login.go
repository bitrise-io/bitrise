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
		withToken     bool
		emailLogin    string
		passwordStdin bool
		oauthLogin    bool
		webLogin      bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save a Bitrise access token",
		Long: `Save a Bitrise access token for future commands to use.

By default, in an interactive terminal, this opens your browser to sign in to
Bitrise (OAuth) and stores a managed, auto-refreshing token. The modes:

  Browser sign-in (default in an interactive terminal; explicit with --oauth).
     Opens your browser to sign in, exchanges the result for a Personal Access
     Token, and refreshes it automatically so you rarely sign in again:

         bitrise auth login
         bitrise auth login --oauth

     This needs the browser on the same machine as the CLI (the sign-in is
     handed back over a loopback address). On a remote/headless host over SSH
     it can't complete — pipe a token instead (see below).

  Token (--with-token, or any non-interactive stdin).
     Reads a Personal Access Token from stdin. This is also used automatically
     when stdin is not a terminal, so CI and pipes keep working without a flag:

         echo "$BITRISE_PAT" | bitrise auth login
         echo "$BITRISE_PAT" | bitrise auth login --with-token

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
		Example: `  bitrise auth login                                     # browser sign-in (OAuth)
  echo "$BITRISE_PAT" | bitrise auth login --with-token  # paste/pipe a token
  bitrise auth login --email alice@example.com           # email/password`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			switch {
			case oauthLogin || webLogin:
				return runOAuthLogin(cmd)
			case emailLogin != "":
				return runEmailLogin(cmd, emailLogin, passwordStdin)
			case withToken:
				return runTokenLogin(cmd)
			case passwordStdin:
				return fmt.Errorf("--password-stdin requires --email (token login reads the token, not a password)")
			case cmdutil.IsTerminal(cmd.InOrStdin()):
				// Interactive and no mode chosen: default to browser OAuth.
				return runOAuthLogin(cmd)
			default:
				// Non-interactive stdin (CI, pipes): read a token from stdin.
				return runTokenLogin(cmd)
			}
		},
	}
	cmd.Flags().BoolVar(&withToken, "with-token", false, "read token from stdin without an interactive prompt")
	cmd.Flags().StringVar(&emailLogin, "email", "", "sign in by email/password and mint a Personal Access Token")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "with --email, read the password from stdin without prompting")
	cmd.Flags().BoolVar(&oauthLogin, "oauth", false, "sign in via the browser (OAuth) and store a managed, auto-refreshing token")
	// --web is a hidden alias for --oauth ("open in the browser").
	cmd.Flags().BoolVar(&webLogin, "web", false, "alias for --oauth")
	_ = cmd.Flags().MarkHidden("web")
	// The three login modes are mutually exclusive. --oauth and --web are
	// aliases, so they're not exclusive with each other.
	for _, mode := range []string{"oauth", "web"} {
		cmd.MarkFlagsMutuallyExclusive(mode, "with-token")
		cmd.MarkFlagsMutuallyExclusive(mode, "email")
		cmd.MarkFlagsMutuallyExclusive(mode, "password-stdin")
	}
	cmd.MarkFlagsMutuallyExclusive("with-token", "email")
	cmd.MarkFlagsMutuallyExclusive("with-token", "password-stdin")
	return cmd
}

// runTokenLogin never prompts — a bare interactive `auth login` defaults to
// OAuth instead, so the token always arrives on stdin.
func runTokenLogin(cmd *cobra.Command) error {
	tok, err := cmdutil.ReadSecretInput(cmd.InOrStdin(), cmd.ErrOrStderr(), "", true)
	if err != nil {
		return err
	}
	if tok == "" {
		return fmt.Errorf("token is empty")
	}
	if err := auth.Save(auth.Auth{Token: tok}); err != nil {
		return err
	}
	confirmLoginSaved(cmd)
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
	confirmLoginSaved(cmd)
	return nil
}

func runOAuthLogin(cmd *cobra.Command) error {
	return doOAuthLogin(cmd, cmdutil.OpenBrowser)
}

// doOAuthLogin takes openBrowser as a param so tests can complete the
// loopback callback without a real browser.
func doOAuthLogin(cmd *cobra.Command, openBrowser func(string) error) error {
	a, err := cmdutil.OAuthConfig().Login(cmd.Context(), openBrowser, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	if err := auth.Save(a); err != nil {
		return err
	}
	confirmLoginSaved(cmd)
	return nil
}

// confirmLoginSaved warns when BITRISE_TOKEN is set, since it shadows the
// token just saved (see liveToken) — otherwise the login silently has no
// effect on later commands.
func confirmLoginSaved(_ *cobra.Command) {
	log.Print("Saved access token")
	if os.Getenv(auth.EnvToken) != "" {
		log.Warnf("%s is set and takes precedence over the token just saved.", auth.EnvToken)
		log.Warnf("Commands will use it, not this login — run 'unset %s' to use the saved token.", auth.EnvToken)
	}
}
