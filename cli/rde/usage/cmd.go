package usage

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	internalrde "github.com/bitrise-io/bitrise/v2/internal/rde"
	"github.com/bitrise-io/bitrise/v2/internal/style"
	"github.com/bitrise-io/bitrise/v2/output"
)

// NewCmd returns the `rde usage` command. A leaf command, not a noun with
// subcommands: the report is a single summary, not a listable collection.
func NewCmd() *cobra.Command {
	var format string
	c := &cobra.Command{
		Use:   "usage",
		Short: "Show the workspace's active session and resource usage",
		Long: `Show a point-in-time snapshot of the workspace's active remote dev sessions:
session counts and vCPU/memory totals split by OS, workspace-wide and per user.

This reports sessions currently consuming resources; it is not a historical or
billing-period report. Requires the workspace's billing-view permission
(workspace owners and billing-managing custom roles).`,
		Example: `  bitrise rde usage
  bitrise rde usage --format json | jq '.totals'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
			res, err := internalrde.NewService(client).GetWorkspaceUsage(cmd.Context(), workspaceID)
			if err != nil {
				return err
			}
			if err := output.Render(cmd.OutOrStdout(), output.Format, res, renderUsage); err != nil {
				return err
			}
			// Diagnostics go to stderr so piped stdout stays parse-safe; the
			// JSON shape carries unknown_machine_type_count as data anyway.
			if res.UnknownMachineTypeCount > 0 {
				s := style.New(cmd.ErrOrStderr())
				ew := cmdutil.NewErrWriter(cmd.ErrOrStderr())
				ew.F("%s active sessions with unrecognized machine types contribute 0 vCPU/memory to the totals, so the sums may undercount (%d affected)\n",
					s.Warn.Render("Warning:"), res.UnknownMachineTypeCount)
				return ew.Err
			}
			return nil
		},
	}
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}

func renderUsage(w io.Writer, res internalrde.WorkspaceUsage) error {
	total := sumBuckets(res.Totals)
	if total.SessionCount == 0 {
		_, err := fmt.Fprintln(w, "No active sessions.")
		return err
	}

	s := style.New(w)
	ew := cmdutil.NewErrWriter(w)
	lbl := func(label string) string {
		return s.Label.Render(fmt.Sprintf("%-12s", label))
	}
	ew.F("%s%-6d%s\n", lbl("Sessions:"), total.SessionCount, osSplit(res.Totals, func(p internalrde.PlatformUsage) int32 { return p.SessionCount }))
	ew.F("%s%-6d%s\n", lbl("vCPU:"), total.VCPU, osSplit(res.Totals, func(p internalrde.PlatformUsage) int32 { return p.VCPU }))
	ew.F("%s%-6d%s\n", lbl("Memory GB:"), total.MemoryGB, osSplit(res.Totals, func(p internalrde.PlatformUsage) int32 { return p.MemoryGB }))

	if len(res.Users) > 0 {
		ew.F("\n")
		if err := renderUserTable(w, s, res.Users); err != nil {
			return err
		}
	}
	return ew.Err
}

// sumBuckets folds the linux/macos/unknown buckets into one grand total.
func sumBuckets(t internalrde.UsageTotals) internalrde.PlatformUsage {
	return internalrde.PlatformUsage{
		SessionCount: t.Linux.SessionCount + t.Macos.SessionCount + t.Unknown.SessionCount,
		VCPU:         t.Linux.VCPU + t.Macos.VCPU + t.Unknown.VCPU,
		MemoryGB:     t.Linux.MemoryGB + t.Macos.MemoryGB + t.Unknown.MemoryGB,
	}
}

// osSplit renders the per-OS breakdown of one metric, e.g.
// "(Linux 64, macOS 24)"; the unknown bucket appears only when non-zero.
func osSplit(t internalrde.UsageTotals, metric func(internalrde.PlatformUsage) int32) string {
	out := fmt.Sprintf("(Linux %d, macOS %d", metric(t.Linux), metric(t.Macos))
	if v := metric(t.Unknown); v != 0 {
		out += fmt.Sprintf(", unknown %d", v)
	}
	return out + ")"
}

func renderUserTable(w io.Writer, s style.Styles, users []internalrde.UserUsage) error {
	const colUser = 0
	// The unknown-OS column appears only when some row needs it, so the
	// common all-known case stays narrow — but when present it lets every
	// row's session total reconcile with its per-OS resources.
	showUnknown := false
	for _, u := range users {
		if u.Totals.Unknown != (internalrde.PlatformUsage{}) {
			showUnknown = true
			break
		}
	}
	headers := []string{"USER", "SESSIONS", "LINUX VCPU/GB", "MACOS VCPU/GB"}
	if showUnknown {
		headers = append(headers, "UNKNOWN VCPU/GB")
	}
	rows := make([][]string, 0, len(users))
	for _, u := range users {
		row := []string{
			userLabel(u),
			strconv.Itoa(int(sumBuckets(u.Totals).SessionCount)),
			vcpuGB(u.Totals.Linux),
			vcpuGB(u.Totals.Macos),
		}
		if showUnknown {
			row = append(row, vcpuGB(u.Totals.Unknown))
		}
		rows = append(rows, row)
	}
	styler := func(row, col int, content string) string {
		if col == colUser && users[row].IsWorkspace {
			return s.Dim.Render(content)
		}
		return content
	}
	return style.Table(w, headers, rows, s.Header, styler)
}

// userLabel identifies a breakdown row: email when known, the workspace
// bucket as "(workspace)", with username/ID fallbacks for user rows
// missing an email.
func userLabel(u internalrde.UserUsage) string {
	switch {
	case u.IsWorkspace:
		return "(workspace)"
	case u.Email != "":
		return u.Email
	case u.Username != "":
		return u.Username
	default:
		return u.UserID
	}
}

func vcpuGB(p internalrde.PlatformUsage) string {
	return fmt.Sprintf("%d / %d", p.VCPU, p.MemoryGB)
}
