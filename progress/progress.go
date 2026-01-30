package progress

import (
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/log/logwriter"
)

// ShowIndicator displays a spinner animation while the action executes.
func ShowIndicator(message string, action func()) {
	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	output := logwriter.NewLogWriter(logger)
	spinner := NewDefaultSpinnerWithOutput(message, output)
	
	spinner.Start()
	action()
	spinner.Stop()
}
