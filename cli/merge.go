package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bitrise-io/bitrise/configmerge"
	"github.com/bitrise-io/bitrise/models"
	"github.com/urfave/cli"
)

var mergeConfigCommand = cli.Command{
	Name:      "merge",
	Usage:     "Resolves includes in a bitrise.yml and merges them into a single file.",
	ArgsUsage: "args[0]: By default it looks for a bitrise.yml in the current directory, custom path can be specified as an argument.",
	Action:    mergeConfig,
	Flags:     []cli.Flag{},
}

func mergeConfig(c *cli.Context) error {
	var configPth string
	if c.Args().Present() {
		configPth = c.Args().First()
	} else {
		configPth = "bitrise.yml"
	}

	merger, err := configmerge.NewMerger("./")
	if err != nil {
		return fmt.Errorf("failed to create config merger: %s", err)
	}

	mergedConfigContent, configFileTree, err := merger.MergeConfig(configPth)
	if err != nil {
		return fmt.Errorf("failed to merge config: %s", err)
	}

	fmt.Println("config tree:")
	printConfigFileTree(*configFileTree)

	fmt.Println()
	fmt.Println("merge config:")
	fmt.Println(mergedConfigContent)

	return nil
}

func printConfigFileTree(configFileTree models.ConfigFileTreeModel) {
	b, err := json.MarshalIndent(configFileTree, "", "\t")
	if err == nil {
		fmt.Println(string(b))
	}
}
