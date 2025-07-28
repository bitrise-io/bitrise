package asdf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAsdfListOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    []string
	}{
		{
			name:    "basic",
			output:  "  1.21.0\n  1.21.11\n  1.21\n  1.22.0\n *1.22\n  1.23.5\n  1.23.7\n  1.23\n  1.24.0\n  1",
			want:    []string{"1.21.0", "1.21.11", "1.21", "1.22.0", "1.22", "1.23.5", "1.23.7", "1.23", "1.24.0", "1"},
		},
		{
			name:    "empty",
			output:  "",
			want:    []string{},
		},
		{
			name:    "single version",
			output:  "1.0.0",
			want:    []string{"1.0.0"},
		},
		{
			name:    "single active version",
			output:  "*1.0.0",
			want:    []string{"1.0.0"},
		},
		{
			name:    "whitespace only",
			output:  "   ",
			want:    []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAsdfListOutput(tt.output)
			if !assert.Equal(t, tt.want, got) {
				t.Errorf("parseAsdfListOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
