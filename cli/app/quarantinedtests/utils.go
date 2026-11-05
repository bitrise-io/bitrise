package quarantinedtests

import (
	"fmt"
	"io"
	"strings"
	"time"

	internalapp "github.com/bitrise-io/bitrise/v3/internal/app"
)

// printQuarantinedTestText is the detail view shared by view, add and
// update.
func printQuarantinedTestText(w io.Writer, t internalapp.QuarantinedTest) error {
	var b strings.Builder
	fmt.Fprintf(&b, "ID:           %s\n", t.ID)
	fmt.Fprintf(&b, "Test case:    %s\n", t.TestCaseName)
	fmt.Fprintf(&b, "Suites:       %s\n", strings.Join(t.TestSuiteName, " › "))
	if t.ClassName != "" {
		fmt.Fprintf(&b, "Class:        %s\n", t.ClassName)
	}
	if t.TestCaseIdentifier != "" {
		fmt.Fprintf(&b, "Identifier:   %s\n", t.TestCaseIdentifier)
	}
	fmt.Fprintf(&b, "Branches:     %s\n", filterText(t.Filters.Branches))
	fmt.Fprintf(&b, "Targets:      %s\n", filterText(t.Filters.Targets))
	if t.Comment != "" {
		fmt.Fprintf(&b, "Comment:      %s\n", t.Comment)
	}
	fmt.Fprintf(&b, "Quarantined:  %s", t.QuarantinedAt.Local().Format(time.RFC3339))
	if t.QuarantinedBy != "" {
		fmt.Fprintf(&b, " by %s", t.QuarantinedBy)
	}
	fmt.Fprintln(&b)
	if src := sourceText(t.QuarantinedFrom); src != "" {
		fmt.Fprintf(&b, "From:         %s\n", src)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func sourceText(src internalapp.QuarantineSource) string {
	switch src.Source {
	case "api":
		return "API"
	case "insights":
		return "Insights"
	case "build", "pipeline":
		if src.Number != 0 {
			return fmt.Sprintf("%s #%d (%s)", src.Source, src.Number, src.ID)
		}
		return src.Source + " " + src.ID
	}
	return src.Source
}
