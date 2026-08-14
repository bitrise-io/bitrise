package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalapp "github.com/bitrise-io/bitrise/v2/internal/app"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewViewCommand returns the `app view` subcommand.
func NewViewCommand() *cobra.Command {
	var web bool

	cmd := &cobra.Command{
		Use:   "view [APP_ID]",
		Short: "Show details of a single app",
		Long: `Show details for a single app identified by its ID.

APP_ID falls back to --app, then $BITRISE_APP_ID, then $BITRISE_APP_SLUG
(injected inside a Bitrise build), then the app_id saved by 'bitrise app
create' or 'bitrise config set app_id', when omitted.`,
		Example: `  bitrise app view stub-app-aaa
  bitrise app view stub-app-aaa --format json
  bitrise app view stub-app-aaa --web
  BITRISE_APP_ID=stub-app-aaa bitrise app view`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(cmd, args, web, cmdutil.OpenBrowser)
		},
	}

	cmdutil.AddAppFlag(cmd.Flags(), "app ID to view (or set BITRISE_APP_ID); overridden by the positional argument")
	cmd.Flags().BoolVar(&web, "web", false, "open the app page in the browser instead of printing")
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

	appSlug, err := cmdutil.ResolveAppSlugArg(cmd, args)
	if err != nil {
		return err
	}

	if web {
		url := fmt.Sprintf("%s/app/%s", cmdutil.ResolveWebBaseURL(cmd), appSlug)
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
	a, err := internalapp.NewService(client).View(cmd.Context(), appSlug)
	if err != nil {
		return err
	}

	return output.Render(cmd.OutOrStdout(), output.Format, a, printAppText)
}

func printAppText(w io.Writer, a internalapp.App) error {
	var b strings.Builder
	fmt.Fprintf(&b, "Title:        %s\n", a.Title)
	fmt.Fprintf(&b, "ID:           %s\n", a.Slug)
	fmt.Fprintf(&b, "Provider:     %s\n", a.Provider)
	fmt.Fprintf(&b, "Repo URL:     %s\n", a.RepoURL)
	if a.OwnerSlug != "" {
		fmt.Fprintf(&b, "Workspace:    %s\n", a.OwnerSlug)
	}
	if a.ProjectType != "" {
		fmt.Fprintf(&b, "Project type: %s\n", a.ProjectType)
	}
	if a.IsDisabled {
		fmt.Fprintln(&b, "Disabled:     yes")
	}
	_, err := io.WriteString(w, b.String())
	return err
}
