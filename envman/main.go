package main

func main() {
	Run()

	/*
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
	*/
}
