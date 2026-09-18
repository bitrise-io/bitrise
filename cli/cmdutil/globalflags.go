package cmdutil

import "github.com/spf13/cobra"

// FlagOutput is the root-persistent output format flag (raw|json|yml, plus
// the human alias for raw). Per-command --format flags (FormatKey) take
// precedence over it when set.
const FlagOutput = "output"

// FlagQuiet suppresses non-error diagnostic messages. Only rde commands act
// on it for now (see IsQuiet).
const FlagQuiet = "quiet"

// FlagNoColor disables ANSI colors regardless of terminal detection.
const FlagNoColor = "no-color"

// FlagTheme selects the color theme (see internal/style.Themes).
const FlagTheme = "theme"

// EnvOutput overrides the output format when neither --output nor the
// "output" config key is set.
const EnvOutput = "BITRISE_OUTPUT"

// EnvTheme overrides the color theme when neither --theme nor the "theme"
// config key is set.
const EnvTheme = "BITRISE_CLI_THEME"

// IsQuiet reports whether the persistent --quiet flag was set.
func IsQuiet(cmd *cobra.Command) bool {
	q, _ := cmd.Root().PersistentFlags().GetBool(FlagQuiet)
	return q
}
