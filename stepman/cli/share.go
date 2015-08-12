package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
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
	var bytes []byte
	bytes, err := json.Marshal(share)
	if err != nil {
		log.Error("[STEPMAN] - Failed to parse json:", err)
		return err
	}

	return fileutil.WriteBytesToFile(shareFilePath, bytes)
}

// GuideTextForStart ...
func GuideTextForStart() string {
	guide := colorstring.Blue("Fork the StepLib repository") + " you want to share your Step in.\n" +
		`   You can find the main ("official") StepLib repository at: ` + colorstring.Green("https://github.com/bitrise-io/bitrise-steplib") + `

   ` + colorstring.Yellow("Note") + `: You can use any StepLib repository you like,
     the StepLib system is decentralized, you don't have to work with the main StepLib repository
     if you don't want to. Feel free to maintain and use your own (or your team's) Step Library.
`
	return guide
}

// GuideTextForShareStart ...
func GuideTextForShareStart() string {
	guide := "Call " + colorstring.Blue("'stepman share start -c STEPLIB_REPO_FORK_GIT_URL'") + ", with the " + colorstring.Yellow("git clone url") + " of " + colorstring.Yellow("your forked StepLib repository") + ".\n" +
		`   This will prepare your forked StepLib locally for sharing.

   For example, if you want to share your Step in the main StepLib repository you should call:
     ` + colorstring.Green("stepman share start -c https://github.com/[your-username]/bitrise-steplib.git") + `
	`
	return guide
}

// GuideTextForShareCreate ...
func GuideTextForShareCreate() string {
	guide := "Next, call " + colorstring.Blue("'stepman share create --tag STEP_VERSION_TAG --git STEP_GIT_URI --stepid STEP_ID'") + `,
   to add your Step to your forked StepLib repository (locally).

   This will copy the required step.yml file from your Step's repository.
   This is all what's required to add your step (or a new version) to a StepLib.

   ` + colorstring.Yellow("Important") + `: You have to add the (version) tag to your Step's repository before you would call this!
     You can do that at: https://github.com/[your-username]/[step-repository]/tags

   An example call:
     ` + colorstring.Green("stepman share create --tag 1.0.0 --git https://github.com/[your-username]/[step-repository].git --stepid my-awesome-step") + `

   ` + colorstring.Yellow("Note") + `: You'll still be able to modify the step.yml in the StepLib after this.
	`
	return guide
}

// GuideTextForShareFinish ...
func GuideTextForShareFinish() string {
	//
	guide := `Almost done! You should review your Step's description file (step.yml) which was created in the previous step,
   and once you're happy with it call: ` + colorstring.Blue("'stepman share finish'") + `

   This will commit & push the step.yml ` + colorstring.Yellow("into your forked StepLib repository") + `.
		`
	return guide
}

// GuideTextForFinish ...
func GuideTextForFinish() string {
	guide := "Everything is ready! The only remaning thing is to " + colorstring.Blue("create a Pull Request") + `.

   If you used the main StepLib repository then you can create a Pull Request
   at: ` + colorstring.Green("https://github.com/bitrise-io/bitrise-steplib/pulls") + `
	`
	return guide
}

func share(c *cli.Context) {
	guide := `
Do you want to ` + colorstring.Green("share ") + colorstring.Yellow("your ") + colorstring.Magenta("own ") + colorstring.Blue("Step") + ` with the world? Awesome!!
To get started you can find a template Step repository at: ` + colorstring.Green("https://github.com/bitrise-io/stepman/tree/master/_step_template") + `

Once you have your Step in a ` + colorstring.Yellow("public git repository") + ` you can share it with others.

To share your Step just follow these steps (pun intended ;) :

1. ` + GuideTextForStart() + `
2. ` + GuideTextForShareStart() + `
3. ` + GuideTextForShareCreate() + `
4. ` + GuideTextForShareFinish() + `
5. ` + GuideTextForFinish()
	fmt.Println(guide)
}

func init() {
	shareFilePath = stepman.StepManDirPath + "/" + ShareFilename
}
