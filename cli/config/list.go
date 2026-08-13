package config

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalconfig "github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewListCommand returns the `config list` subcommand.
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the values currently saved in the global config file",
		Long: `List the values currently saved in the global config file.

This shows what is stored, not what every command will resolve: the
BITRISE_WEB_BASE_URL, BITRISE_APP_ID, BITRISE_APP_SLUG and BITRISE_WORKSPACE_ID
environment variables, and an app_id or default_workspace_id pinned by a
per-directory .bitrise-cli.yml, all take precedence at runtime. Inside a Bitrise
build, BITRISE_APP_SLUG and BITRISE_WORKSPACE_ID are always set.`,
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
			values, err := configValues(cfg)
			if err != nil {
				return err
			}
			// No omitempty on any key: this command's job is to enumerate every
			// recognized key's state, so an unset key must still appear (as "")
			// rather than vanish from the JSON/YAML output.
			values["path"] = p

			return output.Render(cmd.OutOrStdout(), output.Format, values, func(w io.Writer, values map[string]string) error {
				return printListHuman(w, p, values)
			})
		},
	}

	cmd.Flags().StringP(cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return cmd
}

// configValues reads every key in internalconfig.Keys out of cfg, keyed by
// its config-key name, so a key added to Keys shows up here with no matching
// struct field to add.
func configValues(cfg internalconfig.Config) (map[string]string, error) {
	values := make(map[string]string, len(internalconfig.Keys))
	for _, key := range internalconfig.Keys {
		v, err := cfg.Get(key)
		if err != nil {
			return nil, err
		}
		values[key] = v
	}
	return values, nil
}

func printListHuman(w io.Writer, path string, values map[string]string) error {
	if _, err := fmt.Fprintf(w, "Path: %s\n\n", path); err != nil {
		return err
	}
	for _, key := range internalconfig.Keys {
		if _, err := fmt.Fprintf(w, "%s: %s\n", key, unsetLabel(values[key])); err != nil {
			return err
		}
	}
	return nil
}

func unsetLabel(v string) string {
	if v == "" {
		return "(unset)"
	}
	return v
}
