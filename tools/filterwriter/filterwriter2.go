package filterwriter

import (
	"bytes"
	"io"
	"strings"
	"sync"
)

type matchKind byte

const (
	exactChar matchKind = iota
	anyChar
	anyCharExceptNewline
)

type matchState struct {
	kind         matchKind
	matchesChar  byte
	matchesSince int
}

type Writer struct {
	writer  io.Writer
	secrets []string

	buffer []byte
	mux    sync.Mutex

	partialMatches [][]matchState
	fullMatches    []matchRange
}

func New(secrets []string, target io.Writer) *Writer {
	var validSecrets []string
	for _, secret := range secrets {
		if secret != "" {
			validSecrets = append(validSecrets, secret)
		}
	}

	// adding transformed secrets with escaped newline characters to ensure that these are also obscured if found in logs
	for _, secret := range secrets {
		if strings.Contains(secret, "\n") {
			validSecrets = append(validSecrets, strings.ReplaceAll(secret, "\n", `\n`))
		}
	}

	maxSecretLen := 0
	for _, secret := range validSecrets {
		if len(secret) > maxSecretLen {
			maxSecretLen = len(secret)
		}
	}

	partialMatches := make([][]matchState, len(validSecrets))
	for i, secret := range validSecrets {
		lines := strings.Split(secret, string(newLine))

		for j, line := range lines {
			for k := 0; k < len(line); k++ {
				partialMatches[i] = append(partialMatches[i], matchState{
					kind:        exactChar,
					matchesChar: line[k],
				})
			}

			if j <= len(lines)-2 {
				partialMatches[i] = append(partialMatches[i], matchState{
					kind:        exactChar,
					matchesChar: '\n',
				})
				partialMatches[i] = append(partialMatches[i], matchState{kind: anyCharExceptNewline})
			}
		}
	}

	return &Writer{
		writer:         target,
		secrets:        validSecrets,
		buffer:         make([]byte, 0, maxSecretLen),
		partialMatches: partialMatches,
	}
}

func isMatch(c byte, state matchState) bool {
	switch state.kind {
	case exactChar:
		return c == state.matchesChar
	case anyChar:
		return true
	case anyCharExceptNewline:
		return c != '\n'
	}

	return false
}

func updateState(prevState *matchState, curState matchState, c byte) int {
	prevMatchesSince := prevState.matchesSince
	if prevState.kind != anyCharExceptNewline {
		prevState.matchesSince = 0
	}

	if curState.kind == anyCharExceptNewline && curState.matchesSince != 0 && isMatch(c, curState) {
		curState.matchesSince++
	} else {
		curState.matchesSince = 0
	}

	if prevMatchesSince != 0 && isMatch(c, curState) {
		prevMatchesSince++
		if prevMatchesSince > curState.matchesSince {
			curState.matchesSince = prevMatchesSince
		}
	}

	return curState.matchesSince
}

func (w *Writer) nextState(c byte) {
	for _, states := range w.partialMatches {
		for j := len(states) - 1; j > 0; j-- {
			states[j].matchesSince = updateState(&states[j-1], states[j], c)

			if states[j-1].kind == anyCharExceptNewline && j >= 2 {
				states[j].matchesSince = updateState(&states[j-2], states[j], c)
			}
		}

		if isMatch(c, states[0]) {
			states[0].matchesSince = 1
		}
	}

	for _, secretPartialMatches := range w.partialMatches {
		matchesSince := secretPartialMatches[len(secretPartialMatches)-1].matchesSince
		if matchesSince != 0 { // full match
			w.fullMatches = append(w.fullMatches, matchRange{
				first: len(w.buffer) - matchesSince,
				last:  len(w.buffer) - 1,
			})

			secretPartialMatches[len(secretPartialMatches)-1].matchesSince = 0
		}
	}
}

// [NONSECRET]
// [NONSECRET (optional)][SECRET (multiple, could)]
// [NONSECRET (optional)][PARTIAL SECRET (multiple)]
func (w *Writer) writeClearTexts(firstPartialMatchStart int) (int, error) {
	w.fullMatches = mergeAllRanges(w.fullMatches)

	lastEnd := 0
	for _, match := range w.fullMatches {
		if match.first >= firstPartialMatchStart {
			break
		}

		n, err := w.writer.Write(w.buffer[lastEnd:match.first])
		if err != nil {
			return n, err
		}
		lastEnd = match.first

		if match.last >= firstPartialMatchStart {
			break
		}

		linesCount := bytes.Count(w.buffer[match.first:match.last], newLine)
		n, err = w.writer.Write([]byte(RedactStr + strings.Repeat("\n"+RedactStr, linesCount)))
		if err != nil {
			return n, err
		}
		lastEnd = match.last + 1

		w.fullMatches = w.fullMatches[1:]
	}

	if len(w.fullMatches) == 0 {
		n, err := w.writer.Write(w.buffer[lastEnd:firstPartialMatchStart])
		if err != nil {
			return n, err
		}
		lastEnd = firstPartialMatchStart
	}

	if lastEnd > len(w.buffer)-1 {
		w.buffer = w.buffer[:0]
	} else {
		w.buffer = w.buffer[lastEnd:]
	}

	for _, fm := range w.fullMatches {
		fm.first -= lastEnd
		fm.last -= lastEnd
	}

	return lastEnd, nil
}

func (w *Writer) Write(data []byte) (int, error) {
	w.mux.Lock()
	defer func() {
		w.mux.Unlock()
	}()

	for _, char := range data {
		w.buffer = append(w.buffer, char)

		w.nextState(char)

		firstPartialMatch := 0
		for _, states := range w.partialMatches {
			for _, state := range states {
				if state.matchesSince > firstPartialMatch {
					firstPartialMatch = state.matchesSince
				}
			}
		}
		firstPartialMatch = len(w.buffer) - firstPartialMatch

		n, err := w.writeClearTexts(firstPartialMatch)
		if err != nil {
			return n, err
		}
	}

	return len(data), nil
}

func (w *Writer) flush() (int, error) {
	w.mux.Lock()
	defer func() {
		w.mux.Unlock()
	}()

	maxPartialStartRel := len(w.buffer)
	return w.writeClearTexts(maxPartialStartRel)
}

func (w *Writer) Close() error {
	_, err := w.flush()
	return err
}
