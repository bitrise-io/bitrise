package session

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/output"
)

func newUpdateCmd() *cobra.Command {
	var (
		name                 string
		description          string
		autoTerminateMinutes int
		labels               []string
		unsetLabels          []string
		format               string
	)
	c := &cobra.Command{
		Use:   "update SESSION_ID",
		Short: "Update a session's name, description, auto-terminate duration, or labels",
		Long: `Update a session's name, description, auto-terminate duration, or labels.

Labels change incrementally: --label key=value upserts one label (an existing
key is overwritten, other keys are left untouched) and --unset-label key
removes one; both are repeatable. Removing a key the session doesn't have is
a no-op.`,
		Args: cmdutil.RequireArgs("SESSION_ID"),
		Example: `  bitrise rde session update SESSION_ID --name new-name
  bitrise rde session update SESSION_ID --auto-terminate-minutes 0
  bitrise rde session update SESSION_ID --label branch=main --unset-label wip`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			workspaceID, err := cmdutil.ResolveWorkspaceID(cmd)
			if err != nil {
				return err
			}
			req := internalrde.UpdateSessionRequest{}
			if cmd.Flags().Changed("name") {
				n := name
				req.Name = &n
			}
			if cmd.Flags().Changed("description") {
				d := description
				req.Description = &d
			}
			if cmd.Flags().Changed("auto-terminate-minutes") {
				m := autoTerminateMinutes
				req.AutoTerminateMinutes = &m
			}
			labelMap, err := parseLabelFlags("--label", labels)
			if err != nil {
				return err
			}
			req.Labels = labelMap
			// The backend resolves a key that is both set and removed in
			// favor of the removal; a direct contradiction on one command
			// line is a mistake worth stopping before the request goes out.
			for _, k := range unsetLabels {
				if k == "" {
					return fmt.Errorf("--unset-label: key must not be empty")
				}
				if _, ok := labelMap[k]; ok {
					return fmt.Errorf("label %q cannot be both set (--label) and removed (--unset-label)", k)
				}
			}
			req.RemoveLabels = unsetLabels
			if req.Name == nil && req.Description == nil && req.AutoTerminateMinutes == nil && len(req.Labels) == 0 && len(req.RemoveLabels) == 0 {
				return fmt.Errorf("at least one of --name, --description, --auto-terminate-minutes, --label, --unset-label is required")
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
			sess, err := svc.UpdateSession(cmd.Context(), workspaceID, sessionID, req)
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, sess, renderSessionDetail)
		},
	}
	c.Flags().StringVar(&name, "name", "", "new session name")
	c.Flags().StringVar(&description, "description", "", "new session description")
	c.Flags().IntVar(&autoTerminateMinutes, "auto-terminate-minutes", 0, "auto-terminate duration in minutes; 0 disables. Resets the deadline to now + minutes.")
	c.Flags().StringArrayVarP(&labels, "label", "l", nil, "label to set on the session as key=value (repeatable; merged into the existing labels)")
	c.Flags().StringArrayVar(&unsetLabels, "unset-label", nil, "label key to remove from the session (repeatable; unknown keys are ignored)")
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}
