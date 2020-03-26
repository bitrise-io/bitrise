package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise/debug"
	"github.com/bitrise-io/go-utils/command"
)

func Test_DeployDirStructure(t *testing.T) {
	start := time.Now().UnixNano()
	defer func(s int64) {
		debug.W(fmt.Sprintf("[ '%s', %d, %d ],\n", t.Name(), start, time.Now().UnixNano()))
	}(start)

	t.Log("check test deploy dir structure")
	{
		configPth := "test_deploy_structure_check.yml"

		cmd := command.New(binPath(), "run", "test-deploy-dir-structure-check", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			t.Fatal(err, out)
		}
	}
}
