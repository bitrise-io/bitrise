package template

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newViewCmd() *cobra.Command {
	var format string
	c := &cobra.Command{
		Use:   "view TEMPLATE_ID",
		Short: "Show details of a single template",
		Args:  cmdutil.RequireArgs("TEMPLATE_ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
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
			t, err := svc.GetTemplate(cmd.Context(), workspaceID, templateID)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, t, renderDetail)
		},
	}
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}

func renderDetail(w io.Writer, t internalrde.Template) error {
	s := style.New(w)
	ew := cmdutil.NewErrWriter(w)
	lbl := func(label string) string { return s.Label.Render(fmt.Sprintf("%-18s", label)) }

	ew.F("%s%s\n", lbl("Name:"), t.Name)
	ew.F("%s%s\n", lbl("ID:"), s.Slug.Render(t.ID))
	if t.Description != "" {
		ew.F("%s%s\n", lbl("Description:"), t.Description)
	}
	if t.StackID != "" {
		ew.F("%s%s\n", lbl("Stack:"), t.StackID)
	}
	if t.MachineType != "" {
		ew.F("%s%s\n", lbl("Machine type:"), t.MachineType)
	}
	if t.WorkingDirectory != "" {
		ew.F("%s%s\n", lbl("Working dir:"), t.WorkingDirectory)
	}
	if t.CreatedByEmail != "" {
		ew.F("%s%s\n", lbl("Owner:"), t.CreatedByEmail)
	}
	if t.UpdatedAt != nil {
		ew.F("%s%s\n", lbl("Updated:"), t.UpdatedAt.UTC().Format("2006-01-02 15:04 UTC"))
	}
	if len(t.SessionInputs) > 0 {
		ew.Ln()
		ew.Ln(s.Dim.Render("Session inputs"))
		for _, i := range t.SessionInputs {
			req := ""
			if i.Required {
				req = " (required)"
			}
			ew.F("  %s%s\n", i.Key, req)
			if i.Description != "" {
				ew.F("    %s\n", s.Dim.Render(i.Description))
			}
		}
	}
	if len(t.FeatureFlags) > 0 {
		ew.Ln()
		ew.Ln(s.Dim.Render("Feature flags"))
		for _, f := range t.FeatureFlags {
			ew.F("  %s\n", f.Name)
		}
	}
	if len(t.TemplateVariables) > 0 {
		ew.Ln()
		ew.Ln(s.Dim.Render("Template variables"))
		for _, v := range t.TemplateVariables {
			label := v.Key
			if v.IsSecret {
				label += " (secret)"
			}
			ew.F("  %s\n", label)
		}
	}
	return ew.Err
}
