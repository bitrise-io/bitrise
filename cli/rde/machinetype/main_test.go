package machinetype

import (
	"os"
	"testing"

	"github.com/bitrise-io/bitrise/v2/cli/cmdtest"
)

func TestMain(m *testing.M) { os.Exit(cmdtest.RunIsolated(m)) }
