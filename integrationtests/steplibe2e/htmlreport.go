//go:build steplib_e2e

package steplibe2e

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
)

// htmlReportPath mirrors reportPath but for the .html artifact.
func htmlReportPath() string {
	name := "steplib_log_parity_report.html"
	if d := os.Getenv("BITRISE_DEPLOY_DIR"); d != "" {
		return filepath.Join(d, name)
	}
	return filepath.Join(os.TempDir(), name)
}

const htmlStyle = `<style>
  :root { color-scheme: light dark; }
  body { font-family: -apple-system, system-ui, sans-serif; line-height: 1.5; max-width: 1100px; margin: 0 auto; padding: 1.5rem; }
  h1 { font-size: 1.6rem; } h2 { margin-top: 2rem; border-bottom: 1px solid #8884; padding-bottom: .3rem; }
  h3 { margin-top: 1.5rem; } h4 { margin: .8rem 0 .3rem; color: #888; font-weight: 600; }
  code, pre { font-family: ui-monospace, "SF Mono", Menlo, monospace; font-size: .82rem; }
  .badges { display: flex; flex-wrap: wrap; gap: .4rem; margin: .4rem 0; }
  .badge { padding: .1rem .5rem; border-radius: .4rem; font-size: .78rem; white-space: nowrap; }
  .ok { background: #1a7f3722; color: #1a7f37; } .fail { background: #cf222e22; color: #cf222e; }
  .diff { overflow-x: auto; background: #8881; border-radius: .4rem; padding: .5rem .7rem; margin: 0; }
  .diff .del { color: #cf222e; display: block; white-space: pre-wrap; }
  .diff .add { color: #1a7f37; display: block; white-space: pre-wrap; }
  .lvl { opacity: .6; }
  .nodiff { color: #888; font-style: italic; }
  details { margin: .3rem 0; } summary { cursor: pointer; color: #888; font-size: .82rem; }
  details pre { overflow-x: auto; background: #8881; border-radius: .4rem; padding: .5rem .7rem; white-space: pre-wrap; }
  .ref { color: #888; font-weight: normal; font-size: .85rem; }
  .errs { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  .errs pre { overflow-x: auto; background: #8881; border-radius: .4rem; padding: .5rem .7rem; white-space: pre-wrap; }
  @media (max-width: 700px) { .errs { grid-template-columns: 1fr; } }
</style>`

func writeHTMLReport(comparisons []comparison, failures []failureRow, runStatus []string) (string, error) {
	var b strings.Builder
	b.WriteString(htmlStyle)
	b.WriteString(`<h1>Steplib v1 vs v2 — activation log parity</h1>`)
	b.WriteString(`<p>Each step+version was activated through the legacy (v1) and API (v2) paths in JSON+debug log mode. ` +
		`Log lines below are the raw <code>bitrise_cli</code> messages (ANSI stripped); the diff matches them by ` +
		`level + a normalized form so timestamps/paths/durations don't create false diffs.</p>`)

	b.WriteString(`<h2>Matrix run status</h2><div class="badges">`)
	for _, s := range runStatus {
		cls := "ok"
		if strings.HasPrefix(s, "FAILED") {
			cls = "fail"
		}
		b.WriteString(fmt.Sprintf(`<span class="badge %s">%s</span>`, cls, html.EscapeString(s)))
	}
	b.WriteString(`</div>`)

	b.WriteString(`<h2>Per case — raw log diffs (v1 vs v2)</h2>`)
	for _, c := range comparisons {
		b.WriteString(fmt.Sprintf(`<h3>%s @ %s</h3>`, html.EscapeString(c.step), html.EscapeString(displayRef(c.versionRef))))
		b.WriteString(`<div class="badges">`)
		b.WriteString(statusBadge("v1-source", c.v1Status))
		for _, p := range c.pairs {
			b.WriteString(statusBadge(p.v2Variant, p.v2Status))
		}
		b.WriteString(`</div>`)
		b.WriteString(logDetails("v1-source full log", c.v1Logs))
		for _, p := range c.pairs {
			b.WriteString(fmt.Sprintf(`<h4>v1-source vs %s</h4>`, html.EscapeString(p.v2Variant)))
			if len(p.v1Only) == 0 && len(p.v2Only) == 0 {
				b.WriteString(`<p class="nodiff">No log divergences.</p>`)
			} else {
				b.WriteString(`<pre class="diff">`)
				for _, l := range p.v1Only {
					b.WriteString(fmt.Sprintf(`<span class="del">- <span class="lvl">[%s]</span> %s</span>`, html.EscapeString(l.Level), html.EscapeString(l.Raw)))
				}
				for _, l := range p.v2Only {
					b.WriteString(fmt.Sprintf(`<span class="add">+ <span class="lvl">[%s]</span> %s</span>`, html.EscapeString(l.Level), html.EscapeString(l.Raw)))
				}
				b.WriteString(`</pre>`)
			}
			b.WriteString(logDetails(p.v2Variant+" full log", p.v2Logs))
		}
	}

	b.WriteString(`<h2>Failure-case coverage</h2>`)
	b.WriteString(`<p>Each case must fail to activate through both paths; the error each path reports is shown side by side.</p>`)
	for _, f := range failures {
		b.WriteString(fmt.Sprintf(`<h3>%s <span class="ref">%s</span></h3>`, html.EscapeString(f.name), html.EscapeString(f.ref)))
		b.WriteString(fmt.Sprintf(`<p>%s</p>`, html.EscapeString(f.desc)))
		b.WriteString(`<div class="errs">`)
		b.WriteString(fmt.Sprintf(`<div><div class="badges">%s</div><pre>%s</pre></div>`, verdictBadge("v1-source", f.v1Failed), html.EscapeString(f.v1Message)))
		b.WriteString(fmt.Sprintf(`<div><div class="badges">%s</div><pre>%s</pre></div>`, verdictBadge("v2-source", f.v2Failed), html.EscapeString(f.v2Message)))
		b.WriteString(`</div>`)
	}

	path := htmlReportPath()
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func statusBadge(label, status string) string {
	cls := "ok"
	if strings.HasPrefix(status, "FAILED") {
		cls = "fail"
	}
	return fmt.Sprintf(`<span class="badge %s">%s: %s</span>`, cls, html.EscapeString(label), html.EscapeString(status))
}

func verdictBadge(label string, failed bool) string {
	if failed {
		return fmt.Sprintf(`<span class="badge ok">%s: failed (expected)</span>`, html.EscapeString(label))
	}
	return fmt.Sprintf(`<span class="badge fail">%s: UNEXPECTEDLY SUCCEEDED</span>`, html.EscapeString(label))
}

func logDetails(summary string, lines []logLine) string {
	if len(lines) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<details><summary>%s (%d lines)</summary><pre>`, html.EscapeString(summary), len(lines)))
	for _, l := range lines {
		b.WriteString(fmt.Sprintf("<span class=\"lvl\">[%s]</span> %s\n", html.EscapeString(l.Level), html.EscapeString(l.Raw)))
	}
	b.WriteString(`</pre></details>`)
	return b.String()
}
