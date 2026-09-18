package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsQuiet(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().BoolP(FlagQuiet, "q", false, "quiet")
	child := &cobra.Command{Use: "child"}
	root.AddCommand(child)

	assert.False(t, IsQuiet(child), "unset defaults to false")

	require.NoError(t, root.PersistentFlags().Set(FlagQuiet, "true"))
	assert.True(t, IsQuiet(child), "reads the flag off the root, not the child")
}
