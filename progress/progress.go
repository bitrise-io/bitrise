package progress

import (
	"github.com/bitrise-io/bitrise/log"
	"github.com/bitrise-io/bitrise/log/logwriter"
	"github.com/bitrise-io/go-utils/progress"
)

func ShowIndicator(message string, action func()) {
	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	output := logwriter.NewLogWriter(logger)
	progress.NewDefaultWrapperWithOutput(message, output).WrapAction(action)
}
