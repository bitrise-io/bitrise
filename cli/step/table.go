package step

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
)

func renderTable(headers []string, rows [][]string) string {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, row := range rows {
		_, _ = fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	_ = tw.Flush()
	return buf.String()
}
