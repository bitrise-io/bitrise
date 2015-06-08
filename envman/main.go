package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh/terminal"
	"github.com/bitrise-io/go-pathutil"
	"github.com/codegangsta/cli"
)

const envMapName string = "environments.yml"

var envmanDir string = pathutil.UserHomeDir() + "/.envman/"

var envMapPath string = envmanDir + envMapName

var stdinValue string

func createEnvmanDir() error {
	path := envmanDir
	exist, _ := pathutil.IsPathExists(path)
	if exist {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

func loadEnvMap() (envMap, error) {
	environments, err := readEnvMapFromFile(envMapPath)
	if err != nil {
		fmt.Println("Failed to read envlist, err: %s", err)
		return envMap{}, err
	}

	return environments, nil
}

func loadEnvMapOrCreate() (envMap, error) {
	environments, err := loadEnvMap()
	if err != nil {
		if err != (errors.New("No environemt variable list found")) {
			//return envListYMLStruct{}, err
			fmt.Println("Error: %s", err)
		}

		err := createEnvmanDir()
		if err != nil {
			fmt.Println("Failed to create envlist, err: %s", err)
			return envMap{}, err
		}
	}

	return environments, nil
}

func updateOrAddToEnvlist(environments envMap, newEnv envMap) (envMap, error) {
	fmt.Println(environments, newEnv)

	newEnvironments := make(envMap)

	for key, value := range environments {
		newEnvironments[key] = value
	}

	for key, value := range newEnv {
		newEnvironments[key] = value
	}

	err := writeEnvMapToFile(envMapPath, newEnvironments)
	if err != nil {
		fmt.Println("Failed to create store envlist, err: %s", err)
	}

	return newEnvironments, nil
}

func addCommand(c *cli.Context) {
	key := c.String("key")
	value := c.String("value")
	if stdinValue != "" {
		value = stdinValue
	}

	value = strings.Replace(value, "\n", "", -1)

	// Validate input
	if key == "" {
		fmt.Println("Invalid environment variable key")
		return
	}
	if value == "" {
		fmt.Println("Invalid environment variable value")
		return
	}

	// Load envlist, or create if not exist
	environments, err := loadEnvMapOrCreate()
	if err != nil {
		fmt.Println("Failed to load envlist, err: %s", err)
		return
	}

	// Add or update envlist
	newEnv := envMap{key: value}

	fmt.Println("New env: ", newEnv)
	fmt.Println("Old envs: ", environments)

	environments, err = updateOrAddToEnvlist(environments, newEnv)

	//	newEnvStruct := envYMLStruct{envKey, envValue}
	//	newEnvList, err := updateOrAddToEnvlist(envlist, newEnvStruct)
	if err != nil {
		fmt.Println("Failed to create store envlist, err: %s", err)
		return
	}
	fmt.Println("New env list: ", environments)

	return
}

func exportCommand(c *cli.Context) {
	environments, err := loadEnvMap()

	if err != nil {
		fmt.Println("Failed to export environemt variable list, err: %s", err)
		return
	}
	if len(environments) == 0 {
		fmt.Println("Empty environemt variable list")
		return
	}

	for key, value := range environments {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return
}

func runCommand(c *cli.Context) {
	environments, err := loadEnvMap()
	if err != nil {
		fmt.Println("Failed to export environemt variable list, err: %s", err)
	}

	doCmdEnvs := environments
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
