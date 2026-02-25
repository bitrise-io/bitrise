package progress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/bitrise-io/bitrise/v2/log"
)

// Spinner displays an animated progress indicator in the terminal.
type Spinner struct {
	message string
	chars   []string
	delay   time.Duration
	writer  io.Writer
	logger  log.Logger

	stopChan   chan struct{}
	lastOutput string
}

// NewSpinner creates a new Spinner with custom animation characters and timing.
func NewSpinner(message string, chars []string, delay time.Duration, writer io.Writer, logger log.Logger) Spinner {
	return Spinner{
		message: message,
		chars:   chars,
		delay:   delay,
		writer:  writer,
		logger:  logger,
	}
}

// NewDefaultSpinner creates a Spinner with default animation characters and timing, writing to stdout.
func NewDefaultSpinner(message string, logger log.Logger) Spinner {
	return NewDefaultSpinnerWithOutput(message, os.Stdout, logger)
}

// NewDefaultSpinnerWithOutput creates a Spinner with default animation characters and timing.
func NewDefaultSpinnerWithOutput(message string, output io.Writer, logger log.Logger) Spinner {
	chars := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	delay := 100 * time.Millisecond
	return NewSpinner(message, chars, delay, output, logger)
}

func (s *Spinner) erase() {
	n := utf8.RuneCountInString(s.lastOutput)
	for _, c := range []string{"\b", " ", "\b"} {
		for i := 0; i < n; i++ {
			if _, err := fmt.Fprint(s.writer, c); err != nil {
				s.logger.Warnf("Failed to update progress: %s", err)
			}
		}
	}
	s.lastOutput = ""
}

// Run starts the spinner animation and executes the given action.
// It waits for the action to complete before stopping the spinner.
func (s *Spinner) Run(action func()) {
	if s.stopChan != nil {
		s.logger.Warnf("Spinner can only be run once")
		return
	}

	spinnerGroup := sync.WaitGroup{}
	s.stopChan = make(chan struct{})
	defer func() {
		close(s.stopChan)    // Signal the spinner goroutine to stop
		spinnerGroup.Wait()  // Wait for the spinner goroutine to finish
		s.erase()            // Clear the spinner output
	}()

	spinnerGroup.Add(1)
	go func() {
		defer spinnerGroup.Done()
		for {
			for i := 0; i < len(s.chars); i++ {
				select {
				case <-s.stopChan:
					return
				default:
					s.erase()

					out := fmt.Sprintf("%s %s", s.message, s.chars[i])
					if _, err := fmt.Fprint(s.writer, out); err != nil {
						s.logger.Warnf("Failed to update progress: %s", err)
					}
					s.lastOutput = out

					time.Sleep(s.delay)
				}
			}
		}
	}()

	action()
}


