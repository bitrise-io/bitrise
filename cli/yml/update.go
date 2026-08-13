package yml

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalyml "github.com/bitrise-io/bitrise/v2/internal/yml"
)

// NewUpdateCommand returns the `yml update` subcommand.
func NewUpdateCommand() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Upload a new bitrise.yml to Bitrise",
		Long: `Upload a new bitrise.yml configuration to Bitrise for an app.

Reads from --file if provided, otherwise reads from stdin. Pass --file - to
read from stdin explicitly.

Note: if the app is configured to read its bitrise.yml from the repository,
this command succeeds but the change will not affect builds.

Inside a Bitrise build, omitting --app targets the app the build runs for,
overwriting its own stored configuration. Always pass --app when updating a
different app from a build.

Bitrise stores the configuration as structured data rather than as the file
you upload, so comments, key order and YAML anchors are not preserved: a
later 'bitrise yml get' returns an equivalent, reformatted document.`,
		Example: `  bitrise yml update --app my-app-id --file bitrise.yml
  cat bitrise.yml | bitrise yml update --app my-app-id
  bitrise yml update --app my-app-id < bitrise.yml`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			appSlug, err := cmdutil.ResolveAppSlug(cmd)
			if err != nil {
				return err
			}

			rawYAML, err := readInput(cmd.InOrStdin(), filePath)
			if err != nil {
				return fmt.Errorf("read bitrise.yml: %w", err)
			}
			if len(rawYAML) == 0 {
				return fmt.Errorf("bitrise.yml content is empty")
			}

			client, err := cmdutil.NewAPIClient(cmd)
			if err != nil {
				return err
			}

			if err := internalyml.NewService(client).Update(cmd.Context(), appSlug, string(rawYAML)); err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.ErrOrStderr(), "bitrise.yml updated successfully")
			return err
		},
	}

	// No -f shorthand: every other cloud command uses -f for --format, so
	// binding it to --file here would make `yml update -f json` open a file
	// named "json" while `yml get -f json` selects a format.
	cmd.Flags().StringVar(&filePath, "file", "", "path to the bitrise.yml file, or - for stdin (reads from stdin if omitted)")
	cmdutil.AddAppFlag(cmd.Flags(), "app ID to update the bitrise.yml for (or set BITRISE_APP_ID; inside a build, defaults to the app the build runs for)")

	return cmd
}

// readInput reads from filePath when set, otherwise from r (stdin).
func readInput(r io.Reader, filePath string) ([]byte, error) {
	if filePath != "" && filePath != "-" {
		return os.ReadFile(filePath)
	}
	return io.ReadAll(r)
}
