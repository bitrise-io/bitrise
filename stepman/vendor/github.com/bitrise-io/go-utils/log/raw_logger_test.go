package log

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRawPrint(t *testing.T) {
	t.Log("Custom Formattable")
	{
		var b bytes.Buffer
		logger := NewRawLogger(&b)

		test := TestFormattable{
			A: "log",
			B: "test",
		}

		logger.Print(test)
		require.Equal(t, "log test\n", b.String())
	}
}
