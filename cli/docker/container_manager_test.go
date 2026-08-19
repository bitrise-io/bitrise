package docker

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/stretchr/testify/require"
)

func Test_buildCreateArgs_optsKeyIsNotPassedAsEnv(t *testing.T) {
	isFalse := false
	env := envmanModels.EnvironmentItemModel{
		"MYSQL_ROOT_PASSWORD": "password",
		envmanModels.OptionsKey: envmanModels.EnvironmentItemOptionsModel{
			IsExpand: &isFalse,
		},
	}
	container := models.Container{
		Image: "mysql:8",
		Envs:  []envmanModels.EnvironmentItemModel{env},
	}

	args, err := buildCreateContainerCommandArgs(container, containerCreateOptions{name: "mysql"}, nil)
	require.NoError(t, err)
	require.Equal(t, []string{
		"create",
		"--platform", "linux/amd64",
		"--network=bitrise",
		"-e", "MYSQL_ROOT_PASSWORD=password",
		"--name=mysql",
		"mysql:8",
	}, args)
}
