package toolkits

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/bitrise/models"
	"github.com/stretchr/testify/require"
)

func Test_stepBinaryFilename(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	{
		sIDData := models.StepIDData{SteplibSource: "path", IDorURI: "./", Version: ""}
		require.Equal(t, "path-._-", stepBinaryFilename(sIDData))
	}

	{
		sIDData := models.StepIDData{SteplibSource: "git", IDorURI: "https://github.com/bitrise-steplib/steps-go-toolkit-hello-world.git", Version: "master"}
		require.Equal(t, "git-https___github.com_bitrise-steplib_steps-go-toolkit-hello-world.git-master", stepBinaryFilename(sIDData))
	}

	{
		sIDData := models.StepIDData{SteplibSource: "_", IDorURI: "https://github.com/bitrise-steplib/steps-go-toolkit-hello-world.git", Version: "master"}
		require.Equal(t, "_-https___github.com_bitrise-steplib_steps-go-toolkit-hello-world.git-master", stepBinaryFilename(sIDData))
	}

	{
		sIDData := models.StepIDData{SteplibSource: "https://github.com/bitrise-io/bitrise-steplib.git", IDorURI: "script", Version: "1.2.3"}
		require.Equal(t, "https___github.com_bitrise-io_bitrise-steplib.git-script-1.2.3", stepBinaryFilename(sIDData))
	}
}

func Test_parseGoVersionFromGoVersionOutput(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	t.Log("Example OK")
	{
		verStr, err := parseGoVersionFromGoVersionOutput("go version go1.7 darwin/amd64")
		require.NoError(t, err)
		require.Equal(t, "1.7", verStr)
	}

	t.Log("Example OK 2")
	{
		verStr, err := parseGoVersionFromGoVersionOutput(`go version go1.7 darwin/amd64

`)
		require.NoError(t, err)
		require.Equal(t, "1.7", verStr)
	}

	t.Log("Example OK 3")
	{
		verStr, err := parseGoVersionFromGoVersionOutput("go version go1.7.1 darwin/amd64")
		require.NoError(t, err)
		require.Equal(t, "1.7.1", verStr)
	}

	t.Log("Empty")
	{
		verStr, err := parseGoVersionFromGoVersionOutput("")
		require.EqualError(t, err, "Failed to parse Go version, error: version call output was empty")
		require.Equal(t, "", verStr)
	}

	t.Log("Empty 2")
	{
		verStr, err := parseGoVersionFromGoVersionOutput(`

`)
		require.EqualError(t, err, "Failed to parse Go version, error: version call output was empty")
		require.Equal(t, "", verStr)
	}

	t.Log("Invalid")
	{
		verStr, err := parseGoVersionFromGoVersionOutput("go version REMOVED darwin/amd64")
		require.EqualError(t, err, "Failed to parse Go version, error: failed to find version in input: go version REMOVED darwin/amd64")
		require.Equal(t, "", verStr)
	}
}
