package filterwriter

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

/*func Fuzz(data []byte) int {
	secrets := strings.Split(string(data[:50]), "\n")
	var buff bytes.Buffer
	out := New(secrets, &buff)
	wc, err := out.Write(data)
	if err != nil {
		panic("err nil")
	}
	if len(data) != wc {
		panic("data len")
	}
	_, err = out.Flush()
	if err != nil {
		panic("err")
	}
	return 0
}*/

func BenchmarkPerf(t *testing.B) {
	t.Log("multiple secret in the same line")
	{
		// randomReader := rand.New(rand.NewSource(int64(t.N)))
		seed := time.Now().UnixNano()
		t.Logf("Seed: %d", seed)
		randomReader := rand.New(rand.NewSource(time.Now().UnixNano()))

		numSecrets := randomReader.Intn(358)
		t.Logf("Num secrets %d", numSecrets)

		secrets := []string{}
		for i := 0; i < numSecrets; i++ {
			lenSecret := randomReader.Intn(3)
			buf := make([]byte, lenSecret)
			_, err := randomReader.Read(buf)
			if err != nil {
				t.Fatalf("err %s", err)
			}

			secrets = append(secrets, string(buf))
		}

		t.Logf("Secrets %s", secrets)

		dataLen := randomReader.Intn(130000)
		t.Logf("Data len: %d", dataLen)
		log := make([]byte, dataLen)
		_, err := randomReader.Read(log)
		require.NoError(t, err)

		var buff bytes.Buffer
		out := New(secrets, &buff)
		wc, err := out.Write(log)
		require.NoError(t, err)
		require.Equal(t, len(log), wc)

		_, err = out.flush()
		require.NoError(t, err)
		// require.Equal(t, "multiple secrets like: [REDACTED] and [REDACTED]\n[REDACTED] and some extra text", buff.String())
	}
}

func TestSecretsByteList(t *testing.T) {
	{
		secrets := []string{"secret value"}
		byteList := secretsByteList(secrets)
		require.Equal(t, [][][]byte{
			[][]byte{
				[]byte("secret value"),
			},
		}, byteList)
	}

	{
		secrets := []string{"secret value1", "secret value2"}
		byteList := secretsByteList(secrets)
		require.Equal(t, [][][]byte{
			[][]byte{
				[]byte("secret value1"),
			},
			[][]byte{
				[]byte("secret value2"),
			},
		}, byteList)
	}

	{
		secrets := []string{"multi\nline\nsecret"}
		byteList := secretsByteList(secrets)
		require.Equal(t, [][][]byte{
			[][]byte{
				[]byte("multi\n"),
				[]byte("line\n"),
				[]byte("secret"),
			},
		}, byteList)
	}

	{
		secrets := []string{"ending\nwith\nnewline\n"}
		byteList := secretsByteList(secrets)
		require.Equal(t, [][][]byte{
			[][]byte{
				[]byte("ending\n"),
				[]byte("with\n"),
				[]byte("newline\n"),
			},
		}, byteList)
	}

	{
		secrets := []string{"\nstarting\nwith\nnewline"}
		byteList := secretsByteList(secrets)
		require.Equal(t, [][][]byte{
			[][]byte{
				[]byte("\n"),
				[]byte("starting\n"),
				[]byte("with\n"),
				[]byte("newline"),
			},
		}, byteList)
	}

	{
		secrets := []string{"newlines\nin\n\nthe\n\n\nmiddle"}
		byteList := secretsByteList(secrets)
		require.Equal(t, [][][]byte{
			[][]byte{
				[]byte("newlines\n"),
				[]byte("in\n"),
				[]byte("\n"),
				[]byte("the\n"),
				[]byte("\n"),
				[]byte("\n"),
				[]byte("middle"),
			},
		}, byteList)
	}
}

func TestWrite(t *testing.T) {
	t.Log("trivial test")
	{
		var buff bytes.Buffer
		out := New([]string{"abc", "a\nb\nc"}, &buff)
		log := []byte("test with\nnew line\nand single line secret:abc\nand multiline secret:a\nb\nc")
		wc, err := out.Write(log)
		require.NoError(t, err)
		require.Equal(t, len(log), wc)

		err = out.Close()
		require.NoError(t, err)
		require.Equal(t, "test with\nnew line\nand single line secret:[REDACTED]\nand multiline secret:[REDACTED]\n[REDACTED]\n[REDACTED]", buff.String())
	}

	t.Log("chunk without newline")
	{
		var buff bytes.Buffer
		out := New([]string{"ab", "a\nb"}, &buff)
		log := []byte("test without newline, secret:ab")
		wc, err := out.Write(log)
		require.NoError(t, err)
		require.Equal(t, len(log), wc)

		err = out.Close()
		require.NoError(t, err)
		require.Equal(t, "test without newline, secret:[REDACTED]", buff.String())
	}

	t.Log("multiple secret in the same line")
	{
		var buff bytes.Buffer
		out := New([]string{"x1", "x\n2"}, &buff)
		log := []byte("multiple secrets like: x1 and x\n2 and some extra text")
		wc, err := out.Write(log)
		require.NoError(t, err)
		require.Equal(t, len(log), wc)

		err = out.Close()
		require.NoError(t, err)
		require.Equal(t, "multiple secrets like: [REDACTED] and [REDACTED]\n[REDACTED] and some extra text", buff.String())
	}
	/*
	   maxRun := 150000
	   t.Log("multiple secret in the same line with multiple gorutine ")

	   	{
	   		cherr := make(chan error, maxRun)
	   		chStr := make(chan string, maxRun)

	   		var buff bytes.Buffer
	   		out := New([]string{"x1", "x\n2"}, &buff)
	   		log := []byte("multiple secrets like: x1 and x\n2 and some extra text")
	   		for i := 0; i < maxRun; i++ {
	   			go func(buff bytes.Buffer, out *Writer, log []byte) {
	   				runtime.Gosched()
	   				buff.Reset()

	   				wc, err := out.Write(log)
	   				require.NoError(t, err)
	   				require.Equal(t, len(log), wc)

	   				err = out.Close()

	   				cherr <- err
	   				chStr <- buff.String()
	   			}(buff, out, log)
	   		}

	   		errCounter := 0
	   		for err := range cherr {
	   			require.NoError(t, err)

	   			errCounter++
	   			if errCounter == maxRun {
	   				close(cherr)
	   			}
	   		}

	   		strCounter := 0
	   		for str := range chStr {
	   			fmt.Println(str)

	   			strCounter++
	   			if strCounter == maxRun {
	   				close(chStr)
	   			}
	   		}
	   	}
	*/
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
	out := New(secrets, &buff)
	require.Equal(t, [][][]byte{
		[][]byte{[]byte("a\n"), []byte("b\n"), []byte("c")},
		[][]byte{[]byte("b")},
		[][]byte{[]byte("c\n"), []byte("b")},
		[][]byte{[]byte("x\n"), []byte("c\n"), []byte("b\n"), []byte("d")},
		[][]byte{[]byte("f")},
		[][]byte{[]byte(`a\nb\nc`)},
		[][]byte{[]byte(`c\nb`)},
		[][]byte{[]byte(`x\nc\nb\nd`)},
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
		[]byte("x\n"),
		[]byte("a\n"),
		[]byte("a\n"),
		[]byte("b\n"),
		[]byte("c\n"),
		[]byte("x\n"),
		[]byte("c\n"),
		[]byte("b\n")}

	var buff bytes.Buffer
	out := New(secrets, &buff)

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
	// 	[]byte("x\n"),
	// 	[]byte("a\n"),
	// 	[]byte("a\n"),
	// 	[]byte("b\n"),
	// 	[]byte("c\n"),
	// 	[]byte("x\n"), 5.line
	// 	[]byte("c\n"),
	// 	[]byte("b\n")}

	var buff bytes.Buffer
	out := New(secrets, &buff)

	partialMatchMap := map[int]bool{6: true, 2: true, 5: true, 7: true}
	first := out.linesToKeepRange(partialMatchMap) // minimal index in the partialMatchMap
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
		[]byte("x\n"), // 0.
		[]byte("a\n"),
		[]byte("a\n"), // 2.
		[]byte("b\n"),
		[]byte("c\n"), // 4.
		[]byte("x\n"),
		[]byte("c\n"), // 6.
		[]byte("b\n")}

	var buff bytes.Buffer
	out := New(secrets, &buff)

	_, partialMatchMap := out.matchSecrets(lines)
	print, remaining := out.matchLines(lines, partialMatchMap)
	require.Equal(t, [][]byte{
		[]byte("x\n"),
		[]byte("a\n"),
		[]byte("a\n"),
		[]byte("b\n"),
		[]byte("c\n"),
	}, print)
	require.Equal(t, [][]byte{
		[]byte("x\n"),
		[]byte("c\n"),
		[]byte("b\n"),
	}, remaining)
}

func TestSecretLinesToRedact(t *testing.T) {
	secrets := []string{
		"a\nb\nc",
		"b",
	}
	lines := [][]byte{
		[]byte("x\n"),
		[]byte("a\n"),
		[]byte("b\n"),
		[]byte("c\n"),
		[]byte("b\n"),
	}

	var buff bytes.Buffer
	out := New(secrets, &buff)

	matchMap, _ := out.matchSecrets(lines)
	require.Equal(t, map[int][]int{
		0: []int{1},
		1: []int{2, 4},
	}, matchMap)

	secretLines := out.secretLinesToRedact(0, matchMap)
	require.Equal(t, ([][]byte)(nil), secretLines, fmt.Sprintf("%s\n", secretLines))

	secretLines = out.secretLinesToRedact(1, matchMap)
	require.Equal(t, [][]byte{[]byte("a\n")}, secretLines, fmt.Sprintf("%s\n", secretLines))

	secretLines = out.secretLinesToRedact(2, matchMap)
	require.Equal(t, [][]byte{[]byte("b"), []byte("b\n")}, secretLines, fmt.Sprintf("%s\n", secretLines))

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
		[]byte("x\n"),
		[]byte("a\n"),
		[]byte("a\n"),
		[]byte("b\n"),
		[]byte("c\n"),
	}

	var buff bytes.Buffer
	out := New(secrets, &buff)

	matchMap := map[int][]int{0: []int{2}, 1: []int{3}}
	redacted := out.redact(lines, matchMap)
	require.Equal(t, [][]byte{
		[]byte("x\n"),
		[]byte("a\n"),
		[]byte(RedactStr + "\n"),
		[]byte(RedactStr + "\n"),
		[]byte(RedactStr + "\n"),
	}, redacted)

	{
		secrets := []string{
			"106\n105",
			"99",
		}
		lines := [][]byte{
			[]byte("106\n"),
			[]byte("105\n"),
			[]byte("104\n"),
			[]byte("103\n"),
			[]byte("102\n"),
			[]byte("101\n"),
			[]byte("100\n"),
			[]byte("99\n")}

		var buff bytes.Buffer
		out := New(secrets, &buff)

		matchMap := map[int][]int{
			0: []int{0},
			1: []int{7},
		}
		redacted := out.redact(lines, matchMap)
		require.Equal(t, [][]byte{
			[]byte(RedactStr + "\n"),
			[]byte(RedactStr + "\n"),
			[]byte("104" + "\n"),
			[]byte("103" + "\n"),
			[]byte("102" + "\n"),
			[]byte("101" + "\n"),
			[]byte("100" + "\n"),
			[]byte(RedactStr + "\n"),
		}, redacted, fmt.Sprintf("%s", redacted))
	}
}

func TestSplitAfterNewline(t *testing.T) {
	t.Log("bytes")
	{
		require.Equal(t, []byte{}, []byte(""))
	}

	t.Log("empty test")
	{
		b := []byte{}
		lines, chunk := splitAfterNewline(b)
		require.Equal(t, [][]byte(nil), lines)
		require.Equal(t, []byte{}, chunk)
	}

	t.Log("empty test - empty string bytes")
	{
		b := []byte("")
		lines, chunk := splitAfterNewline(b)
		require.Equal(t, [][]byte(nil), lines)
		require.Equal(t, []byte{}, chunk)
	}

	t.Log("newline test")
	{
		b := []byte("\n")
		lines, chunk := splitAfterNewline(b)
		require.Equal(t, [][]byte{[]byte("\n")}, lines)
		require.Equal(t, []byte(nil), chunk)
	}

	t.Log("multi line test")
	{
		b := []byte(`line 1
line 2
line 3
`)
		lines, chunk := splitAfterNewline(b)
		require.Equal(t, 3, len(lines))
		require.Equal(t, []byte("line 1\n"), lines[0])
		require.Equal(t, []byte("line 2\n"), lines[1])
		require.Equal(t, []byte("line 3\n"), lines[2])
		require.Equal(t, []byte(nil), chunk)
	}

	t.Log("multi line test - newlines")
	{
		b := []byte(`

line 1

line 2
`)

		lines, chunk := splitAfterNewline(b)
		require.Equal(t, 5, len(lines))
		require.Equal(t, []byte("\n"), lines[0])
		require.Equal(t, []byte("\n"), lines[1])
		require.Equal(t, []byte("line 1\n"), lines[2])
		require.Equal(t, []byte("\n"), lines[3])
		require.Equal(t, []byte("line 2\n"), lines[4])
		require.Equal(t, []byte(nil), chunk)
	}

	t.Log("chunk test")
	{
		b := []byte("line 1")
		lines, chunk := splitAfterNewline(b)
		require.Equal(t, [][]byte(nil), lines)
		require.Equal(t, []byte("line 1"), chunk)
	}

	t.Log("chunk test")
	{
		b := []byte(`line 1
line 2`)

		lines, chunk := splitAfterNewline(b)
		require.Equal(t, 1, len(lines))
		require.Equal(t, []byte("line 1\n"), lines[0])
		require.Equal(t, []byte("line 2"), chunk)
	}
	t.Log("chunk test")
	{
		b := []byte("test\n\ntest\n")

		lines, chunk := splitAfterNewline(b)
		require.Equal(t, 3, len(lines))
		require.Equal(t, []byte("test\n"), lines[0])
		require.Equal(t, []byte("\n"), lines[1])
		require.Equal(t, []byte("test\n"), lines[2])
		require.Equal(t, []byte(nil), chunk)
	}
}
