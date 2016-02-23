package bitrise

import (
	"os"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise/configs"
	"github.com/bitrise-io/bitrise/models"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func TestEvaluateStepTemplateToBool(t *testing.T) {
	buildRes := models.BuildRunResultsModel{}

	propTempCont := `{{eq 1 1}}`
	isYes, err := EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, isYes)

	propTempCont = `{{eq 1 2}}`
	isYes, err = EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, false, isYes)

	isYes, err = EvaluateTemplateToBool("", false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.NotEqual(t, nil, err)

	// these all should be `true`
	for _, expStr := range []string{
		"true",
		"1",
		`"yes"`,
		`"YES"`,
		`"Yes"`,
		`"YeS"`,
		`"TRUE"`,
		`"True"`,
		`"TrUe"`,
		`"y"`,
	} {
		isYes, err = EvaluateTemplateToBool(expStr, false, false, buildRes, envmanModels.EnvsJSONListModel{})
		require.NoError(t, err)
		require.Equal(t, true, isYes)
	}

	// these all should be `true`
	for _, expStr := range []string{
		"false",
		"0",
		`"no"`,
		`"NO"`,
		`"No"`,
		`"FALSE"`,
		`"False"`,
		`"FaLse"`,
		`"n"`,
	} {
		isYes, err = EvaluateTemplateToBool(expStr, false, false, buildRes, envmanModels.EnvsJSONListModel{})
		require.NoError(t, err)
		require.Equal(t, false, isYes)
	}
}

func TestRegisteredFunctions(t *testing.T) {
	defer func() {
		// env cleanup
		require.Equal(t, nil, os.Unsetenv("TEST_KEY"))
	}()

	buildRes := models.BuildRunResultsModel{}

	propTempCont := `{{getenv "TEST_KEY" | eq "Test value"}}`
	require.Equal(t, nil, os.Setenv("TEST_KEY", "Test value"))
	isYes, err := EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, isYes)

	propTempCont = `{{getenv "TEST_KEY" | eq "A different value"}}`
	require.Equal(t, nil, os.Setenv("TEST_KEY", "Test value"))
	isYes, err = EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, false, isYes)

	propTempCont = `{{enveq "TEST_KEY" "enveq value"}}`
	require.Equal(t, nil, os.Setenv("TEST_KEY", "enveq value"))
	isYes, err = EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, isYes)

	propTempCont = `{{enveq "TEST_KEY" "different enveq value"}}`
	require.Equal(t, nil, os.Setenv("TEST_KEY", "enveq value"))
	isYes, err = EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, false, isYes)
}

func TestCIFlagsAndEnvs(t *testing.T) {
	defer func() {
		// env cleanup
		if err := os.Unsetenv(configs.CIModeEnvKey); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()

	buildRes := models.BuildRunResultsModel{}

	propTempCont := `{{.IsCI}}`
	t.Log("IsCI=true; propTempCont: ", propTempCont)
	if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err := EvaluateTemplateToBool(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{.IsCI}}`
	t.Log("IsCI=fase; propTempCont: ", propTempCont)
	if err := os.Setenv(configs.CIModeEnvKey, "false"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{.IsCI}}`
	t.Log("[unset] IsCI; propTempCont: ", propTempCont)
	if err := os.Unsetenv(configs.CIModeEnvKey); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `$.IsCI`
	t.Log("IsCI=true; short with $; propTempCont: ", propTempCont)
	if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateTemplateToBool(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `.IsCI`
	t.Log("IsCI=true; short, no $; propTempCont: ", propTempCont)
	if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env!")
	}
	isYes, err = EvaluateTemplateToBool(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `not .IsCI`
	t.Log("IsCI=true; NOT; propTempCont: ", propTempCont)
	if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateTemplateToBool(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `not .IsCI`
	t.Log("IsCI=false; NOT; propTempCont: ", propTempCont)
	if err := os.Setenv(configs.CIModeEnvKey, "false"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
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
		if err := os.Unsetenv(configs.PullRequestIDEnvKey); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()

	// env cleanup
	if err := os.Unsetenv(configs.PullRequestIDEnvKey); err != nil {
		t.Error("Failed to unset environment: ", err)
	}

	buildRes := models.BuildRunResultsModel{}

	propTempCont := `{{.IsPR}}`
	t.Log("IsPR [undefined]; propTempCont: ", propTempCont)
	isYes, err := EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `{{.IsPR}}`
	t.Log("IsPR=true; propTempCont: ", propTempCont)
	if err := os.Setenv(configs.PullRequestIDEnvKey, "123"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateTemplateToBool(propTempCont, false, true, buildRes, envmanModels.EnvsJSONListModel{})
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
		if err := os.Unsetenv(configs.PullRequestIDEnvKey); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
		if err := os.Unsetenv(configs.CIModeEnvKey); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
	}()

	buildRes := models.BuildRunResultsModel{}

	propTempCont := `not .IsPR | and (not .IsCI)`
	t.Log("IsPR [undefined] & IsCI [undefined]; propTempCont: ", propTempCont)
	isYes, err := EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `not .IsPR | and .IsCI`
	t.Log("IsPR [undefined] & IsCI [undefined]; propTempCont: ", propTempCont)
	isYes, err = EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `not .IsPR | and .IsCI`
	t.Log("IsPR [undefined] & IsCI=true; propTempCont: ", propTempCont)
	if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateTemplateToBool(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `.IsPR | and .IsCI`
	t.Log("IsPR [undefined] & IsCI=true; propTempCont: ", propTempCont)
	if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateTemplateToBool(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if isYes {
		t.Fatal("Invalid result")
	}

	propTempCont = `.IsPR | and .IsCI`
	t.Log("IsPR=true & IsCI=true; propTempCont: ", propTempCont)
	if err := os.Setenv(configs.PullRequestIDEnvKey, "123"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	if err := os.Setenv(configs.CIModeEnvKey, "true"); err != nil {
		t.Fatal("Failed to set test env! : ", err)
	}
	isYes, err = EvaluateTemplateToBool(propTempCont, true, true, buildRes, envmanModels.EnvsJSONListModel{})
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if !isYes {
		t.Fatal("Invalid result")
	}
}

func TestEvaluateTemplateToString(t *testing.T) {
	buildRes := models.BuildRunResultsModel{}

	propTempCont := ""
	value, err := EvaluateTemplateToString(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.NotEqual(t, nil, err)

	propTempCont = `
{{ if .IsCI }}
value in case of IsCI
{{ end }}
`
	value, err = EvaluateTemplateToString(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, strings.Contains(value, "value in case of IsCI"))

	propTempCont = `
{{ if .IsCI }}
value in case of IsCI
{{ else }}
value in case of not IsCI
{{ end }}
`
	value, err = EvaluateTemplateToString(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, strings.Contains(value, "value in case of IsCI"))

	value, err = EvaluateTemplateToString(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, strings.Contains(value, "value in case of not IsCI"))

	propTempCont = `
This is
{{ if .IsCI }}
value in case of IsCI
{{ end }}
`
	value, err = EvaluateTemplateToString(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, (strings.Contains(value, "This is") && strings.Contains(value, "value in case of IsCI")))

	value, err = EvaluateTemplateToString(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, (strings.Contains(value, "This is")))

	propTempCont = `
This is
{{ if .IsCI }}
value in case of IsCI
{{ end }}
after end
`
	value, err = EvaluateTemplateToString(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, (strings.Contains(value, "This is") && strings.Contains(value, "value in case of IsCI")))

	value, err = EvaluateTemplateToString(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, (strings.Contains(value, "This is") && strings.Contains(value, "after end")))

	propTempCont = `
This is
{{ if .IsCI }}
value in case of IsCI
{{ else }}
value in case of not IsCI
{{ end }}
`
	value, err = EvaluateTemplateToString(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, (strings.Contains(value, "This is") && strings.Contains(value, "value in case of IsCI")))

	value, err = EvaluateTemplateToString(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, (strings.Contains(value, "This is") && strings.Contains(value, "value in case of not IsCI")))

	propTempCont = `
This is
{{ if .IsCI }}
value in case of IsCI
{{ else }}
value in case of not IsCI
{{ end }}
mode
`
	value, err = EvaluateTemplateToString(propTempCont, true, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, (strings.Contains(value, "This is") && strings.Contains(value, "value in case of IsCI") && strings.Contains(value, "mode")))

	value, err = EvaluateTemplateToString(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
	require.Equal(t, nil, err)
	require.Equal(t, true, (strings.Contains(value, "This is") && strings.Contains(value, "value in case of not IsCI") && strings.Contains(value, "mode")))
}
