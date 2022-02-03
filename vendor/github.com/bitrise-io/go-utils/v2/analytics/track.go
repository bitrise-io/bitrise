package analytics

import (
	"bytes"

	"github.com/bitrise-io/go-utils/v2/log"
)

const poolSize = 10
const bufferSize = 100

// Properties ...
type Properties map[string]interface{}

// Merge ...
func (p Properties) Merge(properties Properties) Properties {
	r := Properties{}
	for key, value := range p {
		r[key] = value
	}
	for key, value := range properties {
		r[key] = value
	}
	return r
}

// Tracker ...
type Tracker interface {
	Enqueue(eventName string, properties ...Properties)
	Pin(properties ...Properties) Tracker
}

type tracker struct {
	worker     Worker
	properties []Properties
}

// NewDefaultTracker ...
func NewDefaultTracker(logger log.Logger) Tracker {
	return NewTracker(NewWorker(NewDefaultClient(logger)))
}

// NewTracker ...
func NewTracker(worker Worker, properties ...Properties) Tracker {
	t := tracker{worker: worker, properties: properties}
	return &t
}

// Enqueue ...
func (t tracker) Enqueue(eventName string, properties ...Properties) {
	var b bytes.Buffer
	newEvent(eventName, append(t.properties, properties...)).toJSON(&b)
	t.worker.Run(&b)
}

// Fork ...
func (t tracker) Pin(properties ...Properties) Tracker {
	return NewTracker(t.worker, append(t.properties, properties...)...)
}
