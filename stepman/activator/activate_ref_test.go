package activator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/stretchr/testify/require"
)

const testStepYML = `title: Hello Step
summary: Says hello.
description: A simple example step.
toolkit:
  bash:
    entry_file: step.sh
inputs:
- World: ""
  opts:
    title: Greeting target
`

func writeTestStepYML(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "step.yml"), []byte(testStepYML), 0600))
}

func TestActivatePathRefStep(t *testing.T) {
	stepSrcDir := t.TempDir()
	writeTestStepYML(t, stepSrcDir)

	activatedStepDir := t.TempDir()
	workDir := t.TempDir()

	id := stepid.CanonicalID{SteplibSource: "path", IDorURI: stepSrcDir}

	got, err := ActivatePathRefStep(TestLogger[*testing.T]{t}, id, activatedStepDir, workDir)
	require.NoError(t, err)

	absStepSrcDir, err := pathutil.AbsPath(stepSrcDir)
	require.NoError(t, err)

	require.Equal(t, ActivationTypePathRef, got.ActivationType)
	require.Equal(t, filepath.Join(workDir, "current_step.yml"), got.StepYMLPath)
	require.Equal(t, "path", got.StepInfo.Library)
	require.Equal(t, absStepSrcDir, got.StepInfo.ID)
	require.Empty(t, got.StepInfo.Version)
	require.Equal(t, filepath.Join(absStepSrcDir, "step.yml"), got.StepInfo.DefinitionPth)
	require.NotNil(t, got.StepInfo.Step.Title)
	require.Equal(t, "Hello Step", *got.StepInfo.Step.Title)
}

func TestActivateGitRefStep(t *testing.T) {
	srcRepoDir := initTestStepRepo(t)

	activatedStepDir := t.TempDir()
	workDir := t.TempDir()

	id := stepid.CanonicalID{SteplibSource: "git", IDorURI: srcRepoDir}

	got, err := ActivateGitRefStep(TestLogger[*testing.T]{t}, id, activatedStepDir, workDir)
	require.NoError(t, err)

	require.Equal(t, ActivationTypeGitRef, got.ActivationType)
	require.Equal(t, filepath.Join(workDir, "current_step.yml"), got.StepYMLPath)
	require.Equal(t, "git", got.StepInfo.Library)
	require.Equal(t, srcRepoDir, got.StepInfo.ID)
	require.Empty(t, got.StepInfo.Version)
	require.Equal(t, filepath.Join(activatedStepDir, "step.yml"), got.StepInfo.DefinitionPth)
	require.NotNil(t, got.StepInfo.Step.Title)
	require.Equal(t, "Hello Step", *got.StepInfo.Step.Title)
}

// initTestStepRepo creates a local git repo containing a step.yml and a single
// commit, usable as a clone source for git-ref activation.
func initTestStepRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeTestStepYML(t, dir)

	git := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
	}
	git("init")
	git("add", ".")
	git("commit", "-m", "initial")

	return dir
}
