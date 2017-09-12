package progress

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// Wrapper ...
type Wrapper struct {
	spinner         Spinner
	action          func()
	interactiveMode bool
}

// NewWrapper ...
func NewWrapper(spinner Spinner, interactiveMode bool) Wrapper {
	return Wrapper{
		spinner:         spinner,
		interactiveMode: interactiveMode,
	}
}

// NewDefaultWrapper ...
func NewDefaultWrapper(message string) Wrapper {
	spinner := NewDefaultSpinner(message)
	interactiveMode := OutputDeviceIsTerminal()
	return NewWrapper(spinner, interactiveMode)
}

// WrapAction ...
func (w Wrapper) WrapAction(action func()) {
	if w.interactiveMode {
		w.spinner.Start()
		action()
		w.spinner.Stop()
	} else {
		message := w.spinner.message
		if !strings.HasSuffix(message, ".") {
			message = message + "..."
		}
		fmt.Fprintln(w.spinner.writer, message)
		action()
	}
}

// OutputDeviceIsTerminal ...
func OutputDeviceIsTerminal() bool {
	return terminal.IsTerminal(int(os.Stdout.Fd()))
}
