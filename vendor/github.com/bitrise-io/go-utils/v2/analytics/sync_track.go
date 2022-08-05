package analytics

import (
	"bytes"
	"time"

	"github.com/bitrise-io/go-utils/v2/log"
)

const syncClientTimeout = 10 * time.Second

type syncTracker struct {
	client     Client
	properties []Properties
}

// NewDefaultSyncTracker ...
func NewDefaultSyncTracker(logger log.Logger, properties ...Properties) Tracker {
	return NewSyncTracker(NewDefaultClient(logger, syncClientTimeout), properties...)
}

// NewSyncTracker ...
func NewSyncTracker(client Client, properties ...Properties) Tracker {
	t := syncTracker{client: client, properties: properties}
	return &t
}

// Enqueue ...
func (t syncTracker) Enqueue(eventName string, properties ...Properties) {
	var b bytes.Buffer

	newEvent(eventName, append(t.properties, properties...)).toJSON(&b)
	t.client.Send(&b)
}

// Wait ...
func (t syncTracker) Wait() {
	// no-op in sync tracker
}
