package bitrise

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXcodeDependency(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.SkipNow()

		return
	}

	error := DependencyTryCheckTool("xcode")
	require.Nil(t, error)
}

func TestXcodeDependency_errorOut(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.SkipNow()

		return
	}

	error := DependencyTryCheckTool("xxccode")
	require.NotNil(t, error)
}

func TestXcodeVersion(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.SkipNow()

		return
	}

	error := PrintInstalledXcodeInfos()

	require.Nil(t, error)
}
