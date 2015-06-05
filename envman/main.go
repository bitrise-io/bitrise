package main

import (	
	"fmt"
	"errors"
	"os"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/gkiki90/envman/pathutil"
	"github.com/gkiki90/envman/envutil"
	"code.google.com/p/go.crypto/ssh/terminal"
	"github.com/codegangsta/cli"
)

var stdinValue string
var extMap = map[string]bool{
	".sh" : true,
	".go" : true,
}

func loadEnvlist() (envutil.EnvListYMLStruct, error) {
	envlist, err := envutil.ReadEnvListFromFile(pathutil.EnvlistPath)
	if err != nil {
		fmt.Println("Failed to read envlist, err: %s", err)
		return envutil.EnvListYMLStruct{}, err
	}

	return envlist, nil
}

func loadEnvlistOrCreate() (envutil.EnvListYMLStruct, error) {
	envlist, err := loadEnvlist()
	if err != nil {
		if err != (errors.New("No environemt variable list found")) {
			//return envutil.EnvListYMLStruct{}, err
			fmt.Println("Error: %s", err)
		}

		err := pathutil.CreateEnvmanDir()
		if err != nil {
			fmt.Println("Failed to create envlist, err: %s", err)
			return envutil.EnvListYMLStruct{}, err
		}
	}

	return envlist, nil
}

func updateOrAddToEnvlist(envList envutil.EnvListYMLStruct, newEnvStruct envutil.EnvYMLStruct) (envutil.EnvListYMLStruct, error) {
	alreadyUsedKey := false
	var newEnvList []envutil.EnvYMLStruct
	for i := range envList.Envlist {
		oldEnvStruct := envList.Envlist[i]
		if oldEnvStruct.Key ==  newEnvStruct.Key {
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
	err := envutil.WriteEnvListToFile(pathutil.EnvlistPath, envList)
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
	newEnvStruct := envutil.EnvYMLStruct{ envKey, envValue }
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

	cmdToSend := CommandModel{
		Command: doCommand,
		Environments: doCmdEnvs,
		Argumentums: doArgs,
	}

	executeCmd(cmdToSend)

	return
}

func getCommandEnvironments() []EnvironmentKeyValue {
	cmdEnvs := []EnvironmentKeyValue{}

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
		cmdEnvItem := EnvironmentKeyValue{
			Key:   env.Key,
			Value: os.Getenv(env.Key),
		}
		cmdEnvs = append(cmdEnvs, cmdEnvItem)
	}

	return cmdEnvs
}

func visit(path string, f os.FileInfo, err error) error {
	ext := filepath.Ext(path)
	if extMap[ext] {
		fmt.Printf("Visited: %s\n", path)

		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println("Error reading file, err: %v", err)
			return err
		}

		s := string(bytes)

		fmt.Println(s)
	}
  	
  	return nil
} 

func get_envCommand(c *cli.Context) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
        fmt.Print("Failed to read path, err: %s", err)
    }

	err = filepath.Walk(dir, visit)
  	fmt.Printf("filepath.Walk() returned %v\n", err)
}

func main() {
	// Read piped data
	if ! terminal.IsTerminal(0) {
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
	app.Commands = []cli.Command {
		{
			Name: "add",
			Flags: []cli.Flag {
				cli.StringFlag {
			    Name: "key",
			    Value: "",
			  },
			  cli.StringFlag {
			    Name: "value",
			    Value: "",
			  },
			},
			Action: addCommand,
		},
		{
			Name: "print",
			Action: exportCommand,
		},
		{
			Name: "env",
			Action: exportCommand,
		},
		{
			Name: "run",
			SkipFlagParsing: true,
			Action: runCommand,
		},
		{
			Name: "get_env",
			Action: get_envCommand,
		},
	}

	app.Run(os.Args)
}
