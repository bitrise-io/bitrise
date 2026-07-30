package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internaluser "github.com/bitrise-io/bitrise/v2/internal/user"
	"github.com/bitrise-io/bitrise/v2/internal/webclient"
	"github.com/bitrise-io/bitrise/v2/log"
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
  accounts.`,
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

			if email == "" {
				return fmt.Errorf("--email is required")
			}
			if username == "" {
				return fmt.Errorf("--username is required")
			}
			if firstName == "" {
				return fmt.Errorf("--first-name is required")
			}
			if lastName == "" {
				return fmt.Errorf("--last-name is required")
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

			wc, err := webclient.New(cmdutil.ResolveWebBaseURL())
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

			if output.Format == output.FormatRaw {
				printCreateHuman(acct)
				return nil
			}
			output.Print(acct, output.Format)
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "email address to register (required)")
	cmd.Flags().StringVar(&username, "username", "", "desired username (required)")
	cmd.Flags().StringVar(&firstName, "first-name", "", "first name on the account (required)")
	cmd.Flags().StringVar(&lastName, "last-name", "", "last name on the account (required)")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read the password from stdin without prompting")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}

func printCreateHuman(a internaluser.Account) {
	log.Print("✓ Account created")
	log.Printf("Email:    %s", a.Email)
	if a.Username != "" {
		log.Printf("Username: %s", a.Username)
	}
	if a.Slug != "" {
		log.Printf("ID:       %s", a.Slug)
	}
	if !a.Confirmed {
		log.Print("")
		log.Print("Check your email and click the verification link, then run:")
		log.Printf("  bitrise auth login --email %s", a.Email)
	}
}
