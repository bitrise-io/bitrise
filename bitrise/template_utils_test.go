package bitrise

import (
	"os"
	"testing"

	"github.com/bitrise-io/bitrise-cli/models/models_1_0_0"
)

func TestEvaluateStepTemplateToBool(t *testing.T) {
	buildRes := models.BuildRunResultsModel{}

	propTempCont := `{{eq 1 1}}`
	t.Log("Simple true")
	isYes, err := EvaluateStepTemplateToBool(propTempCont, buildRes, false)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{eq 1 2}}`
	t.Log("Simple false")
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes, false)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}
}

func TestRegisteredFunctions(t *testing.T) {
	buildRes := models.BuildRunResultsModel{}

	propTempCont := `{{getenv "TEST_KEY" | eq "Test value"}}`
	t.Log("getenv - YES - propTempCont: ", propTempCont)
	if err := os.Setenv("TEST_KEY", "Test value"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err := EvaluateStepTemplateToBool(propTempCont, buildRes, false)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{getenv "TEST_KEY" | eq "A different value"}}`
	t.Log("getenv - NO - propTempCont: ", propTempCont)
	if err := os.Setenv("TEST_KEY", "Test value"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes, false)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{enveq "TEST_KEY" "enveq value"}}`
	t.Log("enveq - YES - propTempCont: ", propTempCont)
	if err := os.Setenv("TEST_KEY", "enveq value"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes, false)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{enveq "TEST_KEY" "different enveq value"}}`
	t.Log("enveq - NO - propTempCont: ", propTempCont)
	if err := os.Setenv("TEST_KEY", "enveq value"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes, false)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}
}

func TestRegisteredFlags(t *testing.T) {
	buildRes := models.BuildRunResultsModel{}

	propTempCont := `{{.IsCI}}`
	isCI := true
	t.Log("IsCI=true; propTempCont: ", propTempCont)
	if err := os.Setenv("CI", "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err := EvaluateStepTemplateToBool(propTempCont, buildRes, isCI)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{.IsCI}}`
	isCI = false
	t.Log("IsCI=fase; propTempCont: ", propTempCont)
	if err := os.Setenv("CI", "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes, isCI)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `$.IsCI`
	isCI = true
	t.Log("IsCI=true; short with $; propTempCont: ", propTempCont)
	if err := os.Setenv("CI", "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes, isCI)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `.IsCI`
	isCI = true
	t.Log("IsCI=true; short, no $; propTempCont: ", propTempCont)
	if err := os.Setenv("CI", "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes, isCI)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{not .IsCI}}`
	isCI = true
	t.Log("IsCI=true; NOT; propTempCont: ", propTempCont)
	if err := os.Setenv("CI", "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes, isCI)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	t.Log("Invalid - empty expression")
	isYes, err = EvaluateStepTemplateToBool("", buildRes, true)
	if err == nil {
		t.Fatal("Should return an error!")
	} else {
		t.Log("[expected] Error:", err)
	}
}
