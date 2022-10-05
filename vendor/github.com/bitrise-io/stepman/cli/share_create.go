package cli

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

const maxSummaryLength = 100

func getStepIDFromGit(git string) string {
	splits := strings.Split(git, "/")
	lastPart := splits[len(splits)-1]
	splits = strings.Split(lastPart, ".")
	return splits[0]
}

func validateTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("no Step tag specified")
	}

	parts := strings.Split(tag, ".")
	n := len(parts)

	if n != 3 {
		return fmt.Errorf("invalid semver format %s: %d parts instead of 3", tag, n)
	}

	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return fmt.Errorf("invalid semver format %s: %s", tag, err)
		}
	}

	return nil
}

func getDefaultStepGroupSpec() models.StepGroupInfoModel {
	return models.StepGroupInfoModel{
		Maintainer: "community",
	}
}

func createDefaultStepGroupSpec(route stepman.SteplibRoute, id string) error {
	marshalled, err := yaml.Marshal(getDefaultStepGroupSpec())
	if err != nil {
		return err
	}

	pth := stepman.GetStepGlobalInfoPath(route, id)
	return fileutil.WriteBytesToFile(pth, marshalled)
}

func isStepNew(route stepman.SteplibRoute, id string) (bool, error) {
	stepRootDir := path.Dir(stepman.GetStepGlobalInfoPath(route, id))
	exists, err := pathutil.IsDirExists(stepRootDir)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func create(c *cli.Context) error {
	toolMode := c.Bool(ToolMode)

	log.Infof("Validating Step share params...")

	share, err := ReadShareSteplibFromFile()
	if err != nil {
		log.Errorf(err.Error())
		failf("You need to start sharing with `stepman share start`, you can read instructions with `stepman share`")
	}

	// Input validation
	tag := c.String(TagKey)
	if err := validateTag(tag); err != nil {
		fail("validate tag: %s", err)
	}

	gitURI := c.String(GitKey)
	if gitURI == "" {
		failf("No Step url specified")
	}

	stepID := c.String(StepIDKey)
	if stepID == "" {
		stepID = getStepIDFromGit(gitURI)
	}
	if stepID == "" {
		failf("No Step id specified")
	}
	r := regexp.MustCompile(`[a-z0-9-]+`)
	if find := r.FindString(stepID); find != stepID {
		failf("StepID doesn't conform to: [a-z0-9-]+")
	}

	route, found := stepman.ReadRoute(share.Collection)
	if !found {
		fail("No route found for collectionURI (%s)", share.Collection)
	}

	stepDirInSteplib := stepman.GetStepCollectionDirPath(route, stepID, tag)
	stepYMLPathInSteplib := filepath.Join(stepDirInSteplib, "step.yml")
	if exist, err := pathutil.IsPathExists(stepYMLPathInSteplib); err != nil {
		fail("Failed to check step.yml path in steplib, err: %s", err)
	} else if exist {
		log.Printf("Step already exists in path: %s", stepDirInSteplib)
		log.Warnf("Sharing requires to work in a clean Step repository.")
		if val, err := goinp.AskForBool("Would you like to overwrite the local version of the Step?"); err != nil {
			fail("Failed to get bool, err: %s", err)
		} else {
			if !val {
				log.Errorf("Unfortunately, we can't continue with sharing without overwriting the existing step.yml.")
				failf("Please finish your changes, run this command again, and allow it to overwrite the existing step.yml!")
			}
		}
	}
	log.Donef("all inputs are valid")

	// Clone Step to tmp dir
	fmt.Println()
	log.Infof("Validating the Step...")

	tmp, err := pathutil.NormalizedOSTempDirPath("")
	if err != nil {
		fail("Failed to get temp directory, err: %s", err)
	}

	log.Printf("cloning Step repo from (%s) with tag (%s) to: %s", gitURI, tag, tmp)

	repo, err := git.New(tmp)
	if err != nil {
		return err
	}

	if err := retry.Times(2).Wait(3 * time.Second).Try(func(attempt uint) error {
		return repo.CloneTagOrBranch(gitURI, tag).Run()
	}); err != nil {
		fail("Failed to git-clone (url: %s) version (%s), error: %s",
			gitURI, tag, err)
	}

	// Update step.yml
	tmpStepYMLPath := filepath.Join(tmp, "step.yml")
	bytes, err := fileutil.ReadBytesFromFile(tmpStepYMLPath)
	if err != nil {
		fail("Failed to read Step from file, err: %s", err)
	}
	var stepModel models.StepModel
	if err := yaml.Unmarshal(bytes, &stepModel); err != nil {
		fail("Failed to unmarshal Step, err: %s", err)
	}

	commit, err := repo.RevParse("HEAD").RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		fail("Failed to get commit hash, err: %s", err)
	}

	stepModel.Source = &models.StepSourceModel{
		Git:    gitURI,
		Commit: commit,
	}
	stepModel.PublishedAt = pointers.NewTimePtr(time.Now())

	// Validate step-yml
	if err := stepModel.Audit(); err != nil {
		fail("Failed to validate Step, err: %s", err)
	}
	for _, input := range stepModel.Inputs {
		key, value, err := input.GetKeyValuePair()
		if err != nil {
			fail("Failed to get Step input key-value pair, err: %s", err)
		}

		options, err := input.GetOptions()
		if err != nil {
			fail("Failed to get Step input (%s) options, err: %s", key, err)
		}

		if len(options.ValueOptions) > 0 && value == "" {
			log.Warnf("Step input with 'value_options' should contain a default value!")
			fail("Missing default value for Step input (%s).", key)
		}
	}
	if strings.Contains(*stepModel.Summary, "\n") {
		log.Warnf("Step summary should be one line!")
	}
	if utf8.RuneCountInString(*stepModel.Summary) > maxSummaryLength {
		log.Warnf("Step summary should contain maximum (%d) characters, actual: (%d)!", maxSummaryLength, utf8.RuneCountInString(*stepModel.Summary))
	}
	log.Donef("step is valid")

	// Copy step.yml to steplib
	fmt.Println()
	log.Infof("Integrating the Step into the Steplib...")

	isStepNew, err := isStepNew(route, stepID)
	if err != nil {
		fail("Failed to check if step is new, err: %s", err)
	}

	share.StepID = stepID
	share.StepTag = tag
	if err := WriteShareSteplibToFile(share); err != nil {
		fail("Failed to save share steplib to file, err: %s", err)
	}

	log.Printf("step dir in collection: %s", stepDirInSteplib)
	if exist, err := pathutil.IsPathExists(stepDirInSteplib); err != nil {
		fail("Failed to check path (%s), err: %s", stepDirInSteplib, err)
	} else if !exist {
		if err := os.MkdirAll(stepDirInSteplib, 0777); err != nil {
			fail("Failed to create path (%s), err: %s", stepDirInSteplib, err)
		}
	}

	collectionDir := stepman.GetLibraryBaseDirPath(route)
	steplibRepo, err := git.New(collectionDir)
	if err != nil {
		fail("Failed to init steplib repo: %s", err)
	}
	if err := steplibRepo.Checkout(share.ShareBranchName()).Run(); err != nil {
		if err := steplibRepo.NewBranch(share.ShareBranchName()).Run(); err != nil {
			fail("Git failed to create and checkout branch, err: %s", err)
		}
	}

	stepBytes, err := yaml.Marshal(stepModel)
	if err != nil {
		fail("Failed to marshal Step model, err: %s", err)
	}
	if err := fileutil.WriteBytesToFile(stepYMLPathInSteplib, stepBytes); err != nil {
		fail("Failed to write Step to file, err: %s", err)
	}

	if isStepNew {
		if err := createDefaultStepGroupSpec(route, stepID); err != nil {
			fail("Failed to create step group spec for new step: %s", err)
		}
	}

	log.Printf("your Step (%s@%s) has been added to the local StepLib (%s).", share.StepID, share.StepTag, stepDirInSteplib)

	// Update spec.json
	if err := stepman.ReGenerateLibrarySpec(route); err != nil {
		fail("Failed to re-create steplib, err: %s", err)
	}

	log.Donef("the StepLib changes are now prepared on branch: %s", share.ShareBranchName())

	fmt.Println()
	log.Printf(GuideTextForShareFinish(toolMode))
	fmt.Println()

	return nil
}
