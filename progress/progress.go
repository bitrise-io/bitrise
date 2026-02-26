package progress

import (
	"fmt"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/bitrise/v2/log/logwriter"
	"github.com/bitrise-io/go-utils/v2/progress"
)

// Printer implements the progress.Printer interface using the bitrise CLI logger.
type Printer struct {
	logger log.Logger
	output *logwriter.LogWriter
}

// NewPrinter creates a new Printer for progress indication.
func NewPrinter(logger log.Logger) Printer {
	output := logwriter.NewLogWriter(logger)
	return Printer{
		logger: logger,
		output: output,
	}
}

// PrintWithoutNewline prints text without adding a newline.
func (p Printer) PrintWithoutNewline(text string) {
	if _, err := fmt.Fprint(p.output, text); err != nil {
		p.logger.Warnf("Failed to update progress: %s", err)
	}
}

// Println prints a newline.
func (p Printer) Println() {
	p.logger.Print()
}

// ShowIndicator displays progress dots while the action executes.
func ShowIndicator(message string, action func()) {
	logger := log.NewLogger(log.GetGlobalLoggerOpts())
	printer := NewPrinter(logger)
	
	// Print the message first
	printer.PrintWithoutNewline(message)
	
	if err := progress.NewDefaultSimpleDots(printer).Run(func() error {
		action()
		return nil
	}); err != nil {
		logger.Warnf("Failed to show progress: %s", err)
	}
}
