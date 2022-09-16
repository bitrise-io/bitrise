package cli

import (
	"os"

	log "github.com/bitrise-io/go-utils/v2/advancedlog"
)

func fail(args ...interface{}) {
	log.Error(args...)
	os.Exit(1)
}

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	os.Exit(1)
}

func failln(args ...interface{}) {
	log.Error(args...)
	os.Exit(1)
}
