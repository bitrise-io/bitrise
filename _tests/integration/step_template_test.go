package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_StepTemplate(t *testing.T) {
	start := time.Now().Unix()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().Unix()))
	}(start)

	configPth := "bitrise.yml"

	t.Log("step template test")
	{
		cmd := command.New(binPath(), "run", "test", "--config", configPth)
		cmd.SetDir("step_template")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("bash toolkit step template test")
	{
		cmd := command.New(binPath(), "run", "test", "--config", configPth)
		cmd.SetDir("bash_toolkit_step_template")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}

	t.Log("go toolkit step template test")
	{
		cmd := command.New(binPath(), "run", "test", "--config", configPth)
		cmd.SetDir("go_toolkit_step_template")
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		require.NoError(t, err, out)
	}
}
