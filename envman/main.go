package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh/terminal"
	"github.com/bitrise-io/envman/pathutil"
	"github.com/codegangsta/cli"
)

const envlistName string = "environment_variables.yml"

var envmanDir string = pathutil.UserHomeDir() + "/.envman/"

var envlistPath string = envmanDir + envlistName

var stdinValue string

func createEnvmanDir() error {
	path := envmanDir
	exist, _ := pathutil.IsPathExists(path)
	if exist {
		return nil
	}
	return pathutil.CreateDir(path)
}

func loadEnvlist() (envListYMLStruct, error) {
	envlist, err := readEnvListFromFile(envlistPath)
	if err != nil {
		fmt.Println("Failed to read envlist, err: %s", err)
		return envListYMLStruct{}, err
	}

	return envlist, nil
}

func loadEnvlistOrCreate() (envListYMLStruct, error) {
	envlist, err := loadEnvlist()
	if err != nil {
		if err != (errors.New("No environemt variable list found")) {
			//return envListYMLStruct{}, err
			fmt.Println("Error: %s", err)
		}

		err := createEnvmanDir()
		if err != nil {
			fmt.Println("Failed to create envlist, err: %s", err)
			return envListYMLStruct{}, err
		}
	}

	return envlist, nil
}

func updateOrAddToEnvlist(envList envListYMLStruct, newEnvStruct envYMLStruct) (envListYMLStruct, error) {
	alreadyUsedKey := false
	var newEnvList []envYMLStruct
	for i := range envList.Envlist {
		oldEnvStruct := envList.Envlist[i]
		if oldEnvStruct.Key == newEnvStruct.Key {
			alreadyUsedKey = true
			newEnvList = append(newEnvList, newEnvStruct)
		} else {
			newEnvList = append(newEnvList, oldEnvStruct)
		}
	}
	if alreadyUsedKey == false {
		newEnvList = append(newEnvList, newEnvStruct)
	}
	envList.Envlist = newEnvList
	err := writeEnvListToFile(envlistPath, envList)
	if err != nil {
		fmt.Println("Failed to create store envlist, err: %s", err)
	}

	return envList, nil
}

func addCommand(c *cli.Context) {
	envKey := c.String("key")
	envValue := c.String("value")
	if stdinValue != "" {
		envValue = stdinValue
	}

	envValue = strings.Replace(envValue, "\n", "", -1)

	// Validate input
	if envKey == "" {
		fmt.Println("Invalid environment variable key")
		return
	}
	if envValue == "" {
		fmt.Println("Invalid environment variable value")
		return
	}

	// Load envlist, or create if not exist
	envlist, err := loadEnvlistOrCreate()
	if err != nil {
		fmt.Println("Failed to load envlist, err: %s", err)
		return
	}

	// Add or update envlist
	newEnvStruct := envYMLStruct{envKey, envValue}
	newEnvList, err := updateOrAddToEnvlist(envlist, newEnvStruct)
	if err != nil {
		fmt.Println("Failed to create store envlist, err: %s", err)
		return
	}
	fmt.Println("New env list: ", newEnvList)

	return
}

func exportCommand(c *cli.Context) {
	envlist, err := loadEnvlist()

	if err != nil {
		fmt.Println("Failed to export environemt variable list, err: %s", err)
		return
	}
	if len(envlist.Envlist) == 0 {
		fmt.Println("Empty environemt variable list")
		return
	}

	for i := range envlist.Envlist {
		env := envlist.Envlist[i]
		if os.Getenv(env.Key) == "" {
			os.Setenv(env.Key, env.Value)
		}
	}

	return
}

func runCommand(c *cli.Context) {
	exportCommand(c)

	doCmdEnvs := getCommandEnvironments()
	doCommand := c.Args()[0]
	doArgs := c.Args()[1:]

	cmdToSend := commandModel{
		Command:      doCommand,
		Environments: doCmdEnvs,
		Argumentums:  doArgs,
	}

	executeCmd(cmdToSend)

	return
}

func getCommandEnvironments() []environmentKeyValue {
	cmdEnvs := []environmentKeyValue{}

	envlist, err := loadEnvlist()

	if err != nil {
		fmt.Println("Failed to export environemt variable list, err: %s", err)
		return cmdEnvs
	}
	if len(envlist.Envlist) == 0 {
		fmt.Println("Empty environemt variable list")

		return cmdEnvs
	}

	for i := range envlist.Envlist {
		env := envlist.Envlist[i]
		cmdEnvItem := environmentKeyValue{
			Key:   env.Key,
			Value: os.Getenv(env.Key),
		}
		cmdEnvs = append(cmdEnvs, cmdEnvItem)
	}

	return cmdEnvs
}

func main() {
	// Read piped data
	if !terminal.IsTerminal(0) {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Print("Failed to read stdin, err: %s", err)
		}
		stdinValue = string(bytes)
	}

	// Parse cl
	app := cli.NewApp()
	app.Name = "envman"
	app.Usage = "Environment varaibale manager."
	app.Commands = []cli.Command{
		{
			Name: "add",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "key",
					Value: "",
				},
				cli.StringFlag{
					Name:  "value",
					Value: "",
				},
			},
			Action: addCommand,
		},
		{
			Name:   "print",
			Action: exportCommand,
		},
		{
			Name:   "env",
			Action: exportCommand,
		},
		{
			Name:            "run",
			SkipFlagParsing: true,
			Action:          runCommand,
		},
	}

	app.Run(os.Args)
}
