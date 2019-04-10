package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GOWSConfigModel_WorkspaceForProjectLocation(t *testing.T) {

	gowsConfig := GOWSConfigModel{
		Workspaces: map[string]WorkspaceConfigModel{
			"/proj/path/1": WorkspaceConfigModel{
				WorkspaceRootPath: "/p1/ws/root",
			},
		},
	}

	t.Log("Found")
	{
		wsConfig, isFound := gowsConfig.WorkspaceForProjectLocation("/proj/path/1")
		require.Equal(t, true, isFound)
		require.Equal(t, "/p1/ws/root", wsConfig.WorkspaceRootPath)
	}

	t.Log("Not Found")
	{
		wsConfig, isFound := gowsConfig.WorkspaceForProjectLocation("/proj/path/noproj")
		require.Equal(t, false, isFound)
		require.Equal(t, "", wsConfig.WorkspaceRootPath)
	}
}
