// Package purr implements the `bitrise purr` easter egg: an animated ASCII
// cat mascot with a rainbow-shimmer message.
package purr

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/bitrise-io/bitrise/v2/cli/cmdutil"
	"github.com/bitrise-io/bitrise/v2/internal/style"
)

// purrFrames are the cat-mascot ASCII frames cycled by `bitrise purr`. Each
// frame has the same shape; only the body+tail row differs to suggest a
// happily swinging tail. Keep all frames the same line count and column
// width so the in-place redraw doesn't jiggle.
var purrFrames = []string{
	purrFrame("--==>>"),
	purrFrame("~~==>>"),
	purrFrame("__==>>"),
	purrFrame("~~==>>"),
}

func purrFrame(tail string) string {
	return fmt.Sprintf(`
                  _______
                 /  o  o  \
                ;    \_/   ;
                 \        /
                  '------'
                __|        |__
               |   *  *  *  |%s
               |______________|
                  /\      /\
                 '  '    '  '
`, tail)
}

const purrMessage = "Purr Request is always here to help you!"

// ANSI control sequences used to drive the in-place animation. Kept as
// named constants because the bare escape codes are easy to misread.
const (
	ansiHideCursor = "\x1b[?25l"
	ansiShowCursor = "\x1b[?25h"
	ansiClearBelow = "\x1b[J"   // clear from cursor to end of screen
	ansiCursorPrev = "\x1b[%dF" // CPL: move cursor up N lines, to column 0
)

// hueShiftPerFrame controls how fast the rainbow shimmer scrolls. 15° per
// 250ms tick = one full spectrum every 6s, slow enough to read but fast
// enough to be obviously alive.
const hueShiftPerFrame = 15.0

// NewCmd returns the `bitrise purr` command.
func NewCmd() *cobra.Command {
	var (
		once     bool
		duration time.Duration
		interval time.Duration
	)
	c := &cobra.Command{
		Use:    "purr",
		Short:  "Visit Purr Request, the Bitrise CLI mascot",
		Hidden: true,
		Long: `Visit Purr Request — the rocket-powered cat that's always here to help you.

The mascot animates with a swinging tail. The animation runs for --duration
(default 8s) or until Ctrl-C, advancing one frame every --interval (default
250ms); --once disables animation and prints a single frame. When stdout is
not a terminal (piped output, log file) the command always prints once and
exits regardless of --once.`,
		Example: `  bitrise purr
  bitrise purr --duration 30s
  bitrise purr --interval 100ms
  bitrise purr --once`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmdutil.LogCommandParameters(cmd)
			return runPurr(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), once, duration, interval)
		},
	}
	c.Flags().BoolVar(&once, "once", false, "print a single frame instead of animating")
	c.Flags().DurationVar(&duration, "duration", 8*time.Second, "how long to animate before exiting")
	c.Flags().DurationVar(&interval, "interval", 250*time.Millisecond, "delay between animation frames")
	return c
}

func runPurr(ctx context.Context, in io.Reader, out io.Writer, once bool, duration, interval time.Duration) error {
	s := style.New(out)

	// Static path: piped output, --once, or stdout isn't a TTY. Rainbow
	// auto-degrades to plain text on non-color writers.
	if once || !cmdutil.IsTerminalWriter(out) {
		if _, err := fmt.Fprint(out, purrFrames[0]); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, s.Rainbow(purrMessage, 0))
		return err
	}

	// Hide cursor during animation, restore even on early exit.
	if _, err := fmt.Fprint(out, ansiHideCursor); err != nil {
		return err
	}
	defer func() { _, _ = fmt.Fprint(out, ansiShowCursor) }()

	// Stop cleanly on Ctrl-C.
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// If stdin is a TTY, put it into raw mode and watch for a keypress so
	// any key (including Ctrl-C, which raw mode delivers as a byte rather
	// than a signal) ends the animation. The blocking Read will outlive
	// runPurr when the timer or ctx fires first; that's fine for a short-
	// lived CLI invocation since the OS reaps the goroutine on exit.
	//
	// term.MakeRaw also disables ONLCR on the shared tty device, so plain
	// "\n" stops carriage-returning — without compensation each redrawn
	// line stair-steps to the right. We wrap the output in a tiny "\n →
	// \r\n" translator for the duration of the animation; cursor escape
	// codes don't contain newlines so they pass through untouched.
	keyPress := make(chan struct{})
	if fd, ok := cmdutil.TerminalFd(in); ok {
		if oldState, err := term.MakeRaw(fd); err == nil {
			defer func() { _ = term.Restore(fd, oldState) }()
			out = &crlfWriter{w: out}
			go func() {
				buf := make([]byte, 1)
				_, _ = in.Read(buf)
				close(keyPress)
			}()
		}
	}

	hue := 0.0

	// Initial paint.
	if _, err := fmt.Fprint(out, purrFrames[0]); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, s.Rainbow(purrMessage, hue)); err != nil {
		return err
	}

	// One \n in the frame string + one from Fprintln(message) gives the
	// total number of cursor-down moves we need to undo each tick.
	height := strings.Count(purrFrames[0], "\n") + 1

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	deadline := time.NewTimer(duration)
	defer deadline.Stop()

	frame := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-deadline.C:
			return nil
		case <-keyPress:
			return nil
		case <-ticker.C:
			frame = (frame + 1) % len(purrFrames)
			hue = math.Mod(hue+hueShiftPerFrame, 360)
			// Move cursor back to the top of the frame, clear everything below,
			// and redraw the frame + the (newly hue-shifted) rainbow message.
			if _, err := fmt.Fprintf(out, ansiCursorPrev, height); err != nil {
				return err
			}
			if _, err := fmt.Fprint(out, ansiClearBelow); err != nil {
				return err
			}
			if _, err := fmt.Fprint(out, purrFrames[frame]); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(out, s.Rainbow(purrMessage, hue)); err != nil {
				return err
			}
		}
	}
}

// crlfWriter translates each "\n" byte in its input to "\r\n", forwarding
// everything else unchanged. Used during the animation because raw-mode
// stdin disables ONLCR on the shared tty, so the kernel no longer adds the
// carriage-return for us. Bytes-from-p accounting matches the io.Writer
// contract: a "\n" counts as 1 byte consumed even though 2 bytes were
// emitted to the underlying writer.
type crlfWriter struct{ w io.Writer }

func (c *crlfWriter) Write(p []byte) (int, error) {
	for i, b := range p {
		var chunk []byte
		if b == '\n' {
			chunk = []byte{'\r', '\n'}
		} else {
			chunk = p[i : i+1]
		}
		if _, err := c.w.Write(chunk); err != nil {
			return i, err
		}
	}
	return len(p), nil
}
