//go:build steplib_e2e

package steplibe2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSteplibActivationMatrix activates the curated step matrix through the v1
// and v2 steplib paths (source + precompiled), asserts every cell activates and
// runs, then diffs the v1/v2 activation logs and writes a parity report.
//
// It hits live hosted steplib infra, so it is gated behind the steplib_e2e build
// tag and driven by bitrise_e2e_steplib.yml — it does not run in the default
// integration test suite.
func TestSteplibActivationMatrix(t *testing.T) {
	results := map[string]cellResult{}
	var runStatus []string

	for _, c := range allCells() {
		c := c
		t.Run(c.name(), func(t *testing.T) {
			res := runCell(c)
			results[c.name()] = res
			if res.runErr != nil {
				runStatus = append(runStatus, fmt.Sprintf("FAILED %s: %v", c.name(), res.runErr))
				// Surface a tail of the output to aid debugging, but keep going so
				// the rest of the matrix and the report still complete.
				assert.NoError(t, res.runErr, "cell %s failed; output tail:\n%s", c.name(), tail(res.output, 1200))
				return
			}
			runStatus = append(runStatus, fmt.Sprintf("OK     %s (%d cli log lines)", c.name(), len(res.logs)))
			assert.NotEmpty(t, res.logs, "cell %s produced no bitrise_cli log lines", c.name())
		})
	}

	// Build per step+version comparisons: v1-source baseline vs each v2 variant.
	var comparisons []comparison
	for _, s := range steps() {
		for _, v := range s.versions {
			v1 := results[cellName(s.id, v.label, "v1-source")]
			cmp := comparison{
				step:         s.id,
				versionLabel: v.label,
				versionRef:   v.version,
				v1Status:     statusOf(v1),
			}
			for _, variant := range []string{"v2-source", "v2-precompiled"} {
				v2 := results[cellName(s.id, v.label, variant)]
				v1Only, v2Only := diffLogs(v1.logs, v2.logs)
				cmp.pairs = append(cmp.pairs, pairDiff{
					v2Variant: variant,
					v2Status:  statusOf(v2),
					v1Only:    v1Only,
					v2Only:    v2Only,
				})
			}
			comparisons = append(comparisons, cmp)
		}
	}

	path, err := writeReport(comparisons, runStatus)
	require.NoError(t, err, "write report")
	t.Logf("steplib log parity report written to: %s", path)
}

func statusOf(r cellResult) string {
	if r.runErr != nil {
		return fmt.Sprintf("FAILED: %v", r.runErr)
	}
	return fmt.Sprintf("OK (%d cli log lines)", len(r.logs))
}

func cellName(stepID, versionLabel, variant string) string {
	return fmt.Sprintf("%s_%s_%s", stepID, versionLabel, variant)
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
