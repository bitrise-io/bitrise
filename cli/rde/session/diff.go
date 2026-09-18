package session

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newDiffCmd() *cobra.Command {
	var format string
	c := &cobra.Command{
		Use:   "diff SESSION_ID",
		Short: "Compare a session's template snapshot with the current template",
		Long: `Show what changed between the template config snapshotted at the session's
creation time and the template's current config. Most useful when a session
reports template_outdated=true.

Lists which template variable keys changed (values are never exposed) and
the simple per-field differences (stack, machine type, scripts, working
directory). When the template was deleted, only the snapshot is shown.`,
		Args: cmdutil.RequireArgs("SESSION_ID"),
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
			sessionID, err := svc.ResolveSessionID(cmd.Context(), workspaceID, args[0])
			if err != nil {
				return err
			}
			diff, err := svc.DiffSessionTemplate(cmd.Context(), workspaceID, sessionID)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, diff, renderDiff)
		},
	}
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}

func renderDiff(w io.Writer, d internalrde.SessionTemplateDiff) error {
	s := style.New(w)
	ew := cmdutil.NewErrWriter(w)
	lbl := func(label string) string { return s.Label.Render(fmt.Sprintf("%-22s", label)) }

	if d.Current == nil {
		ew.Ln(s.Dim.Render("Current template: (deleted)"))
		ew.Ln()
	}

	if len(d.ChangedVariableKeys) == 0 && d.Snapshot != nil && d.Current != nil && configsEqual(*d.Snapshot, *d.Current) {
		ew.Ln(s.Dim.Render("No differences."))
		return ew.Err
	}

	if len(d.ChangedVariableKeys) > 0 {
		ew.F("%s%s\n", lbl("Changed variables:"), "")
		for _, k := range d.ChangedVariableKeys {
			ew.F("  %s\n", k)
		}
		ew.Ln()
	}

	if d.Snapshot != nil && d.Current != nil {
		writeFieldDiff(ew, lbl, "Stack", d.Snapshot.StackID, d.Current.StackID)
		writeFieldDiff(ew, lbl, "Machine type", d.Snapshot.MachineType, d.Current.MachineType)
		writeFieldDiff(ew, lbl, "Working dir", d.Snapshot.WorkingDirectory, d.Current.WorkingDirectory)
		writeBoolDiff(ew, lbl, "Startup script set", d.Snapshot.StartupScript != "", d.Current.StartupScript != "")
		writeBoolDiff(ew, lbl, "Warmup script set", d.Snapshot.WarmupScript != "", d.Current.WarmupScript != "")
		writeIntDiff(ew, lbl, "Session inputs", len(d.Snapshot.SessionInputs), len(d.Current.SessionInputs))
		writeIntDiff(ew, lbl, "Feature flags", len(d.Snapshot.FeatureFlags), len(d.Current.FeatureFlags))
	}
	return ew.Err
}

func writeFieldDiff(ew *cmdutil.ErrWriter, lbl func(string) string, name, before, after string) {
	if before == after {
		return
	}
	ew.F("%s%s → %s\n", lbl(name+":"), valueOrPlaceholder(before), valueOrPlaceholder(after))
}

func writeBoolDiff(ew *cmdutil.ErrWriter, lbl func(string) string, name string, before, after bool) {
	if before == after {
		return
	}
	ew.F("%s%t → %t\n", lbl(name+":"), before, after)
}

func writeIntDiff(ew *cmdutil.ErrWriter, lbl func(string) string, name string, before, after int) {
	if before == after {
		return
	}
	ew.F("%s%d → %d\n", lbl(name+":"), before, after)
}

func valueOrPlaceholder(v string) string {
	if v == "" {
		return "(unset)"
	}
	return v
}

// configsEqual reports whether two TemplateConfigs are equivalent on the
// fields the human diff inspects. Scripts compared by presence (not text)
// because the human output reports only "set/unset" for scripts.
func configsEqual(a, b internalrde.TemplateConfig) bool {
	return a.StackID == b.StackID &&
		a.MachineType == b.MachineType &&
		a.WorkingDirectory == b.WorkingDirectory &&
		(a.StartupScript != "") == (b.StartupScript != "") &&
		(a.WarmupScript != "") == (b.WarmupScript != "") &&
		len(a.SessionInputs) == len(b.SessionInputs) &&
		len(a.FeatureFlags) == len(b.FeatureFlags)
}
