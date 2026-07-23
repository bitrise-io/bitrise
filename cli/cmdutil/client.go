package cmdutil

import (
	"errors"
	"os"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/spf13/cobra"
)

var ErrNoToken = errors.New("no Bitrise access token configured (run 'bitrise auth login' or set BITRISE_TOKEN)")

// NewAPIClient builds a *bitriseapi.Client using the token resolved by
// liveToken and the configured API base URL.
func NewAPIClient(cmd *cobra.Command) (*bitriseapi.Client, error) {
	tok, err := liveToken(cmd)
	if err != nil {
		return nil, err
	}
	r := config.FromContext(cmd.Context())
	return bitriseapi.New(r.APIBaseURL, tok), nil
}

// ResolveToken returns the configured token and whether it came from the
// BITRISE_TOKEN environment variable (true) or auth.yaml (false), without
// refreshing it. token is empty when neither is set; err is non-nil only on
// an unexpected auth.yaml load failure (a missing file is not an error).
func ResolveToken() (token string, fromEnv bool, err error) {
	if t := os.Getenv(auth.EnvToken); t != "" {
		return t, true, nil
	}
	a, err := auth.Load()
	if err != nil {
		return "", false, err
	}
	return a.Token, false, nil
}

// liveToken resolves the token to use, refreshing an OAuth-managed token if
// expired. BITRISE_TOKEN, when set, is used verbatim and never refreshed.
func liveToken(cmd *cobra.Command) (string, error) {
	tok, fromEnv, err := ResolveToken()
	if err != nil {
		return "", err
	}
	if tok == "" {
		return "", ErrNoToken
	}
	if fromEnv {
		return tok, nil
	}
	return OAuthConfig().EnsureFreshPAT(cmd.Context(), tok)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
