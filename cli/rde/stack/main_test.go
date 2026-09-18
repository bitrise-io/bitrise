package stack

import (
	"os"
	"testing"

	"github.com/bitrise-io/bitrise/v3/cli/cmdtest"
)

func TestMain(m *testing.M) { os.Exit(cmdtest.RunIsolated(m)) }
