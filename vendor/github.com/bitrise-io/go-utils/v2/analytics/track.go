package analytics

import (
	"bytes"
	"sync"
	"time"

	"github.com/bitrise-io/go-utils/v2/env"
	"github.com/bitrise-io/go-utils/v2/log"
)

const poolSize = 10
const bufferSize = 100
const timeout = 30 * time.Second
const asyncClientTimeout = 30 * time.Second
const analyticsDisabledEnv = "ANALYTICS_DISABLED"

// Tracker ...
type Tracker interface {
	Enqueue(eventName string, properties ...Properties)
	Wait()
	IsTracking() bool
}

type tracker struct {
	jobs        chan *bytes.Buffer
	waitGroup   *sync.WaitGroup
	client      Client
	properties  []Properties
	waitTimeout time.Duration
}

type noopTracker struct{}

// Enqueue ...
func (t noopTracker) Enqueue(eventName string, properties ...Properties) {}

// Wait ...
func (t noopTracker) Wait() {}

// IsTracking ...
func (t noopTracker) IsTracking() bool {
	return false
}

// NewDefaultTracker ...
func NewDefaultTracker(logger log.Logger, envRepo env.Repository, properties ...Properties) Tracker {
	if envRepo.Get(analyticsDisabledEnv) == "true" {
		return noopTracker{}
	}
	return NewTracker(NewDefaultClient(logger, asyncClientTimeout), timeout, properties...)
}

// NewTracker ...
func NewTracker(client Client, waitTimeout time.Duration, properties ...Properties) Tracker {
	t := tracker{client: client, jobs: make(chan *bytes.Buffer, bufferSize), waitGroup: &sync.WaitGroup{}, properties: properties, waitTimeout: waitTimeout}
	t.init(poolSize)
	return &t
}

// Enqueue ...
func (t tracker) Enqueue(eventName string, properties ...Properties) {
	var b bytes.Buffer
	newEvent(eventName, append(t.properties, properties...)).toJSON(&b)
	t.waitGroup.Add(1)
	t.jobs <- &b
}

// Wait ...
func (t tracker) Wait() {
	close(t.jobs)
	c := make(chan struct{})
	go func() {
		defer close(c)
		t.waitGroup.Wait()
	}()
	select {
	case <-c:
	case <-time.After(t.waitTimeout):
	}
}

// IsTracking ...
func (t tracker) IsTracking() bool {
	return true
}

func (t tracker) init(size int) {
	for i := 0; i < size; i++ {
		go t.worker()
	}
}

func (t tracker) worker() {
	for job := range t.jobs {
		t.client.Send(job)
		t.waitGroup.Done()
	}
}
