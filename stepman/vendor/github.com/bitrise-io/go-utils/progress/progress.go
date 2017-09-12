package progress

import (
	"fmt"
	"time"
)

// SimpleProgressE ...
func SimpleProgressE(printChar string, tickInterval time.Duration, action func() error) error {
	var actionError error
	SimpleProgress(printChar, tickInterval, func() {
		actionError = action()
	})
	return actionError
}

// SimpleProgress ...
// action : have to be a synchronous action!
// tickInterval : e.g. : 5000 * time.Millisecond
func SimpleProgress(printChar string, tickInterval time.Duration, action func()) {
	// run async
	finishedChan := make(chan bool)

	go func() {
		action()
		finishedChan <- true
	}()

	isRunFinished := false
	for !isRunFinished {
		select {
		case <-finishedChan:
			isRunFinished = true
		case <-time.Tick(tickInterval):
			fmt.Print(printChar)
		}
	}
	fmt.Println()
}
