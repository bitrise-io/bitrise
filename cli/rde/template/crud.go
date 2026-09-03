package template

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newCreateCmd() *cobra.Command {
	var (
		file   string
		format string
	)
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a new RDE template from a JSON spec file",
		Long: `Create a new RDE template from a JSON spec file. See the annotated
example at the bottom for the JSON shape — or, once you have a template
you like, round-trip from it (audit fields like id/created_at are ignored):

  bitrise rde template view OTHER_TEMPLATE_ID --format json > template.json
  # edit template.json
  bitrise rde template create --file template.json

Pass --file - to read the JSON from stdin.

Required fields:
  name           template name
  stack_id       stack ID (see 'rde stack list')
  machine_type   machine type name (see 'rde machine-type list')

Optional fields:
  description         free-text description
  working_directory   default working directory for sessions
  startup_script      bash script run on every session start
  warmup_script       bash script run once during pre-warming
  session_inputs      array of {key, description, required, default_value,
                      expose_as_env_var} — values callers supply at create time
  template_variables  array of {key, value, is_secret, expose_as_env_var} —
                      baked-in values available to startup/warmup scripts
  feature_flags       array of {name, description}
  workspace_links     array of {label, folder_path, feature_flag_name} — IDE
                      folder shortcuts

Example spec exercising every field (a macOS iOS-app dev environment —
adjust to taste):

  {
    "name": "example-ios-app",
    "description": "Example macOS dev environment for an iOS app.",
    "stack_id": "osx-xcode-16.0.x-edge",
    "machine_type": "g2.mac.m2pro.6c-14g",
    "working_directory": "/Users/vagrant/git",
    "warmup_script": "set -euo pipefail\ncd ~\ngit clone \"https://${GITHUB_PAT}@github.com/example-org/example-app.git\" git\ncd git && bundle install && pod install --project-directory=ios\n",
    "startup_script": "set -euo pipefail\ncd /Users/vagrant/git\ngit pull --ff-only || true\nsudo xcode-select -s \"/Applications/Xcode-${XCODE_VERSION}.app\"\n",
    "session_inputs": [
      {
        "key": "GITHUB_PAT",
        "description": "GitHub PAT with read access to example-org/example-app",
        "required": true,
        "expose_as_env_var": true
      },
      {
        "key": "XCODE_VERSION",
        "description": "Xcode version to select via xcode-select",
        "default_value": "26.3",
        "expose_as_env_var": true
      }
    ],
    "template_variables": [
      {"key": "APP_SCHEME", "value": "ExampleApp", "expose_as_env_var": true},
      {"key": "FASTLANE_API_KEY", "is_secret": true, "expose_as_env_var": true}
    ],
    "feature_flags": [
      {"name": "enable_beta_simulator", "description": "Boot the iOS beta simulator on session start"}
    ],
    "workspace_links": [
      {"label": "Open app in Xcode", "folder_path": "/Users/vagrant/git/ios"},
      {"label": "Open scripts (beta only)", "folder_path": "/Users/vagrant/git/scripts", "feature_flag_name": "enable_beta_simulator"}
    ]
  }`,
		Example: `  bitrise rde template create --file template.json
  cat template.json | bitrise rde template create --file -`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			if file == "" {
				return fmt.Errorf("--file is required")
			}
			spec, err := readTemplateSpec(cmd.InOrStdin(), file)
			if err != nil {
				return err
			}
			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			t, err := internalrde.NewService(client).CreateTemplate(cmd.Context(), workspaceID, spec)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, t, renderDetail)
		},
	}
	c.Flags().StringVarP(&file, "file", "f", "", "path to a JSON spec file (use '-' for stdin)")
	// No -f shorthand: --file owns it on this command.
	c.Flags().StringVar(&format, cmdutil.FormatKey, "", "Output format. Accepted: raw (default), json, yml")
	return c
}

func newUpdateCmd() *cobra.Command {
	var (
		file   string
		format string
	)
	c := &cobra.Command{
		Use:   "update TEMPLATE_ID",
		Short: "Update an existing RDE template from a JSON spec file",
		Long: `Update an existing RDE template from a JSON spec file.

Only fields present in the file are sent. Array fields (template_variables,
session_inputs, feature_flags, workspace_links) replace the server's
existing list wholesale when present — to clear one, include it as [].

Round-trip workflow:

  bitrise rde template view TEMPLATE_ID --format json > template.json
  # edit template.json
  bitrise rde template update TEMPLATE_ID --file template.json

Pass --file - to read the JSON from stdin.`,
		Args: cmdutil.RequireArgs("TEMPLATE_ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			if file == "" {
				return fmt.Errorf("--file is required")
			}
			spec, err := readTemplateSpec(cmd.InOrStdin(), file)
			if err != nil {
				return err
			}
			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			svc := internalrde.NewService(client)
			templateID, err := svc.ResolveTemplateID(cmd.Context(), workspaceID, args[0])
			if err != nil {
				return err
			}
			t, err := svc.UpdateTemplate(cmd.Context(), workspaceID, templateID, spec)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, t, renderDetail)
		},
	}
	c.Flags().StringVarP(&file, "file", "f", "", "path to a JSON spec file (use '-' for stdin)")
	// No -f shorthand: --file owns it on this command.
	c.Flags().StringVar(&format, cmdutil.FormatKey, "", "Output format. Accepted: raw (default), json, yml")
	return c
}

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete TEMPLATE_ID",
		Short: "Delete an RDE template",
		Long: `Delete an RDE template (soft-delete server-side). Existing sessions
created from this template keep working — they reference a snapshot — but
the template can no longer be selected for new sessions.`,
		Args: cmdutil.RequireArgs("TEMPLATE_ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			svc := internalrde.NewService(client)
			templateID, err := svc.ResolveTemplateID(cmd.Context(), workspaceID, args[0])
			if err != nil {
				return err
			}
			if err := svc.DeleteTemplate(cmd.Context(), workspaceID, templateID); err != nil {
				return err
			}
			if !cmdutil.IsQuiet(cmd) {
				_, err := fmt.Fprintf(cmd.ErrOrStderr(), "Deleted template %s\n", templateID)
				return err
			}
			return nil
		},
	}
}

// readTemplateSpec reads a TemplateSpec from path. "-" reads from stdin.
// Extra fields (id, created_at, etc.) are ignored — encoding/json drops
// them naturally — so view-output can be piped straight back in.
func readTemplateSpec(stdin io.Reader, path string) (internalrde.TemplateSpec, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return internalrde.TemplateSpec{}, fmt.Errorf("read template spec: %w", err)
	}
	var spec internalrde.TemplateSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return internalrde.TemplateSpec{}, fmt.Errorf("parse template spec: %w", err)
	}
	return spec, nil
}
