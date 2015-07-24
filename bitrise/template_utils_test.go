package bitrise

import (
	"os"
	"testing"

	"github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
)

func TestEvaluateStepTemplateToBool(t *testing.T) {
	buildRes := models.StepRunResultsModel{}

	propTempCont := `{{eq 1 1}}`
	t.Log("Simple true")
	isYes, err := EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{eq 1 2}}`
	t.Log("Simple false")
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}
}

func TestRegisteredFunctions(t *testing.T) {
	buildRes := models.StepRunResultsModel{}

	propTempCont := `{{getenv "CI" | eq "true"}}`
	t.Log("getenv")
	if err := os.Setenv("CI", "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err := EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}
}
