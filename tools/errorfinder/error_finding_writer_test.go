package errorfinder

import (
	"testing"
)

func Test_errorFindingWriter_findString(t *testing.T) {
	tests := []struct {
		name   string
		inputs []string
		want   *ErrorMessage
	}{
		{
			name: "No color string",
			inputs: []string{
				"Test input",
				"newline\nfoo",
			},
			want: nil,
		},
		{
			name: "Black color string",
			inputs: []string{
				"\x1b[30;1mTest input",
				"newline\nfoo\x1b[0m",
			},
			want: nil,
		},
		{
			name: "Simple red string",
			inputs: []string{
				"\x1b[31;1mTest input\x1b[0m",
			},
			want: &ErrorMessage{
				Message: "Test input",
			},
		},
		{
			name: "Empty red string",
			inputs: []string{
				"Foo\x1b[31;1m\x1b[0mBar",
			},
			want: nil,
		},
		{
			name: "Postfix red string",
			inputs: []string{
				"Foo\x1b[31;1mBar\x1b[0m",
			},
			want: &ErrorMessage{
				Message: "Bar",
			},
		},
		{
			name: "Prefix red string",
			inputs: []string{
				"\x1b[31;1mFoo\x1b[0mBar",
			},
			want: &ErrorMessage{
				Message: "Foo",
			},
		},
		{
			name: "Surrounded red string",
			inputs: []string{
				"Foo\x1b[31;1mBar\x1b[0mBaz",
			},
			want: &ErrorMessage{
				Message: "Bar",
			},
		},
		{
			name: "Multiline red string",
			inputs: []string{
				"Foo\x1b[31;1mBar\nBaz\nQux\x1b[0mTest",
			},
			want: &ErrorMessage{
				Message: "Bar\nBaz\nQux",
			},
		},
		{
			name: "Split red string at content",
			inputs: []string{
				"Foo\x1b[31;1mBa", "r\nBaz\nQux\x1b[0mTest",
			},
			want: &ErrorMessage{
				Message: "Bar\nBaz\nQux",
			},
		},
		{
			name: "Split red string at control",
			inputs: []string{
				"Foo\x1b", "[31", ";1mBar\nBaz\nQux\x1b[0mTest",
			},
			want: &ErrorMessage{
				Message: "Bar\nBaz\nQux",
			},
		},
		{
			name: "Red then black",
			inputs: []string{
				"Foo\x1b[31;1mBar\x1b[30;1mBaz\x1b[0mQux",
			},
			want: &ErrorMessage{
				Message: "Bar",
			},
		},
		{
			name: "Multiple red sections",
			inputs: []string{
				"Foo\x1b[31;1mBar\x1b[0mBaz\x1b[31;1mQux\x1b[0m",
			},
			want: &ErrorMessage{
				Message: "Qux",
			},
		},
		{
			name: "Complex multiple red sections",
			inputs: []string{
				"Foo\x1b[", "31;1mB\na\nr\x1b", "[0mBaz\x1b[31;1mQ", "\nu\nx\x1b[0mTest",
			},
			want: &ErrorMessage{
				Message: "Q\nu\nx",
			},
		},
		{
			name: "Endless red",
			inputs: []string{
				"\x1b[31;1mTest\n in", "put",
			},
			want: &ErrorMessage{
				Message: "Test\n input",
			},
		},
		{
			name: "Repeated reds",
			inputs: []string{
				"\x1b[31;1mTest \x1b[31;1min", "put\x1b[0m",
			},
			want: &ErrorMessage{
				Message: "Test input",
			},
		},
		{
			name: "Endless repeated reds",
			inputs: []string{
				"Foo\n\n\n\x1b[31;1mTest \x1b[31;1mi\x1b[31;1mn", "put",
			},
			want: &ErrorMessage{
				Message: "Test input",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newWriter(nil)
			for _, input := range tt.inputs {
				e.findString(input)
			}
			got := e.getErrorMessage()
			if (tt.want == nil && got != nil) || (tt.want != nil && got == nil) || (tt.want != nil && tt.want.Message != got.Message) {
				t.Errorf("got %v. want %v", got, tt.want)
			}
		})
	}
}
