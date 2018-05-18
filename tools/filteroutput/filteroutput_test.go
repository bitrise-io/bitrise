package filteroutput

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	t.Log("trivial test")
	{
		var buff bytes.Buffer
		out := NewOutput([]string{"abc", "a\nb\nc"}, &buff)
		log := []byte("test with\nnew line\nand single line secret:abc\nand multiline secret:a\nb\nc")
		wc, err := out.Write(log)
		require.NoError(t, err)
		require.Equal(t, len(log), wc)

		_, err = out.Flush()
		require.NoError(t, err)
		require.Equal(t, strings.Join([]string{
			"test with",
			"new line",
			"and single line secret:[REDACTED]",
			"[REDACTED]nd multiline secret:[REDACTED]",
			"[REDACTED]",
			"[REDACTED]",
		}, "\n"), buff.String())
	}

	t.Log("chunk without newline")
	{
		var buff bytes.Buffer
		out := NewOutput([]string{"ab", "a\nb"}, &buff)
		log := []byte("test without newline, secret:ab")
		wc, err := out.Write(log)
		require.NoError(t, err)
		require.Equal(t, len(log), wc)

		_, err = out.Flush()
		require.NoError(t, err)
		require.Equal(t, strings.Join([]string{
			"test without newline, secret:[REDACTED]",
		}, "\n"), buff.String())
	}

	t.Log("multiple secret in the same line")
	{
		var buff bytes.Buffer
		out := NewOutput([]string{"x1", "x\n2"}, &buff)
		log := []byte("multiple secrets like: x1 and x\n2 and some extra text")
		wc, err := out.Write(log)
		require.NoError(t, err)
		require.Equal(t, len(log), wc)

		_, err = out.Flush()
		require.NoError(t, err)
		require.Equal(t, strings.Join([]string{
			"multiple secrets like: [REDACTED] and [REDACTED]",
			"[REDACTED] and some extra text",
		}, "\n"), buff.String())
	}
}

func TestSecrets(t *testing.T) {
	secrets := []string{
		"a\nb\nc",
		"b",
		"c\nb",
		"x\nc\nb\nd",
		"f",
	}

	var buff bytes.Buffer
	out := NewOutput(secrets, &buff)
	require.Equal(t, [][][]byte{
		[][]byte{[]byte("a"), []byte("b"), []byte("c")},
		[][]byte{[]byte("b")},
		[][]byte{[]byte("c"), []byte("b")},
		[][]byte{[]byte("x"), []byte("c"), []byte("b"), []byte("d")},
		[][]byte{[]byte("f")},
	}, out.secrets)
}

func TestMatchSecrets(t *testing.T) {
	secrets := []string{
		"a\nb\nc",
		"b",
		"c\nb",
		"x\nc\nb\nd",
		"f",
	}
	lines := [][]byte{
		[]byte("x"),
		[]byte("a"),
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
		[]byte("x"),
		[]byte("c"),
		[]byte("b")}

	var buff bytes.Buffer
	out := NewOutput(secrets, &buff)

	matchMap, partialMatchMap := out.matchSecrets(lines)
	require.Equal(t, map[int][]int{
		0: []int{2},
		1: []int{3, 7},
		2: []int{6},
	}, matchMap)
	require.Equal(t, map[int]bool{5: true}, partialMatchMap)
}

func TestLinesToKeepRange(t *testing.T) {
	t.Log()
	secrets := []string{
		"a\nb\nc",
		"b",
		"c\nb",
		"x\nc\nb\nd",
	}
	// lines := [][]byte{
	// 	[]byte("x"),
	// 	[]byte("a"),
	// 	[]byte("a"),
	// 	[]byte("b"),
	// 	[]byte("c"),
	// 	[]byte("x"), 5.line
	// 	[]byte("c"),
	// 	[]byte("b")}

	var buff bytes.Buffer
	out := NewOutput(secrets, &buff)

	partialMatchMap := map[int]bool{6: true, 2: true, 5: true, 7: true}
	first := out.linesToKeepRange(partialMatchMap)
	require.Equal(t, 2, first)
}

func TestMatchLine(t *testing.T) {
	secrets := []string{
		"a\nb\nc",
		"b",
		"c\nb",
		"x\nc\nb\nd",
		"f",
	}
	lines := [][]byte{
		[]byte("x"), // 0.
		[]byte("a"),
		[]byte("a"), // 2.
		[]byte("b"),
		[]byte("c"), // 4.
		[]byte("x"),
		[]byte("c"), // 6.
		[]byte("b")}

	var buff bytes.Buffer
	out := NewOutput(secrets, &buff)

	_, partialMatchMap := out.matchSecrets(lines)
	print, remaining := out.matchLines(lines, partialMatchMap)
	require.Equal(t, [][]byte{
		[]byte("x"),
		[]byte("a"),
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}, print)
	require.Equal(t, [][]byte{
		[]byte("x"),
		[]byte("c"),
		[]byte("b"),
	}, remaining)
}

func TestSecretLinesToRedact(t *testing.T) {
	secrets := []string{
		"a\nb\nc",
		"b",
	}
	lines := [][]byte{
		[]byte("x"),
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
		[]byte("b"),
	}

	var buff bytes.Buffer
	out := NewOutput(secrets, &buff)

	matchMap, _ := out.matchSecrets(lines)
	require.Equal(t, map[int][]int{
		0: []int{1},
		1: []int{2, 4},
	}, matchMap)

	secretLines := out.secretLinesToRedact(0, matchMap)
	require.Equal(t, ([][]byte)(nil), secretLines, fmt.Sprintf("%s\n", secretLines))

	secretLines = out.secretLinesToRedact(1, matchMap)
	require.Equal(t, [][]byte{[]byte("a")}, secretLines, fmt.Sprintf("%s\n", secretLines))

	secretLines = out.secretLinesToRedact(2, matchMap)
	require.Equal(t, [][]byte{[]byte("b"), []byte("b")}, secretLines, fmt.Sprintf("%s\n", secretLines))

	secretLines = out.secretLinesToRedact(3, matchMap)
	require.Equal(t, [][]byte{[]byte("c")}, secretLines, fmt.Sprintf("%s\n", secretLines))

	secretLines = out.secretLinesToRedact(4, matchMap)
	require.Equal(t, [][]byte{[]byte("b")}, secretLines, fmt.Sprintf("%s\n", secretLines))
}

func TestRedactLine(t *testing.T) {
	t.Log("redacts the middle of the line")
	{
		line := []byte("asdfabcasdf")
		ranges := []matchRange{ //  asdfabcasdf
			{first: 4, last: 7}, // ****abc****
		}
		redacted := redact(line, ranges)
		require.Equal(t, []byte("asdf[REDACTED]asdf"), redacted, string(redacted))
	}

	t.Log("redacts the begining of the line")
	{
		line := []byte("asdfabcasdf")
		ranges := []matchRange{ //  asdfabcasdf
			{first: 0, last: 5}, // asdfa******
		}
		redacted := redact(line, ranges)
		require.Equal(t, []byte("[REDACTED]bcasdf"), redacted, string(redacted))
	}

	t.Log("redacts the end of the line")
	{
		line := []byte("asdfabcasdf")
		ranges := []matchRange{ //   asdfabcasdf
			{first: 9, last: 11}, // *********df
		}
		redacted := redact(line, ranges)
		require.Equal(t, []byte("asdfabcas[REDACTED]"), redacted, string(redacted))
	}

	t.Log("redacts multiple secrets")
	{
		line := []byte("asdfabcasdf")
		ranges := []matchRange{ //   asdfabcasdf
			{first: 4, last: 7},  // ****abc****
			{first: 8, last: 10}, // ********sd*
		}
		redacted := redact(line, ranges)
		require.Equal(t, []byte("asdf[REDACTED]a[REDACTED]f"), redacted, string(redacted))
	}

	t.Log("redacts the whole line")
	{
		line := []byte("asdfabcasdf")
		ranges := []matchRange{ //   asdfabcasdf
			{first: 0, last: 4},  // asdf*******
			{first: 7, last: 11}, // *******asdf
			{first: 3, last: 9},  // ***fabcas**
		}
		ranges = mergeAllRanges(ranges)
		redacted := redact(line, ranges)
		require.Equal(t, []byte("[REDACTED]"), redacted, string(redacted))
	}
}

func TestRedact(t *testing.T) {
	secrets := []string{
		"a\nb\nc",
		"b",
	}
	lines := [][]byte{
		[]byte("x"),
		[]byte("a"),
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	var buff bytes.Buffer
	out := NewOutput(secrets, &buff)

	matchMap := map[int][]int{0: []int{2}, 1: []int{3}}
	redacted := out.redact(lines, matchMap)
	require.Equal(t, [][]byte{
		[]byte("x"),
		[]byte("a"),
		[]byte(RedactStr),
		[]byte(RedactStr),
		[]byte(RedactStr),
	}, redacted)

	{
		secrets := []string{
			"106\n105",
			"99",
		}
		lines := [][]byte{
			[]byte("106"),
			[]byte("105"),
			[]byte("104"),
			[]byte("103"),
			[]byte("102"),
			[]byte("101"),
			[]byte("100"),
			[]byte("99")}

		var buff bytes.Buffer
		out := NewOutput(secrets, &buff)

		matchMap := map[int][]int{
			0: []int{0},
			1: []int{7},
		}
		redacted := out.redact(lines, matchMap)
		require.Equal(t, [][]byte{
			[]byte(RedactStr),
			[]byte(RedactStr),
			[]byte("104"),
			[]byte("103"),
			[]byte("102"),
			[]byte("101"),
			[]byte("100"),
			[]byte(RedactStr),
		}, redacted, fmt.Sprintf("%s", redacted))
	}
}

func TestSplit(t *testing.T) {
	t.Log("bytes")
	{
		require.Equal(t, []byte{}, []byte(""))
	}

	t.Log("empty test")
	{
		b := []byte{}
		lines, chunk := split(b)
		require.Equal(t, 0, len(lines))
		require.Equal(t, 0, len(chunk))
	}

	t.Log("empty test - empty string bytes")
	{
		b := []byte("")
		lines, chunk := split(b)
		require.Equal(t, 0, len(lines))
		require.Equal(t, 0, len(chunk))
	}

	t.Log("newline test")
	{
		b := []byte("\n")
		lines, chunk := split(b)
		require.Equal(t, 1, len(lines))
		require.Equal(t, []byte("\n"), lines[0])
		require.Equal(t, 0, len(chunk))
	}

	t.Log("multi line test")
	{
		b := []byte(`line 1
line 2
line 3
`)
		lines, chunk := split(b)
		require.Equal(t, 3, len(lines))
		require.Equal(t, []byte("line 1\n"), lines[0])
		require.Equal(t, []byte("line 2\n"), lines[1])
		require.Equal(t, []byte("line 3\n"), lines[2])
		require.Equal(t, 0, len(chunk))
	}

	t.Log("multi line test - newlines")
	{
		b := []byte(`

line 1

line 2
`)

		lines, chunk := split(b)
		require.Equal(t, 5, len(lines))
		require.Equal(t, []byte("\n"), lines[0])
		require.Equal(t, []byte("\n"), lines[1])
		require.Equal(t, []byte("line 1\n"), lines[2])
		require.Equal(t, []byte("\n"), lines[3])
		require.Equal(t, []byte("line 2\n"), lines[4])
		require.Equal(t, 0, len(chunk))
	}

	t.Log("chunk test")
	{
		b := []byte("line 1")
		lines, chunk := split(b)
		require.Equal(t, []byte("line 1"), chunk)
		require.Equal(t, 0, len(lines))
	}

	t.Log("chunk test")
	{
		b := []byte(`line 1
line 2`)

		lines, chunk := split(b)
		require.Equal(t, 1, len(lines))
		require.Equal(t, []byte("line 1\n"), lines[0])
		require.Equal(t, []byte("line 2"), chunk)
	}
}
