package cmdtest

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(RunIsolated(m))
}
