package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/stepman/stepman"
	"github.com/bitrise-tools/colorstring"
	"github.com/urfave/cli"
)

const (
	// ShareFilename ...
	ShareFilename string = "share.json"
)

// ShareModel ...
type ShareModel struct {
	Collection string
	StepID     string
	StepTag    string
}

// ShareBranchName ...
func (share ShareModel) ShareBranchName() string {
	return share.StepID + "-" + share.StepTag
}

// DeleteShareSteplibFile ...
func DeleteShareSteplibFile() error {
	return command.RemoveDir(getShareFilePath())
}

// ReadShareSteplibFromFile ...
func ReadShareSteplibFromFile() (ShareModel, error) {
	if exist, err := pathutil.IsPathExists(getShareFilePath()); err != nil {
		return ShareModel{}, err
	} else if !exist {
		return ShareModel{}, errors.New("No share steplib found")
	}

	bytes, err := fileutil.ReadBytesFromFile(getShareFilePath())
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
	bytes, err := json.MarshalIndent(share, "", "\t")
	if err != nil {
		log.Errorf("Failed to parse json, error: %s", err)
		return err
	}

	return fileutil.WriteBytesToFile(getShareFilePath(), bytes)
}

// GuideTextForStepAudit ...
func GuideTextForStepAudit(toolMode bool) string {
	name := "stepman"
	if toolMode {
		name = "bitrise"
	}

	b := colorstring.NewBuilder()
	b.Plain("First, you need to ensure that your step is stored in a ").Blue("public git repository").NewLine()
	b.Plain("and it follows our ").Blue("step development guideline").Plain(": https://github.com/bitrise-io/bitrise/blob/master/_docs/step-development-guideline.md.").NewLine()
	b.NewLine()
	b.Plain("To audit your step on your local machine call ").Blue("$ %s audit --step-yml path/to/your/step.yml", name)
	return b.String()
}

// GuideTextForStart ...
func GuideTextForStart() string {
	b := colorstring.NewBuilder()
	b.Blue("Fork the StepLib repository ").Plain("you want to share your Step in.").NewLine()
	b.Plain(`You can find the main ("official") StepLib repository at `).Plain("https://github.com/bitrise-io/bitrise-steplib")
	return b.String()
}

// GuideTextForShareStart ...
func GuideTextForShareStart(toolMode bool) string {
	name := "stepman"
	if toolMode {
		name = "bitrise"
	}

	b := colorstring.NewBuilder()
	b.Plain("Call ").Blue("$ %s share start -c https://github.com/[your-username]/bitrise-steplib.git", name).Plain(", with the git clone URL of your forked StepLib repository.").NewLine()
	b.Plain("This will prepare your forked StepLib locally for sharing.")
	return b.String()
}

// GuideTextForShareCreate ...
func GuideTextForShareCreate(toolMode bool) string {
	name := "stepman"
	if toolMode {
		name = "bitrise"
	}

	b := colorstring.NewBuilder()
	b.Plain("Next, call ").Blue("$ %s share create --tag [step-version-tag] --git [step-git-uri].git --stepid [step-id]", name).Plain(",").NewLine()
	b.Plain("to add your Step to your forked StepLib repository (locally).").NewLine()
	b.NewLine()
	b.Yellow("Important: ").Plain("you have to add the (version) tag to your Step's repository.")
	return b.String()
}

// GuideTextForAudit ...
func GuideTextForAudit(toolMode bool) string {
	name := "stepman"
	if toolMode {
		name = "bitrise"
	}

	b := colorstring.NewBuilder()
	b.Plain("You can call ").Blue("$ %s audit -c https://github.com/[your-username]/bitrise-steplib.git ", name).NewLine()
	b.Plain("to perform a complete health-check on your forked StepLib before submitting your Pull Request.").NewLine()
	b.NewLine()
	b.Plain("This can help you catch issues which might prevent your Step from being accepted.")
	return b.String()
}

// GuideTextForShareFinish ...
func GuideTextForShareFinish(toolMode bool) string {
	name := "stepman"
	if toolMode {
		name = "bitrise"
	}

	b := colorstring.NewBuilder()
	b.Plain("Almost done! You should review your Step's step.yml file (the one added to the local StepLib),").NewLine()
	b.Plain("and once you're happy with it call ").Blue("$ %s share finish", name).NewLine()
	b.NewLine()
	b.Plain("This will commit & push the step.yml into your forked StepLib repository.")
	return b.String()
}

// GuideTextForFinish ...
func GuideTextForFinish() string {
	b := colorstring.NewBuilder()
	b.Plain("The only remaining thing is to ").Blue("create a Pull Request").Plain(" in the original StepLib repository. And you are done!")
	return b.String()
}

func share(c *cli.Context) {
	toolMode := c.Bool(ToolMode)

	b := colorstring.NewBuilder()
	b.Plain("Do you want to share your own Step with the world? Awesome!").NewLine()
	b.NewLine()
	b.Plain("Just follow these steps:").NewLine()
	b.NewLine()
	b.Plain("0. ").Plain(GuideTextForStepAudit(toolMode)).NewLine()
	b.NewLine()
	b.Plain("1. ").Plain(GuideTextForStart()).NewLine()
	b.NewLine()
	b.Plain("2. ").Plain(GuideTextForShareStart(toolMode)).NewLine()
	b.NewLine()
	b.Plain("3. ").Plain(GuideTextForShareCreate(toolMode)).NewLine()
	b.NewLine()
	b.Plain("4. ").Plain(GuideTextForAudit(toolMode)).NewLine()
	b.NewLine()
	b.Plain("5. ").Plain(GuideTextForShareFinish(toolMode)).NewLine()
	b.NewLine()
	b.Plain("6. ").Plain(GuideTextForFinish()).NewLine()
	b.NewLine()
	fmt.Printf(b.String())
}

func getShareFilePath() string {
	return path.Join(stepman.GetStepmanDirPath(), ShareFilename)
}
