package cli

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/command/git"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/urfave/cli"
)

const maxSummaryLength = 100

func printFinishCreate(share ShareModel, stepDirInSteplib string, toolMode bool) {
	fmt.Println()
	log.Infof(" * "+colorstring.Green("[OK]")+" Your Step (%s) (%s) added to local StepLib (%s).", share.StepID, share.StepTag, stepDirInSteplib)
	log.Infoln(" *      You can find your Step's step.yml at: " + colorstring.Greenf("%s/step.yml", stepDirInSteplib))
	fmt.Println()
	fmt.Println("   " + GuideTextForShareFinish(toolMode))
}

func getStepIDFromGit(git string) string {
	splits := strings.Split(git, "/")
	lastPart := splits[len(splits)-1]
	splits = strings.Split(lastPart, ".")
	return splits[0]
}

func create(c *cli.Context) error {
	toolMode := c.Bool(ToolMode)

	share, err := ReadShareSteplibFromFile()
	if err != nil {
		log.Error(err)
		log.Fatalln("You have to start sharing with `stepman share start`, or you can read instructions with `stepman share`")
	}

	// Input validation
	tag := c.String(TagKey)
	if tag == "" {
		log.Fatalln("No Step tag specified")
	}

	gitURI := c.String(GitKey)
	if gitURI == "" {
		log.Fatalln("No Step url specified")
	}

	stepID := c.String(StepIDKEy)
	if stepID == "" {
		stepID = getStepIDFromGit(gitURI)
	}
	if stepID == "" {
		log.Fatalln("No Step id specified")
	}
	r := regexp.MustCompile(`[a-z0-9-]+`)
	if find := r.FindString(stepID); find != stepID {
		log.Fatalln("StepID doesn't conforms to: [a-z0-9-]")
	}

	route, found := stepman.ReadRoute(share.Collection)
	if !found {
		log.Fatalf("No route found for collectionURI (%s)", share.Collection)
	}
	stepDirInSteplib := stepman.GetStepCollectionDirPath(route, stepID, tag)
	stepYMLPathInSteplib := path.Join(stepDirInSteplib, "step.yml")
	if exist, err := pathutil.IsPathExists(stepYMLPathInSteplib); err != nil {
		log.Fatalf("Failed to check step.yml path in steplib, err: %s", err)
	} else if exist {
		log.Warnf("Step already exist in path: %s.", stepDirInSteplib)
		if val, err := goinp.AskForBool("Would you like to overwrite local version of Step?"); err != nil {
			log.Fatalf("Failed to get bool, err: %s", err)
		} else {
			if !val {
				log.Errorln("Unfortunately we can't continue with sharing without an overwrite exist step.yml.")
				log.Fatalln("Please finish your changes, run this command again and allow it to overwrite the exist step.yml!")
			}
		}
	}

	// Clone Step to tmp dir
	tmp, err := pathutil.NormalizedOSTempDirPath("")
	if err != nil {
		log.Fatalf("Failed to get temp directory, err: %s", err)
	}

	log.Infof("Cloning Step from (%s) with tag (%s) to temporary path (%s)", gitURI, tag, tmp)

	repo, err := git.New(tmp)
	if err != nil {
		return err
	}

	if err := retry.Times(2).Wait(3 * time.Second).Try(func(attempt uint) error {
		return repo.CloneTagOrBranch(gitURI, tag).Run()
	}); err != nil {
		log.Fatalf("Failed to git-clone (url: %s) version (%s), error: %s",
			gitURI, tag, err)
	}

	// Update step.yml
	tmpStepYMLPath := path.Join(tmp, "step.yml")
	bytes, err := fileutil.ReadBytesFromFile(tmpStepYMLPath)
	if err != nil {
		log.Fatalf("Failed to read Step from file, err: %s", err)
	}
	var stepModel models.StepModel
	if err := yaml.Unmarshal(bytes, &stepModel); err != nil {
		log.Fatalf("Failed to unmarchal Step, err: %s", err)
	}

	commit, err := repo.RevParse("HEAD").RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		log.Fatalf("Failed to get commit hash, err: %s", err)
	}

	stepModel.Source = &models.StepSourceModel{
		Git:    gitURI,
		Commit: commit,
	}
	stepModel.PublishedAt = pointers.NewTimePtr(time.Now())

	// Validate step-yml
	if err := stepModel.Audit(); err != nil {
		log.Fatalf("Failed to validate Step, err: %s", err)
	}
	for _, input := range stepModel.Inputs {
		key, value, err := input.GetKeyValuePair()
		if err != nil {
			log.Fatalf("Failed to get Step input key-value pair, err: %s", err)
		}

		options, err := input.GetOptions()
		if err != nil {
			log.Fatalf("Failed to get Step input (%s) options, err: %s", key, err)
		}

		if len(options.ValueOptions) > 0 && value == "" {
			log.Warn("Step input with 'value_options', should contain default value!")
			log.Fatalf("Missing default value for Step input (%s).", key)
		}
	}
	if strings.Contains(*stepModel.Summary, "\n") {
		log.Warningln("Step summary should be one line!")
	}
	if utf8.RuneCountInString(*stepModel.Summary) > maxSummaryLength {
		log.Warningf("Step summary should contains maximum (%d) characters, actual: (%d)!", maxSummaryLength, utf8.RuneCountInString(*stepModel.Summary))
	}

	// Copy step.yml to steplib
	share.StepID = stepID
	share.StepTag = tag
	if err := WriteShareSteplibToFile(share); err != nil {
		log.Fatalf("Failed to save share steplib to file, err: %s", err)
	}

	log.Info("Step dir in collection:", stepDirInSteplib)
	if exist, err := pathutil.IsPathExists(stepDirInSteplib); err != nil {
		log.Fatalf("Failed to check path (%s), err: %s", stepDirInSteplib, err)
	} else if !exist {
		if err := os.MkdirAll(stepDirInSteplib, 0777); err != nil {
			log.Fatalf("Failed to create path (%s), err: %s", stepDirInSteplib, err)
		}
	}

	log.Infof("Checkout branch: %s", share.ShareBranchName())
	collectionDir := stepman.GetLibraryBaseDirPath(route)
	steplibRepo, err := git.New(collectionDir)
	if err != nil {
		log.Fatalf("Failed to init setplib repo: %s", err)
	}
	if err := steplibRepo.Checkout(share.ShareBranchName()).Run(); err != nil {
		if err := steplibRepo.NewBranch(share.ShareBranchName()).Run(); err != nil {
			log.Fatalf("Git failed to create and checkout branch, err: %s", err)
		}
	}

	stepBytes, err := yaml.Marshal(stepModel)
	if err != nil {
		log.Fatalf("Failed to marcshal Step model, err: %s", err)
	}
	if err := fileutil.WriteBytesToFile(stepYMLPathInSteplib, stepBytes); err != nil {
		log.Fatalf("Failed to write Step to file, err: %s", err)
	}

	// Update spec.json
	if err := stepman.ReGenerateLibrarySpec(route); err != nil {
		log.Fatalf("Failed to re-create steplib, err: %s", err)
	}

	printFinishCreate(share, stepDirInSteplib, toolMode)

	return nil
}
