package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/codegangsta/cli"
)

func Run() {
	// Parse cl
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "Step manager"
	app.Version = "0.0.1"

	app.Author = ""
	app.Email = ""

	/*
		app.Before = func(c *cli.Context) error {
			return nil
		}
	*/

	//app.Flags = flags
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		fmt.Println("Stepman finished:", err)
	}
}
