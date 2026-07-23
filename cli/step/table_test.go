package step

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTable(t *testing.T) {
	got := renderTable(
		[]string{"NAME", "VALUE"},
		[][]string{
			{"branch", "main"},
			{"is_recursive", ""},
		},
	)
	assert.Equal(t, "NAME          VALUE\nbranch        main\nis_recursive  \n", got)
}

func TestRenderTable_NoRows(t *testing.T) {
	got := renderTable([]string{"NAME", "VALUE"}, nil)
	assert.Equal(t, "NAME  VALUE\n", got)
}
