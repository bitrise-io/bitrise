//go:build integration

// Package activation is a native, in-process integration test for the steplib
// activation entry point (activator.ActivateSteplibRefStep). It exercises the v1
// (legacy local steplib) and v2 (API) paths against real URLs across version
// forms, precompiled/source fetch, cache state, offline mode, and error cases,
// and logs the raw activation logs of both paths side by side for eyeballing.
//
// Build-tagged `integration`; needs network. Run with:
//
//	go test -tags integration -v -timeout 30m ./_tests/activation/...
package activation

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bitrise-io/stepman/activator"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/stretchr/testify/assert"
)

const bitriseSteplibURL = "https://github.com/bitrise-io/bitrise-steplib.git"

// logEntry is one captured stepman.Logger call.
type logEntry struct {
	level string
	msg   string
}

// capturingLogger implements stepman.Logger and records every call so a case
// can print the raw activation logs. Threaded in as the real logger.
type capturingLogger struct {
	entries []logEntry
}

func (l *capturingLogger) add(level, format string, v ...any) {
	l.entries = append(l.entries, logEntry{level: level, msg: fmt.Sprintf(format, v...)})
}
func (l *capturingLogger) Debugf(format string, v ...any) { l.add("debug", format, v...) }
func (l *capturingLogger) Infof(format string, v ...any)  { l.add("info", format, v...) }
func (l *capturingLogger) Warnf(format string, v ...any)  { l.add("warn", format, v...) }
func (l *capturingLogger) Errorf(format string, v ...any) { l.add("error", format, v...) }

// variant selects which activation path a case exercises.
type variant struct {
	name        string
	useAPI      bool
	precompiled bool
}

var (
	v1Source      = variant{name: "v1-source", useAPI: false, precompiled: false}
	v1Precompiled = variant{name: "v1-precompiled", useAPI: false, precompiled: true}
	v2Source      = variant{name: "v2-source", useAPI: true, precompiled: false}
	v2Precompiled = variant{name: "v2-precompiled", useAPI: true, precompiled: true}
)

// steplibStep builds a canonical ID for a step in the bitrise steplib.
func steplibStep(id, version string) stepid.CanonicalID {
	return stepid.CanonicalID{SteplibSource: bitriseSteplibURL, IDorURI: id, Version: version}
}

// activationResult is the outcome of one ActivateSteplibRefStep call.
type activationResult struct {
	activated activator.ActivatedStep
	err       error
	logs      []logEntry
	elapsed   time.Duration
}

// activate runs one activation through the given variant, setting the path/fetch
// experiment env for the call and capturing its logs and timing.
func activate(t *testing.T, v variant, id stepid.CanonicalID, offline, didStepLibUpdate bool) activationResult {
	t.Helper()

	// The V2 API path targets the production inventory; there is no longer an
	// override to redirect it at a dev/test inventory.
	if v.useAPI {
		t.Setenv("BITRISE_STEPLIB_API_ENABLE", "true")
	} else {
		t.Setenv("BITRISE_STEPLIB_API_ENABLE", "false")
	}
	if v.precompiled {
		t.Setenv("BITRISE_EXPERIMENT_PRECOMPILED_STEPS", "true")
	} else {
		t.Setenv("BITRISE_EXPERIMENT_PRECOMPILED_STEPS", "false")
	}

	logger := &capturingLogger{}
	start := time.Now()
	activated, err := activator.ActivateSteplibRefStep(logger, id, t.TempDir(), t.TempDir(), didStepLibUpdate, offline)
	elapsed := time.Since(start)

	// Asserted here so every case below covers it, including the failing ones:
	// the inventory source is set before any early return, so a failed activation
	// stays attributable to the path that served it. assert, not require: callers
	// pair two activations and log both afterwards, and FailNow here would cut off
	// the side-by-side diagnostics this harness exists to print.
	wantInventorySource := activator.ActivationInventorySourceSteplib
	if v.useAPI {
		wantInventorySource = activator.ActivationInventorySourceSteplibAPI
	}
	assert.Equal(t, wantInventorySource, activated.ActivationInventorySource, "%s inventory source", v.name)

	return activationResult{activated: activated, err: err, logs: logger.entries, elapsed: elapsed}
}

// logResult prints a case's verdict, key result fields, and raw captured logs.
func logResult(t *testing.T, header string, r activationResult) {
	t.Helper()
	verdict := "OK"
	if r.err != nil {
		verdict = "FAIL"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "\n──────── %s ────────\n", header)
	fmt.Fprintf(&b, "verdict=%s  elapsed=%s  activationType=%q  inventorySource=%q  executable=%t  didStepLibUpdate=%t\n",
		verdict, r.elapsed.Round(time.Millisecond), r.activated.ActivationType, r.activated.ActivationInventorySource,
		r.activated.ExecutablePath != "", r.activated.DidStepLibUpdate)
	if r.err != nil {
		fmt.Fprintf(&b, "error: %s\n", r.err)
	}
	fmt.Fprintf(&b, "logs (%d):\n", len(r.logs))
	for _, e := range r.logs {
		fmt.Fprintf(&b, "  [%-5s] %s\n", e.level, e.msg)
	}
	t.Log(b.String())
}

// logPair runs and logs v1 then v2 for the same ref, adjacent, for an eyeball diff.
// Returns both results so the caller can assert.
func logPair(t *testing.T, id stepid.CanonicalID, v1, v2 variant, offline bool) (activationResult, activationResult) {
	t.Helper()
	r1 := activate(t, v1, id, offline, false)
	r2 := activate(t, v2, id, offline, false)
	t.Logf("=== %s @ %q ===", id.IDorURI, id.Version)
	logResult(t, v1.name, r1)
	logResult(t, v2.name, r2)
	return r1, r2
}

// evictFromCache removes a concrete version's source cache dir so the next v1
// activation is a guaranteed cache miss. No-op if the steplib isn't set up yet.
func evictFromCache(t *testing.T, id stepid.CanonicalID) {
	t.Helper()
	route, found := stepman.ReadRoute(id.SteplibSource)
	if !found {
		return
	}
	if err := os.RemoveAll(stepman.GetStepCacheDirPath(route, id.IDorURI, id.Version)); err != nil {
		t.Fatalf("evict cache for %s@%s: %s", id.IDorURI, id.Version, err)
	}
}
