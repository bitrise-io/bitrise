package timeoutcmd

import (
	"io"
	"os"
	"os/exec"
	"sync"
)

// commandOutput copies a command's stdout and stderr through pipes it owns,
// instead of the ones exec.Cmd creates, so that waiting for the copy doesn't
// mean cmd.Wait, which waits for EOF or, given a WaitDelay, closes the pipes.
type commandOutput struct {
	writeEnds []*os.File
	done      chan struct{}
}

// copyOutput replaces cmd's stdout and stderr writers with pipes, and copies
// from them to the original writers. Like exec.Cmd, it uses one pipe for both
// when they are the same writer, which keeps their lines in order.
func copyOutput(cmd *exec.Cmd) (commandOutput, error) {
	dests := []io.Writer{cmd.Stdout}
	if !sameWriter(cmd.Stdout, cmd.Stderr) {
		dests = append(dests, cmd.Stderr)
	}

	output := commandOutput{done: make(chan struct{})}
	var copies sync.WaitGroup
	for _, dest := range dests {
		r, w, err := os.Pipe()
		if err != nil {
			output.closeWriteEnds()
			return commandOutput{}, err
		}
		output.writeEnds = append(output.writeEnds, w)

		copies.Go(func() {
			defer r.Close()
			_, _ = io.Copy(dest, r)
		})
	}
	go func() {
		copies.Wait()
		close(output.done)
	}()

	cmd.Stdout = output.writeEnds[0]
	cmd.Stderr = output.writeEnds[len(output.writeEnds)-1]
	return output, nil
}

func (o commandOutput) closeWriteEnds() {
	for _, w := range o.writeEnds {
		_ = w.Close()
	}
}

// sameWriter reports whether a and b are the same writer, the way exec.Cmd
// decides to share one pipe: a writer of an uncomparable type is never equal.
func sameWriter(a, b io.Writer) (same bool) {
	defer func() {
		if recover() != nil {
			same = false
		}
	}()
	return a == b
}
