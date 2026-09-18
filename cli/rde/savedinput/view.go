package savedinput

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
		Use:   "view SAVED_INPUT_ID",
		Short: "Show details of a single saved input",
		Args:  cmdutil.RequireArgs("SAVED_INPUT_ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogCommandParameters(cmd)

			if err := output.ConfigureOutputFormat(format); err != nil {
				return fmt.Errorf("failed to configure output format: %w", err)
			}

			client, err := cmdutil.NewRDEClient(cmd)
			if err != nil {
				return err
			}
			in, err := internalrde.NewService(client).GetSavedInput(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return output.Render(cmd.OutOrStdout(), output.Format, in, renderDetail)
		},
	}
	c.Flags().StringVarP(&format, cmdutil.FormatKey, "f", "", "Output format. Accepted: raw (default), json, yml")
	return c
}

func renderDetail(w io.Writer, in internalrde.SavedInput) error {
	s := style.New(w)
	ew := cmdutil.NewErrWriter(w)
	lbl := func(label string) string { return s.Label.Render(fmt.Sprintf("%-12s", label)) }

	ew.F("%s%s\n", lbl("Key:"), in.Key)
	ew.F("%s%s\n", lbl("ID:"), s.Slug.Render(in.ID))
	if in.IsSecret {
		ew.F("%s%s\n", lbl("Value:"), s.Dim.Render("(hidden)"))
		ew.F("%s%s\n", lbl("Secret:"), "yes")
	} else {
		ew.F("%s%s\n", lbl("Value:"), in.Value)
	}
	if in.UpdatedAt != nil {
		ew.F("%s%s\n", lbl("Updated:"), in.UpdatedAt.UTC().Format("2006-01-02 15:04 UTC"))
	}
	return ew.Err
}
