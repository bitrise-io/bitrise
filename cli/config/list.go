package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/output"
)

// configList is the JSON/YAML shape of `config list`.
type configList struct {
	APIBaseURL string `json:"api_base_url,omitempty" yaml:"api_base_url,omitempty"`
	WebBaseURL string `json:"web_base_url,omitempty" yaml:"web_base_url,omitempty"`
	Path       string `json:"path" yaml:"path"`
}

// NewListCommand returns the `config list` subcommand.
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the values currently saved in the global config file",
		Long: `List the values currently saved in the global config file.

This does not show the BITRISE_WEB_BASE_URL environment override, or values
set only in a per-dir .bitrise-cli.yml — those apply at runtime but aren't
reflected here.`,
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
			v := configList{APIBaseURL: cfg.APIBaseURL, WebBaseURL: cfg.WebBaseURL, Path: p}

			if output.Format == output.FormatRaw {
				return printListHuman(cmd, v)
			}
			return output.Print(v, output.Format)
		},
	}

	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return cmd
}

func printListHuman(cmd *cobra.Command, v configList) error {
	w := cmd.OutOrStdout()
	_, err := fmt.Fprintf(w, "Path: %s\n\n%s: %s\n%s: %s\n",
		v.Path,
		internalconfig.KeyAPIBaseURL, unsetLabel(v.APIBaseURL),
		internalconfig.KeyWebBaseURL, unsetLabel(v.WebBaseURL),
	)
	return err
}

func unsetLabel(v string) string {
	if v == "" {
		return "(unset)"
	}
	return v
}
