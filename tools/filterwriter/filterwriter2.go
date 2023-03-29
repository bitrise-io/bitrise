package filterwriter

import (
	"bytes"
	"io"
	"log"
	"strings"
	"sync"
)

type Writer struct {
	writer  io.Writer
	secrets []string

	buffer []byte
	mux    sync.Mutex

	partialMatches [][]bool
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

	partialMatches := make([][]bool, len(validSecrets))
	for i, secret := range validSecrets {
		partialMatches[i] = make([]bool, len(secret))
	}

	return &Writer{
		writer:         target,
		secrets:        validSecrets,
		buffer:         make([]byte, 0, maxSecretLen),
		partialMatches: partialMatches,
	}
}

func (w *Writer) nextState(c byte) {
	for i, secretPartialMatches := range w.partialMatches {
		for matchIndex := len(secretPartialMatches) - 1; matchIndex > 0; matchIndex-- {
			if secretPartialMatches[matchIndex-1] && c == w.secrets[i][matchIndex] {
				secretPartialMatches[matchIndex] = true
			}

			secretPartialMatches[matchIndex-1] = false
		}

		// Secrets have at leat lenght 1
		if c == w.secrets[i][0] {
			secretPartialMatches[0] = true
		}
	}

	for _, secretPartialMatches := range w.partialMatches {
		if secretPartialMatches[len(secretPartialMatches)-1] { // full match
			w.fullMatches = append(w.fullMatches, matchRange{
				first: len(w.buffer) - len(secretPartialMatches),
				last:  len(w.buffer) - 1,
			})

			secretPartialMatches[len(secretPartialMatches)-1] = false
		}
	}
}

// [NONSECRET]
// [NONSECRET (optional)][SECRET (multiple, could)]
// [NONSECRET (optional)][PARTIAL SECRET (multiple)]
func (w *Writer) writeClearTexts(maxPartialStartRel int) (int, error) {
	w.fullMatches = mergeAllRanges(w.fullMatches)

	lastEnd := 0
	for _, match := range w.fullMatches {
		if match.first >= maxPartialStartRel {
			break
		}

		if match.first < 0 {
			log.Printf("sf")
		}

		n, err := w.writer.Write(w.buffer[lastEnd:match.first])
		if err != nil {
			return n, err
		}
		lastEnd = match.first

		if match.last >= maxPartialStartRel {
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
		n, err := w.writer.Write(w.buffer[lastEnd:maxPartialStartRel])
		if err != nil {
			return n, err
		}
		lastEnd = maxPartialStartRel
	}

	if lastEnd+1 >= len(w.buffer) {
		w.buffer = w.buffer[:0]
	} else {
		w.buffer = w.buffer[lastEnd+1:]
	}

	for _, fm := range w.fullMatches {
		fm.first -= lastEnd
		if fm.first < 0 {
			log.Printf("sf")
		}
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
	}

	maxPartialStartRel := len(w.buffer)
	for _, match := range w.partialMatches {
		for i, b := range match {
			if b && i < maxPartialStartRel {
				maxPartialStartRel = i
			}
		}
	}

	n, err := w.writeClearTexts(maxPartialStartRel)
	if err != nil {
		return n, err
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
