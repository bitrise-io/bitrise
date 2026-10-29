package timeoutcmd

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStart_WaitsForTheLastOutput(t *testing.T) {
	out := &slowWriter{delay: 200 * time.Millisecond}
	cmd := New("", "bash", "-c", "echo last line")
	cmd.SetStandardIO(nil, out, out)

	require.NoError(t, cmd.Start())
	require.Equal(t, "last line\n", out.String())
}

func TestStart_DoesNotWaitForAProcessLeftBehind(t *testing.T) {
	out := &slowWriter{}
	// The background sleep inherits stdout, so the pipe stays open after bash exits.
	cmd := New("", "bash", "-c", "sleep 10 & echo started")
	cmd.SetStandardIO(nil, out, out)

	start := time.Now()
	require.NoError(t, cmd.Start())
	require.Less(t, time.Since(start), outputDrainTimeout+2*time.Second)
	require.Equal(t, "started\n", out.String())
}

// slowWriter delays each write, standing in for a log pipeline that falls behind.
type slowWriter struct {
	delay time.Duration
	mux   sync.Mutex
	buf   bytes.Buffer
}

func (w *slowWriter) Write(p []byte) (int, error) {
	time.Sleep(w.delay)
	w.mux.Lock()
	defer w.mux.Unlock()
	return w.buf.Write(p)
}

func (w *slowWriter) String() string {
	w.mux.Lock()
	defer w.mux.Unlock()
	return w.buf.String()
}
