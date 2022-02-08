package bitrise

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXcodeDependency(t *testing.T) {
	error := DependencyTryCheckTool("xxccode")
	require.NotNil(t, error)

	error = DependencyTryCheckTool("xcode")
	require.Nil(t, error)
}

func TestXcodeVersion(t *testing.T) {
	error := PrintInstalledXcodeInfos()

	require.Nil(t, error)
}
