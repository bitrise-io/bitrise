//go:build linux_and_mac
// +build linux_and_mac

package workflow

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
)

func Test_DeployDirStructure(t *testing.T) {
	t.Log("check test deploy dir structure")
	{
		configPth := "test_deploy_structure_check.yml"

		cmd := command.New(testhelpers.BinPath(), "run", "test-deploy-dir-structure-check", "--config", configPth)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			t.Fatal(err, out)
		}
	}
}
