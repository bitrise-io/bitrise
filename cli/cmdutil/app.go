package cmdutil

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagApp is the app slug a command acts on.
const FlagApp = "app"

// AddAppFlag registers --app. Registered per-subcommand rather than as a
// persistent parent flag, since some of these commands are also
// re-registered standalone as legacy top-level aliases (see cli/root.go)
// that never attach to the yml parent.
func AddAppFlag(fs *pflag.FlagSet, help string) {
	fs.String(FlagApp, "", help)
}

// ResolveAppSlug returns the app slug from --app. There is no env var or
// config-file fallback yet — add one when a command needs it.
func ResolveAppSlug(cmd *cobra.Command) (string, error) {
	if v, _ := cmd.Flags().GetString(FlagApp); v != "" {
		return v, nil
	}
	return "", AppSlugRequiredErr()
}

// AppSlugRequiredErr returns the standard missing-app-slug error.
func AppSlugRequiredErr() error {
	return errors.New("--app is required")
}
