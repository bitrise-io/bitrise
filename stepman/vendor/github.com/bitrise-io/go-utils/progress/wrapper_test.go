// +build !race

package progress

import (
	"testing"
	"time"
)

func TestNewWrapper(t *testing.T) {
	message := "loading"
	spinner := NewDefaultSpinner(message)

	isInteractiveMode := true
	NewWrapper(spinner, isInteractiveMode).WrapAction(func() {
		time.Sleep(2 * time.Second)
	})
}

func TestNewDefaultWrapper(t *testing.T) {
	message := "loading"
	NewDefaultWrapper(message).WrapAction(func() {
		time.Sleep(2 * time.Second)
	})
}
