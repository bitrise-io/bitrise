package cli

import (
	"reflect"
	"testing"
)

func Test_insertLatestNVersions(t *testing.T) {
	tests := []struct {
		name       string
		latests    []uint64
		newVersion uint64
		cap        int
		want       []uint64
	}{
		{
			name:       "insertLatestNVersions",
			latests:    []uint64{3, 2, 1},
			newVersion: 4,
			want:       []uint64{4, 3, 2},
		},
		{
			latests:    []uint64{5, 2},
			newVersion: 4,
			want:       []uint64{5, 4},
		},
		{
			latests:    []uint64{5, 2},
			newVersion: 4,
			cap:        3,
			want:       []uint64{5, 4, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if cap(tt.latests) < tt.cap {
				tt.latests = append(make([]uint64, 0, tt.cap), tt.latests...)
			}

			if got := insertLatestNVersions(tt.latests, tt.newVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("insertLatestNVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
