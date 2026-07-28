package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelegateToList_RunsListSubcommand(t *testing.T) {
	var ranWithContext bool
	parent := &cobra.Command{Use: "stack", RunE: DelegateToList}
	list := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ranWithContext = cmd.Context() == parent.Context()
			return nil
		},
	}
	parent.AddCommand(list)
	parent.SetContext(t.Context())

	require.NoError(t, DelegateToList(parent, nil))
	assert.True(t, ranWithContext, "list subcommand should run with the parent's context")
}

func TestDelegateToList_FallsBackToHelpWhenNoListSubcommand(t *testing.T) {
	var helpShown bool
	parent := &cobra.Command{
		Use:  "stack",
		RunE: DelegateToList,
	}
	parent.SetHelpFunc(func(_ *cobra.Command, _ []string) { helpShown = true })

	require.NoError(t, DelegateToList(parent, nil))
	assert.True(t, helpShown, "expected cmd.Help() to be called")
}
