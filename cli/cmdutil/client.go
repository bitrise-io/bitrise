package cmdutil

import (
	"errors"
	"os"

	"github.com/bitrise-io/bitrise/v2/internal/auth"
	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/spf13/cobra"
)

var ErrNoToken = errors.New("no Bitrise access token configured (set BITRISE_TOKEN)")

// NewAPIClient checks BITRISE_TOKEN before internal/auth's stored token,
// since no command can populate auth.yaml yet (auth login is unported).
func NewAPIClient(cmd *cobra.Command) (*bitriseapi.Client, error) {
	tok := os.Getenv("BITRISE_TOKEN")
	if tok == "" {
		a, err := auth.Load()
		if err != nil {
			return nil, err
		}
		tok = a.Token
	}
	if tok == "" {
		return nil, ErrNoToken
	}

	r := config.FromContext(cmd.Context())
	return bitriseapi.New(r.APIBaseURL, tok), nil
}
