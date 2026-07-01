package cmdutil

import (
	"os"

	"github.com/bitrise-io/bitrise/v2/log"
)

// Failf ...
func Failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	globalTracker.Wait()
	os.Exit(1)
}
