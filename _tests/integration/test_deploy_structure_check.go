package integration

import (
	"testing"

	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

func Test_TestDeployDirStructure(t *testing.T) {
	configPth := "test_deploy_structure_check.yml"

	cmd := command.New(binPath(), "run", "test-deploy-dir-structure-check", "--config", configPth)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	require.NoError(t, err, out)
}
