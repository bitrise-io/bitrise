package log

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestFormattable struct {
	A string `json:"a,omitempty"`
	B string `json:"b,omitempty"`
}

// String ...
func (f TestFormattable) String() string {
	return fmt.Sprintf("%s %s", f.A, f.B)
}

// JSON ...
func (f TestFormattable) JSON() string {
	return fmt.Sprintf(`{"a":"%s","b":"%s"}`, f.A, f.B)
}

func TestJSONPrint(t *testing.T) {

	t.Log("Custom Formattable")
	{
		var b bytes.Buffer
		logger := NewJSONLoger(&b)

		test := TestFormattable{
			A: "log",
			B: "test",
		}

		logger.Print(test)
		require.Equal(t, `{"a":"log","b":"test"}`, b.String())
	}
}
