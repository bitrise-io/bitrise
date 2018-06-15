package colorstring

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddColor(t *testing.T) {
	require.Equal(t, string(blackColor+"Hello World!"+reset), addColor(blackColor, "Hello World!"))
}

func TestColors(t *testing.T) {
	require.Equal(t, string(blackColor+"Hello World!"+reset), Black("Hello World!"))
	require.Equal(t, string(redColor+"Hello World!"+reset), Red("Hello World!"))
	require.Equal(t, string(greenColor+"Hello World!"+reset), Green("Hello World!"))
	require.Equal(t, string(yellowColor+"Hello World!"+reset), Yellow("Hello World!"))
	require.Equal(t, string(blueColor+"Hello World!"+reset), Blue("Hello World!"))
	require.Equal(t, string(magentaColor+"Hello World!"+reset), Magenta("Hello World!"))
	require.Equal(t, string(cyanColor+"Hello World!"+reset), Cyan("Hello World!"))
}

func ExampleBlue() {
	fmt.Printf(Blue("Hello World!"))
}
