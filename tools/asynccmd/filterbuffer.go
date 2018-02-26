package asynccmd

import (
	"bytes"
	"io"
	"strings"
	"sync"
)

// RedactStr ...
var RedactStr = "[REDACTED]"

var newLine = []byte("\n")

// Buffer ...
type Buffer struct {
	Buff bytes.Buffer
	sync.Mutex

	secrets [][][]byte

	chunk []byte
	store [][]byte
}

func newBuffer(secrets []string) *Buffer {
	return &Buffer{
		Buff:    bytes.Buffer{},
		secrets: secretsByteList(secrets),
		Mutex:   sync.Mutex{},
	}
}

// Write implements io.Writer interface
// Splits p into lines, the lines are matched against the secrets,
// this determines which lines can be redcted and write into the buffer
// there are may lines which needs to be stored, since partial matching to a secret
// Since we do not know which is the last call of write you need to call Flush
// on buffer to write the remaining lines
func (b *Buffer) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()

	// previous bytes may not ended with newline
	data := append(b.chunk, p...)
	b.chunk = []byte{}

	lastLines := b.lastLines(data)
	if len(lastLines) == 0 {
		// it is neccessary to return the count of incoming bytes
		return len(p), nil
	}

	for _, line := range lastLines {
		lines := b.store
		if lines == nil {
			lines = [][]byte{}
		}
		lines = append(lines, line)

		matchMap, partialMatchMap := b.matchSecrets(lines)

		var linesToPrint [][]byte
		linesToPrint, b.store = b.matchLines(lines, partialMatchMap)
		redactedLines := b.redact(linesToPrint, matchMap)

		redactedBytes := join(redactedLines)
		if c, err := b.Buff.Write(redactedBytes); err != nil {
			return c, err
		}
	}

	// it is neccessary to return the count of incoming bytes
	return len(p), nil
}

// Flush writes the remaining bytes
func (b *Buffer) Flush() error {
	lines := b.store[:]
	// chunk is the remaining part of the last Write call
	if len(b.chunk) > 0 {
		// lines are containing newline, but the chunk needs to be extendid with newline
		chunk := append(b.chunk, newLine...)
		lines = append(lines, chunk)
	}

	matchMap, _ := b.matchSecrets(lines)
	redactedLines := b.redact(lines, matchMap)

	redactedBytes := join(redactedLines)
	if _, err := b.Buff.Write(redactedBytes); err != nil {
		return err
	}
	return nil
}

// ReadLines iterally calls ReadString until it receives EOF
func (b *Buffer) ReadLines() ([]string, error) {
	b.Lock()
	defer b.Unlock()

	lines := []string{}
	eof := false
	for !eof {
		// every line's byte ends with newline
		line, err := b.Buff.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				eof = true
			} else {
				return nil, err
			}
		}
		// nothing red
		if len(line) == 0 {
			continue
		}
		line = strings.TrimSuffix(line, "\n")
		lines = append(lines, line)
	}
	return lines, nil
}

// lastLines splits the buffer's remaining bytes + p bytes by 'func split'
// and updates the buffer's chunk (the remaining bytes)
func (b *Buffer) lastLines(p []byte) [][]byte {
	p = append(b.chunk, p...)
	lines, chunk := split(p)
	b.chunk = chunk
	return lines
}

// matchSecrets collects which secrets matches from which line index
// and which secrets matches partially from which line index
func (b *Buffer) matchSecrets(lines [][]byte) (map[int][]int, map[int][]int) {
	matchMap := map[int][]int{}        // matching line chunk's first line indexes by secret index
	partialMatchMap := map[int][]int{} // partially matching line chunk's first line indexes by secret index

	for secretIdx, secret := range b.secrets {
		secretLine := secret[0] // every match should begin from the secret first line
		lineIndexes := []int{}  // the indexes of lines which contains the secret's first line

		for i, line := range lines {
			if bytes.Contains(line, secretLine) {
				lineIndexes = append(lineIndexes, i)
			}
		}

		if len(lineIndexes) == 0 {
			// this secret can not be found in the lines
			continue
		}

		for _, lineIdx := range lineIndexes {
			if len(secret) == 1 {
				// the single line secret found in the lines
				indexes := matchMap[secretIdx]
				if indexes == nil {
					indexes = []int{}
				}
				matchMap[secretIdx] = append(indexes, lineIdx)
				continue
			}

			// lineIdx. line matches to a multi line secret's first line
			// if lines has more line, every subsequent line must match to the secret's subsequent lines
			partialMatch := true
			match := false

			for i := lineIdx + 1; i < len(lines); i++ {
				secretLineIdx := i - lineIdx

				secretLine = secret[secretLineIdx]
				line := lines[i]

				if !bytes.Contains(line, secretLine) {
					partialMatch = false
					break
				}

				if secretLineIdx == len(secret)-1 {
					// multi line secret found in the lines
					match = true
					break
				}
			}

			if match {
				// multi line secret found in the lines
				indexes := matchMap[secretIdx]
				if indexes == nil {
					indexes = []int{}
				}
				matchMap[secretIdx] = append(indexes, lineIdx)
				continue
			}

			if partialMatch {
				// this secret partially can be found in the lines
				indexes := partialMatchMap[secretIdx]
				if indexes == nil {
					indexes = []int{}
				}
				partialMatchMap[secretIdx] = append(indexes, lineIdx)
			}
		}
	}

	return matchMap, partialMatchMap
}

// linesToKeepRange returns a range (first, last index) of lines needs to be observed
// since they contain partially matching secrets
func (b *Buffer) linesToKeepRange(partialMatchMap map[int][]int) (int, int) {
	first := -1
	last := -1

	for secretIdx, lineIndexes := range partialMatchMap {
		secret := b.secrets[secretIdx]
		secretLength := len(secret)

		for _, lineIdx := range lineIndexes {
			if first == -1 {
				first = lineIdx
				last = first + secretLength
				continue
			}

			if lineIdx+secretLength > last {
				last = lineIdx + secretLength
			}

			if first > lineIdx {
				first = lineIdx
			}
		}
	}

	return first, last
}

// matchLines return which lines can be printed and which should be keept for further observing
func (b *Buffer) matchLines(lines [][]byte, partialMatchMap map[int][]int) ([][]byte, [][]byte) {
	first, last := b.linesToKeepRange(partialMatchMap)
	if first == -1 {
		return lines[:], [][]byte{}
	}

	if first == 0 {
		if last > len(lines)-1 {
			return [][]byte{}, lines[0:len(lines)]
		}
		return lines[last:], lines[0:last]
	}

	if last > len(lines)-1 {
		return lines[:first], lines[first:len(lines)]
	}
	return lines[:first], lines[first:last]
}

// redact hides the given secret from the lines
func (b *Buffer) redact(lines [][]byte, matchMap map[int][]int) [][]byte {
	redacted := lines[:]
	for secretIdx, lineIndexes := range matchMap {
		secret := b.secrets[secretIdx]

		for _, lineIdx := range lineIndexes {
			if lineIdx > len(lines)-1 {
				continue
			}

			for i := lineIdx; i < lineIdx+len(secret); i++ {
				secretLine := secret[i-lineIdx]
				line := redacted[i]
				redacted[i] = bytes.Replace(line, secretLine, []byte(RedactStr), -1)
			}
		}
	}
	return redacted
}

// secretsByteList returns the list of secret byte lines
func secretsByteList(secrets []string) [][][]byte {
	secretByteLinesList := [][][]byte{}
	for _, secret := range secrets {
		secretBytes := []byte(secret)
		secretByteLines := bytes.Split(secretBytes, newLine)
		secretByteLinesList = append(secretByteLinesList, secretByteLines)
	}
	return secretByteLinesList
}

// split p after "\n", the split is assigned to lines
// if last line has no "\n" it is assigned to chunk
func split(p []byte) (lines [][]byte, chunk []byte) {
	if p == nil || len(p) == 0 {
		return [][]byte{}, []byte{}
	}

	lines = [][]byte{}
	chunk = p[:]
	for len(chunk) > 0 {
		idx := bytes.Index(chunk, newLine)
		if idx == -1 {
			return
		}

		lines = append(lines, chunk[:idx+1])

		if idx == len(chunk)-1 {
			chunk = []byte{}
		} else {
			chunk = chunk[idx+1:]
		}
	}
	return
}

// join lines and chunk to restore p after split(p)
func join(lines [][]byte) []byte {
	if lines == nil || len(lines) == 0 {
		return []byte{}
	}
	return bytes.Join(lines, []byte(""))
}
