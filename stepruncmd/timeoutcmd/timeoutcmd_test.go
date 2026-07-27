package timeoutcmd

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// slowWriter delays every Write call, to widen the gap between the child
// process exiting and the stdout-copying goroutine actually delivering data
// to the wrapped writer, making a missed wait deterministic to observe.
type slowWriter struct {
	delay time.Duration

	mux  sync.Mutex
	done bool
}

func (w *slowWriter) Write(p []byte) (int, error) {
	time.Sleep(w.delay)

	w.mux.Lock()
	defer w.mux.Unlock()
	w.done = true

	return len(p), nil
}

func (w *slowWriter) writeCompleted() bool {
	w.mux.Lock()
	defer w.mux.Unlock()
	return w.done
}

func TestStart_WaitsForOutputCopyToFinish(t *testing.T) {
	writer := &slowWriter{delay: 100 * time.Millisecond}

	cmd := New("", "echo", "hello")
	cmd.SetStandardIO(nil, writer, io.Discard)

	err := cmd.Start()
	require.NoError(t, err)

	require.True(t, writer.writeCompleted(), "Start returned before the stdout copy goroutine finished writing")
}
