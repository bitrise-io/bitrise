package cmdtest

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bitrise-io/bitrise/v2/internal/config"
	"github.com/bitrise-io/bitrise/v2/output"
)

func TestRun_WiresConfigFormatArgsAndStdin(t *testing.T) {
	var resolved config.Resolved
	cmd := &cobra.Command{
		Use: "echo",
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved = config.FromContext(cmd.Context())
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), strings.Join(args, ","))
			in, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(cmd.ErrOrStderr(), string(in))
			return nil
		},
	}
	cmd.Flags().StringP(output.FormatKey, "f", output.FormatRaw, "output format")

	stdout, stderr, err := Run(t, cmd, Opts{
		Args:               []string{"hello", "world"},
		RDEAPIBaseURL:      "https://rde.example.test",
		DefaultWorkspaceID: "test-ws",
		Format:             output.FormatJSON,
		Stdin:              "piped-in",
	})

	require.NoError(t, err)
	assert.Equal(t, "hello,world\n", stdout)
	assert.Equal(t, "piped-in", stderr)
	assert.Equal(t, "https://rde.example.test", resolved.RDEAPIBaseURL)
	assert.Equal(t, "test-ws", resolved.DefaultWorkspaceID)

	got, gerr := cmd.Flags().GetString(output.FormatKey)
	require.NoError(t, gerr)
	assert.Equal(t, output.FormatJSON, got)
}

func TestRun_EmptyFormatLeavesFlagDefault(t *testing.T) {
	cmd := &cobra.Command{
		Use:  "echo",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	cmd.Flags().StringP(output.FormatKey, "f", output.FormatRaw, "output format")

	_, _, err := Run(t, cmd, Opts{})
	require.NoError(t, err)

	got, gerr := cmd.Flags().GetString(output.FormatKey)
	require.NoError(t, gerr)
	assert.Equal(t, output.FormatRaw, got, "empty opts.Format must not touch the flag")
}
