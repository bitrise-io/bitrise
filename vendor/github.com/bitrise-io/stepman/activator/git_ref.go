package activator

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/stepman/stepid"
	"github.com/bitrise-io/stepman/stepman"
)

func ActivateGitRefStep(
	log stepman.Logger,
	id stepid.CanonicalID,
	activatedStepDir string,
	workDir string,
) (ActivatedStep, error) {
	repo, err := git.New(activatedStepDir)
	if err != nil {
		return ActivatedStep{}, err
	}

	var cloneCmd *command.Model
	if id.Version == "" {
		cloneCmd = repo.Clone(id.IDorURI, "--depth=1")
	} else {
		cloneCmd = repo.CloneTagOrBranch(id.IDorURI, id.Version, "--depth=1")
	}
	if out, err := cloneCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if strings.HasPrefix(id.IDorURI, "git@") {
			log.Warnf(`Note: if the step's repository is an open source one,
you should probably use a "https://..." git clone URL,
instead of the "git@..." git clone URL which usually requires authentication
even if the repository is open source!`)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ActivatedStep{}, fmt.Errorf("command failed with exit status %d (%s): %w", exitErr.ExitCode(), cloneCmd.PrintableCommandArgs(), errors.New(out))
		}
		return ActivatedStep{}, err
	}

	stepYMLPath := filepath.Join(workDir, "current_step.yml")
	if err := command.CopyFile(filepath.Join(activatedStepDir, "step.yml"), stepYMLPath); err != nil {
		return ActivatedStep{}, err
	}

	return ActivatedStep{
		StepYMLPath:     stepYMLPath,
		DidStepLibUpdate: false,
		ActivationType: ActivationTypeGitRef,
		ExecutablePath: "",
	}, nil
}
