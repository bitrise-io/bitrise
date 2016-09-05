package retry

import (
	"fmt"
	"time"
)

// Action ...
type Action func(attempt uint) error

// Model ...
type Model struct {
	retry   uint
	waitSec uint
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
func Wait(wait uint) *Model {
	Model := Model{}
	return Model.Wait(wait)
}

// Wait ...
func (Model *Model) Wait(waitSec uint) *Model {
	Model.waitSec = waitSec
	return Model
}

// Try ...
func (Model Model) Try(action Action) error {
	if action == nil {
		return fmt.Errorf("no action specified")
	}

	var err error

	for attempt := uint(0); (0 == attempt || nil != err) && attempt <= Model.retry; attempt++ {
		if attempt > 0 && Model.waitSec > 0 {
			time.Sleep(time.Duration(Model.waitSec) * time.Second)
		}

		err = action(attempt)
	}

	return err
}
