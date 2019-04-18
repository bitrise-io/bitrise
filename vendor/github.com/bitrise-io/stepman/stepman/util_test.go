package stepman

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

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

func Test_parseStepModel(t *testing.T) {
	empty := ""
	falseBool := false
	zero := 0
	tests := []struct {
		name     string
		bytes    []byte
		validate bool
		want     models.StepModel
		wantErr  bool
	}{
		{
			name:     "Meta field",
			bytes:    []byte(stepDefinitionMetaFieldOnly),
			validate: false,
			want: models.StepModel{
				Title: &empty,
				Meta: map[string]interface{}{
					"bitrise.io.addons.optional.2": []interface{}{
						map[string]interface{}{
							"addon_id": "addons-testing",
						},
					},
					"bitrise.io.addons.required": []interface{}{
						map[string]interface{}{
							"addon_id": "addons-testing",
							"addon_options": map[string]interface{}{
								"required": true,
								"title":    "Testing Addon",
							},
							"addon_params": "--token TOKEN",
						},
						map[string]interface{}{
							"addon_id": "addons-ship",
							"addon_options": map[string]interface{}{
								"required": true,
								"title":    "Ship Addon",
							},
							"addon_params": "--token TOKEN",
						},
					},
				},
				Summary:             &empty,
				Description:         &empty,
				Website:             &empty,
				SourceCodeURL:       &empty,
				SupportURL:          &empty,
				IsRequiresAdminUser: &falseBool,
				IsAlwaysRun:         &falseBool,
				IsSkippable:         &falseBool,
				RunIf:               &empty,
				Timeout:             &zero,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStepModel(tt.bytes, tt.validate)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStepModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseStepModel() = %+v, want %v,\n Diff: %s", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

const stepDefinitionMetaFieldOnly = `
meta:
  bitrise.io.addons.required: 
    - addon_id: "addons-testing"
      addon_params: "--token TOKEN"
      addon_options: 
        required: true
        title: "Testing Addon"
    - addon_id: "addons-ship"
      addon_params: "--token TOKEN"
      addon_options: 
        required: true
        title: "Ship Addon"
  bitrise.io.addons.optional.2: [{"addon_id":"addons-testing"}]
`
