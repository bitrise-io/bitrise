package progress

import (
	"os"
	"testing"
	"time"
)

func TestNewSpinner(t *testing.T) {
	message := "loading"
	chars := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	delay := 100 * time.Millisecond
	writer := os.Stdout

	spinner := NewSpinner(message, chars, delay, writer)
	spinner.Start()
	time.Sleep(2 * time.Second)
	spinner.Stop()
}

func TestNewDefaultSpinner(t *testing.T) {
	message := "loading"
	spinner := NewDefaultSpinner(message)
	spinner.Start()
	time.Sleep(2 * time.Second)
	spinner.Stop()
}
