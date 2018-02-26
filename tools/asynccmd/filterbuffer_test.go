package asynccmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	RedactStr = "X"

	t.Log("trivial test")
	{
		buff := newBuffer([]string{"abc", "a\nb\nc"})
		log := []byte("test with\nnew line\nand single line secret:abc\nand multiline secret:a\nb\nc")
		wc, err := buff.Write(log)
		require.NoError(t, err)
		require.Equal(t, len(log), wc)

		require.NoError(t, buff.Flush())
		lines, err := buff.ReadLines()
		require.NoError(t, err)
		require.Equal(t, []string{
			"test with",
			"new line",
			"and single line secret:X",
			"Xnd multiline secret:X",
			"X",
			"X",
		}, lines)
	}

	t.Log("chunk without newline")
	{
		buff := newBuffer([]string{"ab", "a\nb"})
		log := []byte("test with without newline, secret:ab")
		wc, err := buff.Write(log)
		require.NoError(t, err)
		require.Equal(t, len(log), wc)

		require.NoError(t, buff.Flush())
		lines, err := buff.ReadLines()
		require.NoError(t, err)
		require.Equal(t, []string{
			"test with without newline, secret:X",
		}, lines)
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

	buff := newBuffer(secrets)
	require.Equal(t, [][][]byte{
		[][]byte{[]byte("a"), []byte("b"), []byte("c")},
		[][]byte{[]byte("b")},
		[][]byte{[]byte("c"), []byte("b")},
		[][]byte{[]byte("x"), []byte("c"), []byte("b"), []byte("d")},
		[][]byte{[]byte("f")},
	}, buff.secrets)
}

func Test_lastLines(t *testing.T) {
	buff := newBuffer([]string{})
	buff.chunk = []byte("te")

	lines := buff.lastLines([]byte("st\nlast\nlines"))
	require.Equal(t, [][]byte{
		[]byte("test\n"),
		[]byte("last\n"),
	}, lines)
	require.Equal(t, []byte("lines"), buff.chunk)
}

func Test_matchSecrets(t *testing.T) {
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

	buff := newBuffer(secrets)

	matchMap, partialMatchMap := buff.matchSecrets(lines)
	require.Equal(t, map[int][]int{
		0: []int{2},
		1: []int{3, 7},
		2: []int{6},
	}, matchMap)
	require.Equal(t, map[int][]int{
		3: []int{5},
	}, partialMatchMap)
}

func Test_linesToKeepRange(t *testing.T) {
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

	buff := newBuffer(secrets)

	partialMatchMap := map[int][]int{
		2: []int{6},
		0: []int{2},
		3: []int{5},
		1: []int{2, 7},
	}
	first, last := buff.linesToKeepRange(partialMatchMap)
	require.Equal(t, 2, first)
	require.Equal(t, 9, last)
}

func Test_matchLine(t *testing.T) {
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

	buff := newBuffer(secrets)

	_, partialMatchMap := buff.matchSecrets(lines)
	print, remaining := buff.matchLines(lines, partialMatchMap)
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

func Test_redact(t *testing.T) {
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

	buff := newBuffer(secrets)

	matchMap := map[int][]int{0: []int{2}, 1: []int{3}}
	redacted := buff.redact(lines, matchMap)
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

		buff := newBuffer(secrets)

		matchMap := map[int][]int{
			0: []int{0},
			1: []int{7},
		}
		redacted := buff.redact(lines, matchMap)
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

func Test_split(t *testing.T) {
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
		require.Equal(t, b, join(lines))
	}

	t.Log("newline test")
	{
		b := []byte("\n")
		lines, chunk := split(b)
		require.Equal(t, 1, len(lines))
		require.Equal(t, []byte("\n"), lines[0])
		require.Equal(t, 0, len(chunk))
		lines = append(lines, chunk)
		require.Equal(t, b, join(lines))
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
		lines = append(lines, chunk)
		require.Equal(t, b, join(lines))
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
		lines = append(lines, chunk)
		require.Equal(t, b, join(lines))
	}

	t.Log("chunk test")
	{
		b := []byte("line 1")
		lines, chunk := split(b)
		require.Equal(t, []byte("line 1"), chunk)
		require.Equal(t, 0, len(lines))
		lines = append(lines, chunk)
		require.Equal(t, b, join(lines), string(join(lines)))
	}

	t.Log("chunk test")
	{
		b := []byte(`line 1
line 2`)

		lines, chunk := split(b)
		require.Equal(t, 1, len(lines))
		require.Equal(t, []byte("line 1\n"), lines[0])
		require.Equal(t, []byte("line 2"), chunk)
		lines = append(lines, chunk)
		require.Equal(t, b, join(lines))
	}
}
