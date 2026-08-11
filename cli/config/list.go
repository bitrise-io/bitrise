package config

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/output"
)

// configList is the JSON/YAML shape of `config list`. No field is omitempty:
// this command's job is to enumerate every recognized key's state, so an
// unset key must still appear (as "") rather than vanish from the JSON/YAML
// output while still showing as "(unset)" in raw mode.
type configList struct {
	APIBaseURL         string `json:"api_base_url" yaml:"api_base_url"`
	WebBaseURL         string `json:"web_base_url" yaml:"web_base_url"`
	AppID              string `json:"app_id" yaml:"app_id"`
	DefaultWorkspaceID string `json:"default_workspace_id" yaml:"default_workspace_id"`
	Path               string `json:"path" yaml:"path"`
}

// NewListCommand returns the `config list` subcommand.
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the values currently saved in the global config file",
		Long: `List the values currently saved in the global config file.

This shows what is stored, not what every command will resolve: the
BITRISE_WEB_BASE_URL and BITRISE_APP_ID environment variables, and an app_id
pinned by a per-directory .bitrise-cli.yml, all take precedence at runtime.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			format, _ := cmd.Flags().GetString(cmdutil.FormatKey)
			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			cfg, err := internalconfig.Load()
			if err != nil {
				return err
			}
			p, err := internalconfig.Path()
			if err != nil {
				return err
			}
			v := configList{APIBaseURL: cfg.APIBaseURL, WebBaseURL: cfg.WebBaseURL, AppID: cfg.AppID, DefaultWorkspaceID: cfg.DefaultWorkspaceID, Path: p}

			if output.Format == output.FormatRaw {
				return printListHuman(cmd.OutOrStdout(), v)
			}
			return output.Print(v, output.Format)
		},
	}

	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return cmd
}

func printListHuman(w io.Writer, v configList) error {
	_, err := fmt.Fprintf(w, "Path: %s\n\n%s: %s\n%s: %s\n%s: %s\n%s: %s\n",
		v.Path,
		internalconfig.KeyAPIBaseURL, unsetLabel(v.APIBaseURL),
		internalconfig.KeyWebBaseURL, unsetLabel(v.WebBaseURL),
		internalconfig.KeyAppID, unsetLabel(v.AppID),
		internalconfig.KeyDefaultWorkspaceID, unsetLabel(v.DefaultWorkspaceID),
	)
	return err
}

func unsetLabel(v string) string {
	if v == "" {
		return "(unset)"
	}
	return v
}
