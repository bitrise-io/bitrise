package progress

import (
	"errors"
	"sync"
	"time"
)

// SimpleDots provides periodic output for long-running operations in CI environments.
// It prints dots at regular intervals to show that work is progressing.
type SimpleDots struct {
	printer Printer
	ticker  Ticker

	stopChan chan struct{}
}

// NewDefaultSimpleDots creates a SimpleDots with a default 5-second interval.
func NewDefaultSimpleDots(printer Printer) *SimpleDots {
	return NewSimpleDotsWithInterval(5*time.Second, printer)
}

// NewSimpleDotsWithInterval creates a new SimpleDots with the given interval.
func NewSimpleDotsWithInterval(interval time.Duration, printer Printer) *SimpleDots {
	return NewSimpleDotsWithTicker(printer, NewTicker(interval))
}

// NewSimpleDotsWithTicker creates a new SimpleDots with a custom Ticker for testing.
func NewSimpleDotsWithTicker(printer Printer, ticker Ticker) *SimpleDots {
	return &SimpleDots{
		printer: printer,
		ticker:  ticker,
	}
}

// Run starts the progress dots and executes the given action.
func (t *SimpleDots) Run(action func() error) error {
	if t.stopChan != nil {
		return errors.New("progress can only be run once")
	}

	tickerGroup := sync.WaitGroup{}
	t.stopChan = make(chan struct{})
	defer func() {
		t.ticker.Stop()
		close(t.stopChan)   // Signal the ticker goroutine to stop
		tickerGroup.Wait()  // Wait for the ticker goroutine to finish, prevent logger race
		t.printer.Println() // Print a newline after the dots
	}()

	tickerGroup.Add(1)
	go func() {
		defer tickerGroup.Done()
		for {
			select {
			case <-t.stopChan:
				return
			case <-t.ticker.Chan():
				t.printer.PrintWithoutNewline(".")
			}
		}
	}()

	return action()
}
