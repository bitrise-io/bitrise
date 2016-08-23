package progress

import (
	"fmt"
	"time"
)

// Action ...
type Action func()

// SimpleProgress ...
// action : have to be a synchronous action!
// tickInterval : e.g. : 5000 * time.Millisecond
func SimpleProgress(printChar string, tickInterval time.Duration, action Action) {
	// run async
	finishedChan := make(chan bool)

	go func() {
		action()
		finishedChan <- true
	}()

	fmt.Print(printChar)
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
