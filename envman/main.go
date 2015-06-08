package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh/terminal"
	"github.com/bitrise-io/go-pathutil"
	"github.com/codegangsta/cli"
)

const envMapName string = ".envstore.yml"

var (
	envmanDir  string = pathutil.UserHomeDir() + "/.envman/"
	envMapPath string = envmanDir + envMapName
	stdinValue string
)

func createEnvmanDir() error {
	if exist, _ := pathutil.IsPathExists(envmanDir); exist {
		return nil
	}
	return os.MkdirAll(envmanDir, 0755)
}

func loadEnvMap() (envMap, error) {
	environments, err := readEnvMapFromFile(envMapPath)
	if err != nil {
		return envMap{}, err
	}

	return environments, nil
}

func loadEnvMapOrCreate() (envMap, error) {
	environments, err := loadEnvMap()
	if err != nil {
		err := createEnvmanDir()
		if err != nil {
			fmt.Println("Failed to create envlist, err:%s", err)
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
		fmt.Println("Failed to create store envlist, err:%s", err)
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
		log.Fatalln("Invalid environment variable key")
	}
	if value == "" {
		log.Fatalln("Invalid environment variable value")
	}

	// Load envlist, or create if not exist
	environments, err := loadEnvMapOrCreate()
	if err != nil {
		log.Fatalln("Failed to load envlist, err:", err)
	}

	// Add or update envlist
	newEnv := envMap{key: value}
	environments, err = updateOrAddToEnvlist(environments, newEnv)
	if err != nil {
		log.Fatalln("Failed to create store envlist, err:", err)
	}
	fmt.Println("New env list: ", environments)

	return
}

func printCommand(c *cli.Context) {
	environments, err := loadEnvMap()
	if err != nil {
		log.Fatalln("Failed to export environment variable list, err:", err)
	}
	if len(environments) == 0 {
		fmt.Println("Empty environment variable list")
	} else {
		fmt.Println(environments)
	}

	return
}

func runCommand(c *cli.Context) {
	environments, err := loadEnvMap()
	if err != nil {
		log.Fatalln("Failed to export environment variable list, err:", err)
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
			log.Fatalln("Failed to read stdin, err:", err)
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
			Action: printCommand,
		},
		{
			Name:            "run",
			SkipFlagParsing: true,
			Action:          runCommand,
		},
	}

	app.Run(os.Args)
}
