package analytics

import (
	"bytes"
	"sync"
)

// Worker ...
type Worker interface {
	Wait()
	Run(message *bytes.Buffer)
}

type worker struct {
	jobs      chan *bytes.Buffer
	waitGroup *sync.WaitGroup
	client    Client
}

// Wait ...
func (w worker) Wait() {
	close(w.jobs)
	w.waitGroup.Wait()
}

// NewWorker ...
func NewWorker(client Client) Worker {
	w := worker{client: client, jobs: make(chan *bytes.Buffer, bufferSize), waitGroup: &sync.WaitGroup{}}
	w.init(poolSize)
	return w
}

// Run ...
func (w worker) Run(message *bytes.Buffer) {
	w.jobs <- message
	w.waitGroup.Add(1)
}

func (w worker) init(size int) {
	for i := 0; i < size; i++ {
		go w.worker()
	}
}

func (w worker) worker() {
	for job := range w.jobs {
		w.client.Send(job)
		w.waitGroup.Done()
	}
}
