package filterwriter

import (
	"io"
	"sort"
	"strings"
	"sync"
)

type pos struct {
	startRel, endRel int
}

type FastWriter struct {
	writer  io.Writer
	secrets []string

	buffer []byte
	mux    sync.Mutex

	partialMatches [][]bool
	fullMatches    []pos
}

func NewFastWriter(secrets []string, target io.Writer) *FastWriter {
	extendedSecrets := secrets
	// adding transformed secrets with escaped newline characters to ensure that these are also obscured if found in logs
	for _, secret := range secrets {
		if strings.Contains(secret, "\n") {
			extendedSecrets = append(extendedSecrets, strings.ReplaceAll(secret, "\n", `\n`))
		}
	}

	maxSecretLen := 0
	for _, secret := range secrets {
		if len(secret) > maxSecretLen {
			maxSecretLen = len(secret)
		}
	}

	partialMatches := make([][]bool, len(secrets))
	for i, secret := range secrets {
		partialMatches[i] = make([]bool, len(secret))
	}

	return &FastWriter{
		writer:         target,
		secrets:        extendedSecrets,
		buffer:         make([]byte, maxSecretLen),
		partialMatches: partialMatches,
	}
}

func (w *FastWriter) nextState(c byte) {
	for i, secretPartialMatches := range w.partialMatches {
		for matchIndex := len(secretPartialMatches) - 1; matchIndex > 0; matchIndex-- {
			if secretPartialMatches[matchIndex-1] && c == w.secrets[i][matchIndex] {
				secretPartialMatches[matchIndex-1] = false
				secretPartialMatches[matchIndex] = true
			}
		}

		if c == w.secrets[i][0] {
			secretPartialMatches[1] = true
		}
	}

	for _, fullmatch := range w.fullMatches {
		fullmatch.startRel--
		fullmatch.endRel--
	}

	for _, secretPartialMatches := range w.partialMatches {
		if secretPartialMatches[len(secretPartialMatches)-1] { // full match
			w.fullMatches = append(w.fullMatches, pos{
				startRel: -len(secretPartialMatches) + 1,
				endRel:   0,
			})

			secretPartialMatches[len(secretPartialMatches)-1] = false
		}
	}
}

// [NONSECRET]
// [NONSECRET (optional)][SECRET (multiple, could)]
// [NONSECRET (optional)][PARTIAL SECRET (multiple)]
func (w *FastWriter) bufferState() (clearText []pos) {
	sort.Slice(w.fullMatches, func(i, j int) bool {
		p1 := w.fullMatches[i]
		p2 := w.fullMatches[j]

		if p1.startRel == p2.startRel {
			return p1.endRel < p2.endRel
		}
		return p1.startRel < p2.startRel
	})

	var mergedFullMatches []pos
	lastStartRel := 0
	for i := 0; i < len(w.fullMatches)-1; i++ {
		cur := w.fullMatches[i]
		if cur.endRel < w.fullMatches[i+1].startRel-1 {
			if lastStartRel != 0 {
				cur.startRel = lastStartRel
				lastStartRel = 0
			}

			mergedFullMatches = append(mergedFullMatches, cur)
			continue
		}
		if lastStartRel == 0 {
			lastStartRel = cur.startRel
		}
	}

	maxPartialStartRel := 0
	for _, match := range w.partialMatches {
		for i, b := range match {
			if b && i+1 > maxPartialStartRel {
				maxPartialStartRel = i + 1
			}
		}
	}

	lastEndRel := -len(w.buffer) + 1
	for _, match := range mergedFullMatches {
		if match.endRel > maxPartialStartRel {
			break
		}

		clearText = append(clearText, pos{
			startRel: lastEndRel,
			endRel:   match.endRel,
		})
		lastEndRel = match.startRel + 1
	}

	return
}

func (w *FastWriter) Write(data []byte) (int, error) {
	w.mux.Lock()
	defer func() {
		w.mux.Unlock()
	}()

	for _, char := range data {
		w.buffer = append(w.buffer, char)

		w.nextState(char)
		clearTexts := w.bufferState()

		if len(clearTexts) != 0 {
			for _, clearText := range clearTexts {
				if clearText.startRel > len(w.buffer)+1 {
					n, err := w.writer.Write([]byte(RedactStr))
					if err != nil {
						return n, err
					}
				}

				start := clearText.startRel + len(w.buffer)
				end := clearText.endRel + len(w.buffer)
				n, err := w.writer.Write(w.buffer[start:end])
				if err != nil {
					return n, err
				}

				w.buffer = w.buffer[end:]
			}
		}

	}

	return len(data), nil
}

func (w *FastWriter) Flush() (int, error) {
	w.mux.Lock()
	defer func() {
		w.mux.Unlock()
	}()

	return 0, nil
	// return w.writer.Write(w.buffer)
}
