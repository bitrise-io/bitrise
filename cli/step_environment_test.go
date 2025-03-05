package cli

import (
	"testing"

	"github.com/bitrise-io/envman/v2/models"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/stretchr/testify/require"
)

type EmptyEnvironment struct{}

func (*EmptyEnvironment) GetEnvironment() map[string]string {
	return map[string]string{}
}

func newBool(v bool) *bool {
	b := v
	return &b
}

func Test_prepareStepEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		params  prepareStepInputParams
		want1   []envmanModels.EnvironmentItemModel
		want2   map[string]string
		want3   map[string]interface{}
		wantErr bool
	}{
		{
			name: "Template expansion works",
			params: prepareStepInputParams{
				environment: []envmanModels.EnvironmentItemModel{},
				inputs: []envmanModels.EnvironmentItemModel{
					{"D": "{{.IsCI}}", "opts": models.EnvironmentItemOptionsModel{IsTemplate: newBool(true)}},
				},
				isCIMode: true,
			},
			want1: []envmanModels.EnvironmentItemModel{
				{"D": "true", "opts": models.EnvironmentItemOptionsModel{IsTemplate: newBool(true)}},
			},
			want2: map[string]string{
				"D": "true",
			},
			want3: map[string]interface{}{},
		},
		{
			name: "Default expansion flag is applied",
			params: prepareStepInputParams{
				environment: []envmanModels.EnvironmentItemModel{
					{"A": "B", "opts": map[string]interface{}{}},
				},
				inputs: []envmanModels.EnvironmentItemModel{
					{"myinput": "$A", "opts": models.EnvironmentItemOptionsModel{IsExpand: nil}},
				},
			},
			want1: []envmanModels.EnvironmentItemModel{
				{"A": "B", "opts": models.EnvironmentItemOptionsModel{IsExpand: newBool(true)}},
				{"myinput": "$A", "opts": models.EnvironmentItemOptionsModel{IsExpand: newBool(true)}},
			},
			want2: map[string]string{
				"A":       "B",
				"myinput": "B",
			},
			want3: map[string]interface{}{},
		},
		{
			name: "Non-string inputs propagated",
			params: prepareStepInputParams{
				environment: []envmanModels.EnvironmentItemModel{
					{"A": "B", "opts": map[string]interface{}{}},
				},
				inputs: []envmanModels.EnvironmentItemModel{
					{"bool": true, "opts": models.EnvironmentItemOptionsModel{IsExpand: nil}},
					{"bool2": true, "opts": models.EnvironmentItemOptionsModel{IsExpand: nil, IsTemplate: newBool(true)}},
					{"number": 12, "opts": models.EnvironmentItemOptionsModel{IsExpand: nil}},
				},
			},
			want1: []envmanModels.EnvironmentItemModel{
				{"A": "B", "opts": models.EnvironmentItemOptionsModel{IsExpand: newBool(true)}},
				{"bool": true, "opts": models.EnvironmentItemOptionsModel{IsExpand: newBool(true)}},
				{"bool2": "true", "opts": models.EnvironmentItemOptionsModel{IsExpand: newBool(true), IsTemplate: newBool(true)}},
				{"number": 12, "opts": models.EnvironmentItemOptionsModel{IsExpand: newBool(true)}},
			},
			want2: map[string]string{
				"A":      "B",
				"bool":   "true",
				"bool2":  "true",
				"number": "12",
			},
			want3: map[string]interface{}{
				"bool":   true,
				"number": 12,
			},
		},
	}

	for _, tt := range tests {
		for _, envVar := range tt.want1 {
			if err := envVar.FillMissingDefaults(); err != nil {
				t.Fatalf("prepare: failed to set missing defaults: %s", err)
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got2, got3, err := prepareStepEnvironment(tt.params, &EmptyEnvironment{})
			if tt.wantErr {
				require.Error(t, err, "prepareStepEnvironment() expected to return error")
			} else {
				require.NoError(t, err, "prepareStepEnvironment()")
			}

			require.Equal(t, tt.want1, got1, "prepareStepEnvironment() first return value")
			require.Equal(t, tt.want2, got2, "prepareStepEnvironment() second return value")
			require.Equal(t, tt.want3, got3, "prepareStepEnvironment() third return value")
		})
	}
}
