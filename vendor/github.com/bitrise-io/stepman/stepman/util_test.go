package stepman

import (
	"testing"

	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/stepman/models"
	"github.com/stretchr/testify/require"
)

func TestAddStepVersionToStepGroup(t *testing.T) {
	step := models.StepModel{
		Title: pointers.NewStringPtr("name 1"),
	}

	group := models.StepGroupModel{
		Versions: map[string]models.StepModel{
			"1.0.0": step,
			"2.0.0": step,
		},
		LatestVersionNumber: "2.0.0",
	}

	group, err := addStepVersionToStepGroup(step, "2.1.0", group)
	require.Equal(t, nil, err)
	require.Equal(t, 3, len(group.Versions))
	require.Equal(t, "2.1.0", group.LatestVersionNumber)
}
