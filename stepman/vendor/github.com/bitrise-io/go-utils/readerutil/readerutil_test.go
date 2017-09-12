package readerutil

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadLongLine(t *testing.T) {
	t.Log("Empty string")
	{
		reader := bufio.NewReader(strings.NewReader(``))
		line, err := ReadLongLine(reader)
		require.Equal(t, io.EOF, err)
		require.Equal(t, "", line)
	}

	t.Log("Single line")
	{
		reader := bufio.NewReader(strings.NewReader(`a single line`))
		line, err := ReadLongLine(reader)
		require.NoError(t, err)
		require.Equal(t, "a single line", line)
		// read one more
		line, err = ReadLongLine(reader)
		require.Equal(t, io.EOF, err)
		require.Equal(t, "", line)
	}

	t.Log("Two lines")
	{
		reader := bufio.NewReader(strings.NewReader(`first line
second line`))
		// first line
		line, readErr := ReadLongLine(reader)
		require.NoError(t, readErr)
		require.Equal(t, "first line", line)
		// second line
		line, readErr = ReadLongLine(reader)
		require.NoError(t, readErr)
		require.Equal(t, "second line", line)
		// read one more
		line, readErr = ReadLongLine(reader)
		require.Equal(t, io.EOF, readErr)
		require.Equal(t, "", line)
	}

	t.Log("Multi line, with long line")
	{
		inputStr := fmt.Sprintf(`first line
second line
third, really long line: %s
  fourth line
`, strings.Repeat("-", 1000000))

		reader := bufio.NewReader(strings.NewReader(inputStr))
		//
		lines := []string{}
		line, readErr := ReadLongLine(reader)
		for ; readErr == nil; line, readErr = ReadLongLine(reader) {
			lines = append(lines, line)
		}
		// ideally the error will be io.EOF
		require.Equal(t, io.EOF, readErr)
		//
		require.Equal(t, 4, len(lines))
		require.Equal(t, "first line", lines[0])
		require.Equal(t, "second line", lines[1])
		// check the start of the long line
		require.Equal(t, "third, really long line: ---", lines[2][0:28])
		require.Equal(t, "  fourth line", lines[3])
	}
}

func TestWalkLinesString(t *testing.T) {
	t.Log("Empty string")
	{
		inputStr := ``
		//
		lines := []string{}
		err := WalkLinesString(inputStr, func(line string) error {
			lines = append(lines, line)
			return nil
		})
		require.Equal(t, nil, err)
		require.Equal(t, []string{}, lines)
	}

	t.Log("Single line")
	{
		inputStr := `a single line`
		//
		lines := []string{}
		err := WalkLinesString(inputStr, func(line string) error {
			lines = append(lines, line)
			return nil
		})
		require.Equal(t, nil, err)
		require.Equal(t, []string{"a single line"}, lines)
	}

	t.Log("Two lines")
	{
		inputStr := `first line
second line`
		//
		lines := []string{}
		err := WalkLinesString(inputStr, func(line string) error {
			lines = append(lines, line)
			return nil
		})
		require.Equal(t, nil, err)
		require.Equal(t, []string{"first line", "second line"}, lines)
	}

	t.Log("Multi line, with long line")
	{
		inputStr := fmt.Sprintf(`first line
second line
third, really long line: %s
  fourth line
`, strings.Repeat("-", 1000000))

		//
		lines := []string{}
		err := WalkLinesString(inputStr, func(line string) error {
			lines = append(lines, line)
			return nil
		})
		require.Equal(t, nil, err)
		//
		require.Equal(t, 4, len(lines))
		require.Equal(t, "first line", lines[0])
		require.Equal(t, "second line", lines[1])
		// check the start of the long line
		require.Equal(t, "third, really long line: ---", lines[2][0:28])
		require.Equal(t, "  fourth line", lines[3])
	}

	t.Log("Break early, with io.EOF")
	{
		inputStr := `first line
second line
break here
this should not be included`
		//
		lines := []string{}
		err := WalkLinesString(inputStr, func(line string) error {
			if line == "break here" {
				return io.EOF
			}
			lines = append(lines, line)
			return nil
		})
		require.Equal(t, nil, err)
		require.Equal(t, []string{"first line", "second line"}, lines)
	}

	t.Log("Break early, with a non io.EOF error")
	{
		inputStr := `first line
second line
break here
this should not be included`
		//
		lines := []string{}
		err := WalkLinesString(inputStr, func(line string) error {
			if line == "break here" {
				return errors.New("Please stop")
			}
			lines = append(lines, line)
			return nil
		})
		require.EqualError(t, err, "Please stop")
		require.Equal(t, []string{"first line", "second line"}, lines)
	}
}
