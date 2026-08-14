package build

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalbuild "github.com/bitrise-io/bitrise/v2/internal/build"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewViewCommand returns the `build view` subcommand.
func NewViewCommand() *cobra.Command {
	var web bool

	cmd := &cobra.Command{
		Use:   "view BUILD_ID",
		Short: "Show details of a single build",
		Long: `Show details for a single build identified by its ID.

BUILD_ID belongs to an app: pass --app ID, or set BITRISE_APP_ID (or run
"bitrise config set app_id ID").`,
		Example: `  bitrise build view abc123 --app my-app-id
  bitrise build view abc123 --app my-app-id --format json
  bitrise build view abc123 --app my-app-id --web`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(cmd, args, web, cmdutil.OpenBrowser)
		},
	}

	cmdutil.AddAppFlag(cmd.Flags(), "app ID (or set BITRISE_APP_ID)")
	cmd.Flags().BoolVar(&web, "web", false, "open the build page in the browser instead of printing")
	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")

	return cmd
}

// runView is NewViewCommand's RunE, with browser-opening injected so tests
// can exercise --web without launching a real browser.
func runView(cmd *cobra.Command, args []string, web bool, openBrowser func(string) error) error {
	cmdutil.LogCommandParameters(cmd)

	format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
	if err := output.ConfigureOutputFormat(format); err != nil {
		return fmt.Errorf("failed to configure output format: %w", err)
	}

	appSlug, err := cmdutil.ResolveAppSlug(cmd)
	if err != nil {
		return err
	}
	buildSlug := args[0]

	if web {
		url := buildWebURL(cmdutil.ResolveWebBaseURL(cmd), appSlug, buildSlug)
		if err := openBrowser(url); err != nil {
			return err
		}
		_, err := fmt.Fprintf(cmd.ErrOrStderr(), "Opened %s\n", url)
		return err
	}

	client, err := cmdutil.NewAPIClient(cmd)
	if err != nil {
		return err
	}
	b, err := internalbuild.NewService(client).View(cmd.Context(), appSlug, buildSlug)
	if err != nil {
		return err
	}
	if b.BuildURL == "" {
		b.BuildURL = buildWebURL(cmdutil.ResolveWebBaseURL(cmd), appSlug, buildSlug)
	}

	return output.Render(cmd.OutOrStdout(), output.Format, b, printBuildText)
}
