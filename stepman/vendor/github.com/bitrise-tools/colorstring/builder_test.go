package colorstring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	{
		b := NewBuilder()
		require.NotNil(t, b.add(nil, "Hello World!"))
		require.Equal(t, "Hello World!", b.String())
	}

	{
		b := NewBuilder()
		require.NotNil(t, b.add(Plain, "Hello World!"))
		require.Equal(t, "Hello World!", b.String())
	}

	{
		b := NewBuilder()
		require.NotNil(t, b.add(Black, "Hello World!"))
		require.Equal(t, Black("Hello World!"), b.String())
	}
}

func TestNewLine(t *testing.T) {
	require.Equal(t, "\n", NewBuilder().NewLine().String())
}

func TestAddColors(t *testing.T) {
	require.Equal(t, "Hello World!", NewBuilder().Plain("Hello World!").String())
	require.Equal(t, Black("Hello World!"), NewBuilder().Black("Hello World!").String())
	require.Equal(t, Red("Hello World!"), NewBuilder().Red("Hello World!").String())
	require.Equal(t, Green("Hello World!"), NewBuilder().Green("Hello World!").String())
	require.Equal(t, Yellow("Hello World!"), NewBuilder().Yellow("Hello World!").String())
	require.Equal(t, Blue("Hello World!"), NewBuilder().Blue("Hello World!").String())
	require.Equal(t, Magenta("Hello World!"), NewBuilder().Magenta("Hello World!").String())
	require.Equal(t, Cyan("Hello World!"), NewBuilder().Cyan("Hello World!").String())
}

func ExampleNewBuilder() {
	b := NewBuilder()
	b.Plain("Hello ").Blue("World").Red("!").NewLine()
	b.Plain("This package helps you to build multiline ").Black("c").Red("o").Green("l").Yellow("o").Blue("r").Magenta("e").Cyan("d").Plain(" messages").NewLine()
	fmt.Printf(b.String())
}
