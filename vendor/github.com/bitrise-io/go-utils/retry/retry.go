package retry

import (
	"fmt"
	"time"
)

// Action ...
type Action func(attempt uint) error

// AbortableAction ...
type AbortableAction func(attempt uint) (error, bool)

// Model ...
type Model struct {
	retry    uint
	waitTime time.Duration
}

// Times ...
func Times(retry uint) *Model {
	Model := Model{}
	return Model.Times(retry)
}

// Times ...
func (Model *Model) Times(retry uint) *Model {
	Model.retry = retry
	return Model
}

// Wait ...
func Wait(waitTime time.Duration) *Model {
	Model := Model{}
	return Model.Wait(waitTime)
}

// Wait ...
func (Model *Model) Wait(waitTime time.Duration) *Model {
	Model.waitTime = waitTime
	return Model
}

// Try continues executing the supplied action while this action parameter returns an error and the configured
// number of times has not been reached. Otherwise, it stops and returns the last received error.
func (Model Model) Try(action Action) error {
	return Model.TryWithAbort(func(attempt uint) (error, bool) {
		return action(attempt), false
	})
}

// TryWithAbort continues executing the supplied action while this action parameter returns an error, a false bool
// value and the configured number of times has not been reached. Returning a true value from the action aborts the
// retry loop.
//
// Good for retrying actions which can return a mix of retryable and non-retryable failures.
func (Model Model) TryWithAbort(action AbortableAction) error {
	if action == nil {
		return fmt.Errorf("no action specified")
	}

	var err error
	var shouldAbort bool

	for attempt := uint(0); (0 == attempt || nil != err) && attempt <= Model.retry; attempt++ {
		if attempt > 0 && Model.waitTime > 0 {
			time.Sleep(Model.waitTime)
		}

		err, shouldAbort = action(attempt)

		if shouldAbort {
			break
		}
	}

	return err
}
