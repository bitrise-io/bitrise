package user

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internaluser "github.com/bitrise-io/bitrise/v2/internal/user"
	"github.com/bitrise-io/bitrise/v2/internal/webclient"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewCreateCommand returns the `user create` subcommand.
func NewCreateCommand() *cobra.Command {
	var (
		email         string
		username      string
		firstName     string
		lastName      string
		passwordStdin bool
	)

	// MarkFlagRequired only checks that a flag was set, so an explicitly empty
	// value would reach the signup request; the non-empty guard in RunE closes
	// that, and both read the flag names from here.
	requiredFlags := []struct {
		name  string
		value *string
		usage string
	}{
		{"email", &email, "email address to register"},
		{"username", &username, "desired username"},
		{"first-name", &firstName, "first name on the account"},
		{"last-name", &lastName, "last name on the account"},
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Bitrise account",
		Long: `Create a new Bitrise account by email and password.

Required flags:
  --email ADDRESS    the email address to register
  --username NAME    desired username (must be unique)
  --first-name N     first name on the account
  --last-name N      last name on the account

Optional flags:
  --password-stdin   read the password from stdin instead of prompting

Password input:
  By default the command prompts for the password (input is masked when stdin
  is a terminal). Use --password-stdin to read it from stdin without a prompt
  — the right choice for piping or scripts:

      printf '%s' "$NEW_PASSWORD" | bitrise user create \
          --email a@b.io --username alice --first-name A --last-name B --password-stdin

Email verification:
  After signup the server emails a verification link. Click it before running
  'bitrise auth login --email <addr>' — sign-in is blocked on unverified
  accounts.

Target host:
  The signup request goes to web_base_url (default app.bitrise.io),
  overridable via $BITRISE_WEB_BASE_URL or 'bitrise config set web_base_url'
  — never by a per-directory .bitrise-cli.yml, so a repo you merely clone
  can't silently redirect where your password is sent.`,
		Example: `  bitrise user create --email alice@example.com --username alice --first-name Alice --last-name L
  printf '%s' "$NEW_PASSWORD" | bitrise user create \
      --email alice@example.com --username alice --first-name Alice --last-name L --password-stdin --format json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			for _, f := range requiredFlags {
				if *f.value == "" {
					return fmt.Errorf("--%s requires a non-empty value", f.name)
				}
			}

			if err := cmdutil.CheckPasswordStdinPiped(passwordStdin, cmdutil.IsTerminal(cmd.InOrStdin()), "bitrise user create --email <email>"); err != nil {
				return err
			}
			pw, err := cmdutil.ReadPasswordInput(cmd.InOrStdin(), cmd.ErrOrStderr(), "Choose a password: ", passwordStdin)
			if err != nil {
				return err
			}
			if pw == "" {
				return fmt.Errorf("password is empty")
			}

			wc, err := webclient.New(cmdutil.ResolveWebBaseURL(cmd))
			if err != nil {
				return err
			}

			acct, err := internaluser.NewService(wc).Signup(cmd.Context(), internaluser.SignupInput{
				Email:     email,
				Username:  username,
				Password:  pw,
				FirstName: firstName,
				LastName:  lastName,
			})
			if err != nil {
				return err
			}

			return output.Render(cmd.OutOrStdout(), output.Format, acct, printCreateHuman)
		},
	}

	for _, f := range requiredFlags {
		cmd.Flags().StringVar(f.value, f.name, "", f.usage+" (required)")
		_ = cmd.MarkFlagRequired(f.name)
	}
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read the password from stdin without prompting")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}

func printCreateHuman(w io.Writer, a internaluser.Account) error {
	var b strings.Builder
	fmt.Fprintln(&b, "✓ Account created")
	fmt.Fprintf(&b, "Email:    %s\n", a.Email)
	if a.Username != "" {
		fmt.Fprintf(&b, "Username: %s\n", a.Username)
	}
	if a.Slug != "" {
		fmt.Fprintf(&b, "ID:       %s\n", a.Slug)
	}
	if !a.Confirmed {
		fmt.Fprintf(&b, "\nCheck your email and click the verification link, then run:\n  bitrise auth login --email %s\n", a.Email)
	}
	_, err := io.WriteString(w, b.String())
	return err
}
