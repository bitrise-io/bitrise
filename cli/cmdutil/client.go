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

// liveToken resolves the token to use, refreshing an OAuth-managed token if
// expired. BITRISE_TOKEN, when set, is used verbatim and never refreshed.
func liveToken(cmd *cobra.Command) (string, error) {
	if t := os.Getenv(auth.EnvToken); t != "" {
		return t, nil
	}
	a, err := auth.Load()
	if err != nil {
		return "", err
	}
	if a.Token == "" {
		return "", ErrNoToken
	}
	return OAuthConfig().EnsureFreshPAT(cmd.Context(), a.Token)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
