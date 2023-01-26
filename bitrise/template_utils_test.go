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
	buildRes := models.BuildRunResultsModel{}

	tests := []struct {
		propTempCont string
		envValue     string
		expected     bool
	}{
		{
			propTempCont: `{{getenv "TEST_KEY" | eq "Test value"}}`,
			envValue:     "Test value",
			expected:     true,
		},
		{
			propTempCont: `{{getenv "TEST_KEY" | eq "A different value"}}`,
			envValue:     "Test value",
			expected:     false,
		},
		{
			propTempCont: `{{enveq "TEST_KEY" "enveq value"}}`,
			envValue:     "enveq value",
			expected:     true,
		},
		{
			propTempCont: `{{enveq "TEST_KEY" "different enveq value"}}`,
			envValue:     "enveq value",
			expected:     false,
		},
	}
	for _, test := range tests {
		t.Run(test.propTempCont, func(t *testing.T) {
			t.Setenv("TEST_KEY", test.envValue)
			isYes, err := EvaluateTemplateToBool(test.propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
			require.NoError(t, err)
			require.Equal(t, test.expected, isYes)
		})
	}
}

func TestCIFlagsAndEnvs(t *testing.T) {
	buildRes := models.BuildRunResultsModel{}

	tests := []struct {
		name         string
		propTempCont string
		setEnvFn     func(t *testing.T)
		isCI         bool
		expected     bool
	}{
		{
			name:         "IsCI=true; propTempCont: {{.IsCI}}",
			propTempCont: "{{.IsCI}}",
			setEnvFn: func(t *testing.T) {
				t.Setenv(configs.CIModeEnvKey, "true")
			},
			isCI:     true,
			expected: true,
		},
		{
			name:         "IsCI=false; propTempCont: {{.IsCI}}",
			propTempCont: "{{.IsCI}}",
			setEnvFn: func(t *testing.T) {
				t.Setenv(configs.CIModeEnvKey, "false")
			},
			isCI:     false,
			expected: false,
		},
		{
			name:         "[unset] IsCI; propTempCont: {{.IsCI}}",
			propTempCont: "{{.IsCI}}",
			setEnvFn: func(t *testing.T) {
				require.NoError(t, os.Unsetenv(configs.CIModeEnvKey))
			},
			isCI:     false,
			expected: false,
		},
		{
			name:         "IsCI=true; short with $; propTempCont: $.IsCI",
			propTempCont: "$.IsCI",
			setEnvFn: func(t *testing.T) {
				t.Setenv(configs.CIModeEnvKey, "true")
			},
			isCI:     true,
			expected: true,
		},
		{
			name:         "IsCI=true; short, no $; propTempCont: .IsCI",
			propTempCont: ".IsCI",
			setEnvFn: func(t *testing.T) {
				t.Setenv(configs.CIModeEnvKey, "true")
			},
			isCI:     true,
			expected: true,
		},
		{
			name:         "IsCI=true; NOT; propTempCont: not .IsCI",
			propTempCont: "not .IsCI",
			setEnvFn: func(t *testing.T) {
				t.Setenv(configs.CIModeEnvKey, "true")
			},
			isCI:     true,
			expected: false,
		},
		{
			name:         "IsCI=false; NOT; propTempCont: not .IsCI",
			propTempCont: "not .IsCI",
			setEnvFn: func(t *testing.T) {
				t.Setenv(configs.CIModeEnvKey, "false")
			},
			isCI:     false,
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setEnvFn(t)
			isYes, err := EvaluateTemplateToBool(test.propTempCont, test.isCI, false, buildRes, envmanModels.EnvsJSONListModel{})
			require.NoError(t, err)
			require.Equal(t, test.expected, isYes)
		})
	}
}

func TestPullRequestFlagsAndEnvs(t *testing.T) {
	buildRes := models.BuildRunResultsModel{}

	t.Run("IsPR [undefined]; propTempCont: {{.IsPR}}", func(t *testing.T) {
		propTempCont := `{{.IsPR}}`
		if err := os.Unsetenv(configs.PullRequestIDEnvKey); err != nil {
			t.Error("Failed to unset environment: ", err)
		}
		isYes, err := EvaluateTemplateToBool(propTempCont, false, false, buildRes, envmanModels.EnvsJSONListModel{})
		require.NoError(t, err)
		require.Equal(t, false, isYes)
	})

	t.Run("IsPR=true; propTempCont: {{.IsPR}}", func(t *testing.T) {
		propTempCont := `{{.IsPR}}`
		t.Setenv(configs.PullRequestIDEnvKey, "123")
		isYes, err := EvaluateTemplateToBool(propTempCont, false, true, buildRes, envmanModels.EnvsJSONListModel{})
		require.NoError(t, err)
		require.Equal(t, true, isYes)
	})
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

	tests := []struct {
		name         string
		propTempCont string
		setEnvFn     func(t *testing.T)
		isCI         bool
		isPR         bool
		expected     bool
	}{
		{
			name:         "IsPR [undefined] & IsCI [undefined]; propTempCont: not .IsPR | and (not .IsCI)",
			propTempCont: "not .IsPR | and (not .IsCI)",
			setEnvFn: func(t *testing.T) {
				require.NoError(t, os.Unsetenv(configs.PullRequestIDEnvKey))
				require.NoError(t, os.Unsetenv(configs.CIModeEnvKey))
			},
			isCI:     false,
			isPR:     false,
			expected: true,
		},
		{
			name:         "IsPR [undefined] & IsCI [undefined]; propTempCont: not .IsPR | and .IsCI",
			propTempCont: "not .IsPR | and .IsCI",
			setEnvFn: func(t *testing.T) {
				require.NoError(t, os.Unsetenv(configs.PullRequestIDEnvKey))
				require.NoError(t, os.Unsetenv(configs.CIModeEnvKey))
			},
			isCI:     false,
			isPR:     false,
			expected: false,
		},
		{
			name:         "IsPR [undefined] & IsCI=true; propTempCont: not .IsPR | and .IsCI",
			propTempCont: "not .IsPR | and .IsCI",
			setEnvFn: func(t *testing.T) {
				require.NoError(t, os.Unsetenv(configs.PullRequestIDEnvKey))
				t.Setenv(configs.CIModeEnvKey, "true")
			},
			isCI:     true,
			isPR:     false,
			expected: true,
		},
		{
			name:         "IsPR [undefined] & IsCI=true; propTempCont: .IsPR | and .IsCI",
			propTempCont: ".IsPR | and .IsCI",
			setEnvFn: func(t *testing.T) {
				require.NoError(t, os.Unsetenv(configs.PullRequestIDEnvKey))
				t.Setenv(configs.CIModeEnvKey, "true")
			},
			isCI:     true,
			isPR:     false,
			expected: false,
		},
		{
			name:         "IsPR=true & IsCI=true; propTempCont: .IsPR | and .IsCI",
			propTempCont: ".IsPR | and .IsCI",
			setEnvFn: func(t *testing.T) {
				t.Setenv(configs.PullRequestIDEnvKey, "123")
				t.Setenv(configs.CIModeEnvKey, "true")
			},
			isCI:     true,
			isPR:     true,
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setEnvFn(t)
			isYes, err := EvaluateTemplateToBool(test.propTempCont, test.isCI, test.isPR, buildRes, envmanModels.EnvsJSONListModel{})
			require.NoError(t, err)
			require.Equal(t, test.expected, isYes)
		})
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
