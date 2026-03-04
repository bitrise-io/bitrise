package progress

import (
	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/log/logwriter"
)

// ShowIndicator displays a spinner animation while the action executes.
// In non-terminal environments (CI), it just executes the action without spinner.
func ShowIndicator(message string, action func()) {
	if !OutputDeviceIsTerminal() {
		action()
		return
	}

	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	output := logwriter.NewLogWriter(logger)
	spinner := NewDefaultSpinnerWithOutput(message, output, logger)
	
	spinner.Run(action)
}
