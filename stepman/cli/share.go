package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-pathutil/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/codegangsta/cli"
)

// ShareModel ...
type ShareModel struct {
	Collection string
	StepID     string
	StepTag    string
}

const (
	// ShareFilename ...
	ShareFilename string = "share.json"
)

var (
	shareFilePath string
)

// DeleteShareSteplibFile ...
func DeleteShareSteplibFile() error {
	if exist, err := pathutil.IsPathExists(shareFilePath); err != nil {
		return err
	} else if exist {
		if err := os.RemoveAll(shareFilePath); err != nil {
			return err
		}
	}
	return nil
}

// ReadShareSteplibFromFile ...
func ReadShareSteplibFromFile() (ShareModel, error) {
	if exist, err := pathutil.IsPathExists(shareFilePath); err != nil {
		return ShareModel{}, err
	} else if !exist {
		return ShareModel{}, errors.New("No share steplib found")
	}

	bytes, err := ioutil.ReadFile(shareFilePath)
	if err != nil {
		return ShareModel{}, err
	}

	share := ShareModel{}
	if err := json.Unmarshal(bytes, &share); err != nil {
		return ShareModel{}, err
	}

	return share, nil
}

// WriteShareSteplibToFile ...
func WriteShareSteplibToFile(share ShareModel) error {
	if exist, err := pathutil.IsPathExists(stepman.StepManDirPath); err != nil {
		return err
	} else if !exist {
		if err := os.MkdirAll(stepman.StepManDirPath, 0777); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(shareFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error("[STEPMAN] - Failed to close file:", err)
		}
	}()

	var bytes []byte
	bytes, err = json.Marshal(share)
	if err != nil {
		log.Error("[STEPMAN] - Failed to parse json:", err)
		return err
	}

	if _, err := file.Write(bytes); err != nil {
		return err
	}
	return nil
}

func share(c *cli.Context) {
	fmt.Println(`
To share your step walk througth this steps:

- Fork the steplib repo.
- Call 'stepman share start -c STEPLIB_REPO_FORK_GIT_URI', this will prepare your forked steplib locally.
- Next call 'stepman share create --tag STEP_VERSION_TAG --git STEP_GIT_URI', to add your step to steplib locally.
- After tehese 'stepman share finish', will automagically push these changes to your forked steplib repo
- Once you're happy with it create pull request


// - stepman setup -share -c https://github.com/gkiki90/bitrise-steplib.git
//   - You can find your step lib repo at: /Users/godrei/.stepman/step_collections/1438870289/collection
// - stepman create -c https://github.com/gkiki90/bitrise-steplib.git new-xcode-archive@1.2.0
//   - NOTE: mkdir -p ./steps/new-xcode-archive/1.2.0
//   - You can find your step's step.yml at: /Users/godrei/.stepman/step_collections/1438870289/collection/steps/new-xcode-archive/1.2.0/step.yml
//   - Open this step.yml, fill out the required infos
//   - Once you're happy with it and want to share:
//     - you can commit your changes in (it's just a regular git repository):
//       - cd /Users/godrei/.stepman/step_collections/1438870289/collection
//       - git checkout -b new-xcode-archive
//       - git add ./steps/new-xcode-archive/1.2.0/step.yml
//       - git commit -m 'new-xcode-archive 1.2.0'
//       - git push
//     - or call: stepman share -c https://github.com/gkiki90/bitrise-steplib.git new-xcode-archive@1.2.0
//       to do it automagically
// - Create a pull request

You can find a template step repository at: https://github.com/bitrise-io/bitrise-steplib/step-template/step.yml
`)
}

func init() {
	shareFilePath = stepman.StepManDirPath + "/" + ShareFilename
}
