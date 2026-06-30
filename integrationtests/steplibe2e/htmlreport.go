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
  :root {
    color-scheme: light;
    --ink: #1b2027; --ground: #fcfdfe; --surface: #eef2f6; --line: #d8dee6;
    --muted: #5b6675; --accent: #2f6c9e; --del: #c0392b; --add: #1e7d4f;
    --sans: ui-sans-serif, -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    --mono: ui-monospace, "SF Mono", "JetBrains Mono", Menlo, Consolas, monospace;
  }
  * { box-sizing: border-box; }
  body { font-family: var(--sans); color: var(--ink); background: var(--ground);
    line-height: 1.55; max-width: 1040px; margin: 0 auto; padding: 2.5rem 1.5rem 4rem;
    font-variant-numeric: tabular-nums; }
  h1 { font-size: 1.7rem; letter-spacing: -.01em; margin: 0 0 .4rem; text-wrap: balance; }
  .lede { color: var(--muted); max-width: 68ch; margin: 0 0 1rem; }
  h2 { font-size: 1.22rem; letter-spacing: -.01em; margin: 2.6rem 0 1rem;
    padding-bottom: .35rem; border-bottom: 2px solid var(--accent); }
  h3 { font-size: 1.05rem; margin: 1.8rem 0 .5rem; font-family: var(--mono); }
  h4 { margin: 1rem 0 .35rem; font-size: .72rem; text-transform: uppercase;
    letter-spacing: .07em; color: var(--muted); font-weight: 700; }
  code { font-family: var(--mono); font-size: .85em; }
  .badges { display: flex; flex-wrap: wrap; gap: .4rem; margin: .5rem 0 .2rem; }
  .badge { padding: .12rem .55rem; border-radius: 999px; font-size: .76rem; font-weight: 600;
    white-space: nowrap; border: 1px solid transparent; }
  .ok { background: #e7f3ec; color: var(--add); border-color: #bfe0cd; }
  .fail { background: #fbe9e7; color: var(--del); border-color: #f0c5be; }
  pre { font-family: var(--mono); font-size: .8rem; margin: 0; overflow-x: auto;
    background: var(--surface); border: 1px solid var(--line); border-radius: .5rem; padding: .6rem .8rem; }
  .diff { border-left: 3px solid var(--accent); }
  .diff .del, .diff .add { display: block; white-space: pre-wrap; word-break: break-word; }
  .diff .del { color: var(--del); } .diff .add { color: var(--add); }
  .lvl { opacity: .55; }
  .nodiff { color: var(--muted); font-style: italic; margin: .2rem 0; }
  details { margin: .4rem 0; }
  summary { cursor: pointer; color: var(--accent); font-size: .8rem; }
  summary:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
  details pre { margin-top: .35rem; white-space: pre-wrap; word-break: break-word; }
  .ref { color: var(--muted); font-weight: 400; font-size: .85rem; font-family: var(--mono); }
  .errs { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-top: .4rem; }
  .errs pre { white-space: pre-wrap; word-break: break-word; }
  @media (max-width: 720px) { .errs { grid-template-columns: 1fr; } }
</style>`

func writeHTMLReport(comparisons []comparison, failures []failureRow, runStatus []string) (string, error) {
	var b strings.Builder
	b.WriteString(htmlStyle)
	b.WriteString(`<h1>Steplib v1 vs v2 — activation log parity</h1>`)
	b.WriteString(`<p class="lede">Each step+version was activated through the legacy (v1) and API (v2) paths in JSON+debug log mode. ` +
		`Log lines below are the raw <code>bitrise_cli</code> messages (ANSI stripped); the diff matches them by ` +
		`level + a normalized form so timestamps, paths and durations don't create false diffs.</p>`)

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
