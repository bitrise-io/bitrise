package models

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_castRecursiveToMapStringInterface(t *testing.T) {
	tests := []struct {
		name    string
		source  interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:    "1 level",
			source:  map[interface{}]interface{}{"aa": "bb"},
			want:    map[string]interface{}{"aa": "bb"},
			wantErr: false,
		},
		{
			name:    "1 level map[string]interface at top",
			source:  map[string]interface{}{"aa": "bb"},
			want:    map[string]interface{}{"aa": "bb"},
			wantErr: false,
		},
		{
			name:    "2 levels",
			source:  map[interface{}]interface{}{"aa": map[interface{}]interface{}{"aa": "bb"}, "b": "c"},
			want:    map[string]interface{}{"aa": map[string]interface{}{"aa": "bb"}, "b": "c"},
			wantErr: false,
		},
		{
			name:    "2 levels map[string]interface at top",
			source:  map[string]interface{}{"aa": map[interface{}]interface{}{"aa": "bb"}, "b": "c"},
			want:    map[string]interface{}{"aa": map[string]interface{}{"aa": "bb"}, "b": "c"},
			wantErr: false,
		},
		{
			name: "Decoded from yml",
			source: map[interface{}]interface{}{
				"bitrise.io.addons.optional.2": []interface{}{
					map[interface{}]interface{}{
						"addon_id": "addons-testing",
					},
				},
				"bitrise.io.addons.required": []interface{}{
					map[interface{}]interface{}{
						"addon_id": "addons-testing",
						"addon_options": map[interface{}]interface{}{
							"required": true,
							"title":    "Testing Addon",
						},
						"addon_params": "--token TOKEN",
					},
					map[interface{}]interface{}{
						"addon_id": "addons-ship",
						"addon_options": map[interface{}]interface{}{
							"required": true,
							"title":    "Ship Addon",
						},
						"addon_params": "--token TOKEN",
					},
				},
			},
			want: map[string]interface{}{
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := recursiveJSONMarshallable(tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("castRecursiveToMapStringInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("castRecursiveToMapStringInterface() = %v, want %v, \n Diff %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}
