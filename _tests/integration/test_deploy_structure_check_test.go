package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
)

func Test_DeployDirStructure(t *testing.T) {
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
