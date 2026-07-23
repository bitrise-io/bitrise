package auth

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/output"
)

// authStatus is the JSON/YML shape of `bitrise auth status`.
type authStatus struct {
	HasToken  bool   `json:"has_token"`
	TokenType string `json:"token_type,omitempty"`
	Source    string `json:"source"`
	// TokenExpiry is set (RFC 3339) only for OAuth-managed tokens.
	TokenExpiry string `json:"token_expiry,omitempty"`
	Path        string `json:"path"`
}

func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show whether an access token is configured and where it came from",
		Long: `Show whether an access token is configured and which source supplied it.

Sources, in precedence order:
  env        BITRISE_TOKEN environment variable
  auth file  auth.yaml, written by 'bitrise auth login' (OAuth or a
             pasted/email token — a new login overwrites the previous one).
             OAuth logins are shown as "oauth (auth file)" and refreshed
             automatically.
  none       no token configured`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.OuputFormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				cmdutil.Failf("Failed to configure output format, error: %s", err)
			}

			s, err := currentStatus()
			if err != nil {
				return err
			}

			if output.Format == output.FormatRaw {
				printStatusHuman(s)
			} else {
				output.Print(s, output.Format)
			}
			return nil
		},
	}
	cmd.Flags().StringP(cmdutil.OuputFormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return cmd
}

// resolveTokenAndSource reports what cmdutil.ResolveToken found, labeled for
// display — status reports what's stored, it doesn't refresh or mutate it.
func resolveTokenAndSource() (token, source string) {
	tok, fromEnv, err := cmdutil.ResolveToken()
	switch {
	case err != nil || tok == "":
		return "", "none"
	case fromEnv:
		return tok, "env (" + auth.EnvToken + ")"
	default:
		return tok, "auth file"
	}
}

// currentStatus is kept separate from NewStatusCommand's RunE so it can be
// tested directly.
func currentStatus() (authStatus, error) {
	p, err := auth.Path()
	if err != nil {
		return authStatus{}, err
	}
	tok, source := resolveTokenAndSource()
	s := authStatus{
		HasToken: tok != "",
		Path:     p,
		Source:   source,
	}
	if tok == "" {
		return s, nil
	}
	s.TokenType = auth.TokenType(tok)
	// Env-sourced tokens skip the auth.yaml OAuth check below entirely.
	if os.Getenv(auth.EnvToken) == "" {
		if a, err := auth.Load(); err == nil && a.IsOAuthManaged() {
			s.Source = "oauth (auth file)"
			if !a.TokenExpiry.IsZero() {
				s.TokenExpiry = a.TokenExpiry.Format(time.RFC3339)
			}
		}
	}
	return s, nil
}

func printStatusHuman(s authStatus) {
	if !s.HasToken {
		log.Print("✗ No access token configured.")
		log.Print("")
		log.Print("Run 'bitrise auth login' to save one,")
		log.Print("or set the BITRISE_TOKEN environment variable.")
		return
	}
	log.Print("✓ Access token configured")
	log.Printf("Type:    %s", s.TokenType)
	log.Printf("Source:  %s", s.Source)
	if s.TokenExpiry != "" {
		log.Printf("Expires: %s", s.TokenExpiry)
	}
	log.Printf("Path:    %s", s.Path)
}
