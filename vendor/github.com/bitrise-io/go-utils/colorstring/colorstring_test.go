package colorstring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddColor(t *testing.T) {
	/*
	  blackColor   Color = "\x1b[30;1m"
	  resetColor   Color = "\x1b[0m"
	*/

	t.Log("colored_string = color + string + reset_color")
	{
		desiredColored := "\x1b[30;1m" + "test" + "\x1b[0m"
		colored := addColor(blackColor, "test")
		require.Equal(t, desiredColored, colored)
	}
}

func TestBlack(t *testing.T) {
	t.Log("Simple string can be blacked")
	{
		desiredColored := "\x1b[30;1m" + "test" + "\x1b[0m"
		colored := Black("test")
		require.Equal(t, desiredColored, colored)
	}

	t.Log("Multiple strings can be blacked")
	{
		desiredColored := "\x1b[30;1m" + "Hello Bitrise !" + "\x1b[0m"
		colored := Black("Hello ", "Bitrise ", "!")
		require.Equal(t, desiredColored, colored)
	}
}

func TestBlackf(t *testing.T) {
	t.Log("Simple format can be blacked")
	{
		desiredColored := "\x1b[30;1m" + fmt.Sprintf("Hello %s", "bitrise") + "\x1b[0m"
		colored := Blackf("Hello %s", "bitrise")
		require.Equal(t, desiredColored, colored)
	}

	t.Log("Complex format can be blacked")
	{
		desiredColored := "\x1b[30;1m" + fmt.Sprintf("Hello %s %s", "bitrise", "!") + "\x1b[0m"
		colored := Blackf("Hello %s %s", "bitrise", "!")
		require.Equal(t, desiredColored, colored)
	}
}
