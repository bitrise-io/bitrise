package cli

import (
	"reflect"
	"testing"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/stretchr/testify/require"
)

func TestExpandStepInputs(t *testing.T) {
	tests := []struct {
		name   string
		envs   []envmanModels.EnvironmentItemModel
		inputs []envmanModels.EnvironmentItemModel
		want   map[string]string
	}{
		{
			name: "Env does not depend on input",
			envs: []envmanModels.EnvironmentItemModel{
				{"simulator_device": "$simulator_major", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_major": "12", "opts": map[string]interface{}{"is_sensitive": false}},
				{"simulator_os_version": "$simulator_device", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"simulator_major":      "12",
				"simulator_os_version": "",
			},
		},
		{
			name: "Env does not depend on input (input order switched)",
			envs: []envmanModels.EnvironmentItemModel{
				{"simulator_device": "$simulator_major", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_os_version": "$simulator_device", "opts": map[string]interface{}{"is_sensitive": false}},
				{"simulator_major": "12", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"simulator_major":      "12",
				"simulator_os_version": "",
			},
		},
		{
			name: "Secrets inputs are removed",
			envs: []envmanModels.EnvironmentItemModel{},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_os_version": "13.3", "opts": map[string]interface{}{"is_sensitive": false}},
				{"secret_input": "top secret", "opts": map[string]interface{}{"is_sensitive": true}},
			},
			want: map[string]string{
				"simulator_os_version": "13.3",
				// "secret_input":         "",
			},
		},
		{
			name: "Secrets environments are redacted",
			envs: []envmanModels.EnvironmentItemModel{
				{"secret_env": "top secret", "opts": map[string]interface{}{"is_sensitive": true}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_device": "iPhone $secret_env", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"simulator_device": "iPhone [REDACTED]",
			},
		},
		{
			name: "Not referencing other envs, missing options (sensive input).",
			envs: []envmanModels.EnvironmentItemModel{},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_os_version": "13.3", "opts": map[string]interface{}{}},
				{"simulator_device": "iPhone 8 Plus", "opts": map[string]interface{}{}},
			},
			want: map[string]string{
				"simulator_os_version": "13.3",
				"simulator_device":     "iPhone 8 Plus",
			},
		},
		{
			name: "Not referencing other envs, options specified.",
			envs: []envmanModels.EnvironmentItemModel{},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_os_version": "13.3", "opts": map[string]interface{}{"is_sensitive": false}},
				{"simulator_device": "iPhone 8 Plus", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"simulator_os_version": "13.3",
				"simulator_device":     "iPhone 8 Plus",
			},
		},
		{
			name: "Input references env var, is_expand is false.",
			envs: []envmanModels.EnvironmentItemModel{
				{"SIMULATOR_OS_VERSION": "13.3", "opts": map[string]interface{}{}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_os_version": "$SIMULATOR_OS_VERSION", "opts": map[string]interface{}{"is_expand": false}},
			},
			want: map[string]string{
				"simulator_os_version": "$SIMULATOR_OS_VERSION",
			},
		},
		{
			name: "Env expansion, input contains env var.",
			envs: []envmanModels.EnvironmentItemModel{
				{"SIMULATOR_OS_VERSION": "13.3", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_os_version": "$SIMULATOR_OS_VERSION", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"simulator_os_version": "13.3",
			},
		},
		{
			name: "Env var expansion, input expansion",
			envs: []envmanModels.EnvironmentItemModel{
				{"SIMULATOR_OS_MAJOR_VERSION": "13", "opts": map[string]interface{}{"is_sensitive": false}},
				{"SIMULATOR_OS_MINOR_VERSION": "3", "opts": map[string]interface{}{"is_sensitive": false}},
				{"SIMULATOR_OS_VERSION": "$SIMULATOR_OS_MAJOR_VERSION.$SIMULATOR_OS_MINOR_VERSION", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_os_version": "$SIMULATOR_OS_VERSION", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"simulator_os_version": "13.3",
			},
		},
		{
			name: "Input expansion, input refers other input",
			envs: []envmanModels.EnvironmentItemModel{
				{"simulator_os_version": "12.1", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_os_version": "13.3", "opts": map[string]interface{}{"is_sensitive": false}},
				{"simulator_device": "iPhone 8 ($simulator_os_version)", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"simulator_os_version": "13.3",
				"simulator_device":     "iPhone 8 (13.3)",
			},
		},
		{
			name: "Input expansion, input can not refer other input declared after it",
			envs: []envmanModels.EnvironmentItemModel{
				{"simulator_os_version": "12.1", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"simulator_device": "iPhone 8 ($simulator_os_version)", "opts": map[string]interface{}{"is_sensitive": false}},
				{"simulator_os_version": "13.3", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"simulator_os_version": "13.3",
				"simulator_device":     "iPhone 8 (12.1)",
			},
		},
		{
			name: "Input refers itself, env refers itself",
			envs: []envmanModels.EnvironmentItemModel{
				{"ENV_LOOP": "$ENV_LOOP", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"loop": "$loop", "opts": map[string]interface{}{"is_sensitive": false}},
				{"env_loop": "$ENV_LOOP", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"loop":     "",
				"env_loop": "",
			},
		},
		{
			name: "Input refers itself, env refers itself; both have prefix included",
			envs: []envmanModels.EnvironmentItemModel{
				{"ENV_LOOP": "Env Something: $ENV_LOOP", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"loop": "Something: $loop", "opts": map[string]interface{}{"is_sensitive": false}},
				{"env_loop": "$ENV_LOOP", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"loop":     "Something: ",
				"env_loop": "Env Something: ",
			},
		},
		{
			name: "Inputs refer inputs in a chain, with prefix included",
			envs: []envmanModels.EnvironmentItemModel{},
			inputs: []envmanModels.EnvironmentItemModel{
				{"similar2": "anything", "opts": map[string]interface{}{"is_sensitive": false}},
				{"similar": "$similar2", "opts": map[string]interface{}{"is_sensitive": false}},
				{"env": "Something: $similar", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"similar2": "anything",
				"similar":  "anything",
				"env":      "Something: anything",
			},
		},
		{
			name: "References in a loop are not expanded",
			envs: []envmanModels.EnvironmentItemModel{
				{"B": "$A", "opts": map[string]interface{}{"is_sensitive": false}},
				{"A": "$B", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			inputs: []envmanModels.EnvironmentItemModel{
				{"a": "$b", "opts": map[string]interface{}{"is_sensitive": false}},
				{"b": "$c", "opts": map[string]interface{}{"is_sensitive": false}},
				{"c": "$a", "opts": map[string]interface{}{"is_sensitive": false}},
				{"env": "$A", "opts": map[string]interface{}{"is_sensitive": false}},
			},
			want: map[string]string{
				"a":   "",
				"b":   "",
				"c":   "",
				"env": "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := expandStepInputsForAnalytics(test.inputs, test.envs)

			require.NotNil(t, got)
			if !reflect.DeepEqual(test.want, got) {
				t.Fatalf("expandStepInputs() actual: %v expected: %v", got, test.want)
			}
		})
	}
}
