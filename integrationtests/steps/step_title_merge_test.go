//go:build linux_and_mac
// +build linux_and_mac

package steps

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/integrationtests/internal/testhelpers"
	"github.com/bitrise-io/go-utils/command"
	"github.com/stretchr/testify/require"
)

// Test_StepTitleMerging covers the step title merging in WorkflowRunner.activateStep across all
// activation types (path, steplib, git): with no explicit workflow title, the step.yml title is
// shown; an explicit workflow title wins; and on activation failure the title falls back to the
// reference seeded by newStepInfoPtr.
func Test_StepTitleMerging(t *testing.T) {
	configPth := "bitrise.yml"

	run := func(workflow string) (string, error) {
		cmd := command.New(testhelpers.BinPath(), "run", workflow, "--config", configPth)
		cmd.SetDir("step_title_merge")
		return cmd.RunAndReturnTrimmedCombinedOutput()
	}

	t.Log("path activation: merges the step.yml title when the workflow doesn't set one")
	{
		out, err := run("path_merged_title")
		require.NoError(t, err, out)
		// The step.yml title is shown in the step header box instead of the raw "path::./" reference.
		require.Contains(t, out, "Local Title From Step YML", out)
	}

	t.Log("steplib activation: merges the step.yml title when the workflow doesn't set one")
	{
		out, err := run("steplib_merged_title")
		require.NoError(t, err, out)
		// The steplib step.yml title ("Script") is shown instead of the raw step id ("script").
		require.Contains(t, out, "Script", out)
	}

	t.Log("git activation: merges the step.yml title when the workflow doesn't set one")
	{
		out, err := run("git_merged_title")
		require.NoError(t, err, out)
		// The git step.yml title ("Script") is shown instead of the raw "git::..." reference.
		require.Contains(t, out, "Script", out)
	}

	t.Log("step.yml title + bitrise.yml title: keeps the explicit workflow title")
	{
		out, err := run("explicit_title")
		require.NoError(t, err, out)
		require.Contains(t, out, "Explicit Workflow Title", out)
		require.NotContains(t, out, "Local Title From Step YML", out)
	}

	t.Log("empty step.yml title + bitrise.yml title: shows the workflow title")
	{
		out, err := run("workflow_title_no_step_title")
		require.NoError(t, err, out)
		require.Contains(t, out, "Workflow Title Without Step Title", out)
	}

	t.Log("empty step.yml title + empty bitrise.yml title: falls back to the step reference")
	{
		out, err := run("no_title_at_all")
		require.NoError(t, err, out)
		// Nothing to merge from either side, so the defaulted reference title is shown.
		require.Contains(t, out, "path::./step_no_title", out)
	}

	t.Log("activation failure, no title: falls back to the step reference")
	{
		out, err := run("activation_error")
		// Activation fails, so the build fails with a non-zero exit code.
		require.Error(t, err, out)
		require.Contains(t, out, "the provided directory doesn't exist", out)
		// activateStep returns before the step info is filled from the activation result, so the
		// header/summary keep the defaulted step reference as the title and no step.yml title is merged.
		require.Contains(t, out, "path::./this-path-does-not-exist", out)
		require.NotContains(t, out, "Local Title From Step YML", out)
	}

	t.Log("activation failure, title set: keeps the explicit title on the failing step")
	{
		out, err := run("activation_error_with_title")
		require.Error(t, err, out)
		require.Contains(t, out, "the provided directory doesn't exist", out)
		// The explicit bitrise.yml title (seeded by newStepInfoPtr) is shown for the failing step.
		require.Contains(t, out, "Title For A Failing Step", out)
	}
}
