package cli

import (
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

func auditStepBeforeShare(pth string) error {
	stepModel, err := stepman.ParseStepDefinition(pth, false)
	if err != nil {
		return err
	}
	return stepModel.AuditBeforeShare()
}

func detectStepIDAndVersionFromPath(pth string) (stepID, stepVersion string, err error) {
	pathComps := strings.Split(pth, "/")
	if len(pathComps) < 4 {
		err = fmt.Errorf("Path should contain at least 4 components: steps, step-id, step-version, step.yml: %s", pth)
		return
	}
	// we only care about the last 4 component of the path
	pathComps = pathComps[len(pathComps)-4:]
	if pathComps[0] != "steps" {
		err = fmt.Errorf("Invalid step.yml path, 'steps' should be included right before the step-id: %s", pth)
		return
	}
	if pathComps[3] != "step.yml" {
		err = fmt.Errorf("Invalid step.yml path, should end with 'step.yml': %s", pth)
		return
	}
	stepID = pathComps[1]
	stepVersion = pathComps[2]
	return
}

func auditStepBeforeSharePullRequest(pth string) error {
	stepID, version, err := detectStepIDAndVersionFromPath(pth)
	if err != nil {
		return err
	}

	stepModel, err := stepman.ParseStepDefinition(pth, false)
	if err != nil {
		return err
	}

	return auditStepModelBeforeSharePullRequest(stepModel, stepID, version)
}

func auditStepModelBeforeSharePullRequest(step models.StepModel, stepID, version string) error {
	if err := step.Audit(); err != nil {
		return fmt.Errorf("Failed to audit step infos, error: %s", err)
	}

	pth, err := pathutil.NormalizedOSTempDirPath(stepID + version)
	if err != nil {
		return fmt.Errorf("Failed to create a temporary directory for the step's audit, error: %s", err)
	}

	if step.Source == nil {
		return fmt.Errorf("Missing Source porperty")
	}

	err = retry.Times(2).Wait(3 * time.Second).Try(func(attempt uint) error {
		gitModel := git.New(pth)
		return gitModel.CloneTagOrBranch(step.Source.Git, pth, version).Run()
	})
	if err != nil {
		return fmt.Errorf("Failed to git-clone the step (url: %s) version (%s), error: %s",
			step.Source.Git, version, err)
	}

	latestCommit, err := git.GetCommitHashOfHead(pth)
	if err != nil {
		return fmt.Errorf("Failed to get git-latest-commit-hash, error: %s", err)
	}
	if latestCommit != step.Source.Commit {
		return fmt.Errorf("Step commit hash (%s) should be the  latest commit hash (%s) on git tag", step.Source.Commit, latestCommit)
	}

	return nil
}

func auditStepLibBeforeSharePullRequest(gitURI string) error {
	if exist, err := stepman.RootExistForLibrary(gitURI); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf("Missing routing for collection, call 'stepman setup -c %s' before audit", gitURI)
	}

	collection, err := stepman.ReadStepSpec(gitURI)
	if err != nil {
		return err
	}

	for stepID, stepGroup := range collection.Steps {
		log.Debugf("Start audit StepGrup, with ID: (%s)", stepID)
		for version, step := range stepGroup.Versions {
			log.Debugf("Start audit Step (%s) (%s)", stepID, version)
			if err := auditStepModelBeforeSharePullRequest(step, stepID, version); err != nil {
				log.Errorf(" * "+colorstring.Redf("[FAILED] ")+"Failed audit (%s) (%s)", stepID, version)
				return fmt.Errorf("   Error: %s", err.Error())
			}
			log.Infof(" * "+colorstring.Greenf("[OK] ")+"Success audit (%s) (%s)", stepID, version)
		}
	}
	return nil
}

func audit(c *cli.Context) error {
	// Input validation
	beforePR := c.Bool("before-pr")

	collectionURI := c.String("collection")
	if collectionURI != "" {
		if beforePR {
			log.Warnln("before-pr flag is used only for Step audit")
		}

		if err := auditStepLibBeforeSharePullRequest(collectionURI); err != nil {
			log.Fatalf("Audit Step Collection failed, err: %s", err)
		}
	} else {
		stepYMLPath := c.String("step-yml")
		if stepYMLPath != "" {
			if exist, err := pathutil.IsPathExists(stepYMLPath); err != nil {
				log.Fatalf("Failed to check path (%s), err: %s", stepYMLPath, err)
			} else if !exist {
				log.Fatalf("step.yml doesn't exist at: %s", stepYMLPath)
			}

			if beforePR {
				if err := auditStepBeforeSharePullRequest(stepYMLPath); err != nil {
					log.Fatalf("Step audit failed, err: %s", err)
				}
			} else {
				if err := auditStepBeforeShare(stepYMLPath); err != nil {
					log.Fatalf("Step audit failed, err: %s", err)
				}
			}

			log.Infof(" * "+colorstring.Greenf("[OK] ")+"Success audit (%s)", stepYMLPath)
		} else {
			log.Fatalln("'stepman audit' command needs --collection or --step-yml flag")
		}
	}

	return nil
}
