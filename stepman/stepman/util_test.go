package stepman

import (
	"testing"

	"github.com/bitrise-io/stepman/models"
)

var (
	title = "name 1"
)

func TestAddStepVersionToStepGroup(t *testing.T) {
	step := models.StepModel{
		Title: &title,
	}

	group := models.StepGroupModel{
		Versions: map[string]models.StepModel{
			"1.0.0": step,
			"2.0.0": step,
		},
		LatestVersionNumber: "2.0.0",
	}

	group, err := addStepVersionToStepGroup(step, "2.1.0", group)
	if err != nil {
		t.Fatal(err)
	}
	if len(group.Versions) != 3 {
		t.Fatal("Failed to add new version")
	}
	if group.LatestVersionNumber != "2.1.0" {
		t.Fatal("Failed to set latest version")
	}
}
