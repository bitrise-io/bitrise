package bitrise

import (
	"os"
	"testing"

	"github.com/bitrise-io/bitrise/models"
)

func TestEvaluateStepTemplateToBool(t *testing.T) {
	buildRes := models.BuildRunResultsModel{}

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

	t.Log("Invalid - empty expression")
	isYes, err = EvaluateStepTemplateToBool("", buildRes)
	if err == nil {
		t.Fatal("Should return an error!")
	} else {
		t.Log("[expected] Error:", err)
	}
}

func TestRegisteredFunctions(t *testing.T) {
	defer func() {
		// env cleanup
		if err := os.Unsetenv("TEST_KEY"); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()

	buildRes := models.BuildRunResultsModel{}

	propTempCont := `{{getenv "TEST_KEY" | eq "Test value"}}`
	t.Log("getenv - YES - propTempCont: ", propTempCont)
	if err := os.Setenv("TEST_KEY", "Test value"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err := EvaluateStepTemplateToBool(propTempCont, buildRes)
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
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
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
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
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
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}
}

func TestCIFlagsAndEnvs(t *testing.T) {
	defer func() {
		// env cleanup
		if err := os.Unsetenv(CIModeEnvKey); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()

	buildRes := models.BuildRunResultsModel{}

	propTempCont := `{{.IsCI}}`
	t.Log("IsCI=true; propTempCont: ", propTempCont)
	if err := os.Setenv(CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err := EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{.IsCI}}`
	t.Log("IsCI=fase; propTempCont: ", propTempCont)
	if err := os.Setenv(CIModeEnvKey, "false"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{.IsCI}}`
	t.Log("[unset] IsCI; propTempCont: ", propTempCont)
	if err := os.Unsetenv(CIModeEnvKey); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `$.IsCI`
	t.Log("IsCI=true; short with $; propTempCont: ", propTempCont)
	if err := os.Setenv(CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `.IsCI`
	t.Log("IsCI=true; short, no $; propTempCont: ", propTempCont)
	if err := os.Setenv(CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `not .IsCI`
	t.Log("IsCI=true; NOT; propTempCont: ", propTempCont)
	if err := os.Setenv(CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `not .IsCI`
	t.Log("IsCI=false; NOT; propTempCont: ", propTempCont)
	if err := os.Setenv(CIModeEnvKey, "false"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}
}

func TestPullRequestFlagsAndEnvs(t *testing.T) {
	defer func() {
		// env cleanup
		if err := os.Unsetenv(PullRequestIDEnvKey); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()

	buildRes := models.BuildRunResultsModel{}

	propTempCont := `{{.IsPR}}`
	t.Log("IsPR [undefined]; propTempCont: ", propTempCont)
	isYes, err := EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{.IsPR}}`
	t.Log("IsPR=true; propTempCont: ", propTempCont)
	if err := os.Setenv(PullRequestIDEnvKey, "123"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}
}

func TestPullRequestAndCIFlagsAndEnvs(t *testing.T) {
	defer func() {
		// env cleanup
		if err := os.Unsetenv(PullRequestIDEnvKey); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
		if err := os.Unsetenv(CIModeEnvKey); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()

	buildRes := models.BuildRunResultsModel{}

	propTempCont := `not .IsPR | and (not .IsCI)`
	t.Log("IsPR [undefined] & IsCI [undefined]; propTempCont: ", propTempCont)
	isYes, err := EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `not .IsPR | and .IsCI`
	t.Log("IsPR [undefined] & IsCI [undefined]; propTempCont: ", propTempCont)
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `not .IsPR | and .IsCI`
	t.Log("IsPR [undefined] & IsCI=true; propTempCont: ", propTempCont)
	if err := os.Setenv(CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `.IsPR | and .IsCI`
	t.Log("IsPR [undefined] & IsCI=true; propTempCont: ", propTempCont)
	if err := os.Setenv(CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `.IsPR | and .IsCI`
	t.Log("IsPR=true & IsCI=true; propTempCont: ", propTempCont)
	if err := os.Setenv(PullRequestIDEnvKey, "123"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	if err := os.Setenv(CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateStepTemplateToBool(propTempCont, buildRes)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}
}
