package asynccmd

import (
	"bytes"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

// RedactStr ...
const RedactStr = "[REDACTED]"

var newLine = []byte("\n")

// Buffer ...
type Buffer struct {
	Buff bytes.Buffer
	sync.Mutex

	secrets [][][]byte

	chunk     []byte
	store     [][]byte
	lastWrite time.Time
}

func newBuffer(secrets []string) *Buffer {
	return &Buffer{
		Buff:    bytes.Buffer{},
		secrets: secretsByteList(secrets),
		Mutex:   sync.Mutex{},
	}
}

// Write implements io.Writer interface.
// Splits p into lines, the lines are matched against the secrets,
// this determines which lines can be redacted and written into the buffer.
// There are may lines which needs to be stored, since they are partial matching to a secret.
// Since we do not know which is the last call of Write we need to call Flush
// on buffer to write the remaining lines.
func (b *Buffer) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()

	// previous bytes may not ended with newline
	data := append(b.chunk, p...)

	var lastLines [][]byte
	lastLines, b.chunk = split(data)
	if len(lastLines) == 0 {
		// it is neccessary to return the count of incoming bytes
		return len(p), nil
	}

	for _, line := range lastLines {
		lines := append(b.store, line)
		matchMap, partialMatchIndexes := b.matchSecrets(lines)

		var linesToPrint [][]byte
		linesToPrint, b.store = b.matchLines(lines, partialMatchIndexes)
		if linesToPrint == nil {
			continue
		}

		redactedLines := b.redact(linesToPrint, matchMap)

		redactedBytes := bytes.Join(redactedLines, nil)
		c, err := b.Buff.Write(redactedBytes)
		b.lastWrite = time.Now()
		if err != nil {
			return c, err
		}
	}

	// it is neccessary to return the count of incoming bytes
	return len(p), nil
}

// Flush writes the remaining bytes.
func (b *Buffer) Flush() error {
	// chunk is the remaining part of the last Write call
	if len(b.chunk) > 0 {
		// lines are containing newline, but the chunk needs to be extendid with newline
		chunk := append(b.chunk, newLine...)
		b.chunk = nil

		b.store = append(b.store, chunk)
	}

	matchMap, _ := b.matchSecrets(b.store)
	redactedLines := b.redact(b.store, matchMap)
	b.store = nil

	redactedBytes := bytes.Join(redactedLines, nil)
	if _, err := b.Buff.Write(redactedBytes); err != nil {
		return err
	}
	b.lastWrite = time.Now()
	return nil
}

// ReadLines iterally calls ReadString until it receives EOF.
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
		// nothing read
		if len(line) == 0 {
			continue
		}
		line = strings.TrimSuffix(line, "\n")
		lines = append(lines, line)
	}
	return lines, nil
}

// matchSecrets collects which secrets matches from which line indexes
// and which secrets matches partially from which line indexes.
// matchMap: matching line chunk's first line indexes by secret index
// partialMatchIndexes: line indexes from which secrets matching but not fully contained in lines
func (b *Buffer) matchSecrets(lines [][]byte) (matchMap map[int][]int, partialMatchIndexes map[int]bool) {
	matchMap = make(map[int][]int)
	partialMatchIndexes = make(map[int]bool)

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
				matchMap[secretIdx] = append(indexes, lineIdx)
				continue
			}

			if partialMatch {
				// this secret partially can be found in the lines
				partialMatchIndexes[lineIdx] = true
			}
		}
	}

	return
}

// linesToKeepRange returns a range (first, last index) of lines needs to be observed
// since they contain partially matching secrets.
func (b *Buffer) linesToKeepRange(partialMatchIndexes map[int]bool) int {
	first := -1

	for lineIdx := range partialMatchIndexes {
		if first == -1 {
			first = lineIdx
			continue
		}

		if first > lineIdx {
			first = lineIdx
		}
	}

	return first
}

// matchLines return which lines can be printed and which should be keept for further observing.
func (b *Buffer) matchLines(lines [][]byte, partialMatchIndexes map[int]bool) ([][]byte, [][]byte) {
	first := b.linesToKeepRange(partialMatchIndexes)
	if first == -1 {
		// no lines needs to be kept
		return lines, nil
	}

	if first == 0 {
		// partial match is always longer then the lines
		return nil, lines
	}

	return lines[:first], lines[first:]
}

// secretLinesToRedact returns which secret lines should be redacted
func (b *Buffer) secretLinesToRedact(lineIdxToRedact int, matchMap map[int][]int) [][]byte {
	// which line is which secrets first matching line
	secretIdxsByLine := make(map[int][]int)
	for secretIdx, lineIndexes := range matchMap {
		for _, lineIdx := range lineIndexes {
			secretIdxsByLine[lineIdx] = append(secretIdxsByLine[lineIdx], secretIdx)
		}
	}

	var secretChunks [][]byte
	for firstMatchingLineIdx, secretIndexes := range secretIdxsByLine {
		if lineIdxToRedact < firstMatchingLineIdx {
			continue
		}

		for _, secretIdx := range secretIndexes {
			secret := b.secrets[secretIdx]

			if lineIdxToRedact > firstMatchingLineIdx+len(secret)-1 {
				continue
			}

			secretLineIdx := lineIdxToRedact - firstMatchingLineIdx
			secretChunks = append(secretChunks, secret[secretLineIdx])
		}
	}

	sort.Slice(secretChunks, func(i, j int) bool { return len(secretChunks[i]) < len(secretChunks[j]) })
	return secretChunks
}

// redact hides the given ranges in the given line.
func redact(line []byte, ranges []matchRange) []byte {
	var offset int // the offset of ranges generated by replacing x bytes by the RedactStr
	for _, r := range ranges {
		length := r.last - r.first
		first := r.first + offset
		last := first + length

		newLine := append([]byte{}, line[:first]...)
		newLine = append(newLine, RedactStr...)
		newLine = append(newLine, line[last:]...)

		offset += len(RedactStr) - length

		line = newLine
	}
	return line
}

// redact hides the given secrets in the given lines.
func (b *Buffer) redact(lines [][]byte, matchMap map[int][]int) [][]byte {
	secretIdxsByLine := map[int][]int{}
	for secretIdx, lineIndexes := range matchMap {
		for _, lineIdx := range lineIndexes {
			secretIdxsByLine[lineIdx] = append(secretIdxsByLine[lineIdx], secretIdx)
		}
	}

	for i, line := range lines {
		linesToRedact := b.secretLinesToRedact(i, matchMap)
		if linesToRedact == nil {
			continue
		}

		var ranges []matchRange
		for _, lineToRedact := range linesToRedact {
			ranges = append(ranges, allRanges(line, lineToRedact)...)
		}

		lines[i] = redact(line, mergeAllRanges(ranges))
	}

	return lines
}

// secretsByteList returns the list of secret byte lines.
func secretsByteList(secrets []string) [][][]byte {
	var s [][][]byte
	for _, secret := range secrets {
		s = append(s, bytes.Split([]byte(secret), newLine))
	}
	return s
}

// split splits p after "\n", the split is assigned to lines
// if last line has no "\n" it is assigned to chunk.
func split(p []byte) (lines [][]byte, chunk []byte) {
	chunk = p
	for len(chunk) > 0 {
		idx := bytes.Index(chunk, newLine)
		if idx == -1 {
			return
		}

		lines = append(lines, chunk[:idx+1])

		if idx == len(chunk)-1 {
			chunk = []byte{}
			continue
		}
		chunk = chunk[idx+1:]
	}
	return
}
