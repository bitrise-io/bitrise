//go:build steplib_e2e

package steplibe2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
)

// cellResult is the outcome of activating+running one matrix cell.
type cellResult struct {
	cell    cell
	output  string // raw combined stdout+stderr (JSON log lines)
	runErr  error  // non-nil if the CLI exited non-zero
	logs    []logLine
}

// stepRef renders the workflow step key, e.g. "git-clone@8.4.0" or "git-clone"
// for the latest (empty) version form.
func (c cell) stepRef() string {
	if c.version.version == "" {
		return c.step.id
	}
	return c.step.id + "@" + c.version.version
}

// name is a stable, filesystem-safe identifier for the cell.
func (c cell) name() string {
	return fmt.Sprintf("%s_%s_%s", c.step.id, c.version.label, c.variant.name)
}

// workflowYML builds a single-step workflow that activates the cell's step.
func (c cell) workflowYML() string {
	var b strings.Builder
	b.WriteString("format_version: \"13\"\n")
	b.WriteString("default_step_lib_source: " + canonicalSteplibURL + "\n")
	b.WriteString("project_type: other\n")
	b.WriteString("workflows:\n")
	b.WriteString("  e2e:\n")
	b.WriteString("    steps:\n")
	b.WriteString("    - " + c.stepRef() + ":\n")
	b.WriteString("        run_if: \"true\"\n")
	if len(c.step.inputs) > 0 {
		b.WriteString("        inputs:\n")
		for k, v := range c.step.inputs {
			b.WriteString("        - " + k + ": " + yamlScalar(v) + "\n")
		}
	}
	return b.String()
}

// yamlScalar quotes a value as a YAML block/flow scalar safe for multi-line and
// special-character content.
func yamlScalar(v string) string {
	if strings.ContainsAny(v, "\n") {
		// Literal block scalar; indent each line under the input mapping.
		var b strings.Builder
		b.WriteString("|-\n")
		for _, ln := range strings.Split(v, "\n") {
			b.WriteString("            " + ln + "\n")
		}
		return strings.TrimRight(b.String(), "\n")
	}
	return fmt.Sprintf("%q", v)
}

// cellEnv builds the per-cell environment. All three experiment flags are set
// explicitly (not just appended) so a cell is deterministic regardless of what
// the orchestrator exported into the process environment.
func (c cell) cellEnv(sourceDir string) []string {
	migrate := "false"
	if c.variant.useAPI {
		migrate = "true"
	}
	precompiled := "false"
	if c.variant.precompiled {
		precompiled = "true"
	}
	return []string{
		"BITRISE_SOURCE_DIR=" + sourceDir,
		"BITRISE_EXPERIMENT_STEPLIB_API_ENABLE_MIGRATE=" + migrate,
		"BITRISE_EXPERIMENT_STEPLIB_API_URL_OVERRIDE=" + devInventoryURL(),
		"BITRISE_EXPERIMENT_PRECOMPILED_STEPS=" + precompiled,
	}
}

// runCell writes the cell's workflow to a temp dir and runs it through the CLI
// in JSON+debug mode, capturing the full structured log stream.
func runCell(c cell) cellResult {
	dir, err := os.MkdirTemp("", "steplibe2e-"+c.step.id+"-")
	if err != nil {
		return cellResult{cell: c, runErr: fmt.Errorf("create temp dir: %w", err)}
	}
	sourceDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		return cellResult{cell: c, runErr: fmt.Errorf("create source dir: %w", err)}
	}
	wfPath := filepath.Join(dir, "workflow.yml")
	if err := os.WriteFile(wfPath, []byte(c.workflowYML()), 0o644); err != nil {
		return cellResult{cell: c, runErr: fmt.Errorf("write workflow: %w", err)}
	}

	cmd := command.New(
		testhelpers.BinPath(),
		"run", "e2e",
		"--config", wfPath,
		"--output-format", "json",
		"--debug",
	).SetDir(sourceDir).AppendEnvs(c.cellEnv(sourceDir)...)

	out, runErr := cmd.RunAndReturnTrimmedCombinedOutput()
	return cellResult{
		cell:   c,
		output: out,
		runErr: runErr,
		logs:   parseCLILogs(out),
	}
}
