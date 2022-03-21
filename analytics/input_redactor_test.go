package analytics

import (
	"reflect"
	"testing"

	bitriseModels "github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/envman/env"
	"github.com/bitrise-io/envman/models"
)

type fields struct {
	secrets           []string
	environment       []models.EnvironmentItemModel
	buildRunResults   bitriseModels.BuildRunResultsModel
	isCIMode          bool
	isPullRequestMode bool
	envSource         env.EnvironmentSource
}

type testCase struct {
	name    string
	fields  fields
	inputs  []models.EnvironmentItemModel
	want    map[string]interface{}
	wantErr bool
}

func Test_inputRedactor_Redact(t *testing.T) {
	tests := []testCase{
		basicString(),
		handleNil(),
		basicNumber(),
		basicBoolean(),
		retainSimpleMap(),
		retainSimpleSlice(),
		retainComplexHierarchy(),
		basicTemplate(),
		basicExpand(),
		complexTemplateExpand(),
		basicSecretRedaction(),
		complexSecretRedaction(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := inputRedactor{
				secrets:           tt.fields.secrets,
				environment:       tt.fields.environment,
				buildRunResults:   tt.fields.buildRunResults,
				isCIMode:          tt.fields.isCIMode,
				isPullRequestMode: tt.fields.isPullRequestMode,
				envSource:         tt.fields.envSource,
			}
			got, err := i.Redact(tt.inputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Redact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Redact() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func basicString() testCase {
	return testCase{
		name:   "basic string",
		fields: fields{},
		inputs: []models.EnvironmentItemModel{{
			"opts":       basicOpts(),
			"test_input": "test_string",
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": "test_string"}},
		wantErr: false,
	}
}

func handleNil() testCase {
	return testCase{
		name:   "basic string",
		fields: fields{},
		inputs: []models.EnvironmentItemModel{{
			"opts":       basicOpts(),
			"test_input": nil,
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": ""}},
		wantErr: false,
	}
}

func basicNumber() testCase {
	return testCase{
		name:   "basic number",
		fields: fields{},
		inputs: []models.EnvironmentItemModel{{
			"opts":       basicOpts(),
			"test_input": 123456789,
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": 123456789}},
		wantErr: false,
	}
}

func basicBoolean() testCase {
	return testCase{
		name:   "basic boolean",
		fields: fields{},
		inputs: []models.EnvironmentItemModel{{
			"opts":       basicOpts(),
			"test_input": true,
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": true}},
		wantErr: false,
	}
}

func retainSimpleMap() testCase {
	return testCase{
		name:   "retain simple map",
		fields: fields{},
		inputs: []models.EnvironmentItemModel{{
			"opts":       basicOpts(),
			"test_input": map[string]interface{}{"test_key": "test_value"},
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": map[string]interface{}{"test_key": "test_value"}}},
		wantErr: false,
	}
}

func retainSimpleSlice() testCase {
	return testCase{
		name:   "retain simple slice",
		fields: fields{},
		inputs: []models.EnvironmentItemModel{{
			"opts":       basicOpts(),
			"test_input": []string{"test_value"},
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": []interface{}{"test_value"}}},
		wantErr: false,
	}
}

func retainComplexHierarchy() testCase {
	return testCase{
		name:   "retain complex hierarchy",
		fields: fields{},
		inputs: []models.EnvironmentItemModel{{
			"opts":       basicOpts(),
			"test_input": map[string]interface{}{"slice_key": []interface{}{map[interface{}]interface{}{"test_key": "test_value"}}, "map_key": map[interface{}]interface{}{"inner_slice_key": []string{"test_value_2"}, "test_key_2": "test_value_3"}},
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": map[string]interface{}{"slice_key": []interface{}{map[string]interface{}{"test_key": "test_value"}}, "map_key": map[string]interface{}{"inner_slice_key": []interface{}{"test_value_2"}, "test_key_2": "test_value_3"}}}},
		wantErr: false,
	}
}

func basicTemplate() testCase {
	return testCase{
		name:   "basic template",
		fields: fields{isCIMode: true, envSource: mockEnvSource{}},
		inputs: []models.EnvironmentItemModel{{
			"opts":       appendTemplate(basicOpts()),
			"test_input": `{{if .IsCI}}echo "CI mode"{{else}}echo "not CI mode"{{end}}`,
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": `echo "CI mode"`, "original_value": `{{if .IsCI}}echo "CI mode"{{else}}echo "not CI mode"{{end}}`}},
		wantErr: false,
	}
}

func basicExpand() testCase {
	return testCase{
		name:   "basic expand",
		fields: fields{isCIMode: true, envSource: mockEnvSource{envs: map[string]string{"TEST_ENV": "TEST_ENV_VALUE"}}},
		inputs: []models.EnvironmentItemModel{{
			"opts":       appendExpand(basicOpts()),
			"test_input": `test_value $TEST_ENV test_value_2`,
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": `test_value TEST_ENV_VALUE test_value_2`, "original_value": `test_value $TEST_ENV test_value_2`}},
		wantErr: false,
	}
}

func complexTemplateExpand() testCase {
	return testCase{
		name:   "complex template expand",
		fields: fields{isCIMode: true, envSource: mockEnvSource{envs: map[string]string{"TEST_ENV": "TEST_ENV_VALUE"}}},
		inputs: []models.EnvironmentItemModel{{
			"opts":       appendTemplate(appendExpand(basicOpts())),
			"test_input": map[string]interface{}{"slice_key": []interface{}{map[interface{}]interface{}{"test_key": `{{if .IsCI}}echo "CI mode"{{else}}echo "not CI mode"{{end}}`}}, "map_key": map[interface{}]interface{}{"inner_slice_key": []string{"test_value_2"}, "test_key_2": `test_value $TEST_ENV test_value_2`}},
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": map[string]interface{}{"slice_key": []interface{}{map[string]interface{}{"test_key": `echo "CI mode"`}}, "map_key": map[string]interface{}{"inner_slice_key": []interface{}{"test_value_2"}, "test_key_2": "test_value TEST_ENV_VALUE test_value_2"}}, "original_value": map[string]interface{}{"slice_key": []interface{}{map[string]interface{}{"test_key": `{{if .IsCI}}echo "CI mode"{{else}}echo "not CI mode"{{end}}`}}, "map_key": map[string]interface{}{"inner_slice_key": []interface{}{"test_value_2"}, "test_key_2": `test_value $TEST_ENV test_value_2`}}}},
		wantErr: false,
	}
}

func basicSecretRedaction() testCase {
	return testCase{
		name:   "basic secret redaction",
		fields: fields{secrets: []string{"TEST_SECRET"}},
		inputs: []models.EnvironmentItemModel{{
			"opts":       basicOpts(),
			"test_input": "test_stringTEST_SECRETtest_string2",
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": "test_string[REDACTED]test_string2"}},
		wantErr: false,
	}
}

func complexSecretRedaction() testCase {
	return testCase{
		name:   "complex template expand secret redaction",
		fields: fields{isCIMode: true, envSource: mockEnvSource{envs: map[string]string{"TEST_ENV": "TEST_ENV_VALUE"}}, secrets: []string{"test", "mode"}},
		inputs: []models.EnvironmentItemModel{{
			"opts":       appendTemplate(appendExpand(basicOpts())),
			"test_input": map[string]interface{}{"slice_key": []interface{}{map[interface{}]interface{}{"test_key": `{{if .IsCI}}echo "CI mode"{{else}}echo "not CI mode"{{end}}`}}, "map_key": map[interface{}]interface{}{"inner_slice_key": []string{"test_value_2"}, "test_key_2": `test_value $TEST_ENV test_value_2`}},
		}},
		want:    map[string]interface{}{"test_input": map[string]interface{}{"value": map[string]interface{}{"slice_key": []interface{}{map[string]interface{}{"test_key": `echo "CI [REDACTED]"`}}, "map_key": map[string]interface{}{"inner_slice_key": []interface{}{"[REDACTED]_value_2"}, "test_key_2": "[REDACTED]_value TEST_ENV_VALUE [REDACTED]_value_2"}}, "original_value": map[string]interface{}{"slice_key": []interface{}{map[string]interface{}{"test_key": `{{if .IsCI}}echo "CI [REDACTED]"{{else}}echo "not CI [REDACTED]"{{end}}`}}, "map_key": map[string]interface{}{"inner_slice_key": []interface{}{"[REDACTED]_value_2"}, "test_key_2": `[REDACTED]_value $TEST_ENV [REDACTED]_value_2`}}}},
		wantErr: false,
	}
}

func basicOpts() map[string]interface{} {
	return map[string]interface{}{
		"title":       "test_title",
		"description": "test_description",
		"summary":     "test_summary",
	}
}

func appendTemplate(input map[string]interface{}) map[string]interface{} {
	input["is_template"] = true
	return input
}

func appendExpand(input map[string]interface{}) map[string]interface{} {
	input["is_expand"] = true
	return input
}

type mockEnvSource struct {
	envs map[string]string
}

func (e mockEnvSource) GetEnvironment() map[string]string {
	return e.envs
}
